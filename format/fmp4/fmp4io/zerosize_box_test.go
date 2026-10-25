package fmp4io_test

import (
	"testing"
	"time"

	"github.com/deepch/vdk/format/fmp4/fmp4io"
	"github.com/deepch/vdk/format/mp4/mp4io"
)

// A crafted moov box whose single child declares size==0 with an unknown tag.
// The child-box loop does `n += size`; with size==0, n never advances, so the
// loop appends a Dummy to Unknowns forever and allocates without bound. The
// mp4io twin caps this (len(self.Unknowns) > 100 -> "too many unknowns"); the
// fmp4io copy had no such guard. With the size<8 guard the loop must error out
// instead of spinning.
//
//	b[0:8]   moov box's own header (Movie.Unmarshal just does n += 8)
//	b[8:12]  child size = 0x00000000  (malicious zero length)
//	b[12:16] child tag  = 0x00000000  (unknown -> Dummy branch)
//	b[16:24] padding so that n+8 (=16) < len(b) (=24) stays true
func craftZeroSizeMoov() []byte { return make([]byte, 24) }

func returnsPromptly(t *testing.T, name string, f func() error) {
	t.Helper()
	done := make(chan error, 1)
	go func() { done <- f() }()
	select {
	case err := <-done:
		if err == nil {
			t.Fatalf("%s: expected an error on a zero-size child box, got nil", name)
		}
		t.Logf("%s returned promptly, err=%v", name, err)
	case <-time.After(2 * time.Second):
		t.Fatalf("%s did NOT return in 2s on a zero-size child box (infinite loop / DoS)", name)
	}
}

func TestFmp4MovieZeroSizeChild(t *testing.T) {
	b := craftZeroSizeMoov()
	returnsPromptly(t, "mp4io.Movie.Unmarshal", func() error {
		var mov mp4io.Movie
		_, err := mov.Unmarshal(b, 0)
		return err
	})
	returnsPromptly(t, "fmp4io.Movie.Unmarshal", func() error {
		var mov fmp4io.Movie
		_, err := mov.Unmarshal(b, 0)
		return err
	})
}
