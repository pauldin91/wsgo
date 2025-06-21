package client

import (
	"os"
)

type Client interface {
	Close()
	Connect()
	GetConnId() string
	HandleInputFrom(source *os.File, handler func(*os.File))
	Send(message string)
	SendError(err error)
}
