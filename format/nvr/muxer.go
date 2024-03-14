package nvr

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"errors"
	"fmt"
	"github.com/deepch/vdk/av"
	"github.com/deepch/vdk/codec/aacparser"
	"github.com/deepch/vdk/codec/h264parser"
	"github.com/deepch/vdk/format/mp4"
	"github.com/google/uuid"
	"github.com/moby/sys/mountinfo"
	"github.com/shirou/gopsutil/v3/disk"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var MIME = []byte{11, 22, 111, 222, 11, 22, 111, 222}

var listTag = []string{"{server_id}", "{host_name}", "{host_name_short}", "{host_name_long}",
	"{stream_name}", "{channel_name}", "{stream_id}", "{channel_id}",
	"{start_year}", "{start_month}", "{start_day}", "{start_hour}", "{start_minute}", "{start_second}",
	"{start_millisecond}", "{start_unix_second}", "{start_unix_millisecond}", "{start_time}", "{start_pts}",
	"{end_year}", "{end_month}", "{end_day}", "{end_hour}", "{end_minute}", "{end_second}",
	"{end_millisecond}", "{end_unix_millisecond}", "{end_unix_second}", "{end_time}", "{end_pts}", "{duration_second}", "{duration_millisecond}"}

const (
	MP4 = "mp4"
	NVR = "nvr"
)

type Muxer struct {
	muxer                                                                       *mp4.Muxer
	format                                                                      string
	limit                                                                       int
	d                                                                           *os.File
	m                                                                           *os.File
	dur                                                                         time.Duration
	h                                                                           int
	gof                                                                         *Gof
	patch                                                                       string
	mpoint                                                                      []string
	start, end                                                                  time.Time
	pstart, pend                                                                time.Duration
	started                                                                     bool
	serverID, streamName, channelName, streamID, channelID, hostLong, hostShort string
}

type Gof struct {
	Streams []av.CodecData
	Packet  []av.Packet
}

type Data struct {
	Time  int64
	Start int64
	Dur   int64
}

const (
	B  = 1
	KB = 1024 * B
	MB = 1024 * KB
	GB = 1024 * MB
)

func init() {
	gob.RegisterName("nvr.Gof", Gof{})
	gob.RegisterName("h264parser.CodecData", h264parser.CodecData{})
	gob.RegisterName("aacparser.CodecData", aacparser.CodecData{})

}

func NewMuxer(serverID, streamName, channelName, streamID, channelID string, mpoint []string, patch, format string, limit int) (m *Muxer, err error) {
	hostLong, _ := os.Hostname()
	var hostShort string
	if p, _, ok := strings.Cut(hostLong, "."); ok {
		hostShort = p
	}
	m = &Muxer{
		mpoint:      mpoint,
		patch:       patch,
		h:           -1,
		gof:         &Gof{},
		format:      format,
		limit:       limit,
		serverID:    serverID,
		streamName:  streamName,
		channelName: channelName,
		streamID:    streamID,
		channelID:   channelID,
		hostLong:    hostLong,
		hostShort:   hostShort,
	}
	return
}

func (m *Muxer) WriteHeader(streams []av.CodecData) (err error) {
	m.gof.Streams = streams
	if m.format == MP4 {
		return m.OpenMP4()
	}

	return
}

func (m *Muxer) WritePacket(pkt av.Packet) (err error) {
	if len(m.gof.Streams) == 0 {
		return
	}
	if !m.started && pkt.IsKeyFrame {
		m.started = true
	}
	if m.started {
		switch m.format {
		case MP4:
			return m.writePacketMP4(pkt)
		case NVR:
			return m.writePacketNVR(pkt)
		}
	}

	return
}

func (m *Muxer) writePacketMP4(pkt av.Packet) (err error) {
	if pkt.IsKeyFrame && m.dur > time.Duration(m.limit)*time.Second {
		m.pstart = pkt.Time
		if err = m.OpenMP4(); err != nil {
			return
		}
		m.dur = 0

	}
	m.dur += pkt.Duration
	m.pend = pkt.Time

	return m.muxer.WritePacket(pkt)
}

func (m *Muxer) writePacketNVR(pkt av.Packet) (err error) {
	if pkt.IsKeyFrame {
		if len(m.gof.Packet) > 0 {
			if err = m.writeGop(); err != nil {
				return
			}
		}
		m.gof.Packet, m.dur = nil, 0
	}
	if pkt.Idx == 0 {
		m.dur += pkt.Duration
	}
	m.gof.Packet = append(m.gof.Packet, pkt)

	return
}

func (m *Muxer) writeGop() (err error) {
	t := time.Now().UTC()
	if m.h != t.Hour() {
		if err = m.OpenNVR(); err != nil {
			return
		}
	}
	f := Data{
		Time: t.UnixNano(),
		Dur:  m.dur.Milliseconds(),
	}
	if f.Start, err = m.d.Seek(0, 2); err != nil {
		return
	}
	enc := gob.NewEncoder(m.d)
	if err = enc.Encode(m.gof); err != nil {
		return
	}
	buf := bytes.NewBuffer([]byte{})
	if err = binary.Write(buf, binary.LittleEndian, f); err != nil {
		return
	}
	if _, err = buf.Write(MIME); err != nil {
		return
	}
	_, err = m.m.Write(buf.Bytes())

	return
}

func (m *Muxer) OpenNVR() (err error) {
	m.WriteTrailer()
	t := time.Now().UTC()
	if err = os.MkdirAll(fmt.Sprintf("%s/%s", m.patch, t.Format("2006/01/02")), 0755); err != nil {
		return
	}
	if m.d, err = os.OpenFile(fmt.Sprintf("%s/%s/%d.d", m.patch, t.Format("2006/01/02"), t.Hour()), os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0660); err != nil {
		return
	}
	if m.m, err = os.OpenFile(fmt.Sprintf("%s/%s/%d.m", m.patch, t.Format("2006/01/02"), t.Hour()), os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0660); err != nil {
		return
	}
	m.h = t.Hour()

	return
}

func (m *Muxer) OpenMP4() (err error) {
	m.WriteTrailer()
	m.start = time.Now().UTC()

	d, err := m.filePatch()
	if err != nil {
		return
	}
	if err = os.MkdirAll(filepath.Dir(d), 0755); err != nil {
		return
	}
	if m.d, err = os.Create(filepath.Join(filepath.Dir(d), fmt.Sprintf("tmp_%s_%d.mp4", uuid.New(), time.Now().Unix()))); err != nil {
		return
	}
	m.muxer = mp4.NewMuxer(m.d)
	m.muxer.NegativeTsMakeZero = true
	if err = m.muxer.WriteHeader(m.gof.Streams); err != nil {
		return
	}

	return
}

func (m *Muxer) filePatch() (string, error) {
	var (
		mu = float64(100)
		ui = -1
	)

	for i, i2 := range m.mpoint {
		if m, err := mountinfo.Mounted(i2); err == nil && m {
			if d, err := disk.Usage(i2); err == nil {
				if d.UsedPercent < mu {
					ui = i
					mu = d.UsedPercent
				}
			}
		}
	}

	if ui == -1 {
		return "", errors.New("not mount ready")
	}

	ts := filepath.Join(m.mpoint[ui], m.patch)
	m.end = time.Now().UTC()

	for _, s := range listTag {
		switch s {
		case "{server_id}":
			ts = strings.Replace(ts, "{server_id}", m.serverID, -1)
		case "{host_name}":
			ts = strings.Replace(ts, "{host_name}", m.hostLong, -1)
		case "{host_name_short}":
			ts = strings.Replace(ts, "{host_name_short}", m.hostShort, -1)
		case "{host_name_long}":
			ts = strings.Replace(ts, "{host_name_long}", m.hostLong, -1)
		case "{stream_name}":
			ts = strings.Replace(ts, "{stream_name}", m.streamName, -1)
		case "{channel_name}":
			ts = strings.Replace(ts, "{channel_name}", m.channelName, -1)
		case "{stream_id}":
			ts = strings.Replace(ts, "{stream_id}", m.streamID, -1)
		case "{channel_id}":
			ts = strings.Replace(ts, "{channel_id}", m.channelID, -1)
		case "{start_year}":
			ts = strings.Replace(ts, "{start_year}", fmt.Sprintf("%d", m.start.Year()), -1)
		case "{start_month}":
			ts = strings.Replace(ts, "{start_month}", fmt.Sprintf("%02d", int(m.start.Month())), -1)
		case "{start_day}":
			ts = strings.Replace(ts, "{start_day}", fmt.Sprintf("%02d", m.start.Day()), -1)
		case "{start_hour}":
			ts = strings.Replace(ts, "{start_hour}", fmt.Sprintf("%02d", m.start.Hour()), -1)
		case "{start_minute}":
			ts = strings.Replace(ts, "{start_minute}", fmt.Sprintf("%02d", m.start.Minute()), -1)
		case "{start_second}":
			ts = strings.Replace(ts, "{start_second}", fmt.Sprintf("%02d", m.start.Second()), -1)
		case "{start_millisecond}":
			ts = strings.Replace(ts, "{start_millisecond}", fmt.Sprintf("%d", m.start.Nanosecond()/1000/1000), -1)
		case "{start_unix_millisecond}":
			ts = strings.Replace(ts, "{start_unix_millisecond}", fmt.Sprintf("%d", m.start.UnixMilli()), -1)
		case "{start_unix_second}":
			ts = strings.Replace(ts, "{start_unix_second}", fmt.Sprintf("%d", m.start.Unix()), -1)
		case "{start_time}":
			ts = strings.Replace(ts, "{start_time}", fmt.Sprintf("%s", m.start.Format("2006-01-02T15:04:05-0700")), -1)
		case "{start_pts}":
			ts = strings.Replace(ts, "{start_pts}", fmt.Sprintf("%d", m.pstart.Milliseconds()), -1)
		case "{end_year}":
			ts = strings.Replace(ts, "{end_year}", fmt.Sprintf("%d", m.end.Year()), -1)
		case "{end_month}":
			ts = strings.Replace(ts, "{end_month}", fmt.Sprintf("%02d", int(m.end.Month())), -1)
		case "{end_day}":
			ts = strings.Replace(ts, "{end_day}", fmt.Sprintf("%02d", m.end.Day()), -1)
		case "{end_hour}":
			ts = strings.Replace(ts, "{end_hour}", fmt.Sprintf("%02d", m.end.Hour()), -1)
		case "{end_minute}":
			ts = strings.Replace(ts, "{end_minute}", fmt.Sprintf("%02d", m.end.Minute()), -1)
		case "{end_second}":
			ts = strings.Replace(ts, "{end_second}", fmt.Sprintf("%02d", m.end.Second()), -1)
		case "{end_millisecond}":
			ts = strings.Replace(ts, "{end_millisecond}", fmt.Sprintf("%d", m.end.Nanosecond()/1000/1000), -1)
		case "{end_unix_millisecond}":
			ts = strings.Replace(ts, "{end_unix_millisecond}", fmt.Sprintf("%d", m.end.UnixMilli()), -1)
		case "{end_unix_second}":
			ts = strings.Replace(ts, "{end_unix_second}", fmt.Sprintf("%d", m.end.Unix()), -1)
		case "{end_time}":
			ts = strings.Replace(ts, "{end_time}", fmt.Sprintf("%s", m.end.Format("2006-01-02T15:04:05-0700")), -1)
		case "{end_pts}":
			ts = strings.Replace(ts, "{end_pts}", fmt.Sprintf("%d", m.pend.Milliseconds()), -1)
		case "{duration_second}":
			ts = strings.Replace(ts, "{duration_second}", fmt.Sprintf("%f", m.dur.Seconds()), -1)
		case "{duration_millisecond}":
			ts = strings.Replace(ts, "{duration_millisecond}", fmt.Sprintf("%d", m.dur.Milliseconds()), -1)
		}
	}

	return ts, nil
}

func (m *Muxer) WriteTrailer() (err error) {
	if m.muxer != nil {
		m.muxer.WriteTrailer()
	}
	if m.m != nil {
		err = m.m.Close()
	}
	if m.d != nil {
		if m.format == MP4 {
			p, err := m.filePatch()
			if err != nil {
				return err
			}
			if err = os.Rename(m.d.Name(), filepath.Join(filepath.Dir(m.d.Name()), filepath.Base(p))); err != nil {
				return err
			}
		}
		err = m.d.Close()
	}

	return
}
