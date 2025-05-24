package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type WsServer struct {
	address string
	mutex   sync.Mutex
	sockets map[string]*websocket.Conn
	ctx     context.Context
	cancel  context.CancelFunc
	wg      *sync.WaitGroup
	errChan chan error
}

func NewWsServer(ctx context.Context, serveAddress string) WsServer {
	cotx, cancel := context.WithCancel(ctx)
	return WsServer{
		address: serveAddress,
		ctx:     cotx,
		cancel:  cancel,
		sockets: make(map[string]*websocket.Conn),
		errChan: make(chan error),
		mutex:   sync.Mutex{},
		wg:      &sync.WaitGroup{},
	}
}

func (ws *WsServer) Start() {
	go func() {
		http.HandleFunc("/ws", ws.wsHandler)
		log.Printf("INFO: WS server started on %s\n", ws.address)
		if err := http.ListenAndServe(ws.address, nil); err != nil {
			log.Fatal("Could not start WebSocket server:", err)
		}
	}()
	ws.waitForSignal()
}

func (ws *WsServer) StartTls(certFile, certKey string) {

	go func() {
		http.HandleFunc("/ws", ws.wsHandler)
		log.Printf("INFO: WS server started on %s\n", ws.address)
		if err := http.ListenAndServeTLS(ws.address, certFile, certKey, nil); err != nil {
			log.Fatal("Could not start WebSocket server:", err)
		}
	}()
}

func (ws *WsServer) waitForSignal() {
	go func() {
		for {
			select {
			case <-ws.ctx.Done():
				log.Println("[server] caught interrupt signal")
				ws.cancel()
				return
			case err := <-ws.errChan:
				log.Printf("error : %s\n", err)
				return
			}
		}
	}()
}

func (ws *WsServer) Shutdown() {
	for _, c := range ws.sockets {
		c.Close()
	}
	close(ws.errChan)
	ws.wg.Wait()
}

func (ws *WsServer) sendToClient(clientID string, message any) {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()

	if conn, ok := ws.sockets[clientID]; ok {
		if err := conn.WriteJSON(message); err != nil {
			log.Println("Error sending to client", clientID, ":", err)
			conn.Close()
			delete(ws.sockets, clientID)
		}
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
		http.Error(w, "Failed to upgrade connection", http.StatusBadRequest)
		return
	}
	clientID := fmt.Sprintf("%p", conn)

	ws.mutex.Lock()
	ws.sockets[clientID] = conn
	ws.mutex.Unlock()

	initialMsg := fmt.Sprintf("New client connected with id : %s\n", clientID)

	log.Print(initialMsg)
	ws.sendToClient(clientID, initialMsg)

	defer ws.closeConnection(clientID)

	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			break
		}
		msg := fmt.Sprintf("[client] %s says : %s\n", clientID, string(p))
		ws.broadcastMessage(msg)
		log.Printf("[client] %s: says %s\n", clientID, string(p))
	}
}

func (ws *WsServer) broadcastMessage(message string) {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()

	for clientID, conn := range ws.sockets {
		if err := conn.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
			log.Println("Error writing to client", clientID, ":", err)
			conn.Close()
			delete(ws.sockets, clientID)
		}
	}
}

func (ws *WsServer) closeConnection(clientID string) {
	ws.mutex.Lock()
	ws.sockets[clientID].Close()
	delete(ws.sockets, clientID)
	ws.mutex.Unlock()
	fmt.Println("Client disconnected:", clientID)
}
