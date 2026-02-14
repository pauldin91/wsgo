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
	address                  string
	connectionsMutex         sync.RWMutex
	websocketConns           map[string]*websocket.Conn
	wg                       *sync.WaitGroup
	errorChan                chan error
	onMessageReceivedHandler func([]byte)
	tlsConfig                *tls.Config
	httpServer               *http.Server
	listener                 net.Listener
	mux                      *http.ServeMux
}

func NewWsServerWithCerts(serveAddress string, tlsConfig *tls.Config) *WsServer {
	server := &WsServer{
		address:                  serveAddress,
		websocketConns:           make(map[string]*websocket.Conn),
		errorChan:                make(chan error, 1),
		wg:                       &sync.WaitGroup{},
		mux:                      http.NewServeMux(),
		onMessageReceivedHandler: func(bytes []byte) { log.Printf("Echo: %v\n", string(bytes)) },
	}
	server.httpServer = &http.Server{
		Addr:      serveAddress,
		TLSConfig: tlsConfig,
		Handler:   server.mux,
	}
	server.mux.HandleFunc("/ws", server.wsHandler)
	return server
}

func (s *WsServer) GetConnections() map[string]net.Conn {
	s.connectionsMutex.RLock()
	defer s.connectionsMutex.RUnlock()
	conns := make(map[string]net.Conn, len(s.websocketConns))
	for k, v := range s.websocketConns {
		conns[k] = v.UnderlyingConn()
	}
	return conns
}

func (s *WsServer) Start(ctx context.Context) {
	s.wg.Add(2)
	go func() {
		defer s.wg.Done()
		var err error
		ln, err := net.Listen("tcp", s.address)
		if err != nil {
			select {
			case s.errorChan <- err:
			default:
			}
			return
		}

		if s.tlsConfig != nil {
			s.listener = tls.NewListener(ln, s.tlsConfig)
		} else {
			s.listener = ln
		}
		err = s.httpServer.Serve(s.listener)
		if err != nil && err != http.ErrServerClosed {
			select {
			case s.errorChan <- err:
			default:
			}
		}
	}()
	go func() {
		defer s.wg.Done()
		s.waitForShutdown(ctx)
	}()
}

func (s *WsServer) waitForShutdown(ctx context.Context) {
	select {
	case rcv := <-ctx.Done():
		log.Printf("shutdown signal received %v\n", rcv)
	case <-s.errorChan:
	}
}

func (s *WsServer) OnMessageReceived(handler func([]byte)) {
	s.onMessageReceivedHandler = handler
}

func (s *WsServer) Shutdown() {
	if s.httpServer != nil {
		s.httpServer.Shutdown(context.Background())
	}
	s.connectionsMutex.Lock()
	for _, c := range s.websocketConns {
		c.Close()
	}
	s.connectionsMutex.Unlock()
	s.wg.Wait()
	close(s.errorChan)
}

func (s *WsServer) wsHandler(w http.ResponseWriter, r *http.Request) {
	var upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		select {
		case s.errorChan <- err:
		default:
		}
		return
	}
	clientID := conn.RemoteAddr().String()

	s.connectionsMutex.Lock()
	s.websocketConns[clientID] = conn
	s.connectionsMutex.Unlock()
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.handleConnection(conn)
	}()
}

func (s *WsServer) handleConnection(conn *websocket.Conn) {
	defer s.closeConnection(conn.RemoteAddr().String())

	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			break
		}
		s.onMessageReceivedHandler(p)
	}
}

func (s *WsServer) closeConnection(clientID string) {
	s.connectionsMutex.Lock()
	if conn, exists := s.websocketConns[clientID]; exists {
		err := conn.Close()
		if err != nil {
			log.Printf("error closing connection %v", err)
		}
		delete(s.websocketConns, clientID)
	}
	s.connectionsMutex.Unlock()
}
