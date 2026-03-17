package wsgo

import (
	"github.com/pauldin91/wsgo/client"
	"github.com/pauldin91/wsgo/p2p"
	"github.com/pauldin91/wsgo/server"
)

type Server = server.Server
type TcpServer = server.TCPServer
type WsServer = server.WSServer
type QuicServer = server.QUICServer
type Client = client.Client
type TcpClient = client.TcpClient
type WsClient = client.WsClient
type QuicClient = client.QuicClient
type Peer = p2p.P2PServer
