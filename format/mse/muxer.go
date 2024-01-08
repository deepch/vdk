package mse

import (
	"github.com/deepch/vdk/av"
	"github.com/deepch/vdk/format/mp4f"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"net"
	"net/http"
)

var Debug bool

type Muxer struct {
	m    *mp4f.Muxer
	r    *http.Request
	w    http.ResponseWriter
	conn net.Conn
}

func NewMuxer(r *http.Request, w http.ResponseWriter) (*Muxer, error) {
	conn, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		return nil, err
	}
	go func() {
		defer func() {
			conn.Close()
		}()
		for {
			if _, _, err = wsutil.NextReader(conn, ws.StateServerSide); err != nil {
				return
			}
		}
	}()

	return &Muxer{
		conn: conn,
		m:    mp4f.NewMuxer(nil),
		r:    r,
		w:    w,
	}, nil
}

func (m *Muxer) WriteHeader(streams []av.CodecData) (err error) {
	if err = m.m.WriteHeader(streams); err != nil {
		return
	}
	meta, fist := m.m.GetInit(streams)
	if err = wsutil.WriteServerText(m.conn, []byte(meta)); err != nil {
		return
	}
	if err = wsutil.WriteServerBinary(m.conn, fist); err != nil {
		return
	}

	return
}

func (m *Muxer) WritePacket(pkt av.Packet) (err error) {
	gotFrame, buffer, err := m.m.WritePacket(pkt, false)
	if err != nil {
		return
	}
	if gotFrame {
		if err = wsutil.WriteServerBinary(m.conn, buffer); err != nil {
			return
		}
	}

	return
}

func (m *Muxer) WriteTrailer() (err error) {

	return m.conn.Close()
}
