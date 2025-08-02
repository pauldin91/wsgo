package server

import (
	"context"
	"net"

	"github.com/pauldin91/wsgo/internal"
)

type Server interface {
	Start() error
	OnMessageReceived(handler func([]byte))
	GetConnections() map[string]net.Conn
	Shutdown()
}

func NewServer(ctx context.Context, addr string, protocol internal.Protocol) Server {
	switch protocol {
	case internal.TCP:
		return NewTcpServer(ctx, addr)
	case internal.WebSocket:
		return NewWsServerWithCerts(ctx, addr, nil)
	case internal.QUIC:
		panic("unimplemented")
	case internal.WebRTC:
		panic("unimplemented")
	default:
		panic("unsupported")
	}
}
