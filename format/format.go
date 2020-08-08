package format

import (
	"github.com/deepch/vdk/av/avutil"
	"github.com/deepch/vdk/format/aac"
	"github.com/deepch/vdk/format/flv"
	"github.com/deepch/vdk/format/mp4"
	"github.com/deepch/vdk/format/rtmp"
	"github.com/deepch/vdk/format/rtsp"
	"github.com/deepch/vdk/format/ts"
)

func RegisterAll() {
	avutil.DefaultHandlers.Add(mp4.Handler)
	avutil.DefaultHandlers.Add(ts.Handler)
	avutil.DefaultHandlers.Add(rtmp.Handler)
	avutil.DefaultHandlers.Add(rtsp.Handler)
	avutil.DefaultHandlers.Add(flv.Handler)
	avutil.DefaultHandlers.Add(aac.Handler)
}
