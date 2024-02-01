package mp4f

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/deepch/vdk/av"
	"github.com/deepch/vdk/codec/aacparser"
	"github.com/deepch/vdk/codec/h264parser"
	"github.com/deepch/vdk/codec/h265parser"
	"github.com/deepch/vdk/format/fmp4/fmp4io"
	"github.com/deepch/vdk/format/mp4/mp4io"
	"github.com/deepch/vdk/format/mp4f/mp4fio"
	"github.com/deepch/vdk/utils/bits/pio"
)

type Muxer struct {
	maxFrames     int
	bufw          *bufio.Writer
	wpos          int64
	fragmentIndex int
	streams       []*Stream
	path          string
}

func NewMuxer(w *os.File) *Muxer {
	return &Muxer{}
}
func (self *Muxer) SetPath(path string) {
	self.path = path
}
func (self *Muxer) SetMaxFrames(count int) {
	self.maxFrames = count
}
func (self *Muxer) newStream(codec av.CodecData) (err error) {
	switch codec.Type() {
	case av.H264, av.H265, av.AAC:
	default:
		err = fmt.Errorf("fmp4: codec type=%v is not supported", codec.Type())
		return
	}
	stream := &Stream{CodecData: codec}

	stream.sample = &mp4io.SampleTable{
		SampleDesc:    &mp4io.SampleDesc{},
		TimeToSample:  &mp4io.TimeToSample{},
		SampleToChunk: &mp4io.SampleToChunk{},
		SampleSize:    &mp4io.SampleSize{},
		ChunkOffset:   &mp4io.ChunkOffset{},
	}

	stream.trackAtom = &mp4io.Track{
		Header: &mp4io.TrackHeader{
			TrackId:  int32(len(self.streams) + 1),
			Flags:    0x0007,
			Duration: 0,
			Matrix:   [9]int32{0x10000, 0, 0, 0, 0x10000, 0, 0, 0, 0x40000000},
		},
		Media: &mp4io.Media{
			Header: &mp4io.MediaHeader{
				TimeScale: 1000,
				Duration:  0,
				Language:  21956,
			},
			Info: &mp4io.MediaInfo{
				Sample: stream.sample,
				Data: &mp4io.DataInfo{
					Refer: &mp4io.DataRefer{
						Url: &mp4io.DataReferUrl{
							Flags: 0x000001,
						},
					},
				},
			},
		},
	}
	switch codec.Type() {
	case av.H264:
		stream.sample.SyncSample = &mp4io.SyncSample{}
		stream.timeScale = 90000
	case av.H265:
		stream.sample.SyncSample = &mp4io.SyncSample{}
		stream.timeScale = 90000
	case av.AAC:
		stream.timeScale = int64(codec.(av.AudioCodecData).SampleRate())
	}

	stream.muxer = self
	self.streams = append(self.streams, stream)

	return
}

func (self *Stream) buildEsds(conf []byte) *FDummy {
	esds := &mp4fio.ElemStreamDesc{DecConfig: conf}

	b := make([]byte, esds.Len())
	esds.Marshal(b)

	esdsDummy := FDummy{
		Data: b,
		Tag_: mp4io.Tag(uint32(mp4io.ESDS)),
	}
	return &esdsDummy
}

func (self *Stream) buildHdlr() *FDummy {
	hdlr := FDummy{
		Data: []byte{
			0x00, 0x00, 0x00, 0x35, 0x68, 0x64, 0x6C, 0x72,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x76, 0x69,
			0x64, 0x65, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x47, 0x6F, 0x6F, 0x64, 0x67, 0x61,
			0x6D, 0x65, 0x20, 0x47, 0x4F, 0x20, 0x53, 0x65, 0x72, 0x76,
			0x65, 0x72, 0x00, 0x00, 0x00},

		Tag_: mp4io.Tag(uint32(mp4io.HDLR)),
	}
	return &hdlr
}

func (self *Stream) buildAudioHdlr() *FDummy {
	hdlr := FDummy{
		Data: []byte{
			0x00, 0x00, 0x00, 0x35, 0x68, 0x64, 0x6C, 0x72,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x73, 0x6F,
			0x75, 0x6E, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x42, 0x65, 0x6E, 0x74, 0x6F, 0x34,
			0x20, 0x53, 0x6F, 0x75, 0x6E, 0x64, 0x20, 0x48, 0x61, 0x6E,
			0x64, 0x6C, 0x65, 0x72, 0x00},

		Tag_: mp4io.Tag(uint32(mp4io.HDLR)),
	}
	return &hdlr
}

func (self *Stream) buildEdts() *FDummy {
	edts := FDummy{
		Data: []byte{
			0x00, 0x00, 0x00, 0x30, 0x65, 0x64, 0x74, 0x73,
			0x00, 0x00, 0x00, 0x28, 0x65, 0x6C, 0x73, 0x74, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x21,
			0xFF, 0xFF, 0xFF, 0xFF, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00,
			0x9B, 0x24, 0x00, 0x00, 0x02, 0x10, 0x00, 0x01, 0x00, 0x00,
		},
		Tag_: mp4io.Tag(0x65647473),
	}
	return &edts
}

func (self *Stream) fillTrackAtom() (err error) {
	self.trackAtom.Media.Header.TimeScale = int32(self.timeScale)
	self.trackAtom.Media.Header.Duration = int32(self.duration)

	if self.Type() == av.H264 {
		codec := self.CodecData.(h264parser.CodecData)
		width, height := codec.Width(), codec.Height()
		self.sample.SampleDesc.AVC1Desc = &mp4io.AVC1Desc{
			DataRefIdx:           1,
			HorizontalResolution: 72,
			VorizontalResolution: 72,
			Width:                int16(width),
			Height:               int16(height),
			FrameCount:           1,
			Depth:                24,
			ColorTableId:         -1,
			Conf:                 &mp4io.AVC1Conf{Data: codec.AVCDecoderConfRecordBytes()},
		}
		self.trackAtom.Header.TrackWidth = float64(width)
		self.trackAtom.Header.TrackHeight = float64(height)

		self.trackAtom.Media.Handler = &mp4io.HandlerRefer{
			SubType: [4]byte{'v', 'i', 'd', 'e'},
			Name:    []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 'G', 'G', 0, 0, 0},
		}
		self.trackAtom.Media.Info.Video = &mp4io.VideoMediaInfo{
			Flags: 0x000001,
		}
		self.codecString = fmt.Sprintf("avc1.%02X%02X%02X", codec.RecordInfo.AVCProfileIndication, codec.RecordInfo.ProfileCompatibility, codec.RecordInfo.AVCLevelIndication)
	} else if self.Type() == av.H265 {
		codec := self.CodecData.(h265parser.CodecData)
		width, height := codec.Width(), codec.Height()

		self.sample.SampleDesc.HV1Desc = &mp4io.HV1Desc{
			DataRefIdx:           1,
			HorizontalResolution: 72,
			VorizontalResolution: 72,
			Width:                int16(width),
			Height:               int16(height),
			FrameCount:           1,
			Depth:                24,
			ColorTableId:         -1,
			Conf:                 &mp4io.HV1Conf{Data: codec.AVCDecoderConfRecordBytes()},
		}

		self.trackAtom.Media.Handler = &mp4io.HandlerRefer{
			SubType: [4]byte{'v', 'i', 'd', 'e'},
			Name:    []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 'G', 'G', 0, 0, 0},
		}
		self.trackAtom.Media.Info.Video = &mp4io.VideoMediaInfo{
			Flags: 0x000001,
		}
		//self.codecString = fmt.Sprintf("hvc1.%02X%02X%02X", codec.RecordInfo.AVCProfileIndication, codec.RecordInfo.ProfileCompatibility, codec.RecordInfo.AVCLevelIndication)
		self.codecString = "hev1.1.6.L120.90"

	} else if self.Type() == av.AAC {
		codec := self.CodecData.(aacparser.CodecData)
		self.sample.SampleDesc.MP4ADesc = &mp4io.MP4ADesc{
			DataRefIdx:       1,
			NumberOfChannels: int16(codec.ChannelLayout().Count()),
			SampleSize:       int16(codec.SampleFormat().BytesPerSample() * 4),
			SampleRate:       float64(codec.SampleRate()),
			Unknowns:         []mp4io.Atom{self.buildEsds(codec.MPEG4AudioConfigBytes())},
		}

		self.trackAtom.Header.Volume = 1
		self.trackAtom.Header.AlternateGroup = 1
		self.trackAtom.Header.Duration = 0

		self.trackAtom.Media.Handler = &mp4io.HandlerRefer{
			SubType: [4]byte{'s', 'o', 'u', 'n'},
			Name:    []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 'G', 'G', 0, 0, 0},
		}

		self.trackAtom.Media.Info.Sound = &mp4io.SoundMediaInfo{}
		self.codecString = "mp4a.40.2"

	} else {
		err = fmt.Errorf("fmp4: codec type=%d invalid", self.Type())
	}

	return
}

func (self *Muxer) WriteTrailer() (err error) {
	return
}

func (element *Muxer) WriteHeader(streams []av.CodecData) error {
	element.streams = []*Stream{}
	for _, stream := range streams {
		if err := element.newStream(stream); err != nil {
			log.Println("WriteHeader", err)
		}
	}

	return nil
}

func (element *Muxer) GetInit(streams []av.CodecData) (string, []byte) {
	moov := &mp4io.Movie{
		Header: &mp4io.MovieHeader{
			PreferredRate:     1,
			PreferredVolume:   1,
			Matrix:            [9]int32{0x10000, 0, 0, 0, 0x10000, 0, 0, 0, 0x40000000},
			NextTrackId:       3,
			Duration:          0,
			TimeScale:         1000,
			CreateTime:        time0(),
			ModifyTime:        time0(),
			PreviewTime:       time0(),
			PreviewDuration:   time0(),
			PosterTime:        time0(),
			SelectionTime:     time0(),
			SelectionDuration: time0(),
			CurrentTime:       time0(),
		},
		Unknowns: []mp4io.Atom{element.buildMvex()},
	}
	var meta string
	for _, stream := range element.streams {
		if err := stream.fillTrackAtom(); err != nil {
			return meta, []byte{}
		}
		moov.Tracks = append(moov.Tracks, stream.trackAtom)
		meta += stream.codecString + ","
	}
	meta = meta[:len(meta)-1]
	ftypeData := []byte{0x00, 0x00, 0x00, 0x18, 0x66, 0x74, 0x79, 0x70, 0x69, 0x73, 0x6f, 0x36, 0x00, 0x00, 0x00, 0x01, 0x69, 0x73, 0x6f, 0x36, 0x64, 0x61, 0x73, 0x68}
	file := make([]byte, moov.Len()+len(ftypeData))
	copy(file, ftypeData)
	moov.Marshal(file[len(ftypeData):])
	return meta, file
}

func (element *Muxer) WritePacket(pkt av.Packet, GOP bool) (bool, []byte, error) {
	if pkt.Idx+1 > int8(len(element.streams)) {
		return false, nil, nil
	}
	stream := element.streams[pkt.Idx]
	if GOP {
		ts := time.Duration(0)
		if stream.lastpkt != nil {
			ts = pkt.Time - stream.lastpkt.Time
		}
		got, buf, err := stream.writePacketV3(pkt, ts, 5)
		stream.lastpkt = &pkt
		if err != nil {
			return false, []byte{}, err
		}
		return got, buf, err
	}
	ts := time.Duration(0)
	if stream.lastpkt != nil {
		ts = pkt.Time - stream.lastpkt.Time
	}
	got, buf, err := stream.writePacketV2(pkt, ts, 5)
	stream.lastpkt = &pkt
	if err != nil {
		return false, []byte{}, err
	}
	return got, buf, err
}

func (element *Muxer) WritePacketPrepush(pkt av.Packet, dur time.Duration, GOP bool) (bool, []byte, error) {
	stream := element.streams[pkt.Idx]
	if GOP {

		got, buf, err := stream.writePacketV3(pkt, dur, 0)
		stream.lastpkt = &pkt
		if err != nil {
			return false, []byte{}, err
		}
		return got, buf, err
	}

	got, buf, err := stream.writePacketV2(pkt, dur, 0)
	stream.lastpkt = &pkt
	if err != nil {
		return false, []byte{}, err
	}
	return got, buf, err
}

func (element *Muxer) WritePacket4(pkt av.Packet) error {
	stream := element.streams[pkt.Idx]
	return stream.writePacketV4(pkt)
}
func (element *Stream) writePacketV4(pkt av.Packet) error {
	//pkt.Data = pkt.Data[4:]
	defaultFlags := fmp4io.SampleNonKeyframe
	if pkt.IsKeyFrame {
		defaultFlags = fmp4io.SampleNoDependencies
	}
	trackID := pkt.Idx + 1
	if element.sampleIndex == 0 {
		element.moof.Header = &mp4fio.MovieFragHeader{Seqnum: uint32(element.muxer.fragmentIndex + 1)}
		element.moof.Tracks = []*mp4fio.TrackFrag{
			&mp4fio.TrackFrag{
				Header: &mp4fio.TrackFragHeader{
					Data: []byte{0x00, 0x02, 0x00, 0x20, 0x00, 0x00, 0x00, uint8(trackID), 0x01, 0x01, 0x00, 0x00},
				},
				DecodeTime: &mp4fio.TrackFragDecodeTime{
					Version: 1,
					Flags:   0,
					Time:    uint64(element.timeToTs(pkt.Time)),
				},
				Run: &mp4fio.TrackFragRun{
					Flags:            0x000b05,
					FirstSampleFlags: uint32(defaultFlags),
					DataOffset:       0,
					Entries:          []mp4io.TrackFragRunEntry{},
				},
			},
		}
		element.buffer = []byte{0x00, 0x00, 0x00, 0x00, 0x6d, 0x64, 0x61, 0x74}
	}
	runEnrty := mp4io.TrackFragRunEntry{
		Duration: uint32(element.timeToTs(pkt.Duration)),
		Size:     uint32(len(pkt.Data)),
		Cts:      uint32(element.timeToTs(pkt.CompositionTime)),
		Flags:    uint32(defaultFlags),
	}
	//log.Println("packet", defaultFlags,pkt.Duration,  pkt.CompositionTime)
	element.moof.Tracks[0].Run.Entries = append(element.moof.Tracks[0].Run.Entries, runEnrty)
	element.buffer = append(element.buffer, pkt.Data...)
	element.sampleIndex++
	element.dts += element.timeToTs(pkt.Duration)

	return nil
}
func (element *Muxer) SetIndex(val int) {
	element.fragmentIndex = val
}
func (element *Stream) writePacketV3(pkt av.Packet, rawdur time.Duration, maxFrames int) (bool, []byte, error) {
	trackID := pkt.Idx + 1
	var out []byte
	var got bool
	if element.sampleIndex > maxFrames && pkt.IsKeyFrame {
		element.moof.Tracks[0].Run.DataOffset = uint32(element.moof.Len() + 8)
		out = make([]byte, element.moof.Len()+len(element.buffer))
		element.moof.Marshal(out)
		pio.PutU32BE(element.buffer, uint32(len(element.buffer)))
		copy(out[element.moof.Len():], element.buffer)
		element.sampleIndex = 0
		element.muxer.fragmentIndex++
		got = true
	}
	if element.sampleIndex == 0 {
		element.moof.Header = &mp4fio.MovieFragHeader{Seqnum: uint32(element.muxer.fragmentIndex + 1)}
		element.moof.Tracks = []*mp4fio.TrackFrag{
			&mp4fio.TrackFrag{
				Header: &mp4fio.TrackFragHeader{
					Data: []byte{0x00, 0x02, 0x00, 0x20, 0x00, 0x00, 0x00, uint8(trackID), 0x01, 0x01, 0x00, 0x00},
				},
				DecodeTime: &mp4fio.TrackFragDecodeTime{
					Version: 1,
					Flags:   0,
					Time:    uint64(element.dts),
				},
				Run: &mp4fio.TrackFragRun{
					Flags:            0x000b05,
					FirstSampleFlags: 0x02000000,
					DataOffset:       0,
					Entries:          []mp4io.TrackFragRunEntry{},
				},
			},
		}
		element.buffer = []byte{0x00, 0x00, 0x00, 0x00, 0x6d, 0x64, 0x61, 0x74}
	}
	runEnrty := mp4io.TrackFragRunEntry{
		Duration: uint32(element.timeToTs(rawdur)),
		Size:     uint32(len(pkt.Data)),
		Cts:      uint32(element.timeToTs(pkt.CompositionTime)),
	}
	element.moof.Tracks[0].Run.Entries = append(element.moof.Tracks[0].Run.Entries, runEnrty)
	element.buffer = append(element.buffer, pkt.Data...)
	element.sampleIndex++
	element.dts += element.timeToTs(rawdur)
	return got, out, nil
}
func (element *Muxer) Finalize() []byte {
	stream := element.streams[0]
	stream.moof.Tracks[0].Run.DataOffset = uint32(stream.moof.Len() + 8)
	out := make([]byte, stream.moof.Len()+len(stream.buffer))
	stream.moof.Marshal(out)
	PutU32BE(stream.buffer, uint32(len(stream.buffer)))
	copy(out[stream.moof.Len():], stream.buffer)
	stream.sampleIndex = 0
	stream.muxer.fragmentIndex++
	return out

}

// PutU32BE func
func PutU32BE(b []byte, v uint32) {
	b[0] = byte(v >> 24)
	b[1] = byte(v >> 16)
	b[2] = byte(v >> 8)
	b[3] = byte(v)
}
func (element *Stream) writePacketV2(pkt av.Packet, rawdur time.Duration, maxFrames int) (bool, []byte, error) {
	trackID := pkt.Idx + 1
	if element.sampleIndex == 0 {
		element.moof.Header = &mp4fio.MovieFragHeader{Seqnum: uint32(element.muxer.fragmentIndex + 1)}
		element.moof.Tracks = []*mp4fio.TrackFrag{
			&mp4fio.TrackFrag{
				Header: &mp4fio.TrackFragHeader{
					Data: []byte{0x00, 0x02, 0x00, 0x20, 0x00, 0x00, 0x00, uint8(trackID), 0x01, 0x01, 0x00, 0x00},
				},
				DecodeTime: &mp4fio.TrackFragDecodeTime{
					Version: 1,
					Flags:   0,
					Time:    uint64(element.dts),
				},
				Run: &mp4fio.TrackFragRun{
					Flags:            0x000b05,
					FirstSampleFlags: 0x02000000,
					DataOffset:       0,
					Entries:          []mp4io.TrackFragRunEntry{},
				},
			},
		}
		element.buffer = []byte{0x00, 0x00, 0x00, 0x00, 0x6d, 0x64, 0x61, 0x74}
	}
	runEnrty := mp4io.TrackFragRunEntry{
		Duration: uint32(element.timeToTs(rawdur)),
		Size:     uint32(len(pkt.Data)),
		Cts:      uint32(element.timeToTs(pkt.CompositionTime)),
	}
	element.moof.Tracks[0].Run.Entries = append(element.moof.Tracks[0].Run.Entries, runEnrty)
	element.buffer = append(element.buffer, pkt.Data...)
	element.sampleIndex++
	element.dts += element.timeToTs(rawdur)
	if element.sampleIndex > maxFrames { // Количество фреймов в пакете
		element.moof.Tracks[0].Run.DataOffset = uint32(element.moof.Len() + 8)
		file := make([]byte, element.moof.Len()+len(element.buffer))
		element.moof.Marshal(file)
		pio.PutU32BE(element.buffer, uint32(len(element.buffer)))
		copy(file[element.moof.Len():], element.buffer)
		element.sampleIndex = 0
		element.muxer.fragmentIndex++
		return true, file, nil
	}
	return false, []byte{}, nil
}

func (self *Muxer) buildMvex() *FDummy {
	mvex := &FDummy{
		Data: []byte{
			0x00, 0x00, 0x00, 0x38, 0x6D, 0x76, 0x65, 0x78,
			0x00, 0x00, 0x00, 0x10, 0x6D, 0x65, 0x68, 0x64, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		Tag_: mp4io.Tag(0x6D766578),
	}
	for i := 1; i <= len(self.streams); i++ {
		trex := self.buildTrex(i)
		mvex.Data = append(mvex.Data, trex...)
	}

	pio.PutU32BE(mvex.Data, uint32(len(mvex.Data)))
	return mvex
}

func (self *Muxer) buildTrex(trackId int) []byte {
	return []byte{
		0x00, 0x00, 0x00, 0x20, 0x74, 0x72, 0x65, 0x78,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, uint8(trackId), 0x00, 0x00,
		0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00}
}

func time0() time.Time {
	return time.Date(1904, time.January, 1, 0, 0, 0, 0, time.UTC)
}
