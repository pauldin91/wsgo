package server

import (
	"context"
	"net"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type WsServer struct {
	address    string
	mutex      sync.Mutex
	sockets    map[string]*websocket.Conn
	ctx        context.Context
	wg         *sync.WaitGroup
	errChan    chan error
	msgHandler func([]byte)
	certFile   string
	certKey    string
}

func NewWsServer(ctx context.Context, serveAddress string) *WsServer {
	return &WsServer{
		address:    serveAddress,
		ctx:        ctx,
		sockets:    make(map[string]*websocket.Conn),
		errChan:    make(chan error),
		mutex:      sync.Mutex{},
		wg:         &sync.WaitGroup{},
		msgHandler: func(b []byte) {},
	}
}

func (server *WsServer) GetConnections() map[string]net.Conn {
	panic("unimplemented")
}

func (ws *WsServer) Start() error {
	http.HandleFunc("/ws", ws.wsHandler)
	go func() {

		if err := http.ListenAndServe(ws.address, nil); err != nil {
			ws.errChan <- err
		}
	}()

	ws.waitForSignal()
	return nil
}

func (ws *WsServer) StartTls() error {

	http.HandleFunc("/ws", ws.wsHandler)
	go func() {

		if err := http.ListenAndServeTLS(ws.address, ws.certFile, ws.certKey, nil); err != nil {
			ws.errChan <- err
		}
	}()
	ws.waitForSignal()
	return nil
}

func (ws *WsServer) waitForSignal() {
	go func() {
		for {
			select {
			case <-ws.ctx.Done():
				return
			case <-ws.errChan:
				return
			}
		}
	}()
}

func (server *WsServer) OnMessageReceived(handler func([]byte)) {
	server.msgHandler = handler
}

func (ws *WsServer) Shutdown() {
	for _, c := range ws.sockets {
		c.Close()
	}
	close(ws.errChan)
	ws.wg.Wait()
}

func (ws *WsServer) wsHandler(w http.ResponseWriter, r *http.Request) {
	var upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		ws.errChan <- err
		return
	}
	clientID := conn.RemoteAddr().String()

	ws.mutex.Lock()
	ws.sockets[clientID] = conn
	ws.mutex.Unlock()
	ws.handleConnection(ws.sockets[clientID])
}

func (ws *WsServer) handleConnection(conn *websocket.Conn) {
	defer ws.closeConnection(conn.RemoteAddr().String())

	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			break
		}
		ws.msgHandler(p)
		// con := ws.sockets[conn.RemoteAddr().String()].NetConn()
	}
}

func (ws *WsServer) closeConnection(clientID string) {
	ws.mutex.Lock()
	ws.sockets[clientID].Close()
	delete(ws.sockets, clientID)
	ws.mutex.Unlock()
}
