package mkvio

import (
	"errors"
	"io"
)

var (
	ErrParse         = errors.New("Parse error")
	ErrUnexpectedEOF = errors.New("Unexpected EOF")
)

// InitDocument creates a MKV/WebM document containing the file data
// It does not do any parsing
func InitDocument(r io.Reader) *Document {
	doc := new(Document)
	doc.r = r

	return doc
}

// ParseAll parses the entire MKV/WebM document
// When an EBML/WebM element is encountered, it calls the provided function
// and passes the newly parsed element
func (doc *Document) ParseAll(c func(Element)) error {
	for {
		el, err := doc.ParseElement()
		if err != nil {
			return err
		}

		c(el)
	}

	return nil
}

func (doc *Document) GetVideoCodec() (*Element, error) {
	for {

		el, err := doc.ParseElement()

		if err != nil {
			return nil, err
		}

		if el.ElementRegister.ID == ElementCodecPrivate.ID {
			return &el, nil
		}

	}

	return nil, errors.New("not found")
}

// ParseElement parses an EBML element starting at the document's current cursor position.
// Because of its nature, it does not set the elements's parent or level.
func (doc *Document) ParseElement() (Element, error) {
	var el Element

	id, err := doc.GetElementID(&el)
	if err != nil {
		return el, err
	}

	size, err := doc.GetElementSize(&el)
	if err != nil {
		return el, err
	}

	reg := GetElementRegister(id)
	el.ID = reg.ID
	el.Type = reg.Type
	el.Name = reg.Name
	el.Size = size

	if el.Type != ElementTypeMaster {
		d, err := doc.GetElementContent(&el)
		if err != nil {
			return el, err
		}

		el.Content = d
	}

	return el, nil
}

// GetElementID tries to parse the next element's id,
// starting from the document's current cursor position.
func (doc *Document) GetElementID(el *Element) (uint32, error) {
	b := make([]byte, 1)

	_, err := io.ReadFull(doc.r, b)
	if err != nil {
		return 0, err
	}

	if ((b[0] & 0x80) >> 7) == 1 { // Class A ID (on 1 byte)
		el.Bytes = append(el.Bytes, b[0])
		return uint32(b[0]), nil
	}
	if ((b[0] & 0x40) >> 6) == 1 { // Class B ID (on 2 byte)
		bb := make([]byte, 2)
		copy(bb, b)

		_, err = io.ReadFull(doc.r, bb[1:])
		if err != nil {
			return 0, err
		}

		el.Bytes = append(el.Bytes, bb...)
		return uint32(pack(2, bb)), nil
	}
	if ((b[0] & 0x20) >> 5) == 1 { // Class C ID (on 3 bytes)
		bb := make([]byte, 3)
		copy(bb, b)

		_, err = io.ReadFull(doc.r, bb[1:])
		if err != nil {
			return 0, err
		}

		el.Bytes = append(el.Bytes, bb...)
		return uint32(pack(3, bb)), nil
	}
	if ((b[0] & 0x10) >> 4) == 1 { // Class D ID (on 4 bytes)
		bb := make([]byte, 4)
		copy(bb, b)

		_, err = io.ReadFull(doc.r, bb[1:])
		if err != nil {
			return 0, err
		}

		el.Bytes = append(el.Bytes, bb...)
		return uint32(pack(4, bb)), nil
	}

	return 0, ErrParse
}

// GetElementSize tries to parse the next element's size,
// starting from the document's current cursor position.
func (doc *Document) GetElementSize(el *Element) (uint64, error) {
	b := make([]byte, 1)

	_, err := io.ReadFull(doc.r, b)
	if err != nil {
		return 0, err
	}

	var mask byte
	var length uint64

	if b[0] >= 0x80 {
		length = 1
		mask = 0x7f
	} else if b[0] >= 0x40 {
		length = 2
		mask = 0x3f
	} else if b[0] >= 0x20 {
		length = 3
		mask = 0x1f
	} else if b[0] >= 0x10 {
		length = 4
		mask = 0x0f
	} else if b[0] >= 0x08 {
		length = 5
		mask = 0x07
	} else if b[0] >= 0x04 {
		length = 6
		mask = 0x03
	} else if b[0] >= 0x02 {
		length = 7
		mask = 0x01
	} else if b[0] >= 0x01 {
		length = 8
		mask = 0x00
	} else {
		return 0, ErrParse
	}

	bb := make([]byte, length)
	bb[0] = b[0]

	if length > 1 {
		_, err = io.ReadFull(doc.r, bb[1:])
		if err != nil {
			return 0, err
		}
	}

	el.Bytes = append(el.Bytes, bb...)

	bb[0] &= mask
	v := pack(int(length), bb)

	return v, nil
}

// GetElementContent returns the element's data (if any)
// Data is present if the element's type is not Master
func (doc *Document) GetElementContent(el *Element) ([]byte, error) {
	buf := make([]byte, el.Size)

	_, err := io.ReadFull(doc.r, buf)
	if err != nil {
		return nil, err
	}

	el.Bytes = append(el.Bytes, buf...)
	return buf, nil
}
