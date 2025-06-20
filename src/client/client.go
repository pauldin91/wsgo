package client

import (
	"bufio"
)

type Client interface {
	Close()
	Connect()
	GetConnId() string
	Input(reader *bufio.Reader, handler func(*bufio.Reader))
	ListenForInput(reader *bufio.Reader)
	Send(message string)
}
