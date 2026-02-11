package protocol

type Type string

const (
	TCP       Type = "tcp"
	WebSocket Type = "websocket"
	QUIC      Type = "quic"
	WebRTC    Type = "webrtc"
)

func (p Type) String() string {
	return string(p)
}

func (p Type) IsValid() bool {
	switch p {
	case TCP, WebSocket, QUIC, WebRTC:
		return true
	default:
		return false
	}
}
