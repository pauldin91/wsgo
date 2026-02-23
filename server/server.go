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
	Broadcast([]byte) error
	Shutdown()
}

func NewServer(addr string, protocol string) (Server, error) {
	switch protocol {
	case "tcp":
		return NewTcpServer(addr), nil
	case "websocket", "ws":
		return NewWsServerWithCerts(addr, nil), nil
	case "quic":
		return NewQuicServer(addr), nil
	case "webrtc":
		return nil, fmt.Errorf("WebRTC protocol not yet implemented")
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", protocol)
	}
}
