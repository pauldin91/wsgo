package client

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"testing"
)

func TestClient(t *testing.T) {
	// Signal handling setup
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	host := "ws://localhost:6443/ws" //flag.String("host", "wss://localhost:6443/ws", "Server host")

	flag.Parse()

	client := NewWsClient(ctx, host)
	client.Connect()
	client.Send("test")

	p, _ := os.FindProcess(os.Getpid())
	p.Signal(syscall.SIGTERM)

	<-ctx.Done()

	log.Println("[main] shutdown signal received")

	client.Close()

}
