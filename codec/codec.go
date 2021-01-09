package codec

import (
	"time"

	"github.com/deepch/vdk/av"
	"github.com/deepch/vdk/codec/fake"
)

type OpusCodecData struct {
	typ            av.CodecType
	SampleRate_    int
	ChannelLayout_ av.ChannelLayout
}

func (self OpusCodecData) Type() av.CodecType {
	return self.typ
}

func (self OpusCodecData) SampleRate() int {
	return self.SampleRate_
}

func (self OpusCodecData) ChannelLayout() av.ChannelLayout {
	return self.ChannelLayout_
}

func (self OpusCodecData) PacketDuration(data []byte) (time.Duration, error) {
	return time.Duration(20) * time.Millisecond, nil
}

func (self OpusCodecData) SampleFormat() av.SampleFormat {
	return av.FLT
}

type PCMUCodecData struct {
	typ av.CodecType
}

func (self PCMUCodecData) Type() av.CodecType {
	return self.typ
}

func (self PCMUCodecData) SampleRate() int {
	return 8000
}

func (self PCMUCodecData) ChannelLayout() av.ChannelLayout {
	return av.CH_MONO
}

func (self PCMUCodecData) SampleFormat() av.SampleFormat {
	return av.S16
}

func (self PCMUCodecData) PacketDuration(data []byte) (time.Duration, error) {
	return time.Duration(len(data)) * time.Second / time.Duration(8000), nil
}

func NewPCMMulawCodecData() av.AudioCodecData {
	return PCMUCodecData{
		typ: av.PCM_MULAW,
	}
}

func NewPCMCodecData() av.AudioCodecData {
	return PCMUCodecData{
		typ: av.PCM,
	}
}

func NewPCMAlawCodecData() av.AudioCodecData {
	return PCMUCodecData{
		typ: av.PCM_ALAW,
	}
}
func NewOpusCodecData(sr int, cc av.ChannelLayout) av.AudioCodecData {
	return OpusCodecData{
		typ:            av.OPUS,
		SampleRate_:    sr,
		ChannelLayout_: cc,
	}
}

type SpeexCodecData struct {
	fake.CodecData
}

func (self SpeexCodecData) PacketDuration(data []byte) (time.Duration, error) {
	return time.Millisecond * 20, nil
}

func NewSpeexCodecData(sr int, cl av.ChannelLayout) SpeexCodecData {
	codec := SpeexCodecData{}
	codec.CodecType_ = av.SPEEX
	codec.SampleFormat_ = av.S16
	codec.SampleRate_ = sr
	codec.ChannelLayout_ = cl
	return codec
}
