package format

import (
	"github.com/honuworx/vdk/av/avutil"
	"github.com/honuworx/vdk/format/aac"
	"github.com/honuworx/vdk/format/flv"
	"github.com/honuworx/vdk/format/mp4"
	"github.com/honuworx/vdk/format/rtmp"
	"github.com/honuworx/vdk/format/rtsp"
	"github.com/honuworx/vdk/format/ts"
)

func RegisterAll() {
	avutil.DefaultHandlers.Add(mp4.Handler)
	avutil.DefaultHandlers.Add(ts.Handler)
	avutil.DefaultHandlers.Add(rtmp.Handler)
	avutil.DefaultHandlers.Add(rtsp.Handler)
	avutil.DefaultHandlers.Add(flv.Handler)
	avutil.DefaultHandlers.Add(aac.Handler)
}
