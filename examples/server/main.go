package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
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

	server := server.NewTcpServer(*host)
	server.Start(ctx)

	server.OnMessageReceived(func(msg []byte) {
		fmt.Printf("Received: %s\n", string(msg))
	})
	<-ctx.Done()
	slog.Info("[main] shutdown signal received")
	server.Shutdown()
}
