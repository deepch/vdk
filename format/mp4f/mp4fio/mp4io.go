package mp4fio

import (
	"github.com/deepch/vdk/format/mp4/mp4io"
	"github.com/deepch/vdk/utils/bits/pio"
)

type ElemStreamDesc struct {
	DecConfig []byte
	TrackId   uint16
	mp4io.AtomPos
}

func (self ElemStreamDesc) Children() []mp4io.Atom {
	return nil
}

func (self ElemStreamDesc) fillLength(b []byte, length int) (n int) {
	b[n] = uint8(length & 0x7f)
	n++
	return
}

func (self ElemStreamDesc) lenDescHdr() (n int) {
	return 2
}

func (self ElemStreamDesc) fillDescHdr(b []byte, tag uint8, datalen int) (n int) {
	b[n] = tag
	n++
	n += self.fillLength(b[n:], datalen)
	return
}

func (self ElemStreamDesc) lenESDescHdr() (n int) {
	return self.lenDescHdr() + 3
}

func (self ElemStreamDesc) fillESDescHdr(b []byte, datalen int) (n int) {
	n += self.fillDescHdr(b[n:], mp4io.MP4ESDescrTag, datalen)
	pio.PutU16BE(b[n:], self.TrackId)
	n += 2
	b[n] = 0 // flags
	n++
	return
}

func (self ElemStreamDesc) lenDecConfigDescHdr() (n int) {
	return self.lenDescHdr() + 2 + 3 + 4 + 4 + self.lenDescHdr()
}

func (self ElemStreamDesc) fillDecConfigDescHdr(b []byte, datalen int) (n int) {
	n += self.fillDescHdr(b[n:], mp4io.MP4DecConfigDescrTag, datalen)
	b[n] = 0x40 // objectid
	n++
	b[n] = 0x15 // streamtype
	n++
	// buffer size db
	pio.PutU24BE(b[n:], 0)
	n += 3
	// max bitrage
	pio.PutU32BE(b[n:], uint32(200000))
	n += 4
	// avg bitrage
	pio.PutU32BE(b[n:], uint32(0))
	n += 4
	n += self.fillDescHdr(b[n:], mp4io.MP4DecSpecificDescrTag, datalen-n)
	return
}

func (self ElemStreamDesc) Len() (n int) {
	// len + tag
	return 8 +
		// ver + flags
		4 +
		self.lenESDescHdr() +
		self.lenDecConfigDescHdr() +
		len(self.DecConfig) +
		self.lenDescHdr() + 1
}

// Version(4)
// ESDesc(
//   MP4ESDescrTag
//   ESID(2)
//   ESFlags(1)
//   DecConfigDesc(
//     MP4DecConfigDescrTag
//     objectId streamType bufSize avgBitrate
//     DecSpecificDesc(
//       MP4DecSpecificDescrTag
//       decConfig
//     )
//   )
//   ?Desc(lenDescHdr+1)
// )

func (self ElemStreamDesc) Marshal(b []byte) (n int) {
	pio.PutU32BE(b[4:], uint32(mp4io.ESDS))
	n += 8
	pio.PutU32BE(b[n:], 0) // Version
	n += 4
	datalen := self.Len()
	n += self.fillESDescHdr(b[n:], datalen-n-self.lenESDescHdr()+3)
	n += self.fillDecConfigDescHdr(b[n:], datalen-n-self.lenDescHdr()-3)
	copy(b[n:], self.DecConfig)
	n += len(self.DecConfig)
	n += self.fillDescHdr(b[n:], 0x06, datalen-n-self.lenDescHdr())
	b[n] = 0x02
	n++
	pio.PutU32BE(b[0:], uint32(n))
	return
}
