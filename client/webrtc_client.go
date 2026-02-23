package client

import (
	"context"
	"net"
)

type WebRTCClient struct {
	connID                string
	onMessageParseHandler func(net.Conn)
}

func NewWebRTCClient() WebRTCClient {
	return WebRTCClient{
		onMessageParseHandler: func(c net.Conn) {},
	}
}

func (c *WebRTCClient) Close() {}
func (c *WebRTCClient) Connect(context.Context) error {
	return nil
}
func (c *WebRTCClient) GetConnId() string {
	return c.connID
}
func (c *WebRTCClient) OnMessageReceivedHandler(func([]byte)) {}
func (c *WebRTCClient) OnMessageParseHandler(func(net.Conn)) {

}
func SendError(err error) {}
func Send([]byte) error {
	return nil
}
