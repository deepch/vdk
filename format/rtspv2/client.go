package rtspv2

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"html"
	"io"
	"log"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/deepch/vdk/av"
	"github.com/deepch/vdk/codec"
	"github.com/deepch/vdk/codec/aacparser"
	"github.com/deepch/vdk/codec/h264parser"
	"github.com/deepch/vdk/codec/h265parser"
	"github.com/deepch/vdk/format/rtsp/sdp"
)

const (
	SignalStreamRTPStop = iota
	SignalCodecUpdate
)

const (
	VIDEO = "video"
	AUDIO = "audio"
)

const (
	RTPHeaderSize = 12
)
const (
	DESCRIBE = "DESCRIBE"
	OPTIONS  = "OPTIONS"
	PLAY     = "PLAY"
	PAUSE    = "PAUSE"
	SETUP    = "SETUP"
	RECORD   = "RECORD"
	ANNOUNCE = "ANNOUNCE"
	TEARDOWN = "TEARDOWN"
)

type RTSPClient struct {
	control             string
	seq                 int
	session             string
	realm               string
	nonce               string
	username            string
	password            string
	startVideoTS        int64
	startAudioTS        int64
	videoID             int
	audioID             int
	videoIDX            int8
	audioIDX            int8
	mediaSDP            []sdp.Media
	SDPRaw              []byte
	conn                net.Conn
	connRW              *bufio.ReadWriter
	pURL                *url.URL
	headers             map[string]string
	Signals             chan int
	OutgoingProxyQueue  chan *[]byte
	OutgoingPacketQueue chan *av.Packet
	clientDigest        bool
	clientBasic         bool
	fuStarted           bool
	options             RTSPClientOptions
	BufferRtpPacket     *bytes.Buffer
	vps                 []byte
	sps                 []byte
	pps                 []byte
	CodecData           []av.CodecData
	AudioTimeLine       time.Duration
	AudioTimeScale      int64
	audioCodec          av.CodecType
	videoCodec          av.CodecType
	PreAudioTS          int64
	PreVideoTS          int64
	PreSequenceNumber   int
	FPS                 int
	WaitCodec           bool
}

type RTSPClientOptions struct {
	Debug            bool
	URL              string
	DialTimeout      time.Duration
	ReadWriteTimeout time.Duration
	DisableAudio     bool
	OutgoingProxy    bool
}

func Dial(options RTSPClientOptions) (*RTSPClient, error) {
	client := &RTSPClient{
		headers:             make(map[string]string),
		Signals:             make(chan int, 100),
		OutgoingProxyQueue:  make(chan *[]byte, 3000),
		OutgoingPacketQueue: make(chan *av.Packet, 3000),
		BufferRtpPacket:     bytes.NewBuffer([]byte{}),
		videoID:             -1,
		audioID:             -2,
		videoIDX:            -1,
		audioIDX:            -2,
		options:             options,
		AudioTimeScale:      8000,
	}
	client.headers["User-Agent"] = "Lavf58.20.100"
	Debug = options.Debug
	err := client.parseURL(html.UnescapeString(client.options.URL))
	if err != nil {
		return nil, err
	}
	conn, err := net.DialTimeout("tcp", client.pURL.Host, client.options.DialTimeout)
	if err != nil {
		return nil, err
	}
	err = conn.SetDeadline(time.Now().Add(client.options.ReadWriteTimeout))
	if err != nil {
		return nil, err
	}
	client.conn = conn
	client.connRW = bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	err = client.request(OPTIONS, nil, client.pURL.String(), nil, false, false)
	if err != nil {
		return nil, err
	}
	err = client.request(DESCRIBE, map[string]string{"Accept": "application/sdp"}, client.pURL.String(), nil, false, false)
	if err != nil {
		return nil, err
	}
	var ch int
	for _, m := range client.mediaSDP {
		if (m.AVType != VIDEO && m.AVType != AUDIO) || (client.options.DisableAudio && m.AVType == AUDIO) {
			continue
		}
		err = client.request(SETUP, map[string]string{"Transport": "RTP/AVP/TCP;unicast;interleaved=" + strconv.Itoa(ch) + "-" + strconv.Itoa(ch+1)}, client.ControlTrack(m.Control), nil, false, false)
		if err != nil {
			return nil, err
		}
		if m.AVType == VIDEO {
			if m.Type == av.H264 {
				if len(m.SpropParameterSets) > 1 {
					if codecData, err := h264parser.NewCodecDataFromSPSAndPPS(m.SpropParameterSets[0], m.SpropParameterSets[1]); err == nil {
						client.sps = m.SpropParameterSets[0]
						client.pps = m.SpropParameterSets[1]
						client.CodecData = append(client.CodecData, codecData)
					}
				} else {
					client.CodecData = append(client.CodecData, h264parser.CodecData{})
					client.WaitCodec = true
				}
				client.FPS = m.FPS
				client.videoCodec = av.H264
			} else if m.Type == av.H265 {
				if len(m.SpropVPS) > 1 && len(m.SpropSPS) > 1 && len(m.SpropPPS) > 1 {
					if codecData, err := h265parser.NewCodecDataFromVPSAndSPSAndPPS(m.SpropVPS, m.SpropSPS, m.SpropPPS); err == nil {
						client.vps = m.SpropVPS
						client.sps = m.SpropSPS
						client.pps = m.SpropPPS
						client.CodecData = append(client.CodecData, codecData)
					}
				} else {
					client.CodecData = append(client.CodecData, h265parser.CodecData{})
				}
				client.videoCodec = av.H265
			} else {
				client.Println("SDP Video Codec Type Not Supported", m.Type)
			}
			client.videoIDX = int8(len(client.CodecData) - 1)
			client.videoID = ch
		}
		if m.AVType == AUDIO {
			client.audioID = ch
			var CodecData av.AudioCodecData
			switch m.Type {
			case av.AAC:
				CodecData, err = aacparser.NewCodecDataFromMPEG4AudioConfigBytes(m.Config)
				if err == nil {
					client.Println("Audio AAC bad config")
				}
			case av.OPUS:
				var cl av.ChannelLayout
				switch m.ChannelCount {
				case 1:
					cl = av.CH_MONO
				case 2:
					cl = av.CH_STEREO
				default:
					cl = av.CH_MONO
				}
				CodecData = codec.NewOpusCodecData(m.TimeScale, cl)
			case av.PCM_MULAW:
				CodecData = codec.NewPCMMulawCodecData()
			case av.PCM_ALAW:
				CodecData = codec.NewPCMAlawCodecData()
			case av.PCM:
				CodecData = codec.NewPCMCodecData()
			default:
				client.Println("Audio Codec", m.Type, "not supported")
			}
			if CodecData != nil {
				client.CodecData = append(client.CodecData, CodecData)
				client.audioIDX = int8(len(client.CodecData) - 1)
				client.audioCodec = CodecData.Type()
				if m.TimeScale != 0 {
					client.AudioTimeScale = int64(m.TimeScale)
				}
			}
		}
		ch += 2
	}

	err = client.request(PLAY, nil, client.control, nil, false, false)
	if err != nil {
		return nil, err
	}
	go client.startStream()
	return client, nil
}

func Publish(options RTSPClientOptions, sdp []byte) (*RTSPClient, error) {
	client := &RTSPClient{
		headers:             make(map[string]string),
		Signals:             make(chan int, 100),
		OutgoingProxyQueue:  make(chan *[]byte, 3000),
		OutgoingPacketQueue: make(chan *av.Packet, 3000),
		BufferRtpPacket:     bytes.NewBuffer([]byte{}),
		videoID:             -1,
		audioID:             -2,
		videoIDX:            -1,
		audioIDX:            -2,
		options:             options,
		AudioTimeScale:      8000,
	}
	client.headers["User-Agent"] = "Lavf58.20.100"
	err := client.parseURL(html.UnescapeString(client.options.URL))
	if err != nil {
		return nil, err
	}
	conn, err := net.DialTimeout("tcp", client.pURL.Host, client.options.DialTimeout)
	if err != nil {
		return nil, err
	}
	err = conn.SetDeadline(time.Now().Add(client.options.ReadWriteTimeout))
	if err != nil {
		return nil, err
	}
	client.conn = conn
	client.connRW = bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	err = client.request(OPTIONS, nil, client.pURL.String(), nil, false, false)
	if err != nil {
		return nil, err
	}

	err = client.request(ANNOUNCE, map[string]string{"Content-Type": "application/sdp", "Content-Length": fmt.Sprintf("%d", len(sdp))}, client.pURL.String(), sdp, false, false)
	if err != nil {
		return nil, err
	}

	//err = client.request(OPTIONS, nil, client.pURL.String(), nil, false, false)
	//if err != nil {
	//	return nil, err
	//}
	//
	//err = client.request(DESCRIBE, map[string]string{"Accept": "application/sdp"}, client.pURL.String(), nil,false, false)
	//if err != nil {
	//	return nil, err
	//}

	// interleved
	err = client.request(SETUP, map[string]string{"Transport": "RTP/AVP/TCP;unicast;interleaved=0-1;mode=record"}, client.pURL.String(), nil, false, false)
	if err != nil {
		return nil, err
	}

	err = client.request(RECORD, map[string]string{"Range": "npt=0.000-"}, client.pURL.String(), nil, false, false)
	if err != nil {
		return nil, err
	}
	return client, err
}

func (client *RTSPClient) SendRTP(pack []byte) {
	client.conn.Write(pack)
}

func (client *RTSPClient) ControlTrack(track string) string {
	if strings.Contains(track, "rtsp://") {
		return track
	}
	if !strings.HasSuffix(client.control, "/") {
		track = "/" + track
	}
	return client.control + track
}

func (client *RTSPClient) startStream() {
	defer func() {
		client.Signals <- SignalStreamRTPStop
	}()
	timer := time.Now()
	oneb := make([]byte, 1)
	header := make([]byte, 4)
	var fixed bool
	for {
		err := client.conn.SetDeadline(time.Now().Add(client.options.ReadWriteTimeout))
		if err != nil {
			client.Println("RTSP Client RTP SetDeadline", err)
			return
		}
		if int(time.Now().Sub(timer).Seconds()) > 25 {
			err := client.request(OPTIONS, map[string]string{"Require": "implicit-play"}, client.control, nil, false, true)
			if err != nil {
				client.Println("RTSP Client RTP keep-alive", err)
				return
			}
			timer = time.Now()
		}
		if !fixed {
			nb, err := io.ReadFull(client.connRW, header)
			if err != nil || nb != 4 {
				client.Println("RTSP Client RTP Read Header", err)
				return
			}
		}
		fixed = false
		switch header[0] {
		case 0x24:
			length := int32(binary.BigEndian.Uint16(header[2:]))
			if length > 65535 || length < 12 {
				client.Println("RTSP Client RTP Incorrect Packet Size")
				return
			}
			content := make([]byte, length+4)
			content[0] = header[0] // magic 0x24
			content[1] = header[1] // channel
			content[2] = header[2] // length byte high
			content[3] = header[3] // length byte low
			n, rerr := io.ReadFull(client.connRW, content[4:length+4])
			if rerr != nil || n != int(length) {
				client.Println("RTSP Client RTP ReadFull", err)
				return
			}
			//atomic.AddInt64(&client.Bitrate, int64(length+4))
			if client.options.OutgoingProxy {
				if len(client.OutgoingProxyQueue) < 2000 {
					client.OutgoingProxyQueue <- &content
				} else {
					client.Println("RTSP Client OutgoingProxy Chanel Full")
					return
				}
			}
			pkt, got := client.RTPDemuxer(&content)
			if !got {
				continue
			}
			for _, i2 := range pkt {
				if len(client.OutgoingPacketQueue) > 2000 {
					client.Println("RTSP Client OutgoingPacket Chanel Full")
					return
				}
				client.OutgoingPacketQueue <- i2
			}
		case 0x52:
			var responseTmp []byte
			for {
				n, rerr := io.ReadFull(client.connRW, oneb)
				if rerr != nil || n != 1 {
					client.Println("RTSP Client RTP Read Keep-Alive Header", rerr)
					return
				}
				responseTmp = append(responseTmp, oneb...)
				if (len(responseTmp) > 4 && bytes.Compare(responseTmp[len(responseTmp)-4:], []byte("\r\n\r\n")) == 0) || len(responseTmp) > 768 {
					if strings.Contains(string(responseTmp), "Content-Length:") {
						si, err := strconv.Atoi(stringInBetween(string(responseTmp), "Content-Length: ", "\r\n"))
						if err != nil {
							client.Println("RTSP Client RTP Read Keep-Alive Content-Length", err)
							return
						}
						cont := make([]byte, si)
						_, err = io.ReadFull(client.connRW, cont)
						if err != nil {
							client.Println("RTSP Client RTP Read Keep-Alive ReadFull", err)
							return
						}
					}
					break
				}
			}
			client.Println("RTSP RESPONSE<<< ", string(responseTmp))
		default:
			client.Println("RTSP Client RTP Read DeSync")
			return
		}
	}
}

func (client *RTSPClient) Options() error {
	err := client.request(OPTIONS, nil, client.pURL.String(), nil, false, true)
	return err
}

func (client *RTSPClient) Play(customHeaders map[string]string) error {
	err := client.request(PLAY, customHeaders, client.pURL.String(), nil, false, true)
	return err
}

func (client *RTSPClient) Pause() error {
	err := client.request(PAUSE, nil, client.pURL.String(), nil, false, true)
	return err
}

func (client *RTSPClient) Announce(sdp []byte) error {
	err := client.request(ANNOUNCE, nil, client.pURL.String(), sdp, false, true)
	return err
}

// request nores 如果流已经在拉的话，nores必须为true，否则client.connRW.ReadLine() 和 startStream中会因为并发读取导致slice的数量不够，导致outbounds of slice panic.
func (client *RTSPClient) request(method string, customHeaders map[string]string, uri string, body []byte, one bool, nores bool) (err error) {
	err = client.conn.SetDeadline(time.Now().Add(client.options.ReadWriteTimeout))
	if err != nil {
		return
	}
	client.seq++
	builder := bytes.Buffer{}
	builder.WriteString(fmt.Sprintf("%s %s RTSP/1.0\r\n", method, uri))
	builder.WriteString(fmt.Sprintf("CSeq: %d\r\n", client.seq))
	if client.clientDigest {
		builder.WriteString(fmt.Sprintf("Authorization: %s\r\n", client.createDigest(method, uri)))
	}
	if customHeaders != nil {
		for k, v := range customHeaders {
			builder.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
		}
	}
	for k, v := range client.headers {
		builder.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	builder.WriteString(fmt.Sprintf("\r\n"))
	if body != nil && len(body) > 0 {
		builder.WriteString(string(body))
		builder.WriteString(fmt.Sprintf("\r\n"))
	}

	client.Println(builder.String())
	s := builder.String()
	_, err = client.connRW.WriteString(s)
	if err != nil {
		return
	}
	err = client.connRW.Flush()
	if err != nil {
		return
	}
	builder.Reset()
	if !nores {
		var isPrefix bool
		var line []byte
		var contentLen int
		res := make(map[string]string)
		for {
			line, isPrefix, err = client.connRW.ReadLine()
			if err != nil {
				return
			}
			if strings.Contains(string(line), "RTSP/1.0") && (!strings.Contains(string(line), "200") && !strings.Contains(string(line), "401")) {
				time.Sleep(1 * time.Second)
				err = errors.New("Camera send status" + string(line))
				return
			}
			builder.Write(line)
			if !isPrefix {
				builder.WriteString("\r\n")
			}
			if len(line) == 0 {
				break
			}
			splits := strings.SplitN(string(line), ":", 2)
			if len(splits) == 2 {
				if splits[0] == "Content-length" {
					splits[0] = "Content-Length"
				}
				res[splits[0]] = splits[1]
			}
		}
		client.Println(builder.String())
		if val, ok := res["WWW-Authenticate"]; ok {
			if strings.Contains(val, "Digest") {
				client.realm = stringInBetween(val, "realm=\"", "\"")
				client.nonce = stringInBetween(val, "nonce=\"", "\"")
				client.clientDigest = true
			} else if strings.Contains(val, "Basic") {
				client.headers["Authorization"] = "Basic " + base64.StdEncoding.EncodeToString([]byte(client.username+":"+client.password))
				client.clientBasic = true
			}
			if !one {
				err = client.request(method, customHeaders, uri, nil, true, false)
				return
			}
			err = errors.New("RTSP Client Unauthorized 401")
			return
		}
		if val, ok := res["Session"]; ok {
			splits2 := strings.Split(val, ";")
			client.session = strings.TrimSpace(splits2[0])
			client.headers["Session"] = strings.TrimSpace(splits2[0])
		}
		if val, ok := res["Content-Base"]; ok {
			client.control = strings.TrimSpace(val)
		}
		if val, ok := res["RTP-Info"]; ok {
			splits := strings.Split(val, ",")
			for _, v := range splits {
				splits2 := strings.Split(v, ";")
				for _, vs := range splits2 {
					if strings.Contains(vs, "rtptime") {
						splits3 := strings.Split(vs, "=")
						if len(splits3) == 2 {
							if client.startVideoTS == 0 {
								ts, _ := strconv.Atoi(strings.TrimSpace(splits3[1]))
								client.startVideoTS = int64(ts)
							} else {
								ts, _ := strconv.Atoi(strings.TrimSpace(splits3[1]))
								client.startAudioTS = int64(ts)
							}
						}
					}
				}
			}
		}
		if method == DESCRIBE {
			if val, ok := res["Content-Length"]; ok {
				contentLen, err = strconv.Atoi(strings.TrimSpace(val))
				if err != nil {
					return
				}
				client.SDPRaw = make([]byte, contentLen)
				_, err = io.ReadFull(client.connRW, client.SDPRaw)
				if err != nil {
					return
				}
				builder.Write(client.SDPRaw)
				_, client.mediaSDP = sdp.Parse(string(client.SDPRaw))
			}
		}
		client.Println(builder.String())
	}
	return
}

func (client *RTSPClient) Close() {
	if client.conn != nil {
		client.conn.SetDeadline(time.Now().Add(time.Second))
		client.request(TEARDOWN, nil, client.control, nil, false, true)
		err := client.conn.Close()
		client.Println("RTSP Client Close", err)
	}
}

func (client *RTSPClient) parseURL(rawURL string) error {
	l, err := url.Parse(rawURL)
	if err != nil {
		return err
	}
	username := l.User.Username()
	password, _ := l.User.Password()
	l.User = nil
	if l.Port() == "" {
		l.Host = fmt.Sprintf("%s:%s", l.Host, "554")
	}
	if l.Scheme != "rtsp" {
		l.Scheme = "rtsp"
	}
	client.pURL = l
	client.username = username
	client.password = password
	client.control = l.String()
	return nil
}

func (client *RTSPClient) createDigest(method string, uri string) string {
	md5UserRealmPwd := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%s:%s:%s", client.username, client.realm, client.password))))
	md5MethodURL := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%s:%s", method, uri))))
	response := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%s:%s:%s", md5UserRealmPwd, client.nonce, md5MethodURL))))
	Authorization := fmt.Sprintf("Digest username=\"%s\", realm=\"%s\", nonce=\"%s\", uri=\"%s\", response=\"%s\"", client.username, client.realm, client.nonce, uri, response)
	return Authorization
}

func stringInBetween(str string, start string, end string) (result string) {
	s := strings.Index(str, start)
	if s == -1 {
		return
	}
	str = str[s+len(start):]
	e := strings.Index(str, end)
	if e == -1 {
		return
	}
	str = str[:e]
	return str
}

func (client *RTSPClient) RTPDemuxer(payloadRAW *[]byte) ([]*av.Packet, bool) {
	content := *payloadRAW
	firstByte := content[4]
	version := (firstByte >> 6) & 0x03
	padding := (firstByte>>5)&1 == 1
	extension := (firstByte>>4)&1 == 1
	CSRCCnt := int(firstByte & 0x0f)
	marker := (content[5]>>7)&1 == 1
	payloadType := content[5] & 0x7f
	SequenceNumber := int(binary.BigEndian.Uint16(content[6:8]))
	timestamp := int64(binary.BigEndian.Uint32(content[8:12]))
	SSRC := binary.BigEndian.Uint32(content[12:16])
	if Debug {
		log.Printf("version: %d, padding: %v, extension: %v, csrccnt: %d, marker: %v, payload type: %d, sequence number: %d. timestamp: %d, ssrc: %d",
			version, padding, extension, CSRCCnt, marker, payloadType, SequenceNumber, timestamp, SSRC)
	}

	offset := RTPHeaderSize

	end := len(content)
	if end-offset >= 4*CSRCCnt {
		offset += 4 * CSRCCnt
	}
	if extension && len(content) < 4+offset+2+2 {
		return nil, false
	}
	if extension && end-offset >= 4 {
		extWords := int(binary.BigEndian.Uint16(content[4+offset+2:]))
		extLen := 4 * extWords
		offset += 4 // this is profile(2 byte) + ext length(2 byte)
		if Debug {
			for i := 0; i < extWords; i++ {
				ext := binary.BigEndian.Uint32(content[offset+4*i : offset+4*(i+1)])
				log.Printf("extension: %d: %d", i+1, ext)
				if i == 1 {
					log.Print("timestamp: ", time.Unix(int64(ext), 0).String())
				}
			}
		}
		if end-offset >= extLen {
			offset += extLen
		}
	}
	if padding && end-offset > 0 {
		paddingLen := int(content[end-1])
		if end-offset >= paddingLen {
			end -= paddingLen
		}
	}
	offset += 4
	switch int(content[1]) {
	case client.videoID:
		if client.PreVideoTS == 0 {
			client.PreVideoTS = timestamp
		}
		if client.PreSequenceNumber != 0 && SequenceNumber-client.PreSequenceNumber != 1 {
			client.Println("drop packet", SequenceNumber-1)
		}
		client.PreSequenceNumber = SequenceNumber
		if client.BufferRtpPacket.Len() > 4048576 {
			client.Println("Big Buffer Flush")
			client.BufferRtpPacket.Truncate(0)
			client.BufferRtpPacket.Reset()
		}
		nalRaw, _ := h264parser.SplitNALUs(content[offset:end])
		if len(nalRaw) == 0 || len(nalRaw[0]) == 0 {
			return nil, false
		}
		var retmap []*av.Packet
		for _, nal := range nalRaw {
			if client.videoCodec == av.H265 {
				naluType := (nal[0] >> 1) & 0x3f
				switch naluType {
				case h265parser.NAL_UNIT_CODED_SLICE_TRAIL_R:
					retmap = append(retmap, &av.Packet{
						Data:            append(binSize(len(nal)), nal...),
						CompositionTime: time.Duration(1) * time.Millisecond,
						Idx:             client.videoIDX,
						IsKeyFrame:      false,
						NaluType:        naluType,
						Duration:        time.Duration(float32(timestamp-client.PreVideoTS)/90) * time.Millisecond,
						Time:            time.Duration(timestamp/90) * time.Millisecond,
					})
				case h265parser.NAL_UNIT_VPS:
					client.CodecUpdateVPS(nal)
				case h265parser.NAL_UNIT_SPS:
					client.CodecUpdateSPS(nal)
				case h265parser.NAL_UNIT_PPS:
					client.CodecUpdatePPS(nal)
				case h265parser.NAL_UNIT_UNSPECIFIED_49:
					se := nal[2] >> 6
					naluType := nal[2] & 0x3f
					if se == 2 {
						client.BufferRtpPacket.Truncate(0)
						client.BufferRtpPacket.Reset()
						client.BufferRtpPacket.Write([]byte{(nal[0] & 0x81) | (naluType << 1), nal[1]})
						r := make([]byte, 2)
						r[1] = nal[1]
						r[0] = (nal[0] & 0x81) | (naluType << 1)
						client.BufferRtpPacket.Write(nal[3:])
					} else if se == 1 {
						client.BufferRtpPacket.Write(nal[3:])
						retmap = append(retmap, &av.Packet{
							Data:            append(binSize(client.BufferRtpPacket.Len()), client.BufferRtpPacket.Bytes()...),
							CompositionTime: time.Duration(1) * time.Millisecond,
							Idx:             client.videoIDX,
							IsKeyFrame:      naluType == h265parser.NAL_UNIT_CODED_SLICE_IDR_W_RADL,
							NaluType:        naluType,
							Duration:        time.Duration(float32(timestamp-client.PreVideoTS)/90) * time.Millisecond,
							Time:            time.Duration(timestamp/90) * time.Millisecond,
						})
					} else {
						client.BufferRtpPacket.Write(nal[3:])
					}
				default:
					client.Println("Unsupported Nal", naluType)
				}

			} else if client.videoCodec == av.H264 {
				naluType := nal[0] & 0x1f
				//log.Printf("nalu type: %d", naluType)
				switch {
				case naluType >= 1 && naluType <= 5:
					retmap = append(retmap, &av.Packet{
						Data:            append(binSize(len(nal)), nal...),
						CompositionTime: time.Duration(1) * time.Millisecond,
						Idx:             client.videoIDX,
						NaluType:        naluType,
						IsKeyFrame:      naluType == 5,
						Duration:        time.Duration(float32(timestamp-client.PreVideoTS)/90) * time.Millisecond,
						Time:            time.Duration(timestamp/90) * time.Millisecond,
					})
				case naluType == 7:
					client.CodecUpdateSPS(nal)
				case naluType == 8:
					client.CodecUpdatePPS(nal)
				case naluType == 24:
					client.Println("24 Type need add next version report https://github.com/deepch/vdk")
				case naluType == 28:
					fuIndicator := content[offset]
					fuHeader := content[offset+1]
					isStart := fuHeader&0x80 != 0
					isEnd := fuHeader&0x40 != 0
					if isStart {
						client.fuStarted = true
						client.BufferRtpPacket.Truncate(0)
						client.BufferRtpPacket.Reset()
						client.BufferRtpPacket.Write([]byte{fuIndicator&0xe0 | fuHeader&0x1f})
					}
					if client.fuStarted {
						client.BufferRtpPacket.Write(content[offset+2 : end])
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
									case naluTypefs == 7:
										client.CodecUpdateSPS(v)
									case naluTypefs == 8:
										client.CodecUpdatePPS(v)
									}
								}
							}
							retmap = append(retmap, &av.Packet{
								Data:            append(binSize(client.BufferRtpPacket.Len()), client.BufferRtpPacket.Bytes()...),
								CompositionTime: time.Duration(1) * time.Millisecond,
								Duration:        time.Duration(float32(timestamp-client.PreVideoTS)/90) * time.Millisecond,
								Idx:             client.videoIDX,
								IsKeyFrame:      naluTypef == 5,
								NaluType:        naluTypef,
								Time:            time.Duration(timestamp/90) * time.Millisecond,
							})
						}
					}
				default:
					client.Println("Unsupported NAL Type", naluType)
				}
			}
		}
		if len(retmap) > 0 {
			client.PreVideoTS = timestamp
			return retmap, true
		}
	case client.audioID:
		if client.PreAudioTS == 0 {
			client.PreAudioTS = timestamp
		}
		nalRaw, _ := h264parser.SplitNALUs(content[offset:end])
		var retmap []*av.Packet
		for _, nal := range nalRaw {
			var duration time.Duration
			switch client.audioCodec {
			case av.PCM_MULAW:
				duration = time.Duration(len(nal)) * time.Second / time.Duration(client.AudioTimeScale)
				client.AudioTimeLine += duration
				retmap = append(retmap, &av.Packet{
					Data:            nal,
					CompositionTime: time.Duration(1) * time.Millisecond,
					Duration:        duration,
					Idx:             client.audioIDX,
					IsKeyFrame:      false,
					Time:            client.AudioTimeLine,
				})
			case av.PCM_ALAW:
				duration = time.Duration(len(nal)) * time.Second / time.Duration(client.AudioTimeScale)
				client.AudioTimeLine += duration
				retmap = append(retmap, &av.Packet{
					Data:            nal,
					CompositionTime: time.Duration(1) * time.Millisecond,
					Duration:        duration,
					Idx:             client.audioIDX,
					IsKeyFrame:      false,
					Time:            client.AudioTimeLine,
				})
			case av.OPUS:
				duration = time.Duration(20) * time.Millisecond
				client.AudioTimeLine += duration
				retmap = append(retmap, &av.Packet{
					Data:            nal,
					CompositionTime: time.Duration(1) * time.Millisecond,
					Duration:        duration,
					Idx:             client.audioIDX,
					IsKeyFrame:      false,
					Time:            client.AudioTimeLine,
				})
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
					client.AudioTimeLine += duration
					retmap = append(retmap, &av.Packet{
						Data:            frame,
						CompositionTime: time.Duration(1) * time.Millisecond,
						Duration:        duration,
						Idx:             client.audioIDX,
						IsKeyFrame:      false,
						Time:            client.AudioTimeLine,
					})
				}
			}
		}
		if len(retmap) > 0 {
			client.PreAudioTS = timestamp
			return retmap, true
		}
	default:
		//client.Println("Unsuported Intervaled data packet", int(content[1]), content[offset:end])
	}
	return nil, false
}

func (client *RTSPClient) CodecUpdateSPS(val []byte) {
	if client.videoCodec != av.H264 && client.videoCodec != av.H265 {
		return
	}
	if bytes.Compare(val, client.sps) == 0 {
		return
	}
	client.sps = val
	if (client.videoCodec == av.H264 && len(client.pps) == 0) || (client.videoCodec == av.H265 && (len(client.vps) == 0 || len(client.pps) == 0)) {
		return
	}
	var codecData av.VideoCodecData
	var err error
	switch client.videoCodec {
	case av.H264:
		client.Println("Codec Update SPS", val)
		codecData, err = h264parser.NewCodecDataFromSPSAndPPS(val, client.pps)
		if err != nil {
			client.Println("Parse Codec Data Error", err)
			return
		}
	case av.H265:
		codecData, err = h265parser.NewCodecDataFromVPSAndSPSAndPPS(client.vps, val, client.pps)
		if err != nil {
			client.Println("Parse Codec Data Error", err)
			return
		}
	}
	if len(client.CodecData) > 0 {
		for i, i2 := range client.CodecData {
			if i2.Type().IsVideo() {
				client.CodecData[i] = codecData
			}
		}
	} else {
		client.CodecData = append(client.CodecData, codecData)
	}
	client.Signals <- SignalCodecUpdate
}

func (client *RTSPClient) CodecUpdatePPS(val []byte) {
	if client.videoCodec != av.H264 && client.videoCodec != av.H265 {
		return
	}
	if bytes.Compare(val, client.pps) == 0 {
		return
	}
	client.pps = val
	if (client.videoCodec == av.H264 && len(client.sps) == 0) || (client.videoCodec == av.H265 && (len(client.vps) == 0 || len(client.sps) == 0)) {
		return
	}
	var codecData av.VideoCodecData
	var err error
	switch client.videoCodec {
	case av.H264:
		client.Println("Codec Update PPS", val)
		codecData, err = h264parser.NewCodecDataFromSPSAndPPS(client.sps, val)
		if err != nil {
			client.Println("Parse Codec Data Error", err)
			return
		}
	case av.H265:
		codecData, err = h265parser.NewCodecDataFromVPSAndSPSAndPPS(client.vps, client.sps, val)
		if err != nil {
			client.Println("Parse Codec Data Error", err)
			return
		}
	}
	if len(client.CodecData) > 0 {
		for i, i2 := range client.CodecData {
			if i2.Type().IsVideo() {
				client.CodecData[i] = codecData
			}
		}
	} else {
		client.CodecData = append(client.CodecData, codecData)
	}
	client.Signals <- SignalCodecUpdate
}

func (client *RTSPClient) CodecUpdateVPS(val []byte) {
	if client.videoCodec != av.H265 {
		return
	}
	if bytes.Compare(val, client.vps) == 0 {
		return
	}
	client.vps = val
	if len(client.sps) == 0 || len(client.pps) == 0 {
		return
	}
	codecData, err := h265parser.NewCodecDataFromVPSAndSPSAndPPS(val, client.sps, client.pps)
	if err != nil {
		client.Println("Parse Codec Data Error", err)
		return
	}
	if len(client.CodecData) > 0 {
		for i, i2 := range client.CodecData {
			if i2.Type().IsVideo() {
				client.CodecData[i] = codecData
			}
		}
	} else {
		client.CodecData = append(client.CodecData, codecData)
	}
	client.Signals <- SignalCodecUpdate
}

//Println mini logging functions
func (client *RTSPClient) Println(v ...interface{}) {
	if client.options.Debug {
		log.Println(v)
	}
}

//binSize
func binSize(val int) []byte {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(val))
	return buf
}
