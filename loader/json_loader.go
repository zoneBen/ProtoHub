package loader

import (
	"encoding/json"
	"errors"
	"github.com/zoneBen/ProtoHub/modu"
	"io"
	"os"
)

type JsonLoader struct {
}

func (j *JsonLoader) Load(filePath string) (modu.EParser, error) {
	var dev modu.EParser
	f, err := os.Open(filePath)
	defer f.Close()
	if err != nil {
		err = errors.New("Load file error")
		return dev, err
	}
	data, err := io.ReadAll(f)
	err = json.Unmarshal(data, &dev)
	return dev, err
}
