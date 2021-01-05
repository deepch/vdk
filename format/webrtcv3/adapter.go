package webrtc

import (
	"bytes"
	"encoding/base64"
	"errors"
	"time"

	"github.com/pion/webrtc/v3"

	"github.com/deepch/vdk/av"
	"github.com/deepch/vdk/codec/h264parser"
	"github.com/pion/webrtc/v3/pkg/media"
)

var (
	ErrorNotFound          = errors.New("stream not found")
	ErrorCodecNotSupported = errors.New("codec not supported")
	ErrorClientOffline     = errors.New("client offline")
	Label                  = "track_"
)

type Muxer struct {
	streams map[int8]*Stream
	status  webrtc.ICEConnectionState
	stop    bool
	pc      *webrtc.PeerConnection
	pt      *time.Timer
	ps      chan bool
}
type Stream struct {
	codec av.CodecData
	ts    time.Duration
	track *webrtc.TrackLocalStaticSample
}

func NewMuxer() *Muxer {
	tmp := Muxer{ps: make(chan bool, 100), pt: time.NewTimer(time.Second * 20), streams: make(map[int8]*Stream)}
	go tmp.WaitCloser()
	return &tmp
}

func (element *Muxer) WriteHeader(streams []av.CodecData, sdp64 string) (string, error) {
	if len(streams) == 0 {
		return "", ErrorNotFound
	}
	sdpB, err := base64.StdEncoding.DecodeString(sdp64)
	if err != nil {
		return "", err
	}
	offer := webrtc.SessionDescription{
		Type: webrtc.SDPTypeOffer,
		SDP:  string(sdpB),
	}
	peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		return "", err
	}
	for i, i2 := range streams {
		var track *webrtc.TrackLocalStaticSample
		if i2.Type().IsVideo() {
			if i2.Type() != av.H264 {
				return "", errors.New("Video Not h264 codec not supported")
			}
			track, err = webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{
				MimeType: "video/h264",
			}, "pion-rtsp-video", "pion-rtsp-video")
			if err != nil {
				return "", err
			}
			if _, err = peerConnection.AddTrack(track); err != nil {
				return "", err
			}
		} else if i2.Type().IsAudio() {
			AudioCodecString := "audio/PCMU"
			switch i2.Type() {
			case av.PCM_ALAW:
				AudioCodecString = "audio/PCMA"
			case av.PCM_MULAW:
				AudioCodecString = "audio/PCMU"
			default:
				return "", errors.New("No Audio Codec Supported")
				continue
			}
			//log.Fatalln(i2.Type())
			track, err = webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{
				MimeType: AudioCodecString,
			}, "pion-rtsp-audio", "pion-rtsp-audio")
			if err != nil {
				return "", err
			}
			if _, err = peerConnection.AddTrack(track); err != nil {
				return "", err
			}
		}
		element.streams[int8(i)] = &Stream{track: track, codec: i2}
	}
	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		element.status = connectionState
		if connectionState == webrtc.ICEConnectionStateDisconnected {
			element.ps <- true
		}
	})
	peerConnection.OnDataChannel(func(d *webrtc.DataChannel) {
		d.OnMessage(func(msg webrtc.DataChannelMessage) {
			element.pt.Reset(5 * time.Second)
		})
	})

	if err = peerConnection.SetRemoteDescription(offer); err != nil {
		return "", err
	}
	gatherCompletePromise := webrtc.GatheringCompletePromise(peerConnection)
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		return "", err
	}
	if err = peerConnection.SetLocalDescription(answer); err != nil {
		return "", err
	}
	element.pc = peerConnection
	waitT := time.NewTimer(time.Second * 20)
	select {
	case <-waitT.C:
		return "", errors.New("gatherCompletePromise wait")
	case <-gatherCompletePromise:
	}
	resp := peerConnection.LocalDescription()
	return base64.StdEncoding.EncodeToString([]byte(resp.SDP)), nil

}

func (element *Muxer) WritePacket(pkt av.Packet) (err error) {
	if element.stop {
		return ErrorClientOffline
	}
	if element.status != webrtc.ICEConnectionStateConnected {
		return nil
	}
	if tmp, ok := element.streams[pkt.Idx]; ok {
		if tmp.ts == 0 {
			tmp.ts = pkt.Time
		}
		switch tmp.codec.Type() {
		case av.H264:
			codec := tmp.codec.(h264parser.CodecData)
			if pkt.IsKeyFrame {
				pkt.Data = append([]byte{0, 0, 0, 1}, bytes.Join([][]byte{codec.SPS(), codec.PPS(), pkt.Data[4:]}, []byte{0, 0, 0, 1})...)
			} else {
				pkt.Data = pkt.Data[4:]
			}
			//log.Println("video", pkt.Time-tmp.ts)
		case av.PCM_MULAW:
			//log.Println("audio", pkt.Time-tmp.ts)
		case av.PCM_ALAW:
			//log.Println("audio", pkt.Time-tmp.ts)
		default:
			return ErrorCodecNotSupported
		}
		//log.Println(tmp.codec.Type(), pkt.Time-tmp.ts)
		err := tmp.track.WriteSample(media.Sample{Data: pkt.Data, Duration: pkt.Time - tmp.ts})
		element.streams[pkt.Idx].ts = pkt.Time
		return err
	}
	return ErrorNotFound

}
func (element *Muxer) WaitCloser() {
	select {
	case <-element.ps:
		element.stop = true
		element.Close()
	case <-element.pt.C:
		element.stop = true
		element.Close()
	}
}
func (element *Muxer) Close() error {
	if element.pc != nil {
		err := element.pc.Close()
		if err != nil {
			return err
		}
	}
	return nil
}
