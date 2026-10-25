package h265parser

import (
	"testing"
)

// A truncated HEVC decoder-config record (30 zero bytes: vpscount=0, spscount=0)
// drives Unmarshal past the end of the slice: after the SPS loop it does
// n++; n++; b[n] with no bounds check (the h264parser twin checks len(b) < n+1
// before reading ppscount). On the unfixed code this panics with
// index-out-of-range; with the fix it must return ErrDecconfInvalid and not panic.
func TestH265ConfRecordTruncated(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Unmarshal panicked on truncated input (bug present): %v", r)
		}
	}()

	b := make([]byte, 30) // >= 30 passes the only length guard; all zero
	var rec AVCDecoderConfRecord
	if _, err := rec.Unmarshal(b); err == nil {
		t.Fatalf("expected an error on truncated input, got nil")
	}
}
