package client

import (
	"bufio"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
)

type TcpClient struct {
	address                  string
	conn                     net.Conn
	wg                       *sync.WaitGroup
	tlsConfig                *tls.Config
	onMessageReceivedHandler func([]byte)
	onConnectionEstablished  func(net.Conn)
}

func NewTcpClient(address string) *TcpClient {
	return &TcpClient{
		wg:                       &sync.WaitGroup{},
		address:                  address,
		onMessageReceivedHandler: func(b []byte) {},
		onConnectionEstablished:  func(c net.Conn) {},
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
	var conn net.Conn
	var err error

	if c.tlsConfig == nil {
		conn, err = net.Dial("tcp", c.address)
	} else {
		conn, err = tls.Dial("tcp", c.address, c.tlsConfig)
	}

	if err != nil {
		return err
	}

	log.Printf("connected to server %s", c.address)

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		c.readMessages(ctx)
	}()

	c.onConnectionEstablished(conn)
	c.wg.Wait()

	return nil
}

func (c *TcpClient) GetConnId() string {
	return c.conn.LocalAddr().String()
}

func (c *TcpClient) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}

func (c *TcpClient) readMessages(ctx context.Context) {

	reader := bufio.NewReader(c.conn)
	for {
		buffer, _, err := reader.ReadLine()
		if err != nil {
			if !errors.Is(err, net.ErrClosed) && !errors.Is(err, io.EOF) {
				select {
				case <-ctx.Done():
					c.conn.Close()
					break
				default:
				}
			}
			return
		}
		c.onMessageReceivedHandler(buffer)
	}
}

func (c *TcpClient) Disconnect() error {
	return c.conn.Close()
}
