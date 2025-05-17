package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/pauldin91/wsgo/src/core"
)

func main() {
	// Signal handling setup
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	host := flag.String("host", "wss://localhost:6443/ws", "Server host")

	flag.Parse()

	client := core.NewWsClient(*host)
	go func() {
		client.Wait(sigChan)
	}()
	readFromConsole(sigChan, client)

}

func readFromConsole(sigChan chan os.Signal, client *core.WsClient) {

	// Input reader setup
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Type something (Ctrl+C to exit):")

	for {
		select {
		case sig := <-sigChan:
			fmt.Printf("\nReceived signal: %s. Exiting...\n", sig)
			return

		default:
			fmt.Print("> ")
			input, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println("Read error:", err)
				return
			}

			text := strings.TrimSpace(input)
			if text == "exit" {
				fmt.Println("Exiting by user command.")
				return
			}

			fmt.Printf("You send: %s\n", text)
			client.Send(text)
		}
	}
}
