package timescale

import (
	"math/bits"
	"time"
)

// ToScale converts a decode time from time.Duration to a specified timescale
func ToScale(t time.Duration, scale uint32) uint64 {
	hi, lo := bits.Mul64(uint64(t), uint64(scale))
	dts, rem := bits.Div64(hi, lo, uint64(time.Second))
	if rem >= uint64(time.Second/2) {
		// round up
		dts++
	}
	return dts
}

// Relative converts a sub-second relative time (which may be negative) to a specified timescale
func Relative(t time.Duration, scale uint32) int32 {
	rel := int64(t) * int64(scale) / int64(time.Second/2)
	if (rel&1 != 0) == (t > 0) {
		// round up
		rel++
	}
	return int32(rel >> 1)
}
