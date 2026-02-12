package client

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

type WsClient struct {
	address                 string
	errorChan               chan error
	conn                    *websocket.Conn
	connMutex               sync.RWMutex
	wg                      *sync.WaitGroup
	cancel                  context.CancelFunc
	onMessageReceivedFunc   func([]byte)
	onConnectionEstablished func(net.Conn)
	onInputReadyHandler     func()
}

func NewWsClient(address string) *WsClient {
	if address == "" {
		address = "ws://localhost:8080"
	}
	return &WsClient{
		wg:                      &sync.WaitGroup{},
		address:                 address,
		errorChan:               make(chan error, 1),
		onMessageReceivedFunc:   func(bytes []byte) { log.Printf("Received: %v\n", bytes) },
		onConnectionEstablished: func(c net.Conn) {},
		onInputReadyHandler:     func() {},
	}
}

func (c *WsClient) Send(msg []byte) error {
	c.connMutex.RLock()
	conn := c.conn
	c.connMutex.RUnlock()

	if conn == nil {
		return fmt.Errorf("connection not established")
	}
	return conn.WriteMessage(websocket.TextMessage, msg)
}

func (c *WsClient) OnMessageReceivedHandler(handler func([]byte)) {
	c.onMessageReceivedFunc = handler
}

func (c *WsClient) OnMessageParseHandler(handler func(net.Conn)) {
	c.onConnectionEstablished = handler
}
func (c *WsClient) OnParseMsgHandler(src *os.File) {
	c.onInputReadyHandler = func() {
		reader := bufio.NewReader(src)
		for {
			input, _, err := reader.ReadLine()
			if err != nil {
				c.SendError(err)
				return
			}
			text := strings.TrimSpace(string(input))
			if text == "exit" {
				return
			}
			if err := c.Send([]byte(text + "\n")); err != nil {
				c.SendError(err)
				return
			}
		}
	}
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		c.onInputReadyHandler()
	}()
}

func (c *WsClient) Connect(ctx context.Context) error {
	dialer := websocket.Dialer{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	conn, _, err := dialer.Dial(c.address, nil)
	if err != nil {
		return err
	}

	c.connMutex.Lock()
	c.conn = conn
	c.connMutex.Unlock()

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
	return nil
}

func (c *WsClient) GetConnId() string {
	c.connMutex.RLock()
	defer c.connMutex.RUnlock()
	return fmt.Sprintf("%p", c.conn)
}

func (c *WsClient) Close() {
	if c.cancel != nil {
		c.cancel()
	}
	c.connMutex.Lock()
	if c.conn != nil {
		c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		c.conn.Close()
	}
	c.connMutex.Unlock()
	c.wg.Wait()
	close(c.errorChan)
}

func (c *WsClient) readMessages() {
	c.connMutex.RLock()
	conn := c.conn
	c.connMutex.RUnlock()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			select {
			case c.errorChan <- err:
			default:
			}
			return
		}
		c.onMessageReceivedFunc(message)
	}
}

func (c *WsClient) SendError(err error) {
	select {
	case c.errorChan <- err:
	default:
	}
}

func (c *WsClient) handleShutdown(ctx context.Context) {
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
