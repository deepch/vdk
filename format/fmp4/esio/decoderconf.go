package esio

import (
	"errors"
	"fmt"

	"github.com/deepch/vdk/av"
	"github.com/deepch/vdk/codec/aacparser"
	"github.com/deepch/vdk/utils/bits/pio"
)

type DecoderConfigDescriptor struct {
	ObjectType ObjectType
	StreamType StreamType
	BufferSize uint32
	MaxBitrate uint32
	AvgBitrate uint32

	AudioSpecific []byte
}

type ObjectType uint8

// ISO/IEC 14496-1 7.2.6.6.2 Table 5
const ObjectTypeAudio = ObjectType(0x40)

type StreamType uint8

// ISO/IEC 14496-1 7.2.6.6.2 Table 6
const StreamTypeAudioStream = StreamType(0x05)

func parseDecoderConfig(d []byte) (*DecoderConfigDescriptor, error) {
	if len(d) < 13 {
		return nil, errors.New("DecoderConfigDescriptor short")
	}
	conf := &DecoderConfigDescriptor{
		ObjectType: ObjectType(d[0]),
		StreamType: StreamType(d[1] >> 2),
		BufferSize: pio.U24BE(d[2:]),
		MaxBitrate: pio.U32BE(d[5:]),
		AvgBitrate: pio.U32BE(d[9:]),
	}
	d = d[13:]
	for len(d) > 0 {
		tag, contents, remainder, err := parseHeader(d)
		if err != nil {
			return nil, fmt.Errorf("DecoderConfigDescriptor: %w", err)
		}
		d = remainder
		switch tag {
		case TagDecoderSpecificInfo:
			switch conf.ObjectType {
			case ObjectTypeAudio:
				conf.AudioSpecific = contents
			}
		}
	}
	return conf, nil
}

func (c *DecoderConfigDescriptor) appendTo(b *builder) error {
	if c == nil {
		return nil
	}
	cursor := b.Descriptor(TagDecoderConfigDescriptor)
	defer cursor.DescriptorDone(-1)
	b.WriteByte(byte(c.ObjectType))
	b.WriteByte(byte(c.StreamType<<2) | 1)
	b.WriteU24(c.BufferSize)
	b.WriteU32(c.MaxBitrate)
	b.WriteU32(c.AvgBitrate)
	switch {
	case c.AudioSpecific != nil:
		// ISO/IEC 14496-3
		// 1.6.2.1 - base AudioSpecificConfig
		// 4.4.1 - GASpecificConfig
		// but we don't actually need to inspect this right now so just preserve the bytes
		c2 := b.Descriptor(TagDecoderSpecificInfo)
		b.Write(c.AudioSpecific)
		c2.DescriptorDone(-1)
	}
	return nil
}

func DecoderConfigFromCodecData(stream av.CodecData) (*DecoderConfigDescriptor, error) {
	switch cd := stream.(type) {
	case aacparser.CodecData:
		return &DecoderConfigDescriptor{
			ObjectType:    ObjectTypeAudio,
			StreamType:    StreamTypeAudioStream,
			AudioSpecific: cd.MPEG4AudioConfigBytes(),
		}, nil
	}
	return nil, fmt.Errorf("can't marshal %T to DecoderConfigDescriptor", stream)
}
