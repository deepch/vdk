package mp4fio

import (
	"github.com/deepch/vdk/format/mp4/mp4io"
	"github.com/deepch/vdk/utils/bits/pio"
)

func (self MovieFrag) Tag() mp4io.Tag {
	return mp4io.MOOF
}

type MovieFrag struct {
	Header   *MovieFragHeader
	Tracks   []*TrackFrag
	Unknowns []mp4io.Atom
	mp4io.AtomPos
}

func (self MovieFrag) Marshal(b []byte) (n int) {
	pio.PutU32BE(b[4:], uint32(mp4io.MOOF))
	n += self.marshal(b[8:]) + 8
	pio.PutU32BE(b[0:], uint32(n))
	return
}
func (self MovieFrag) marshal(b []byte) (n int) {
	if self.Header != nil {
		n += self.Header.Marshal(b[n:])
	}
	for _, atom := range self.Tracks {
		n += atom.Marshal(b[n:])
	}
	for _, atom := range self.Unknowns {
		n += atom.Marshal(b[n:])
	}
	return
}
func (self MovieFrag) Len() (n int) {
	n += 8
	if self.Header != nil {
		n += self.Header.Len()
	}
	for _, atom := range self.Tracks {
		n += atom.Len()
	}
	for _, atom := range self.Unknowns {
		n += atom.Len()
	}
	return
}
func (self *MovieFrag) Unmarshal(b []byte, offset int) (n int, err error) {

	return
}
func (self MovieFrag) Children() (r []mp4io.Atom) {
	if self.Header != nil {
		r = append(r, self.Header)
	}
	for _, atom := range self.Tracks {
		r = append(r, atom)
	}
	r = append(r, self.Unknowns...)
	return
}

func (self MovieFragHeader) Tag() mp4io.Tag {
	return mp4io.MFHD
}

type MovieFragHeader struct {
	Version uint8
	Flags   uint32
	Seqnum  uint32
	mp4io.AtomPos
}

func (self MovieFragHeader) Marshal(b []byte) (n int) {
	pio.PutU32BE(b[4:], uint32(mp4io.MFHD))
	n += self.marshal(b[8:]) + 8
	pio.PutU32BE(b[0:], uint32(n))
	return
}
func (self MovieFragHeader) marshal(b []byte) (n int) {
	pio.PutU8(b[n:], self.Version)
	n += 1
	pio.PutU24BE(b[n:], self.Flags)
	n += 3
	pio.PutU32BE(b[n:], self.Seqnum)
	n += 4
	return
}
func (self MovieFragHeader) Len() (n int) {
	n += 8
	n += 1
	n += 3
	n += 4
	return
}
func (self *MovieFragHeader) Unmarshal(b []byte, offset int) (n int, err error) {

	return
}
func (self MovieFragHeader) Children() (r []mp4io.Atom) {
	return
}

func (self TrackFragRun) Tag() mp4io.Tag {
	return mp4io.TRUN
}

type TrackFragRun struct {
	Version          uint8
	Flags            uint32
	DataOffset       uint32
	FirstSampleFlags uint32
	Entries          []mp4io.TrackFragRunEntry
	mp4io.AtomPos
}

func (self TrackFragRun) Marshal(b []byte) (n int) {
	pio.PutU32BE(b[4:], uint32(mp4io.TRUN))
	n += self.marshal(b[8:]) + 8
	pio.PutU32BE(b[0:], uint32(n))
	return
}
func (self TrackFragRun) marshal(b []byte) (n int) {
	pio.PutU8(b[n:], self.Version)
	n += 1
	pio.PutU24BE(b[n:], self.Flags)
	n += 3
	pio.PutU32BE(b[n:], uint32(len(self.Entries)))
	n += 4
	if self.Flags&mp4io.TRUN_DATA_OFFSET != 0 {
		{
			pio.PutU32BE(b[n:], self.DataOffset)
			n += 4
		}
	}
	if self.Flags&mp4io.TRUN_FIRST_SAMPLE_FLAGS != 0 {
		{
			pio.PutU32BE(b[n:], self.FirstSampleFlags)
			n += 4
		}
	}

	for i, entry := range self.Entries {
		var flags uint32
		if i > 0 {
			flags = self.Flags
		} else {
			flags = self.FirstSampleFlags
		}
		//if flags&TRUN_SAMPLE_DURATION != 0 {
		pio.PutU32BE(b[n:], entry.Duration)
		n += 4
		//}
		//if flags&TRUN_SAMPLE_SIZE != 0 {
		pio.PutU32BE(b[n:], entry.Size)
		n += 4
		//}
		if flags&mp4io.TRUN_SAMPLE_FLAGS != 0 {
			pio.PutU32BE(b[n:], entry.Flags)
			n += 4
		}
		//if flags&TRUN_SAMPLE_CTS != 0 {
		pio.PutU32BE(b[n:], entry.Cts)
		n += 4
		//}
	}
	return
}
func (self TrackFragRun) Len() (n int) {
	n += 8
	n += 1
	n += 3
	n += 4
	if self.Flags&mp4io.TRUN_DATA_OFFSET != 0 {
		{
			n += 4
		}
	}
	if self.Flags&mp4io.TRUN_FIRST_SAMPLE_FLAGS != 0 {
		{
			n += 4
		}
	}

	for i := range self.Entries {
		var flags uint32
		if i > 0 {
			flags = self.Flags
		} else {
			flags = self.FirstSampleFlags
		}
		//if flags&TRUN_SAMPLE_DURATION != 0 {
		n += 4
		//}
		//if flags&TRUN_SAMPLE_SIZE != 0 {
		n += 4
		//}
		if flags&mp4io.TRUN_SAMPLE_FLAGS != 0 {
			n += 4
		}
		//if flags&TRUN_SAMPLE_CTS != 0 {
		n += 4
		//}
	}
	return
}
func (self *TrackFragRun) Unmarshal(b []byte, offset int) (n int, err error) {

	return
}
func (self TrackFragRun) Children() (r []mp4io.Atom) {
	return
}

func (self TrackFrag) Tag() mp4io.Tag {
	return mp4io.TRAF
}

type TrackFrag struct {
	Header     *TrackFragHeader
	DecodeTime *TrackFragDecodeTime
	Run        *TrackFragRun
	Unknowns   []mp4io.Atom
	mp4io.AtomPos
}

func (self TrackFrag) Marshal(b []byte) (n int) {
	pio.PutU32BE(b[4:], uint32(mp4io.TRAF))
	n += self.marshal(b[8:]) + 8
	pio.PutU32BE(b[0:], uint32(n))
	return
}
func (self TrackFrag) marshal(b []byte) (n int) {
	if self.Header != nil {
		n += self.Header.Marshal(b[n:])
	}
	if self.DecodeTime != nil {
		n += self.DecodeTime.Marshal(b[n:])
	}
	if self.Run != nil {
		n += self.Run.Marshal(b[n:])
	}
	for _, atom := range self.Unknowns {
		n += atom.Marshal(b[n:])
	}
	return
}
func (self TrackFrag) Len() (n int) {
	n += 8
	if self.Header != nil {
		n += self.Header.Len()
	}
	if self.DecodeTime != nil {
		n += self.DecodeTime.Len()
	}
	if self.Run != nil {
		n += self.Run.Len()
	}
	for _, atom := range self.Unknowns {
		n += atom.Len()
	}
	return
}
func (self *TrackFrag) Unmarshal(b []byte, offset int) (n int, err error) {

	return
}
func (self TrackFrag) Children() (r []mp4io.Atom) {
	if self.Header != nil {
		r = append(r, self.Header)
	}
	if self.DecodeTime != nil {
		r = append(r, self.DecodeTime)
	}
	if self.Run != nil {
		r = append(r, self.Run)
	}
	r = append(r, self.Unknowns...)
	return
}

const LenTrackFragRunEntry = 16

func (self TrackFragHeader) Tag() mp4io.Tag {
	return mp4io.TFHD
}

type TrackFragHeader struct {
	Data []byte
	mp4io.AtomPos
}

func (self TrackFragHeader) Marshal(b []byte) (n int) {
	pio.PutU32BE(b[4:], uint32(mp4io.TFHD))
	n += self.marshal(b[8:]) + 8
	pio.PutU32BE(b[0:], uint32(n))
	return
}
func (self TrackFragHeader) marshal(b []byte) (n int) {
	copy(b, self.Data)
	n += len(self.Data)
	return
}
func (self TrackFragHeader) Len() (n int) {
	return len(self.Data) + 8
}
func (self *TrackFragHeader) Unmarshal(b []byte, offset int) (n int, err error) {

	return
}
func (self TrackFragHeader) Children() (r []mp4io.Atom) {
	return
}

func (self TrackFragDecodeTime) Tag() mp4io.Tag {
	return mp4io.TFDT
}

type TrackFragDecodeTime struct {
	Version uint8
	Flags   uint32
	Time    uint64
	mp4io.AtomPos
}

func (self TrackFragDecodeTime) Marshal(b []byte) (n int) {
	pio.PutU32BE(b[4:], uint32(mp4io.TFDT))
	n += self.marshal(b[8:]) + 8
	pio.PutU32BE(b[0:], uint32(n))
	return
}
func (self TrackFragDecodeTime) marshal(b []byte) (n int) {
	pio.PutU8(b[n:], self.Version)
	n += 1
	pio.PutU24BE(b[n:], self.Flags)
	n += 3
	pio.PutU64BE(b[n:], self.Time)
	n += 8
	return
}
func (self TrackFragDecodeTime) Len() (n int) {
	n += 8
	n += 1
	n += 3
	n += 8
	return
}
func (self *TrackFragDecodeTime) Unmarshal(b []byte, offset int) (n int, err error) {

	return
}
func (self TrackFragDecodeTime) Children() (r []mp4io.Atom) {
	return
}
