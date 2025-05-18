package protocols

import (
	"ProtoHub/core"
	"ProtoHub/modu"
	"ProtoHub/parser"
	"fmt"
	"log"
	"time"
)

type SimpleTextProtocol struct{}

// GenerateCommands 生成命令键与内容的映射
func (p *SimpleTextProtocol) GenerateCommands(dev *modu.EParser) (map[string][]byte, error) {
	commands := make(map[string][]byte)
	for _, addr := range dev.Addrs {
		cmdKey := p.GenerateKey(dev, addr)
		cmd := fmt.Sprintf("%s%s%s", addr.CID1, addr.Command, addr.CommandExtra)
		commands[cmdKey] = []byte(cmd)
	}
	return nil, nil
}

func (p *SimpleTextProtocol) GenerateKey(dev *modu.EParser, addr modu.EAddr) string {
	return fmt.Sprintf("%s_%s_%s_%s", dev.Dev.Cid1, addr.CID1, addr.Command, addr.CommandExtra)
}

// Send 根据命令键发送对应命令内容
func (p *SimpleTextProtocol) Send(transport core.Transport, sendBuf []byte, dev *modu.EParser) ([]byte, error) {
	err := transport.Connect()
	if err != nil {
		return nil, err
	}
	defer transport.Close()
	err = transport.Write(sendBuf)
	if err != nil {
		return nil, err
	}
	time.Sleep(100 * time.Millisecond)
	return transport.Read()
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
