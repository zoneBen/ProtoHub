package parser

import (
	"ProtoHub/modu"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
)

type SimpleParser struct {
}

func (p *SimpleParser) Extract(data []byte, dev *modu.EParser, addr modu.EAddr) ([]byte, error) {

	separator := dev.Dev.Separator
	if separator == "空格" {
		separator = " "
	}
	strData := strings.Replace(string(data), "\r", "", -1)
	tmps := strings.Split(strData, separator)
	maxlen := len(tmps)
	var valTmp string
	//优先根据测点Index
	if addr.MetricIndex > 0 && addr.MetricIndex <= maxlen {
		valTmp = tmps[addr.MetricIndex-1]
	} else {
		if addr.Length > 0 {
			valTmp = string(data)[addr.StartAt : addr.StartAt+addr.Length]
		}
	}
	// 第一个测点需要排除前缀
	if addr.MetricIndex == 1 {
		_, _, _, revPre, _ := getQRealCommand(dev, addr)
		if strings.HasPrefix(valTmp, revPre) {
			valTmp = valTmp[len(revPre):]
		}
	}
	//再截取
	if addr.CutLength > 0 {
		// 截取反转替换为正转
		if len(valTmp) >= (addr.CutOffset + addr.CutLength) {
			valTmp = valTmp[addr.CutOffset : addr.CutOffset+addr.CutLength]
		}
	}

	return []byte(valTmp), nil
}

func (p *SimpleParser) Parse(buf []byte, dev *modu.EParser, addr modu.EAddr) (modu.ParseValue, error) {
	var parseValue modu.ParseValue
	parseValue.Addr = addr
	valTmp := string(buf)
	var err error
	var v float64
	if valTmp != "" {
		var valFloat float64
		switch addr.DataType {
		case "FLOAT":
			{
				// 浮点数提取强化兼容性对于包含字符的自动过滤处理
				valFloat, err = extractNumber(valTmp)
				if err != nil {
					log.Println("coverValue() FLOAT ParseFloat error ", err.Error())
				}
				break
			}
		case "MAP":
			{
				var jMap map[string]float64
				if addr.ReMap != "" {
					err = json.Unmarshal([]byte(addr.ReMap), &jMap)
					if err == nil {
						if vTmp, ok := jMap[valTmp]; ok {
							valFloat = vTmp
						}
					} else {
						log.Println("coverValue() MAP json.Unmarshal error ", err.Error())
					}
				}
				break
			}
		case "BIN2INT":
			{
				num, err := strconv.ParseInt(valTmp, 2, 64)
				if err == nil {
					valFloat = float64(num)
				} else {
					log.Println("coverValue() BIN2INT ParseInt error ", valTmp, err.Error())
				}
				break
			}
		case "HEX2INT":
			{
				num, err := strconv.ParseInt(valTmp, 16, 64)
				if err == nil {
					valFloat = float64(num)
				} else {
					log.Println("coverValue() HEX2INT ParseInt error ", valTmp, err.Error())
				}
				break
			}
		}
		if addr.Scale == 0.0 {
			addr.Scale = 1.0
		}
		if err == nil {
			v = (valFloat * addr.Scale) + addr.Foundation
			parseValue.Value = v
		}
	} else {
		err = errors.New(fmt.Sprintf("coverValue() error: %s\n", addr.MetricName))
	}
	return parseValue, err
}

func getQRealCommand(dev *modu.EParser, eAddr modu.EAddr) (realCommand string, sendPre, sendSuf, revPre, revSuf string) {
	if dev.Dev.SendPre != "" {
		sendPre = replacementSpecialCharacters(dev.Dev.SendPre)
	}
	if eAddr.SendPre != "" {
		sendPre = replacementSpecialCharacters(eAddr.SendPre)
	}

	if dev.Dev.SendSuf != "" {
		sendSuf = replacementSpecialCharacters(dev.Dev.SendSuf)
	}
	if eAddr.SendSuf != "" {
		sendSuf = replacementSpecialCharacters(eAddr.SendSuf)
	}

	if dev.Dev.RevPre != "" {
		revPre = replacementSpecialCharacters(dev.Dev.RevPre)
	}
	if eAddr.RevPre != "" {
		revPre = replacementSpecialCharacters(eAddr.RevPre)
	}

	if dev.Dev.RevSuf != "" {
		revSuf = replacementSpecialCharacters(dev.Dev.RevSuf)
	}
	if eAddr.RevSuf != "" {
		revSuf = replacementSpecialCharacters(eAddr.RevSuf)
	}

	realCommand = fmt.Sprintf("%s%s%s%s", sendPre, eAddr.Command, eAddr.CommandExtra, sendSuf)
	return
}

func replacementNRCharacters(oldVal string) (val string) {
	val = strings.Replace(oldVal, "\r", "", -1)
	val = strings.Replace(val, "\n", "", -1)
	return
}

func replacementSpecialCharacters(oldVal string) (val string) {
	val = strings.Replace(oldVal, "\\r", "\r", -1)
	val = strings.Replace(val, "\\n", "\n", -1)
	val = strings.Replace(val, "\\n", "\n", -1)
	return
}

// extractNumber 尝试从字符串中提取第一个有效的浮点数或整型数
func extractNumber(input string) (float64, error) {
	re := regexp.MustCompile(`[-+]?\d+(\.\d+)?`)
	// 查找所有匹配项
	matches := re.FindAllString(input, -1)
	if len(matches) > 0 {
		return strconv.ParseFloat(matches[0], 64)
	}
	// 如果没有找到有效的数字，则返回错误
	return 0, fmt.Errorf("no valid number found in input string")
}
