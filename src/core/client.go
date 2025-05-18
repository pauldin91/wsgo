package core

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

type WsClient struct {
	address string
	send    chan []byte
	conn    *websocket.Conn
	ctx     context.Context
	wg      *sync.WaitGroup
}

func NewWsClient(ctx context.Context, address string) *WsClient {
	var ws = &WsClient{
		address: address,
		send:    make(chan []byte),
		ctx:     ctx,
		wg:      &sync.WaitGroup{},
	}
	ws.connect()
	ws.wg.Add(1)
	go func() {
		defer ws.wg.Done()
		ws.write()
		ws.read()
	}()
	return ws
}

func (ws *WsClient) Close() {
	ws.wg.Wait()
	for {
		select {

		case sig := <-ws.ctx.Done():
			fmt.Printf("\nReceived signal: %s. Exiting...\n", sig)
			if ws.conn != nil {
				close(ws.send)
				ws.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				ws.conn.Close()
				ws.conn = nil
			}

			return
		}
	}
}

func (ws *WsClient) Send(message string) {
	ws.send <- []byte(message)
}

func (ws *WsClient) connect() {
	dialer := websocket.Dialer{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	conn, _, err := dialer.Dial(ws.address, nil)
	if err != nil {
		return
	}

	ws.conn = conn
	log.Printf("connected to server %s", ws.address)
}

func (ws *WsClient) read() {
	for {
		_, message, err := ws.conn.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			return
		}
		log.Printf("recv: %s", message)
	}
}

func (ws *WsClient) write() {
	for {
		select {
		case msg, ok := <-ws.send:
			if !ok {
				log.Println("send channel closed")
				return
			}
			err := ws.conn.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				log.Println("write:", err)
				return
			}
		}
	}
}
