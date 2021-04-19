package esio

type builder struct {
	buf []byte
}

func (b *builder) Bytes() []byte {
	return b.buf
}

// Grow the buffer by n bytes and return a slice holding the new area.
// The slice is only valid until the next method called on the builder.
func (b *builder) Grow(n int) []byte {
	pos := len(b.buf)
	b.buf = append(b.buf, make([]byte, n)...)
	return b.buf[pos:]
}

// WriteByte appends a uint8
func (b *builder) WriteByte(v byte) error {
	b.buf = append(b.buf, v)
	return nil
}

// WriteU16 appends a 16-but unsigned big-endian integer
func (b *builder) WriteU16(v uint16) {
	b.buf = append(b.buf, uint8(v>>8), uint8(v))
}

// WriteU24 appends a 24-bit unsigned big-endian integer
func (b *builder) WriteU24(v uint32) {
	b.buf = append(b.buf, uint8(v>>16), uint8(v>>8), uint8(v))
}

// WriteU32 appends a 32-bit unsigned big-endian integer
func (b *builder) WriteU32(v uint32) {
	b.buf = append(b.buf, uint8(v>>24), uint8(v>>16), uint8(v>>8), uint8(v))
}

// WriteU64 appends a 64-bit unsigned big-endian integer
func (b *builder) WriteU64(v uint64) {
	b.buf = append(b.buf,
		uint8(v>>56),
		uint8(v>>48),
		uint8(v>>40),
		uint8(v>>32),
		uint8(v>>24),
		uint8(v>>16),
		uint8(v>>8),
		uint8(v),
	)
}

// Write appends a slice. It never returns an error, but implements io.Writer
func (b *builder) Write(d []byte) (int, error) {
	b.buf = append(b.buf, d...)
	return len(d), nil
}

// Cursor allocates length bytes and returns a pointer that can be used to access the allocated region later, even after the buffer has grown
func (b *builder) Cursor(length int) cursor {
	c := cursor{builder: b, i: len(b.buf)}
	b.Grow(length)
	c.j = len(b.buf)
	return c
}

// Descriptor writes a descriptor tag and leaves room for a length later.
// Call DescriptorDone on the returned cursor to complete it.
func (b *builder) Descriptor(tag Tag) cursor {
	b.WriteByte(byte(tag))
	return b.Cursor(4)
}

type cursor struct {
	builder *builder
	i, j    int
}

func (c cursor) Bytes() []byte {
	return c.builder.buf[c.i:c.j]
}

// DescriptorDone completes a descriptor tag by writing the length of its contents.
// Either pass the length of the contents, or -1 if the current end of the buffer is the end of the contents.
func (c cursor) DescriptorDone(length int) {
	if length < 0 {
		length = len(c.builder.buf) - c.j
	}
	buf := c.Bytes()
	for i := 3; i >= 0; i-- {
		v := byte(length >> uint(7*i) & 0x7f)
		if i != 0 {
			v |= 0x80
		}
		buf[3-i] = v
	}
}
