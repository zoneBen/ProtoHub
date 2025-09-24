package protocols

import (
	"context"
	"errors"
	"fmt"
	"github.com/zoneBen/ProtoHub/core"
	"github.com/zoneBen/ProtoHub/modu"
	"github.com/zoneBen/ProtoHub/parser"
	"log"
	"strings"
	"time"
)

type SimpleTextProtocol struct{}

func replacementSpecialCharacters(oldVal string) (val string) {
	val = strings.Replace(oldVal, "空格", " ", -1)
	val = strings.Replace(oldVal, "\\r", "\r", -1)
	val = strings.Replace(val, "\\n", "\n", -1)
	return
}

// GenerateCommands 生成命令键与内容的映射
func (p *SimpleTextProtocol) GenerateCommands(dev *modu.EParser) (map[string][]byte, error) {
	commands := make(map[string][]byte)
	var sendPre = replacementSpecialCharacters(dev.Dev.SendPre)
	var sendSuf = replacementSpecialCharacters(dev.Dev.SendSuf)
	for _, addr := range dev.Addrs {
		if addr.SendPre != "" {
			sendPre = replacementSpecialCharacters(addr.SendPre)
		}
		if addr.SendSuf != "" {
			sendSuf = replacementSpecialCharacters(addr.SendSuf)
		}
		cmdKey := p.GenerateKey(dev, addr)
		cmd := fmt.Sprintf("%s%s%s%s%s", sendPre, addr.CID1, addr.Command, addr.CommandExtra, sendSuf)
		commands[cmdKey] = []byte(cmd)
	}
	return commands, nil
}

func (p *SimpleTextProtocol) GenerateKey(dev *modu.EParser, addr modu.EAddr) string {
	return fmt.Sprintf("%s_%s_%s_%s", dev.Dev.Cid1, addr.CID1, addr.Command, addr.CommandExtra)
}

// Send 根据命令键发送对应命令内容
func (p *SimpleTextProtocol) Send(transport core.Transport, sendBuf []byte, dev *modu.EParser) ([]byte, error) {
	err := transport.Connect()
	if err != nil {
		log.Println("SimpleTextProtocol Send connect err:", err)
		return nil, err
	}
	defer func() {
		if closeErr := transport.Close(); closeErr != nil {
			log.Printf("Warning: failed to close transport: %v", closeErr)
		}
	}()

	err = transport.Write(sendBuf)
	if err != nil {
		return nil, fmt.Errorf("write failed: %w", err)
	}

	var endFlag1, endFlag2 byte = 0x0D, 0x0A // \r and \n
	timeout := 1 * time.Second
	endTime := time.Now().Add(timeout)
	var received []byte

	for time.Now().Before(endTime) {
		remaining := time.Until(endTime)
		if remaining <= 0 {
			break
		}

		// 设置本次读取的最大等待时间（例如 100ms，避免单次阻塞太久）
		readTimeout := 100 * time.Millisecond
		if remaining < readTimeout {
			readTimeout = remaining
		}

		ctx, cancel := context.WithTimeout(context.Background(), readTimeout)
		data, err := transport.ReadWithContext(ctx)
		cancel()

		if err != nil {
			// 如果是 context 超时，继续下一轮
			if errors.Is(err, context.DeadlineExceeded) {
				continue
			}
			// 其他错误（如连接断开）
			return nil, fmt.Errorf("read error: %w", err)
		}

		if len(data) > 0 {
			received = append(received, data...)
			// 检查最后一个字节是否为结束符
			last := received[len(received)-1]
			if last == endFlag1 || last == endFlag2 {
				return received, nil
			}
		}
	}

	return nil, fmt.Errorf("read timeout after %v", timeout)
}

// GetCommandAddrs 获取命令对应的测点
func (p *SimpleTextProtocol) GetCommandAddrs(dev *modu.EParser, commandKey string) (addrs []modu.EAddr) {
	for _, addr := range dev.Addrs {
		cmdKey := p.GenerateKey(dev, addr)
		if cmdKey == commandKey {
			addrs = append(addrs, addr)
		}
	}
	return addrs
}

// ParseResponse 解析响应数据
func (p *SimpleTextProtocol) ParseResponse(data []byte, dev *modu.EParser, addrs []modu.EAddr) (map[string]modu.ParseValue, error) {
	var par parser.SimpleParser
	var r = make(map[string]modu.ParseValue)
	for _, addr := range addrs {
		extract, err := par.Extract(data, dev, addr)
		if err != nil {
			log.Printf("提取数据失败")
			continue
		}
		v, err := par.Parse(extract, dev, addr)
		r[addr.MetricCode] = v
	}
	return r, nil
}
