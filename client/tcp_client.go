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
	errorChan                chan error
	conn                     net.Conn
	connMutex                sync.RWMutex
	wg                       *sync.WaitGroup
	cancel                   context.CancelFunc
	tlsConfig                *tls.Config
	onMessageReceivedHandler func([]byte)
	onConnectionEstablished  func(net.Conn)
}

func NewTcpClient(address string) *TcpClient {
	return &TcpClient{
		wg:                       &sync.WaitGroup{},
		address:                  address,
		errorChan:                make(chan error, 1),
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
	c.connMutex.RLock()
	conn := c.conn
	c.connMutex.RUnlock()

	if conn == nil {
		return fmt.Errorf("connection not established")
	}
	_, err := conn.Write([]byte(string(msg) + "\n"))
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

	c.connMutex.Lock()
	c.conn = conn
	c.connMutex.Unlock()

	log.Printf("connected to server %s", c.address)

	ctx, c.cancel = context.WithCancel(ctx)
	c.wg.Add(2)
	go func() {
		defer c.wg.Done()
		c.readMessages()
	}()
	go func() {
		defer c.wg.Done()
		c.handleShutdown(ctx)
	}()

	c.onConnectionEstablished(conn)

	return nil
}

func (c *TcpClient) GetConnId() string {
	c.connMutex.RLock()
	defer c.connMutex.RUnlock()
	return fmt.Sprintf("%p", c.conn)
}

func (c *TcpClient) Close() {
	if c.cancel != nil {
		c.cancel()
	}
	c.connMutex.Lock()
	if c.conn != nil {
		c.conn.Close()
	}
	c.connMutex.Unlock()
	c.wg.Wait()
	close(c.errorChan)
}

func (c *TcpClient) readMessages() {
	c.connMutex.RLock()
	conn := c.conn
	c.connMutex.RUnlock()

	reader := bufio.NewReader(conn)
	for {
		buffer, _, err := reader.ReadLine()
		if err != nil {
			if !errors.Is(err, net.ErrClosed) && !errors.Is(err, io.EOF) {
				select {
				case c.errorChan <- err:
				default:
				}
			}
			return
		}
		c.onMessageReceivedHandler(buffer)
	}
}

func (c *TcpClient) SendError(err error) {
	select {
	case c.errorChan <- err:
	default:
	}
}

func (c *TcpClient) handleShutdown(ctx context.Context) {
	select {
	case <-ctx.Done():
	case <-c.errorChan:
	}
	c.connMutex.Lock()
	if c.conn != nil {
		c.conn.Close()
	}
	c.connMutex.Unlock()
}
