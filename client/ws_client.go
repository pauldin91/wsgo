package client

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"sync"

	"github.com/gorilla/websocket"
)

type WsClient struct {
	address              string
	errorChan            chan error
	conn                 *websocket.Conn
	wg                   *sync.WaitGroup
	ctx                  context.Context
	incomingMsgHandler   func([]byte)
	outgoingMsgHandler   func(net.Conn)
	outgoingWsMsgHandler func(*websocket.Conn)
}

func NewWsClient(ctx context.Context, address string) *WsClient {
	var ws = &WsClient{
		wg:                 &sync.WaitGroup{},
		address:            address,
		ctx:                ctx,
		incomingMsgHandler: func(b []byte) {},
		outgoingMsgHandler: func(c net.Conn) {},
	}
	return ws
}

func (ws *WsClient) OnMessageReceivedHandler(handler func([]byte)) {
	ws.incomingMsgHandler = handler
}

func (ws *WsClient) OnMessageParseHandler(handler func(net.Conn)) {
	ws.outgoingMsgHandler = handler
}
func (ws *WsClient) OnMessageParseWsHandler(handler func(*websocket.Conn)) {
	ws.outgoingWsMsgHandler = handler
	ws.wg.Add(1)
	go func() {
		defer ws.wg.Done()
		go ws.outgoingWsMsgHandler(ws.conn)
		ws.handle()
	}()
}

func (ws *WsClient) Connect() error {
	dialer := websocket.Dialer{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	conn, _, err := dialer.Dial(ws.address, nil)
	if err != nil {
		return err
	}

	ws.conn = conn
	if ws.conn != nil {
		ws.readFromServer()
	}
	return nil
}

func (ws *WsClient) GetConnId() string {
	return fmt.Sprintf("%p", ws.conn)

}

func (ws *WsClient) Close() {

	if ws.conn != nil {
		ws.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		ws.conn.Close()
	}
	ws.wg.Wait()
}

func (ws *WsClient) readFromServer() {
	ws.wg.Add(1)
	go func() {
		defer ws.wg.Done()

		ws.readSocketBuffer()
		ws.handle()
	}()
}

func (ws *WsClient) readSocketBuffer() {
	go func() {
		for {
			_, message, err := ws.conn.ReadMessage()
			if err != nil {
				ws.errorChan <- err
				return
			}
			ws.incomingMsgHandler(message)
		}
	}()
}

func (ws *WsClient) SendError(err error) {
	ws.errorChan <- err
}

func (ws *WsClient) handle() {
	for {
		select {
		case <-ws.ctx.Done():
			_ = ws.conn.Close()
			return

		case <-ws.errorChan:
			return

		}
	}
}
