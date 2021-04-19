package fmp4

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/deepch/vdk/av"
	"github.com/deepch/vdk/format/fmp4/fmp4io"
	"github.com/deepch/vdk/format/fmp4/fragment"
)

var (
	shdrOnce sync.Once
	shdr     []byte
)

// MovieFragmenter breaks a stream into segments each containing both tracks from the original stream
type MovieFragmenter struct {
	tracks []*TrackFragmenter
	fhdr   []byte
	vidx   int
	seqNum uint32
	shdrw  bool
}

// NewMovie creates a movie fragmenter from a stream
func NewMovie(streams []av.CodecData) (*MovieFragmenter, error) {
	f := &MovieFragmenter{
		tracks: make([]*TrackFragmenter, len(streams)),
		vidx:   -1,
	}
	atoms := make([]*fmp4io.Track, len(streams))
	var err error
	for i, cd := range streams {
		f.tracks[i], err = NewTrack(cd)
		if err != nil {
			return nil, fmt.Errorf("track %d: %w", i, err)
		}
		atoms[i] = f.tracks[i].atom
		if cd.Type().IsVideo() {
			f.vidx = i
		}
	}
	if f.vidx < 0 {
		return nil, errors.New("no video track found")
	}
	f.fhdr, err = MovieHeader(atoms)
	if err != nil {
		return nil, err
	}
	return f, err
}

// Fragment produces a fragment out of the currently-queued packets.
func (f *MovieFragmenter) Fragment() (fragment.Fragment, error) {
	dur := f.tracks[f.vidx].Duration()
	var tracks []fragmentWithData
	for _, track := range f.tracks {
		tf := track.makeFragment()
		if tf.trackFrag != nil {
			tracks = append(tracks, tf)
		}
	}
	if len(tracks) == 0 {
		return fragment.Fragment{}, nil
	}
	f.seqNum++
	initial := !f.shdrw
	f.shdrw = true
	frag := marshalFragment(tracks, f.seqNum, initial)
	frag.Duration = dur
	return frag, nil
}

// WritePacket formats and queues a packet for the next fragment to be written
func (f *MovieFragmenter) WritePacket(pkt av.Packet) error {
	return f.tracks[pkt.Idx].WritePacket(pkt)
}

// Duration calculates the elapsed duration between the first and last pending video frame
func (f *MovieFragmenter) Duration() time.Duration {
	return f.tracks[f.vidx].Duration()
}

// MovieHeader marshals an init.mp4 for the fragmenter's tracks
func (f *MovieFragmenter) MovieHeader() (filename, contentType string, blob []byte) {
	return "init.mp4", "video/mp4", f.fhdr
}

// NewSegment indicates that a new segment has begun and  the next call to
// Fragment() should include a leading FTYP header
func (f *MovieFragmenter) NewSegment() {
	f.shdrw = false
}

func (f *MovieFragmenter) TimeScale() uint32 {
	return 90000
}
