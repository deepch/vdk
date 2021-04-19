package fmp4io

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/deepch/vdk/utils/bits/pio"
)

type Tag uint32

func (a Tag) String() string {
	var b [4]byte
	pio.PutU32BE(b[:], uint32(a))
	for i := 0; i < 4; i++ {
		if b[i] == 0 {
			b[i] = ' '
		}
	}
	return string(b[:])
}

type Atom interface {
	Pos() (int, int)
	Tag() Tag
	Marshal([]byte) int
	Unmarshal([]byte, int) (int, error)
	Len() int
	Children() []Atom
}

type AtomPos struct {
	Offset int
	Size   int
}

func (a AtomPos) Pos() (int, int) {
	return a.Offset, a.Size
}

func (a *AtomPos) setPos(offset int, size int) {
	a.Offset, a.Size = offset, size
}

type Dummy struct {
	Data []byte
	Tag_ Tag
	AtomPos
}

func (a Dummy) Children() []Atom {
	return nil
}

func (a Dummy) Tag() Tag {
	return a.Tag_
}

func (a Dummy) Len() int {
	return len(a.Data)
}

func (a Dummy) Marshal(b []byte) int {
	copy(b, a.Data)
	return len(a.Data)
}

func (a *Dummy) Unmarshal(b []byte, offset int) (n int, err error) {
	(&a.AtomPos).setPos(offset, len(b))
	a.Data = b
	n = len(b)
	return
}

type FullAtom struct {
	Version uint8
	Flags   uint32
	AtomPos
}

func (f FullAtom) marshalAtom(b []byte, tag Tag) (n int) {
	pio.PutU32BE(b[4:], uint32(tag))
	pio.PutU8(b[8:], f.Version)
	pio.PutU24BE(b[9:], f.Flags)
	return 12
}

func (f FullAtom) atomLen() int {
	return 12
}

func (f *FullAtom) unmarshalAtom(b []byte, offset int) (n int, err error) {
	f.AtomPos.setPos(offset, len(b))
	n = 8
	if len(b) < n+4 {
		return 0, parseErr("fullAtom", offset, nil)
	}
	f.Version = pio.U8(b[n:])
	f.Flags = pio.U24BE(b[n+1:])
	n += 4
	return
}

func StringToTag(tag string) Tag {
	var b [4]byte
	copy(b[:], []byte(tag))
	return Tag(pio.U32BE(b[:]))
}

func FindChildrenByName(root Atom, tag string) Atom {
	return FindChildren(root, StringToTag(tag))
}

func FindChildren(root Atom, tag Tag) Atom {
	if root.Tag() == tag {
		return root
	}
	for _, child := range root.Children() {
		if r := FindChildren(child, tag); r != nil {
			return r
		}
	}
	return nil
}

func ReadFileAtoms(r io.ReadSeeker) (atoms []Atom, err error) {
	for {
		offset, _ := r.Seek(0, 1)
		taghdr := make([]byte, 8)
		if _, err = io.ReadFull(r, taghdr); err != nil {
			if err == io.EOF {
				err = nil
			}
			return
		}
		size := pio.U32BE(taghdr[0:])
		tag := Tag(pio.U32BE(taghdr[4:]))

		var atom Atom
		switch tag {
		case FTYP:
			atom = &FileType{}
		case STYP:
			atom = &SegmentType{}
		case MOOV:
			atom = &Movie{}
		case MOOF:
			atom = &MovieFrag{}
		case SIDX:
			atom = &SegmentIndex{}
		}

		if atom != nil {
			b := make([]byte, int(size))
			if _, err = io.ReadFull(r, b[8:]); err != nil {
				return
			}
			copy(b, taghdr)
			if _, err = atom.Unmarshal(b, int(offset)); err != nil {
				return
			}
			atoms = append(atoms, atom)
		} else {
			dummy := &Dummy{Tag_: tag}
			dummy.setPos(int(offset), int(size))
			if _, err = r.Seek(int64(size)-8, 1); err != nil {
				return
			}
			atoms = append(atoms, dummy)
		}
	}
}

func printatom(out io.Writer, root Atom, depth int) {
	offset, size := root.Pos()

	type stringintf interface {
		String() string
	}

	fmt.Fprintf(out,
		"%s%s offset=%d size=%d",
		strings.Repeat(" ", depth*2), root.Tag(), offset, size,
	)
	if str, ok := root.(stringintf); ok {
		fmt.Fprint(out, " ", str.String())
	}
	fmt.Fprintln(out)

	children := root.Children()
	for _, child := range children {
		printatom(out, child, depth+1)
	}
}

func FprintAtom(out io.Writer, root Atom) {
	printatom(out, root, 0)
}

func PrintAtom(root Atom) {
	FprintAtom(os.Stdout, root)
}
