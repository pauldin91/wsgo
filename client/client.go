package client

import (
	"context"
	"net"

	"github.com/pauldin91/wsgo/internal"
)

type Client interface {
	Close()
	Connect() error
	GetConnId() string
	OnMessageReceivedHandler(func([]byte))
	OnMessageParseHandler(func(net.Conn))
	SendError(err error)
	Send([]byte)
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
