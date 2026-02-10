package client

import (
	"context"
	"net"

	model "github.com/pauldin91/wsgo/model"
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

func NewClient(ctx context.Context, addr string, protocol model.Protocol) Client {
	switch protocol {
	case model.TCP:
		return NewTcpClient(ctx, addr)
	case model.WebSocket:
		return NewWsClient(ctx, addr)
	case model.QUIC:
		panic("unimplemented")
	case model.WebRTC:
		panic("unimplemented")
	default:
		panic("unsupported")
	}
}
