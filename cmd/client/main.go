package main

import (
	"bufio"
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/gorilla/websocket"
	"github.com/pauldin91/wsgo/client"
)

func main() {
	// Signal handling setup
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	host := flag.String("host", "ws://localhost:4443/ws", "Server host")

	flag.Parse()

	client := client.NewWsClient(ctx, *host)
	client.Connect()

	client.OnMessageParseWsHandler(func(conn *websocket.Conn) {
		reader := bufio.NewReader(os.Stdin)
		for {
			input, _, err := reader.ReadLine()
			if err != nil {
				client.SendError(err)
				return
			}
			text := strings.TrimSpace(string(input))
			if text == "exit" {
				return
			}
			conn.WriteMessage(websocket.TextMessage, []byte(text+"\n"))

		}
	})

	<-ctx.Done()

	log.Println("[main] shutdown signal received")

	client.Close()
}
