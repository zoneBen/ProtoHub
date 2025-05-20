package excel

import (
	"errors"
	"fmt"
	"github.com/wxnacy/wgo/arrays"
	excelize "github.com/xuri/excelize/v2"
	"github.com/zoneBen/ProtoHub/internal/pinyin"
	"github.com/zoneBen/ProtoHub/modu"
	"strconv"
	"strings"
)

const (
	AddrSheetName      = "Sheet1"
	AddrSheetStartRow  = 6
	AlarmSheetName     = "alarm"
	AlarmSheetStartRow = 2
	HmiSheetName       = "hmi"
	HmiSheetStartRow   = 2
)

var AddrTitles = []string{"命令（CID1）", "命令（CID2）", "命令内容", "发送前缀", "发送后缀", "接收前缀", "接收后缀", "指标名称", "指标单位", "测点名称", "测点序号", "起始位", "数据长度", "枚举", "截取偏移", "截取长度", "缩放", "偏置", "字节排序", "数据类型", "值映射", "非零告警", "告警描述"}
var AlarmTitles = []string{"指标名称", "操作符", "值", "告警描述"}
var HmiTitles = []string{"指标名称", "功能码", "寄存器地址", "缩放", "步长"}

var HDataType = []string{"INT8", "UINT8", "BIN2INT", "INT16", "UINT16", "UINT32", "INT64", "UINT64", "FLOAT32-IEEE", "FLOAT64-IEEE", "FLOAT32", "FIXED", "UFIXED"}
var Q1DataType = []string{"FLOAT", "MAP", "BIN2INT", "HEX2INT", "INT2BIN"}
var YesNoDropList = []string{"是", "否"}

var ByteOrderDropList = []string{"AB", "ABCD", "BA", "DCBA", "BADC", "CDAB", "ABCDEFGH", "GHEFCDAB", "BADCFEHG", "HGFEDCBA"}
var ScaleDropList = []string{"1.0", "0.1", "0.01", "0.001", "10", "100"}
var HmiScaleDropList = []string{"1", "10", "100"}
var WriteFuncDropList = []string{"无", "5", "6", "15", "16"}

var OperatorDropList = []string{">=", "<=", "!=", "==", ">", "<"}
var TransmissionModeDropList = []string{"Q1", "电总"}

func ReadExcelData(excelFile string) (ex modu.EParser, err error) {
	f, err := excelize.OpenFile(excelFile)
	defer f.Close()
	if err != nil {
		return
	}
	ex.Dev.Name, err = f.GetCellValue(AddrSheetName, "B1")
	ex.Dev.Code, err = f.GetCellValue(AddrSheetName, "B2")
	ex.Dev.DevType, err = f.GetCellValue(AddrSheetName, "B3")
	ex.Dev.RevPre, err = f.GetCellValue(AddrSheetName, "F1")
	ex.Dev.RevSuf, err = f.GetCellValue(AddrSheetName, "F2")
	ex.Dev.SendPre, err = f.GetCellValue(AddrSheetName, "F3")
	ex.Dev.SendSuf, err = f.GetCellValue(AddrSheetName, "J1")
	ex.Dev.Cid1, err = f.GetCellValue(AddrSheetName, "J2")
	ex.Dev.TransmissionMode, err = f.GetCellValue(AddrSheetName, "J3")
	ex.Dev.Separator, err = f.GetCellValue(AddrSheetName, "O1")
	ex.Dev.Version, err = f.GetCellValue(AddrSheetName, "O2")
	addr, err := f.GetCellValue(AddrSheetName, "O3")
	ex.Dev.Addr = addr
	ex.Addrs = reMapAddrs(f)
	ex.Alarms = reMapAlarms(f)
	ex.Hmis = reMapHmis(f)
	return
}

func reMapAddrs(f *excelize.File) (eAddrs []modu.EAddr) {
	items, err := getExcel2Map(f, AddrSheetName, AddrSheetStartRow)
	if err != nil {
		return
	}
	for _, v := range items {
		var t modu.EAddr
		t.CID1 = getStringFromMap("命令（CID1）", v, "")
		t.CommandExtra = getStringFromMap("命令内容", v, "")
		t.SendPre = getStringFromMap("发送前缀", v, "")
		t.SendSuf = getStringFromMap("发送后缀", v, "")
		t.RevPre = getStringFromMap("接收前缀", v, "")
		t.RevSuf = getStringFromMap("接收后缀", v, "")
		t.Command = getStringFromMap("命令（CID2）", v, "")
		t.MetricName = getStringFromMap("指标名称", v, "")
		t.MetricUnit = getStringFromMap("指标单位", v, "")
		t.MetricCode = getStringFromMap("测点名称", v, "")
		if t.MetricCode == "" {
			t.MetricCode = pinyin.GetCode(t.MetricName)
		}
		t.MetricIndex = getIntFromMap("测点序号", v, 0)
		t.StartAt = getIntFromMap("起始位", v, 0)
		t.Length = getIntFromMap("数据长度", v, 0)
		t.EnumStr = getStringFromMap("枚举", v, "")
		t.CutOffset = getIntFromMap("截取偏移", v, 0)
		t.CutLength = getIntFromMap("截取长度", v, 0)
		t.Scale = getFloat64FromMap("缩放", v, 1.0)
		t.ByteOrder = getStringFromMap("字节排序", v, "")
		t.DataType = getStringFromMap("数据类型", v, "")
		t.ReMap = getStringFromMap("值映射", v, "")
		t.NotZeroAlarm = getStringFromMap("非零告警", v, "")
		t.AlarmCont = getStringFromMap("告警描述", v, "")
		t.Foundation = getFloat64FromMap("偏置", v, 0.0)
		eAddrs = append(eAddrs, t)
	}
	return
}

func reMapAlarms(f *excelize.File) (eAlarms []modu.EAlarm) {
	items, err := getExcel2Map(f, AlarmSheetName, AlarmSheetStartRow)
	if err != nil {
		return
	}
	for _, v := range items {
		var t modu.EAlarm
		t.MetricName = getStringFromMap("指标名称", v, "")
		t.Operator = getStringFromMap("操作符", v, "")
		t.Value = getFloat64FromMap("告警值", v, 0.0)
		t.AlarmCont = getStringFromMap("告警描述", v, "")
		eAlarms = append(eAlarms, t)
	}
	return
}

func reMapHmis(f *excelize.File) (eHims []modu.EHmi) {
	items, err := getExcel2Map(f, HmiSheetName, HmiSheetStartRow)
	if err != nil {
		return
	}
	for _, v := range items {
		var t modu.EHmi
		t.MetricName = getStringFromMap("指标名称", v, "")
		t.FunCode = getIntFromMap("功能码", v, 6)
		t.Address = getIntFromMap("寄存器地址", v, 0)
		t.Scale = getFloat64FromMap("缩放", v, 0)
		t.StepNum = getIntFromMap("步长", v, 100)
		eHims = append(eHims, t)
	}
	return
}

func getStringFromMap(k string, m map[string]string, defaultVal string) string {
	if v, ok := m[k]; ok {
		return v
	}
	return defaultVal
}

func getIntFromMap(k string, m map[string]string, defaultVal int) int {
	if v, ok := m[k]; ok {
		cv, err := strconv.ParseInt(v, 10, 64)
		if err == nil {
			return int(cv)
		}
	}
	return defaultVal
}

func getFloat64FromMap(k string, m map[string]string, defaultVal float64) float64 {
	if v, ok := m[k]; ok {
		cv, err := strconv.ParseFloat(v, 64)
		if err == nil {
			return cv
		}
	}
	return defaultVal
}

func getExcel2Map(f *excelize.File, sheetName string, startRow int) (r []map[string]string, err error) {
	r = make([]map[string]string, 0)
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return
	}
	var titles []string
	titleIndex := startRow - 2

	for index, row := range rows {
		if index < titleIndex {
			continue
		}
		v := make(map[string]string)
		var jumpThis bool
		for j, colCell := range row {
			if index == titleIndex {
				titles = append(titles, colCell)
				jumpThis = true
			} else {
				v[titles[j]] = colCell
			}
		}
		if !jumpThis {
			r = append(r, v)
		}

	}
	return
}

func setExcelDefaultCellValue(f *excelize.File) {
	index, err := f.NewSheet(AddrSheetName)
	if err != nil {
		return
	}
	f.SetActiveSheet(index)

	f.SetCellValue(AddrSheetName, "A1", "设备名称")
	f.SetCellValue(AddrSheetName, "A2", "协议编码")
	f.SetCellValue(AddrSheetName, "A3", "设备类型")

	f.SetCellValue(AddrSheetName, "E1", "接收前缀")
	setDropListCell(f, AddrSheetName, []string{"("}, "F1")
	f.SetCellValue(AddrSheetName, "E2", "接收后缀")
	setDropListCell(f, AddrSheetName, []string{"\\r", "\\n"}, "F2")
	f.SetCellValue(AddrSheetName, "E3", "发送前缀")
	//setDropListCell(f, AddrSheetName, []string{"\\r", "\\n"}, "F3")

	f.SetCellValue(AddrSheetName, "I1", "发送后缀")
	setDropListCell(f, AddrSheetName, []string{"\\r", "\\n"}, "K1")
	f.SetCellValue(AddrSheetName, "I2", "CID1")
	f.SetCellValue(AddrSheetName, "I3", "通讯协议")
	setDropListCell(f, AddrSheetName, TransmissionModeDropList, "J3")
	f.SetCellValue(AddrSheetName, "N1", "指标分割符")
	f.SetCellValue(AddrSheetName, "N2", "版本号")
	setDropListCell(f, AddrSheetName, []string{"空格"}, "O1")
	f.SetCellValue(AddrSheetName, "N3", "通讯地址")

	setDropListToEnd(f, AddrSheetName, ByteOrderDropList, convertToExcelColumn(getColumnNameIndex("字节排序", AddrTitles)), AddrSheetStartRow)

	dataTypeClumn := convertToExcelColumn(getColumnNameIndex("数据类型", AddrTitles))
	f.AddComment(AddrSheetName, excelize.Comment{
		Cell:   fmt.Sprintf("%s%d", dataTypeClumn, AddrSheetStartRow-1),
		Author: "DTCT",
		Paragraph: []excelize.RichTextRun{
			{Text: "通讯类型为Q1时，", Font: &excelize.Font{Bold: true}},
			{Text: "\"FLOAT\", \"MAP\", \"BIN2INT\", \"HEX2INT\","},
		},
	})

	setDropListToEnd(f, AddrSheetName, ScaleDropList, convertToExcelColumn(getColumnNameIndex("缩放", AddrTitles)), AddrSheetStartRow)
	setDropListToEnd(f, AddrSheetName, YesNoDropList, convertToExcelColumn(getColumnNameIndex("非零告警", AddrTitles)), AddrSheetStartRow)

	cutClumn := convertToExcelColumn(getColumnNameIndex("截取偏移", AddrTitles))
	f.AddComment(AddrSheetName, excelize.Comment{
		Cell:   fmt.Sprintf("%s%d", cutClumn, AddrSheetStartRow-1),
		Author: "DTCT",
		Paragraph: []excelize.RichTextRun{
			{Text: "截取偏移: ", Font: &excelize.Font{Bold: true}},
			{Text: "从右往左,从0开始"},
		},
	})

	for columnindex, t := range AddrTitles {
		columnName := convertToExcelColumn(columnindex+1) + fmt.Sprintf("%d", AddrSheetStartRow-1)
		f.SetCellValue(AddrSheetName, columnName, t)
	}

	index, err = f.NewSheet(AlarmSheetName)
	f.SetActiveSheet(index)
	for columnindex, t := range AlarmTitles {
		columnName := convertToExcelColumn(columnindex+1) + fmt.Sprintf("%d", AlarmSheetStartRow-1)
		f.SetCellValue(AlarmSheetName, columnName, t)
	}

	setDropListToEnd(f, AlarmSheetName, OperatorDropList, convertToExcelColumn(getColumnNameIndex("操作符", AlarmTitles)), AlarmSheetStartRow)

	index, err = f.NewSheet(HmiSheetName)
	f.SetActiveSheet(index)
	for columnindex, t := range HmiTitles {
		columnName := convertToExcelColumn(columnindex+1) + fmt.Sprintf("%d", HmiSheetStartRow-1)
		f.SetCellValue(HmiSheetName, columnName, t)
	}
	setDropListToEnd(f, HmiSheetName, HmiScaleDropList, convertToExcelColumn(getColumnNameIndex("缩放", HmiTitles)), HmiSheetStartRow)
	setDropListToEnd(f, HmiSheetName, WriteFuncDropList, convertToExcelColumn(getColumnNameIndex("功能码", HmiTitles)), HmiSheetStartRow)

	c1 := convertToExcelColumn(getColumnNameIndex("指标名称", HmiTitles))
	dv := excelize.NewDataValidation(true)
	dv.SetSqref(fmt.Sprintf("%s%d:%s65535", c1, HmiSheetStartRow, c1))
	c2 := convertToExcelColumn(getColumnNameIndex("指标名称", AddrTitles))
	dv.SetSqrefDropList(fmt.Sprintf("%s!$%s$%d:$%s$65535", AddrSheetName, c2, AddrSheetStartRow, c2))
	f.AddDataValidation(HmiSheetName, dv)

	index, err = f.NewSheet("hide")
	f.SetActiveSheet(index)
	f.SetDefinedName(&excelize.DefinedName{
		Name:     "_Q1",
		RefersTo: "hide!$A:$A",
		Comment:  "_Q1",
		Scope:    "_Q1",
	})

	f.SetDefinedName(&excelize.DefinedName{
		Name:     "_电总",
		RefersTo: "hide!$B:$B",
		Comment:  "_电总",
		Scope:    "_电总",
	})

	dv = excelize.NewDataValidation(true)
	dv.SetSqref(fmt.Sprintf("%s%d:%s65535", dataTypeClumn, AddrSheetStartRow, dataTypeClumn))
	dv.SetSqrefDropList("=INDIRECT(CONCAT(\"_\",$J$3))")
	f.AddDataValidation(AddrSheetName, dv)

	for index, v := range Q1DataType {
		f.SetCellValue("hide", fmt.Sprintf("A%d", index+1), v)
	}

	for index, v := range HDataType {
		f.SetCellValue("hide", fmt.Sprintf("B%d", index+1), v)
	}

	err = f.SetSheetVisible("hide", false)
	if err != nil {
		fmt.Printf(err.Error())
	}
}

func convertToExcelColumn(num int) string {
	if num <= 0 {
		return ""
	}
	var result strings.Builder
	for num > 0 {
		remainder := (num - 1) % 26
		result.WriteRune('A' + rune(remainder))
		num = (num - 1) / 26
	}
	return reverseString(result.String())
}

func getColumnNameIndex(title string, titles []string) (row int) {
	for index, t := range titles {
		if t == title {
			row = index + 1
		}
	}
	return
}

func reverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}
func setDropListToEnd(f *excelize.File, sheetName string, dropList []string, columnName string, startRow int) (err error) {
	drop := excelize.NewDataValidation(true)
	drop.Sqref = fmt.Sprintf("%s%d:%s65535", columnName, startRow, columnName)
	err = drop.SetDropList(dropList)
	err = f.AddDataValidation(sheetName, drop)
	return
}

func setDropListCell(f *excelize.File, sheetName string, dropList []string, cell string) (err error) {
	drop := excelize.NewDataValidation(true)
	drop.Sqref = cell
	err = drop.SetDropList(dropList)
	err = f.AddDataValidation(sheetName, drop)
	return
}

func CheckEParser(parser *modu.EParser) (err error) {
	if len(parser.Addrs) < 1 {
		return errors.New(fmt.Sprintf("%s 协议不完善未能发现有效命令配置", parser.Dev.Name))
	}
	err = checkMainInfo(parser)
	if err != nil {
		return errors.New(fmt.Sprintf("%s %s", parser.Dev.Name, err.Error()))
	}
	err = checkAddrs(parser)
	if err != nil {
		return errors.New(fmt.Sprintf("%s %s", parser.Dev.Name, err.Error()))
	}
	err = checkHmis(parser)
	if err != nil {
		return errors.New(fmt.Sprintf("%s %s", parser.Dev.Name, err.Error()))
	}
	return
}

func checkMainInfo(parser *modu.EParser) (err error) {
	if parser.Dev.DevType == "" || parser.Dev.Code == "" {
		return errors.New("设备类型或编码错误")
	}
	if arrays.ContainsString([]string{"Q1", "电总"}, parser.Dev.TransmissionMode) == -1 {
		return errors.New("通讯协议错误，未设置Q1或电总")
	}
	if parser.Dev.TransmissionMode == "Q1" || parser.Dev.TransmissionMode == "Q1分组" {
		if parser.Dev.RevPre == "" {
			return errors.New("接收前缀未填写")
		}
		if parser.Dev.RevSuf == "" {
			return errors.New("接收后缀未填写")
		}
		if parser.Dev.SendSuf == "" {
			return errors.New("发送后缀未填写")
		}
		if parser.Dev.Separator == "" {
			return errors.New("指标分割符未填写")
		}
	}

	if parser.Dev.TransmissionMode == "电总" {
		if parser.Dev.Cid1 == "" {
			return errors.New("CID1 未填写")
		}
		if parser.Dev.CrcNum == 0 {
			return errors.New("CRC数未填写")
		}
		if parser.Dev.Version == "" {
			return errors.New("版本号 未填写")
		}
		if parser.Dev.Addr == "" {
			return errors.New("通讯地址 未填写")
		}
	}

	return
}
func checkAddrs(parser *modu.EParser) (err error) {
	for _, v := range parser.Addrs {
		if v.MetricName == "" {
			return errors.New(fmt.Sprintf("%s 指标名称未填写", v.MetricName))
		}
		if v.Command == "" {
			return errors.New(fmt.Sprintf("%s 命令/CID2未填写", v.MetricName))
		}

		if parser.Dev.TransmissionMode == "Q1" {
			err = checkQAddr(&v)
			if err != nil {
				return
			}
		}

		if parser.Dev.TransmissionMode == "电总" {
			err = checkHAddr(&v, parser)
			if err != nil {
				return
			}
		}

	}
	return
}

func checkQAddr(addr *modu.EAddr) (err error) {

	if addr.MetricIndex == 0 && addr.Length < 1 {
		return errors.New(fmt.Sprintf("%s 测点序号或数据长度不正确", addr.MetricName))
	}

	if addr.DataType == "" {
		return errors.New(fmt.Sprintf("%s 请填写数据类型", addr.MetricName))
	}

	if addr.DataType == "MAP" && addr.ReMap == "" {
		return errors.New(fmt.Sprintf("%s 数据类型为MAP时需填写值映射", addr.MetricName))
	}
	return
}

func checkHAddr(addr *modu.EAddr, parser *modu.EParser) (err error) {

	if addr.Length < 1 {
		return errors.New(fmt.Sprintf("%s 数据长度不正确", addr.MetricName))
	}

	if addr.ByteOrder == "" {
		return errors.New(fmt.Sprintf("%s 字节排序未填写", addr.MetricName))
	}
	if addr.DataType == "" {
		return errors.New(fmt.Sprintf("%s 数据类型未填写", addr.MetricName))
	}
	dataLength := addr.Length
	if parser.Dev.TransmissionMode == "电总" {
		dataLength = addr.Length / 2
	}
	if dataLength == 1 && addr.DataType != "BIN2INT" {
		if arrays.ContainsString([]string{"AB", "BA"}, addr.ByteOrder) == -1 {
			return errors.New(fmt.Sprintf("%s 数据长度为1时字节序只能为AB或BA", addr.MetricName))
		}
		if arrays.ContainsString([]string{"INT16", "UINT16", "FIXED", "UFIXED"}, addr.DataType) == -1 {
			return errors.New(fmt.Sprintf("%s 数据长度为1时数据类型只能为 \"INT16\", \"UINT16\",\"FIXED\", \"UFIXED\"", addr.MetricName))
		}
	}
	if dataLength == 2 && addr.DataType != "BIN2INT" {
		if arrays.ContainsString([]string{"AB", "BA", "ABCD", "DCBA", "BADC", "CDAB"}, addr.ByteOrder) == -1 {
			return errors.New(fmt.Sprintf("%s 数据长度为2时字节序只能为\"ABCD\", \"DCBA\", \"BADC\", \"CDAB\"", addr.MetricName))
		}
		if arrays.ContainsString([]string{"INT32", "UINT32", "FLOAT32", "FLOAT32-IEEE"}, addr.DataType) == -1 {
			return errors.New(fmt.Sprintf("%s 数据长度为2时数据类型只能为 \"INT32\", \"UINT32\", \"FLOAT32\", \"FLOAT32-IEEE\"", addr.MetricName))
		}
	}
	if dataLength == 4 && addr.DataType != "BIN2INT" {
		if arrays.ContainsString([]string{"AB", "BA", "ABCDEFGH", "GHEFCDAB", "BADCFEHG", "HGFEDCBA"}, addr.ByteOrder) == -1 {
			return errors.New(fmt.Sprintf("%s 数据长度为4时字节序只能为\"ABCDEFGH\", \"GHEFCDAB\", \"BADCFEHG\", \"HGFEDCBA\"", addr.MetricName))
		}
		if arrays.ContainsString([]string{"INT64", "UINT64", "FLOAT64-IEEE"}, addr.DataType) == -1 {
			return errors.New(fmt.Sprintf("%s 数据长度为4时数据类型只能为 \"INT64\", \"UINT64\", \"FLOAT64-IEEE\"", addr.MetricName))
		}
	}

	return
}

func checkHmis(parser *modu.EParser) (err error) {
	for _, v := range parser.Hmis {
		if v.MetricName == "" {
			return errors.New("写屏指标名称不能为空")
		}
		if arrays.ContainsInt([]int64{5, 6, 15, 16}, int64(v.FunCode)) == -1 && v.FunCode != 0 {
			return errors.New(fmt.Sprintf("%s 写屏无效功能码", v.MetricName))
		}
		if v.Address > 65535 {
			return errors.New(fmt.Sprintf("%s 写屏无效寄存器", v.MetricName))
		}
		if v.StepNum == 0 && v.FunCode > 0 {
			return errors.New(fmt.Sprintf("%s 写屏无效步长", v.MetricName))
		}
	}
	return
}
