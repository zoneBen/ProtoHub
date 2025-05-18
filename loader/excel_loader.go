package loader

import (
	"ProtoHub/excel"
	"ProtoHub/modu"
)

type ExcelLoader struct {
}

func (e *ExcelLoader) Load(path string) (modu.EParser, error) {
	return excel.ReadExcelData(path)
}
