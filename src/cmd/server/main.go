package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/pauldin91/wsgo/src/server"
)

func main() {
	// Signal handling setup
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	host := flag.String("host", "localhost:6443", "Server host")
	flag.Parse()

	server := server.NewWsServer(ctx, *host)

	server.Start()
	<-ctx.Done()
	log.Println("[main] shutdown signal received")
	server.Shutdown()
}
