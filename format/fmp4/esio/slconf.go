package esio

import "errors"

type SLConfigDescriptor struct {
	Predefined SLConfigPredefined
	Custom     []byte
}

// SLConfigPredefined references a standard SL config by index
type SLConfigPredefined uint8

// ISO/IEC 14496-1:2004 7.3.2.3.2 Table 12
const (
	SLConfigCustom = SLConfigPredefined(iota)
	SLConfigNull
	SLConfigMP4
)

func parseSLConfig(d []byte) (*SLConfigDescriptor, error) {
	// ISO/IEC 14496-1:2004 7.3.2.3
	if len(d) == 0 {
		return nil, errors.New("SLConfigDescriptor short")
	}
	sl := &SLConfigDescriptor{Predefined: SLConfigPredefined(d[0])}
	if sl.Predefined == SLConfigCustom {
		sl.Custom = d[1:]
	}
	return sl, nil
}

func (c *SLConfigDescriptor) appendTo(b *builder) error {
	if c == nil {
		return nil
	}
	cursor := b.Descriptor(TagSLConfigDescriptor)
	defer cursor.DescriptorDone(-1)
	b.WriteByte(byte(c.Predefined))
	if c.Predefined == SLConfigCustom {
		b.Write(c.Custom)
	}
	return nil
}
