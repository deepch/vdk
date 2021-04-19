package fmp4io

import (
	"math"
	"time"

	"github.com/deepch/vdk/utils/bits/pio"
)

func GetTime32(b []byte) (t time.Time) {
	sec := pio.U32BE(b)
	if sec != 0 {
		t = time.Date(1904, time.January, 1, 0, 0, 0, 0, time.UTC)
		t = t.Add(time.Second * time.Duration(sec))
	}
	return
}

func PutTime32(b []byte, t time.Time) {
	var sec uint32
	if !t.IsZero() {
		dur := t.Sub(time.Date(1904, time.January, 1, 0, 0, 0, 0, time.UTC))
		sec = uint32(dur / time.Second)
	}
	pio.PutU32BE(b, sec)
}

func GetTime64(b []byte) (t time.Time) {
	sec := pio.U64BE(b)
	if sec != 0 {
		t = time.Date(1904, time.January, 1, 0, 0, 0, 0, time.UTC)
		t = t.Add(time.Second * time.Duration(sec))
	}
	return
}

func PutTime64(b []byte, t time.Time) {
	var sec uint64
	if !t.IsZero() {
		dur := t.Sub(time.Date(1904, time.January, 1, 0, 0, 0, 0, time.UTC))
		sec = uint64(dur / time.Second)
	}
	pio.PutU64BE(b, sec)
}

func PutFixed16(b []byte, f float64) {
	intpart, fracpart := math.Modf(f)
	b[0] = uint8(intpart)
	b[1] = uint8(fracpart * 256.0)
}

func GetFixed16(b []byte) float64 {
	return float64(b[0]) + float64(b[1])/256.0
}

func PutFixed32(b []byte, f float64) {
	intpart, fracpart := math.Modf(f)
	pio.PutU16BE(b[0:2], uint16(intpart))
	pio.PutU16BE(b[2:4], uint16(fracpart*65536.0))
}

func GetFixed32(b []byte) float64 {
	return float64(pio.U16BE(b[0:2])) + float64(pio.U16BE(b[2:4]))/65536.0
}
