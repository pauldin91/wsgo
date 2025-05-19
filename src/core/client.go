package core

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

type WsClient struct {
	address string
	send    chan []byte
	conn    *websocket.Conn
	wg      *sync.WaitGroup
	ctx     context.Context
	cnl     context.CancelFunc
}

func NewWsClient(ctx context.Context, address string) *WsClient {
	context, cancel := context.WithCancel(ctx)
	var ws = &WsClient{
		address: address,
		send:    make(chan []byte),
		ctx:     context,
		wg:      &sync.WaitGroup{},
		cnl:     cancel,
	}
	ws.connect()
	ws.read()
	ws.write()
	ws.readFromConsole()
	return ws
}

func (ws *WsClient) Close() {

	if ws.conn != nil {
		ws.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		ws.conn.Close()
	}
	close(ws.send)
	ws.wg.Wait()
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

func (client *WsClient) readFromConsole() {
	client.wg.Add(1)
	go func() {
		defer client.wg.Done()
		// Input reader setup
		reader := bufio.NewReader(os.Stdin)

		messageChan := make(chan []byte)
		errorChan := make(chan error)

		// Reader goroutine (blocks on ReadMessage)
		go func() {
			for {
				input, err := reader.ReadString('\n')
				if err != nil {
					fmt.Println("[console] Read error:", err)
					return
				}

				text := strings.TrimSpace(input)
				if text == "exit" {
					fmt.Println("[console] Exiting by user command.")
					return
				}
				client.Send(text)
			}
		}()

		fmt.Println("[console] Type something (Ctrl+C to exit):")

		for {
			select {
			case <-client.ctx.Done():
				fmt.Println("[console] Exiting...")
				return

			case msg := <-messageChan:
				log.Printf("[read] msg: %s", msg)

			case err := <-errorChan:
				log.Printf("[read] error: %v", err)
				return
			}
		}
	}()
}

func (ws *WsClient) read() {
	ws.wg.Add(1)
	go func() {
		defer ws.wg.Done()

		messageChan := make(chan []byte)
		errorChan := make(chan error)

		// Reader goroutine (blocks on ReadMessage)
		go func() {
			for {
				_, message, err := ws.conn.ReadMessage()
				if err != nil {
					errorChan <- err
					return
				}
				messageChan <- message
			}
		}()

		for {
			select {
			case <-ws.ctx.Done():
				log.Println("[read] context cancelled, closing WebSocket...")
				_ = ws.conn.Close() // unblock ReadMessage if needed
				return

			case msg := <-messageChan:
				log.Printf("[read] msg: %s", msg)

			case err := <-errorChan:
				log.Printf("[read] error: %v", err)
				return
			}
		}
	}()
}

func (ws *WsClient) write() {
	ws.wg.Add(1)
	go func() {
		defer ws.wg.Done()
		for {
			select {
			case <-ws.ctx.Done():
				log.Println("[write] Exiting...")
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
	}()
}
