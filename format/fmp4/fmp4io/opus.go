package fmp4io

import (
	"github.com/deepch/vdk/utils/bits/pio"
)

const (
	OPUS = Tag(0x4f707573)
	DOPS = Tag(0x644f7073)
)

type OpusSampleEntry struct {
	DataRefIdx       uint16
	NumberOfChannels uint16
	SampleSize       uint16
	CompressionID    uint16
	SampleRate       float64
	Conf             *OpusSpecificConfiguration
	AtomPos
}

func (a OpusSampleEntry) Tag() Tag { return OPUS }

func (a OpusSampleEntry) Marshal(b []byte) (n int) {
	pio.PutU32BE(b[4:], uint32(OPUS))
	n += a.marshal(b[8:]) + 8
	pio.PutU32BE(b[0:], uint32(n))
	return
}

func (a OpusSampleEntry) marshal(b []byte) (n int) {
	n += 6
	pio.PutU16BE(b[n:], a.DataRefIdx)
	n += 2
	n += 8
	pio.PutU16BE(b[n:], a.NumberOfChannels)
	n += 2
	pio.PutU16BE(b[n:], a.SampleSize)
	n += 2
	n += 4
	PutFixed32(b[n:], a.SampleRate)
	n += 4
	if a.Conf != nil {
		n += a.Conf.Marshal(b[n:])
	}
	return
}

func (a OpusSampleEntry) Len() (n int) {
	n += 8
	n += 6
	n += 2
	n += 8
	n += 2
	n += 2
	n += 4
	n += 4
	if a.Conf != nil {
		n += a.Conf.Len()
	}
	return
}

func (a *OpusSampleEntry) Unmarshal(b []byte, offset int) (n int, err error) {
	(&a.AtomPos).setPos(offset, len(b))
	n += 8
	n += 6
	if len(b) < n+2 {
		err = parseErr("DataRefIdx", n+offset, err)
		return
	}
	a.DataRefIdx = pio.U16BE(b[n:])
	n += 2
	n += 2
	n += 2
	n += 4
	if len(b) < n+2 {
		err = parseErr("NumberOfChannels", n+offset, err)
		return
	}
	a.NumberOfChannels = pio.U16BE(b[n:])
	n += 2
	if len(b) < n+2 {
		err = parseErr("SampleSize", n+offset, err)
		return
	}
	a.SampleSize = pio.U16BE(b[n:])
	n += 2
	n += 2
	n += 2
	if len(b) < n+4 {
		err = parseErr("SampleRate", n+offset, err)
		return
	}
	a.SampleRate = GetFixed32(b[n:])
	n += 4
	for n+8 < len(b) {
		tag := Tag(pio.U32BE(b[n+4:]))
		size := int(pio.U32BE(b[n:]))
		if len(b) < n+size {
			err = parseErr("TagSizeInvalid", n+offset, err)
			return
		}
		switch tag {
		case DOPS:
			{
				atom := &OpusSpecificConfiguration{}
				if _, err = atom.Unmarshal(b[n:n+size], offset+n); err != nil {
					err = parseErr("esds", n+offset, err)
					return
				}
				a.Conf = atom
			}
		}
		n += size
	}
	return
}

func (a OpusSampleEntry) Children() (r []Atom) {
	if a.Conf != nil {
		r = append(r, a.Conf)
	}
	return
}

type OpusSpecificConfiguration struct {
	Version              uint8
	OutputChannelCount   uint8
	PreSkip              uint16
	InputSampleRate      uint32
	OutputGain           int16
	ChannelMappingFamily uint8
	AtomPos
}

func (a OpusSpecificConfiguration) Tag() Tag         { return DOPS }
func (a OpusSpecificConfiguration) Children() []Atom { return nil }

func (a OpusSpecificConfiguration) Len() (n int) {
	n += 8
	n++
	n++
	n += 2
	n += 4
	n += 2
	n++
	return
}

func (a OpusSpecificConfiguration) Marshal(b []byte) (n int) {
	pio.PutU32BE(b[4:], uint32(DOPS))
	n += a.marshal(b[8:]) + 8
	pio.PutU32BE(b[0:], uint32(n))
	return
}

func (a OpusSpecificConfiguration) marshal(b []byte) (n int) {
	pio.PutU8(b[n:], a.Version)
	n++
	pio.PutU8(b[n:], a.OutputChannelCount)
	n++
	pio.PutU16BE(b[n:], a.PreSkip)
	n += 2
	pio.PutU32BE(b[n:], a.InputSampleRate)
	n += 4
	pio.PutI16BE(b[n:], a.OutputGain)
	n += 2
	pio.PutU8(b[n:], a.ChannelMappingFamily)
	n++
	return
}

func (a *OpusSpecificConfiguration) Unmarshal(b []byte, offset int) (n int, err error) {
	a.setPos(offset, len(b))
	n += 8
	if len(b) < 8+11 {
		err = parseErr("OpusSpecificConfiguration", offset, nil)
		return
	}
	a.Version = b[n]
	if a.Version != 0 {
		err = parseErr("unknown version", offset, nil)
		return
	}
	n++
	a.OutputChannelCount = b[n]
	n++
	a.PreSkip = pio.U16BE(b[n:])
	n += 2
	a.InputSampleRate = pio.U32BE(b[n:])
	n += 4
	a.OutputGain = pio.I16BE(b[n:])
	n += 2
	a.ChannelMappingFamily = b[n]
	if a.ChannelMappingFamily != 0 {
		err = parseErr("ChannelMappingFamily", offset+n, nil)
		return
	}
	n++
	return
}
