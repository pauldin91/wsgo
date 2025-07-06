package p2p

type p2p interface {
	Start()
	StartTls()
	Connect(...string)
	Shutdown()
}
