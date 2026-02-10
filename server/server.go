package server

import (
	"context"
	"net"

	model "github.com/pauldin91/wsgo/model"
)

type Server interface {
	Start() error
	OnMessageReceived(handler func([]byte))
	GetConnections() map[string]net.Conn
	Shutdown()
}

func NewServer(ctx context.Context, addr string, protocol model.Protocol) Server {
	switch protocol {
	case model.TCP:
		return NewTcpServer(ctx, addr)
	case model.WebSocket:
		return NewWsServerWithCerts(ctx, addr, nil)
	case model.QUIC:
		panic("unimplemented")
	case model.WebRTC:
		panic("unimplemented")
	default:
		panic("unsupported")
	}
}
