package fmp4io

import (
	"fmt"

	"github.com/deepch/vdk/utils/bits/pio"
)

const MOOF = Tag(0x6d6f6f66)

type MovieFrag struct {
	Header   *MovieFragHeader
	Tracks   []*TrackFrag
	Unknowns []Atom
	AtomPos
}

func (a MovieFrag) Marshal(b []byte) (n int) {
	pio.PutU32BE(b[4:], uint32(MOOF))
	n += a.marshal(b[8:]) + 8
	pio.PutU32BE(b[0:], uint32(n))
	return
}

func (a MovieFrag) marshal(b []byte) (n int) {
	if a.Header != nil {
		n += a.Header.Marshal(b[n:])
	}
	for _, atom := range a.Tracks {
		n += atom.Marshal(b[n:])
	}
	for _, atom := range a.Unknowns {
		n += atom.Marshal(b[n:])
	}
	return
}

func (a MovieFrag) Len() (n int) {
	n += 8
	if a.Header != nil {
		n += a.Header.Len()
	}
	for _, atom := range a.Tracks {
		n += atom.Len()
	}
	for _, atom := range a.Unknowns {
		n += atom.Len()
	}
	return
}

func (a *MovieFrag) Unmarshal(b []byte, offset int) (n int, err error) {
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
		case MFHD:
			{
				atom := &MovieFragHeader{}
				if _, err = atom.Unmarshal(b[n:n+size], offset+n); err != nil {
					err = parseErr("mfhd", n+offset, err)
					return
				}
				a.Header = atom
			}
		case TRAF:
			{
				atom := &TrackFrag{}
				if _, err = atom.Unmarshal(b[n:n+size], offset+n); err != nil {
					err = parseErr("traf", n+offset, err)
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

func (a MovieFrag) Children() (r []Atom) {
	if a.Header != nil {
		r = append(r, a.Header)
	}
	for _, atom := range a.Tracks {
		r = append(r, atom)
	}
	r = append(r, a.Unknowns...)
	return
}

func (a MovieFrag) Tag() Tag {
	return MOOF
}

const MFHD = Tag(0x6d666864)

type MovieFragHeader struct {
	Version uint8
	Flags   uint32
	Seqnum  uint32
	AtomPos
}

func (a MovieFragHeader) Marshal(b []byte) (n int) {
	pio.PutU32BE(b[4:], uint32(MFHD))
	n += a.marshal(b[8:]) + 8
	pio.PutU32BE(b[0:], uint32(n))
	return
}

func (a MovieFragHeader) marshal(b []byte) (n int) {
	pio.PutU8(b[n:], a.Version)
	n += 1
	pio.PutU24BE(b[n:], a.Flags)
	n += 3
	pio.PutU32BE(b[n:], a.Seqnum)
	n += 4
	return
}

func (a MovieFragHeader) Len() (n int) {
	n += 8
	n += 1
	n += 3
	n += 4
	return
}

func (a *MovieFragHeader) Unmarshal(b []byte, offset int) (n int, err error) {
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
		err = parseErr("Seqnum", n+offset, err)
		return
	}
	a.Seqnum = pio.U32BE(b[n:])
	n += 4
	return
}

func (a MovieFragHeader) Children() (r []Atom) {
	return
}

func (a MovieFragHeader) Tag() Tag {
	return MFHD
}

// TRUN is the atom type for TrackFragRun
const TRUN = Tag(0x7472756e)

// TrackFragRun atom
type TrackFragRun struct {
	Version          uint8
	Flags            TrackRunFlags
	DataOffset       uint32
	FirstSampleFlags SampleFlags
	Entries          []TrackFragRunEntry
	AtomPos
}

// TrackRunFlags is the type of TrackFragRun's Flags
type TrackRunFlags uint32

// Defined flags for TrackFragRun
const (
	TrackRunDataOffset       TrackRunFlags = 0x01
	TrackRunFirstSampleFlags TrackRunFlags = 0x04
	TrackRunSampleDuration   TrackRunFlags = 0x100
	TrackRunSampleSize       TrackRunFlags = 0x200
	TrackRunSampleFlags      TrackRunFlags = 0x400
	TrackRunSampleCTS        TrackRunFlags = 0x800
)

func (a TrackFragRun) Marshal(b []byte) (n int) {
	pio.PutU32BE(b[4:], uint32(TRUN))
	n += a.marshal(b[8:]) + 8
	pio.PutU32BE(b[0:], uint32(n))
	return
}

func (a TrackFragRun) marshal(b []byte) (n int) {
	pio.PutU8(b[n:], a.Version)
	n += 1
	pio.PutU24BE(b[n:], uint32(a.Flags))
	n += 3
	pio.PutU32BE(b[n:], uint32(len(a.Entries)))
	n += 4
	if a.Flags&TrackRunDataOffset != 0 {
		{
			pio.PutU32BE(b[n:], a.DataOffset)
			n += 4
		}
	}
	if a.Flags&TrackRunFirstSampleFlags != 0 {
		{
			pio.PutU32BE(b[n:], uint32(a.FirstSampleFlags))
			n += 4
		}
	}

	for _, entry := range a.Entries {
		if a.Flags&TrackRunSampleDuration != 0 {
			pio.PutU32BE(b[n:], entry.Duration)
			n += 4
		}
		if a.Flags&TrackRunSampleSize != 0 {
			pio.PutU32BE(b[n:], entry.Size)
			n += 4
		}
		if a.Flags&TrackRunSampleFlags != 0 {
			pio.PutU32BE(b[n:], uint32(entry.Flags))
			n += 4
		}
		if a.Flags&TrackRunSampleCTS != 0 {
			if a.Version > 0 {
				pio.PutI32BE(b[:n], int32(entry.CTS))
			} else {
				pio.PutU32BE(b[n:], uint32(entry.CTS))
			}
			n += 4
		}
	}
	return
}

func (a TrackFragRun) Len() (n int) {
	n += 8
	n += 1
	n += 3
	n += 4
	if a.Flags&TrackRunDataOffset != 0 {
		{
			n += 4
		}
	}
	if a.Flags&TrackRunFirstSampleFlags != 0 {
		{
			n += 4
		}
	}

	for range a.Entries {
		if a.Flags&TrackRunSampleDuration != 0 {
			n += 4
		}
		if a.Flags&TrackRunSampleSize != 0 {
			n += 4
		}
		if a.Flags&TrackRunSampleFlags != 0 {
			n += 4
		}
		if a.Flags&TrackRunSampleCTS != 0 {
			n += 4
		}
	}
	return
}

func (a *TrackFragRun) Unmarshal(b []byte, offset int) (n int, err error) {
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
	a.Flags = TrackRunFlags(pio.U24BE(b[n:]))
	n += 3
	var _len_Entries uint32
	_len_Entries = pio.U32BE(b[n:])
	n += 4
	a.Entries = make([]TrackFragRunEntry, _len_Entries)
	if a.Flags&TrackRunDataOffset != 0 {
		{
			if len(b) < n+4 {
				err = parseErr("DataOffset", n+offset, err)
				return
			}
			a.DataOffset = pio.U32BE(b[n:])
			n += 4
		}
	}
	if a.Flags&TrackRunFirstSampleFlags != 0 {
		{
			if len(b) < n+4 {
				err = parseErr("FirstSampleFlags", n+offset, err)
				return
			}
			a.FirstSampleFlags = SampleFlags(pio.U32BE(b[n:]))
			n += 4
		}
	}

	for i := 0; i < int(_len_Entries); i++ {
		entry := &a.Entries[i]
		if a.Flags&TrackRunSampleDuration != 0 {
			entry.Duration = pio.U32BE(b[n:])
			n += 4
		}
		if a.Flags&TrackRunSampleSize != 0 {
			entry.Size = pio.U32BE(b[n:])
			n += 4
		}
		if a.Flags&TrackRunSampleFlags != 0 {
			entry.Flags = SampleFlags(pio.U32BE(b[n:]))
			n += 4
		}
		if a.Flags&TrackRunSampleCTS != 0 {
			if a.Version > 0 {
				entry.CTS = int32(pio.I32BE(b[n:]))
			} else {
				entry.CTS = int32(pio.U32BE(b[n:]))
			}
			n += 4
		}
	}
	return
}

func (a TrackFragRun) Children() (r []Atom) {
	return
}

type TrackFragRunEntry struct {
	Duration uint32
	Size     uint32
	Flags    SampleFlags
	CTS      int32
}

func (a TrackFragRun) Tag() Tag {
	return TRUN
}

const TFDT = Tag(0x74666474)

type TrackFragDecodeTime struct {
	Version uint8
	Flags   uint32
	Time    uint64
	AtomPos
}

func (a TrackFragDecodeTime) Marshal(b []byte) (n int) {
	pio.PutU32BE(b[4:], uint32(TFDT))
	n += a.marshal(b[8:]) + 8
	pio.PutU32BE(b[0:], uint32(n))
	return
}

func (a TrackFragDecodeTime) marshal(b []byte) (n int) {
	pio.PutU8(b[n:], a.Version)
	n += 1
	pio.PutU24BE(b[n:], a.Flags)
	n += 3
	if a.Version != 0 {
		pio.PutU64BE(b[n:], a.Time)
		n += 8
	} else {
		pio.PutU32BE(b[n:], uint32(a.Time))
		n += 4
	}
	return
}

func (a TrackFragDecodeTime) Len() (n int) {
	n += 8
	n += 1
	n += 3
	if a.Version != 0 {
		n += 8
	} else {

		n += 4
	}
	return
}

func (a *TrackFragDecodeTime) Unmarshal(b []byte, offset int) (n int, err error) {
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
	if a.Version != 0 {
		a.Time = pio.U64BE(b[n:])
		n += 8
	} else {
		a.Time = uint64(pio.U32BE(b[n:]))
		n += 4
	}
	return
}

func (a TrackFragDecodeTime) Children() (r []Atom) {
	return
}

func (a TrackFragDecodeTime) Tag() Tag {
	return TFDT
}

const TRAF = Tag(0x74726166)

type TrackFrag struct {
	Header     *TrackFragHeader
	DecodeTime *TrackFragDecodeTime
	Run        *TrackFragRun
	Unknowns   []Atom
	AtomPos
}

func (a TrackFrag) Marshal(b []byte) (n int) {
	pio.PutU32BE(b[4:], uint32(TRAF))
	n += a.marshal(b[8:]) + 8
	pio.PutU32BE(b[0:], uint32(n))
	return
}

func (a TrackFrag) marshal(b []byte) (n int) {
	if a.Header != nil {
		n += a.Header.Marshal(b[n:])
	}
	if a.DecodeTime != nil {
		n += a.DecodeTime.Marshal(b[n:])
	}
	if a.Run != nil {
		n += a.Run.Marshal(b[n:])
	}
	for _, atom := range a.Unknowns {
		n += atom.Marshal(b[n:])
	}
	return
}

func (a TrackFrag) Len() (n int) {
	n += 8
	if a.Header != nil {
		n += a.Header.Len()
	}
	if a.DecodeTime != nil {
		n += a.DecodeTime.Len()
	}
	if a.Run != nil {
		n += a.Run.Len()
	}
	for _, atom := range a.Unknowns {
		n += atom.Len()
	}
	return
}

func (a *TrackFrag) Unmarshal(b []byte, offset int) (n int, err error) {
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
		case TFHD:
			{
				atom := &TrackFragHeader{}
				if _, err = atom.Unmarshal(b[n:n+size], offset+n); err != nil {
					err = parseErr("tfhd", n+offset, err)
					return
				}
				a.Header = atom
			}
		case TFDT:
			{
				atom := &TrackFragDecodeTime{}
				if _, err = atom.Unmarshal(b[n:n+size], offset+n); err != nil {
					err = parseErr("tfdt", n+offset, err)
					return
				}
				a.DecodeTime = atom
			}
		case TRUN:
			{
				atom := &TrackFragRun{}
				if _, err = atom.Unmarshal(b[n:n+size], offset+n); err != nil {
					err = parseErr("trun", n+offset, err)
					return
				}
				a.Run = atom
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

func (a TrackFrag) Children() (r []Atom) {
	if a.Header != nil {
		r = append(r, a.Header)
	}
	if a.DecodeTime != nil {
		r = append(r, a.DecodeTime)
	}
	if a.Run != nil {
		r = append(r, a.Run)
	}
	r = append(r, a.Unknowns...)
	return
}

func (a TrackFrag) Tag() Tag {
	return TRAF
}

func (a TrackFragRun) String() string {
	return fmt.Sprintf("dataoffset=%d", a.DataOffset)
}

// TFHD is the atom type for TrackFragHeader
const TFHD = Tag(0x74666864)

// TrackFragHeader atom
type TrackFragHeader struct {
	Version         uint8
	Flags           TrackFragFlags
	TrackID         uint32
	BaseDataOffset  uint64
	StsdID          uint32
	DefaultDuration uint32
	DefaultSize     uint32
	DefaultFlags    SampleFlags
	AtomPos
}

// TrackFragFlags is the type of TrackFragHeader's Flags
type TrackFragFlags uint32

// Defined flags for TrackFragHeader
const (
	TrackFragBaseDataOffset    TrackFragFlags = 0x01
	TrackFragStsdID            TrackFragFlags = 0x02
	TrackFragDefaultDuration   TrackFragFlags = 0x08
	TrackFragDefaultSize       TrackFragFlags = 0x10
	TrackFragDefaultFlags      TrackFragFlags = 0x20
	TrackFragDurationIsEmpty   TrackFragFlags = 0x010000
	TrackFragDefaultBaseIsMOOF TrackFragFlags = 0x020000
)

func (a TrackFragHeader) Marshal(b []byte) (n int) {
	pio.PutU32BE(b[4:], uint32(TFHD))
	n += a.marshal(b[8:]) + 8
	pio.PutU32BE(b[0:], uint32(n))
	return
}

func (a TrackFragHeader) marshal(b []byte) (n int) {
	pio.PutU8(b[n:], a.Version)
	n += 1
	pio.PutU24BE(b[n:], uint32(a.Flags))
	n += 3
	pio.PutU32BE(b[n:], a.TrackID)
	n += 4
	if a.Flags&TrackFragBaseDataOffset != 0 {
		{
			pio.PutU64BE(b[n:], a.BaseDataOffset)
			n += 8
		}
	}
	if a.Flags&TrackFragStsdID != 0 {
		{
			pio.PutU32BE(b[n:], a.StsdID)
			n += 4
		}
	}
	if a.Flags&TrackFragDefaultDuration != 0 {
		{
			pio.PutU32BE(b[n:], a.DefaultDuration)
			n += 4
		}
	}
	if a.Flags&TrackFragDefaultSize != 0 {
		{
			pio.PutU32BE(b[n:], a.DefaultSize)
			n += 4
		}
	}
	if a.Flags&TrackFragDefaultFlags != 0 {
		{
			pio.PutU32BE(b[n:], uint32(a.DefaultFlags))
			n += 4
		}
	}
	return
}

func (a TrackFragHeader) Len() (n int) {
	n += 8
	n += 1
	n += 3
	n += 4
	if a.Flags&TrackFragBaseDataOffset != 0 {
		{
			n += 8
		}
	}
	if a.Flags&TrackFragStsdID != 0 {
		{
			n += 4
		}
	}
	if a.Flags&TrackFragDefaultDuration != 0 {
		{
			n += 4
		}
	}
	if a.Flags&TrackFragDefaultSize != 0 {
		{
			n += 4
		}
	}
	if a.Flags&TrackFragDefaultFlags != 0 {
		{
			n += 4
		}
	}
	return
}

func (a *TrackFragHeader) Unmarshal(b []byte, offset int) (n int, err error) {
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
	a.Flags = TrackFragFlags(pio.U24BE(b[n:]))
	n += 3
	if len(b) < n+4 {
		err = parseErr("TrackID", n+offset, err)
		return
	}
	a.TrackID = pio.U32BE(b[n:])
	n += 4
	if a.Flags&TrackFragBaseDataOffset != 0 {
		{
			if len(b) < n+8 {
				err = parseErr("BaseDataOffset", n+offset, err)
				return
			}
			a.BaseDataOffset = pio.U64BE(b[n:])
			n += 8
		}
	}
	if a.Flags&TrackFragStsdID != 0 {
		{
			if len(b) < n+4 {
				err = parseErr("StsdId", n+offset, err)
				return
			}
			a.StsdID = pio.U32BE(b[n:])
			n += 4
		}
	}
	if a.Flags&TrackFragDefaultDuration != 0 {
		{
			if len(b) < n+4 {
				err = parseErr("DefaultDuration", n+offset, err)
				return
			}
			a.DefaultDuration = pio.U32BE(b[n:])
			n += 4
		}
	}
	if a.Flags&TrackFragDefaultSize != 0 {
		{
			if len(b) < n+4 {
				err = parseErr("DefaultSize", n+offset, err)
				return
			}
			a.DefaultSize = pio.U32BE(b[n:])
			n += 4
		}
	}
	if a.Flags&TrackFragDefaultFlags != 0 {
		{
			if len(b) < n+4 {
				err = parseErr("DefaultFlags", n+offset, err)
				return
			}
			a.DefaultFlags = SampleFlags(pio.U32BE(b[n:]))
			n += 4
		}
	}
	return
}

func (a TrackFragHeader) Children() (r []Atom) {
	return
}

func (a TrackFragHeader) Tag() Tag {
	return TFHD
}

func (a TrackFragHeader) String() string {
	return fmt.Sprintf("basedataoffset=%d", a.BaseDataOffset)
}
