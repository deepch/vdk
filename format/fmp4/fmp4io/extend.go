package fmp4io

import "github.com/deepch/vdk/utils/bits/pio"

const MVEX = Tag(0x6d766578)

type MovieExtend struct {
	Tracks   []*TrackExtend
	Unknowns []Atom
	AtomPos
}

func (a MovieExtend) Marshal(b []byte) (n int) {
	pio.PutU32BE(b[4:], uint32(MVEX))
	n += a.marshal(b[8:]) + 8
	pio.PutU32BE(b[0:], uint32(n))
	return
}

func (a MovieExtend) marshal(b []byte) (n int) {
	for _, atom := range a.Tracks {
		n += atom.Marshal(b[n:])
	}
	for _, atom := range a.Unknowns {
		n += atom.Marshal(b[n:])
	}
	return
}

func (a MovieExtend) Len() (n int) {
	n += 8
	for _, atom := range a.Tracks {
		n += atom.Len()
	}
	for _, atom := range a.Unknowns {
		n += atom.Len()
	}
	return
}

func (a *MovieExtend) Unmarshal(b []byte, offset int) (n int, err error) {
	(&a.AtomPos).setPos(offset, len(b))
	n += 8
	for n+8 < len(b) {
		tag := Tag(pio.U32BE(b[n+4:]))
		size := int(pio.U32BE(b[n:]))
		if len(b) < n+size {
			err = parseErr("TagSizeInvalid", n+offset, err)
			return
		}
		switch tag {
		case TREX:
			{
				atom := &TrackExtend{}
				if _, err = atom.Unmarshal(b[n:n+size], offset+n); err != nil {
					err = parseErr("trex", n+offset, err)
					return
				}
				a.Tracks = append(a.Tracks, atom)
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

func (a MovieExtend) Children() (r []Atom) {
	for _, atom := range a.Tracks {
		r = append(r, atom)
	}
	r = append(r, a.Unknowns...)
	return
}

func (a MovieExtend) Tag() Tag {
	return MVEX
}

const TREX = Tag(0x74726578)

type TrackExtend struct {
	Version               uint8
	Flags                 uint32
	TrackID               uint32
	DefaultSampleDescIdx  uint32
	DefaultSampleDuration uint32
	DefaultSampleSize     uint32
	DefaultSampleFlags    uint32
	AtomPos
}

func (a TrackExtend) Marshal(b []byte) (n int) {
	pio.PutU32BE(b[4:], uint32(TREX))
	n += a.marshal(b[8:]) + 8
	pio.PutU32BE(b[0:], uint32(n))
	return
}

func (a TrackExtend) marshal(b []byte) (n int) {
	pio.PutU8(b[n:], a.Version)
	n += 1
	pio.PutU24BE(b[n:], a.Flags)
	n += 3
	pio.PutU32BE(b[n:], a.TrackID)
	n += 4
	pio.PutU32BE(b[n:], a.DefaultSampleDescIdx)
	n += 4
	pio.PutU32BE(b[n:], a.DefaultSampleDuration)
	n += 4
	pio.PutU32BE(b[n:], a.DefaultSampleSize)
	n += 4
	pio.PutU32BE(b[n:], a.DefaultSampleFlags)
	n += 4
	return
}

func (a TrackExtend) Len() (n int) {
	n += 8
	n += 1
	n += 3
	n += 4
	n += 4
	n += 4
	n += 4
	n += 4
	return
}

func (a *TrackExtend) Unmarshal(b []byte, offset int) (n int, err error) {
	(&a.AtomPos).setPos(offset, len(b))
	n += 8
	if len(b) < n+1 {
		err = parseErr("Version", n+offset, err)
		return
	}
	a.Version = pio.U8(b[n:])
	n += 1
	if len(b) < n+3 {
		err = parseErr("Flags", n+offset, err)
		return
	}
	a.Flags = pio.U24BE(b[n:])
	n += 3
	if len(b) < n+4 {
		err = parseErr("TrackID", n+offset, err)
		return
	}
	a.TrackID = pio.U32BE(b[n:])
	n += 4
	if len(b) < n+4 {
		err = parseErr("DefaultSampleDescIdx", n+offset, err)
		return
	}
	a.DefaultSampleDescIdx = pio.U32BE(b[n:])
	n += 4
	if len(b) < n+4 {
		err = parseErr("DefaultSampleDuration", n+offset, err)
		return
	}
	a.DefaultSampleDuration = pio.U32BE(b[n:])
	n += 4
	if len(b) < n+4 {
		err = parseErr("DefaultSampleSize", n+offset, err)
		return
	}
	a.DefaultSampleSize = pio.U32BE(b[n:])
	n += 4
	if len(b) < n+4 {
		err = parseErr("DefaultSampleFlags", n+offset, err)
		return
	}
	a.DefaultSampleFlags = pio.U32BE(b[n:])
	n += 4
	return
}

func (a TrackExtend) Children() (r []Atom) {
	return
}

func (a TrackExtend) Tag() Tag {
	return TREX
}
