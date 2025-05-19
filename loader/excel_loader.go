package loader

import (
	"github.com/zoneBen/ProtoHub/excel"
	"github.com/zoneBen/ProtoHub/modu"
)

type ExcelLoader struct {
}

func (e *ExcelLoader) Load(path string) (modu.EParser, error) {
	return excel.ReadExcelData(path)
}
