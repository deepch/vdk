package fmp4io

import (
	"github.com/deepch/vdk/format/fmp4/esio"
	"github.com/deepch/vdk/utils/bits/pio"
)

const MP4A = Tag(0x6d703461)

type MP4ADesc struct {
	DataRefIdx       int16
	Version          int16
	RevisionLevel    int16
	Vendor           int32
	NumberOfChannels int16
	SampleSize       int16
	CompressionId    int16
	SampleRate       float64
	Conf             *ElemStreamDesc
	Unknowns         []Atom
	AtomPos
}

func (a MP4ADesc) Marshal(b []byte) (n int) {
	pio.PutU32BE(b[4:], uint32(MP4A))
	n += a.marshal(b[8:]) + 8
	pio.PutU32BE(b[0:], uint32(n))
	return
}

func (a MP4ADesc) marshal(b []byte) (n int) {
	n += 6
	pio.PutI16BE(b[n:], a.DataRefIdx)
	n += 2
	pio.PutI16BE(b[n:], a.Version)
	n += 2
	pio.PutI16BE(b[n:], a.RevisionLevel)
	n += 2
	pio.PutI32BE(b[n:], a.Vendor)
	n += 4
	pio.PutI16BE(b[n:], a.NumberOfChannels)
	n += 2
	pio.PutI16BE(b[n:], a.SampleSize)
	n += 2
	pio.PutI16BE(b[n:], a.CompressionId)
	n += 2
	n += 2
	PutFixed32(b[n:], a.SampleRate)
	n += 4
	if a.Conf != nil {
		n += a.Conf.Marshal(b[n:])
	}
	for _, atom := range a.Unknowns {
		n += atom.Marshal(b[n:])
	}
	return
}

func (a MP4ADesc) Len() (n int) {
	n += 8
	n += 6
	n += 2
	n += 2
	n += 2
	n += 4
	n += 2
	n += 2
	n += 2
	n += 2
	n += 4
	if a.Conf != nil {
		n += a.Conf.Len()
	}
	for _, atom := range a.Unknowns {
		n += atom.Len()
	}
	return
}

func (a *MP4ADesc) Unmarshal(b []byte, offset int) (n int, err error) {
	(&a.AtomPos).setPos(offset, len(b))
	n += 8
	n += 6
	if len(b) < n+2 {
		err = parseErr("DataRefIdx", n+offset, err)
		return
	}
	a.DataRefIdx = pio.I16BE(b[n:])
	n += 2
	if len(b) < n+2 {
		err = parseErr("Version", n+offset, err)
		return
	}
	a.Version = pio.I16BE(b[n:])
	n += 2
	if len(b) < n+2 {
		err = parseErr("RevisionLevel", n+offset, err)
		return
	}
	a.RevisionLevel = pio.I16BE(b[n:])
	n += 2
	if len(b) < n+4 {
		err = parseErr("Vendor", n+offset, err)
		return
	}
	a.Vendor = pio.I32BE(b[n:])
	n += 4
	if len(b) < n+2 {
		err = parseErr("NumberOfChannels", n+offset, err)
		return
	}
	a.NumberOfChannels = pio.I16BE(b[n:])
	n += 2
	if len(b) < n+2 {
		err = parseErr("SampleSize", n+offset, err)
		return
	}
	a.SampleSize = pio.I16BE(b[n:])
	n += 2
	if len(b) < n+2 {
		err = parseErr("CompressionId", n+offset, err)
		return
	}
	a.CompressionId = pio.I16BE(b[n:])
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
		case ESDS:
			{
				atom := &ElemStreamDesc{}
				if _, err = atom.Unmarshal(b[n:n+size], offset+n); err != nil {
					err = parseErr("esds", n+offset, err)
					return
				}
				a.Conf = atom
			}
		default:
			{
				atom := &Dummy{Tag_: tag, Data: b[n : n+size]}
				if _, err = atom.Unmarshal(b[n:n+size], offset+n); err != nil {
					err = parseErr("", n+offset, err)
					return
				}
				a.Unknowns = append(a.Unknowns, atom)
			}
		}
		n += size
	}
	return
}

func (a MP4ADesc) Children() (r []Atom) {
	if a.Conf != nil {
		r = append(r, a.Conf)
	}
	r = append(r, a.Unknowns...)
	return
}

func (a MP4ADesc) Tag() Tag {
	return MP4A
}

const ESDS = Tag(0x65736473)

type ElemStreamDesc struct {
	StreamDescriptor *esio.StreamDescriptor
	AtomPos
}

func (a ElemStreamDesc) Children() []Atom {
	return nil
}

func (a ElemStreamDesc) Len() (n int) {
	blob, _ := a.StreamDescriptor.Marshal()
	return 8 + 4 + len(blob)
}

func (a ElemStreamDesc) Marshal(b []byte) (n int) {
	pio.PutU32BE(b[4:], uint32(ESDS))
	n += 8
	pio.PutU32BE(b[n:], 0) // Version
	n += 4
	blob, err := a.StreamDescriptor.Marshal()
	if err != nil {
		panic(err)
	}
	copy(b[n:], blob)
	n += len(blob)
	pio.PutU32BE(b[0:], uint32(n))
	return
}

func (a *ElemStreamDesc) Unmarshal(b []byte, offset int) (n int, err error) {
	if len(b) < n+12 {
		err = parseErr("hdr", offset+n, err)
		return
	}
	a.AtomPos.setPos(offset, len(b))
	var remainder []byte
	a.StreamDescriptor, remainder, err = esio.ParseStreamDescriptor(b[12:])
	n += len(b) - len(remainder)
	return
}

func (a ElemStreamDesc) Tag() Tag {
	return ESDS
}
