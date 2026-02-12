package protocol

type Type string

const (
	TCP       Type = "tcp"
	WebSocket Type = "websocket"
	QUIC      Type = "quic"
	WebRTC    Type = "webrtc"
)

func (t Type) String() string {
	return string(t)
}

func (t Type) IsValid() bool {
	switch t {
	case TCP, WebSocket, QUIC, WebRTC:
		return true
	default:
		return false
	}
}

// ParseType converts a string to a Type
func ParseType(s string) (Type, bool) {
	switch s {
	case "tcp":
		return TCP, true
	case "websocket", "ws":
		return WebSocket, true
	case "quic":
		return QUIC, true
	case "webrtc":
		return WebRTC, true
	default:
		return "", false
	}
}
