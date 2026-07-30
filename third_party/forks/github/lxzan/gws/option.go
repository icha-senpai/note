package gws

import (
	"bufio"
	"crypto/tls"
	"net"
	"net/http"
	"time"

	"github.com/icha-senpai/note/third_party/forks/github/klauspost/compress/flate"
	"github.com/icha-senpai/note/third_party/forks/github/lxzan/gws/internal"
)

const (
	// Default parallel goroutine limit
	defaultParallelGolimit = 8

	// Default compression level
	defaultCompressLevel = flate.BestSpeed

	// Default maximum payload size for reading
	defaultReadMaxPayloadSize = 16 * 1024 * 1024

	// Default maximum payload size for writing
	defaultWriteMaxPayloadSize = 16 * 1024 * 1024

	// Default compression threshold
	defaultCompressThreshold = 512

	// Default compressor pool size
	defaultCompressorPoolSize = 32

	// Default read buffer size
	defaultReadBufferSize = 4 * 1024

	// Default write buffer size
	defaultWriteBufferSize = 4 * 1024

	// Default handshake timeout
	defaultHandshakeTimeout = 5 * time.Second

	// Default dial timeout
	defaultDialTimeout = 5 * time.Second
)

type (
	// For gws client, it is recommended to enable contextual takeover and not modify the sliding window index to provide the best compatibility.
	// For gws server, if you turn on context-side takeover, each connection takes up more memory, configure the sliding window index appropriately.
	PermessageDeflate struct {
		// Whether to turn on compression
		Enabled bool

		// Compress level
		Level int

		// Compression threshold, messages below the threshold will not be compressed, only for context-free takeover mode.
		Threshold int

		// Compressor memory pool size
		// The higher the value the lower the probability of competition, but it will consume a lot of memory
		PoolSize int

		// Server side context takeover
		ServerContextTakeover bool

		// Client side context takeover
		ClientContextTakeover bool

		// The server-side sliding window index
		// Range 8<=n<=15, means pow(2,n) bytes.
		ServerMaxWindowBits int

		// The client-side sliding window index
		// Range 8<=n<=15, means pow(2,n) bytes.
		ClientMaxWindowBits int
	}

	Config struct {
		// Memory pool for bufio.Reader
		brPool *internal.Pool[*bufio.Reader]

		// Big File Compressor
		bdPool *internal.Pool[*bigDeflater]

		// Memory pool for compressor sliding window
		cswPool *internal.Pool[[]byte]

		// Memory pool for decompressor sliding window
		dswPool *internal.Pool[[]byte]

		// Whether to enable parallel message processing
		ParallelEnabled bool

		// Limit on the number of concurrent goroutines used for parallel message processing (single connection)
		ParallelGolimit int

		// Maximum read message content length
		ReadMaxPayloadSize int

		// Size of the read buffer
		ReadBufferSize int

		// Maximum length of written message content
		WriteMaxPayloadSize int

		// Deprecated: Size of the write buffer, v1.4.5 version of this parameter is deprecated
		WriteBufferSize int

		// Whether to check the text utf8 encoding, turn off the performance will be better
		CheckUtf8Enabled bool

		// Message callback (OnMessage) recovery program
		Recovery func(logger Logger)

		// Logging tools
		Logger Logger
	}

	// Server configurations
	ServerOption struct {
		// Configuration
		config *Config

		// Deprecated: Size of the write buffer, v1.4.5 version of this parameter is deprecated
		WriteBufferSize int

		// PermessageDeflate configuration
		PermessageDeflate PermessageDeflate

		// Whether parallel processing is enabled
		ParallelEnabled bool

		// Parallel goroutine limit
		ParallelGolimit int

		// Maximum payload size for reading
		ReadMaxPayloadSize int

		// Read buffer size
		ReadBufferSize int

		// Maximum payload size for writing
		WriteMaxPayloadSize int

		// Whether UTF-8 check is enabled
		CheckUtf8Enabled bool

		// Logger
		Logger Logger

		// Recovery function
		Recovery func(logger Logger)

		// TLS configuration
		TlsConfig *tls.Config

		// Handshake timeout duration
		HandshakeTimeout time.Duration

		// WebSocket sub-protocol, handshake failure disconnects the connection
		SubProtocols []string

		// Additional response headers (may not be supported by the client)
		// https://www.rfc-editor.org/rfc/rfc6455.html#section-1.3
		ResponseHeader http.Header

		// Authentication function for connection establishment requests
		Authorize func(r *http.Request, session SessionStorage) bool

		// Create session storage space for custom SessionStorage implementations
		NewSession func() SessionStorage
	}
)

// When context takeover is enabled, all messages must be compressed regardless of length,
// otherwise the browser will report an error.
func (c *PermessageDeflate) setThreshold(isServer bool) {
	if (isServer && c.ServerContextTakeover) || (!isServer && c.ClientContextTakeover) {
		c.Threshold = 0
	}
}

// Removes protected WebSocket header fields
func (c *ServerOption) deleteProtectedHeaders() {
	c.ResponseHeader.Del(internal.Upgrade.Key)
	c.ResponseHeader.Del(internal.Connection.Key)
	c.ResponseHeader.Del(internal.SecWebSocketAccept.Key)
	c.ResponseHeader.Del(internal.SecWebSocketExtensions.Key)
	c.ResponseHeader.Del(internal.SecWebSocketProtocol.Key)
}

// Initialize server options
func initServerOption(c *ServerOption) *ServerOption {
	if c == nil {
		c = new(ServerOption)
	}
	if c.ReadMaxPayloadSize <= 0 {
		c.ReadMaxPayloadSize = defaultReadMaxPayloadSize
	}
	if c.ParallelGolimit <= 0 {
		c.ParallelGolimit = defaultParallelGolimit
	}
	if c.ReadBufferSize <= 0 {
		c.ReadBufferSize = defaultReadBufferSize
	}
	if c.WriteMaxPayloadSize <= 0 {
		c.WriteMaxPayloadSize = defaultWriteMaxPayloadSize
	}
	if c.WriteBufferSize <= 0 {
		c.WriteBufferSize = defaultWriteBufferSize
	}
	if c.Authorize == nil {
		c.Authorize = func(r *http.Request, session SessionStorage) bool { return true }
	}
	if c.NewSession == nil {
		c.NewSession = func() SessionStorage { return newSmap() }
	}
	if c.ResponseHeader == nil {
		c.ResponseHeader = http.Header{}
	}
	if c.HandshakeTimeout <= 0 {
		c.HandshakeTimeout = defaultHandshakeTimeout
	}
	if c.Logger == nil {
		c.Logger = defaultLogger
	}
	if c.Recovery == nil {
		c.Recovery = func(logger Logger) {}
	}

	if c.PermessageDeflate.Enabled {
		if c.PermessageDeflate.ServerMaxWindowBits < 8 || c.PermessageDeflate.ServerMaxWindowBits > 15 {
			c.PermessageDeflate.ServerMaxWindowBits = internal.SelectValue(c.PermessageDeflate.ServerContextTakeover, 12, 15)
		}
		if c.PermessageDeflate.ClientMaxWindowBits < 8 || c.PermessageDeflate.ClientMaxWindowBits > 15 {
			c.PermessageDeflate.ClientMaxWindowBits = internal.SelectValue(c.PermessageDeflate.ClientContextTakeover, 12, 15)
		}
		if c.PermessageDeflate.Threshold <= 0 {
			c.PermessageDeflate.Threshold = defaultCompressThreshold
		}
		if c.PermessageDeflate.Level == 0 {
			c.PermessageDeflate.Level = defaultCompressLevel
		}
		if c.PermessageDeflate.PoolSize <= 0 {
			c.PermessageDeflate.PoolSize = defaultCompressorPoolSize
		}
		c.PermessageDeflate.PoolSize = internal.ToBinaryNumber(c.PermessageDeflate.PoolSize)
	}

	c.deleteProtectedHeaders()

	c.config = &Config{
		ParallelEnabled:     c.ParallelEnabled,
		ParallelGolimit:     c.ParallelGolimit,
		ReadMaxPayloadSize:  c.ReadMaxPayloadSize,
		ReadBufferSize:      c.ReadBufferSize,
		WriteMaxPayloadSize: c.WriteMaxPayloadSize,
		WriteBufferSize:     c.WriteBufferSize,
		CheckUtf8Enabled:    c.CheckUtf8Enabled,
		Recovery:            c.Recovery,
		Logger:              c.Logger,
		brPool: internal.NewPool(func() *bufio.Reader {
			return bufio.NewReaderSize(nil, c.ReadBufferSize)
		}),
	}

	if c.PermessageDeflate.Enabled {
		c.config.bdPool = internal.NewPool[*bigDeflater](func() *bigDeflater {
			return newBigDeflater(true, c.PermessageDeflate)
		})
		if c.PermessageDeflate.ServerContextTakeover {
			windowSize := internal.BinaryPow(c.PermessageDeflate.ServerMaxWindowBits)
			c.config.cswPool = internal.NewPool[[]byte](func() []byte {
				return make([]byte, 0, windowSize)
			})
		}
		if c.PermessageDeflate.ClientContextTakeover {
			windowSize := internal.BinaryPow(c.PermessageDeflate.ClientMaxWindowBits)
			c.config.dswPool = internal.NewPool[[]byte](func() []byte {
				return make([]byte, 0, windowSize)
			})
		}
	}

	return c
}

// Get server configuration
func (c *ServerOption) getConfig() *Config { return c.config }

// Client configurations
type ClientOption struct {
	// Deprecated: Size of the write buffer, v1.4.5 version of this parameter is deprecated
	WriteBufferSize int

	// PermessageDeflate configuration
	PermessageDeflate PermessageDeflate

	// Whether parallel processing is enabled
	ParallelEnabled bool

	// Parallel goroutine limit
	ParallelGolimit int

	// Maximum payload size for reading
	ReadMaxPayloadSize int

	// Read buffer size
	ReadBufferSize int

	// Maximum payload size for writing
	WriteMaxPayloadSize int

	// Whether UTF-8 check is enabled
	CheckUtf8Enabled bool

	// Logger
	Logger Logger

	// Recovery function
	Recovery func(logger Logger)

	// Server address, e.g., wss://example.com/connect
	Addr string

	// Extra request headers
	RequestHeader http.Header

	// Handshake timeout duration
	HandshakeTimeout time.Duration

	// TLS configuration
	TlsConfig *tls.Config

	// The default is to return the net.Dialer instance.
	// Can also be used to set a proxy, for example:
	// NewDialer: func() (proxy.Dialer, error) {
	//     return proxy.SOCKS5("tcp", "127.0.0.1:1080", nil, nil)
	// },
	NewDialer func() (Dialer, error)

	// For custom SessionStorage implementations
	NewSession func() SessionStorage
}

// Initialize client options
func initClientOption(c *ClientOption) *ClientOption {
	if c == nil {
		c = new(ClientOption)
	}
	if c.ReadMaxPayloadSize <= 0 {
		c.ReadMaxPayloadSize = defaultReadMaxPayloadSize
	}
	if c.ParallelGolimit <= 0 {
		c.ParallelGolimit = defaultParallelGolimit
	}
	if c.ReadBufferSize <= 0 {
		c.ReadBufferSize = defaultReadBufferSize
	}
	if c.WriteMaxPayloadSize <= 0 {
		c.WriteMaxPayloadSize = defaultWriteMaxPayloadSize
	}
	if c.WriteBufferSize <= 0 {
		c.WriteBufferSize = defaultWriteBufferSize
	}
	if c.HandshakeTimeout <= 0 {
		c.HandshakeTimeout = defaultHandshakeTimeout
	}
	if c.RequestHeader == nil {
		c.RequestHeader = http.Header{}
	}
	if c.NewDialer == nil {
		c.NewDialer = func() (Dialer, error) { return &net.Dialer{Timeout: defaultDialTimeout}, nil }
	}
	if c.NewSession == nil {
		c.NewSession = func() SessionStorage { return newSmap() }
	}
	if c.Logger == nil {
		c.Logger = defaultLogger
	}
	if c.Recovery == nil {
		c.Recovery = func(logger Logger) {}
	}
	if c.PermessageDeflate.Enabled {
		if c.PermessageDeflate.ServerMaxWindowBits < 8 || c.PermessageDeflate.ServerMaxWindowBits > 15 {
			c.PermessageDeflate.ServerMaxWindowBits = 15
		}
		if c.PermessageDeflate.ClientMaxWindowBits < 8 || c.PermessageDeflate.ClientMaxWindowBits > 15 {
			c.PermessageDeflate.ClientMaxWindowBits = 15
		}
		if c.PermessageDeflate.Threshold <= 0 {
			c.PermessageDeflate.Threshold = defaultCompressThreshold
		}
		if c.PermessageDeflate.Level == 0 {
			c.PermessageDeflate.Level = defaultCompressLevel
		}
		c.PermessageDeflate.PoolSize = 1
	}
	return c
}

// Converts the ClientOption configuration to Config and returns it
func (c *ClientOption) getConfig() *Config {
	config := &Config{
		ParallelEnabled:     c.ParallelEnabled,
		ParallelGolimit:     c.ParallelGolimit,
		ReadMaxPayloadSize:  c.ReadMaxPayloadSize,
		ReadBufferSize:      c.ReadBufferSize,
		WriteMaxPayloadSize: c.WriteMaxPayloadSize,
		WriteBufferSize:     c.WriteBufferSize,
		CheckUtf8Enabled:    c.CheckUtf8Enabled,
		Recovery:            c.Recovery,
		Logger:              c.Logger,
	}
	return config
}
