package protocols

import (
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
		log.Println("SimpleTextProtocol Send err", err)
		return nil, err
	}
	defer transport.Close()
	err = transport.Write(sendBuf)
	if err != nil {
		return nil, err
	}
	var endFlag byte
	endFlag = 0x0D
	timeout := 1 * time.Second
	timeoutChan := time.After(timeout)
	resultChan := make(chan []byte)
	errChan := make(chan error)
	go func() {
		var received []byte
		for {
			// 读取数据
			data, err := transport.Read()
			if err != nil {
				errChan <- err
				return
			}
			received = append(received, data...)
			if len(received) > 0 && (received[len(received)-1] == endFlag || received[len(received)-1] == 0x0A) {
				resultChan <- received
				return
			}
		}
	}()

	select {
	case <-timeoutChan:
		return nil, fmt.Errorf("读取超时，超时时间：%v", timeout)
	case result := <-resultChan:
		return result, nil
	case err := <-errChan:
		return nil, err
	}
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
