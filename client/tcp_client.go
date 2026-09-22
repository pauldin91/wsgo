package client

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"sync"
)

type TcpClient struct {
	conn net.Conn
	wg   *sync.WaitGroup

	onMessageReceivedHandler func([]byte)
	onConnectionEstablished  func(net.Conn)
}

func NewTcpClient(address string, msgReceivedHandler func([]byte), tlsConfig *tls.Config) *TcpClient {
	var err error
	var conn net.Conn

	if tlsConfig == nil {
		conn, err = net.Dial("tcp", address)
	} else {
		conn, err = tls.Dial("tcp", address, tlsConfig)
	}

	log.Printf("connected to server %s", address)
	if err != nil {
		return nil
	}

	return &TcpClient{
		wg:                       &sync.WaitGroup{},
		onMessageReceivedHandler: msgReceivedHandler,
		conn:                     conn,
	}
}

func (c *TcpClient) OnMessageReceivedHandler(handler func([]byte)) {
	c.onMessageReceivedHandler = handler
}

func (c *TcpClient) OnMessageParseHandler(handler func(net.Conn)) {
	c.onConnectionEstablished = handler
}

func (c *TcpClient) Send(msg []byte) error {

	if c.conn == nil {
		return fmt.Errorf("connection not established")
	}
	_, err := c.conn.Write([]byte(string(msg) + "\n"))
	return err
}

func (c *TcpClient) Connect(ctx context.Context) error {
	go func() {
		<-ctx.Done()
		c.Disconnect()
	}()

	reader := bufio.NewReader(c.conn)
	for {
		buffer, _, err := reader.ReadLine()
		if err != nil {
			return nil
		}
		c.onMessageReceivedHandler(buffer)
	}
}

func (c *TcpClient) GetConnId() string {
	return c.conn.LocalAddr().String()
}

func (c *TcpClient) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}

func (c *TcpClient) Disconnect() error {
	return c.conn.Close()
}
