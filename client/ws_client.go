package client

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

type WsClient struct {
	address              string
	errorChan            chan error
	conn                 *websocket.Conn
	wg                   *sync.WaitGroup
	cancel               context.CancelFunc
	incomingMsgHandler   func([]byte)
	outgoingMsgHandler   func(net.Conn)
	outgoingWsMsgHandler func()
}

func NewWsClient(address string) *WsClient {
	if address == "" {
		address = "ws://localhost:8080"
	}
	var ws = &WsClient{
		wg:        &sync.WaitGroup{},
		address:   address,
		errorChan: make(chan error, 1),
	}
	return ws
}

func (ws *WsClient) Send(msg []byte) error {
	if ws.conn == nil {
		return fmt.Errorf("connection not established")
	}
	return ws.conn.WriteMessage(websocket.TextMessage, msg)
}

func (ws *WsClient) OnMessageReceivedHandler(handler func([]byte)) {
	ws.incomingMsgHandler = handler
}

func (ws *WsClient) OnMessageParseHandler(handler func(net.Conn)) {
	ws.outgoingMsgHandler = handler
}
func (ws *WsClient) OnParseMsgHandler(src *os.File) {
	ws.outgoingWsMsgHandler = func() {
		reader := bufio.NewReader(src)
		for {
			input, _, err := reader.ReadLine()
			if err != nil {
				ws.SendError(err)
				return
			}
			text := strings.TrimSpace(string(input))
			if text == "exit" {
				return
			}
			if err := ws.Send([]byte(text + "\n")); err != nil {
				ws.SendError(err)
				return
			}
		}
	}
	ws.wg.Add(1)
	go func() {
		defer ws.wg.Done()
		ws.outgoingWsMsgHandler()
	}()
}

func (ws *WsClient) Connect(ctx context.Context) error {
	dialer := websocket.Dialer{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	conn, _, err := dialer.Dial(ws.address, nil)
	if err != nil {
		return err
	}

	ws.conn = conn
	ctx, ws.cancel = context.WithCancel(ctx)
	ws.readFromServer(ctx)
	return nil
}

func (ws *WsClient) GetConnId() string {
	return fmt.Sprintf("%p", ws.conn)

}

func (ws *WsClient) Close() {
	if ws.cancel != nil {
		ws.cancel()
	}
	if ws.conn != nil {
		ws.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		ws.conn.Close()
	}
	ws.wg.Wait()
	if ws.errorChan != nil {
		close(ws.errorChan)
	}
}

func (ws *WsClient) readFromServer(ctx context.Context) {
	ws.wg.Add(2)
	go func() {
		defer ws.wg.Done()
		ws.readSocketBuffer()
	}()
	go func() {
		defer ws.wg.Done()
		ws.handle(ctx)
	}()
}

func (ws *WsClient) readSocketBuffer() {
	for {
		_, message, err := ws.conn.ReadMessage()
		if err != nil {
			select {
			case ws.errorChan <- err:
			default:
			}
			return
		}
		ws.incomingMsgHandler(message)
	}
}

func (ws *WsClient) SendError(err error) {
	select {
	case ws.errorChan <- err:
	default:
	}
}

func (ws *WsClient) handle(ctx context.Context) {
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
