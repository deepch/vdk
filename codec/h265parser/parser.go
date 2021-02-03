package h265parser

import (
	"bytes"
	"fmt"

	"github.com/deepch/vdk/av"
	"github.com/deepch/vdk/utils/bits"
	"github.com/deepch/vdk/utils/bits/pio"
)

const (
	NAL_UNIT_CODED_SLICE_TRAIL_N    = 0
	NAL_UNIT_CODED_SLICE_TRAIL_R    = 1
	NAL_UNIT_CODED_SLICE_TSA_N      = 2
	NAL_UNIT_CODED_SLICE_TSA_R      = 3
	NAL_UNIT_CODED_SLICE_STSA_N     = 4
	NAL_UNIT_CODED_SLICE_STSA_R     = 5
	NAL_UNIT_CODED_SLICE_RADL_N     = 6
	NAL_UNIT_CODED_SLICE_RADL_R     = 7
	NAL_UNIT_CODED_SLICE_RASL_N     = 8
	NAL_UNIT_CODED_SLICE_RASL_R     = 9
	NAL_UNIT_RESERVED_VCL_N10       = 10
	NAL_UNIT_RESERVED_VCL_R11       = 11
	NAL_UNIT_RESERVED_VCL_N12       = 12
	NAL_UNIT_RESERVED_VCL_R13       = 13
	NAL_UNIT_RESERVED_VCL_N14       = 14
	NAL_UNIT_RESERVED_VCL_R15       = 15
	NAL_UNIT_CODED_SLICE_BLA_W_LP   = 16
	NAL_UNIT_CODED_SLICE_BLA_W_RADL = 17
	NAL_UNIT_CODED_SLICE_BLA_N_LP   = 18
	NAL_UNIT_CODED_SLICE_IDR_W_RADL = 19
	NAL_UNIT_CODED_SLICE_IDR_N_LP   = 20
	NAL_UNIT_CODED_SLICE_CRA        = 21
	NAL_UNIT_RESERVED_IRAP_VCL22    = 22
	NAL_UNIT_RESERVED_IRAP_VCL23    = 23
	NAL_UNIT_RESERVED_VCL24         = 24
	NAL_UNIT_RESERVED_VCL25         = 25
	NAL_UNIT_RESERVED_VCL26         = 26
	NAL_UNIT_RESERVED_VCL27         = 27
	NAL_UNIT_RESERVED_VCL28         = 28
	NAL_UNIT_RESERVED_VCL29         = 29
	NAL_UNIT_RESERVED_VCL30         = 30
	NAL_UNIT_RESERVED_VCL31         = 31
	NAL_UNIT_VPS                    = 32
	NAL_UNIT_SPS                    = 33
	NAL_UNIT_PPS                    = 34
	NAL_UNIT_ACCESS_UNIT_DELIMITER  = 35
	NAL_UNIT_EOS                    = 36
	NAL_UNIT_EOB                    = 37
	NAL_UNIT_FILLER_DATA            = 38
	NAL_UNIT_PREFIX_SEI             = 39
	NAL_UNIT_SUFFIX_SEI             = 40
	NAL_UNIT_RESERVED_NVCL41        = 41
	NAL_UNIT_RESERVED_NVCL42        = 42
	NAL_UNIT_RESERVED_NVCL43        = 43
	NAL_UNIT_RESERVED_NVCL44        = 44
	NAL_UNIT_RESERVED_NVCL45        = 45
	NAL_UNIT_RESERVED_NVCL46        = 46
	NAL_UNIT_RESERVED_NVCL47        = 47
	NAL_UNIT_UNSPECIFIED_48         = 48
	NAL_UNIT_UNSPECIFIED_49         = 49
	NAL_UNIT_UNSPECIFIED_50         = 50
	NAL_UNIT_UNSPECIFIED_51         = 51
	NAL_UNIT_UNSPECIFIED_52         = 52
	NAL_UNIT_UNSPECIFIED_53         = 53
	NAL_UNIT_UNSPECIFIED_54         = 54
	NAL_UNIT_UNSPECIFIED_55         = 55
	NAL_UNIT_UNSPECIFIED_56         = 56
	NAL_UNIT_UNSPECIFIED_57         = 57
	NAL_UNIT_UNSPECIFIED_58         = 58
	NAL_UNIT_UNSPECIFIED_59         = 59
	NAL_UNIT_UNSPECIFIED_60         = 60
	NAL_UNIT_UNSPECIFIED_61         = 61
	NAL_UNIT_UNSPECIFIED_62         = 62
	NAL_UNIT_UNSPECIFIED_63         = 63
	NAL_UNIT_INVALID                = 64
)

const (
	MAX_VPS_COUNT  = 16
	MAX_SUB_LAYERS = 7
	MAX_SPS_COUNT  = 32
)

func IsDataNALU(b []byte) bool {
	typ := b[0] & 0x1f
	return typ >= 1 && typ <= 5
}

var StartCodeBytes = []byte{0, 0, 1}
var AUDBytes = []byte{0, 0, 0, 1, 0x9, 0xf0, 0, 0, 0, 1} // AUD

func CheckNALUsType(b []byte) (typ int) {
	_, typ = SplitNALUs(b)
	return
}

const (
	NALU_RAW = iota
	NALU_AVCC
	NALU_ANNEXB
)

func SplitNALUs(b []byte) (nalus [][]byte, typ int) {
	if len(b) < 4 {
		return [][]byte{b}, NALU_RAW
	}
	val3 := pio.U24BE(b)
	val4 := pio.U32BE(b)
	if val4 <= uint32(len(b)) {
		_val4 := val4
		_b := b[4:]
		nalus := [][]byte{}
		for {
			nalus = append(nalus, _b[:_val4])
			_b = _b[_val4:]
			if len(_b) < 4 {
				break
			}
			_val4 = pio.U32BE(_b)
			_b = _b[4:]
			if _val4 > uint32(len(_b)) {
				break
			}
		}
		if len(_b) == 0 {
			return nalus, NALU_AVCC
		}
	}
	// is Annex B
	if val3 == 1 || val4 == 1 {
		_val3 := val3
		_val4 := val4
		start := 0
		pos := 0
		for {
			if start != pos {
				nalus = append(nalus, b[start:pos])
			}
			if _val3 == 1 {
				pos += 3
			} else if _val4 == 1 {
				pos += 4
			}
			start = pos
			if start == len(b) {
				break
			}
			_val3 = 0
			_val4 = 0
			for pos < len(b) {
				if pos+2 < len(b) && b[pos] == 0 {
					_val3 = pio.U24BE(b[pos:])
					if _val3 == 0 {
						if pos+3 < len(b) {
							_val4 = uint32(b[pos+3])
							if _val4 == 1 {
								break
							}
						}
					} else if _val3 == 1 {
						break
					}
					pos++
				} else {
					pos++
				}
			}
		}
		typ = NALU_ANNEXB
		return
	}

	return [][]byte{b}, NALU_RAW
}

type SPSInfo struct {
	ProfileIdc uint
	LevelIdc   uint

	MbWidth  uint
	MbHeight uint

	CropLeft   uint
	CropRight  uint
	CropTop    uint
	CropBottom uint

	Width  uint
	Height uint
}

func ParseSPS(data []byte) (self SPSInfo, err error) {
	r := &bits.GolombBitReader{R: bytes.NewReader(data)}

	if _, err = r.ReadBits(8); err != nil {
		return
	}

	if self.ProfileIdc, err = r.ReadBits(8); err != nil {
		return
	}

	// constraint_set0_flag-constraint_set6_flag,reserved_zero_2bits
	if _, err = r.ReadBits(8); err != nil {
		return
	}

	// level_idc
	if self.LevelIdc, err = r.ReadBits(8); err != nil {
		return
	}

	// seq_parameter_set_id
	if _, err = r.ReadExponentialGolombCode(); err != nil {
		return
	}

	if self.ProfileIdc == 100 || self.ProfileIdc == 110 ||
		self.ProfileIdc == 122 || self.ProfileIdc == 244 ||
		self.ProfileIdc == 44 || self.ProfileIdc == 83 ||
		self.ProfileIdc == 86 || self.ProfileIdc == 118 {

		var chroma_format_idc uint
		if chroma_format_idc, err = r.ReadExponentialGolombCode(); err != nil {
			return
		}

		if chroma_format_idc == 3 {
			// residual_colour_transform_flag
			if _, err = r.ReadBit(); err != nil {
				return
			}
		}

		// bit_depth_luma_minus8
		if _, err = r.ReadExponentialGolombCode(); err != nil {
			return
		}
		// bit_depth_chroma_minus8
		if _, err = r.ReadExponentialGolombCode(); err != nil {
			return
		}
		// qpprime_y_zero_transform_bypass_flag
		if _, err = r.ReadBit(); err != nil {
			return
		}

		var seq_scaling_matrix_present_flag uint
		if seq_scaling_matrix_present_flag, err = r.ReadBit(); err != nil {
			return
		}

		if seq_scaling_matrix_present_flag != 0 {
			for i := 0; i < 8; i++ {
				var seq_scaling_list_present_flag uint
				if seq_scaling_list_present_flag, err = r.ReadBit(); err != nil {
					return
				}
				if seq_scaling_list_present_flag != 0 {
					var sizeOfScalingList uint
					if i < 6 {
						sizeOfScalingList = 16
					} else {
						sizeOfScalingList = 64
					}
					lastScale := uint(8)
					nextScale := uint(8)
					for j := uint(0); j < sizeOfScalingList; j++ {
						if nextScale != 0 {
							var delta_scale uint
							if delta_scale, err = r.ReadSE(); err != nil {
								return
							}
							nextScale = (lastScale + delta_scale + 256) % 256
						}
						if nextScale != 0 {
							lastScale = nextScale
						}
					}
				}
			}
		}
	}

	// log2_max_frame_num_minus4
	if _, err = r.ReadExponentialGolombCode(); err != nil {
		return
	}

	var pic_order_cnt_type uint
	if pic_order_cnt_type, err = r.ReadExponentialGolombCode(); err != nil {
		return
	}
	if pic_order_cnt_type == 0 {
		// log2_max_pic_order_cnt_lsb_minus4
		if _, err = r.ReadExponentialGolombCode(); err != nil {
			return
		}
	} else if pic_order_cnt_type == 1 {
		// delta_pic_order_always_zero_flag
		if _, err = r.ReadBit(); err != nil {
			return
		}
		// offset_for_non_ref_pic
		if _, err = r.ReadSE(); err != nil {
			return
		}
		// offset_for_top_to_bottom_field
		if _, err = r.ReadSE(); err != nil {
			return
		}
		var num_ref_frames_in_pic_order_cnt_cycle uint
		if num_ref_frames_in_pic_order_cnt_cycle, err = r.ReadExponentialGolombCode(); err != nil {
			return
		}
		for i := uint(0); i < num_ref_frames_in_pic_order_cnt_cycle; i++ {
			if _, err = r.ReadSE(); err != nil {
				return
			}
		}
	}

	// max_num_ref_frames
	if _, err = r.ReadExponentialGolombCode(); err != nil {
		return
	}

	// gaps_in_frame_num_value_allowed_flag
	if _, err = r.ReadBit(); err != nil {
		return
	}

	if self.MbWidth, err = r.ReadExponentialGolombCode(); err != nil {
		return
	}
	self.MbWidth++

	if self.MbHeight, err = r.ReadExponentialGolombCode(); err != nil {
		return
	}
	self.MbHeight++

	var frame_mbs_only_flag uint
	if frame_mbs_only_flag, err = r.ReadBit(); err != nil {
		return
	}
	if frame_mbs_only_flag == 0 {
		// mb_adaptive_frame_field_flag
		if _, err = r.ReadBit(); err != nil {
			return
		}
	}

	// direct_8x8_inference_flag
	if _, err = r.ReadBit(); err != nil {
		return
	}

	var frame_cropping_flag uint
	if frame_cropping_flag, err = r.ReadBit(); err != nil {
		return
	}
	if frame_cropping_flag != 0 {
		if self.CropLeft, err = r.ReadExponentialGolombCode(); err != nil {
			return
		}
		if self.CropRight, err = r.ReadExponentialGolombCode(); err != nil {
			return
		}
		if self.CropTop, err = r.ReadExponentialGolombCode(); err != nil {
			return
		}
		if self.CropBottom, err = r.ReadExponentialGolombCode(); err != nil {
			return
		}
	}

	self.Width = (self.MbWidth * 16) - self.CropLeft*2 - self.CropRight*2
	self.Height = ((2 - frame_mbs_only_flag) * self.MbHeight * 16) - self.CropTop*2 - self.CropBottom*2

	return
}

type CodecData struct {
	Record     []byte
	RecordInfo AVCDecoderConfRecord
	SPSInfo    SPSInfo
}

func (self CodecData) Type() av.CodecType {
	return av.H265
}

func (self CodecData) AVCDecoderConfRecordBytes() []byte {
	return self.Record
}

func (self CodecData) SPS() []byte {
	return self.RecordInfo.SPS[0]
}

func (self CodecData) PPS() []byte {
	return self.RecordInfo.PPS[0]
}

func (self CodecData) VPS() []byte {
	return self.RecordInfo.VPS[0]
}

func (self CodecData) Width() int {
	return int(self.SPSInfo.Width)
}

func (self CodecData) Height() int {
	return int(self.SPSInfo.Height)
}

func NewCodecDataFromAVCDecoderConfRecord(record []byte) (self CodecData, err error) {
	self.Record = record
	if _, err = (&self.RecordInfo).Unmarshal(record); err != nil {
		return
	}
	if len(self.RecordInfo.SPS) == 0 {
		err = fmt.Errorf("h265parser: no SPS found in AVCDecoderConfRecord")
		return
	}
	if len(self.RecordInfo.PPS) == 0 {
		err = fmt.Errorf("h265parser: no PPS found in AVCDecoderConfRecord")
		return
	}
	if len(self.RecordInfo.VPS) == 0 {
		err = fmt.Errorf("h265parser: no VPS found in AVCDecoderConfRecord")
		return
	}
	if self.SPSInfo, err = ParseSPS(self.RecordInfo.SPS[0]); err != nil {
		err = fmt.Errorf("h265parser: parse SPS failed(%s)", err)
		return
	}
	return
}

func NewCodecDataFromVPSAndSPSAndPPS(vps, sps, pps []byte) (self CodecData, err error) {
	recordinfo := AVCDecoderConfRecord{}
	recordinfo.AVCProfileIndication = sps[1]
	recordinfo.ProfileCompatibility = sps[2]
	recordinfo.AVCLevelIndication = sps[3]
	recordinfo.SPS = [][]byte{sps}
	recordinfo.PPS = [][]byte{pps}
	recordinfo.VPS = [][]byte{vps}
	recordinfo.LengthSizeMinusOne = 3

	buf := make([]byte, recordinfo.Len())
	recordinfo.Marshal(buf)

	self.RecordInfo = recordinfo
	self.Record = buf

	if self.SPSInfo, err = ParseSPS(sps); err != nil {
		return
	}
	return
}

type AVCDecoderConfRecord struct {
	AVCProfileIndication uint8
	ProfileCompatibility uint8
	AVCLevelIndication   uint8
	LengthSizeMinusOne   uint8
	VPS                  [][]byte
	SPS                  [][]byte
	PPS                  [][]byte
}

var ErrDecconfInvalid = fmt.Errorf("h265parser: AVCDecoderConfRecord invalid")

func (self *AVCDecoderConfRecord) Unmarshal(b []byte) (n int, err error) {
	if len(b) < 7 {
		err = ErrDecconfInvalid
		return
	}

	self.AVCProfileIndication = b[1]
	self.ProfileCompatibility = b[2]
	self.AVCLevelIndication = b[3]
	self.LengthSizeMinusOne = b[4] & 0x03
	spscount := int(b[5] & 0x1f)
	n += 6

	for i := 0; i < spscount; i++ {
		if len(b) < n+2 {
			err = ErrDecconfInvalid
			return
		}
		spslen := int(pio.U16BE(b[n:]))
		n += 2

		if len(b) < n+spslen {
			err = ErrDecconfInvalid
			return
		}
		self.SPS = append(self.SPS, b[n:n+spslen])
		n += spslen
	}

	if len(b) < n+1 {
		err = ErrDecconfInvalid
		return
	}

	ppscount := int(b[n])
	n++

	for i := 0; i < ppscount; i++ {
		if len(b) < n+2 {
			err = ErrDecconfInvalid
			return
		}
		ppslen := int(pio.U16BE(b[n:]))
		n += 2

		if len(b) < n+ppslen {
			err = ErrDecconfInvalid
			return
		}
		self.PPS = append(self.PPS, b[n:n+ppslen])
		n += ppslen
	}

	vpscount := int(b[n])
	n++

	for i := 0; i < vpscount; i++ {
		if len(b) < n+2 {
			err = ErrDecconfInvalid
			return
		}
		vpslen := int(pio.U16BE(b[n:]))
		n += 2

		if len(b) < n+vpslen {
			err = ErrDecconfInvalid
			return
		}
		self.VPS = append(self.VPS, b[n:n+vpslen])
		n += vpslen
	}
	return
}

func (self AVCDecoderConfRecord) Len() (n int) {
	n = 7
	for _, sps := range self.SPS {
		n += 2 + len(sps)
	}
	for _, pps := range self.PPS {
		n += 2 + len(pps)
	}
	for _, vps := range self.VPS {
		n += 2 + len(vps)
	}
	return
}

func (self AVCDecoderConfRecord) Marshal(b []byte) (n int) {
	b[0] = 1
	b[1] = self.AVCProfileIndication
	b[2] = self.ProfileCompatibility
	b[3] = self.AVCLevelIndication
	b[4] = self.LengthSizeMinusOne | 0xfc
	b[5] = uint8(len(self.SPS)) | 0xe0
	n += 6

	for _, sps := range self.SPS {
		pio.PutU16BE(b[n:], uint16(len(sps)))
		n += 2
		copy(b[n:], sps)
		n += len(sps)
	}

	b[n] = uint8(len(self.PPS))
	n++

	for _, pps := range self.PPS {
		pio.PutU16BE(b[n:], uint16(len(pps)))
		n += 2
		copy(b[n:], pps)
		n += len(pps)
	}

	b[n] = uint8(len(self.VPS))
	n++

	for _, vps := range self.VPS {
		pio.PutU16BE(b[n:], uint16(len(vps)))
		n += 2
		copy(b[n:], vps)
		n += len(vps)
	}

	return
}

type SliceType uint

func (self SliceType) String() string {
	switch self {
	case SLICE_P:
		return "P"
	case SLICE_B:
		return "B"
	case SLICE_I:
		return "I"
	}
	return ""
}

const (
	SLICE_P = iota + 1
	SLICE_B
	SLICE_I
)

func ParseSliceHeaderFromNALU(packet []byte) (sliceType SliceType, err error) {

	if len(packet) <= 1 {
		err = fmt.Errorf("h265parser: packet too short to parse slice header")
		return
	}

	nal_unit_type := packet[0] & 0x1f
	switch nal_unit_type {
	case 1, 2, 5, 19:
		// slice_layer_without_partitioning_rbsp
		// slice_data_partition_a_layer_rbsp

	default:
		err = fmt.Errorf("h265parser: nal_unit_type=%d has no slice header", nal_unit_type)
		return
	}

	r := &bits.GolombBitReader{R: bytes.NewReader(packet[1:])}

	// first_mb_in_slice
	if _, err = r.ReadExponentialGolombCode(); err != nil {
		return
	}

	// slice_type
	var u uint
	if u, err = r.ReadExponentialGolombCode(); err != nil {
		return
	}

	switch u {
	case 0, 3, 5, 8:
		sliceType = SLICE_P
	case 1, 6:
		sliceType = SLICE_B
	case 2, 4, 7, 9:
		sliceType = SLICE_I
	default:
		err = fmt.Errorf("h265parser: slice_type=%d invalid", u)
		return
	}

	return
}
