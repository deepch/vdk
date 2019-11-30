package mp4f

import "github.com/deepch/vdk/format/mp4/mp4io"

type FDummy struct {
	Data []byte
	Tag_ mp4io.Tag
	mp4io.AtomPos
}

func (self FDummy) Children() []mp4io.Atom {
	return nil
}

func (self FDummy) Tag() mp4io.Tag {
	return self.Tag_
}

func (self FDummy) Len() int {
	return len(self.Data)
}

func (self FDummy) Marshal(b []byte) int {
	copy(b, self.Data)
	return len(self.Data)
}

func (self FDummy) Unmarshal(b []byte, offset int) (n int, err error) {
	return
}
