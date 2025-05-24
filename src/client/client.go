package client

import (
	"context"
	"crypto/tls"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

type WsClient struct {
	address    string
	send       chan []byte
	socketChan chan []byte
	conn       *websocket.Conn
	wg         *sync.WaitGroup
	ctx        context.Context
	cnl        context.CancelFunc
}

func NewWsClient(ctx context.Context, address string) *WsClient {
	context, cancel := context.WithCancel(ctx)
	var ws = &WsClient{
		wg:         &sync.WaitGroup{},
		address:    address,
		ctx:        context,
		cnl:        cancel,
		send:       make(chan []byte),
		socketChan: make(chan []byte),
	}
	return ws
}

func (ws *WsClient) Connect() {
	ws.connect()
	if ws.conn != nil {
		ws.ReadFromServer()
	}
}

func (ws *WsClient) connect() {
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
}

func (ws *WsClient) Close() {

	if ws.conn != nil {
		ws.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		ws.conn.Close()
	}
	close(ws.send)
	ws.wg.Wait()
}

func (ws *WsClient) ReadFromServer() {
	ws.wg.Add(1)
	go func() {
		defer ws.wg.Done()

		errorChan := make(chan error)
		ws.readSocketBuffer(errorChan)
		ws.handle(errorChan)
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

func (ws *WsClient) WriteToServer() {
	ws.wg.Add(1)
	go func() {
		defer ws.wg.Done()
		ws.handle(make(chan error))
	}()
}

func (ws *WsClient) handle(errorChan chan error) {
	for {
		select {
		case <-ws.ctx.Done():
			log.Println("[read] context cancelled, closing WebSocket...")
			_ = ws.conn.Close()
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
			err := ws.conn.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				log.Println("[write] error", err)
				return
			}
		}
	}
}

/*
func (ws *WsClient) readInput(errorChan chan error) {
	reader := bufio.NewReader(os.Stdin)

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

func (ws *WsClient) readFromInput() {
	ws.wg.Add(1)
	go func() {
		defer ws.wg.Done()

		errorChan := make(chan error)
		ws.readInput(errorChan)

		ws.handle(errorChan)
	}()
}*/
