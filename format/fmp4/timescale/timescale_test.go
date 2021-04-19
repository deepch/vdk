package timescale

import (
	"testing"
	"time"
)

func TestToScale(t *testing.T) {
	const scale uint32 = 90000
	values := []struct {
		T time.Duration
		V uint64
	}{
		{0, 0},
		{time.Second/60 - 1, 1500},
		{time.Second/60 + 0, 1500},
		{time.Second/60 + 1, 1500},
		{(time.Second/60)*60 - 1, 90000},
		{(time.Second/60)*60 + 0, 90000},
		{(time.Second/60)*60 + 1, 90000},
		{time.Second * (1 << 32), 90000 * (1 << 32)},
		{time.Second*(1<<32) + time.Second/60 - 1, 90000*(1<<32) + 1500},
		{time.Second*(1<<32) + time.Second/60 + 0, 90000*(1<<32) + 1500},
		{time.Second*(1<<32) + time.Second/60 + 1, 90000*(1<<32) + 1500},
	}
	for _, ex := range values {
		n := ToScale(ex.T, scale)
		if n != ex.V {
			t.Errorf("%d (%s): expected %d, got %d", ex.T, ex.T, ex.V, n)
		}
	}
}

func TestRelative(t *testing.T) {
	const scale uint32 = 90000
	values := []struct {
		T time.Duration
		V int32
	}{
		{0, 0},
		{time.Second/60 - 1, 1500},
		{time.Second/60 + 0, 1500},
		{time.Second/60 + 1, 1500},
		{(time.Second/60)*5 - 1, 7500},
		{(time.Second/60)*5 + 0, 7500},
		{(time.Second/60)*5 + 1, 7500},
		{-time.Second/60 - 1, -1500},
		{-time.Second/60 + 0, -1500},
		{-time.Second/60 + 1, -1500},
		{(-time.Second/60)*5 - 1, -7500},
		{(-time.Second/60)*5 + 0, -7500},
		{(-time.Second/60)*5 + 1, -7500},
	}
	for _, ex := range values {
		n := Relative(ex.T, scale)
		if n != ex.V {
			t.Errorf("%d (%s): expected %d, got %d", ex.T, ex.T, ex.V, n)
		}
	}
}
