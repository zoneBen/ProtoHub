package parser

const (
	zero  = byte('0')
	one   = byte('1')
	lsb   = byte('[') // left square brackets
	rsb   = byte(']') // right square brackets
	space = byte(' ')
)

// append bytes of string in binary format.
func appendBinaryString(bs []byte, b byte) []byte {
	var a byte
	for i := 0; i < 8; i++ {
		a = b
		b <<= 1
		b >>= 1
		switch a {
		case b:
			bs = append(bs, zero)
		default:
			bs = append(bs, one)
		}
		b <<= 1
	}
	return bs
}

// ByteToBinaryString get the string in binary format of a byte or uint8.
func ByteToBinaryString(b byte) string {
	buf := make([]byte, 0, 8)
	buf = appendBinaryString(buf, b)
	return string(buf)
}

// BytesToBinaryString get the string in binary format of a []byte or []int8.
func BytesToBinaryString(bs []byte) string {
	l := len(bs)
	bl := l*8 + l + 1
	buf := make([]byte, 0, bl)
	//buf = append(buf, lsb)
	for _, b := range bs {
		buf = appendBinaryString(buf, b)
		//buf = append(buf, space)
	}
	//buf[bl-1] = rsb
	return string(buf)
}

func parseSignMagnitude(b byte) int8 {
	// 提取符号位（最高位）
	signBit := (b >> 7) & 1 // 0 或 1

	// 提取数值部分（低7位）
	magnitude := b & 0x7F // 取低7位

	// 转换为有符号整数
	if signBit == 1 {
		return -int8(magnitude)
	}
	return int8(magnitude)
}

func parseSignMagnitude16(data [2]byte) int16 {
	// 合并两个字节为一个 uint16
	val := uint16(data[0])<<8 | uint16(data[1])

	// 提取符号位
	signBit := (val >> 15) & 1

	// 提取数值部分（低15位）
	magnitude := val & 0x7FFF

	if signBit == 1 {
		return -int16(magnitude)
	}
	return int16(magnitude)
}
