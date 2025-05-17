package core

import (
	"crypto/tls"
	"fmt"
	"log"
	"os"

	"github.com/gorilla/websocket"
)

type WsClient struct {
	address string
	done    chan bool
	err     chan error
	send    chan []byte
	conn    *websocket.Conn
}

func NewWsClient(address string) *WsClient {
	var ws = &WsClient{
		address: address,
		done:    make(chan bool),
		err:     make(chan error),
		send:    make(chan []byte),
	}

	go ws.connect()
	return ws
}

func (ws *WsClient) Wait(sigChan chan os.Signal) {
	for {
		select {
		case sig := <-sigChan:
			fmt.Printf("\nReceived signal: %s. Exiting...\n", sig)
			return
		case <-ws.done:
			log.Printf("dial %s error \n", ws.address)
			return
		case t := <-ws.err:
			log.Printf("error %s error \n", t)
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
		ws.err <- err
		return
	}

	ws.conn = conn
	log.Printf("connected to server %s", ws.address)

	go ws.write()
	ws.read()
}

func (ws *WsClient) read() {
	defer ws.conn.Close()

	for {
		_, message, err := ws.conn.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			ws.done <- true
			break
		}
		log.Printf("recv: %s", message)
	}
}

func (ws *WsClient) write() {
	for {
		select {
		case msg := <-ws.send:
			err := ws.conn.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				log.Println("write:", err)
				ws.done <- true
				return
			}
		case <-ws.done:
			return
		}
	}
}
