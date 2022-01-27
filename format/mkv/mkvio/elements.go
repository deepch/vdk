package mkvio

const (
	ElementType        uint8 = 0x0
	ElementTypeUnknown uint8 = 0x0
	ElementTypeMaster  uint8 = 0x1
	ElementTypeUint    uint8 = 0x2
	ElementTypeInt     uint8 = 0x3
	ElementTypeString  uint8 = 0x4
	ElementTypeUnicode uint8 = 0x5
	ElementTypeBinary  uint8 = 0x6
	ElementTypeFloat   uint8 = 0x7
	ElementTypeDate    uint8 = 0x8
)

var (
	ElementUnknown                     = ElementRegister{0x0, ElementTypeUnknown, "Unknown"}
	ElementEBML                        = ElementRegister{0x1a45dfa3, ElementTypeMaster, "EBML"}
	ElementEBMLVersion                 = ElementRegister{0x4286, ElementTypeUint, "EBMLVersion"}
	ElementEBMLReadVersion             = ElementRegister{0x42f7, ElementTypeUint, "EBMLReadVersion"}
	ElementEBMLMaxIDLength             = ElementRegister{0x42f2, ElementTypeUint, "EBMLMaxIDLength"}
	ElementEBMLMaxSizeLength           = ElementRegister{0x42f3, ElementTypeUint, "EBMLMaxSizeLength"}
	ElementDocType                     = ElementRegister{0x4282, ElementTypeString, "DocType"}
	ElementDocTypeVersion              = ElementRegister{0x4287, ElementTypeUint, "DocTypeVersion"}
	ElementDocTypeReadVersion          = ElementRegister{0x4285, ElementTypeUint, "DocTypeReadVersion"}
	ElementVoid                        = ElementRegister{0xec, ElementTypeBinary, "Void"}
	ElementCRC32                       = ElementRegister{0xbf, ElementTypeBinary, "CRC-32"}
	ElementSegment                     = ElementRegister{0x18538067, ElementTypeMaster, "Segment"}
	ElementSeekHead                    = ElementRegister{0x114d9b74, ElementTypeMaster, "SeekHead"}
	ElementSeek                        = ElementRegister{0x4dbb, ElementTypeMaster, "Seek"}
	ElementSeekID                      = ElementRegister{0x53ab, ElementTypeBinary, "SeekID"}
	ElementSeekPosition                = ElementRegister{0x53ac, ElementTypeUint, "SeekPosition"}
	ElementInfo                        = ElementRegister{0x1549a966, ElementTypeMaster, "Info"}
	ElementSegmentUID                  = ElementRegister{0x73a4, ElementTypeBinary, "SegmentUID"}
	ElementSegmentFilename             = ElementRegister{0x7384, ElementTypeUnicode, "SegmentFilename"}
	ElementPrevUID                     = ElementRegister{0x3cb923, ElementTypeBinary, "PrevUID"}
	ElementPrevFilename                = ElementRegister{0x3c83ab, ElementTypeUnicode, "PrevFilename"}
	ElementNextUID                     = ElementRegister{0x3eb923, ElementTypeBinary, "NextUID"}
	ElementNextFilename                = ElementRegister{0x3e83bb, ElementTypeUnicode, "NextFilename"}
	ElementSegmentFamily               = ElementRegister{0x4444, ElementTypeBinary, "SegmentFamily"}
	ElementChapterTranslate            = ElementRegister{0x6924, ElementTypeMaster, "ChapterTranslate"}
	ElementChapterTranslateEditionUID  = ElementRegister{0x69fc, ElementTypeUint, "ChapterTranslateEditionUID"}
	ElementChapterTranslateCodec       = ElementRegister{0x69bf, ElementTypeUint, "ChapterTranslateCodec"}
	ElementChapterTranslateID          = ElementRegister{0x69a5, ElementTypeBinary, "ChapterTranslateID"}
	ElementTimecodeScale               = ElementRegister{0x2ad7b1, ElementTypeUint, "TimecodeScale"}
	ElementDuration                    = ElementRegister{0x4489, ElementTypeFloat, "Duration"}
	ElementDateUTC                     = ElementRegister{0x4461, ElementTypeDate, "DateUTC"}
	ElementTitle                       = ElementRegister{0x7ba9, ElementTypeUnicode, "Title"}
	ElementMuxingApp                   = ElementRegister{0x4d80, ElementTypeUnicode, "MuxingApp"}
	ElementWritingApp                  = ElementRegister{0x5741, ElementTypeUnicode, "WritingApp"}
	ElementCluster                     = ElementRegister{0x1f43b675, ElementTypeMaster, "Cluster"}
	ElementTimecode                    = ElementRegister{0xe7, ElementTypeUint, "Timecode"}
	ElementSlientTracks                = ElementRegister{0x5854, ElementTypeMaster, "SlientTracks"}
	ElementSlientTrackNumber           = ElementRegister{0x58d7, ElementTypeUint, "SlientTrackNumber"}
	ElementPosition                    = ElementRegister{0xa7, ElementTypeUint, "Position"}
	ElementPrevSize                    = ElementRegister{0xab, ElementTypeUint, "PrevSize"}
	ElementSimpleBlock                 = ElementRegister{0xa3, ElementTypeBinary, "SimpleBlock"}
	ElementBlockGroup                  = ElementRegister{0xa0, ElementTypeMaster, "BlockGroup"}
	ElementBlock                       = ElementRegister{0xa1, ElementTypeBinary, "Block"}
	ElementBlockAdditions              = ElementRegister{0x75a1, ElementTypeMaster, "BlockAdditions"}
	ElementBlockMore                   = ElementRegister{0xa6, ElementTypeMaster, "BlockMore"}
	ElementBlockAddID                  = ElementRegister{0xee, ElementTypeUint, "BlockAddID"}
	ElementBlockAdditional             = ElementRegister{0xa5, ElementTypeBinary, "BlockAdditional"}
	ElementBlockDuration               = ElementRegister{0x9b, ElementTypeUint, "BlockDuration"}
	ElementReferencePriority           = ElementRegister{0xfa, ElementTypeUint, "ReferencePriority"}
	ElementReferenceBlock              = ElementRegister{0xfb, ElementTypeInt, "ReferenceBlock"}
	ElementCodecState                  = ElementRegister{0xa4, ElementTypeBinary, "CodecState"}
	ElementDiscardPadding              = ElementRegister{0x75a2, ElementTypeInt, "DiscardPadding"}
	ElementSlices                      = ElementRegister{0x8e, ElementTypeMaster, "Slices"}
	ElementTimeSlice                   = ElementRegister{0xe8, ElementTypeMaster, "TimeSlice"}
	ElementLaceNumber                  = ElementRegister{0xcc, ElementTypeUint, "LaceNumber"}
	ElementTracks                      = ElementRegister{0x1654ae6b, ElementTypeMaster, "Tracks"}
	ElementTrackEntry                  = ElementRegister{0xae, ElementTypeMaster, "TrackEntry"}
	ElementTrackNumber                 = ElementRegister{0xd7, ElementTypeUint, "TrackNumber"}
	ElementTrackUID                    = ElementRegister{0x73c5, ElementTypeUint, "TrackUID"}
	ElementTrackType                   = ElementRegister{0x83, ElementTypeUint, "TrackType"}
	ElementFlagEnabled                 = ElementRegister{0xb9, ElementTypeUint, "FlagEnabled"}
	ElementFlagDefault                 = ElementRegister{0x88, ElementTypeUint, "FlagDefault"}
	ElementFlagForced                  = ElementRegister{0x55aa, ElementTypeUint, "FlagForced"}
	ElementFlagLacing                  = ElementRegister{0x9c, ElementTypeUint, "FlagLacing"}
	ElementMinCache                    = ElementRegister{0x6de7, ElementTypeUint, "MinCache"}
	ElementMaxCache                    = ElementRegister{0x6df8, ElementTypeUint, "MaxCache"}
	ElementDefaultDuration             = ElementRegister{0x23e383, ElementTypeUint, "DefaultDuration"}
	ElementDefaultDecodedFieldDuration = ElementRegister{0x234e7a, ElementTypeUint, "DefaultDecodedFieldDuration"}
	ElementMaxBlockAdditionID          = ElementRegister{0x55ee, ElementTypeUint, "MaxBlockAdditionID"}
	ElementName                        = ElementRegister{0x536e, ElementTypeUnicode, "Name"}
	ElementLanguage                    = ElementRegister{0x22b59c, ElementTypeString, "Language"}
	ElementCodecID                     = ElementRegister{0x86, ElementTypeString, "CodecID"}
	ElementCodecPrivate                = ElementRegister{0x63a2, ElementTypeBinary, "CodecPrivate"}
	ElementCodecName                   = ElementRegister{0x258688, ElementTypeUnicode, "CodecName"}
	ElementAttachmentLink              = ElementRegister{0x7446, ElementTypeUint, "AttachmentLink"}
	ElementCodecDecodeAll              = ElementRegister{0xaa, ElementTypeUint, "CodecDecodeAll"}
	ElementTrackOverlay                = ElementRegister{0x6fab, ElementTypeUint, "TrackOverlay"}
	ElementCodecDelay                  = ElementRegister{0x56aa, ElementTypeUint, "CodecDelay"}
	ElementSeekPreRoll                 = ElementRegister{0x56bb, ElementTypeUint, "SeekPreRoll"}
	ElementTrackTranslate              = ElementRegister{0x6624, ElementTypeMaster, "TrackTranslate"}
	ElementTrackTranslateEditionUID    = ElementRegister{0x66fc, ElementTypeUint, "TrackTranslateEditionUID"}
	ElementTrackTranslateCodec         = ElementRegister{0x66bf, ElementTypeUint, "TrackTranslateCodec"}
	ElementTrackTranslateTrackID       = ElementRegister{0x66a5, ElementTypeBinary, "TrackTranslateTrackID"}
	ElementVideo                       = ElementRegister{0xe0, ElementTypeMaster, "Video"}
	ElementFlagInterlaced              = ElementRegister{0x9a, ElementTypeUint, "FlagInterlaced"}
	ElementStereoMode                  = ElementRegister{0x53b8, ElementTypeUint, "StereoMode"}
	ElementAlphaMode                   = ElementRegister{0x53c0, ElementTypeUint, "AlphaMode"}
	ElementPixelWidth                  = ElementRegister{0xb0, ElementTypeUint, "PixelWidth"}
	ElementPixelHeight                 = ElementRegister{0xba, ElementTypeUint, "PixelHeight"}
	ElementPixelCropBottom             = ElementRegister{0x54aa, ElementTypeUint, "PixelCropBottom"}
	ElementPixelCropTop                = ElementRegister{0x54bb, ElementTypeUint, "PixelCropTop"}
	ElementPixelCropLeft               = ElementRegister{0x54cc, ElementTypeUint, "PixelCropLeft"}
	ElementPixelCropRight              = ElementRegister{0x54dd, ElementTypeUint, "PixelCropRight"}
	ElementDisplayWidth                = ElementRegister{0x54b0, ElementTypeUint, "DisplayWidth"}
	ElementDisplayHeight               = ElementRegister{0x54ba, ElementTypeUint, "DisplayHeight"}
	ElementDisplayUint                 = ElementRegister{0x54b2, ElementTypeUint, "DisplayUint"}
	ElementAspectRatioType             = ElementRegister{0x54b3, ElementTypeUint, "AspectRatioType"}
	ElementColourSpace                 = ElementRegister{0x2eb524, ElementTypeBinary, "ColourSpace"}
	ElementAudio                       = ElementRegister{0xe1, ElementTypeMaster, "Audio"}
	ElementSamplingFrequency           = ElementRegister{0xb5, ElementTypeFloat, "SamplingFrequency"}
	ElementOutputSamplingFrequency     = ElementRegister{0x78b5, ElementTypeFloat, "OutputSamplingFrequency"}
	ElementChannels                    = ElementRegister{0x9f, ElementTypeUint, "Channels"}
	ElementBitDepth                    = ElementRegister{0x6264, ElementTypeUint, "BitDepth"}
	ElementTrackOperation              = ElementRegister{0xe2, ElementTypeMaster, "TrackOperation"}
	ElementTrackCombinePlanes          = ElementRegister{0xe3, ElementTypeMaster, "TrackCombinePlanes"}
	ElementTrackPlane                  = ElementRegister{0xe4, ElementTypeMaster, "TrackPlane"}
	ElementTrackPlaneUID               = ElementRegister{0xe5, ElementTypeUint, "TrackPlaneUID"}
	ElementTrackPlaneType              = ElementRegister{0xe6, ElementTypeUint, "TrackPlaneType"}
	ElementTrackJoinBlocks             = ElementRegister{0xe9, ElementTypeMaster, "TrackJoinBlocks"}
	ElementTrackJoinUID                = ElementRegister{0xed, ElementTypeUint, "TrackJoinUID"}
	ElementContentEncodings            = ElementRegister{0x6d80, ElementTypeMaster, "ContentEncodings"}
	ElementContentEncoding             = ElementRegister{0x6240, ElementTypeMaster, "ContentEncoding"}
	ElementContentEncodingOrder        = ElementRegister{0x5031, ElementTypeUint, "ContentEncodingOrder"}
	ElementContentEncodingScope        = ElementRegister{0x5032, ElementTypeUint, "ContentEncodingScope"}
	ElementContentEncodingType         = ElementRegister{0x5033, ElementTypeUint, "ContentEncodingType"}
	ElementContentCompression          = ElementRegister{0x5034, ElementTypeMaster, "ContentCompression"}
	ElementContentCompAlgo             = ElementRegister{0x4254, ElementTypeUint, "ContentCompAlgo"}
	ElementContentCompSettings         = ElementRegister{0x4255, ElementTypeBinary, "ContentCompSettings"}
	ElementContentEncryption           = ElementRegister{0x5035, ElementTypeMaster, "ContentEncryption"}
	ElementContentEncAlgo              = ElementRegister{0x47e1, ElementTypeUint, "ContentEncAlgo"}
	ElementContentEncKeyID             = ElementRegister{0x47e2, ElementTypeUint, "ContentEncKeyID"}
	ElementContentSignature            = ElementRegister{0x47e3, ElementTypeBinary, "ContentSignature"}
	ElementContentSigKeyID             = ElementRegister{0x47e4, ElementTypeBinary, "ContentSigKeyID"}
	ElementContentSigAlgo              = ElementRegister{0x47e5, ElementTypeUint, "ContentSigAlgo"}
	ElementContentSigHashAlgo          = ElementRegister{0x47e6, ElementTypeUint, "ContentSigHashAlgo"}
	ElementCues                        = ElementRegister{0x1c53bb6b, ElementTypeMaster, "Cues"}
	ElementCuePoint                    = ElementRegister{0xbb, ElementTypeMaster, "CuePoint"}
	ElementCueTime                     = ElementRegister{0xb3, ElementTypeUint, "CueTime"}
	ElementCueTrackPositions           = ElementRegister{0xb7, ElementTypeMaster, "CueTrackPositions"}
	ElementCueTrack                    = ElementRegister{0xf7, ElementTypeUint, "CueTrack"}
	ElementCueClusterPosition          = ElementRegister{0xf1, ElementTypeUint, "CueClusterPosition"}
	ElementCueRelativePosition         = ElementRegister{0xf0, ElementTypeUint, "CueRelativePosition"}
	ElementCueDuration                 = ElementRegister{0xb2, ElementTypeUint, "CueDuration"}
	ElementCueBlockNumber              = ElementRegister{0x5378, ElementTypeUint, "CueBlockNumber"}
	ElementCueCodecState               = ElementRegister{0xea, ElementTypeUint, "CueCodecState"}
	ElementCueReference                = ElementRegister{0xdb, ElementTypeMaster, "CueReference"}
	ElementCueRefTime                  = ElementRegister{0x96, ElementTypeUint, "CueRefTime"}
	ElementAttachments                 = ElementRegister{0x1941a469, ElementTypeMaster, "Attachments"}
	ElementAttachedFile                = ElementRegister{0x61a7, ElementTypeMaster, "AttachedFile"}
	ElementFileDescription             = ElementRegister{0x467e, ElementTypeUnicode, "FileDescription"}
	ElementFileName                    = ElementRegister{0x466e, ElementTypeUnicode, "FileName"}
	ElementFileMimeType                = ElementRegister{0x6460, ElementTypeString, "FileMimeType"}
	ElementFileData                    = ElementRegister{0x465c, ElementTypeBinary, "FileData"}
	ElementFileUID                     = ElementRegister{0x46ae, ElementTypeUint, "FileUID"}
	ElementChapters                    = ElementRegister{0x1043a770, ElementTypeMaster, "Chapters"}
	ElementEditionEntry                = ElementRegister{0x45b9, ElementTypeMaster, "EditionEntry"}
	ElementEditionUID                  = ElementRegister{0x45bc, ElementTypeUint, "EditionUID"}
	ElementEditionFlagHidden           = ElementRegister{0x45bd, ElementTypeUint, "EditionFlagHidden"}
	ElementEditionFlagDefault          = ElementRegister{0x45db, ElementTypeUint, "EditionFlagDefault"}
	ElementEditionFlagOrdered          = ElementRegister{0x45dd, ElementTypeUint, "EditionFlagOrdered"}
	ElementChapterAtom                 = ElementRegister{0xb6, ElementTypeMaster, "ChapterAtom"}
	ElementChapterUID                  = ElementRegister{0x73c4, ElementTypeUint, "ChapterUID"}
	ElementChapterStringUID            = ElementRegister{0x5654, ElementTypeUnicode, "ChapterStringUID"}
	ElementChapterTimeStart            = ElementRegister{0x91, ElementTypeUint, "ChapterTimeStart"}
	ElementChapterTimeEnd              = ElementRegister{0x92, ElementTypeUint, "ChapterTimeEnd"}
	ElementChapterFlagHidden           = ElementRegister{0x98, ElementTypeUint, "ChapterFlagHidden"}
	ElementChapterFlagEnabled          = ElementRegister{0x4598, ElementTypeUint, "ChapterFlagEnabled"}
	ElementChapterSegmentUID           = ElementRegister{0x6e67, ElementTypeBinary, "ChapterSegmentUID"}
	ElementChapterSegmentEditionUID    = ElementRegister{0x6ebc, ElementTypeUint, "ChapterSegmentEditionUID"}
	ElementChapterPhysicalEquiv        = ElementRegister{0x63c3, ElementTypeUint, "ChapterPhysicalEquiv"}
	ElementChapterTrack                = ElementRegister{0x8f, ElementTypeMaster, "ChapterTrack"}
	ElementChapterTrackNumber          = ElementRegister{0x89, ElementTypeUint, "ChapterTrackNumber"}
	ElementChapterDisplay              = ElementRegister{0x80, ElementTypeMaster, "ChapterDisplay"}
	ElementChapString                  = ElementRegister{0x85, ElementTypeUnicode, "ChapString"}
	ElementChapLanguage                = ElementRegister{0x437c, ElementTypeString, "ChapLanguage"}
	ElementChapCountry                 = ElementRegister{0x437e, ElementTypeString, "ChapCountry"}
	ElementChapProcess                 = ElementRegister{0x6944, ElementTypeMaster, "ChapProcess"}
	ElementChapProcessCodecID          = ElementRegister{0x6955, ElementTypeUint, "ChapProcessCodecID"}
	ElementChapProcessPrivate          = ElementRegister{0x450d, ElementTypeBinary, "ChapProcessPrivate"}
	ElementChapProcessCommand          = ElementRegister{0x6911, ElementTypeMaster, "ChapProcessCommand"}
	ElementChapProcessTime             = ElementRegister{0x6922, ElementTypeUint, "ChapProcessTime"}
	ElementChapProcessData             = ElementRegister{0x6933, ElementTypeBinary, "ChapProcessData"}

	// TODO: Add tags
)

// GetElementRegister returns the infos concerning the provided element ID
func GetElementRegister(id uint32) ElementRegister {
	switch id {
	case ElementEBML.ID:
		return ElementEBML
	case ElementEBMLVersion.ID:
		return ElementEBMLVersion
	case ElementEBMLReadVersion.ID:
		return ElementEBMLReadVersion
	case ElementEBMLMaxIDLength.ID:
		return ElementEBMLMaxIDLength
	case ElementEBMLMaxSizeLength.ID:
		return ElementEBMLMaxSizeLength
	case ElementDocType.ID:
		return ElementDocType
	case ElementDocTypeVersion.ID:
		return ElementDocTypeVersion
	case ElementDocTypeReadVersion.ID:
		return ElementDocTypeReadVersion
	case ElementVoid.ID:
		return ElementVoid
	case ElementCRC32.ID:
		return ElementCRC32
	case ElementSegment.ID:
		return ElementSegment
	case ElementSeekHead.ID:
		return ElementSeekHead
	case ElementSeek.ID:
		return ElementSeek
	case ElementSeekID.ID:
		return ElementSeekID
	case ElementSeekPosition.ID:
		return ElementSeekPosition
	case ElementInfo.ID:
		return ElementInfo
	case ElementTimecodeScale.ID:
		return ElementTimecodeScale
	case ElementDuration.ID:
		return ElementDuration
	case ElementDateUTC.ID:
		return ElementDateUTC
	case ElementTitle.ID:
		return ElementTitle
	case ElementMuxingApp.ID:
		return ElementMuxingApp
	case ElementWritingApp.ID:
		return ElementWritingApp
	case ElementCluster.ID:
		return ElementCluster
	case ElementTimecode.ID:
		return ElementTimecode
	case ElementPrevSize.ID:
		return ElementPrevSize
	case ElementSimpleBlock.ID:
		return ElementSimpleBlock
	case ElementBlockGroup.ID:
		return ElementBlockGroup
	case ElementBlock.ID:
		return ElementBlock
	case ElementBlockAdditions.ID:
		return ElementBlockAdditions
	case ElementBlockMore.ID:
		return ElementBlockMore
	case ElementBlockAddID.ID:
		return ElementBlockAddID
	case ElementBlockAdditional.ID:
		return ElementBlockAdditional
	case ElementBlockDuration.ID:
		return ElementBlockDuration
	case ElementReferenceBlock.ID:
		return ElementReferenceBlock
	case ElementDiscardPadding.ID:
		return ElementDiscardPadding
	case ElementTracks.ID:
		return ElementTracks
	case ElementTrackEntry.ID:
		return ElementTrackEntry
	case ElementTrackNumber.ID:
		return ElementTrackNumber
	case ElementTrackUID.ID:
		return ElementTrackUID
	case ElementTrackType.ID:
		return ElementTrackType
	case ElementFlagEnabled.ID:
		return ElementFlagEnabled
	case ElementFlagDefault.ID:
		return ElementFlagDefault
	case ElementFlagForced.ID:
		return ElementFlagForced
	case ElementFlagLacing.ID:
		return ElementFlagLacing
	case ElementDefaultDuration.ID:
		return ElementDefaultDuration
	case ElementName.ID:
		return ElementName
	case ElementLanguage.ID:
		return ElementLanguage
	case ElementCodecID.ID:
		return ElementCodecID
	case ElementCodecPrivate.ID:
		return ElementCodecPrivate
	case ElementCodecName.ID:
		return ElementCodecName
	case ElementCodecDelay.ID:
		return ElementCodecDelay
	case ElementSeekPreRoll.ID:
		return ElementSeekPreRoll
	case ElementVideo.ID:
		return ElementVideo
	case ElementFlagInterlaced.ID:
		return ElementFlagInterlaced
	case ElementStereoMode.ID:
		return ElementStereoMode
	case ElementAlphaMode.ID:
		return ElementAlphaMode
	case ElementPixelWidth.ID:
		return ElementPixelWidth
	case ElementPixelHeight.ID:
		return ElementPixelHeight
	case ElementPixelCropBottom.ID:
		return ElementPixelCropBottom
	case ElementPixelCropTop.ID:
		return ElementPixelCropTop
	case ElementPixelCropLeft.ID:
		return ElementPixelCropLeft
	case ElementPixelCropRight.ID:
		return ElementPixelCropRight
	case ElementDisplayWidth.ID:
		return ElementDisplayWidth
	case ElementDisplayHeight.ID:
		return ElementDisplayHeight
	case ElementDisplayUint.ID:
		return ElementDisplayUint
	case ElementAspectRatioType.ID:
		return ElementAspectRatioType
	case ElementAudio.ID:
		return ElementAudio
	case ElementSamplingFrequency.ID:
		return ElementSamplingFrequency
	case ElementOutputSamplingFrequency.ID:
		return ElementOutputSamplingFrequency
	case ElementChannels.ID:
		return ElementChannels
	case ElementBitDepth.ID:
		return ElementBitDepth
	case ElementContentEncodings.ID:
		return ElementContentEncodings
	case ElementContentEncoding.ID:
		return ElementContentEncoding
	case ElementContentEncodingOrder.ID:
		return ElementContentEncodingOrder
	case ElementContentEncodingScope.ID:
		return ElementContentEncodingScope
	case ElementContentEncodingType.ID:
		return ElementContentEncodingType
	case ElementContentEncryption.ID:
		return ElementContentEncryption
	case ElementContentEncAlgo.ID:
		return ElementContentEncAlgo
	case ElementContentEncKeyID.ID:
		return ElementContentEncKeyID
	case ElementUnknown.ID:
		return ElementUnknown
	default:
		return ElementUnknown
	}
}
