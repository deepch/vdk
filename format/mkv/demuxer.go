package mkv

import (
	"encoding/binary"
	"errors"
	"io"
	"time"

	"github.com/deepch/vdk/av"
	"github.com/deepch/vdk/codec/h264parser"
	"github.com/deepch/vdk/format/mkv/mkvio"
)

type Demuxer struct {
	r       *mkvio.Document
	pkts    []av.Packet
	sps     []byte
	pps     []byte
	streams []*Stream
	ps      uint32
	stage   int
	fc      int
	ls      time.Duration
}

func NewDemuxer(r io.Reader) *Demuxer {

	return &Demuxer{
		r: mkvio.InitDocument(r),
	}

}

func (self *Demuxer) Streams() (streams []av.CodecData, err error) {

	if err = self.probe(); err != nil {
		return
	}

	for _, stream := range self.streams {
		streams = append(streams, stream.CodecData)
	}

	if len(streams) == 0 {
		return nil, errors.New("streams not found")
	}

	return streams, err

}

func (self *Demuxer) probe() (err error) {
	if self.stage == 0 {

		var el *mkvio.Element
		el, err = self.r.GetVideoCodec()
		if err != nil {
			return
		}

		if el.ElementRegister.ID == mkvio.ElementCodecPrivate.ID {
			payload := el.Content[6:]
			var reader int
			for pos := 0; pos < len(payload); pos = reader {
				lens := int(binary.BigEndian.Uint16(payload[reader:]))
				reader += 2
				nal := payload[reader : reader+lens]
				naluType := nal[0] & 0x1f
				switch naluType {
				case h264parser.NALU_SPS:
					self.sps = nal
				case h264parser.NALU_PPS:
					self.pps = nal
				}
				reader += lens
				reader++
			}
		}

		if len(self.sps) > 0 && len(self.pps) > 0 {
			var codec av.CodecData
			codec, err = h264parser.NewCodecDataFromSPSAndPPS(self.sps, self.pps)

			if err != nil {
				return
			}

			stream := &Stream{}
			stream.idx = 0
			stream.demuxer = self
			stream.CodecData = codec
			self.streams = append(self.streams, stream)

		}
		self.stage++
	}
	return
}

func (self *Demuxer) ReadPacket() (pkt av.Packet, err error) {

	var el mkvio.Element

	for {
		el, err = self.r.ParseElement()
		if err != nil {
			return
		}

		if el.Type == 6 && el.ElementRegister.ID == mkvio.ElementSimpleBlock.ID {
			self.fc++
			nals, _ := h264parser.SplitNALUs(el.Content[4:])
			for _, nal := range nals {

				naluType := nal[0] & 0x1f

				if naluType == 5 {
					l1 := int(binary.BigEndian.Uint16(el.Content[2:4]))
					dur := time.Duration(uint32(l1)) * time.Millisecond
					self.ls += time.Duration(uint32(l1)) * time.Millisecond
					self.ps = 0
					pkt = av.Packet{IsKeyFrame: true, Idx: 0, Duration: dur, Time: self.ls, Data: append(binSize(len(nal)), nal...)}
					return

				} else if naluType == 1 {

					l1 := int(binary.BigEndian.Uint16(el.Content[1:3]))
					dur := time.Duration(uint32(l1)-self.ps) * time.Millisecond
					self.ls += time.Duration(uint32(l1)-self.ps) * time.Millisecond
					self.ps = uint32(l1)
					pkt = av.Packet{Idx: 0, Duration: dur, Time: self.ls, Data: append(binSize(len(nal)), nal...)}
					return

				}
			}
		}
	}

	return

}

func binSize(val int) []byte {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(val))
	return buf
}
