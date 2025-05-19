package transport

import (
	"errors"
	"time"

	"github.com/tarm/serial"
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
	port   *serial.Port
}

func NewSerialTransport(config *SerialConfig) *SerialTransport {
	return &SerialTransport{config: config}
}

func (s *SerialTransport) Connect() error {
	c := &serial.Config{
		Name:        s.config.PortName,
		Baud:        s.config.BaudRate,
		Size:        byte(s.config.DataBits),
		StopBits:    serial.StopBits(s.config.StopBits),
		Parity:      serial.Parity(s.config.Parity),
		ReadTimeout: time.Duration(s.config.ReadTimeout),
	}
	port, err := serial.OpenPort(c)
	if err != nil {
		return err
	}
	s.port = port
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
