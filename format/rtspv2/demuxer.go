package rtspv2

import (
	"encoding/binary"
	"math"
	"time"

	"github.com/deepch/vdk/av"
	"github.com/deepch/vdk/codec/aacparser"
	"github.com/deepch/vdk/codec/h264parser"
	"github.com/deepch/vdk/codec/h265parser"
)

const (
	TimeBaseFactor = 90
	TimeDelay      = 1
)

func (client *RTSPClient) containsPayloadType(pt int) bool {
	var exist bool
	for _, sdp := range client.mediaSDP {
		if sdp.Rtpmap == pt {
			exist = true
		}
	}
	return exist
}

// func (client *RTSPClient) durationFromSDP() time.Duration {
// 	for _, sdp := range client.mediaSDP {
// 		if sdp.AVType == VIDEO && sdp.FPS != 0 {
// 			return time.Duration((int(1000) / sdp.FPS) * int(time.Millisecond))
// 		}
// 	}
// 	return 0
// }

func (client *RTSPClient) RTPDemuxer(payloadRAW *[]byte) ([]*av.Packet, bool) {
	content := *payloadRAW
	firstByte := content[4]
	padding := (firstByte>>5)&1 == 1
	extension := (firstByte>>4)&1 == 1
	CSRCCnt := int(firstByte & 0x0f)
	payloadType := int(content[5] & 0x7f)
	sequenceNumber := int(binary.BigEndian.Uint16(content[6:8]))
	timestamp := int64(binary.BigEndian.Uint32(content[8:12]))
	// SSRC := binary.BigEndian.Uint32(content[12:16])
	if isRTCPPacket(content) {
		client.Println("skipping RTCP packet")
		return nil, false
	}

	if !client.containsPayloadType(payloadType) {
		// client.Println(fmt.Sprintf("skipping RTP packet, paytload type: %v", payloadType))
		return nil, false
	}

	// client.Println(fmt.Sprintf("padding: %v, extension: %v, csrccnt: %d, sequence number: %d.payload type: %d, timestamp: %d", padding, extension, CSRCCnt, sequenceNumber, payloadType, timestamp))
	client.offset = RTPHeaderSize
	client.sequenceNumber = sequenceNumber
	client.timestamp = timestamp
	client.end = len(content)
	if client.end-client.offset >= 4*CSRCCnt {
		client.offset += 4 * CSRCCnt
	}
	if extension && len(content) < 4+client.offset+2+2 {
		return nil, false
	}
	var realTimestamp int64
	if extension && client.end-client.offset >= 4 {
		extWords := int(binary.BigEndian.Uint16(content[4+client.offset+2:]))
		extLen := 4 * extWords
		client.offset += 4 // this is profile(2 byte) + ext length(2 byte)
		realTimestamp = int64(binary.BigEndian.Uint32(content[client.offset+4 : client.offset+4*2]))
		if client.end-client.offset >= extLen {
			client.offset += extLen
		}
	}
	client.realVideoTs = realTimestamp

	if padding && client.end-client.offset > 0 {
		paddingLen := int(content[client.end-1])
		if client.end-client.offset >= paddingLen {
			client.end -= paddingLen
		}
	}
	client.offset += 4
	if len(content) < client.end {
		return nil, false
	}

	switch int(content[1]) {
	case client.videoID:
		return client.handleVideo(content)
	case client.audioID:
		return client.handleAudio(content)
	}
	return nil, false
}

func (client *RTSPClient) handleVideo(content []byte) ([]*av.Packet, bool) {
	if client.PreVideoTS == 0 {
		client.PreVideoTS = client.timestamp
	}
	if client.timestamp-client.PreVideoTS < 0 {
		if math.MaxUint32-client.PreVideoTS < 90*100 { //100 ms
			client.PreVideoTS = 0
			client.PreVideoTS -= (math.MaxUint32 - client.PreVideoTS)
		} else {
			client.PreVideoTS = 0
		}
	}
	if client.PreSequenceNumber != 0 && client.sequenceNumber-client.PreSequenceNumber != 1 {
		client.Println("drop packet", client.sequenceNumber-1)
	}
	client.PreSequenceNumber = client.sequenceNumber
	if client.BufferRtpPacket.Len() > 4048576 {
		client.Println("Big Buffer Flush")
		client.BufferRtpPacket.Truncate(0)
		client.BufferRtpPacket.Reset()
	}
	nalRaw, _ := h264parser.SplitNALUs(content[client.offset:client.end])
	if len(nalRaw) == 0 || len(nalRaw[0]) == 0 {
		client.Println("nal Raw 0", nalRaw)
		return nil, false
	}
	var retmap []*av.Packet
	for _, nal := range nalRaw {
		if client.videoCodec == av.H265 {
			retmap = client.handleH265Payload(nal, retmap)
		} else if client.videoCodec == av.H264 {
			retmap = client.handleH264Payload(content, nal, retmap)
		}
	}
	if len(retmap) > 0 {
		client.PreVideoTS = client.timestamp
		return retmap, true
	}

	return nil, false
}

func (client *RTSPClient) handleH264Payload(content, nal []byte, retmap []*av.Packet) []*av.Packet {
	naluType := nal[0] & 0x1f
	switch {
	case naluType >= 1 && naluType <= 5:
		retmap = client.appendVideoPacket(retmap, nal, naluType == 5)
	case naluType == h264parser.NALU_SPS:
		client.CodecUpdateSPS(nal)
	case naluType == h264parser.NALU_PPS:
		client.CodecUpdatePPS(nal)
	case naluType == 24:
		packet := nal[1:]
		for len(packet) >= 2 {
			size := int(packet[0])<<8 | int(packet[1])
			if size+2 > len(packet) {
				break
			}
			naluTypefs := packet[2] & 0x1f
			switch {
			case naluTypefs >= 1 && naluTypefs <= 5:
				retmap = client.appendVideoPacket(retmap, packet[2:size+2], naluTypefs == 5)
			case naluTypefs == h264parser.NALU_SPS:
				client.CodecUpdateSPS(packet[2 : size+2])
			case naluTypefs == h264parser.NALU_PPS:
				client.CodecUpdatePPS(packet[2 : size+2])
			}
			packet = packet[size+2:]
		}
	case naluType == 28:
		fuIndicator := content[client.offset]
		fuHeader := content[client.offset+1]
		isStart := fuHeader&0x80 != 0
		isEnd := fuHeader&0x40 != 0
		if isStart {
			client.fuStarted = true
			client.BufferRtpPacket.Truncate(0)
			client.BufferRtpPacket.Reset()
			client.BufferRtpPacket.Write([]byte{fuIndicator&0xe0 | fuHeader&0x1f})
		}
		if client.fuStarted {
			client.BufferRtpPacket.Write(content[client.offset+2 : client.end])
			if isEnd {
				client.fuStarted = false
				naluTypef := client.BufferRtpPacket.Bytes()[0] & 0x1f
				if naluTypef == 7 || naluTypef == 9 {
					bufered, _ := h264parser.SplitNALUs(append([]byte{0, 0, 0, 1}, client.BufferRtpPacket.Bytes()...))
					for _, v := range bufered {
						naluTypefs := v[0] & 0x1f
						switch {
						case naluTypefs == 5:
							client.BufferRtpPacket.Reset()
							client.BufferRtpPacket.Write(v)
							naluTypef = 5
						case naluTypefs == h264parser.NALU_SPS:
							client.CodecUpdateSPS(v)
						case naluTypefs == h264parser.NALU_PPS:
							client.CodecUpdatePPS(v)
						}
					}
				}
				retmap = client.appendVideoPacket(retmap, client.BufferRtpPacket.Bytes(), naluTypef == 5)
			}
		}
	default:
		//client.Println("Unsupported NAL Type", naluType)
	}

	return retmap
}

func (client *RTSPClient) handleH265Payload(nal []byte, retmap []*av.Packet) []*av.Packet {
	naluType := (nal[0] >> 1) & 0x3f
	switch naluType {
	case h265parser.NAL_UNIT_CODED_SLICE_TRAIL_R:
		retmap = client.appendVideoPacket(retmap, nal, false)
	case h265parser.NAL_UNIT_VPS:
		client.CodecUpdateVPS(nal)
	case h265parser.NAL_UNIT_SPS:
		client.CodecUpdateSPS(nal)
	case h265parser.NAL_UNIT_PPS:
		client.CodecUpdatePPS(nal)
	case h265parser.NAL_UNIT_UNSPECIFIED_49:
		se := nal[2] >> 6
		naluType := nal[2] & 0x3f
		switch se {
		case 2:
			client.BufferRtpPacket.Truncate(0)
			client.BufferRtpPacket.Reset()
			client.BufferRtpPacket.Write([]byte{(nal[0] & 0x81) | (naluType << 1), nal[1]})
			r := make([]byte, 2)
			r[1] = nal[1]
			r[0] = (nal[0] & 0x81) | (naluType << 1)
			client.BufferRtpPacket.Write(nal[3:])
		case 1:
			client.BufferRtpPacket.Write(nal[3:])
			retmap = client.appendVideoPacket(retmap, client.BufferRtpPacket.Bytes(), naluType == h265parser.NAL_UNIT_CODED_SLICE_IDR_W_RADL)
		default:
			client.BufferRtpPacket.Write(nal[3:])
		}
	default:
		//client.Println("Unsupported Nal", naluType)
	}
	return retmap
}

func (client *RTSPClient) handleAudio(content []byte) ([]*av.Packet, bool) {
	if client.PreAudioTS == 0 {
		client.PreAudioTS = client.timestamp
	}
	nalRaw, _ := h264parser.SplitNALUs(content[client.offset:client.end])
	var retmap []*av.Packet
	for _, nal := range nalRaw {
		var duration time.Duration
		switch client.audioCodec {
		case av.PCM_MULAW, av.PCM_ALAW:
			duration = time.Duration(len(nal)) * time.Second / time.Duration(client.AudioTimeScale)
			retmap = client.appendAudioPacket(retmap, nal, duration)
		case av.OPUS:
			duration = time.Duration(20) * time.Millisecond
			retmap = client.appendAudioPacket(retmap, nal, duration)
		case av.AAC:
			auHeadersLength := uint16(0) | (uint16(nal[0]) << 8) | uint16(nal[1])
			auHeadersCount := auHeadersLength >> 4
			framesPayloadOffset := 2 + int(auHeadersCount)<<1
			auHeaders := nal[2:framesPayloadOffset]
			framesPayload := nal[framesPayloadOffset:]
			for i := 0; i < int(auHeadersCount); i++ {
				auHeader := uint16(0) | (uint16(auHeaders[0]) << 8) | uint16(auHeaders[1])
				frameSize := auHeader >> 3
				frame := framesPayload[:frameSize]
				auHeaders = auHeaders[2:]
				framesPayload = framesPayload[frameSize:]
				if _, _, _, _, err := aacparser.ParseADTSHeader(frame); err == nil {
					frame = frame[7:]
				}
				duration = time.Duration((float32(1024)/float32(client.AudioTimeScale))*1000*1000*1000) * time.Nanosecond
				retmap = client.appendAudioPacket(retmap, frame, duration)
			}
		}
	}
	if len(retmap) > 0 {
		client.PreAudioTS = client.timestamp
		return retmap, true
	}
	return nil, false
}

func (client *RTSPClient) appendAudioPacket(retmap []*av.Packet, nal []byte, duration time.Duration) []*av.Packet {
	client.AudioTimeLine += duration
	return append(retmap, &av.Packet{
		Data:            nal,
		CompositionTime: time.Duration(1) * time.Millisecond,
		Duration:        duration,
		Idx:             client.audioIDX,
		IsKeyFrame:      false,
		Time:            client.AudioTimeLine,
	})
}

func (client *RTSPClient) appendVideoPacket(retmap []*av.Packet, nal []byte, isKeyFrame bool) []*av.Packet {
	// Playback has real video ts
	if client.realVideoTs != 0 {
		return client.appendPlaybackVideoPacket(retmap, nal, isKeyFrame)
	} else {
		// LiveView
		return client.appendLiveViewVideoPacket(retmap, nal, isKeyFrame)
	}
}

func (client *RTSPClient) appendPlaybackVideoPacket(retmap []*av.Packet, nal []byte, isKeyFrame bool) []*av.Packet {
	prePkt := client.PrePacket
	curPkt := &av.Packet{
		Data:            append(binSize(len(nal)), nal...),
		CompositionTime: time.Duration(TimeDelay) * time.Millisecond,
		Idx:             client.videoIDX,
		IsKeyFrame:      isKeyFrame,
		Duration:        client.PreDuration,
		Time:            time.Duration(client.timestamp/TimeBaseFactor) * time.Millisecond,
		RealTimestamp:   0,
		RealTs:          client.realVideoTs,
	}
	client.PrePacket = append(retmap, curPkt)
	if len(prePkt) == 0 {
		return nil
	} else {
		var prePktTime time.Duration
		lenPrePkt := len(prePkt)
		for i := lenPrePkt - 1; i >= 0; i-- {
			prePktTime = prePkt[i].Time
			if i+1 == lenPrePkt {
				prePkt[i].Duration = time.Duration(client.timestamp/TimeBaseFactor)*time.Millisecond - prePkt[i].Time
			} else {
				prePkt[i].Duration = (prePkt[i].Time - prePktTime) * time.Millisecond
			}
			client.PreDuration = prePkt[i].Duration

			if prePkt[i].IsKeyFrame {
				if prePkt[i].RealTs == client.preRealVideoMs/1000 {
					client.iterateDruation += prePkt[i].Duration
				} else {
					client.iterateDruation = 0
					client.preKeyRealVideoTs = prePkt[i].RealTs
				}
			} else {
				client.iterateDruation += prePkt[i].Duration
			}

			prePkt[i].RealTimestamp = client.preKeyRealVideoTs*1000 + client.iterateDruation.Milliseconds()
			client.preRealVideoMs = prePkt[i].RealTimestamp
			// fmt.Println("playback duration", prePkt[i].IsKeyFrame, prePkt[i].RealTs, client.preRealVideoMs, prePkt[i].Duration, client.iterateDruation)
		}
		return prePkt
	}
}

func (client *RTSPClient) appendLiveViewVideoPacket(retmap []*av.Packet, nal []byte, isKeyFrame bool) []*av.Packet {
	prePkt := client.PrePacket
	client.PrePacket = append(retmap, &av.Packet{
		Data:            append(binSize(len(nal)), nal...),
		CompositionTime: time.Duration(TimeDelay) * time.Millisecond,
		Idx:             client.videoIDX,
		IsKeyFrame:      isKeyFrame,
		Duration:        client.PreDuration,
		Time:            time.Duration(client.timestamp/TimeBaseFactor) * time.Millisecond,
	})
	if len(prePkt) == 0 {
		return nil
	} else {
		var prePktTime time.Duration
		lenPrePkt := len(prePkt)
		for i := lenPrePkt - 1; i >= 0; i-- {
			if isKeyFrame {
				prePkt[i].Duration = client.PreDuration
			} else {
				prePktTime = prePkt[i].Time
				if i+1 == lenPrePkt {
					prePkt[i].Duration = time.Duration(client.timestamp/TimeBaseFactor)*time.Millisecond - prePkt[i].Time
				} else {
					prePkt[i].Duration = (prePkt[i].Time - prePktTime) * time.Millisecond
				}
				client.PreDuration = prePkt[i].Duration
			}
			// fmt.Println("liveview duration", prePkt[i].IsKeyFrame, prePkt[i].Time.Milliseconds(), prePkt[i].Duration)

		}
		return prePkt
	}
}
