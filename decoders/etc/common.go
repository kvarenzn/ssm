package etc

func clampU8(n int) byte {
	if n < 0 {
		return 0
	} else if n > 0xff {
		return 0xff
	} else {
		return byte(n)
	}
}

func applicateColorRaw(c [3]byte) []byte {
	return []byte{c[0], c[1], c[2], 0xff}
}

func applicateColor(c [3]byte, m int16) []byte {
	return []byte{clampU8(int(int16(c[0]) + m)), clampU8(int(int16(c[1]) + m)), clampU8(int(int16(c[2]) + m)), 0xff}
}
