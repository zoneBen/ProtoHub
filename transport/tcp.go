package transport

import (
	"context"
	"errors"
	"net"
	"time"
)

type TCPConfig struct {
	Address string
	Timeout time.Duration
}

type TCPTransport struct {
	config *TCPConfig
	conn   net.Conn
}

func NewTCPTransport(config *TCPConfig) *TCPTransport {
	return &TCPTransport{config: config}
}

func (t *TCPTransport) Connect() error {
	//if t.conn != nil {
	//	return errors.New("tcp connection already established")
	//}
	conn, err := net.DialTimeout("tcp", t.config.Address, t.config.Timeout)
	if err != nil {
		return err
	}

	t.conn = conn
	return nil
}

func (t *TCPTransport) Write(data []byte) error {
	if t.conn == nil {
		return errors.New("tcp connection not established")
	}
	_, err := t.conn.Write(data)
	return err
}

func (t *TCPTransport) Read() ([]byte, error) {
	if t.conn == nil {
		return nil, errors.New("tcp connection not established")
	}
	// 设置读取超时时间
	if err := t.conn.SetReadDeadline(time.Now().Add(t.config.Timeout)); err != nil {
		return nil, err
	}
	buf := make([]byte, 1024)
	n, err := t.conn.Read(buf)
	if err != nil {
		// 处理超时错误
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return nil, errors.New("read timeout")
		}
		return nil, err
	}
	return buf[:n], nil
}

func (t *TCPTransport) ReadWithContext(ctx context.Context) ([]byte, error) {
	if t.conn == nil {
		return nil, errors.New("tcp connection not established")
	}

	// 获取 context 的 deadline
	deadline, ok := ctx.Deadline()
	if ok {
		// 设置读取截止时间为 context 的 deadline
		if err := t.conn.SetReadDeadline(deadline); err != nil {
			return nil, err
		}
	} else {
		// 如果 context 没有 deadline，使用默认超时（避免无限阻塞）
		if err := t.conn.SetReadDeadline(time.Now().Add(t.config.Timeout)); err != nil {
			return nil, err
		}
	}

	buf := make([]byte, 1024)
	n, err := t.conn.Read(buf)
	if err != nil {
		// 检查是否是超时错误
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			// 判断是 context 超时还是默认超时
			select {
			case <-ctx.Done():
				return nil, ctx.Err() // 被 context 取消
			default:
				return nil, errors.New("read timeout")
			}
		}
		return nil, err
	}

	return buf[:n], nil
}

func (t *TCPTransport) Close() error {
	if t.conn == nil {
		return nil
	}
	err := t.conn.Close()
	t.conn = nil
	return err
}
