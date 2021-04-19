package fmp4

import (
	"fmt"

	"github.com/deepch/vdk/av"
	"github.com/deepch/vdk/codec/aacparser"
	"github.com/deepch/vdk/codec/h264parser"
	"github.com/deepch/vdk/codec/opusparser"
	"github.com/deepch/vdk/format/fmp4/esio"
	"github.com/deepch/vdk/format/fmp4/fmp4io"
)

// Track creates a TRAK atom for this stream
func (f *TrackFragmenter) Track() (*fmp4io.Track, error) {
	sample := &fmp4io.SampleTable{
		SampleDesc:    &fmp4io.SampleDesc{},
		TimeToSample:  &fmp4io.TimeToSample{},
		SampleToChunk: &fmp4io.SampleToChunk{},
		SampleSize:    &fmp4io.SampleSize{},
		ChunkOffset:   &fmp4io.ChunkOffset{},
	}
	switch cd := f.codecData.(type) {
	case h264parser.CodecData:
		f.timeScale = 90000
		cd.RecordInfo.LengthSizeMinusOne = 3
		conf := make([]byte, cd.RecordInfo.Len())
		cd.RecordInfo.Marshal(conf)
		sample.SampleDesc.AVC1Desc = &fmp4io.AVC1Desc{
			DataRefIdx:           1,
			HorizontalResolution: 72,
			VorizontalResolution: 72,
			Width:                int16(cd.Width()),
			Height:               int16(cd.Height()),
			FrameCount:           1,
			Depth:                24,
			ColorTableId:         -1,
			Conf:                 &fmp4io.AVC1Conf{Data: conf},
		}
	case aacparser.CodecData:
		f.timeScale = 48000
		dc, err := esio.DecoderConfigFromCodecData(cd)
		if err != nil {
			return nil, fmt.Errorf("decoding AAC configuration: %w", err)
		}
		sample.SampleDesc.MP4ADesc = &fmp4io.MP4ADesc{
			DataRefIdx:       1,
			NumberOfChannels: int16(cd.ChannelLayout().Count()),
			SampleSize:       16,
			SampleRate:       float64(cd.SampleRate()),
			Conf: &fmp4io.ElemStreamDesc{
				StreamDescriptor: &esio.StreamDescriptor{
					ESID:          uint16(f.trackID),
					DecoderConfig: dc,
					SLConfig:      &esio.SLConfigDescriptor{Predefined: esio.SLConfigMP4},
				},
			},
		}
	case *opusparser.CodecData:
		f.timeScale = 48000
		sample.SampleDesc.OpusDesc = &fmp4io.OpusSampleEntry{
			DataRefIdx:       1,
			NumberOfChannels: uint16(cd.ChannelLayout().Count()),
			SampleSize:       16,
			SampleRate:       float64(cd.SampleRate()),
			Conf: &fmp4io.OpusSpecificConfiguration{
				OutputChannelCount: uint8(cd.ChannelLayout().Count()),
				PreSkip:            3840, // 80ms
			},
		}
	default:
		return nil, fmt.Errorf("mp4: codec type=%v is not supported", f.codecData.Type())
	}
	trackAtom := &fmp4io.Track{
		Header: &fmp4io.TrackHeader{
			Flags:   0x0003, // Track enabled | Track in movie
			Matrix:  [9]int32{0x10000, 0, 0, 0, 0x10000, 0, 0, 0, 0x40000000},
			TrackID: f.trackID,
		},
		Media: &fmp4io.Media{
			Header: &fmp4io.MediaHeader{
				Language:  21956,
				TimeScale: f.timeScale,
			},
			Info: &fmp4io.MediaInfo{
				Sample: sample,
				Data: &fmp4io.DataInfo{
					Refer: &fmp4io.DataRefer{
						Url: &fmp4io.DataReferUrl{
							Flags: 0x000001, // Self reference
						},
					},
				},
			},
		},
	}
	if f.codecData.Type().IsVideo() {
		vc := f.codecData.(av.VideoCodecData)
		trackAtom.Media.Handler = &fmp4io.HandlerRefer{
			Type: fmp4io.VideoHandler,
			Name: "VideoHandler",
		}
		trackAtom.Media.Info.Video = &fmp4io.VideoMediaInfo{
			Flags: 0x000001,
		}
		trackAtom.Header.TrackWidth = float64(vc.Width())
		trackAtom.Header.TrackHeight = float64(vc.Height())
	} else {
		trackAtom.Header.Volume = 1
		trackAtom.Header.AlternateGroup = 1
		trackAtom.Media.Handler = &fmp4io.HandlerRefer{
			Type: fmp4io.SoundHandler,
			Name: "SoundHandler",
		}
		trackAtom.Media.Info.Sound = &fmp4io.SoundMediaInfo{}
	}
	return trackAtom, nil
}

// MovieHeader marshals an init.mp4 for the given tracks
func MovieHeader(tracks []*fmp4io.Track) ([]byte, error) {
	ftyp := fmp4io.FileType{
		MajorBrand: 0x69736f36, // iso6
		CompatibleBrands: []uint32{
			0x69736f35, // iso5
			0x6d703431, // mp41
		},
	}
	moov := &fmp4io.Movie{
		Header: &fmp4io.MovieHeader{
			PreferredRate:   1,
			PreferredVolume: 1,
			Matrix:          [9]int32{0x10000, 0, 0, 0, 0x10000, 0, 0, 0, 0x40000000},
			TimeScale:       1000,
		},
		Tracks:      tracks,
		MovieExtend: &fmp4io.MovieExtend{},
	}
	for _, track := range tracks {
		if track.Header.TrackID >= moov.Header.NextTrackID {
			moov.Header.NextTrackID = track.Header.TrackID + 1
		}
		moov.MovieExtend.Tracks = append(moov.MovieExtend.Tracks,
			&fmp4io.TrackExtend{TrackID: track.Header.TrackID, DefaultSampleDescIdx: 1})
	}
	// marshal init segment
	fhdr := make([]byte, ftyp.Len()+moov.Len())
	n := ftyp.Marshal(fhdr)
	moov.Marshal(fhdr[n:])
	return fhdr, nil
}

// FragmentHeader returns the header needed for the beginning of a MP4 segment file
func FragmentHeader() []byte {
	styp := fmp4io.SegmentType{
		MajorBrand:       0x6d736468,           // msdh
		CompatibleBrands: []uint32{0x6d736978}, // msix
	}
	shdr := make([]byte, styp.Len())
	styp.Marshal(shdr)
	return shdr
}
