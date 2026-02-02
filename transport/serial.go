package transport

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"

	serial "github.com/albenik/go-serial/v2"
)

// SerialConfig 串口配置
type SerialConfig struct {
	PortName     string
	BaudRate     int
	DataBits     int    // 5, 6, 7, 8
	StopBits     string // "1", "1.5", "2"
	Parity       string // "N", "E", "O"
	ReadTimeout  int    // 毫秒
	WriteTimeout int    // 毫秒
}

// SerialTransport 串口传输层
type SerialTransport struct {
	config *SerialConfig
	port   io.ReadWriteCloser
	mu     sync.RWMutex
	closed bool
}

var (
	ErrNotConnected = errors.New("serial port not connected")
	ErrClosed       = errors.New("serial port already closed")
)

// NewSerialTransport 创建实例（带默认值）
func NewSerialTransport(config *SerialConfig) *SerialTransport {
	if config.DataBits == 0 {
		config.DataBits = 8
	}
	if config.StopBits == "" {
		config.StopBits = "1"
	}
	if config.Parity == "" {
		config.Parity = "N"
	}
	//if config.ReadTimeout == 0 {
	//	config.ReadTimeout = 50
	//}
	config.ReadTimeout = 200
	if config.WriteTimeout == 0 {
		config.WriteTimeout = 1000
	}
	return &SerialTransport{config: config}
}

// Connect 打开串口并配置参数
func (s *SerialTransport) Connect() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.port != nil {
		return errors.New("already connected")
	}
	// Step 1: Open with default settings
	port, err := serial.Open(s.config.PortName)
	if err != nil {
		return fmt.Errorf("failed to open serial port %s: %w", s.config.PortName, err)
	}

	// Step 2: Reconfigure with desired settings
	opts := []serial.Option{
		serial.WithBaudrate(s.config.BaudRate),
		serial.WithDataBits(s.config.DataBits),
		serial.WithReadTimeout(s.config.ReadTimeout),
		serial.WithWriteTimeout(s.config.WriteTimeout),
	}

	// Map StopBits
	switch s.config.StopBits {
	case "1":
		opts = append(opts, serial.WithStopBits(serial.OneStopBit))
	case "1.5":
		opts = append(opts, serial.WithStopBits(serial.OnePointFiveStopBits))
	case "2":
		opts = append(opts, serial.WithStopBits(serial.TwoStopBits))
	default:
		port.Close()
		return fmt.Errorf("invalid stop bits: %s", s.config.StopBits)
	}

	// Map Parity
	switch s.config.Parity {
	case "N":
		opts = append(opts, serial.WithParity(serial.NoParity))
	case "E":
		opts = append(opts, serial.WithParity(serial.EvenParity))
	case "O":
		opts = append(opts, serial.WithParity(serial.OddParity))
	default:
		port.Close()
		return fmt.Errorf("invalid parity: %s", s.config.Parity)
	}

	if err := port.Reconfigure(opts...); err != nil {
		port.Close()
		return fmt.Errorf("failed to reconfigure serial port: %w", err)
	}

	s.port = port
	s.closed = false
	return nil
}

// Write 写入数据
func (s *SerialTransport) Write(data []byte) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.port == nil {
		return ErrNotConnected
	}
	if s.closed {
		return ErrClosed
	}

	_, err := s.port.Write(data)
	if err != nil {
		return fmt.Errorf("serial write failed: %w", err)
	}
	return nil
}

// Read 读取数据（基础版）
func (s *SerialTransport) Read() ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.port == nil {
		return nil, ErrNotConnected
	}
	if s.closed {
		return nil, ErrClosed
	}

	totalBuf := make([]byte, 0)
	buf := make([]byte, 1024)

	// 第一次读取
	n, err := s.port.Read(buf)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil, nil
		}
		return nil, fmt.Errorf("serial read failed: %w", err)
	}
	totalBuf = append(totalBuf, buf[:n]...)

	// 继续尝试读取更多数据
	for {
		n, err := s.port.Read(buf)
		if err != nil || n == 0 {
			break
		}
		totalBuf = append(totalBuf, buf[:n]...)
	}

	return totalBuf, nil
}

// ReadWithContext 支持 context 超时
func (s *SerialTransport) ReadWithContext(ctx context.Context) ([]byte, error) {
	s.mu.RLock()
	port := s.port
	s.mu.RUnlock()

	if port == nil {
		return nil, ErrNotConnected
	}
	if s.isClosed() {
		return nil, ErrClosed
	}

	type result struct {
		data []byte
		err  error
	}
	ch := make(chan result, 1)

	go func() {
		totalBuf := make([]byte, 0)
		buf := make([]byte, 1024)

		// 第一次读取
		n, err := port.Read(buf)
		if err != nil {
			ch <- result{nil, err}
			return
		}
		totalBuf = append(totalBuf, buf[:n]...)

		// 继续尝试读取更多数据，直到超时或无数据
		for {
			n, err := port.Read(buf)
			if err != nil || n == 0 {
				break
			}
			totalBuf = append(totalBuf, buf[:n]...)
		}

		ch <- result{totalBuf, nil}
	}()

	select {
	case r := <-ch:
		return r.data, r.err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (s *SerialTransport) isClosed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.closed
}

// Close 安全关闭
func (s *SerialTransport) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil
	}
	s.closed = true

	if s.port == nil {
		return nil
	}

	err := s.port.Close()
	s.port = nil
	return err
}

// IsConnected 检查连接状态
func (s *SerialTransport) IsConnected() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.port != nil && !s.closed
}
