package server

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/pion/webrtc/v4"
)

type WebRTCServer struct {
	address                  string
	connectionsMutex         sync.RWMutex
	dataChannels             map[string]*webrtc.DataChannel
	wg                       *sync.WaitGroup
	errorChan                chan error
	onMessageReceivedHandler func([]byte)
	httpServer               *http.Server
	mux                      *http.ServeMux
}

func NewWebRTCServer(address string) *WebRTCServer {
	s := &WebRTCServer{
		address:                  address,
		dataChannels:             make(map[string]*webrtc.DataChannel),
		wg:                       &sync.WaitGroup{},
		errorChan:                make(chan error, 1),
		onMessageReceivedHandler: func(msg []byte) { log.Printf("Echo: %s\n", string(msg)) },
		mux:                      http.NewServeMux(),
	}
	s.httpServer = &http.Server{Addr: address, Handler: s.mux}
	s.mux.HandleFunc("/webrtc/offer", s.offerHandler)
	return s
}

func (s *WebRTCServer) Start(ctx context.Context) {
	s.wg.Add(2)
	go func() {
		defer s.wg.Done()
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			select {
			case s.errorChan <- err:
			default:
			}
		}
	}()
	go func() {
		defer s.wg.Done()
		select {
		case <-ctx.Done():
			log.Printf("shutdown signal received\n")
		case err := <-s.errorChan:
			log.Printf("error %v\n", err)
		}
	}()
}

func (s *WebRTCServer) offerHandler(w http.ResponseWriter, r *http.Request) {
	var offer webrtc.SessionDescription
	if err := json.NewDecoder(r.Body).Decode(&offer); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	pc, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pc.OnDataChannel(func(dc *webrtc.DataChannel) {
		clientID := dc.Label()
		s.connectionsMutex.Lock()
		s.dataChannels[clientID] = dc
		s.connectionsMutex.Unlock()

		dc.OnMessage(func(msg webrtc.DataChannelMessage) {
			s.onMessageReceivedHandler(msg.Data)
		})
		dc.OnClose(func() {
			s.connectionsMutex.Lock()
			delete(s.dataChannels, clientID)
			s.connectionsMutex.Unlock()
		})
	})

	if err = pc.SetRemoteDescription(offer); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	answer, err := pc.CreateAnswer(nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	gatherDone := webrtc.GatheringCompletePromise(pc)
	if err = pc.SetLocalDescription(answer); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	<-gatherDone

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pc.LocalDescription())
}

func (s *WebRTCServer) OnMessageReceived(handler func([]byte)) {
	s.onMessageReceivedHandler = handler
}

func (s *WebRTCServer) GetConnections() map[string]string {
	conn := make(map[string]string)

	return conn
}

func (s *WebRTCServer) Broadcast(msg []byte) error {
	s.connectionsMutex.Lock()
	defer s.connectionsMutex.Unlock()
	for _, dc := range s.dataChannels {
		if err := dc.Send(msg); err != nil {
			return err
		}
	}
	return nil
}

func (s *WebRTCServer) Shutdown() {
	if s.httpServer != nil {
		s.httpServer.Shutdown(context.Background())
	}
	s.connectionsMutex.Lock()
	for _, dc := range s.dataChannels {
		dc.Close()
	}
	s.connectionsMutex.Unlock()
	s.wg.Wait()
	close(s.errorChan)
}
