package esio

import (
	"errors"
	"fmt"

	"github.com/deepch/vdk/utils/bits/pio"
)

type StreamDescriptor struct {
	ESID      uint16
	DependsOn *uint16
	URL       *string
	OCR       *uint16

	DecoderConfig *DecoderConfigDescriptor
	SLConfig      *SLConfigDescriptor
}

// Tag identifies element stream descriptor types
type Tag uint8

// ISO/IEC 14496-1:2004 7.2.2 Table 1
const (
	TagForbidden = Tag(iota)
	TagObjectDescriptor
	TagInitialObjectDescriptor
	TagESDescriptor
	TagDecoderConfigDescriptor
	TagDecoderSpecificInfo
	TagSLConfigDescriptor
)

const (
	esFlagStreamDependence = 0x80
	esFlagURL              = 0x40
	esFlagOCR              = 0x20
)

func ParseStreamDescriptor(start []byte) (desc *StreamDescriptor, remainder []byte, err error) {
	// ISO/IEC 14496-1:2004 7.2.6.5.1
	tag, d, remainder, err := parseHeader(start)
	if err != nil {
		err = fmt.Errorf("ES_Descriptor: %w", err)
		return
	} else if tag != TagESDescriptor {
		err = fmt.Errorf("expected ES_Descriptor but got tag %02X", tag)
		return
	}
	desc = &StreamDescriptor{ESID: pio.U16BE(d)}
	flags := d[2]
	d = d[3:]
	if flags&esFlagStreamDependence != 0 {
		v := pio.U16BE(d)
		desc.DependsOn = &v
		d = d[2:]
	}
	if flags&esFlagURL != 0 {
		urlLength := d[0]
		v := string(d[1 : 1+urlLength])
		desc.URL = &v
		d = d[1+urlLength:]
	}
	if flags&esFlagOCR != 0 {
		v := pio.U16BE(d)
		desc.OCR = &v
		d = d[2:]
	}
	for len(d) > 0 {
		var child []byte
		tag, child, d, err = parseHeader(d)
		if err != nil {
			err = fmt.Errorf("ES_Descriptor: %w", err)
			return
		}
		switch tag {
		case TagDecoderConfigDescriptor:
			desc.DecoderConfig, err = parseDecoderConfig(child)
		case TagSLConfigDescriptor:
			desc.SLConfig, err = parseSLConfig(child)
		}
		if err != nil {
			return
		}
	}
	remainder = d
	return
}

func (s *StreamDescriptor) Marshal() ([]byte, error) {
	var b builder
	cursor := b.Descriptor(TagESDescriptor)
	b.WriteU16(s.ESID)
	var flags uint8
	if s.DependsOn != nil {
		flags |= esFlagStreamDependence
	}
	if s.URL != nil {
		flags |= esFlagURL
	}
	if s.OCR != nil {
		flags |= esFlagOCR
	}
	b.WriteByte(flags)
	if s.DependsOn != nil {
		b.WriteU16(*s.DependsOn)
	}
	if s.URL != nil {
		b.WriteByte(byte(len(*s.URL)))
		b.Write([]byte(*s.URL))
	}
	if s.OCR != nil {
		b.WriteU16(*s.OCR)
	}
	if err := s.DecoderConfig.appendTo(&b); err != nil {
		return nil, err
	}
	if err := s.SLConfig.appendTo(&b); err != nil {
		return nil, err
	}
	cursor.DescriptorDone(-1)
	return b.Bytes(), nil
}

func parseLength(start []byte) (length int, d []byte, err error) {
	// ISO/IEC 14496-1:2004 8.3.3
	d = start
	for i := 0; i < 4; i++ {
		if len(d) == 0 {
			err = errors.New("short tag")
			return
		}
		v := d[0]
		d = d[1:]
		length <<= 7
		length |= int(v & 0x7f)
		if v&0x80 == 0 {
			break
		}
	}
	return
}

func parseHeader(start []byte) (tag Tag, contents, d []byte, err error) {
	d = start
	if len(d) < 2 {
		err = errors.New("short tag")
		return
	}
	tag = Tag(d[0])
	length, d, err := parseLength(d[1:])
	if err != nil {
		return
	}
	if length > len(d) {
		err = fmt.Errorf("short tag: %02x: expected %d bytes but only got %d", tag, length, len(d))
		return
	}
	contents = d[:length]
	d = d[length:]
	return
}
