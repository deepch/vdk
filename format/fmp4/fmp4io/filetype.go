package fmp4io

import "github.com/deepch/vdk/utils/bits/pio"

const FTYP = Tag(0x66747970)

type FileType struct {
	MajorBrand       uint32
	MinorVersion     uint32
	CompatibleBrands []uint32
	AtomPos
}

func (t FileType) Tag() Tag {
	return FTYP
}

func (f FileType) Marshal(b []byte) (n int) {
	l := 16 + 4*len(f.CompatibleBrands)
	pio.PutU32BE(b, uint32(l))
	pio.PutU32BE(b[4:], uint32(FTYP))
	pio.PutU32BE(b[8:], f.MajorBrand)
	pio.PutU32BE(b[12:], f.MinorVersion)
	for i, v := range f.CompatibleBrands {
		pio.PutU32BE(b[16+4*i:], v)
	}
	return l
}

func (f FileType) Len() int {
	return 16 + 4*len(f.CompatibleBrands)
}

func (f *FileType) Unmarshal(b []byte, offset int) (n int, err error) {
	f.AtomPos.setPos(offset, len(b))
	n = 8
	if len(b) < n+8 {
		return 0, parseErr("MajorBrand", offset+n, nil)
	}
	f.MajorBrand = pio.U32BE(b[n:])
	n += 4
	f.MinorVersion = pio.U32BE(b[n:])
	n += 4
	for n < len(b)-3 {
		f.CompatibleBrands = append(f.CompatibleBrands, pio.U32BE(b[n:]))
		n += 4
	}
	return
}

func (f FileType) Children() []Atom {
	return nil
}

const STYP = Tag(0x73747970)

type SegmentType struct {
	MajorBrand       uint32
	MinorVersion     uint32
	CompatibleBrands []uint32
	AtomPos
}

func (t SegmentType) Tag() Tag {
	return STYP
}

func (f SegmentType) Marshal(b []byte) (n int) {
	l := 16 + 4*len(f.CompatibleBrands)
	pio.PutU32BE(b, uint32(l))
	pio.PutU32BE(b[4:], uint32(STYP))
	pio.PutU32BE(b[8:], f.MajorBrand)
	pio.PutU32BE(b[12:], f.MinorVersion)
	for i, v := range f.CompatibleBrands {
		pio.PutU32BE(b[16+4*i:], v)
	}
	return l
}

func (f SegmentType) Len() int {
	return 16 + 4*len(f.CompatibleBrands)
}

func (f *SegmentType) Unmarshal(b []byte, offset int) (n int, err error) {
	f.AtomPos.setPos(offset, len(b))
	n = 8
	if len(b) < n+8 {
		return 0, parseErr("MajorBrand", offset+n, nil)
	}
	f.MajorBrand = pio.U32BE(b[n:])
	n += 4
	f.MinorVersion = pio.U32BE(b[n:])
	n += 4
	for n < len(b)-3 {
		f.CompatibleBrands = append(f.CompatibleBrands, pio.U32BE(b[n:]))
		n += 4
	}
	return
}

func (f SegmentType) Children() []Atom {
	return nil
}
