package h265parser

import (
	"bytes"
	"errors"
	"fmt"
	"time"

	"github.com/deepch/vdk/av"
	"github.com/deepch/vdk/utils/bits"
	"github.com/deepch/vdk/utils/bits/pio"
)

type SPSInfo struct {
	ProfileIdc                       uint
	LevelIdc                         uint
	MbWidth                          uint
	MbHeight                         uint
	CropLeft                         uint
	CropRight                        uint
	CropTop                          uint
	CropBottom                       uint
	Width                            uint
	Height                           uint
	numTemporalLayers                uint
	temporalIdNested                 uint
	chromaFormat                     uint
	PicWidthInLumaSamples            uint
	PicHeightInLumaSamples           uint
	bitDepthLumaMinus8               uint
	bitDepthChromaMinus8             uint
	generalProfileSpace              uint
	generalTierFlag                  uint
	generalProfileIDC                uint
	generalProfileCompatibilityFlags uint32
	generalConstraintIndicatorFlags  uint64
	generalLevelIDC                  uint
	fps                              uint
}

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

var (
	ErrorH265IncorectUnitSize = errors.New("Invorect Unit Size")
	ErrorH265IncorectUnitType = errors.New("Incorect Unit Type")
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

func ParseSPS(sps []byte) (ctx SPSInfo, err error) {
	if len(sps) < 2 {
		err = ErrorH265IncorectUnitSize
		return
	}
	rbsp := nal2rbsp(sps[2:])
	br := &bits.GolombBitReader{R: bytes.NewReader(rbsp)}
	if _, err = br.ReadBits(4); err != nil {
		return
	}
	spsMaxSubLayersMinus1, err := br.ReadBits(3)
	if err != nil {
		return
	}

	if spsMaxSubLayersMinus1+1 > ctx.numTemporalLayers {
		ctx.numTemporalLayers = spsMaxSubLayersMinus1 + 1
	}
	if ctx.temporalIdNested, err = br.ReadBit(); err != nil {
		return
	}
	if err = parsePTL(br, &ctx, spsMaxSubLayersMinus1); err != nil {
		return
	}
	if _, err = br.ReadExponentialGolombCode(); err != nil {
		return
	}
	var cf uint
	if cf, err = br.ReadExponentialGolombCode(); err != nil {
		return
	}
	ctx.chromaFormat = uint(cf)
	if ctx.chromaFormat == 3 {
		if _, err = br.ReadBit(); err != nil {
			return
		}
	}
	if ctx.PicWidthInLumaSamples, err = br.ReadExponentialGolombCode(); err != nil {
		return
	}
	ctx.Width = uint(ctx.PicWidthInLumaSamples)
	if ctx.PicHeightInLumaSamples, err = br.ReadExponentialGolombCode(); err != nil {
		return
	}
	ctx.Height = uint(ctx.PicHeightInLumaSamples)
	conformanceWindowFlag, err := br.ReadBit()
	if err != nil {
		return
	}
	if conformanceWindowFlag != 0 {
		if _, err = br.ReadExponentialGolombCode(); err != nil {
			return
		}
		if _, err = br.ReadExponentialGolombCode(); err != nil {
			return
		}
		if _, err = br.ReadExponentialGolombCode(); err != nil {
			return
		}
		if _, err = br.ReadExponentialGolombCode(); err != nil {
			return
		}
	}

	var bdlm8 uint
	if bdlm8, err = br.ReadExponentialGolombCode(); err != nil {
		return
	}
	ctx.bitDepthChromaMinus8 = uint(bdlm8)
	var bdcm8 uint
	if bdcm8, err = br.ReadExponentialGolombCode(); err != nil {
		return
	}
	ctx.bitDepthChromaMinus8 = uint(bdcm8)

	_, err = br.ReadExponentialGolombCode()
	if err != nil {
		return
	}
	spsSubLayerOrderingInfoPresentFlag, err := br.ReadBit()
	if err != nil {
		return
	}
	var i uint
	if spsSubLayerOrderingInfoPresentFlag != 0 {
		i = 0
	} else {
		i = spsMaxSubLayersMinus1
	}
	for ; i <= spsMaxSubLayersMinus1; i++ {
		if _, err = br.ReadExponentialGolombCode(); err != nil {
			return
		}
		if _, err = br.ReadExponentialGolombCode(); err != nil {
			return
		}
		if _, err = br.ReadExponentialGolombCode(); err != nil {
			return
		}
	}

	if _, err = br.ReadExponentialGolombCode(); err != nil {
		return
	}
	if _, err = br.ReadExponentialGolombCode(); err != nil {
		return
	}
	if _, err = br.ReadExponentialGolombCode(); err != nil {
		return
	}
	if _, err = br.ReadExponentialGolombCode(); err != nil {
		return
	}
	if _, err = br.ReadExponentialGolombCode(); err != nil {
		return
	}
	if _, err = br.ReadExponentialGolombCode(); err != nil {
		return
	}
	return
}

func parsePTL(br *bits.GolombBitReader, ctx *SPSInfo, maxSubLayersMinus1 uint) error {
	var err error
	var ptl SPSInfo
	if ptl.generalProfileSpace, err = br.ReadBits(2); err != nil {
		return err
	}
	if ptl.generalTierFlag, err = br.ReadBit(); err != nil {
		return err
	}
	if ptl.generalProfileIDC, err = br.ReadBits(5); err != nil {
		return err
	}
	if ptl.generalProfileCompatibilityFlags, err = br.ReadBits32(32); err != nil {
		return err
	}
	if ptl.generalConstraintIndicatorFlags, err = br.ReadBits64(48); err != nil {
		return err
	}
	if ptl.generalLevelIDC, err = br.ReadBits(8); err != nil {
		return err
	}
	updatePTL(ctx, &ptl)
	if maxSubLayersMinus1 == 0 {
		return nil
	}
	subLayerProfilePresentFlag := make([]uint, maxSubLayersMinus1)
	subLayerLevelPresentFlag := make([]uint, maxSubLayersMinus1)
	for i := uint(0); i < maxSubLayersMinus1; i++ {
		if subLayerProfilePresentFlag[i], err = br.ReadBit(); err != nil {
			return err
		}
		if subLayerLevelPresentFlag[i], err = br.ReadBit(); err != nil {
			return err
		}
	}
	if maxSubLayersMinus1 > 0 {
		for i := maxSubLayersMinus1; i < 8; i++ {
			if _, err = br.ReadBits(2); err != nil {
				return err
			}
		}
	}
	for i := uint(0); i < maxSubLayersMinus1; i++ {
		if subLayerProfilePresentFlag[i] != 0 {
			if _, err = br.ReadBits32(32); err != nil {
				return err
			}
			if _, err = br.ReadBits32(32); err != nil {
				return err
			}
			if _, err = br.ReadBits32(24); err != nil {
				return err
			}
		}

		if subLayerLevelPresentFlag[i] != 0 {
			if _, err = br.ReadBits(8); err != nil {
				return err
			}
		}
	}
	return nil
}

func updatePTL(ctx, ptl *SPSInfo) {
	ctx.generalProfileSpace = ptl.generalProfileSpace

	if ptl.generalTierFlag > ctx.generalTierFlag {
		ctx.generalLevelIDC = ptl.generalLevelIDC

		ctx.generalTierFlag = ptl.generalTierFlag
	} else {
		if ptl.generalLevelIDC > ctx.generalLevelIDC {
			ctx.generalLevelIDC = ptl.generalLevelIDC
		}
	}

	if ptl.generalProfileIDC > ctx.generalProfileIDC {
		ctx.generalProfileIDC = ptl.generalProfileIDC
	}

	ctx.generalProfileCompatibilityFlags &= ptl.generalProfileCompatibilityFlags

	ctx.generalConstraintIndicatorFlags &= ptl.generalConstraintIndicatorFlags
}

func nal2rbsp(nal []byte) []byte {
	return bytes.Replace(nal, []byte{0x0, 0x0, 0x3}, []byte{0x0, 0x0}, -1)
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

func (self CodecData) FPS() int {
	return int(self.SPSInfo.fps)
}

func (self CodecData) Resolution() string {
	return fmt.Sprintf("%vx%v", self.Width(), self.Height())
}

func (self CodecData) Tag() string {
	//return fmt.Sprintf("hvc1.%02X%02X%02X", self.RecordInfo.AVCProfileIndication, self.RecordInfo.ProfileCompatibility, self.RecordInfo.AVCLevelIndication)
	return "hev1.1.6.L120.90"
}

func (self CodecData) Bandwidth() string {
	return fmt.Sprintf("%v", (int(float64(self.Width())*(float64(1.71)*(30/float64(self.FPS())))))*1000)
}

func (self CodecData) PacketDuration(data []byte) time.Duration {
	return time.Duration(1000./float64(self.FPS())) * time.Millisecond
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
	recordinfo.AVCProfileIndication = sps[3]
	recordinfo.ProfileCompatibility = sps[4]
	recordinfo.AVCLevelIndication = sps[5]
	recordinfo.SPS = [][]byte{sps}
	recordinfo.PPS = [][]byte{pps}
	recordinfo.VPS = [][]byte{vps}
	recordinfo.LengthSizeMinusOne = 3
	if self.SPSInfo, err = ParseSPS(sps); err != nil {
		return
	}
	buf := make([]byte, recordinfo.Len())
	recordinfo.Marshal(buf, self.SPSInfo)
	self.RecordInfo = recordinfo
	self.Record = buf
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
	if len(b) < 30 {
		err = ErrDecconfInvalid
		return
	}
	self.AVCProfileIndication = b[1]
	self.ProfileCompatibility = b[2]
	self.AVCLevelIndication = b[3]
	self.LengthSizeMinusOne = b[4] & 0x03

	vpscount := int(b[25] & 0x1f)
	n += 26
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

	if len(b) < n+1 {
		err = ErrDecconfInvalid
		return
	}

	n++
	n++

	spscount := int(b[n])
	n++

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

	n++
	n++

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
	return
}

func (self AVCDecoderConfRecord) Len() (n int) {
	n = 23
	for _, sps := range self.SPS {
		n += 5 + len(sps)
	}
	for _, pps := range self.PPS {
		n += 5 + len(pps)
	}
	for _, vps := range self.VPS {
		n += 5 + len(vps)
	}
	return
}

func (self AVCDecoderConfRecord) Marshal(b []byte, si SPSInfo) (n int) {
	b[0] = 1
	b[1] = self.AVCProfileIndication
	b[2] = self.ProfileCompatibility
	b[3] = self.AVCLevelIndication
	b[21] = 3
	b[22] = 3
	n += 23
	b[n] = (self.VPS[0][0] >> 1) & 0x3f
	n++
	b[n] = byte(len(self.VPS) >> 8)
	n++
	b[n] = byte(len(self.VPS))
	n++
	for _, vps := range self.VPS {
		pio.PutU16BE(b[n:], uint16(len(vps)))
		n += 2
		copy(b[n:], vps)
		n += len(vps)
	}
	b[n] = (self.SPS[0][0] >> 1) & 0x3f
	n++
	b[n] = byte(len(self.SPS) >> 8)
	n++
	b[n] = byte(len(self.SPS))
	n++
	for _, sps := range self.SPS {
		pio.PutU16BE(b[n:], uint16(len(sps)))
		n += 2
		copy(b[n:], sps)
		n += len(sps)
	}
	b[n] = (self.PPS[0][0] >> 1) & 0x3f
	n++
	b[n] = byte(len(self.PPS) >> 8)
	n++
	b[n] = byte(len(self.PPS))
	n++
	for _, pps := range self.PPS {
		pio.PutU16BE(b[n:], uint16(len(pps)))
		n += 2
		copy(b[n:], pps)
		n += len(pps)
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

	default:
		err = fmt.Errorf("h265parser: nal_unit_type=%d has no slice header", nal_unit_type)
		return
	}

	r := &bits.GolombBitReader{R: bytes.NewReader(packet[1:])}
	if _, err = r.ReadExponentialGolombCode(); err != nil {
		return
	}
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
