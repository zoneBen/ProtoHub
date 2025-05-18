package main

import (
	"ProtoHub/excel"
	"ProtoHub/loader"
	"ProtoHub/protocols"
	"ProtoHub/transport"
	"fmt"
	"log"
	"strings"
	"time"
)

func main() {
	a := loader.ExcelLoader{}
	dev, err := a.Load("a.xlsx")
	if err != nil {
		log.Println(err)
	}
	err = excel.CheckEParser(&dev)
	if err != nil {
		log.Println(err)
	}

	protocol := protocols.ACProtocol{}
	cmds, err := protocol.GenerateCommands(&dev)
	if err != nil {
		log.Println(err)
	}
	fmt.Printf("共有%d条命令\n", len(cmds))
	var index = 1
	var conf = transport.TCPConfig{"127.0.0.1:8080", time.Duration(1000)}
	clent := transport.NewTCPTransport(&conf)
	clent.Connect()
	time.Sleep(1 * time.Second)
	for {
		for cmdKey, c := range cmds {
			fmt.Printf("第%d命令 : %s bytes:%02X", index, strings.Replace(string(c), "\r", "", -1), c)
			index++
			buf, err := protocol.Send(clent, c, &dev)
			if err == nil {
				fmt.Printf(" ---> 收到数据%s bytes:%02X\n", strings.Replace(string(buf), "\r", "", -1), buf)
			} else {
				fmt.Println("")
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
	fmt.Printf("名称 %s\n", dev.Dev.Name)
}
