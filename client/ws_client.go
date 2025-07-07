package client

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/gorilla/websocket"
)

type WsClient struct {
	address    string
	send       chan []byte
	socketChan chan []byte
	errorChan  chan error
	conn       *websocket.Conn
	wg         *sync.WaitGroup
	ctx        context.Context
}

func NewWsClient(ctx context.Context, address string) *WsClient {
	var ws = &WsClient{
		wg:         &sync.WaitGroup{},
		address:    address,
		ctx:        ctx,
		send:       make(chan []byte),
		socketChan: make(chan []byte),
	}
	return ws
}

func (ws *WsClient) Connect() {
	dialer := websocket.Dialer{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	conn, _, err := dialer.Dial(ws.address, nil)
	if err != nil {
		log.Printf("connection error :%s\n", err)
		return
	}

	ws.conn = conn
	log.Printf("connected to server %s", ws.address)
	if ws.conn != nil {
		ws.readFromServer()
	}
}

func (ws *WsClient) GetConnId() string {
	return fmt.Sprintf("%p", ws.conn)

}

func (ws *WsClient) Close() {

	if ws.conn != nil {
		ws.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		ws.conn.Close()
	}
	close(ws.send)
	ws.wg.Wait()
}

func (ws *WsClient) readFromServer() {
	ws.wg.Add(1)
	go func() {
		defer ws.wg.Done()

		errorChan := make(chan error)
		ws.readSocketBuffer(errorChan)
		ws.handle()
	}()
}

func (ws *WsClient) readSocketBuffer(errorChan chan error) {
	go func() {
		for {
			_, message, err := ws.conn.ReadMessage()
			if err != nil {
				errorChan <- err
				return
			}
			ws.socketChan <- message
		}
	}()
}

func (ws *WsClient) Send(message string) {
	if ws.conn != nil {
		ws.send <- []byte(message)
	}
}

func (ws *WsClient) SendError(err error) {
	ws.errorChan <- err
}

func (ws *WsClient) HandleInputFrom(source *os.File, handler func(*os.File)) {
	ws.wg.Add(1)
	go func() {
		defer ws.wg.Done()
		handler(source)
		ws.handle()
	}()
}

func (ws *WsClient) handle() {
	for {
		select {
		case <-ws.ctx.Done():
			log.Println("[read] context cancelled, closing WebSocket...")
			_ = ws.conn.Close()
			return

		case msg := <-ws.socketChan:
			log.Printf("[read] msg: %s", msg)

		case err := <-ws.errorChan:
			log.Printf("[read] error: %v", err)
			return

		case msg, ok := <-ws.send:
			if !ok {
				log.Println("[write] send channel closed")
				return
			}
			err := ws.conn.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				log.Println("[write] error", err)
				return
			}
		}
	}
}
