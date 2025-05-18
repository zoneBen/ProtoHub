package modu

type EParser struct {
	Dev    EDev     `json:"dev"`    // 设备信息
	Addrs  []EAddr  `json:"addrs"`  // 测点配置
	Alarms []EAlarm `json:"alarms"` // 告警设定
	Hmis   []EHmi   `json:"hmis"`   // 写屏设定
}

type EDev struct {
	Name             string `json:"name"`             // 设备名称
	Code             string `json:"code"`             // 协议编码
	DevType          string `json:"devType"`          // 设备类型
	SendPre          string `json:"sendPre"`          // 发送前缀
	SendSuf          string `json:"sendSuf"`          // 发送后缀
	RevPre           string `json:"revPre"`           // 接收前缀
	RevSuf           string `json:"revSuf"`           // 发送后缀
	Cid1             string `json:"cid1"`             // CID1
	TransmissionMode string `json:"transmissionMode"` // 传输方式
	Separator        string `json:"separator"`        // 指标分割符
	Version          string `json:"version"`          // 版本号
	Addr             string `json:"addr"`             // 通讯地址
	CrcNum           int    `json:"crcNum"`           //CRC数
}

type EAddr struct {
	RevPre       string  `json:"revPre"`       // 接收前缀
	CID1         string  `json:"cid1"`         // 命令（CID1）
	Command      string  `json:"command"`      // 命令（CID2）
	CommandExtra string  `json:"commandExtra"` // 命令内容
	MetricName   string  `json:"metricName"`   // 指标名称
	MetricUnit   string  `json:"metricUnit"`   // 指标单位
	MetricCode   string  `json:"metricCode"`   // 测点名称
	MetricIndex  int     `json:"metricIndex"`  // 测点序号
	StartAt      int     `json:"startAt"`      // 起始位
	Length       int     `json:"length"`       // 数据长度
	EnumStr      string  `json:"enumStr"`      // 枚举
	CutOffset    int     `json:"cutOffset"`    // 截取偏移
	CutLength    int     `json:"cutLength"`    // 截取长度
	Scale        float64 `json:"scale"`        // 缩放
	ByteOrder    string  `json:"byteOrder"`    // 字节排序
	DataType     string  `json:"dataType"`     // 数据类型
	ReMap        string  `json:"reMap"`        // 值映射
	NotZeroAlarm string  `json:"notZeroAlarm"` // 非零告警
	AlarmCont    string  `json:"alarmCont"`    // 告警描述
	Foundation   float64 `json:"foundation"`   //偏置
	SendPre      string  `json:"sendPre"`      // 发送前缀
	SendSuf      string  `json:"sendSuf"`      // 发送后缀
	RevSuf       string  `json:"revSuf"`       // 发送后缀
}

type addrSortByCommond []EAddr

func (a addrSortByCommond) Len() int           { return len(a) }
func (a addrSortByCommond) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a addrSortByCommond) Less(i, j int) bool { return a[i].Command < a[j].Command }

type addrSortByStart []EAddr

func (a addrSortByStart) Len() int           { return len(a) }
func (a addrSortByStart) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a addrSortByStart) Less(i, j int) bool { return a[i].StartAt < a[j].StartAt }

type addrSortByMetricIndex []EAddr

func (a addrSortByMetricIndex) Len() int           { return len(a) }
func (a addrSortByMetricIndex) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a addrSortByMetricIndex) Less(i, j int) bool { return a[i].MetricIndex < a[j].MetricIndex }

type EAlarm struct {
	MetricName string  // 指标名称
	Operator   string  // 操作符
	Value      float64 // 告警值
	AlarmCont  string  // 告警描述
}

type EHmi struct {
	MetricName string  // 指标名称
	FunCode    int     // 功能码
	Address    int     // 寄存器地址
	Scale      float64 // 缩放
	StepNum    int     // 步长
}

type ParseValue struct {
	Addr  EAddr
	Value float64
}
