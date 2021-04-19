package fmp4io

import (
	"time"

	"github.com/deepch/vdk/utils/bits/pio"
)

const MDIA = Tag(0x6d646961)

type Media struct {
	Header   *MediaHeader
	Handler  *HandlerRefer
	Info     *MediaInfo
	Unknowns []Atom
	AtomPos
}

func (a Media) Marshal(b []byte) (n int) {
	pio.PutU32BE(b[4:], uint32(MDIA))
	n += a.marshal(b[8:]) + 8
	pio.PutU32BE(b[0:], uint32(n))
	return
}

func (a Media) marshal(b []byte) (n int) {
	if a.Header != nil {
		n += a.Header.Marshal(b[n:])
	}
	if a.Handler != nil {
		n += a.Handler.Marshal(b[n:])
	}
	if a.Info != nil {
		n += a.Info.Marshal(b[n:])
	}
	for _, atom := range a.Unknowns {
		n += atom.Marshal(b[n:])
	}
	return
}

func (a Media) Len() (n int) {
	n += 8
	if a.Header != nil {
		n += a.Header.Len()
	}
	if a.Handler != nil {
		n += a.Handler.Len()
	}
	if a.Info != nil {
		n += a.Info.Len()
	}
	for _, atom := range a.Unknowns {
		n += atom.Len()
	}
	return
}

func (a *Media) Unmarshal(b []byte, offset int) (n int, err error) {
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
		case MDHD:
			{
				atom := &MediaHeader{}
				if _, err = atom.Unmarshal(b[n:n+size], offset+n); err != nil {
					err = parseErr("mdhd", n+offset, err)
					return
				}
				a.Header = atom
			}
		case HDLR:
			{
				atom := &HandlerRefer{}
				if _, err = atom.Unmarshal(b[n:n+size], offset+n); err != nil {
					err = parseErr("hdlr", n+offset, err)
					return
				}
				a.Handler = atom
			}
		case MINF:
			{
				atom := &MediaInfo{}
				if _, err = atom.Unmarshal(b[n:n+size], offset+n); err != nil {
					err = parseErr("minf", n+offset, err)
					return
				}
				a.Info = atom
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

func (a Media) Children() (r []Atom) {
	if a.Header != nil {
		r = append(r, a.Header)
	}
	if a.Handler != nil {
		r = append(r, a.Handler)
	}
	if a.Info != nil {
		r = append(r, a.Info)
	}
	r = append(r, a.Unknowns...)
	return
}

func (a Media) Tag() Tag {
	return MDIA
}

const MDHD = Tag(0x6d646864)

type MediaHeader struct {
	Version    uint8
	Flags      uint32
	CreateTime time.Time
	ModifyTime time.Time
	TimeScale  uint32
	Duration   uint32
	Language   int16
	Quality    int16
	AtomPos
}

func (a MediaHeader) Marshal(b []byte) (n int) {
	pio.PutU32BE(b[4:], uint32(MDHD))
	n += a.marshal(b[8:]) + 8
	pio.PutU32BE(b[0:], uint32(n))
	return
}

func (a MediaHeader) marshal(b []byte) (n int) {
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
	pio.PutI16BE(b[n:], a.Language)
	n += 2
	pio.PutI16BE(b[n:], a.Quality)
	n += 2
	return
}

func (a MediaHeader) Len() (n int) {
	n += 8
	n += 1
	n += 3
	n += 4
	n += 4
	n += 4
	n += 4
	n += 2
	n += 2
	return
}

func (a *MediaHeader) Unmarshal(b []byte, offset int) (n int, err error) {
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
	if len(b) < n+2 {
		err = parseErr("Language", n+offset, err)
		return
	}
	a.Language = pio.I16BE(b[n:])
	n += 2
	if len(b) < n+2 {
		err = parseErr("Quality", n+offset, err)
		return
	}
	a.Quality = pio.I16BE(b[n:])
	n += 2
	return
}

func (a MediaHeader) Children() (r []Atom) {
	return
}

func (a MediaHeader) Tag() Tag {
	return MDHD
}

const MINF = Tag(0x6d696e66)

type MediaInfo struct {
	Sound    *SoundMediaInfo
	Video    *VideoMediaInfo
	Data     *DataInfo
	Sample   *SampleTable
	Unknowns []Atom
	AtomPos
}

func (a MediaInfo) Marshal(b []byte) (n int) {
	pio.PutU32BE(b[4:], uint32(MINF))
	n += a.marshal(b[8:]) + 8
	pio.PutU32BE(b[0:], uint32(n))
	return
}

func (a MediaInfo) marshal(b []byte) (n int) {
	if a.Sound != nil {
		n += a.Sound.Marshal(b[n:])
	}
	if a.Video != nil {
		n += a.Video.Marshal(b[n:])
	}
	if a.Data != nil {
		n += a.Data.Marshal(b[n:])
	}
	if a.Sample != nil {
		n += a.Sample.Marshal(b[n:])
	}
	for _, atom := range a.Unknowns {
		n += atom.Marshal(b[n:])
	}
	return
}

func (a MediaInfo) Len() (n int) {
	n += 8
	if a.Sound != nil {
		n += a.Sound.Len()
	}
	if a.Video != nil {
		n += a.Video.Len()
	}
	if a.Data != nil {
		n += a.Data.Len()
	}
	if a.Sample != nil {
		n += a.Sample.Len()
	}
	for _, atom := range a.Unknowns {
		n += atom.Len()
	}
	return
}

func (a *MediaInfo) Unmarshal(b []byte, offset int) (n int, err error) {
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
		case SMHD:
			{
				atom := &SoundMediaInfo{}
				if _, err = atom.Unmarshal(b[n:n+size], offset+n); err != nil {
					err = parseErr("smhd", n+offset, err)
					return
				}
				a.Sound = atom
			}
		case VMHD:
			{
				atom := &VideoMediaInfo{}
				if _, err = atom.Unmarshal(b[n:n+size], offset+n); err != nil {
					err = parseErr("vmhd", n+offset, err)
					return
				}
				a.Video = atom
			}
		case DINF:
			{
				atom := &DataInfo{}
				if _, err = atom.Unmarshal(b[n:n+size], offset+n); err != nil {
					err = parseErr("dinf", n+offset, err)
					return
				}
				a.Data = atom
			}
		case STBL:
			{
				atom := &SampleTable{}
				if _, err = atom.Unmarshal(b[n:n+size], offset+n); err != nil {
					err = parseErr("stbl", n+offset, err)
					return
				}
				a.Sample = atom
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

func (a MediaInfo) Children() (r []Atom) {
	if a.Sound != nil {
		r = append(r, a.Sound)
	}
	if a.Video != nil {
		r = append(r, a.Video)
	}
	if a.Data != nil {
		r = append(r, a.Data)
	}
	if a.Sample != nil {
		r = append(r, a.Sample)
	}
	r = append(r, a.Unknowns...)
	return
}

func (a MediaInfo) Tag() Tag {
	return MINF
}

const VMHD = Tag(0x766d6864)

type VideoMediaInfo struct {
	Version      uint8
	Flags        uint32
	GraphicsMode int16
	Opcolor      [3]int16
	AtomPos
}

func (a VideoMediaInfo) Marshal(b []byte) (n int) {
	pio.PutU32BE(b[4:], uint32(VMHD))
	n += a.marshal(b[8:]) + 8
	pio.PutU32BE(b[0:], uint32(n))
	return
}

func (a VideoMediaInfo) marshal(b []byte) (n int) {
	pio.PutU8(b[n:], a.Version)
	n += 1
	pio.PutU24BE(b[n:], a.Flags)
	n += 3
	pio.PutI16BE(b[n:], a.GraphicsMode)
	n += 2
	for _, entry := range a.Opcolor {
		pio.PutI16BE(b[n:], entry)
		n += 2
	}
	return
}

func (a VideoMediaInfo) Len() (n int) {
	n += 8
	n += 1
	n += 3
	n += 2
	n += 2 * len(a.Opcolor[:])
	return
}

func (a *VideoMediaInfo) Unmarshal(b []byte, offset int) (n int, err error) {
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
	if len(b) < n+2 {
		err = parseErr("GraphicsMode", n+offset, err)
		return
	}
	a.GraphicsMode = pio.I16BE(b[n:])
	n += 2
	if len(b) < n+2*len(a.Opcolor) {
		err = parseErr("Opcolor", n+offset, err)
		return
	}
	for i := range a.Opcolor {
		a.Opcolor[i] = pio.I16BE(b[n:])
		n += 2
	}
	return
}

func (a VideoMediaInfo) Children() (r []Atom) {
	return
}

func (a VideoMediaInfo) Tag() Tag {
	return VMHD
}

const SMHD = Tag(0x736d6864)

type SoundMediaInfo struct {
	Version uint8
	Flags   uint32
	Balance int16
	AtomPos
}

func (a SoundMediaInfo) Marshal(b []byte) (n int) {
	pio.PutU32BE(b[4:], uint32(SMHD))
	n += a.marshal(b[8:]) + 8
	pio.PutU32BE(b[0:], uint32(n))
	return
}

func (a SoundMediaInfo) marshal(b []byte) (n int) {
	pio.PutU8(b[n:], a.Version)
	n += 1
	pio.PutU24BE(b[n:], a.Flags)
	n += 3
	pio.PutI16BE(b[n:], a.Balance)
	n += 2
	n += 2
	return
}

func (a SoundMediaInfo) Len() (n int) {
	n += 8
	n += 1
	n += 3
	n += 2
	n += 2
	return
}

func (a *SoundMediaInfo) Unmarshal(b []byte, offset int) (n int, err error) {
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
	if len(b) < n+2 {
		err = parseErr("Balance", n+offset, err)
		return
	}
	a.Balance = pio.I16BE(b[n:])
	n += 2
	n += 2
	return
}

func (a SoundMediaInfo) Children() (r []Atom) {
	return
}

func (a SoundMediaInfo) Tag() Tag {
	return SMHD
}

const DINF = Tag(0x64696e66)

func (a DataInfo) Tag() Tag {
	return DINF
}

type DataInfo struct {
	Refer    *DataRefer
	Unknowns []Atom
	AtomPos
}

func (a DataInfo) Marshal(b []byte) (n int) {
	pio.PutU32BE(b[4:], uint32(DINF))
	n += a.marshal(b[8:]) + 8
	pio.PutU32BE(b[0:], uint32(n))
	return
}

func (a DataInfo) marshal(b []byte) (n int) {
	if a.Refer != nil {
		n += a.Refer.Marshal(b[n:])
	}
	for _, atom := range a.Unknowns {
		n += atom.Marshal(b[n:])
	}
	return
}

func (a DataInfo) Len() (n int) {
	n += 8
	if a.Refer != nil {
		n += a.Refer.Len()
	}
	for _, atom := range a.Unknowns {
		n += atom.Len()
	}
	return
}

func (a *DataInfo) Unmarshal(b []byte, offset int) (n int, err error) {
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
		case DREF:
			{
				atom := &DataRefer{}
				if _, err = atom.Unmarshal(b[n:n+size], offset+n); err != nil {
					err = parseErr("dref", n+offset, err)
					return
				}
				a.Refer = atom
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

func (a DataInfo) Children() (r []Atom) {
	if a.Refer != nil {
		r = append(r, a.Refer)
	}
	r = append(r, a.Unknowns...)
	return
}
