package opusparser

import (
	"errors"
	"time"

	"github.com/deepch/vdk/av"
)

type CodecData struct {
	Channels int
}

func NewCodecData(channels int) *CodecData {
	return &CodecData{Channels: channels}
}

func (d CodecData) Type() av.CodecType {
	return av.OPUS
}

func (d CodecData) SampleRate() int {
	return 48000
}

func (d CodecData) ChannelLayout() av.ChannelLayout {
	switch d.Channels {
	case 1:
		return av.CH_MONO
	case 2:
		return av.CH_STEREO
	default:
		panic("not implemented")
	}
}

func (d CodecData) SampleFormat() av.SampleFormat {
	return av.S16
}

func (d CodecData) PacketDuration(pkt []byte) (time.Duration, error) {
	return PacketDuration(pkt)
}

func Channels(pkt []byte) int {
	if len(pkt) > 0 && (pkt[0]&0x4) == 0 {
		return 1
	}
	return 2
}

func PacketDuration(pkt []byte) (time.Duration, error) {
	if len(pkt) < 1 {
		return 0, errors.New("empty opus packet")
	}
	toc := pkt[0]
	config := toc >> 3
	//stereo := (toc & 0x4) != 0
	code := toc & 0x3
	numFr := 0
	switch code {
	case 0:
		// one frame
		if len(pkt) > 1 {
			numFr = 1
		}
	case 1, 2:
		// two frames
		if len(pkt) > 2 {
			numFr = 2
		}
	case 3:
		// N frames
		if len(pkt) < 2 {
			return 0, errors.New("invalid opus packet")
		}
		numFr = int(pkt[1] & 0x3f)
	}
	return time.Duration(numFr) * opusFrameTimes[config], nil
}

var opusFrameTimes = []time.Duration{
	// SILK NB
	10 * time.Millisecond,
	20 * time.Millisecond,
	40 * time.Millisecond,
	60 * time.Millisecond,
	// SILK MB
	10 * time.Millisecond,
	20 * time.Millisecond,
	40 * time.Millisecond,
	60 * time.Millisecond,
	// SILK WB
	10 * time.Millisecond,
	20 * time.Millisecond,
	40 * time.Millisecond,
	60 * time.Millisecond,
	// Hybrid SWB
	10 * time.Millisecond,
	20 * time.Millisecond,
	// Hybrid FB
	10 * time.Millisecond,
	20 * time.Millisecond,
	// CELT NB
	2500 * time.Microsecond,
	5 * time.Millisecond,
	10 * time.Millisecond,
	20 * time.Millisecond,
	// CELT WB
	2500 * time.Microsecond,
	5 * time.Millisecond,
	10 * time.Millisecond,
	20 * time.Millisecond,
	// CELT SWB
	2500 * time.Microsecond,
	5 * time.Millisecond,
	10 * time.Millisecond,
	20 * time.Millisecond,
	// CELT FB
	2500 * time.Microsecond,
	5 * time.Millisecond,
	10 * time.Millisecond,
	20 * time.Millisecond,
}
