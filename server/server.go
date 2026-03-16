package server

import (
	"context"
	"fmt"

	"github.com/pauldin91/wsgo/protocol"
)

type Server interface {
	Start(context.Context)
	SendTo(protocol.Message) error
	OnMessageReceived(handler func([]byte))
	Broadcast([]byte) error
	Shutdown()
}

func NewServer(addr string, protocol string) (Server, error) {
	switch protocol {
	case "tcp":
		return NewTCPServer(addr), nil
	case "websocket", "ws":
		return NewWSServerWithCerts(addr, nil), nil
	case "quic":
		return NewQuicServer(addr), nil
	case "webrtc":
		return nil, fmt.Errorf("WebRTC protocol not yet implemented")
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", protocol)
	}
}
