package fmp4io

import (
	"fmt"
	"time"

	"github.com/deepch/vdk/utils/bits/pio"
)

const MOOV = Tag(0x6d6f6f76)

type Movie struct {
	Header      *MovieHeader
	MovieExtend *MovieExtend
	Tracks      []*Track
	Unknowns    []Atom
	AtomPos
}

func (a Movie) Marshal(b []byte) (n int) {
	pio.PutU32BE(b[4:], uint32(MOOV))
	n += a.marshal(b[8:]) + 8
	pio.PutU32BE(b[0:], uint32(n))
	return
}

func (a Movie) marshal(b []byte) (n int) {
	if a.Header != nil {
		n += a.Header.Marshal(b[n:])
	}
	for _, atom := range a.Tracks {
		n += atom.Marshal(b[n:])
	}
	if a.MovieExtend != nil {
		n += a.MovieExtend.Marshal(b[n:])
	}
	for _, atom := range a.Unknowns {
		n += atom.Marshal(b[n:])
	}
	return
}

func (a Movie) Len() (n int) {
	n += 8
	if a.Header != nil {
		n += a.Header.Len()
	}
	for _, atom := range a.Tracks {
		n += atom.Len()
	}
	if a.MovieExtend != nil {
		n += a.MovieExtend.Len()
	}
	for _, atom := range a.Unknowns {
		n += atom.Len()
	}
	return
}

func (a *Movie) Unmarshal(b []byte, offset int) (n int, err error) {
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
		case MVHD:
			{
				atom := &MovieHeader{}
				if _, err = atom.Unmarshal(b[n:n+size], offset+n); err != nil {
					err = parseErr("mvhd", n+offset, err)
					return
				}
				a.Header = atom
			}
		case MVEX:
			{
				atom := &MovieExtend{}
				if _, err = atom.Unmarshal(b[n:n+size], offset+n); err != nil {
					err = parseErr("mvex", n+offset, err)
					return
				}
				a.MovieExtend = atom
			}
		case TRAK:
			{
				atom := &Track{}
				if _, err = atom.Unmarshal(b[n:n+size], offset+n); err != nil {
					err = parseErr("trak", n+offset, err)
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

func (a Movie) Children() (r []Atom) {
	if a.Header != nil {
		r = append(r, a.Header)
	}
	if a.MovieExtend != nil {
		r = append(r, a.MovieExtend)
	}
	for _, atom := range a.Tracks {
		r = append(r, atom)
	}
	r = append(r, a.Unknowns...)
	return
}

func (a Movie) Tag() Tag {
	return MOOV
}

const MVHD = Tag(0x6d766864)

type MovieHeader struct {
	Version         uint8
	Flags           uint32
	CreateTime      time.Time
	ModifyTime      time.Time
	TimeScale       uint32
	Duration        uint32
	PreferredRate   float64
	PreferredVolume float64
	Matrix          [9]int32
	NextTrackID     uint32
	AtomPos
}

func (a MovieHeader) Marshal(b []byte) (n int) {
	pio.PutU32BE(b[4:], uint32(MVHD))
	n += a.marshal(b[8:]) + 8
	pio.PutU32BE(b[0:], uint32(n))
	return
}

func (a MovieHeader) marshal(b []byte) (n int) {
	pio.PutU8(b[n:], a.Version)
	n += 1
	pio.PutU24BE(b[n:], a.Flags)
	n += 3
	PutTime32(b[n:], a.CreateTime)
	n += 4
	PutTime32(b[n:], a.ModifyTime)
	n += 4
	pio.PutU32BE(b[n:], a.TimeScale)
	n += 4
	pio.PutU32BE(b[n:], a.Duration)
	n += 4
	PutFixed32(b[n:], a.PreferredRate)
	n += 4
	PutFixed16(b[n:], a.PreferredVolume)
	n += 2
	n += 10
	for _, entry := range a.Matrix {
		pio.PutI32BE(b[n:], entry)
		n += 4
	}
	n += 24
	pio.PutU32BE(b[n:], a.NextTrackID)
	n += 4
	return
}

func (a MovieHeader) Len() (n int) {
	n += 8
	n += 1
	n += 3
	n += 4
	n += 4
	n += 4
	n += 4
	n += 4
	n += 2
	n += 10
	n += 4 * len(a.Matrix[:])
	n += 24
	n += 4
	return
}

func (a *MovieHeader) Unmarshal(b []byte, offset int) (n int, err error) {
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
		err = parseErr("CreateTime", n+offset, err)
		return
	}
	a.CreateTime = GetTime32(b[n:])
	n += 4
	if len(b) < n+4 {
		err = parseErr("ModifyTime", n+offset, err)
		return
	}
	a.ModifyTime = GetTime32(b[n:])
	n += 4
	if len(b) < n+4 {
		err = parseErr("TimeScale", n+offset, err)
		return
	}
	a.TimeScale = pio.U32BE(b[n:])
	n += 4
	if len(b) < n+4 {
		err = parseErr("Duration", n+offset, err)
		return
	}
	a.Duration = pio.U32BE(b[n:])
	n += 4
	if len(b) < n+4 {
		err = parseErr("PreferredRate", n+offset, err)
		return
	}
	a.PreferredRate = GetFixed32(b[n:])
	n += 4
	if len(b) < n+2 {
		err = parseErr("PreferredVolume", n+offset, err)
		return
	}
	a.PreferredVolume = GetFixed16(b[n:])
	n += 2
	n += 10
	if len(b) < n+4*len(a.Matrix) {
		err = parseErr("Matrix", n+offset, err)
		return
	}
	for i := range a.Matrix {
		a.Matrix[i] = pio.I32BE(b[n:])
		n += 4
	}
	n += 24
	if len(b) < n+4 {
		err = parseErr("NextTrackID", n+offset, err)
		return
	}
	a.NextTrackID = pio.U32BE(b[n:])
	n += 4
	return
}

func (a MovieHeader) Children() (r []Atom) {
	return
}

func (a MovieHeader) Tag() Tag {
	return MVHD
}

func (a MovieHeader) String() string {
	return fmt.Sprintf("dur=%d", a.Duration)
}

const TRAK = Tag(0x7472616b)

type Track struct {
	Header   *TrackHeader
	Media    *Media
	Unknowns []Atom
	AtomPos
}

func (a Track) Marshal(b []byte) (n int) {
	pio.PutU32BE(b[4:], uint32(TRAK))
	n += a.marshal(b[8:]) + 8
	pio.PutU32BE(b[0:], uint32(n))
	return
}

func (a Track) marshal(b []byte) (n int) {
	if a.Header != nil {
		n += a.Header.Marshal(b[n:])
	}
	if a.Media != nil {
		n += a.Media.Marshal(b[n:])
	}
	for _, atom := range a.Unknowns {
		n += atom.Marshal(b[n:])
	}
	return
}

func (a Track) Len() (n int) {
	n += 8
	if a.Header != nil {
		n += a.Header.Len()
	}
	if a.Media != nil {
		n += a.Media.Len()
	}
	for _, atom := range a.Unknowns {
		n += atom.Len()
	}
	return
}

func (a *Track) Unmarshal(b []byte, offset int) (n int, err error) {
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
		case TKHD:
			{
				atom := &TrackHeader{}
				if _, err = atom.Unmarshal(b[n:n+size], offset+n); err != nil {
					err = parseErr("tkhd", n+offset, err)
					return
				}
				a.Header = atom
			}
		case MDIA:
			{
				atom := &Media{}
				if _, err = atom.Unmarshal(b[n:n+size], offset+n); err != nil {
					err = parseErr("mdia", n+offset, err)
					return
				}
				a.Media = atom
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

func (a Track) Children() (r []Atom) {
	if a.Header != nil {
		r = append(r, a.Header)
	}
	if a.Media != nil {
		r = append(r, a.Media)
	}
	r = append(r, a.Unknowns...)
	return
}

func (a Track) Tag() Tag {
	return TRAK
}

func (a *Track) GetAVC1Conf() (conf *AVC1Conf) {
	atom := FindChildren(a, AVCC)
	conf, _ = atom.(*AVC1Conf)
	return
}

func (a *Track) GetElemStreamDesc() (esds *ElemStreamDesc) {
	atom := FindChildren(a, ESDS)
	esds, _ = atom.(*ElemStreamDesc)
	return
}

const TKHD = Tag(0x746b6864)

type TrackHeader struct {
	Version        uint8
	Flags          uint32
	CreateTime     time.Time
	ModifyTime     time.Time
	TrackID        uint32
	Duration       uint32
	Layer          int16
	AlternateGroup int16
	Volume         float64
	Matrix         [9]int32
	TrackWidth     float64
	TrackHeight    float64
	AtomPos
}

func (a TrackHeader) Marshal(b []byte) (n int) {
	pio.PutU32BE(b[4:], uint32(TKHD))
	n += a.marshal(b[8:]) + 8
	pio.PutU32BE(b[0:], uint32(n))
	return
}

func (a TrackHeader) marshal(b []byte) (n int) {
	pio.PutU8(b[n:], a.Version)
	n += 1
	pio.PutU24BE(b[n:], a.Flags)
	n += 3
	PutTime32(b[n:], a.CreateTime)
	n += 4
	PutTime32(b[n:], a.ModifyTime)
	n += 4
	pio.PutU32BE(b[n:], a.TrackID)
	n += 4
	n += 4
	pio.PutU32BE(b[n:], a.Duration)
	n += 4
	n += 8
	pio.PutI16BE(b[n:], a.Layer)
	n += 2
	pio.PutI16BE(b[n:], a.AlternateGroup)
	n += 2
	PutFixed16(b[n:], a.Volume)
	n += 2
	n += 2
	for _, entry := range a.Matrix {
		pio.PutI32BE(b[n:], entry)
		n += 4
	}
	PutFixed32(b[n:], a.TrackWidth)
	n += 4
	PutFixed32(b[n:], a.TrackHeight)
	n += 4
	return
}

func (a TrackHeader) Len() (n int) {
	n += 8
	n += 1
	n += 3
	n += 4
	n += 4
	n += 4
	n += 4
	n += 4
	n += 8
	n += 2
	n += 2
	n += 2
	n += 2
	n += 4 * len(a.Matrix[:])
	n += 4
	n += 4
	return
}

func (a *TrackHeader) Unmarshal(b []byte, offset int) (n int, err error) {
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
		err = parseErr("CreateTime", n+offset, err)
		return
	}
	a.CreateTime = GetTime32(b[n:])
	n += 4
	if len(b) < n+4 {
		err = parseErr("ModifyTime", n+offset, err)
		return
	}
	a.ModifyTime = GetTime32(b[n:])
	n += 4
	if len(b) < n+4 {
		err = parseErr("TrackId", n+offset, err)
		return
	}
	a.TrackID = pio.U32BE(b[n:])
	n += 4
	n += 4
	if len(b) < n+4 {
		err = parseErr("Duration", n+offset, err)
		return
	}
	a.Duration = pio.U32BE(b[n:])
	n += 4
	n += 8
	if len(b) < n+2 {
		err = parseErr("Layer", n+offset, err)
		return
	}
	a.Layer = pio.I16BE(b[n:])
	n += 2
	if len(b) < n+2 {
		err = parseErr("AlternateGroup", n+offset, err)
		return
	}
	a.AlternateGroup = pio.I16BE(b[n:])
	n += 2
	if len(b) < n+2 {
		err = parseErr("Volume", n+offset, err)
		return
	}
	a.Volume = GetFixed16(b[n:])
	n += 2
	n += 2
	if len(b) < n+4*len(a.Matrix) {
		err = parseErr("Matrix", n+offset, err)
		return
	}
	for i := range a.Matrix {
		a.Matrix[i] = pio.I32BE(b[n:])
		n += 4
	}
	if len(b) < n+4 {
		err = parseErr("TrackWidth", n+offset, err)
		return
	}
	a.TrackWidth = GetFixed32(b[n:])
	n += 4
	if len(b) < n+4 {
		err = parseErr("TrackHeight", n+offset, err)
		return
	}
	a.TrackHeight = GetFixed32(b[n:])
	n += 4
	return
}

func (a TrackHeader) Children() (r []Atom) {
	return
}

func (a TrackHeader) Tag() Tag {
	return TKHD
}

const MDAT = Tag(0x6d646174)
