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
	"github.com/deepch/vdk/codec/h264parser"
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
	SETUP    = "SETUP"
)

type RTSPClient struct {
	control         string
	seq             int
	session         string
	realm           string
	nonce           string
	username        string
	password        string
	startVideoTS    int64
	startAudioTS    int64
	videoID         int
	audioID         int
	mediaSDP        []sdp.Media
	SDPRaw          []byte
	conn            net.Conn
	connRW          *bufio.ReadWriter
	pURL            *url.URL
	headers         map[string]string
	Signals         chan int
	OutgoingProxy   chan *[]byte
	OutgoingPacket  chan *av.Packet
	clientDigest    bool
	clientBasic     bool
	fuStarted       bool
	options         RTSPClientOptions
	BufferRtpPacket *bytes.Buffer
	sps             []byte
	pps             []byte
	CodecData       []av.CodecData
}

type RTSPClientOptions struct {
	Debug            bool
	URL              string
	DialTimeout      time.Duration
	ReadWriteTimeout time.Duration
	DisableAudio     bool
}

func Dial(options RTSPClientOptions) (*RTSPClient, error) {
	client := &RTSPClient{
		headers:         make(map[string]string),
		Signals:         make(chan int, 100),
		OutgoingProxy:   make(chan *[]byte, 3000),
		OutgoingPacket:  make(chan *av.Packet, 3000),
		BufferRtpPacket: bytes.NewBuffer([]byte{}),
		videoID:         0,
		audioID:         2,
		options:         options,
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
	err = client.request(OPTIONS, nil, client.pURL.String(), false, false)
	if err != nil {
		return nil, err
	}
	err = client.request(DESCRIBE, map[string]string{"Accept": "application/sdp"}, client.pURL.String(), false, false)
	if err != nil {
		return nil, err
	}
	var ch int
	for _, i2 := range client.mediaSDP {
		if (i2.AVType != VIDEO && i2.AVType != AUDIO) || (client.options.DisableAudio && i2.AVType == AUDIO) {
			continue
		}
		err = client.request(SETUP, map[string]string{"Transport": "RTP/AVP/TCP;unicast;interleaved=" + strconv.Itoa(ch) + "-" + strconv.Itoa(ch+1)}, client.ControlTrack(i2.Control), false, false)
		if err != nil {
			return nil, err
		}
		if i2.AVType == VIDEO {
			if i2.Type == av.H264 && len(i2.SpropParameterSets) > 1 {
				if codecData, err := h264parser.NewCodecDataFromSPSAndPPS(i2.SpropParameterSets[0], i2.SpropParameterSets[1]); err == nil {
					client.sps = i2.SpropParameterSets[0]
					client.pps = i2.SpropParameterSets[1]
					client.CodecData = append(client.CodecData, codecData)
				}
			} else {
				client.Println("SDP Video Codec Type Not Supported", i2.Type)
			}
			client.videoID = ch
		}
		if i2.AVType == AUDIO {
			client.audioID = ch
		}
		ch += 2
	}

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
			nb, err := io.ReadFull(client.conn, header)
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
			n, rerr := io.ReadFull(client.conn, content[4:length+4])
			if rerr != nil || n != int(length) {
				client.Println("RTSP Client RTP ReadFull", err)
				return
			}
			//atomic.AddInt64(&client.Bitrate, int64(length+4))
			if len(client.OutgoingProxy) < 2000 {
				client.OutgoingProxy <- &content
			} else {
				client.Println("RTSP Client OutgoingProxy Chanel Full")
				return
			}
			pkt, got := client.RTPDemuxer(&content)
			if !got {
				continue
			}
			for _, i2 := range pkt {
				if len(client.OutgoingPacket) > 2000 {
					client.Println("RTSP Client OutgoingPacket Chanel Full")
					return
				}
				client.OutgoingPacket <- i2
			}
		case 0x52:
			var responseTmp []byte
			for {
				n, rerr := io.ReadFull(client.conn, oneb)
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
						_, err = io.ReadFull(client.conn, cont)
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
		client.Println(builder.String())
	}
	return
}

func (client *RTSPClient) Close() {
	if client.conn != nil {
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
	padding := (firstByte>>5)&1 == 1
	extension := (firstByte>>4)&1 == 1
	CSRCCnt := int(firstByte & 0x0f)
	timestamp := int64(binary.BigEndian.Uint32(content[8:12]))
	offset := RTPHeaderSize

	end := len(content)
	if end-offset >= 4*CSRCCnt {
		offset += 4 * CSRCCnt
	}
	if extension && len(content) < 4+offset+2+2 {
		return nil, false
	}
	if extension && end-offset >= 4 {
		extLen := 4 * int(binary.BigEndian.Uint16(content[4+offset+2:]))
		offset += 4
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
		if client.BufferRtpPacket.Len() > 4048576 {
			client.Println("Big Buffer Flush")
			client.BufferRtpPacket.Truncate(0)
			client.BufferRtpPacket.Reset()
		}

		nalRaw, _ := h264parser.SplitNALUs(content[offset:end])
		var retmap []*av.Packet
		for _, nal := range nalRaw {
			naluType := nal[0] & 0x1f
			switch {
			case naluType >= 1 && naluType <= 5:
				retmap = append(retmap, &av.Packet{
					Data:            append(binSize(len(nal)), nal...),
					CompositionTime: time.Duration(1) * time.Millisecond,
					Idx:             0,
					IsKeyFrame:      naluType == 5,
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
						retmap = append(retmap, &av.Packet{
							Data:            append(binSize(client.BufferRtpPacket.Len()), client.BufferRtpPacket.Bytes()...),
							CompositionTime: time.Duration(1) * time.Millisecond,
							Idx:             0,
							IsKeyFrame:      naluTypef == 5,
							Time:            time.Duration(timestamp/90) * time.Millisecond,
						})
					}
				}

			default:
				client.Println("Unsupported NAL Type", naluType)
			}
		}
		if len(retmap) > 0 {
			return retmap, true
		}
	case client.audioID:
		client.Println("Audio Unsupported try report to https://github.com/deepch/vdk", padding, extension, timestamp)
	default:
		client.Println("Unsuported Intervaled data packet", int(content[1]), content[offset:end])
	}
	return nil, false
}

func (client *RTSPClient) CodecUpdateSPS(val []byte) {
	if bytes.Compare(val, client.sps) != 0 {
		if len(client.sps) > 0 && len(client.pps) > 0 {
			codecData, err := h264parser.NewCodecDataFromSPSAndPPS(val, client.pps)
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
		}
		client.Signals <- SignalCodecUpdate
		client.sps = val
	}
}

func (client *RTSPClient) CodecUpdatePPS(val []byte) {
	if bytes.Compare(val, client.pps) != 0 {
		if len(client.sps) > 0 && len(client.pps) > 0 {
			codecData, err := h264parser.NewCodecDataFromSPSAndPPS(client.sps, val)
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
		}
		client.Signals <- SignalCodecUpdate
		client.pps = val
	}
}

//Println mini logging functions
func (client *RTSPClient) Println(v ...interface{}) {
	if client.options.Debug {
		log.Println(v)
	}
}
func binSize(val int) []byte {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(val))
	return buf
}
