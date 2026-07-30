package gws

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"runtime"
	"unsafe"

	"github.com/icha-senpai/note/third_party/forks/github/lxzan/gws/internal"
)

const frameHeaderSize = 14

type Opcode uint8

const (
	OpcodeContinuation    Opcode = 0x0
	OpcodeText            Opcode = 0x1
	OpcodeBinary          Opcode = 0x2
	OpcodeCloseConnection Opcode = 0x8
	OpcodePing            Opcode = 0x9
	OpcodePong            Opcode = 0xA
)

// Checks if the opcode is a data frame
func (c Opcode) isDataFrame() bool {
	return c <= OpcodeBinary
}

type CloseError struct {
	// Close code, indicating the reason for closing the connection
	Code uint16

	// Close reason, providing a detailed description of the closure
	Reason []byte
}

// Returns a description of the close error
func (c *CloseError) Error() string {
	return fmt.Sprintf("gws: connection closed, code=%d, reason=%s", c.Code, string(c.Reason))
}

var (
	errEmpty = errors.New("")

	// Failure to pass forensic authentication
	ErrUnauthorized = errors.New("gws: unauthorized")

	// Handshake error, request header does not pass checksum.
	ErrHandshake = errors.New("gws: handshake error")

	// Compression extension negotiation failed, please try to disable compression.
	ErrCompressionNegotiation = errors.New("gws: invalid compression negotiation")

	// Sub-protocol negotiation failed
	ErrSubprotocolNegotiation = errors.New("gws: sub-protocol negotiation failed")

	// Text message encoding error (must be utf8)
	ErrTextEncoding = errors.New("gws: invalid text encoding")

	// message is too large
	ErrMessageTooLarge = errors.New("gws: message too large")

	// Connection closed
	ErrConnClosed = net.ErrClosed

	// Unsupported network protocols
	ErrUnsupportedProtocol = errors.New("gws: unsupported protocol")
)

type Event interface {
	// WebSocket connection was successfully established
	OnOpen(socket *Conn)

	// Received a close frame from the other end of the network connection, or disconnected voluntarily due to an error in the IO process
	// In the former case, err can be asserted as *CloseError
	OnClose(socket *Conn, err error)

	// Received a ping frame
	OnPing(socket *Conn, payload []byte)

	// Received a pong frame
	OnPong(socket *Conn, payload []byte)

	// If ParallelEnabled is enabled, OnMessage is called in parallel. No recover is done.
	OnMessage(socket *Conn, message *Message)
}

type BuiltinEventHandler struct{}

// Default callback when a WebSocket connection is established
func (b BuiltinEventHandler) OnOpen(socket *Conn) {}

// Default callback when a WebSocket connection is closed
func (b BuiltinEventHandler) OnClose(socket *Conn, err error) {}

// Default callback when a ping frame is received, automatically replies with a pong frame
func (b BuiltinEventHandler) OnPing(socket *Conn, payload []byte) { _ = socket.WritePong(nil) }

// Default callback when a pong frame is received
func (b BuiltinEventHandler) OnPong(socket *Conn, payload []byte) {}

// Default callback when a text/binary message is received
func (b BuiltinEventHandler) OnMessage(socket *Conn, message *Message) {}

type frameHeader [frameHeaderSize]byte

// Returns the value of the FIN bit
func (c *frameHeader) GetFIN() bool {
	return ((*c)[0] >> 7) == 1
}

// Returns the value of the RSV1 bit
func (c *frameHeader) GetRSV1() bool {
	return ((*c)[0] << 1 >> 7) == 1
}

// Returns the value of the RSV2 bit
func (c *frameHeader) GetRSV2() bool {
	return ((*c)[0] << 2 >> 7) == 1
}

// Returns the value of the RSV3 bit
func (c *frameHeader) GetRSV3() bool {
	return ((*c)[0] << 3 >> 7) == 1
}

// Returns the opcode
func (c *frameHeader) GetOpcode() Opcode {
	return Opcode((*c)[0] << 4 >> 4)
}

// Returns the value of the mask bytes
func (c *frameHeader) GetMask() bool {
	return ((*c)[1] >> 7) == 1
}

// Returns the length code
func (c *frameHeader) GetLengthCode() uint8 {
	return (*c)[1] << 1 >> 1
}

// Sets the Mask bit to 1
func (c *frameHeader) SetMask() {
	(*c)[1] |= uint8(128)
}

// Sets the frame length and returns the offset
func (c *frameHeader) SetLength(n uint64) (offset int) {
	if n <= internal.ThresholdV1 {
		(*c)[1] += uint8(n)
		return 0
	} else if n <= internal.ThresholdV2 {
		(*c)[1] += 126
		binary.BigEndian.PutUint16((*c)[2:4], uint16(n))
		return 2
	} else {
		(*c)[1] += 127
		binary.BigEndian.PutUint64((*c)[2:10], n)
		return 8
	}
}

// Sets the mask
func (c *frameHeader) SetMaskKey(offset int, key [4]byte) {
	copy((*c)[offset:offset+4], key[0:])
}

// Generates a frame header
// Consider having a random number generator for each client connection
func (c *frameHeader) GenerateHeader(isServer bool, fin bool, compress bool, opcode Opcode, length int) (headerLength int, maskBytes []byte) {
	headerLength = 2
	var b0 = uint8(opcode)
	if fin {
		b0 += 128
	}
	if compress {
		b0 += 64
	}
	(*c)[0] = b0
	headerLength += c.SetLength(uint64(length))

	if !isServer {
		(*c)[1] |= 128
		maskNum := internal.AlphabetNumeric.Uint32()
		binary.LittleEndian.PutUint32((*c)[headerLength:headerLength+4], maskNum)
		maskBytes = (*c)[headerLength : headerLength+4]
		headerLength += 4
	}
	return
}

// Parses the complete protocol header, up to 14 bytes, and returns the payload length
func (c *frameHeader) Parse(reader io.Reader) (int, error) {
	if err := internal.ReadN(reader, (*c)[0:2]); err != nil {
		return 0, err
	}

	var payloadLength = 0
	var lengthCode = c.GetLengthCode()
	switch lengthCode {
	case 126:
		if err := internal.ReadN(reader, (*c)[2:4]); err != nil {
			return 0, err
		}
		payloadLength = int(binary.BigEndian.Uint16((*c)[2:4]))

	case 127:
		if err := internal.ReadN(reader, (*c)[2:10]); err != nil {
			return 0, err
		}
		payloadLength = int(binary.BigEndian.Uint64((*c)[2:10]))
	default:
		payloadLength = int(lengthCode)
	}

	var maskOn = c.GetMask()
	if maskOn {
		if err := internal.ReadN(reader, (*c)[10:14]); err != nil {
			return 0, err
		}
	}

	return payloadLength, nil
}

// Returns the mask
func (c *frameHeader) GetMaskKey() []byte {
	return (*c)[10:14]
}

type Message struct {
	// if the message is compressed
	compressed bool

	// opcode of the message
	Opcode Opcode

	// content of the message
	Data *bytes.Buffer
}

// Reads data from the message into the given byte slice p
func (c *Message) Read(p []byte) (n int, err error) {
	return c.Data.Read(p)
}

// Returns the byte slice of the message's data buffer
func (c *Message) Bytes() []byte {
	return c.Data.Bytes()
}

// Close message, recycling resources
func (c *Message) Close() error {
	binaryPool.Put(c.Data)
	c.Data = nil
	return nil
}

type continuationFrame struct {
	// Indicates if the frame is initialized
	initialized bool

	// Indicates if the frame is compressed
	compressed bool

	// The opcode of the frame
	opcode Opcode

	// The buffer for the frame data
	buffer *bytes.Buffer
}

// Resets the state of the continuation frame
func (c *continuationFrame) reset() {
	c.initialized = false
	c.compressed = false
	c.opcode = 0
	c.buffer = nil
}

// Logger interface
type Logger interface {
	// Printing the error log
	Error(v ...any)
}

// Standard Log Library
type stdLogger struct{}

// Printing the error log
func (c *stdLogger) Error(v ...any) {
	log.Println(v...)
}

// Exception recovery with logging of error messages
func Recovery(logger Logger) {
	if e := recover(); e != nil {
		const size = 64 << 10
		buf := make([]byte, size)
		buf = buf[:runtime.Stack(buf, false)]
		msg := *(*string)(unsafe.Pointer(&buf))
		logger.Error("fatal error:", e, msg)
	}
}
