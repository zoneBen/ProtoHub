package core

import "github.com/zoneBen/ProtoHub/modu"

type Protocol interface {
	GenerateCommands(dev *modu.EParser) (map[string][]byte, error)
	GetCommandAddrs(dev *modu.EParser, commandKey string) []modu.EAddr
	GenerateKey(dev *modu.EParser, addr modu.EAddr) string
	Send(transport Transport, sendBuf []byte, dev *modu.EParser) ([]byte, error)
	ParseResponse(data []byte, dev *modu.EParser, addrs []modu.EAddr) (map[string]modu.ParseValue, error)
}
