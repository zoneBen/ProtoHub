package loader

import "github.com/zoneBen/ProtoHub/modu"

type Loader interface {
	Load(path string) (modu.EParser, error)
}
