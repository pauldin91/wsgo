package wsgo

import (
	"github.com/pauldin91/wsgo/client"
	"github.com/pauldin91/wsgo/server"
)

type Server = server.Server
type TcpServer = server.TcpServer
type WsServer = server.WsServer
type QuicServer = server.QuicServer
type Client = client.Client
type TcpClient = client.TcpClient
type WsClient = client.WsClient
type QuicClient = client.QuicClient
