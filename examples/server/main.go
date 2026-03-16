package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/pauldin91/wsgo/protocol"
	"github.com/pauldin91/wsgo/server"
)

func main() {
	// Signal handling setup
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	host := flag.String("host", ":8080", "Server host")
	flag.Parse()

	server := server.NewTcpServer(*host)

	server.OnMessageReceived(func(msg []byte) {

		message := protocol.Message{}
		err := json.Unmarshal(msg, &message)
		if err != nil {
			fmt.Printf("Failed to parse message %s with error: %v\n", string(msg), err)
			return
		}
		server.SendTo(message)
	})

	reader := bufio.NewReader(os.Stdin)
	go func() {
		for {
			input, _, err := reader.ReadLine()
			if err != nil {
				log.Printf("error: %v", err)
				return
			}
			if string(input) == "exit" {
				stop()
				return
			}
			if err := server.Broadcast(input); err != nil {
				log.Printf("Broadcast error: %v", err)
			}
			log.Printf("Broadcasted msg %s to all clients\n", input)
		}
	}()

	server.Start(ctx)
	<-ctx.Done()
	slog.Info("[main] shutdown signal received")
	server.Shutdown()
}
