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
