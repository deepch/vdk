package webrtc

import (
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/deepch/vdk/av"
	"github.com/deepch/vdk/codec/h264parser"
	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
	"math/rand"
)

type Muxer struct {
	streams map[int8]*Stream
}
type Stream struct {
	codec av.CodecData
	track *webrtc.Track
}
func NewMuxer() *Muxer {
	return &Muxer{streams: make(map[int8]*Stream)}
}
func (self *Muxer) WriteHeader(streams []av.CodecData, sdp64 string) (string, error) {
	if len(streams) == 0 {
		return "",errors.New("No Stream Forund")
	}
	mediaEngine := webrtc.MediaEngine{}
	sdpB, err := base64.StdEncoding.DecodeString(sdp64)
	if err != nil {
		return "",err
	}
	offer := webrtc.SessionDescription{
		Type: webrtc.SDPTypeOffer,
		SDP:  string(sdpB),
	}
	if err = mediaEngine.PopulateFromSDP(offer); err != nil {
		return "",err
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
		return "",err
	}
	for i, i2 := range streams {
		var track *webrtc.Track
		if i2.Type().IsVideo() {
			track, err = peerConnection.NewTrack(getPayloadType(mediaEngine, webrtc.RTPCodecTypeVideo, i2.Type().String()), rand.Uint32(), "video", "pion")
			if err != nil {
				return "",err
			}
		} else if i2.Type().IsAudio() {
			track, err = peerConnection.NewTrack(getPayloadType(mediaEngine, webrtc.RTPCodecTypeAudio, i2.Type().String()), rand.Uint32(), "audio", "pion")
			if err != nil {
				return "",err
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
		_, err = peerConnection.AddTrack(track);
		if err != nil {
			return "",err
		}
		self.streams[int8(i)] = &Stream{track: track , codec: i2}
	}
	//iceConnectedCtx, iceConnectedCtxCancel := context.WithCancel(context.Background())
	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		fmt.Printf("Connection State has changed %s \n", connectionState.String())
		if connectionState == webrtc.ICEConnectionStateConnected {
		//	iceConnectedCtxCancel()
		}
	})
	if err = peerConnection.SetRemoteDescription(offer); err != nil {
		return "",err
	}
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		return "",err
	}
	//gatherComplete := webrtc.GatheringCompletePromise(peerConnection)
	if err = peerConnection.SetLocalDescription(answer); err != nil {
		return "",err
	}
	//<-gatherComplete


	return  base64.StdEncoding.EncodeToString([]byte(answer.SDP)), nil
}
func (self *Muxer) WritePacket(pkt av.Packet) (err error) {
	if tmp, ok := self.streams[pkt.Idx]; ok {
		switch tmp.codec.Type() {
		case av.H264:
			codec := tmp.codec.(h264parser.CodecData)
			if pkt.IsKeyFrame {
				pkt.Data = append([]byte("\000\000\001"+string(codec.SPS())+"\000\000\001"+string(codec.PPS())+"\000\000\001"), pkt.Data[4:]...)

			} else {
				pkt.Data = pkt.Data[4:]
			}
			return tmp.track.WriteSample(media.Sample{Data: pkt.Data, Samples: 90000})
		default:
			return errors.New("Media Track Not Found")
		}
	}
	return errors.New("Media Track Not Found")
}
func getPayloadType(m webrtc.MediaEngine, codecType webrtc.RTPCodecType, codecName string) uint8 {
	for _, codec := range m.GetCodecsByKind(codecType) {
		if codec.Name == codecName {
			return codec.PayloadType
		}
	}
	panic(fmt.Sprintf("Remote peer does not support %s", codecName))
}

/*
//MuxerRTC muxer
type MuxerRTC struct {
	Peer  *webrtc.PeerConnection
	Muxer string
	Track []*webrtc.Track
}

//NewWebRT adapter
func NewWebRT() *MuxerRTC {
	peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	})
	if err != nil {
		panic(err)
	}
	log.Println(peerConnection)
	return &MuxerRTC{Peer: peerConnection}
}

//AddTrackVideo video
func (element *MuxerRTC) AddTrackVideo() error {
	Track, err := element.Peer.NewTrack(webrtc.DefaultPayloadTypeH264, rand.Uint32(), "video", "video")
	if err != nil {
		return err
	}
	_, err = element.Peer.AddTrack(Track)
	if err != nil {
		return err
	}
	element.Track = append(element.Track, Track)
	return nil
}

//AddTrackAudio audio
func (element *MuxerRTC) AddTrackAudio() error {
	Track, err := element.Peer.NewTrack(webrtc.DefaultPayloadTypePCMA, rand.Uint32(), "audio", "audio")
	if err != nil {
		return err
	}
	_, err = element.Peer.AddTrack(Track)
	if err != nil {
		return err
	}
	_, err = element.Peer.AddTransceiverFromTrack(Track, webrtc.RtpTransceiverInit{Direction: webrtc.RTPTransceiverDirectionSendonly})
	if err != nil {
		return err
	}
	element.Track = append(element.Track, Track)
	return nil
}

//CreateAnswer func
func (element *MuxerRTC) CreateAnswer(data string) ([]byte, error) {
	//	log.Println(string(data))

	//	sd, err := base64.StdEncoding.DecodeString(data)
	//	if err != nil {
	//		log.Println("DecodeString", err)
	//		return nil, err
	//	}
	offer := webrtc.SessionDescription{
		Type: webrtc.SDPTypeOffer,
		SDP:  data,
	}
	if err := element.Peer.SetRemoteDescription(offer); err != nil {
		log.Println("SetRemoteDescription", err)

		return nil, err
	}
	answer, err := element.Peer.CreateAnswer(nil)
	if err != nil {
		log.Println("CreateAnswer", err)

		return nil, err
	}
	return []byte(answer.SDP), nil
}

//WritePacket to stream
func (element *MuxerRTC) WritePacket(id int, data []byte, smaple uint32) error {
	err := element.Track[id].WriteSample(media.Sample{Data: data, Samples: smaple})
	return err
}
func (element *MuxerRTC) Close() {
	return
}
*/
