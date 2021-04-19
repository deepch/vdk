package fragment

import (
	"time"

	"github.com/deepch/vdk/av"
)

type Fragment struct {
	Bytes       []byte
	Length      int
	Independent bool
	Duration    time.Duration
}

type Fragmenter interface {
	av.PacketWriter
	Fragment() (Fragment, error)
	Duration() time.Duration
	TimeScale() uint32
	MovieHeader() (filename, contentType string, contents []byte)
	NewSegment()
}
