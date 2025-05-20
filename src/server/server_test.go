package server

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"testing"
)

func TestServer(t *testing.T) {
	// Signal handling setup
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	host := "localhost:6443" //flag.String("host", "wss://localhost:6443/ws", "Server host")

	server := NewWsServer(ctx, host)

	server.Start()
	<-ctx.Done()
	log.Println("[main] shutdown signal received")
	server.Shutdown()
}
