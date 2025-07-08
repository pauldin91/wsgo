package internal

type Conn interface {
	Write(p []byte) (int, error)
	Close() error
}
