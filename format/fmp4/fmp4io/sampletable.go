package fmp4io

import (
	"fmt"

	"github.com/deepch/vdk/utils/bits/pio"
)

const STBL = Tag(0x7374626c)

type SampleTable struct {
	SampleDesc        *SampleDesc
	TimeToSample      *TimeToSample
	CompositionOffset *CompositionOffset
	SampleToChunk     *SampleToChunk
	SyncSample        *SyncSample
	ChunkOffset       *ChunkOffset
	SampleSize        *SampleSize
	AtomPos
}

func (a SampleTable) Marshal(b []byte) (n int) {
	pio.PutU32BE(b[4:], uint32(STBL))
	n += a.marshal(b[8:]) + 8
	pio.PutU32BE(b[0:], uint32(n))
	return
}

func (a SampleTable) marshal(b []byte) (n int) {
	if a.SampleDesc != nil {
		n += a.SampleDesc.Marshal(b[n:])
	}
	if a.TimeToSample != nil {
		n += a.TimeToSample.Marshal(b[n:])
	}
	if a.CompositionOffset != nil {
		n += a.CompositionOffset.Marshal(b[n:])
	}
	if a.SampleToChunk != nil {
		n += a.SampleToChunk.Marshal(b[n:])
	}
	if a.SyncSample != nil {
		n += a.SyncSample.Marshal(b[n:])
	}
	if a.SampleSize != nil {
		n += a.SampleSize.Marshal(b[n:])
	}
	if a.ChunkOffset != nil {
		n += a.ChunkOffset.Marshal(b[n:])
	}
	return
}

func (a SampleTable) Len() (n int) {
	n += 8
	if a.SampleDesc != nil {
		n += a.SampleDesc.Len()
	}
	if a.TimeToSample != nil {
		n += a.TimeToSample.Len()
	}
	if a.CompositionOffset != nil {
		n += a.CompositionOffset.Len()
	}
	if a.SampleToChunk != nil {
		n += a.SampleToChunk.Len()
	}
	if a.SyncSample != nil {
		n += a.SyncSample.Len()
	}
	if a.ChunkOffset != nil {
		n += a.ChunkOffset.Len()
	}
	if a.SampleSize != nil {
		n += a.SampleSize.Len()
	}
	return
}

func (a *SampleTable) Unmarshal(b []byte, offset int) (n int, err error) {
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
		case STSD:
			{
				atom := &SampleDesc{}
				if _, err = atom.Unmarshal(b[n:n+size], offset+n); err != nil {
					err = parseErr("stsd", n+offset, err)
					return
				}
				a.SampleDesc = atom
			}
		case STTS:
			{
				atom := &TimeToSample{}
				if _, err = atom.Unmarshal(b[n:n+size], offset+n); err != nil {
					err = parseErr("stts", n+offset, err)
					return
				}
				a.TimeToSample = atom
			}
		case CTTS:
			{
				atom := &CompositionOffset{}
				if _, err = atom.Unmarshal(b[n:n+size], offset+n); err != nil {
					err = parseErr("ctts", n+offset, err)
					return
				}
				a.CompositionOffset = atom
			}
		case STSC:
			{
				atom := &SampleToChunk{}
				if _, err = atom.Unmarshal(b[n:n+size], offset+n); err != nil {
					err = parseErr("stsc", n+offset, err)
					return
				}
				a.SampleToChunk = atom
			}
		case STSS:
			{
				atom := &SyncSample{}
				if _, err = atom.Unmarshal(b[n:n+size], offset+n); err != nil {
					err = parseErr("stss", n+offset, err)
					return
				}
				a.SyncSample = atom
			}
		case STCO:
			{
				atom := &ChunkOffset{}
				if _, err = atom.Unmarshal(b[n:n+size], offset+n); err != nil {
					err = parseErr("stco", n+offset, err)
					return
				}
				a.ChunkOffset = atom
			}
		case STSZ:
			{
				atom := &SampleSize{}
				if _, err = atom.Unmarshal(b[n:n+size], offset+n); err != nil {
					err = parseErr("stsz", n+offset, err)
					return
				}
				a.SampleSize = atom
			}
		}
		n += size
	}
	return
}

func (a SampleTable) Children() (r []Atom) {
	if a.SampleDesc != nil {
		r = append(r, a.SampleDesc)
	}
	if a.TimeToSample != nil {
		r = append(r, a.TimeToSample)
	}
	if a.CompositionOffset != nil {
		r = append(r, a.CompositionOffset)
	}
	if a.SampleToChunk != nil {
		r = append(r, a.SampleToChunk)
	}
	if a.SyncSample != nil {
		r = append(r, a.SyncSample)
	}
	if a.ChunkOffset != nil {
		r = append(r, a.ChunkOffset)
	}
	if a.SampleSize != nil {
		r = append(r, a.SampleSize)
	}
	return
}

func (a SampleTable) Tag() Tag {
	return STBL
}

const STSD = Tag(0x73747364)

type SampleDesc struct {
	Version  uint8
	AVC1Desc *AVC1Desc
	MP4ADesc *MP4ADesc
	OpusDesc *OpusSampleEntry
	Unknowns []Atom
	AtomPos
}

func (a SampleDesc) Marshal(b []byte) (n int) {
	pio.PutU32BE(b[4:], uint32(STSD))
	n += a.marshal(b[8:]) + 8
	pio.PutU32BE(b[0:], uint32(n))
	return
}

func (a SampleDesc) marshal(b []byte) (n int) {
	pio.PutU8(b[n:], a.Version)
	n += 1
	n += 3
	_childrenNR := 0
	if a.AVC1Desc != nil {
		_childrenNR++
	}
	if a.MP4ADesc != nil {
		_childrenNR++
	}
	if a.OpusDesc != nil {
		_childrenNR++
	}
	_childrenNR += len(a.Unknowns)
	pio.PutI32BE(b[n:], int32(_childrenNR))
	n += 4
	if a.AVC1Desc != nil {
		n += a.AVC1Desc.Marshal(b[n:])
	}
	if a.MP4ADesc != nil {
		n += a.MP4ADesc.Marshal(b[n:])
	}
	if a.OpusDesc != nil {
		n += a.OpusDesc.Marshal(b[n:])
	}
	for _, atom := range a.Unknowns {
		n += atom.Marshal(b[n:])
	}
	return
}

func (a SampleDesc) Len() (n int) {
	n += 8
	n += 1
	n += 3
	n += 4
	if a.AVC1Desc != nil {
		n += a.AVC1Desc.Len()
	}
	if a.MP4ADesc != nil {
		n += a.MP4ADesc.Len()
	}
	if a.OpusDesc != nil {
		n += a.OpusDesc.Len()
	}
	for _, atom := range a.Unknowns {
		n += atom.Len()
	}
	return
}

func (a *SampleDesc) Unmarshal(b []byte, offset int) (n int, err error) {
	(&a.AtomPos).setPos(offset, len(b))
	n += 8
	if len(b) < n+1 {
		err = parseErr("Version", n+offset, err)
		return
	}
	a.Version = pio.U8(b[n:])
	n += 1
	n += 3
	n += 4
	for n+8 < len(b) {
		tag := Tag(pio.U32BE(b[n+4:]))
		size := int(pio.U32BE(b[n:]))
		if len(b) < n+size {
			err = parseErr("TagSizeInvalid", n+offset, err)
			return
		}
		switch tag {
		case AVC1:
			{
				atom := &AVC1Desc{}
				if _, err = atom.Unmarshal(b[n:n+size], offset+n); err != nil {
					err = parseErr("avc1", n+offset, err)
					return
				}
				a.AVC1Desc = atom
			}
		case MP4A:
			{
				atom := &MP4ADesc{}
				if _, err = atom.Unmarshal(b[n:n+size], offset+n); err != nil {
					err = parseErr("mp4a", n+offset, err)
					return
				}
				a.MP4ADesc = atom
			}
		case OPUS:
			{
				atom := &OpusSampleEntry{}
				if _, err = atom.Unmarshal(b[n:n+size], offset+n); err != nil {
					err = parseErr("OPUS", n+offset, err)
					return
				}
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

func (a SampleDesc) Children() (r []Atom) {
	if a.AVC1Desc != nil {
		r = append(r, a.AVC1Desc)
	}
	if a.MP4ADesc != nil {
		r = append(r, a.MP4ADesc)
	}
	if a.OpusDesc != nil {
		r = append(r, a.OpusDesc)
	}
	r = append(r, a.Unknowns...)
	return
}

func (a SampleDesc) Tag() Tag {
	return STSD
}

const STTS = Tag(0x73747473)

type TimeToSample struct {
	Version uint8
	Flags   uint32
	Entries []TimeToSampleEntry
	AtomPos
}

func (a TimeToSample) Marshal(b []byte) (n int) {
	pio.PutU32BE(b[4:], uint32(STTS))
	n += a.marshal(b[8:]) + 8
	pio.PutU32BE(b[0:], uint32(n))
	return
}

func (a TimeToSample) marshal(b []byte) (n int) {
	pio.PutU8(b[n:], a.Version)
	n += 1
	pio.PutU24BE(b[n:], a.Flags)
	n += 3
	pio.PutU32BE(b[n:], uint32(len(a.Entries)))
	n += 4
	for _, entry := range a.Entries {
		putTimeToSampleEntry(b[n:], entry)
		n += lenTimeToSampleEntry
	}
	return
}

func (a TimeToSample) Len() (n int) {
	n += 8
	n += 1
	n += 3
	n += 4
	n += lenTimeToSampleEntry * len(a.Entries)
	return
}

func (a *TimeToSample) Unmarshal(b []byte, offset int) (n int, err error) {
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
	var _len_Entries uint32
	_len_Entries = pio.U32BE(b[n:])
	n += 4
	a.Entries = make([]TimeToSampleEntry, _len_Entries)
	if len(b) < n+lenTimeToSampleEntry*len(a.Entries) {
		err = parseErr("TimeToSampleEntry", n+offset, err)
		return
	}
	for i := range a.Entries {
		a.Entries[i] = getTimeToSampleEntry(b[n:])
		n += lenTimeToSampleEntry
	}
	return
}

func (a TimeToSample) Children() (r []Atom) {
	return
}

func (a TimeToSample) String() string {
	return fmt.Sprintf("entries=%d", len(a.Entries))
}

type TimeToSampleEntry struct {
	Count    uint32
	Duration uint32
}

func getTimeToSampleEntry(b []byte) (a TimeToSampleEntry) {
	a.Count = pio.U32BE(b[0:])
	a.Duration = pio.U32BE(b[4:])
	return
}

func putTimeToSampleEntry(b []byte, a TimeToSampleEntry) {
	pio.PutU32BE(b[0:], a.Count)
	pio.PutU32BE(b[4:], a.Duration)
}

const lenTimeToSampleEntry = 8

func (a TimeToSample) Tag() Tag {
	return STTS
}

const STSC = Tag(0x73747363)

type SampleToChunk struct {
	Version uint8
	Flags   uint32
	Entries []SampleToChunkEntry
	AtomPos
}

func (a SampleToChunk) Marshal(b []byte) (n int) {
	pio.PutU32BE(b[4:], uint32(STSC))
	n += a.marshal(b[8:]) + 8
	pio.PutU32BE(b[0:], uint32(n))
	return
}

func (a SampleToChunk) marshal(b []byte) (n int) {
	pio.PutU8(b[n:], a.Version)
	n += 1
	pio.PutU24BE(b[n:], a.Flags)
	n += 3
	pio.PutU32BE(b[n:], uint32(len(a.Entries)))
	n += 4
	for _, entry := range a.Entries {
		putSampleToChunkEntry(b[n:], entry)
		n += lenSampleToChunkEntry
	}
	return
}

func (a SampleToChunk) Len() (n int) {
	n += 8
	n += 1
	n += 3
	n += 4
	n += lenSampleToChunkEntry * len(a.Entries)
	return
}

func (a *SampleToChunk) Unmarshal(b []byte, offset int) (n int, err error) {
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
	var _len_Entries uint32
	_len_Entries = pio.U32BE(b[n:])
	n += 4
	a.Entries = make([]SampleToChunkEntry, _len_Entries)
	if len(b) < n+lenSampleToChunkEntry*len(a.Entries) {
		err = parseErr("SampleToChunkEntry", n+offset, err)
		return
	}
	for i := range a.Entries {
		a.Entries[i] = getSampleToChunkEntry(b[n:])
		n += lenSampleToChunkEntry
	}
	return
}

func (a SampleToChunk) Children() (r []Atom) {
	return
}

func (a SampleToChunk) String() string {
	return fmt.Sprintf("entries=%d", len(a.Entries))
}

type SampleToChunkEntry struct {
	FirstChunk      uint32
	SamplesPerChunk uint32
	SampleDescId    uint32
}

func getSampleToChunkEntry(b []byte) (a SampleToChunkEntry) {
	a.FirstChunk = pio.U32BE(b[0:])
	a.SamplesPerChunk = pio.U32BE(b[4:])
	a.SampleDescId = pio.U32BE(b[8:])
	return
}

func putSampleToChunkEntry(b []byte, a SampleToChunkEntry) {
	pio.PutU32BE(b[0:], a.FirstChunk)
	pio.PutU32BE(b[4:], a.SamplesPerChunk)
	pio.PutU32BE(b[8:], a.SampleDescId)
}

const lenSampleToChunkEntry = 12

func (a SampleToChunk) Tag() Tag {
	return STSC
}

const CTTS = Tag(0x63747473)

type CompositionOffset struct {
	Version uint8
	Flags   uint32
	Entries []CompositionOffsetEntry
	AtomPos
}

func (a CompositionOffset) Marshal(b []byte) (n int) {
	pio.PutU32BE(b[4:], uint32(CTTS))
	n += a.marshal(b[8:]) + 8
	pio.PutU32BE(b[0:], uint32(n))
	return
}

func (a CompositionOffset) marshal(b []byte) (n int) {
	pio.PutU8(b[n:], a.Version)
	n += 1
	pio.PutU24BE(b[n:], a.Flags)
	n += 3
	pio.PutU32BE(b[n:], uint32(len(a.Entries)))
	n += 4
	for _, entry := range a.Entries {
		putCompositionOffsetEntry(b[n:], entry)
		n += lenCompositionOffsetEntry
	}
	return
}

func (a CompositionOffset) Len() (n int) {
	n += 8
	n += 1
	n += 3
	n += 4
	n += lenCompositionOffsetEntry * len(a.Entries)
	return
}

func (a *CompositionOffset) Unmarshal(b []byte, offset int) (n int, err error) {
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
	var _len_Entries uint32
	_len_Entries = pio.U32BE(b[n:])
	n += 4
	a.Entries = make([]CompositionOffsetEntry, _len_Entries)
	if len(b) < n+lenCompositionOffsetEntry*len(a.Entries) {
		err = parseErr("CompositionOffsetEntry", n+offset, err)
		return
	}
	for i := range a.Entries {
		a.Entries[i] = getCompositionOffsetEntry(b[n:])
		n += lenCompositionOffsetEntry
	}
	return
}

func (a CompositionOffset) Children() (r []Atom) {
	return
}

func (a CompositionOffset) String() string {
	return fmt.Sprintf("entries=%d", len(a.Entries))
}

type CompositionOffsetEntry struct {
	Count  uint32
	Offset uint32
}

func getCompositionOffsetEntry(b []byte) (a CompositionOffsetEntry) {
	a.Count = pio.U32BE(b[0:])
	a.Offset = pio.U32BE(b[4:])
	return
}

func putCompositionOffsetEntry(b []byte, a CompositionOffsetEntry) {
	pio.PutU32BE(b[0:], a.Count)
	pio.PutU32BE(b[4:], a.Offset)
}

const lenCompositionOffsetEntry = 8

func (a CompositionOffset) Tag() Tag {
	return CTTS
}

const STSS = Tag(0x73747373)

type SyncSample struct {
	Version uint8
	Flags   uint32
	Entries []uint32
	AtomPos
}

func (a SyncSample) Marshal(b []byte) (n int) {
	pio.PutU32BE(b[4:], uint32(STSS))
	n += a.marshal(b[8:]) + 8
	pio.PutU32BE(b[0:], uint32(n))
	return
}

func (a SyncSample) marshal(b []byte) (n int) {
	pio.PutU8(b[n:], a.Version)
	n += 1
	pio.PutU24BE(b[n:], a.Flags)
	n += 3
	pio.PutU32BE(b[n:], uint32(len(a.Entries)))
	n += 4
	for _, entry := range a.Entries {
		pio.PutU32BE(b[n:], entry)
		n += 4
	}
	return
}

func (a SyncSample) Len() (n int) {
	n += 8
	n += 1
	n += 3
	n += 4
	n += 4 * len(a.Entries)
	return
}

func (a *SyncSample) Unmarshal(b []byte, offset int) (n int, err error) {
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
	var _len_Entries uint32
	_len_Entries = pio.U32BE(b[n:])
	n += 4
	a.Entries = make([]uint32, _len_Entries)
	if len(b) < n+4*len(a.Entries) {
		err = parseErr("uint32", n+offset, err)
		return
	}
	for i := range a.Entries {
		a.Entries[i] = pio.U32BE(b[n:])
		n += 4
	}
	return
}

func (a SyncSample) Children() (r []Atom) {
	return
}

func (a SyncSample) Tag() Tag {
	return STSS
}

func (a SyncSample) String() string {
	return fmt.Sprintf("entries=%d", len(a.Entries))
}

const STCO = Tag(0x7374636f)

type ChunkOffset struct {
	Version uint8
	Flags   uint32
	Entries []uint32
	AtomPos
}

func (a ChunkOffset) Marshal(b []byte) (n int) {
	pio.PutU32BE(b[4:], uint32(STCO))
	n += a.marshal(b[8:]) + 8
	pio.PutU32BE(b[0:], uint32(n))
	return
}

func (a ChunkOffset) marshal(b []byte) (n int) {
	pio.PutU8(b[n:], a.Version)
	n += 1
	pio.PutU24BE(b[n:], a.Flags)
	n += 3
	pio.PutU32BE(b[n:], uint32(len(a.Entries)))
	n += 4
	for _, entry := range a.Entries {
		pio.PutU32BE(b[n:], entry)
		n += 4
	}
	return
}

func (a ChunkOffset) Len() (n int) {
	n += 8
	n += 1
	n += 3
	n += 4
	n += 4 * len(a.Entries)
	return
}

func (a *ChunkOffset) Unmarshal(b []byte, offset int) (n int, err error) {
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
	var _len_Entries uint32
	_len_Entries = pio.U32BE(b[n:])
	n += 4
	a.Entries = make([]uint32, _len_Entries)
	if len(b) < n+4*len(a.Entries) {
		err = parseErr("uint32", n+offset, err)
		return
	}
	for i := range a.Entries {
		a.Entries[i] = pio.U32BE(b[n:])
		n += 4
	}
	return
}

func (a ChunkOffset) Children() (r []Atom) {
	return
}

func (a ChunkOffset) Tag() Tag {
	return STCO
}

func (a ChunkOffset) String() string {
	return fmt.Sprintf("entries=%d", len(a.Entries))
}

const STSZ = Tag(0x7374737a)

type SampleSize struct {
	Version    uint8
	Flags      uint32
	SampleSize uint32
	Entries    []uint32
	AtomPos
}

func (a SampleSize) Marshal(b []byte) (n int) {
	pio.PutU32BE(b[4:], uint32(STSZ))
	n += a.marshal(b[8:]) + 8
	pio.PutU32BE(b[0:], uint32(n))
	return
}

func (a SampleSize) marshal(b []byte) (n int) {
	pio.PutU8(b[n:], a.Version)
	n += 1
	pio.PutU24BE(b[n:], a.Flags)
	n += 3
	pio.PutU32BE(b[n:], a.SampleSize)
	n += 4
	if a.SampleSize != 0 {
		return
	}
	pio.PutU32BE(b[n:], uint32(len(a.Entries)))
	n += 4
	for _, entry := range a.Entries {
		pio.PutU32BE(b[n:], entry)
		n += 4
	}
	return
}

func (a SampleSize) Len() (n int) {
	n += 8
	n += 1
	n += 3
	n += 4
	if a.SampleSize != 0 {
		return
	}
	n += 4
	n += 4 * len(a.Entries)
	return
}

func (a *SampleSize) Unmarshal(b []byte, offset int) (n int, err error) {
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
		err = parseErr("SampleSize", n+offset, err)
		return
	}
	a.SampleSize = pio.U32BE(b[n:])
	n += 4
	if a.SampleSize != 0 {
		return
	}
	var _len_Entries uint32
	_len_Entries = pio.U32BE(b[n:])
	n += 4
	a.Entries = make([]uint32, _len_Entries)
	if len(b) < n+4*len(a.Entries) {
		err = parseErr("uint32", n+offset, err)
		return
	}
	for i := range a.Entries {
		a.Entries[i] = pio.U32BE(b[n:])
		n += 4
	}
	return
}

func (a SampleSize) Children() (r []Atom) {
	return
}

func (a SampleSize) Tag() Tag {
	return STSZ
}

func (a SampleSize) String() string {
	return fmt.Sprintf("entries=%d", len(a.Entries))
}
