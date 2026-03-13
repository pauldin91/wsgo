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
			fmt.Printf("Received: %s\n", string(msg))
			return
		}
		conns := server.GetConnections()
		conns[message.Receiver].Write(msg)
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
			conns := server.GetConnections()
			cid := make([]string, 0, len(conns))
			for _, c := range conns {
				cid = append(cid, c.RemoteAddr().String())
			}

			log.Printf("Broadcasted msg %s to clients %v\n", input, cid)
		}
	}()

	server.Start(ctx)
	<-ctx.Done()
	slog.Info("[main] shutdown signal received")
	server.Shutdown()
}
