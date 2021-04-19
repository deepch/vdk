package fmp4io

import "github.com/deepch/vdk/utils/bits/pio"

const AVC1 = Tag(0x61766331)

type AVC1Desc struct {
	DataRefIdx           int16
	Version              int16
	Revision             int16
	Vendor               int32
	TemporalQuality      int32
	SpatialQuality       int32
	Width                int16
	Height               int16
	HorizontalResolution float64
	VorizontalResolution float64
	FrameCount           int16
	CompressorName       [32]byte
	Depth                int16
	ColorTableId         int16
	Conf                 *AVC1Conf
	PixelAspect          *PixelAspect
	Unknowns             []Atom
	AtomPos
}

func (a AVC1Desc) Tag() Tag {
	return AVC1
}

func (a AVC1Desc) Marshal(b []byte) (n int) {
	pio.PutU32BE(b[4:], uint32(AVC1))
	n += a.marshal(b[8:]) + 8
	pio.PutU32BE(b[0:], uint32(n))
	return
}

func (a AVC1Desc) marshal(b []byte) (n int) {
	n += 6
	pio.PutI16BE(b[n:], a.DataRefIdx)
	n += 2
	pio.PutI16BE(b[n:], a.Version)
	n += 2
	pio.PutI16BE(b[n:], a.Revision)
	n += 2
	pio.PutI32BE(b[n:], a.Vendor)
	n += 4
	pio.PutI32BE(b[n:], a.TemporalQuality)
	n += 4
	pio.PutI32BE(b[n:], a.SpatialQuality)
	n += 4
	pio.PutI16BE(b[n:], a.Width)
	n += 2
	pio.PutI16BE(b[n:], a.Height)
	n += 2
	PutFixed32(b[n:], a.HorizontalResolution)
	n += 4
	PutFixed32(b[n:], a.VorizontalResolution)
	n += 4
	n += 4
	pio.PutI16BE(b[n:], a.FrameCount)
	n += 2
	copy(b[n:], a.CompressorName[:])
	n += len(a.CompressorName[:])
	pio.PutI16BE(b[n:], a.Depth)
	n += 2
	pio.PutI16BE(b[n:], a.ColorTableId)
	n += 2
	if a.Conf != nil {
		n += a.Conf.Marshal(b[n:])
	}
	if a.PixelAspect != nil {
		n += a.PixelAspect.Marshal(b[n:])
	}
	for _, atom := range a.Unknowns {
		n += atom.Marshal(b[n:])
	}
	return
}

func (a AVC1Desc) Len() (n int) {
	n += 8
	n += 6
	n += 2
	n += 2
	n += 2
	n += 4
	n += 4
	n += 4
	n += 2
	n += 2
	n += 4
	n += 4
	n += 4
	n += 2
	n += len(a.CompressorName[:])
	n += 2
	n += 2
	if a.Conf != nil {
		n += a.Conf.Len()
	}
	if a.PixelAspect != nil {
		n += a.PixelAspect.Len()
	}
	for _, atom := range a.Unknowns {
		n += atom.Len()
	}
	return
}

func (a *AVC1Desc) Unmarshal(b []byte, offset int) (n int, err error) {
	a.AtomPos.setPos(offset, len(b))
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
		err = parseErr("Revision", n+offset, err)
		return
	}
	a.Revision = pio.I16BE(b[n:])
	n += 2
	if len(b) < n+4 {
		err = parseErr("Vendor", n+offset, err)
		return
	}
	a.Vendor = pio.I32BE(b[n:])
	n += 4
	if len(b) < n+4 {
		err = parseErr("TemporalQuality", n+offset, err)
		return
	}
	a.TemporalQuality = pio.I32BE(b[n:])
	n += 4
	if len(b) < n+4 {
		err = parseErr("SpatialQuality", n+offset, err)
		return
	}
	a.SpatialQuality = pio.I32BE(b[n:])
	n += 4
	if len(b) < n+2 {
		err = parseErr("Width", n+offset, err)
		return
	}
	a.Width = pio.I16BE(b[n:])
	n += 2
	if len(b) < n+2 {
		err = parseErr("Height", n+offset, err)
		return
	}
	a.Height = pio.I16BE(b[n:])
	n += 2
	if len(b) < n+4 {
		err = parseErr("HorizontalResolution", n+offset, err)
		return
	}
	a.HorizontalResolution = GetFixed32(b[n:])
	n += 4
	if len(b) < n+4 {
		err = parseErr("VorizontalResolution", n+offset, err)
		return
	}
	a.VorizontalResolution = GetFixed32(b[n:])
	n += 4
	n += 4
	if len(b) < n+2 {
		err = parseErr("FrameCount", n+offset, err)
		return
	}
	a.FrameCount = pio.I16BE(b[n:])
	n += 2
	if len(b) < n+len(a.CompressorName) {
		err = parseErr("CompressorName", n+offset, err)
		return
	}
	copy(a.CompressorName[:], b[n:])
	n += len(a.CompressorName)
	if len(b) < n+2 {
		err = parseErr("Depth", n+offset, err)
		return
	}
	a.Depth = pio.I16BE(b[n:])
	n += 2
	if len(b) < n+2 {
		err = parseErr("ColorTableId", n+offset, err)
		return
	}
	a.ColorTableId = pio.I16BE(b[n:])
	n += 2
	for n+8 < len(b) {
		tag := Tag(pio.U32BE(b[n+4:]))
		size := int(pio.U32BE(b[n:]))
		if len(b) < n+size {
			err = parseErr("TagSizeInvalid", n+offset, err)
			return
		}
		switch tag {
		case AVCC:
			{
				atom := &AVC1Conf{}
				if _, err = atom.Unmarshal(b[n:n+size], offset+n); err != nil {
					err = parseErr("avcC", n+offset, err)
					return
				}
				a.Conf = atom
			}
		case PASP:
			{
				atom := &PixelAspect{}
				if _, err = atom.Unmarshal(b[n:n+size], offset+n); err != nil {
					err = parseErr("pasp", n+offset, err)
					return
				}
				a.PixelAspect = atom
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

func (a AVC1Desc) Children() (r []Atom) {
	if a.Conf != nil {
		r = append(r, a.Conf)
	}
	if a.PixelAspect != nil {
		r = append(r, a.PixelAspect)
	}
	r = append(r, a.Unknowns...)
	return
}

const AVCC = Tag(0x61766343)

type AVC1Conf struct {
	Data []byte
	AtomPos
}

func (a AVC1Conf) Tag() Tag {
	return AVCC
}

func (a AVC1Conf) Marshal(b []byte) (n int) {
	pio.PutU32BE(b[4:], uint32(AVCC))
	n += a.marshal(b[8:]) + 8
	pio.PutU32BE(b[0:], uint32(n))
	return
}

func (a AVC1Conf) marshal(b []byte) (n int) {
	copy(b[n:], a.Data[:])
	n += len(a.Data[:])
	return
}

func (a AVC1Conf) Len() (n int) {
	n += 8
	n += len(a.Data[:])
	return
}

func (a *AVC1Conf) Unmarshal(b []byte, offset int) (n int, err error) {
	a.AtomPos.setPos(offset, len(b))
	n += 8
	a.Data = b[n:]
	n += len(b[n:])
	return
}

func (a AVC1Conf) Children() (r []Atom) {
	return
}

const PASP = Tag(0x70617370)

type PixelAspect struct {
	HorizontalSpacing uint32
	VerticalSpacing   uint32
	AtomPos
}

func (a PixelAspect) Tag() Tag {
	return PASP
}

func (a PixelAspect) Marshal(b []byte) (n int) {
	pio.PutU32BE(b[4:], uint32(PASP))
	n += a.marshal(b[8:]) + 8
	pio.PutU32BE(b[0:], uint32(n))
	return
}

func (a PixelAspect) marshal(b []byte) (n int) {
	pio.PutU32BE(b[n:], a.HorizontalSpacing)
	n += 4
	pio.PutU32BE(b[n:], a.VerticalSpacing)
	n += 4
	return
}

func (a PixelAspect) Len() (n int) {
	return 8 + 8
}

func (a *PixelAspect) Unmarshal(b []byte, offset int) (n int, err error) {
	a.AtomPos.setPos(offset, len(b))
	n += 8
	if len(b) < n+4 {
		err = parseErr("HorizontalSpacing", n+offset, err)
		return
	}
	a.HorizontalSpacing = pio.U32BE(b[n:])
	n += 4
	if len(b) < n+4 {
		err = parseErr("VerticalSpacing", n+offset, err)
		return
	}
	a.VerticalSpacing = pio.U32BE(b[n:])
	n += 4
	return
}

func (a *PixelAspect) Children() (r []Atom) {
	return nil
}
