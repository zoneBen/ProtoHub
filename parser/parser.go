package parser

import "ProtoHub/modu"

type Parser interface {
	Extract(data []byte, dev *modu.EParser, addr modu.EAddr) ([]byte, error)
	Parse(buf []byte, dev *modu.EParser, addr modu.EAddr) (modu.ParseValue, error)
}
