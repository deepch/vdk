package fmp4io

import (
	"bytes"

	"github.com/deepch/vdk/utils/bits/pio"
)

const DREF = Tag(0x64726566)

type DataRefer struct {
	Version uint8
	Flags   uint32
	Url     *DataReferUrl
	AtomPos
}

func (a DataRefer) Tag() Tag {
	return DREF
}

func (a DataRefer) Marshal(b []byte) (n int) {
	pio.PutU32BE(b[4:], uint32(DREF))
	n += a.marshal(b[8:]) + 8
	pio.PutU32BE(b[0:], uint32(n))
	return
}

func (a DataRefer) marshal(b []byte) (n int) {
	pio.PutU8(b[n:], a.Version)
	n += 1
	pio.PutU24BE(b[n:], a.Flags)
	n += 3
	_childrenNR := 0
	if a.Url != nil {
		_childrenNR++
	}
	pio.PutI32BE(b[n:], int32(_childrenNR))
	n += 4
	if a.Url != nil {
		n += a.Url.Marshal(b[n:])
	}
	return
}

func (a DataRefer) Len() (n int) {
	n += 8
	n += 1
	n += 3
	n += 4
	if a.Url != nil {
		n += a.Url.Len()
	}
	return
}

func (a *DataRefer) Unmarshal(b []byte, offset int) (n int, err error) {
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
	n += 4
	for n+8 < len(b) {
		tag := Tag(pio.U32BE(b[n+4:]))
		size := int(pio.U32BE(b[n:]))
		if len(b) < n+size {
			err = parseErr("TagSizeInvalid", n+offset, err)
			return
		}
		switch tag {
		case URL:
			{
				atom := &DataReferUrl{}
				if _, err = atom.Unmarshal(b[n:n+size], offset+n); err != nil {
					err = parseErr("url ", n+offset, err)
					return
				}
				a.Url = atom
			}
		}
		n += size
	}
	return
}

func (a DataRefer) Children() (r []Atom) {
	if a.Url != nil {
		r = append(r, a.Url)
	}
	return
}

const URL = Tag(0x75726c20)

type DataReferUrl struct {
	Version uint8
	Flags   uint32
	AtomPos
}

func (a DataReferUrl) Tag() Tag {
	return URL
}

func (a DataReferUrl) Marshal(b []byte) (n int) {
	pio.PutU32BE(b[4:], uint32(URL))
	n += a.marshal(b[8:]) + 8
	pio.PutU32BE(b[0:], uint32(n))
	return
}

func (a DataReferUrl) marshal(b []byte) (n int) {
	pio.PutU8(b[n:], a.Version)
	n += 1
	pio.PutU24BE(b[n:], a.Flags)
	n += 3
	return
}

func (a DataReferUrl) Len() (n int) {
	n += 8
	n += 1
	n += 3
	return
}

func (a *DataReferUrl) Unmarshal(b []byte, offset int) (n int, err error) {
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
	return
}

func (a DataReferUrl) Children() (r []Atom) {
	return
}

const HDLR = Tag(0x68646c72)

type HandlerRefer struct {
	Version    uint8
	Flags      uint32
	Predefined uint32
	Type       uint32
	Reserved   [3]uint32
	Name       string
	AtomPos
}

const (
	VideoHandler = 0x76696465 // vide
	SoundHandler = 0x736f756e // soun
)

func (a HandlerRefer) Marshal(b []byte) (n int) {
	pio.PutU32BE(b[4:], uint32(HDLR))
	n += a.marshal(b[8:]) + 8
	pio.PutU32BE(b[0:], uint32(n))
	return
}

func (a HandlerRefer) marshal(b []byte) (n int) {
	pio.PutU8(b[n:], a.Version)
	n += 1
	pio.PutU24BE(b[n:], a.Flags)
	n += 3
	pio.PutU32BE(b[n:], a.Predefined)
	n += 4
	pio.PutU32BE(b[n:], a.Type)
	n += 4
	n += 3 * 4
	copy(b[n:], a.Name)
	n += len(a.Name) + 1
	return
}

func (a HandlerRefer) Len() (n int) {
	n += 8
	n += 1
	n += 3
	n += 4
	n += 4
	n += 3 * 4
	n += len(a.Name) + 1
	return
}

func (a *HandlerRefer) Unmarshal(b []byte, offset int) (n int, err error) {
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
		err = parseErr("Predefined", n+offset, err)
		return
	}
	a.Predefined = pio.U32BE(b[n:])
	n += 4
	if len(b) < n+4 {
		err = parseErr("Type", n+offset, err)
		return
	}
	a.Type = pio.U32BE(b[n:])
	n += 4
	n += 3 * 4
	i := bytes.IndexByte(b[n:], 0)
	if i > 0 {
		a.Name = string(b[n : n+i])
		n += i + 1
	}
	return
}

func (a HandlerRefer) Children() (r []Atom) {
	return
}

func (a HandlerRefer) Tag() Tag {
	return HDLR
}
