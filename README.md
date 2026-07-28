# wsgo

[![Go Reference](https://pkg.go.dev/badge/github.com/pauldin91/wsgo.svg)](https://pkg.go.dev/github.com/pauldin91/wsgo)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

A lightweight Go library for building peer-to-peer applications. Provides a unified interface across TCP, WebSocket, QUIC, and WebRTC transports with minimal setup.

## Protocols

| Protocol  | Transport | TLS        | Use case                          |
|-----------|-----------|------------|-----------------------------------|
| TCP       | TCP       | Optional   | Low-level, reliable streams       |
| WebSocket | TCP       | Optional   | Browser-compatible, HTTP upgrade  |
| QUIC      | UDP       | Required   | Low-latency, multiplexed streams  |
| WebRTC    | UDP       | Built-in   | Browser P2P, NAT traversal        |

## Installation

```bash
go get github.com/pauldin91/wsgo
```

## Usage

### Server

```go
import "github.com/pauldin91/wsgo/server"

srv, err := server.NewServer(":4443", "tcp") // tcp | websocket | quic | webrtc
if err != nil {
    log.Fatal(err)
}

srv.OnMessageReceived(func(msg []byte) {
    fmt.Println("received:", string(msg))
})

srv.Start(ctx)
defer srv.Shutdown()
```

### Client

```go
import "github.com/pauldin91/wsgo/client"

c, err := client.NewClient(ctx, ":4443", "tcp") // tcp | websocket | quic | webrtc
if err != nil {
    log.Fatal(err)
}

c.OnMessageReceivedHandler(func(msg []byte) {
    fmt.Println("received:", string(msg))
})

if err := c.Connect(ctx); err != nil {
    log.Fatal(err)
}
defer c.Close()

c.Send([]byte("hello"))
```

### Protocol-specific constructors

```go
// TCP
server.NewTcpServer(":4443")
client.NewTcpClient(":4443")

// WebSocket
server.NewWsServerWithCerts(":6443", nil)
client.NewWsClient("ws://localhost:6443/ws")

// QUIC (self-signed TLS generated automatically)
server.NewQuicServer(":7443")
client.NewQuicClient(":7443")

// WebRTC (signaling over HTTP)
server.NewWebRTCServer(":8443")          // exposes POST /webrtc/offer
client.NewWebRTCClient("http://localhost:8443/webrtc/offer")
```

## Running the examples

Both examples accept `-host` and `-protocol` flags.

```bash
# TCP
go run ./examples/server -host=:4443 -protocol=tcp
go run ./examples/client -host=:4443 -protocol=tcp

# WebSocket
go run ./examples/server -host=:6443 -protocol=websocket
go run ./examples/client -host=ws://localhost:6443/ws -protocol=websocket

# QUIC
go run ./examples/server -host=:7443 -protocol=quic
go run ./examples/client -host=:7443 -protocol=quic

# WebRTC
go run ./examples/server -host=:8443 -protocol=webrtc
go run ./examples/client -host=http://localhost:8443/webrtc/offer -protocol=webrtc
```

Type `exit` in the client to disconnect gracefully.

## Docker

All services are built from a single unified image. Run individual protocol pairs or all at once.

```bash
# All protocols
cd docker && docker compose up

# Single pair
docker compose up tcp.server tcp.client
docker compose up ws.server ws.client
docker compose up quic.server quic.client
docker compose up webrtc.server webrtc.client
```

| Service         | Port            | Protocol  |
|-----------------|-----------------|-----------|
| `tcp.server`    | `4443/tcp`      | TCP       |
| `ws.server`     | `6443/tcp`      | WebSocket |
| `quic.server`   | `7443/udp`      | QUIC      |
| `webrtc.server` | `8443/tcp`      | WebRTC    |

## Project structure

```
wsgo/
├── client/
│   ├── client.go          # Client interface + factory
│   ├── tcp_client.go
│   ├── ws_client.go
│   ├── quic_client.go
│   └── webrtc_client.go
├── server/
│   ├── server.go          # Server interface + factory
│   ├── tcp_server.go
│   ├── ws_server.go
│   ├── quic_server.go
│   └── webrtc_server.go
├── protocol/
│   ├── protocol.go        # Protocol type + parsing
│   └── message.go         # Message struct
├── internal/
│   └── crypto/
│       └── tls.go         # Self-signed TLS for QUIC
├── examples/
│   ├── server/main.go     # Runnable server (all protocols)
│   └── client/main.go     # Runnable client (all protocols)
├── docker/
│   ├── server.dockerfile
│   ├── client.dockerfile
│   └── docker-compose.yml
└── wsgo.go                # Top-level type aliases
```

## WebRTC signaling

WebRTC requires an SDP offer/answer exchange before the data channel opens. `wsgo` handles this with a lightweight HTTP signaling endpoint built into the server — no external signaling infrastructure needed.

```
client                        server
  |                              |
  |-- POST /webrtc/offer (SDP) ->|
  |<-- 200 OK (SDP answer) ------|
  |                              |
  |<====== DataChannel =========>|
```

## QUIC notes

QUIC requires TLS. The server auto-generates a self-signed certificate via `internal/crypto`. The client connects with `InsecureSkipVerify: true`, suitable for development. For production, supply a real certificate to `crypto.GenerateTLSConfig` or replace it with your own `tls.Config`.

## License

[MIT](LICENSE)
