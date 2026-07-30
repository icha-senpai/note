package internal

// map status codes to error messages
var closeErrorMap = map[StatusCode]string{
	0:                     "empty code",
	CloseNormalClosure:    "close normal",
	CloseGoingAway:        "client going away",
	CloseProtocolError:    "protocol error",
	CloseUnsupported:      "unsupported data",
	CloseNoStatusReceived: "no status",
	CloseAbnormalClosure:  "abnormal closure",
	CloseUnsupportedData:  "invalid payload data",
	ClosePolicyViolation:  "policy violation",
	CloseMessageTooLarge:  "message too large",
	CloseMissingExtension: "mandatory extension missing",
	CloseInternalErr:      "internal error",
	CloseServiceRestart:   "server restarting",
	CloseTryAgainLater:    "try again later",
	CloseTLSHandshake:     "TLS handshake error",
}

// websocket error code
type StatusCode uint16

const (
	CloseNormalClosure StatusCode = 1000

	CloseGoingAway StatusCode = 1001

	CloseProtocolError StatusCode = 1002

	CloseUnsupported StatusCode = 1003

	CloseNoStatusReceived StatusCode = 1005

	CloseAbnormalClosure StatusCode = 1006

	CloseUnsupportedData StatusCode = 1007

	ClosePolicyViolation StatusCode = 1008

	CloseMessageTooLarge StatusCode = 1009

	CloseMissingExtension StatusCode = 1010

	CloseInternalErr StatusCode = 1011

	CloseServiceRestart StatusCode = 1012

	CloseTryAgainLater StatusCode = 1013

	CloseTLSHandshake StatusCode = 1015
)

func (c StatusCode) Uint16() uint16 {
	return uint16(c)
}

func (c StatusCode) Bytes() []byte {
	if c == 0 {
		return []byte{}
	}
	return []byte{uint8(c >> 8), uint8(c << 8 >> 8)}
}

func (c StatusCode) Error() string {
	return "gws: " + closeErrorMap[c]
}

func NewError(code StatusCode, err error) *Error {
	return &Error{Code: code, Err: err}
}

type Error struct {
	Err  error
	Code StatusCode
}

func (c *Error) Error() string {
	return c.Err.Error()
}

// executes the passed functions in sequence and returns the first encountered error
func Errors(funcs ...func() error) error {
	for _, f := range funcs {
		if err := f(); err != nil {
			return err
		}
	}
	return nil
}
