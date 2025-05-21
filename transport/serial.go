package transport

import (
	"errors"
	serial "github.com/tarm/goserial"
	"io"
	"time"
)

type SerialConfig struct {
	PortName    string
	BaudRate    int
	DataBits    int
	StopBits    int
	Parity      byte
	ReadTimeout int
}

type SerialTransport struct {
	config *SerialConfig
	port   io.ReadWriteCloser
}

func NewSerialTransport(config *SerialConfig) *SerialTransport {
	return &SerialTransport{config: config}
}

func (s *SerialTransport) Connect() error {
	c := &serial.Config{
		Name:        s.config.PortName,
		Baud:        s.config.BaudRate,
		ReadTimeout: time.Duration(s.config.ReadTimeout),
	}
	var err error
	s.port, err = serial.OpenPort(c)
	if err != nil {
		return err
	}
	return nil
}

func (s *SerialTransport) Write(data []byte) error {
	if s.port == nil {
		return errors.New("serial port not connected")
	}
	_, err := s.port.Write(data)
	return err
}

func (s *SerialTransport) Read() ([]byte, error) {
	if s.port == nil {
		return nil, errors.New("serial port not connected")
	}
	buf := make([]byte, 1024)
	n, err := s.port.Read(buf)
	if err != nil {
		return nil, err
	}
	return buf[:n], nil
}

func (s *SerialTransport) Close() error {
	if s.port == nil {
		return nil
	}
	err := s.port.Close()
	s.port = nil
	return err
}
