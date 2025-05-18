package parser

import (
	"ProtoHub/modu"
	"encoding/hex"
	"errors"
	"strconv"
)

type HexParser struct {
}

func (p *HexParser) Extract(data []byte, dev *modu.EParser, addr modu.EAddr) ([]byte, error) {
	if addr.Length > 0 {
		end := addr.StartAt + addr.Length
		if len(data) >= end {
			return data[addr.StartAt:end], nil
		}
	}
	return nil, errors.New("数据不足")
}

func (p *HexParser) Parse(buf []byte, dev *modu.EParser, addr modu.EAddr) (modu.ParseValue, error) {
	var parseValue modu.ParseValue
	parseValue.Addr = addr
	bytes, err := hex.DecodeString(string(buf))
	if err != nil {
		return parseValue, err
	}
	var v float64
	if addr.DataType == "BIN2INT" {
		if addr.CutLength < 1 {
			err = errors.New("BIN2INT长度不足")
		}
		tmps := BytesToBinaryString(bytes)
		st := len(tmps) - addr.CutOffset
		t := tmps[st-addr.CutLength : st]
		ti, err1 := strconv.ParseInt(string(t), 2, 64)
		if err1 != nil {
			err = err1
			return parseValue, errors.New("strconv.ParseInt 转换失败")
		}
		v = float64(ti)
	} else {
		v = convertToFloat(bytes, addr.DataType, addr.ByteOrder)
	}
	if addr.Scale != 0 {
		v = v * addr.Scale
	}
	v = v + addr.Foundation
	parseValue.Value = v
	return parseValue, nil
}
