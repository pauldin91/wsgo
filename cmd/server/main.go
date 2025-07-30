package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/pauldin91/wsgo/server"
)

func main() {
	// Signal handling setup
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	host := flag.String("host", ":4443", "Server host")
	flag.Parse()

	log.Printf("[main] starting TCP server on %s", *host)
	server := server.NewTcpServer(ctx, *host)
	server.Start()
	server.OnMessageReceived(func(msg []byte) {

		fmt.Printf("Received: %s\n", string(msg))
	})
	<-ctx.Done()
	log.Println("[main] shutdown signal received")
	server.Shutdown()
}
