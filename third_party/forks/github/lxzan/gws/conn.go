package gws

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/binary"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/icha-senpai/note/third_party/forks/github/lxzan/gws/internal"
)

// WebSocket connection
type Conn struct {
	// Mutex to protect shared resources
	mu sync.Mutex

	// Session storage for storing session data
	ss SessionStorage

	// Atomic value for storing errors
	ev atomic.Value

	// Indicates if this is a server-side connection
	isServer bool

	// Subprotocol
	subprotocol string

	// Underlying network connection
	conn net.Conn

	// Configuration information
	config *Config

	// Buffered reader
	br *bufio.Reader

	// Continuation frame
	continuationFrame continuationFrame

	// Frame header
	fh frameHeader

	// Event handler
	handler Event

	// Closed state
	closed uint32

	// Read queue
	readQueue channel

	// Write queue
	writeQueue workerQueue

	// Deflater
	deflater *deflater

	// Decompressing dictionary sliding window
	dpsWindow slideWindow

	// Compressed dictionary sliding window
	cpsWindow slideWindow

	// Compression extension configuration
	pd PermessageDeflate
}

// ReadLoop
// Read messages in a loop.
// If HTTP Server is reused, it is recommended to enable goroutine, as blocking will prevent the context from being GC.
func (c *Conn) ReadLoop() {
	c.handler.OnOpen(c)

	// Infinite loop to read messages, if an error occurs, trigger the error event and exit the loop
	for {
		if err := c.readMessage(); err != nil {
			c.emitError(true, err)
			break
		}
	}

	err, ok := c.ev.Load().(error)
	_ = c.dispatchControl(OpcodeCloseConnection, nil, internal.SelectValue(ok, err, errEmpty))

	// Reclaim resources
	if c.isServer {
		c.br.Reset(nil)
		c.config.brPool.Put(c.br)
		c.br = nil
		if c.cpsWindow.enabled {
			c.config.cswPool.Put(c.cpsWindow.dict)
			c.cpsWindow.dict = nil
		}
		if c.dpsWindow.enabled {
			c.config.dswPool.Put(c.dpsWindow.dict)
			c.dpsWindow.dict = nil
		}
	}
}

// Checks if the connection is closed
func (c *Conn) isClosed() bool {
	return atomic.LoadUint32(&c.closed) == 1
}

// Handle the error event
func (c *Conn) emitError(reading bool, err error) {
	if err == nil {
		return
	}

	if atomic.CompareAndSwapUint32(&c.closed, 0, 1) {
		// Error code to be sent and cause of error
		var sendCode, sendErr = internal.CloseGoingAway, error(internal.CloseGoingAway)
		if reading {
			switch v := err.(type) {
			case internal.StatusCode:
				sendCode, sendErr = v, v
			case *internal.Error:
				sendCode, sendErr, err = v.Code, v.Err, v.Err
			default:
				sendCode, sendErr = internal.CloseNormalClosure, err
			}
		}

		var reason = append(sendCode.Bytes(), sendErr.Error()...)
		_ = c.writeClose(err, reason)
	}
}

// Handles the close event
func (c *Conn) emitClose(buf *bytes.Buffer) error {
	var responseCode = internal.CloseNormalClosure
	var realCode = internal.CloseNormalClosure.Uint16()
	switch buf.Len() {
	case 0:
		responseCode = 0
		realCode = 0
	case 1:
		responseCode = internal.CloseProtocolError
		realCode = uint16(buf.Bytes()[0])
		buf.Reset()
	default:
		var b [2]byte
		_, _ = buf.Read(b[0:])
		realCode = binary.BigEndian.Uint16(b[0:])
		switch realCode {
		case 1004, 1005, 1006, 1014, 1015:
			responseCode = internal.CloseProtocolError
		default:
			if realCode < 1000 || realCode >= 5000 || (realCode >= 1016 && realCode < 3000) {
				responseCode = internal.CloseProtocolError
			} else if realCode < 1016 {
				responseCode = internal.CloseNormalClosure
			} else {
				responseCode = internal.StatusCode(realCode)
			}
		}
		if !internal.CheckEncoding(c.config.CheckUtf8Enabled, uint8(OpcodeCloseConnection), buf.Bytes()) {
			responseCode = internal.CloseUnsupportedData
		}
	}
	if atomic.CompareAndSwapUint32(&c.closed, 0, 1) {
		_ = c.writeClose(&CloseError{Code: realCode, Reason: buf.Bytes()}, responseCode.Bytes())
	}
	return internal.CloseNormalClosure
}

// Sets the deadline for the connection
func (c *Conn) SetDeadline(t time.Time) error {
	err := c.conn.SetDeadline(t)
	c.emitError(false, err)
	return err
}

// Sets the deadline for read operations
func (c *Conn) SetReadDeadline(t time.Time) error {
	err := c.conn.SetReadDeadline(t)
	c.emitError(false, err)
	return err
}

// Sets the deadline for write operations
func (c *Conn) SetWriteDeadline(t time.Time) error {
	err := c.conn.SetWriteDeadline(t)
	c.emitError(false, err)
	return err
}

// Returns the local network address
func (c *Conn) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

// Returns the remote network address
func (c *Conn) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

// NetConn
// Gets the underlying TCP/TLS/KCP... connection
func (c *Conn) NetConn() net.Conn {
	return c.conn
}

// Controls whether the operating system should delay packet transmission in hopes of sending fewer packets (Nagle's algorithm).
// The default is true (no delay), meaning that data is sent as soon as possible after a Write.
func (c *Conn) SetNoDelay(noDelay bool) error {
	switch v := c.conn.(type) {
	case *net.TCPConn:
		return v.SetNoDelay(noDelay)

	case *tls.Conn:
		if netConn, ok := v.NetConn().(*net.TCPConn); ok {
			return netConn.SetNoDelay(noDelay)
		}
	}
	return nil
}

// Gets the negotiated sub-protocol
func (c *Conn) SubProtocol() string { return c.subprotocol }

// Gets the session storage
func (c *Conn) Session() SessionStorage { return c.ss }
