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

	host := flag.String("host", ":4443", "Server host")
	flag.Parse()

	log.Printf("[main] starting TCP server on %s", *host)
	server := server.NewTcpServer(ctx, *host)
	server.Start()

	// p, _ := os.FindProcess(os.Getpid())
	// p.Signal(syscall.SIGTERM)
	<-ctx.Done()
	log.Println("[main] shutdown signal received")
	server.Shutdown()
}
