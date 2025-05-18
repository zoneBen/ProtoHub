package loader

import "ProtoHub/modu"

type JsonLoader struct {
}

func (j *JsonLoader) Load(path string) (modu.EParser, error) {
	return modu.EParser{}, nil
}
