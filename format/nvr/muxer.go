package nvr

import (
	"log"
	"os"
	"time"

	"github.com/deepch/vdk/av"
)

type Muxer struct {
	name      string
	patch     string
	started   bool
	file      *os.File
	codec     []av.CodecData
	buffer    []*av.Packet
	bufferDur time.Duration
	seqDur    time.Duration
}

//NewMuxer func
func NewMuxer(codec []av.CodecData, name, patch string, seqDur time.Duration) *Muxer {
	return &Muxer{
		codec:  codec,
		name:   name,
		patch:  patch,
		seqDur: seqDur,
	}
}

//WritePacket func
func (obj *Muxer) CodecUpdate(val []av.CodecData) {
	obj.codec = val
}

//WritePacket func
func (obj *Muxer) WritePacket(pkt *av.Packet) (err error) {
	if !obj.started && pkt.IsKeyFrame {
		obj.started = true
	}
	if obj.started {
		if pkt.IsKeyFrame && obj.bufferDur >= obj.seqDur {
			log.Println("write to drive", len(obj.buffer), obj.bufferDur)
			obj.buffer = nil
			obj.bufferDur = 0
		}
		obj.buffer = append(obj.buffer, pkt)
		if pkt.Idx == 0 {
			obj.bufferDur += pkt.Duration
		}
	}
	return nil
}

//Close func
func (obj *Muxer) Close() {
	return
}
