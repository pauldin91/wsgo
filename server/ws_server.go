package server

import (
	"context"
	"crypto/tls"
	"errors"
	"log"
	"net"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/pauldin91/wsgo/protocol"
)

type WSServer struct {
	address                  string
	connectionsMutex         sync.RWMutex
	connections              map[string]*websocket.Conn
	wg                       *sync.WaitGroup
	errorChan                chan error
	onMessageReceivedHandler func([]byte)
	tlsConfig                *tls.Config
	httpServer               *http.Server
	listener                 net.Listener
	mux                      *http.ServeMux
}

func NewWSServerWithCerts(serveAddress string, tlsConfig *tls.Config) *WSServer {
	server := &WSServer{
		address:                  serveAddress,
		connections:              make(map[string]*websocket.Conn),
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

func (s *WSServer) Start(ctx context.Context) {
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
}

func (s *WSServer) OnMessageReceived(handler func([]byte)) {
	if handler != nil {
		s.onMessageReceivedHandler = handler
	}
}

func (s *WSServer) Shutdown() {
	if s.httpServer != nil {
		s.httpServer.Shutdown(context.Background())
	}
	s.connectionsMutex.Lock()
	for _, c := range s.connections {
		c.Close()
	}
	s.connectionsMutex.Unlock()
	s.wg.Wait()
	close(s.errorChan)
}

func (s *WSServer) wsHandler(w http.ResponseWriter, r *http.Request) {
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
	s.connections[clientID] = conn
	s.connectionsMutex.Unlock()
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.handleConnection(conn)
	}()
}

func (s *WSServer) handleConnection(conn *websocket.Conn) {
	defer s.closeConnection(conn.RemoteAddr().String())

	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			break
		}
		s.onMessageReceivedHandler(p)
	}
}

func (s *WSServer) closeConnection(clientID string) {
	s.connectionsMutex.Lock()
	if conn, exists := s.connections[clientID]; exists {
		err := conn.Close()
		if err != nil {
			log.Printf("error closing connection %v", err)
		}
		delete(s.connections, clientID)
	}
	s.connectionsMutex.Unlock()
}

func (s *WSServer) Broadcast(msg []byte) error {
	s.connectionsMutex.Lock()
	defer s.connectionsMutex.Unlock()
	for _, c := range s.connections {
		if err := c.WriteMessage(websocket.TextMessage, msg); err != nil {
			return err
		}
	}
	return nil
}

func (s *WSServer) SendTo(msg protocol.Message) error {
	s.connectionsMutex.RLock()
	defer s.connectionsMutex.RUnlock()
	if conn, ok := s.connections[msg.Receiver]; ok {
		message := protocol.Message{Sender: msg.Sender, Content: msg.Content}
		return conn.WriteJSON(message)

	}
	return errors.New("address not found")

}

func (s *WSServer) GetConnections() map[string]string {
	result := make(map[string]string)
	s.connectionsMutex.Lock()
	defer s.connectionsMutex.Unlock()
	for _, c := range s.connections {
		result[c.RemoteAddr().String()] = c.RemoteAddr().String()
	}
	return result

}
