package server

import (
	"net"
)

type Server interface {
	Start()
	StartTls(certFile, certKey string)
	GetConnections() map[string]net.Conn
	Shutdown()
}
