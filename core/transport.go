package core

type Transport interface {
	Connect() error
	Write(data []byte) error
	Read() ([]byte, error)
	Close() error
}
