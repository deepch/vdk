package fmp4io

import "github.com/deepch/vdk/utils/bits/pio"

const SIDX = Tag(0x73696478)

type SegmentIndex struct {
	FullAtom
	ReferenceID uint32
	TimeScale   uint32
	EarliestPTS uint64
	FirstOffset uint64
	References  []SegmentReference
}

type SegmentReference struct {
	ReferencesBox      bool
	ReferencedSize     uint32
	SubsegmentDuration uint32
	StartsWithSAP      bool
	SAPType            uint8
	SAPDeltaTime       uint32
}

func (s SegmentIndex) Tag() Tag {
	return SIDX
}

func (s SegmentIndex) Len() (n int) {
	n = s.FullAtom.atomLen()
	n += 4
	n += 4
	if s.Version == 0 {
		n += 4
		n += 4
	} else {
		n += 8
		n += 8
	}
	n += 2
	n += 2
	n += 12 * len(s.References)
	return
}

func (s SegmentIndex) Marshal(b []byte) (n int) {
	n = s.FullAtom.marshalAtom(b, SIDX)
	pio.PutU32BE(b[n:], s.ReferenceID)
	n += 4
	pio.PutU32BE(b[n:], s.TimeScale)
	n += 4
	if s.Version == 0 {
		pio.PutU32BE(b[n:], uint32(s.EarliestPTS))
		n += 4
		pio.PutU32BE(b[n:], uint32(s.FirstOffset))
		n += 4
	} else {
		pio.PutU64BE(b[n:], s.EarliestPTS)
		n += 8
		pio.PutU64BE(b[n:], s.FirstOffset)
		n += 8
	}
	n += 2
	pio.PutU16BE(b[n:], uint16(len(s.References)))
	n += 2
	for _, ref := range s.References {
		v := ref.ReferencedSize
		if ref.ReferencesBox {
			v |= 1 << 31
		}
		pio.PutU32BE(b[n:], v)
		n += 4
		pio.PutU32BE(b[n:], ref.SubsegmentDuration)
		n += 4
		v = (uint32(ref.SAPType) << 28) | ref.SAPDeltaTime
		if ref.StartsWithSAP {
			v |= 1 << 31
		}
		pio.PutU32BE(b[n:], v)
		n += 4
	}
	pio.PutU32BE(b, uint32(n))
	return
}

func (s *SegmentIndex) Unmarshal(b []byte, offset int) (n int, err error) {
	n, err = s.FullAtom.unmarshalAtom(b, offset)
	if err != nil {
		return
	}
	if len(b) < n+8 {
		return 0, parseErr("ReferenceID", n+offset, nil)
	}
	s.ReferenceID = pio.U32BE(b[n:])
	n += 4
	s.TimeScale = pio.U32BE(b[n:])
	n += 4
	if s.Version == 0 {
		if len(b) < n+8 {
			return 0, parseErr("EarliestPTS", n+offset, nil)
		}
		s.EarliestPTS = uint64(pio.U32BE(b[n:]))
		n += 4
		s.FirstOffset = uint64(pio.U32BE(b[n:]))
		n += 4
	} else {
		if len(b) < n+16 {
			return 0, parseErr("EarliestPTS", n+offset, nil)
		}
		s.EarliestPTS = pio.U64BE(b[n:])
		n += 8
		s.FirstOffset = pio.U64BE(b[n:])
		n += 8
	}
	if len(b) < n+4 {
		return 0, parseErr("ReferenceCount", n+offset, nil)
	}
	n += 2
	refCount := int(pio.U16BE(b[n:]))
	n += 2
	if len(b) < n+(12*refCount) {
		return 0, parseErr("SegmentReference", n+offset, nil)
	}
	s.References = make([]SegmentReference, refCount)
	for i := range s.References {
		ref := &s.References[i]
		refSize := pio.U32BE(b[n:])
		n += 4
		if refSize&(1<<31) != 0 {
			ref.ReferencesBox = true
		}
		ref.ReferencedSize = refSize &^ ((1 << 31) - 1)
		ref.SubsegmentDuration = pio.U32BE(b[n:])
		n += 4
		sapDelta := pio.U32BE(b[:n])
		n += 4
		if sapDelta&(1<<31) != 0 {
			ref.StartsWithSAP = true
		}
		ref.SAPType = uint8(0x7 & (sapDelta >> 28))
		ref.SAPDeltaTime = sapDelta &^ ((1 << 28) - 1)
	}
	return
}

func (s SegmentIndex) Children() []Atom {
	return nil
}
