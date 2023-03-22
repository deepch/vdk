package main

import (
	"context"
	"github.com/deepch/vdk/format/rtspv2"
	"github.com/deepch/vdk/format/ts"
	"log"
	"os/exec"
	"time"
)

func main() {
	RTSPClient, err := rtspv2.Dial(rtspv2.RTSPClientOptions{URL: "rtsp://url", DisableAudio: true, DialTimeout: 3 * time.Second, ReadWriteTimeout: 5 * time.Second, Debug: true, OutgoingProxy: false})
	if err != nil {
		panic(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cmd := exec.CommandContext(ctx, "ffmpeg", "-flags", "low_delay", "-analyzeduration", "1", "-fflags", "-nobuffer", "-probesize", "1024k", "-f", "mpegts", "-i", "-", "-vcodec", "libx264", "-preset", "ultrafast", "-bf", "0", "-f", "mpegts", "-max_muxing_queue_size", "400", "-pes_payload_size", "0", "pipe:1")
	inPipe, _ := cmd.StdinPipe()
	outPipe, _ := cmd.StdoutPipe()
	//cmd.Stderr = os.Stderr
	mux := ts.NewMuxer(inPipe)
	demuxer := ts.NewDemuxer(outPipe)
	codec := RTSPClient.CodecData
	mux.WriteHeader(codec)
	go func() {
		imNewCodec, err := demuxer.Streams()
		log.Println("new codec data", imNewCodec, err)
		for i, data := range imNewCodec {
			log.Println(i, data)
		}
		for {
			pkt, err := demuxer.ReadPacket()
			if err != nil {
				log.Panic(err)
			}
			log.Println("im new pkt ===>", pkt.Idx, pkt.Time)
		}
	}()
	cmd.Start()
	var start bool
	for {
		select {
		case signals := <-RTSPClient.Signals:
			switch signals {
			case rtspv2.SignalCodecUpdate:
				//?
			case rtspv2.SignalStreamRTPStop:
				return
			}
		case packetAV := <-RTSPClient.OutgoingPacketQueue:
			if packetAV.IsKeyFrame {
				start = true
			}
			if !start {
				continue
			}
			if err = mux.WritePacket(*packetAV); err != nil {
				return
			}
		}
	}
}
