package parser

import (
	"encoding/binary"
	"fmt"
	"math"
)

// 自定义混合字节序类型
type MixedByteOrder struct {
	order []int // 表示字节排列顺序
}

func (m *MixedByteOrder) Uint16(b []byte) uint16 {
	if len(b) < 2 {
		return 0
	}
	var res [2]byte
	for i := 0; i < 2; i++ {
		res[i] = b[m.order[i]]
	}
	return binary.BigEndian.Uint16(res[:])
}

func (m *MixedByteOrder) Uint32(b []byte) uint32 {
	if len(b) < 4 {
		return 0
	}
	var res [4]byte
	for i := 0; i < 4; i++ {
		res[i] = b[m.order[i]]
	}
	return binary.BigEndian.Uint32(res[:])
}

func (m *MixedByteOrder) Uint64(b []byte) uint64 {
	if len(b) < 8 {
		return 0
	}
	var res [8]byte
	for i := 0; i < 8; i++ {
		res[i] = b[m.order[i]]
	}
	return binary.BigEndian.Uint64(res[:])
}

func (m *MixedByteOrder) PutUint16(b []byte, v uint16) {
	if len(b) < 2 {
		return
	}
	src := make([]byte, 2)
	binary.BigEndian.PutUint16(src, v)
	for i := 0; i < 2; i++ {
		b[m.order[i]] = src[i]
	}
}

func (m *MixedByteOrder) PutUint32(b []byte, v uint32) {
	if len(b) < 4 {
		return
	}
	src := make([]byte, 4)
	binary.BigEndian.PutUint32(src, v)
	for i := 0; i < 4; i++ {
		b[m.order[i]] = src[i]
	}
}

func (m *MixedByteOrder) PutUint64(b []byte, v uint64) {
	if len(b) < 8 {
		return
	}
	src := make([]byte, 8)
	binary.BigEndian.PutUint64(src, v)
	for i := 0; i < 8; i++ {
		b[m.order[i]] = src[i]
	}
}

func (m *MixedByteOrder) String() string {
	return fmt.Sprintf("Custom(%v)", m.order)
}

// 获取字节序
func getByteOrder(order string) binary.ByteOrder {
	switch order {
	case "AB", "ABCD", "ABCDEFGH":
		return binary.BigEndian
	case "BA", "DCBA", "HGFEDCBA":
		return binary.LittleEndian
	case "BADC":
		return &MixedByteOrder{order: []int{1, 0, 3, 2}} // Swap each 16-bit word
	case "CDAB":
		return &MixedByteOrder{order: []int{2, 3, 0, 1}} // Swap two 16-bit words
	case "GHEFCDAB":
		return &MixedByteOrder{order: []int{6, 7, 4, 5, 2, 3, 0, 1}}
	default:
		panic("Unsupported byte order: " + order)
	}
}

// 自定义混合顺序处理函数
func reorderBytes(data []byte, order string) []byte {
	if len(order) != len(data) {
		panic("Byte order length must match data length")
	}
	res := make([]byte, len(data))
	for i := 0; i < len(order); i++ {
		srcPos := order[i] - 'A'
		if int(srcPos) >= len(data) {
			panic("Invalid byte order mapping")
		}
		res[i] = data[srcPos]
	}
	return res
}

// 将 bytes 转换为 float64
func convertToFloat(data []byte, dataType, byteOrder string) float64 {
	order := getByteOrder(byteOrder)

	switch dataType {
	case "INT8":
		val := int8(data[0])
		return float64(val)
	case "UINT8":
		val := data[0]
		return float64(val)
	case "INT16":
		if order == nil {
			data = reorderBytes(data[:2], byteOrder)
		} else {
			data = data[:2]
		}
		val := int16(order.Uint16(data))
		return float64(val)
	case "UINT16":
		if order == nil {
			data = reorderBytes(data[:2], byteOrder)
		} else {
			data = data[:2]
		}
		val := order.Uint16(data)
		return float64(val)
	case "UINT32":
		if order == nil {
			data = reorderBytes(data[:4], byteOrder)
		} else {
			data = data[:4]
		}
		val := order.Uint32(data)
		return float64(val)
	case "INT64":
		if order == nil {
			data = reorderBytes(data[:8], byteOrder)
		} else {
			data = data[:8]
		}
		val := int64(order.Uint64(data))
		return float64(val)
	case "UINT64":
		if order == nil {
			data = reorderBytes(data[:8], byteOrder)
		} else {
			data = data[:8]
		}
		val := order.Uint64(data)
		return float64(val)
	case "FLOAT32-IEEE", "FLOAT32":
		if order == nil {
			data = reorderBytes(data[:4], byteOrder)
		} else {
			data = data[:4]
		}
		val := math.Float32frombits(order.Uint32(data))
		return float64(val)
	case "FLOAT64-IEEE", "FLOAT64":
		if order == nil {
			data = reorderBytes(data[:8], byteOrder)
		} else {
			data = data[:8]
		}
		val := math.Float64frombits(order.Uint64(data))
		return val
	case "FIXED":
		// 假设是 32-bit fixed: 16.16 格式
		if order == nil {
			data = reorderBytes(data[:4], byteOrder)
		} else {
			data = data[:4]
		}
		val := int64(order.Uint32(data))
		return float64(val) / (1 << 16)
	case "UFIXED":
		// 假设是 32-bit ufixed: 16.16 格式
		if order == nil {
			data = reorderBytes(data[:4], byteOrder)
		} else {
			data = data[:4]
		}
		val := order.Uint32(data)
		return float64(val) / (1 << 16)
	default:
		panic("Unsupported data type: " + dataType)
	}
}
