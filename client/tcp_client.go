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
	address            string
	errorChan          chan error
	conn               net.Conn
	wg                 *sync.WaitGroup
	cancel             context.CancelFunc
	tlsConf            *tls.Config
	msgReceivedHandler func([]byte)
	msgParseHandler    func(net.Conn)
}

func NewTcpClient(address string) *TcpClient {
	var ws = &TcpClient{
		wg:                 &sync.WaitGroup{},
		address:            address,
		errorChan:          make(chan error, 1),
		msgReceivedHandler: func(b []byte) {},
		msgParseHandler:    func(c net.Conn) {},
	}
	return ws
}

func (ws *TcpClient) OnMessageReceivedHandler(handler func([]byte)) {
	ws.msgReceivedHandler = handler
}

func (ws *TcpClient) OnMessageParseHandler(handler func(net.Conn)) {
	ws.msgParseHandler = handler
}

func (ws *TcpClient) Send(msg []byte) error {
	if ws.conn == nil {
		return fmt.Errorf("connection not established")
	}
	_, err := ws.conn.Write([]byte(string(msg) + "\n"))
	return err
}

func (ws *TcpClient) Connect(ctx context.Context) error {
	var conn net.Conn
	var err error

	if ws.tlsConf == nil {
		conn, err = net.Dial("tcp", ws.address)
	} else {
		conn, err = tls.Dial("tcp", ws.address, ws.tlsConf)
	}

	if err != nil {
		return err
	}
	ws.conn = conn

	log.Printf("connected to server %s", ws.address)

	ctx, ws.cancel = context.WithCancel(ctx)
	ws.wg.Add(2)
	go func() {
		defer ws.wg.Done()
		ws.readSocketBuffer()
	}()
	go func() {
		defer ws.wg.Done()
		ws.handle(ctx)
	}()

	if ws.msgParseHandler != nil {
		ws.msgParseHandler(ws.conn)
	}

	return nil
}

func (ws *TcpClient) GetConnId() string {
	return fmt.Sprintf("%p", ws.conn)

}

func (ws *TcpClient) Close() {
	if ws.cancel != nil {
		ws.cancel()
	}
	if ws.conn != nil {
		ws.conn.Write([]byte("Remote host terminated connection"))
		ws.conn.Close()
	}
	ws.wg.Wait()
	if ws.errorChan != nil {
		close(ws.errorChan)
	}
}

func (ws *TcpClient) readSocketBuffer() {
	reader := bufio.NewReader(ws.conn)
	for {
		buffer, _, err := reader.ReadLine()
		if err != nil {
			if errors.Is(err, net.ErrClosed) || errors.Is(err, io.EOF) {
				return
			}
			select {
			case ws.errorChan <- err:
			default:
			}
			return
		}
		ws.msgReceivedHandler(buffer)
	}
}

func (ws *TcpClient) SendError(err error) {
	select {
	case ws.errorChan <- err:
	default:
	}
}

func (ws *TcpClient) handle(ctx context.Context) {

	for {
		select {
		case <-ctx.Done():
			_ = ws.conn.Close()
			return

		case <-ws.errorChan:
			return
		}
	}
}
