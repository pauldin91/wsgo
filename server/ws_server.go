package server

import (
	"context"
	"crypto/tls"
	"encoding/json"
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
	connections              map[string]net.Conn
	wg                       *sync.WaitGroup
	onMessageReceivedHandler func([]byte)
	httpServer               *http.Server
	listener                 net.Listener
	mux                      *http.ServeMux
}

func NewWSServerWithCerts(serveAddress string, tlsConfig *tls.Config) *WSServer {
	var err error
	ln, err := net.Listen("tcp", serveAddress)
	if err != nil {
		return nil
	}

	server := &WSServer{
		address:                  serveAddress,
		connections:              make(map[string]net.Conn),
		wg:                       &sync.WaitGroup{},
		mux:                      http.NewServeMux(),
		listener:                 ln,
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
	s.httpServer.Serve(s.listener)

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
}

func (s *WSServer) wsHandler(w http.ResponseWriter, r *http.Request) {
	var upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	clientID := conn.RemoteAddr().String()

	s.connectionsMutex.Lock()
	s.connections[clientID] = conn.NetConn()
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
		if _, err := c.Write([]byte(string(msg) + "\n")); err != nil {
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
		json.Marshal(message)
		_, err := conn.Write([]byte(string(msg.Content) + "\n"))
		return err

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
