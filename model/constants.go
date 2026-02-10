package model

type Protocol int

const (
	TCP Protocol = iota
	WebSocket
	QUIC
	WebRTC
)
