package server

import (
	"context"
	"fmt"
	"net"
)

type Server interface {
	Start(context.Context)
	OnMessageReceived(handler func([]byte))
	GetConnections() map[string]net.Conn
	Shutdown()
}

func NewServer(ctx context.Context, addr string, protocol string) (Server, error) {
	switch protocol {
	case "tcp":
		return NewTcpServer(addr), nil
	case "websocket", "ws":
		return NewWsServerWithCerts(addr, nil), nil
	case "quic":
		return nil, fmt.Errorf("QUIC protocol not yet implemented")
	case "webrtc":
		return nil, fmt.Errorf("WebRTC protocol not yet implemented")
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", protocol)
	}
}
