package webrtc

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/pion/webrtc/v2"

	"github.com/deepch/vdk/av"
	"github.com/deepch/vdk/codec/h264parser"
	"github.com/pion/webrtc/v2/pkg/media"
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
	track *webrtc.Track
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
	mediaEngine := webrtc.MediaEngine{}
	sdpB, err := base64.StdEncoding.DecodeString(sdp64)
	if err != nil {
		return "", err
	}
	offer := webrtc.SessionDescription{
		Type: webrtc.SDPTypeOffer,
		SDP:  string(sdpB),
	}
	if err = mediaEngine.PopulateFromSDP(offer); err != nil {
		return "", err
	}
	api := webrtc.NewAPI(webrtc.WithMediaEngine(mediaEngine))
	peerConnection, err := api.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	})
	if err != nil {
		return "", err
	}
	for i, i2 := range streams {
		var track *webrtc.Track
		if i2.Type().IsVideo() {
			track, err = peerConnection.NewTrack(getPayloadType(mediaEngine, webrtc.RTPCodecTypeVideo, i2.Type().String()), rand.Uint32(), "video", Label)
			if err != nil {
				return "", err
			}
		} else if i2.Type().IsAudio() {
			track, err = peerConnection.NewTrack(getPayloadType(mediaEngine, webrtc.RTPCodecTypeAudio, i2.Type().String()), rand.Uint32(), "audio", Label)
			if err != nil {
				return "", err
			}
		}
		_, err = peerConnection.AddTransceiverFromTrack(track,
			webrtc.RtpTransceiverInit{
				Direction: webrtc.RTPTransceiverDirectionSendonly,
			},
		)
		if err != nil {
			return "", err
		}
		_, err = peerConnection.AddTrack(track)
		if err != nil {
			return "", err
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
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		return "", err
	}
	if err = peerConnection.SetLocalDescription(answer); err != nil {
		return "", err
	}
	element.pc = peerConnection
	return base64.StdEncoding.EncodeToString([]byte(answer.SDP)), nil
}

func (element *Muxer) WritePacket(pkt av.Packet) (err error) {
	if element.stop {
		return ErrorClientOffline
	}
	if element.status != webrtc.ICEConnectionStateConnected {
		return nil
	}
	if tmp, ok := element.streams[pkt.Idx]; ok {
		switch tmp.codec.Type() {
		case av.H264:
			codec := tmp.codec.(h264parser.CodecData)
			if pkt.IsKeyFrame {
				pkt.Data = append([]byte{0, 0, 0, 1}, bytes.Join([][]byte{codec.SPS(), codec.PPS(), pkt.Data[4:]}, []byte{0, 0, 0, 1})...)
			} else {
				pkt.Data = pkt.Data[4:]
			}
			return tmp.track.WriteSample(media.Sample{Data: pkt.Data, Samples: 90000})
		default:
			return ErrorCodecNotSupported
		}
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

func getPayloadType(m webrtc.MediaEngine, codecType webrtc.RTPCodecType, codecName string) uint8 {
	for _, codec := range m.GetCodecsByKind(codecType) {
		if codec.Name == codecName {
			return codec.PayloadType
		}
	}
	panic(fmt.Sprintf("Remote peer does not support %s", codecName))
}
