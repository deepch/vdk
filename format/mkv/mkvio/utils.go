package mkvio

func pack(n int, b []byte) uint64 {
	var v uint64
	var k uint64 = (uint64(n) - 1) * 8

	for i := 0; i < n; i++ {
		v |= uint64(b[i]) << k
		k -= 8
	}

	return v
}

func unpack(n int, v uint64) []byte {
	var b []byte

	for i := uint(n); i > 0; i-- {
		b = append(b, byte(v>>(8*i))&0xff)
	}

	return b
}
