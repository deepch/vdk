package dvrip

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"html"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/deepch/vdk/av"
	"github.com/deepch/vdk/codec"
	"github.com/deepch/vdk/codec/h264parser"
)

const (
	SignalStreamStop = iota
	SignalCodecUpdate
)

type Client struct {
	conn                net.Conn
	login               string
	password            string
	host                string
	stream              string
	sequenceNumber      int32
	session             int32
	aliveInterval       time.Duration
	CodecData           []av.CodecData
	OutgoingPacketQueue chan *av.Packet
	Signals             chan int
	options             ClientOptions
	sps                 []byte
	pps                 []byte
}

type ClientOptions struct {
	Debug            bool
	URL              string
	DialTimeout      time.Duration
	ReadWriteTimeout time.Duration
	DisableAudio     bool
}

//Dial func
func Dial(options ClientOptions) (*Client, error) {
	client := &Client{
		Signals:             make(chan int, 100),
		OutgoingPacketQueue: make(chan *av.Packet, 3000),
		options:             options,
	}
	err := client.parseURL(html.UnescapeString(client.options.URL))
	if err != nil {
		return nil, err
	}
	client.conn, err = net.DialTimeout("tcp", client.host, time.Second*2)
	if err != nil {
		return nil, err
	}
	err = client.conn.SetDeadline(time.Now().Add(5 * time.Second))
	if err != nil {
		return nil, err
	}
	err = client.Login()
	if err != nil {
		return nil, err
	}
	err = client.SetTime()
	if err != nil {
		return nil, err
	}
	go client.Monitor()
	return client, nil
}

//Close func
func (client *Client) Close() error {
	err := client.conn.Close()
	return err
}

//SetKeepAlive func
func (client *Client) SetKeepAlive() error {
	body, err := json.Marshal(map[string]string{
		"Name":      "KeepAlive",
		"SessionID": fmt.Sprintf("0x%08X", client.session),
	})
	if err != nil {
		return err
	}
	err = client.send(codeKeepAlive, body)
	if err != nil {
		return err
	}
	return nil
}

//Monitor func
func (client *Client) Monitor() {
	defer func() {
		client.Signals <- SignalStreamStop
	}()
	_, _, err := client.Command(codeOPMonitor, map[string]interface{}{
		"Action": "Claim",
		"Parameter": map[string]interface{}{
			"Channel":    0,
			"CombinMode": "NONE",
			"StreamType": client.stream,
			"TransMode":  "TCP",
		},
	})
	if err != nil {
		return
	}
	payload, err := json.Marshal(map[string]interface{}{
		"Name":      "OPMonitor",
		"SessionID": fmt.Sprintf("0x%08X", client.session),
		"OPMonitor": map[string]interface{}{
			"Action": "Start",
			"Parameter": map[string]interface{}{
				"Channel":    0,
				"CombinMode": "NONE",
				"StreamType": client.stream,
				"TransMode":  "TCP",
			},
		},
	})
	err = client.send(1410, payload)
	if err != nil {
		return
	}
	var length uint32 = 0
	var dataType uint32
	timer := time.Now()
	var fps int
	for {
		if time.Now().Sub(timer).Milliseconds() > client.aliveInterval.Milliseconds() {
			err = client.SetKeepAlive()
			if err != nil {
				return
			}
			timer = time.Now()
		}
		_, body, err := client.recv(false)
		if err != nil {
			return
		}
		buf := bytes.NewReader(body)
		err = binary.Read(buf, binary.BigEndian, &dataType)
		if err != nil {
			return
		}
		switch dataType {
		case 0x1FC, 0x1FE:
			frame := struct {
				Media    byte
				FPS      byte
				Width    byte
				Height   byte
				DateTime uint32
				Length   uint32
			}{}
			err = binary.Read(buf, binary.LittleEndian, &frame)
			fps = int(frame.FPS)
			if err != nil {
				return
			}
			var packet bytes.Buffer
			if frame.Length > uint32(buf.Len()) {
				need := frame.Length - uint32(buf.Len())
				_, err = buf.WriteTo(&packet)
				if err != nil {
					return
				}
				_, err := client.recvSize(&packet, need)
				if err != nil {
					return
				}
			} else {
				_, err = buf.WriteTo(&packet)
				if err != nil {
					return
				}
			}
			if parseMediaType(dataType, frame.Media) == av.H264.String() {
				packets, _ := h264parser.SplitNALUs(packet.Bytes())
				for _, i2 := range packets {
					naluType := i2[0] & 0x1f
					switch {
					case naluType >= 1 && naluType <= 5:
						client.OutgoingPacketQueue <- &av.Packet{Duration: time.Duration(1000/fps) * time.Millisecond, Idx: 0, IsKeyFrame: naluType == 5, Data: append(binSize(len(i2)), i2...)}
					case naluType == 7:
						client.CodecUpdateSPS(i2)
					case naluType == 8:
						client.CodecUpdatePPS(i2)
					}
				}
			}
		case 0x1FD:
			err = binary.Read(buf, binary.LittleEndian, &length)
			if err != nil {
				return
			}
			var packet bytes.Buffer
			if length > uint32(buf.Len()) {
				need := length - uint32(buf.Len())
				_, err = buf.WriteTo(&packet)
				if err != nil {
					return
				}
				_, err := client.recvSize(&packet, need)
				if err != nil {
					return
				}
			} else {
				_, err = buf.WriteTo(&packet)
				if err != nil {
					return
				}
			}
			packets, _ := h264parser.SplitNALUs(packet.Bytes())
			for _, i2 := range packets {
				naluType := i2[0] & 0x1f
				switch {
				case naluType >= 1 && naluType <= 5:
					if fps != 0 {
						client.OutgoingPacketQueue <- &av.Packet{Duration: time.Duration(1000/fps) * time.Millisecond, Idx: 0, IsKeyFrame: naluType == 5, Data: append(binSize(len(i2)), i2...)}
					}
				case naluType == 7:
					client.CodecUpdateSPS(i2)
				case naluType == 8:
					client.CodecUpdatePPS(i2)
				}
			}
		case 0x1FA, 0x1F9:
			if client.options.DisableAudio {
				continue
			}
			frame := struct {
				Media      byte
				SampleRate byte
				Length     uint16
			}{}
			err = binary.Read(buf, binary.LittleEndian, &frame)
			if err != nil {
				return
			}
			var packet bytes.Buffer
			if uint32(frame.Length) > uint32(buf.Len()) {
				need := uint32(frame.Length) - uint32(buf.Len())
				_, err = buf.WriteTo(&packet)
				if err != nil {
					return
				}
				_, err := client.recvSize(&packet, need)
				if err != nil {
					return
				}
			} else {
				_, err = buf.WriteTo(&packet)
				if err != nil {
					return
				}
			}
			if parseMediaType(dataType, frame.Media) == av.PCM_ALAW.String() {
				if client.CodecData != nil {
					if len(client.CodecData) == 1 {
						client.CodecUpdatePCMAlaw()
					}
					client.OutgoingPacketQueue <- &av.Packet{Duration: time.Duration(8000/packet.Len()) * time.Millisecond, Idx: 1, Data: packet.Bytes()}
				}
			}
		case 0xFFD8FFE0:
		default:
			continue
		}
	}
}

func (client *Client) SetTime() error {
	_, _, err := client.Command(codeOPTimeSetting, time.Now().Format("2006-01-02 15:04:05"))
	return err
}
func (client *Client) Login() error {
	body, err := json.Marshal(map[string]string{
		"EncryptType": "MD5",
		"LoginType":   "DVRIP-WEB",
		"PassWord":    sofiaHash(client.password),
		"UserName":    client.login,
	})
	if err != nil {
		return err
	}
	err = client.send(codeLogin, body)
	if err != nil {
		return err
	}
	_, resp, err := client.recv(true)
	if err != nil {
		return err
	}
	res := LoginResp{}
	err = json.Unmarshal(resp, &res)
	if err != nil {
		return err
	}
	if (statusCode(res.Ret) != statusOK) && (statusCode(res.Ret) != statusUpgradeSuccessful) {
		return fmt.Errorf("unexpected status code: %v - %v", res.Ret, statusCodes[statusCode(res.Ret)])
	}
	client.aliveInterval = time.Duration(res.AliveInterval) * time.Second
	session, err := strconv.ParseUint(res.SessionID, 0, 32)
	if err != nil {
		return err
	}
	client.session = int32(session)
	return nil
}

//Command func
func (client *Client) Command(command requestCode, data interface{}) (*Payload, []byte, error) {
	params, err := json.Marshal(map[string]interface{}{
		"Name":                requestCodes[command],
		"SessionID":           fmt.Sprintf("0x%08X", client.session),
		requestCodes[command]: data,
	})
	if err != nil {
		return nil, nil, err
	}
	err = client.send(command, params)
	if err != nil {
		return nil, nil, err
	}
	resp, body, err := client.recv(true)
	return resp, body, err
}

//send func
func (client *Client) send(msgID requestCode, data []byte) error {
	var buf bytes.Buffer
	if err := binary.Write(&buf, binary.LittleEndian, Payload{
		Head:           255,
		Version:        0,
		Session:        client.session,
		SequenceNumber: client.sequenceNumber,
		MsgID:          int16(msgID),
		BodyLength:     int32(len(data)) + 2,
	}); err != nil {
		return err
	}
	err := client.conn.SetDeadline(time.Now().Add(5 * time.Second))
	if err != nil {
		return err
	}
	err = binary.Write(&buf, binary.LittleEndian, data)
	if err != nil {
		return err
	}
	err = binary.Write(&buf, binary.LittleEndian, magicEnd)
	if err != nil {
		return err
	}
	_, err = client.conn.Write(buf.Bytes())
	if err != nil {
		return err
	}
	client.sequenceNumber++
	return nil
}

//recvSize func
func (client *Client) recvSize(buffer *bytes.Buffer, size uint32) ([]byte, error) {
	all := uint32(0)
	for {
		_, body, err := client.recv(false)
		if err != nil {
			return nil, err
		}
		all += uint32(len(body))
		buffer.Write(body)
		if all == size {
			break
		} else if all > size {
			return nil, fmt.Errorf("invalid read size")
		}
	}
	return nil, nil
}

//recv func
func (client *Client) recv(text bool) (*Payload, []byte, error) {
	var p Payload
	var b = make([]byte, 20)
	err := client.conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	if err != nil {
		return nil, nil, err
	}
	_, err = client.conn.Read(b)
	if err != nil {
		return nil, nil, err
	}
	err = binary.Read(bytes.NewReader(b), binary.LittleEndian, &p)
	if err != nil {
		return nil, nil, err
	}
	client.sequenceNumber += 1
	if p.BodyLength <= 0 || p.BodyLength >= 100000 {
		return nil, nil, fmt.Errorf("invalid bodylength: %v", p.BodyLength)
	}
	body := make([]byte, p.BodyLength)
	err = binary.Read(client.conn, binary.LittleEndian, &body)
	if err != nil {
		return nil, nil, err
	}
	if text && len(body) > 2 && bytes.Compare(body[len(body)-2:], []byte{10, 0}) == 0 {
		body = body[:len(body)-2]
	}
	return &p, body, nil
}

//parseURL func
func (client *Client) parseURL(rawURL string) error {
	l, err := url.Parse(rawURL)
	if err != nil {
		return err
	}
	username := l.User.Username()
	password, _ := l.User.Password()
	l.User = nil
	if l.Port() == "" {
		l.Host = fmt.Sprintf("%s:%s", l.Host, "34567")
	}
	if username == "" {
		username = "admin"
	}
	if password == "" {
		password = "admin"
	}
	client.login = username
	client.password = password
	client.host = l.Host
	client.stream = strings.Trim(l.EscapedPath(), "/")
	return nil
}

func (client *Client) CodecUpdateSPS(val []byte) {
	if bytes.Compare(val, client.sps) == 0 {
		return
	}
	client.sps = val
	if len(client.pps) == 0 {
		return
	}
	codecData, err := h264parser.NewCodecDataFromSPSAndPPS(val, client.pps)
	if err != nil {
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

func (client *Client) CodecUpdatePPS(val []byte) {
	if bytes.Compare(val, client.pps) == 0 {
		return
	}
	client.pps = val
	if len(client.sps) == 0 {
		return
	}
	codecData, err := h264parser.NewCodecDataFromSPSAndPPS(client.sps, val)
	if err != nil {
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

func (client *Client) CodecUpdatePCMAlaw() {
	CodecData := codec.NewPCMAlawCodecData()
	client.CodecData = append(client.CodecData, CodecData)
	client.Signals <- SignalCodecUpdate
}
