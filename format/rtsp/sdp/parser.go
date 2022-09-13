package sdp

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/deepch/vdk/av"
)

type Session struct {
	Uri string
}

type Media struct {
	AVType             string
	Type               av.CodecType
	FPS                int
	TimeScale          int
	Control            string
	Rtpmap             int
	ChannelCount       int
	Config             []byte
	SpropParameterSets [][]byte
	SpropVPS           []byte
	SpropSPS           []byte
	SpropPPS           []byte
	PayloadType        int
	SizeLength         int
	IndexLength        int
}

func Parse(content string) (sess Session, medias []Media) {
	var media *Media

	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		////Camera [BUG] a=x-framerate: 25
		if strings.Contains(line, "x-framerate") {
			line = strings.Replace(line, " ", "", -1)
		}
		typeval := strings.SplitN(line, "=", 2)
		if len(typeval) == 2 {
			fields := strings.SplitN(typeval[1], " ", 2)

			switch typeval[0] {
			case "m":
				if len(fields) > 0 {
					switch fields[0] {
					case "audio", "video":
						medias = append(medias, Media{AVType: fields[0]})
						media = &medias[len(medias)-1]
						mfields := strings.Split(fields[1], " ")
						if len(mfields) >= 3 {
							media.PayloadType, _ = strconv.Atoi(mfields[2])
						}
						switch media.PayloadType {
						case 0:
							media.Type = av.PCM_MULAW
						case 8:
							media.Type = av.PCM_ALAW
						}
					default:
						media = nil
					}
				}

			case "u":
				sess.Uri = typeval[1]

			case "a":
				if media != nil {
					for _, field := range fields {
						keyval := strings.SplitN(field, ":", 2)
						if len(keyval) >= 2 {
							key := keyval[0]
							val := keyval[1]
							switch key {
							case "control":
								media.Control = val
							case "rtpmap":
								media.Rtpmap, _ = strconv.Atoi(val)
							case "x-framerate":
								media.FPS, _ = strconv.Atoi(val)
							}
						}
						keyval = strings.Split(field, "/")
						if len(keyval) >= 2 {
							key := keyval[0]
							switch strings.ToUpper(key) {
							case "MPEG4-GENERIC":
								media.Type = av.AAC
							case "L16":
								media.Type = av.PCM
							case "OPUS":
								media.Type = av.OPUS
								if len(keyval) > 2 {
									if i, err := strconv.Atoi(keyval[2]); err == nil {
										media.ChannelCount = i
									}
								}
							case "H264":
								media.Type = av.H264
							case "JPEG":
								media.Type = av.JPEG
							case "H265":
								media.Type = av.H265
							case "HEVC":
								media.Type = av.H265
							case "PCMA":
								media.Type = av.PCM_ALAW
							case "PCMU":
								media.Type = av.PCM_MULAW
							}
							if i, err := strconv.Atoi(keyval[1]); err == nil {
								media.TimeScale = i
							}
							if false {
								fmt.Println("sdp:", keyval[1], media.TimeScale)
							}
						}
						keyval = strings.Split(field, ";")
						if len(keyval) > 1 {
							for _, field := range keyval {
								keyval := strings.SplitN(field, "=", 2)
								if len(keyval) == 2 {
									key := strings.TrimSpace(keyval[0])
									val := keyval[1]
									switch key {
									case "config":
										media.Config, _ = hex.DecodeString(val)
									case "sizelength":
										media.SizeLength, _ = strconv.Atoi(val)
									case "indexlength":
										media.IndexLength, _ = strconv.Atoi(val)
									case "sprop-vps":
										val, err := base64.StdEncoding.DecodeString(val)
										if err == nil {
											media.SpropVPS = val
										} else {
											log.Println("SDP: decode vps error", err)
										}
									case "sprop-sps":
										val, err := base64.StdEncoding.DecodeString(val)
										if err == nil {
											media.SpropSPS = val
										} else {
											log.Println("SDP: decode sps error", err)
										}
									case "sprop-pps":
										val, err := base64.StdEncoding.DecodeString(val)
										if err == nil {
											media.SpropPPS = val
										} else {
											log.Println("SDP: decode pps error", err)
										}
									case "sprop-parameter-sets":
										fields := strings.Split(val, ",")
										for _, field := range fields {
											if field == "" {
												continue
											}
											val, _ := base64.StdEncoding.DecodeString(field)
											media.SpropParameterSets = append(media.SpropParameterSets, val)
										}
									}
								}
							}
						}
					}
				}

			}

		}
	}
	return
}
