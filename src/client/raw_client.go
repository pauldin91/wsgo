package client

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
)

type TcpClient struct {
	address    string
	send       chan []byte
	socketChan chan []byte
	conn       *net.Conn
	wg         *sync.WaitGroup
	ctx        context.Context
	cnl        context.CancelFunc
}

func NewTcpClient(ctx context.Context, address string) *TcpClient {
	context, cancel := context.WithCancel(ctx)
	var ws = &TcpClient{
		wg:         &sync.WaitGroup{},
		address:    address,
		ctx:        context,
		cnl:        cancel,
		send:       make(chan []byte),
		socketChan: make(chan []byte),
	}
	return ws
}

func (ws *TcpClient) Connect() {
	ws.init()
	if ws.conn != nil {
		ws.readFromServer()
	}
}

func (ws *TcpClient) init() {

	conn, err := net.Dial("tcp", ws.address)
	if err != nil {
		log.Printf("connection error :%s\n", err)
		return
	}

	ws.conn = &conn
	log.Printf("connected to server %s", ws.address)
}

func (ws *TcpClient) GetConnId() string {
	return fmt.Sprintf("%p", ws.conn)

}

func (ws *TcpClient) Close() {

	if ws.conn != nil {
		(*ws.conn).Write([]byte(fmt.Sprint("Remote host terminated connection")))
		(*ws.conn).Close()
	}
	close(ws.send)
	ws.wg.Wait()
}

func (ws *TcpClient) readFromServer() {
	ws.wg.Add(1)
	go func() {
		defer ws.wg.Done()

		errorChan := make(chan error)
		ws.readSocketBuffer(errorChan)
		ws.handle(errorChan)
	}()
}

func (ws *TcpClient) readSocketBuffer(errorChan chan error) {
	go func() {
		for {
			var byteBuffer []byte
			_, err := (*ws.conn).Read(byteBuffer)
			if err != nil {
				errorChan <- err
				return
			}
			ws.socketChan <- byteBuffer
		}
	}()
}

func (ws *TcpClient) Send(message string) {
	if ws.conn != nil {
		ws.send <- []byte(message)
	}
}

func (ws *TcpClient) handle(errorChan chan error) {
	for {
		select {
		case <-ws.ctx.Done():
			log.Println("[read] context cancelled, closing WebSocket...")
			_ = (*ws.conn).Close()
			return

		case msg := <-ws.socketChan:
			log.Printf("[read] msg: %s", msg)

		case err := <-errorChan:
			log.Printf("[read] error: %v", err)
			return

		case msg, ok := <-ws.send:
			if !ok {
				log.Println("[write] send channel closed")
				return
			}
			_, err := (*ws.conn).Write(msg)
			if err != nil {
				log.Println("[write] error", err)
				return
			}
		}
	}
}

func (ws *TcpClient) readInput(reader *bufio.Reader, errorChan chan error) {

	go func() {
		for {
			input, err := reader.ReadString('\n')
			if err != nil {
				errorChan <- err
				fmt.Println("[console] Read error:", err)
				return
			}

			text := strings.TrimSpace(input)
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

		errorChan := make(chan error)
		ws.readInput(reader, errorChan)
		ws.handle(errorChan)

	}()
}
