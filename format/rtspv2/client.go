package rtspv2

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"crypto/tls"
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
	RTPHeaderSize      = 12
	RTCPSenderReport   = 200
	RTCPReceiverReport = 201
)

const (
	DESCRIBE = "DESCRIBE"
	OPTIONS  = "OPTIONS"
	PLAY     = "PLAY"
	SETUP    = "SETUP"
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
	chTMP               int
	timestamp           int64
	sequenceNumber      int
	end                 int
	offset              int
}

type RTSPClientOptions struct {
	Debug              bool
	URL                string
	DialTimeout        time.Duration
	ReadWriteTimeout   time.Duration
	DisableAudio       bool
	OutgoingProxy      bool
	InsecureSkipVerify bool
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
	client.headers["User-Agent"] = "Lavf58.76.100"
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
	if client.pURL.Scheme == "rtsps" {
		tlsConn := tls.Client(conn, &tls.Config{InsecureSkipVerify: options.InsecureSkipVerify, ServerName: client.pURL.Hostname()})
		err = tlsConn.Handshake()
		if err != nil {
			return nil, err
		}
		conn = tlsConn
	}
	client.conn = conn
	client.connRW = bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	err = client.request(OPTIONS, nil, client.pURL.String(), false, false)
	if err != nil {
		return nil, err
	}
	err = client.request(DESCRIBE, map[string]string{"Accept": "application/sdp"}, client.pURL.String(), false, false)
	if err != nil {
		return nil, err
	}
	for _, i2 := range client.mediaSDP {
		if (i2.AVType != VIDEO && i2.AVType != AUDIO) || (client.options.DisableAudio && i2.AVType == AUDIO) {
			//TODO check it
			if strings.Contains(string(client.SDPRaw), "LaunchDigital") {
				client.chTMP += 2
			}
			continue
		}
		err = client.request(SETUP, map[string]string{"Transport": "RTP/AVP/TCP;unicast;interleaved=" + strconv.Itoa(client.chTMP) + "-" + strconv.Itoa(client.chTMP+1)}, client.ControlTrack(i2.Control), false, false)
		if err != nil {
			return nil, err
		}
		if i2.AVType == VIDEO {
			if i2.Type == av.H264 {
				if len(i2.SpropParameterSets) > 1 {
					if codecData, err := h264parser.NewCodecDataFromSPSAndPPS(i2.SpropParameterSets[0], i2.SpropParameterSets[1]); err == nil {
						client.sps = i2.SpropParameterSets[0]
						client.pps = i2.SpropParameterSets[1]
						client.CodecData = append(client.CodecData, codecData)
					}
				} else {
					client.CodecData = append(client.CodecData, h264parser.CodecData{})
					client.WaitCodec = true
				}
				client.FPS = i2.FPS
				client.videoCodec = av.H264
			} else if i2.Type == av.H265 {
				if len(i2.SpropVPS) > 1 && len(i2.SpropSPS) > 1 && len(i2.SpropPPS) > 1 {
					if codecData, err := h265parser.NewCodecDataFromVPSAndSPSAndPPS(i2.SpropVPS, i2.SpropSPS, i2.SpropPPS); err == nil {
						client.vps = i2.SpropVPS
						client.sps = i2.SpropSPS
						client.pps = i2.SpropPPS
						client.CodecData = append(client.CodecData, codecData)
					}
				} else {
					client.CodecData = append(client.CodecData, h265parser.CodecData{})
				}
				client.videoCodec = av.H265

			} else {
				client.Println("SDP Video Codec Type Not Supported", i2.Type)
			}
			client.videoIDX = int8(len(client.CodecData) - 1)
			client.videoID = client.chTMP
		}
		if i2.AVType == AUDIO {
			client.audioID = client.chTMP
			var CodecData av.AudioCodecData
			switch i2.Type {
			case av.AAC:
				CodecData, err = aacparser.NewCodecDataFromMPEG4AudioConfigBytes(i2.Config)
				if err == nil {
					client.Println("Audio AAC bad config")
				}
			case av.OPUS:
				var cl av.ChannelLayout
				switch i2.ChannelCount {
				case 1:
					cl = av.CH_MONO
				case 2:
					cl = av.CH_STEREO
				default:
					cl = av.CH_MONO
				}
				CodecData = codec.NewOpusCodecData(i2.TimeScale, cl)
			case av.PCM_MULAW:
				CodecData = codec.NewPCMMulawCodecData()
			case av.PCM_ALAW:
				CodecData = codec.NewPCMAlawCodecData()
			case av.PCM:
				CodecData = codec.NewPCMCodecData()
			default:
				client.Println("Audio Codec", i2.Type, "not supported")
			}
			if CodecData != nil {
				client.CodecData = append(client.CodecData, CodecData)
				client.audioIDX = int8(len(client.CodecData) - 1)
				client.audioCodec = CodecData.Type()
				if i2.TimeScale != 0 {
					client.AudioTimeScale = int64(i2.TimeScale)
				}
			}
		}
		client.chTMP += 2
	}
	//test := map[string]string{"Scale": "1.000000", "Speed": "1.000000", "Range": "clock=20210929T210000Z-20210929T211000Z"}
	err = client.request(PLAY, nil, client.control, false, false)
	if err != nil {
		return nil, err
	}
	go client.startStream()
	return client, nil
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
			err := client.request(OPTIONS, map[string]string{"Require": "implicit-play"}, client.control, false, true)
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
			content[0] = header[0]
			content[1] = header[1]
			content[2] = header[2]
			content[3] = header[3]
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
		default:
			client.Println("RTSP Client RTP Read DeSync")
			return
		}
	}
}

func (client *RTSPClient) request(method string, customHeaders map[string]string, uri string, one bool, nores bool) (err error) {
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
				err = client.request(method, customHeaders, uri, true, false)
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
		if method == SETUP {
			//deep := stringInBetween(builder.String(), "interleaved=", ";")
			if val, ok := res["Transport"]; ok {
				splits2 := strings.Split(val, ";")
				for _, vs := range splits2 {
					if strings.Contains(vs, "interleaved") {
						splits3 := strings.Split(vs, "=")
						if len(splits3) == 2 {
							splits4 := strings.Split(splits3[1], "-")
							if len(splits4) == 2 {
								if val, err := strconv.Atoi(splits4[0]); err == nil {
									client.chTMP = val
								}
							}
						}
					}
				}
			}

		}
		client.Println(builder.String())
	}
	return
}

func (client *RTSPClient) Close() {
	if client.conn != nil {
		client.conn.SetDeadline(time.Now().Add(time.Second))
		client.request(TEARDOWN, nil, client.control, false, true)
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
	if l.Scheme != "rtsp" && l.Scheme != "rtsps" {
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

// Println mini logging functions
func (client *RTSPClient) Println(v ...interface{}) {
	if client.options.Debug {
		log.Println(v)
	}
}

// binSize
func binSize(val int) []byte {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(val))
	return buf
}

func isRTCPPacket(content []byte) bool {
	rtcpPacketType := content[5]
	return rtcpPacketType == RTCPSenderReport || rtcpPacketType == RTCPReceiverReport
}
