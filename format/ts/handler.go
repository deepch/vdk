package ts

import (
	"io"

	"github.com/honuworx/vdk/av"
	"github.com/honuworx/vdk/av/avutil"
)

func Handler(h *avutil.RegisterHandler) {
	h.Ext = ".ts"

	h.Probe = func(b []byte) bool {
		return b[0] == 0x47 && b[188] == 0x47
	}

	h.ReaderDemuxer = func(r io.Reader) av.Demuxer {
		return NewDemuxer(r)
	}

	h.WriterMuxer = func(w io.Writer) av.Muxer {
		return NewMuxer(w)
	}

	h.CodecTypes = CodecTypes
}
