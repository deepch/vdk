package mp4f

import (
	"time"

	"github.com/deepch/vdk/av"
	"github.com/deepch/vdk/format/mp4"
	"github.com/deepch/vdk/format/mp4/mp4io"
	"github.com/deepch/vdk/format/mp4f/mp4fio"
)

type Stream struct {
	av.CodecData
	codecString            string
	trackAtom              *mp4io.Track
	idx                    int
	lastpkt                *av.Packet
	timeScale              int64
	duration               int64
	muxer                  *Muxer
	demuxer                *mp4.Demuxer
	sample                 *mp4io.SampleTable
	sampleIndex            int
	sampleOffsetInChunk    int64
	syncSampleIndex        int
	dts                    int64
	sttsEntryIndex         int
	sampleIndexInSttsEntry int
	cttsEntryIndex         int
	sampleIndexInCttsEntry int
	chunkGroupIndex        int
	chunkIndex             int
	sampleIndexInChunk     int
	sttsEntry              *mp4io.TimeToSampleEntry
	cttsEntry              *mp4io.CompositionOffsetEntry
	moof                   mp4fio.MovieFrag
	buffer                 []byte
}

func timeToTs(tm time.Duration, timeScale int64) int64 {
	return int64(tm * time.Duration(timeScale) / time.Second)
}

func tsToTime(ts int64, timeScale int64) time.Duration {
	return time.Duration(ts) * time.Second / time.Duration(timeScale)
}

func (obj *Stream) timeToTs(tm time.Duration) int64 {
	return int64(tm * time.Duration(obj.timeScale) / time.Second)
}

func (obj *Stream) tsToTime(ts int64) time.Duration {
	return time.Duration(ts) * time.Second / time.Duration(obj.timeScale)
}
