package rtspv2

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

var Debug bool

type ProxyConn struct {
	URL      *url.URL
	netconn  net.Conn
	readbuf  []byte
	writebuf []byte
	sdp      []byte
	playing  bool
	options  bool
	cseq     int
	session  string
	protocol int
	in       int
}

type Proxy struct {
	Addr          string
	HandleConn    func(*ProxyConn)
	HandleOptions func(*ProxyConn)
	HandlePlay    func(*ProxyConn)
}

func NewProxyConn(netconn net.Conn) *ProxyConn {
	conn := &ProxyConn{}
	conn.netconn = netconn
	conn.writebuf = make([]byte, 4096)
	conn.readbuf = make([]byte, 4096)
	conn.session = uuid.New().String()
	return conn
}

func (self *ProxyConn) Close() (err error) {
	return nil
}

func (self *ProxyConn) WritePacket(pkt *[]byte) (err error) {
	err = self.netconn.SetDeadline(time.Now().Add(time.Second * 5))
	if err != nil {
		return err
	}
	_, err = self.netconn.Write(*pkt)
	if err != nil {
		return err
	}
	return nil
}

func (self *ProxyConn) WriteHeader(sdp []byte) {
	self.sdp = sdp
}

func (self *ProxyConn) NetConn() net.Conn {
	return self.netconn
}

func (self *Proxy) ListenAndServe() (err error) {
	addr := self.Addr
	if addr == "" {
		addr = ":554"
	}
	var tcpaddr *net.TCPAddr
	if tcpaddr, err = net.ResolveTCPAddr("tcp", addr); err != nil {
		err = fmt.Errorf("rtsp: ListenAndServe: %s", err)
		return
	}

	var listener *net.TCPListener
	if listener, err = net.ListenTCP("tcp", tcpaddr); err != nil {
		return
	}

	if Debug {
		fmt.Println("rtsp: server: listening on", addr)
	}

	for {
		var netconn net.Conn
		if netconn, err = listener.Accept(); err != nil {
			return
		}

		if Debug {
			fmt.Println("rtsp: server: accepted")
		}
		conn := NewProxyConn(netconn)
		go func() {
			err := self.handleConn(conn)
			if Debug {
				fmt.Println("rtsp: server: client closed err:", err)
			}
			//defer conn.Close()
		}()
	}
}

func (self *Proxy) handleConn(conn *ProxyConn) (err error) {
	if self.HandleConn != nil {
		self.HandleConn(conn)
	} else {
		for {
			if err = conn.prepare(); err != nil {
				return
			}
			if conn.options {
				if self.HandleOptions != nil {
					self.HandleOptions(conn)
				}
			}
			if conn.playing {
				if self.HandlePlay != nil {
					self.HandlePlay(conn)
				}
			}
		}
	}

	return
}

func (self *ProxyConn) prepare() error {

	self.options = false
	self.cseq++
	err := self.netconn.SetDeadline(time.Now().Add(time.Second * 5))
	if err != nil {
		return err
	}

	n, err := self.netconn.Read(self.readbuf)
	if err != nil {
		return err
	}
	allStringsSlice := strings.Split(string(self.readbuf[:n]), "\r\n")
	if len(allStringsSlice) == 0 {
		return errors.New("no cmd")
	}

	fistStringsSlice := strings.Split(allStringsSlice[0], " ")

	if len(fistStringsSlice) == 0 {
		return errors.New("no fist cmd")
	}

	cseq := strings.TrimSpace(stringInBetween(string(self.readbuf[:n]), "CSeq:", "\r\n"))
	switch fistStringsSlice[0] {
	case OPTIONS:

		if len(fistStringsSlice) < 2 {
			return errors.New("return bad OPTIONS")
		}
		if self.URL, err = url.Parse(fistStringsSlice[1]); err != nil {
			return err
		}
		_, err := self.netconn.Write([]byte("RTSP/1.0 200 OK\r\nPublic: OPTIONS, DESCRIBE, SETUP, PLAY\r\nSession: " + self.session + "\r\nCSeq: " + cseq + "\r\n\r\n"))
		if err != nil {
			return err
		}
		self.options = true

	case SETUP:
		if strings.Contains(string(self.readbuf[:n]), "RTP/AVP/UDP") {
			_, err := self.netconn.Write([]byte("RTSP/1.0 461 Unsupported transport\r\nCSeq: " + cseq + "\r\nSession: " + self.session + "\r\n\r\n"))
			if err != nil {
				return err
			}
			return nil
		}
		_, err := self.netconn.Write([]byte("RTSP/1.0 200 OK\r\nCSeq: " + cseq + "\r\nSession: " + self.session + "\r\nTransport: RTP/AVP/TCP;unicast;interleaved=" + strconv.Itoa(self.in) + "-" + strconv.Itoa(self.in+1) + "\r\n\r\n"))
		if err != nil {
			return err
		}
		self.in = self.in + 2
	case DESCRIBE:

		buf := "RTSP/1.0 200 OK\r\nContent-Type: application/sdp\r\nSession: " + self.session + "\r\nContent-Length: " + strconv.Itoa(len(self.sdp)) + "\r\nCSeq: " + cseq + "\r\n\r\n"
		_, err := self.netconn.Write([]byte(buf + string(self.sdp)))
		if err != nil {
			return err
		}

	case PLAY:
		_, err := self.netconn.Write([]byte("RTSP/1.0 200 OK\r\nSession: " + self.session + ";timeout=60\r\nCSeq: " + cseq + "\r\n\r\n"))
		if err != nil {
			return err
		}
		self.playing = true
	case TEARDOWN:
		self.netconn.Close()
		return errors.New("exit")

	default:

		return errors.New("metod not found " + fistStringsSlice[0])

	}
	return nil
}
