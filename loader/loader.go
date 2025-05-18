package loader

import "ProtoHub/modu"

type Loader interface {
	Load(path string) (modu.EParser, error)
}
