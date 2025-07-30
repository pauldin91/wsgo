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
	ctx                context.Context
	tlsConf            *tls.Config
	msgReceivedHandler func([]byte)
	msgParseHandler    func(net.Conn)
}

func NewTcpClient(ctx context.Context, address string) *TcpClient {
	var ws = &TcpClient{
		wg:                 &sync.WaitGroup{},
		address:            address,
		ctx:                ctx,
		errorChan:          make(chan error),
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

func (ws *TcpClient) Send(msg []byte) {
	ws.conn.Write([]byte(string(msg) + "\n"))
}

func (ws *TcpClient) Connect() error {
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

	ws.wg.Add(1)
	go func() {
		defer ws.wg.Done()
		ws.readSocketBuffer()
		ws.handle()
	}()

	return nil
}

func (ws *TcpClient) GetConnId() string {
	return fmt.Sprintf("%p", ws.conn)

}

func (ws *TcpClient) Close() {

	if ws.conn != nil {
		ws.conn.Write([]byte("Remote host terminated connection"))
		ws.conn.Close()
	}
	ws.wg.Wait()
}

func (ws *TcpClient) readSocketBuffer() {
	go func() {
		for {
			reader := bufio.NewReader(ws.conn)
			buffer, _, err := reader.ReadLine()
			if err != nil {
				if errors.Is(err, net.ErrClosed) || errors.Is(err, io.EOF) {
					return
				}
				ws.errorChan <- err
				return
			}
			ws.msgReceivedHandler(buffer)
		}
	}()
}

func (ws *TcpClient) SendError(err error) {
	ws.errorChan <- err
}

func (ws *TcpClient) handle() {
	go func() {

		for {
			select {
			case <-ws.ctx.Done():
				_ = ws.conn.Close()
				return

			case <-ws.errorChan:
				return
			}
		}
	}()
}
