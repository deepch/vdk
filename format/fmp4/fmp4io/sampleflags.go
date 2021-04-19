package fmp4io

type SampleFlags uint32

// fragment sample flags
const (
	SampleIsNonSync       SampleFlags = 0x00010000
	SampleHasDependencies SampleFlags = 0x01000000
	SampleNoDependencies  SampleFlags = 0x02000000

	SampleNonKeyframe = SampleHasDependencies | SampleIsNonSync
)
