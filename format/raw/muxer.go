package raw

import (
	"bytes"
	"io"

	"github.com/deepch/vdk/codec/h265parser"

	"github.com/deepch/vdk/codec/h264parser"

	"github.com/deepch/vdk/av"
)

var startCode = []byte{0, 0, 0, 1}

type Muxer struct {
	idx int8
	w   io.WriteSeeker
}

func NewMuxer(v io.WriteSeeker) *Muxer {
	return &Muxer{w: v}
}
func (element *Muxer) WriteHeader(streams []av.CodecData) (err error) {
	for i, stream := range streams {
		switch stream.Type() {
		case av.H264:
			_, err = element.w.Write(append(startCode, bytes.Join([][]byte{stream.(h264parser.CodecData).SPS(), stream.(h264parser.CodecData).PPS()}, startCode)...))
			element.idx = int8(i)
		case av.H265:
			_, err = element.w.Write(append(startCode, bytes.Join([][]byte{stream.(h265parser.CodecData).SPS(), stream.(h265parser.CodecData).PPS(), stream.(h265parser.CodecData).VPS()}, startCode)...))
			element.idx = int8(i)
		}
	}
	return
}
func (element *Muxer) WritePacket(pkt *av.Packet) (err error) {
	if pkt.Idx == element.idx {
		_, err = element.w.Write(startCode)
		if err != nil {
			return
		}
		_, err = element.w.Write(pkt.Data[4:])
	}
	return
}
