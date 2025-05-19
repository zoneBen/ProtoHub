package main

import (
	"ProtoHub/core"
	"ProtoHub/excel"
	"ProtoHub/loader"
	"ProtoHub/protocols"
	"ProtoHub/transport"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"
)

func main() {
	var controller string
	var filePath string
	var baudRate int64
	var dataBits int64
	var stopBits int64
	var parity string
	flag.Int64Var(&dataBits, "db", 8, "数据位")
	flag.Int64Var(&stopBits, "sb", 1, "停止位")
	flag.Int64Var(&baudRate, "b", 9600, "默认波特率")
	flag.StringVar(&parity, "p", "N", "N O E")
	flag.StringVar(&filePath, "f", "a.xlsx", "IP地址:端口")
	flag.StringVar(&controller, "c", "", "IP地址:端口")
	flag.Parse()
	var load loader.Loader
	if strings.Contains(filePath, ".json") {
		load = &loader.JsonLoader{}
	} else {
		load = &loader.ExcelLoader{}
	}
	dev, err := load.Load(filePath)
	if err != nil {
		log.Println(err)
	}
	err = excel.CheckEParser(&dev)
	if err != nil {
		log.Println(err)
	}
	var protocol core.Protocol
	if dev.Dev.TransmissionMode == "电总" {
		protocol = &protocols.ACProtocol{SOI: 0x7E, EOI: 0x0D}
	} else {
		protocol = &protocols.SimpleTextProtocol{}
	}
	cmds, err := protocol.GenerateCommands(&dev)
	if err != nil {
		log.Println(err)
	}
	fmt.Printf("共有%d条命令\n", len(cmds))
	var clent core.Transport
	var index = 1
	if strings.HasPrefix(controller, "/dev") {
		var conf = transport.SerialConfig{PortName: controller,
			BaudRate:    int(baudRate),
			DataBits:    int(dataBits),
			StopBits:    int(stopBits),
			Parity:      parity[0],
			ReadTimeout: 1000}
		clent = transport.NewSerialTransport(&conf)
	} else {
		var conf = transport.TCPConfig{controller, time.Duration(1000) * time.Millisecond}
		clent = transport.NewTCPTransport(&conf)
	}

	clent.Connect()
	time.Sleep(1 * time.Second)
	for {
		for cmdKey, c := range cmds {
			fmt.Printf("第%d命令 : %s bytes:%02X\n", index, strings.Replace(string(c), "\r", "", -1), c)
			index++
			buf, err := protocol.Send(clent, c, &dev)
			if err == nil {
				fmt.Printf(" ---> 收到数据%s bytes:%02X\n", strings.Replace(string(buf), "\r", "", -1), buf)
			} else {
				fmt.Println(err)
			}
			if len(buf) > 0 {
				addrs := protocol.GetCommandAddrs(&dev, cmdKey)
				rs, _ := protocol.ParseResponse(buf, &dev, addrs)
				for _, r := range rs {
					fmt.Printf("%s -> %g %s\n", r.Addr.MetricName, r.Value, r.Addr.EnumStr)
				}
			}
			time.Sleep(1 * time.Second)
		}
	}
}
