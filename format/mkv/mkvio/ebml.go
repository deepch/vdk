package mkvio

import (
	"io"
)

// Document represents a WebM file
type Document struct {
	r io.Reader
}

// ElementRegister contains the ID, type and name of the
// standard WebM/Matroska elements
type ElementRegister struct {
	ID   uint32
	Type uint8
	Name string
}

// Element is a Matroska/WebM/EBML element
type Element struct {
	ElementRegister

	Parent  *Element
	Level   int32
	Size    uint64
	Content []byte // Data contained in the element, nil if it is a master element
	Bytes   []byte // Whole binary representation of the element (nil if data is missing)
}
