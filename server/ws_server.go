package server

import (
	"context"
	"crypto/tls"
	"log"
	"net"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type WsServer struct {
	address    string
	mutex      sync.Mutex
	sockets    map[string]*websocket.Conn
	wg         *sync.WaitGroup
	errChan    chan error
	msgHandler func([]byte)
	certFile   string
	certKey    string
	tls        *tls.Config
	server     *http.Server
	listener   net.Listener
	mux        *http.ServeMux
}

func NewWsServerWithCerts(serveAddress string, tlsConfig *tls.Config) *WsServer {
	server := &WsServer{
		address: serveAddress,
		sockets: make(map[string]*websocket.Conn),
		errChan: make(chan error, 1),
		mutex:   sync.Mutex{},
		wg:      &sync.WaitGroup{},
		mux:     http.NewServeMux(),
	}
	server.server = &http.Server{
		Addr:      serveAddress,
		TLSConfig: tlsConfig,
		Handler:   server.mux,
	}
	server.mux.HandleFunc("/ws", server.wsHandler)
	return server
}

func (server *WsServer) GetConnections() map[string]net.Conn {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	conns := make(map[string]net.Conn)
	for k, v := range server.sockets {
		conns[k] = v.UnderlyingConn()
	}
	return conns
}

func (ws *WsServer) Start(ctx context.Context) {
	ws.wg.Add(2)
	go func() {
		defer ws.wg.Done()
		var err error
		ln, err := net.Listen("tcp", ws.address)
		if err != nil {
			select {
			case ws.errChan <- err:
			default:
			}
			return
		}

		if ws.tls != nil {
			ws.listener = tls.NewListener(ln, ws.tls)
		} else {
			ws.listener = ln
		}
		err = ws.server.Serve(ws.listener)
		if err != nil && err != http.ErrServerClosed {
			select {
			case ws.errChan <- err:
			default:
			}
		}
	}()
	go func() {
		defer ws.wg.Done()
		ws.waitForSignal(ctx)
	}()
}

func (ws *WsServer) waitForSignal(ctx context.Context) {
	for {
		select {
		case rcv := <-ctx.Done():
			log.Printf("shutdown signal received %v\n", rcv)
			return
		case <-ws.errChan:
			return
		}
	}
}

func (server *WsServer) OnMessageReceived(handler func([]byte)) {
	server.msgHandler = handler
}

func (ws *WsServer) Shutdown() {
	if ws.server != nil {
		ws.server.Shutdown(context.Background())
	}
	ws.mutex.Lock()
	for _, c := range ws.sockets {
		c.Close()
	}
	ws.mutex.Unlock()
	ws.wg.Wait()
	if ws.errChan != nil {
		close(ws.errChan)
	}
}

func (ws *WsServer) wsHandler(w http.ResponseWriter, r *http.Request) {
	var upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		select {
		case ws.errChan <- err:
		default:
		}
		return
	}
	clientID := conn.RemoteAddr().String()

	ws.mutex.Lock()
	ws.sockets[clientID] = conn
	ws.mutex.Unlock()
	ws.wg.Add(1)
	go func() {
		defer ws.wg.Done()
		ws.handleConnection(conn)
	}()
}

func (ws *WsServer) handleConnection(conn *websocket.Conn) {
	defer ws.closeConnection(conn.RemoteAddr().String())

	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			break
		}
		ws.msgHandler(p)

	}
}

func (ws *WsServer) closeConnection(clientID string) {
	ws.mutex.Lock()
	err := ws.sockets[clientID].Close()
	if err != nil {
		log.Printf("error closing connection %v", err)
	}
	delete(ws.sockets, clientID)
	ws.mutex.Unlock()
}
