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
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	host := flag.String("host", ":4443", "Server listen address")
	proto := flag.String("protocol", "tcp", "Protocol to use: tcp, websocket, quic, webrtc")
	flag.Parse()

	srv, err := server.NewServer(*host, *proto)
	if err != nil {
		slog.Error("failed to create server", "error", err)
		os.Exit(1)
	}

	srv.OnMessageReceived(func(msg []byte) {
		fmt.Printf("Received: %s\n", string(msg))
	})

	srv.Start(ctx)
	slog.Info("server started", "protocol", *proto, "address", *host)

	<-ctx.Done()
	slog.Info("shutdown signal received")
	srv.Shutdown()
}
