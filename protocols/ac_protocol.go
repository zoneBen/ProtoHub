package protocols

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/zoneBen/ProtoHub/core"
	"github.com/zoneBen/ProtoHub/modu"
	"github.com/zoneBen/ProtoHub/parser"
	"log"
	"time"
)

type ACProtocol struct {
	SOI byte
	EOI byte
}

// GenerateCommands 生成命令键与内容的映射
func (p *ACProtocol) GenerateCommands(dev *modu.EParser) (map[string][]byte, error) {
	commands := make(map[string][]byte)
	devCid1, err := getByte(dev.Dev.Cid1)
	if err != nil {
		log.Println("设备CID1未设置或设置错误")
		return nil, err
	}
	devVer, err := getByte(dev.Dev.Version)
	if err != nil {
		log.Println("设备版本号未设置或设置错误")
		return nil, err
	}
	devAdr, err := getByte(dev.Dev.Addr)
	if err != nil {
		log.Println("设备通讯地址未设置或设置错误")
		return nil, err
	}
	for _, addr := range dev.Addrs {
		cmdKey := p.GenerateKey(dev, addr)
		if _, ok := commands[cmdKey]; ok {
			continue
		}
		cid1, err := getByte(addr.CID1)
		if err != nil {
			cid1 = devCid1
		}
		cid2, err := getByte(addr.Command)
		if err != nil {
			cid1 = devCid1
		}
		ext, err := getBytes(addr.CommandExtra)
		frame, err := buildFrame(p.SOI, devVer, devAdr, cid1, cid2, ext, p.EOI)
		if err != nil {
			return nil, err
		}
		commands[cmdKey] = frame
	}
	return commands, nil
}

func (p *ACProtocol) GenerateKey(dev *modu.EParser, addr modu.EAddr) string {
	return fmt.Sprintf("%s@%s@%s@%s", dev.Dev.Cid1, addr.CID1, addr.Command, addr.CommandExtra)
}

func getByte(s string) (byte, error) {
	if s == "" {
		return 0, errors.New(fmt.Sprintf("转换HEX数据为空"))
	}
	if len(s) == 1 {
		s = "0" + s
	}
	if len(s) > 2 {
		return 0, errors.New(fmt.Sprintf("转换HEX数据过大请检查数据%s\n", s))
	}
	data, err := hex.DecodeString(s)
	if err != nil {
		return 0, errors.New(fmt.Sprintf("Error decoding hex string:%w", err))
	}
	return data[0], nil
}

func getBytes(s string) ([]byte, error) {
	data, err := hex.DecodeString(s)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error decoding hex string:%w", err))
	}
	return data, nil
}

// buildFrame 构建协议帧
func buildFrame(soi byte, ver byte, adr byte, cid1 byte, cid2 byte, info []byte, eoi byte) ([]byte, error) {
	lenID := uint16(len(info) * 2)
	seg1 := (lenID >> 8) & 0x0F
	seg2 := (lenID >> 4) & 0x0F
	seg3 := lenID & 0x0F
	lCheckSum := ^(seg1+seg2+seg3)&0x0F + 1
	length := (uint16(lCheckSum) << 12) | lenID
	var sum uint16

	// 组装完整帧
	frame := []byte{soi}
	body := []byte{}
	body = append(body, bytesToASCII([]byte{ver})...)
	body = append(body, bytesToASCII([]byte{adr})...)
	body = append(body, bytesToASCII([]byte{cid1})...)
	body = append(body, bytesToASCII([]byte{cid2})...)
	lenData := Uint16ToBytes(length, false)
	body = append(body, bytesToASCII(lenData)...)
	if lenID > 0 {
		body = append(body, bytesToASCII(info)...)
	}
	for _, b := range body {
		sum += uint16(b)
	}
	chkSum := ^sum + 1 // 取反加1
	frame = append(frame, body...)
	sunData := bytesToASCII(Uint16ToBytes(chkSum, false))
	frame = append(frame, sunData...)
	frame = append(frame, eoi)
	return frame, nil
}

func bytesToASCII(buf []byte) []byte {
	return []byte(fmt.Sprintf("%02X", buf))
}

func Uint16ToBytes(n uint16, littleEndian bool) []byte {
	data := make([]byte, 2)
	if littleEndian {
		binary.LittleEndian.PutUint16(data, n)
	} else {
		binary.BigEndian.PutUint16(data, n)
	}
	return data
}

// Send 根据命令键发送对应命令内容
func (p *ACProtocol) Send(transport core.Transport, sendBuf []byte, dev *modu.EParser) ([]byte, error) {
	err := transport.Connect()
	if err != nil {
		log.Println("ACProtocol Send connect err:", err)
		return nil, err
	}
	defer transport.Close()

	err = transport.Write(sendBuf)
	if err != nil {
		return nil, fmt.Errorf("write failed: %w", err)
	}

	// ⬇️ 关键：总超时设为 3 秒（足够设备响应 + 发送完整帧）
	totalTimeout := 3 * time.Second
	endTime := time.Now().Add(totalTimeout)
	var received []byte

	for time.Now().Before(endTime) {
		// 每次读取最多等待 200ms（应 >= 串口 ReadTimeout）
		remaining := time.Until(endTime)
		if remaining <= 0 {
			break
		}
		readTimeout := 200 * time.Millisecond
		if remaining < readTimeout {
			readTimeout = remaining
		}

		ctx, cancel := context.WithTimeout(context.Background(), readTimeout)
		data, err := transport.ReadWithContext(ctx)
		cancel()

		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				// 本次读取超时，但总时间还没到，继续等
				continue
			}
			return nil, fmt.Errorf("read error: %w", err)
		}

		if len(data) > 0 {
			received = append(received, data...)
			// 🔍 严格检查：只要 buffer 中最后一个字节是 EOI，就认为帧完整
			if received[len(received)-1] == p.EOI {
				return received, nil // ✅ 完整帧，立即返回
			}
		}
	}

	// 总超时到了，但没等到 EOI
	if len(received) > 0 {
		log.Printf("Warning: response missing EOI (0x%02X). Got: % X", p.EOI, received)
	}
	return nil, fmt.Errorf("timeout waiting for EOI (0x%02X) after %v", p.EOI, totalTimeout)
}

// GetCommandAddrs 获取命令对应的测点
func (p *ACProtocol) GetCommandAddrs(dev *modu.EParser, commandKey string) (addrs []modu.EAddr) {
	for _, addr := range dev.Addrs {
		cmdKey := p.GenerateKey(dev, addr)
		if cmdKey == commandKey {
			addrs = append(addrs, addr)
		}
	}
	return addrs
}

// ParseResponse 解析响应数据
func (p *ACProtocol) ParseResponse(data []byte, dev *modu.EParser, addrs []modu.EAddr) (map[string]modu.ParseValue, error) {
	var par parser.HexParser
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
