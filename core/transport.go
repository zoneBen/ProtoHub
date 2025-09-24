package core

import "context"

type Transport interface {
	Connect() error
	Write(data []byte) error
	Read() ([]byte, error)
	ReadWithContext(ctx context.Context) ([]byte, error)
	Close() error
}
