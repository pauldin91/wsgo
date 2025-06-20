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
	"strings"
	"sync"
)

type TcpClient struct {
	address    string
	send       chan []byte
	socketChan chan []byte
	errorChan  chan error
	conn       net.Conn
	wg         *sync.WaitGroup
	ctx        context.Context
}

func NewTcpClient(ctx context.Context, address string) *TcpClient {
	context := context.WithoutCancel(ctx)
	var ws = &TcpClient{
		wg:         &sync.WaitGroup{},
		address:    address,
		ctx:        context,
		send:       make(chan []byte),
		socketChan: make(chan []byte),
		errorChan:  make(chan error),
	}
	return ws
}

func (ws *TcpClient) ConnectTls(conf *tls.Config) {

	conn, err := tls.Dial("tcp", ws.address, conf)
	if err != nil {
		log.Printf("connection error :%s\n", err)
		return
	}

	ws.conn = conn

	log.Printf("connected to server %s", ws.address)

	ws.wg.Add(1)
	go func() {
		defer ws.wg.Done()
		ws.readSocketBuffer()
		ws.handle()
	}()
}

func (ws *TcpClient) Connect() {
	conn, err := net.Dial("tcp", ws.address)
	if err != nil {
		log.Printf("connection error :%s\n", err)
		return
	}
	ws.conn = conn

	log.Printf("connected to server %s", ws.address)

	ws.wg.Add(1)
	go func() {
		defer ws.wg.Done()
		ws.readSocketBuffer()
		ws.handle()
	}()
}

func (ws *TcpClient) GetConnId() string {
	return fmt.Sprintf("%p", ws.conn)

}

func (ws *TcpClient) Close() {

	if ws.conn != nil {
		ws.conn.Write([]byte(fmt.Sprint("Remote host terminated connection")))
		ws.conn.Close()
	}
	close(ws.send)
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
			ws.socketChan <- buffer
		}
	}()
}

func (ws *TcpClient) Send(message string) {
	if ws.conn != nil {
		ws.send <- []byte(message + "\n")
	}
}

func (ws *TcpClient) handle() {
	for {
		select {
		case <-ws.ctx.Done():
			fmt.Println("[read] context cancelled, closing WebSocket...")
			_ = ws.conn.Close()
			return

		case msg, ok := <-ws.socketChan:
			if len(msg) > 0 && ok {
				log.Printf("[read] msg: %s", msg)
			}

		case err := <-ws.errorChan:
			fmt.Printf("[read] error: %v", err)
			return

		case msg, ok := <-ws.send:
			if !ok {
				fmt.Println("[write] send channel closed")
				return
			}
			_, err := (ws.conn).Write(msg)

			if err != nil {
				fmt.Println("[write] error", err)
				return
			}
		}
	}
}

func (ws *TcpClient) readInput(reader *bufio.Reader) {

	go func() {
		for {
			input, _, err := reader.ReadLine()
			if err != nil {
				ws.errorChan <- err
				fmt.Println("[console] Read error:", err)
				return
			}

			text := strings.TrimSpace(string(input))
			if text == "exit" {
				fmt.Println("[console] Exiting by user command.")
				return
			}
			ws.Send(text)
		}
	}()
}

func (ws *TcpClient) ListenForInput(reader *bufio.Reader) {
	ws.wg.Add(1)
	go func() {
		defer ws.wg.Done()

		ws.readInput(reader)
		ws.handle()

	}()
}

func (ws *TcpClient) Input(reader *bufio.Reader, handler func(*bufio.Reader)) {
	ws.wg.Add(1)
	go func() {
		defer ws.wg.Done()
		handler(reader)
	}()
}
