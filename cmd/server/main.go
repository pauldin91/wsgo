package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/websocket"
	"github.com/pauldin91/wsgo/server"
)

func main() {
	// Signal handling setup
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	host := flag.String("host", ":4443", "Server host")
	flag.Parse()

	log.Printf("[main] starting TCP server on %s", *host)
	server := server.NewWsServer(ctx, *host)
	server.Start()
	server.OnWsMessageReceived(func(conn *websocket.Conn, msg []byte) {
		conn.WriteMessage(websocket.PingMessage, []byte(string(msg)+"\n"))
		conn.SetPongHandler(func(appData string) error {
			fmt.Printf("Received: %s\n", appData)
			return nil
		})
	})

	// p, _ := os.FindProcess(os.Getpid())
	// p.Signal(syscall.SIGTERM)
	<-ctx.Done()
	log.Println("[main] shutdown signal received")
	server.Shutdown()
}
