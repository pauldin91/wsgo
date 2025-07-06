package client

import (
	"context"
	"os"

	"github.com/pauldin91/wsgo/internal"
)

type Client interface {
	Close()
	Connect()
	GetConnId() string
	HandleInputFrom(source *os.File, handler func(*os.File))
	Send(message string)
	SendError(err error)
}

func NewClient(ctx context.Context, addr string, protocol internal.Protocol) Client {
	switch protocol {
	case internal.TCP:
		return NewTcpClient(ctx, addr)
	case internal.WebSocket:
		return NewWsClient(ctx, addr)
	case internal.QUIC:
		panic("unimplemented")
	case internal.WebRTC:
		panic("unimplemented")
	default:
		panic("unsupported")
	}
}
