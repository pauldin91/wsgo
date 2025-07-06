package server

import (
	"context"

	"github.com/pauldin91/wsgo/internal"
)

type Server interface {
	Start()
	StartTls()
	OnMessageReceived(handler func(msg string))
	Shutdown()
}

func NewServer(ctx context.Context, addr string, protocol internal.Protocol) Server {
	switch protocol {
	case internal.TCP:
		return NewTcpServer(ctx, addr)
	case internal.WebSocket:
		return NewWsServer(ctx, addr)
	case internal.QUIC:
		panic("unimplemented")
	case internal.WebRTC:
		panic("unimplemented")
	default:
		panic("unsupported")
	}
}
