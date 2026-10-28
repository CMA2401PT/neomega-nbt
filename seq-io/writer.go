package seq_io

import "io"

type Writer interface {
	// must fully writed
	Write(b []byte) (err error)
	WriteByte(byte) error
}

// OffsetWriter is a wrapper around an io.Writer which keeps track of the amount of bytes written, so that it
// may be used in errors.
type OffsetWriter struct {
	w         io.Writer
	writeByte func(byte) error
}

func NewOffsetWriter(w io.Writer) OffsetWriter {
	var writer OffsetWriter
	if byteWriter, ok := w.(io.ByteWriter); ok {
		writer = OffsetWriter{w: w, writeByte: byteWriter.WriteByte}
	} else {
		writer = OffsetWriter{w: w, writeByte: func(b byte) error {
			var buf [1]byte
			buf[0] = b
			n, err := w.Write(buf[:])
			if err == nil && n != 1 {
				return io.ErrShortWrite
			}
			return err
		}}
	}
	return writer
}

// Write writes a byte slice to the underlying io.Writer. It increases the byte offset by exactly n.
func (w OffsetWriter) Write(b []byte) (err error) {
	n, err := w.w.Write(b)
	if err == nil && n != len(b) {
		return io.ErrShortWrite
	}
	return err
}

func (w OffsetWriter) WriteByte(b byte) (err error) {
	return w.writeByte(b)
}

type SliceGrowWriter struct {
	buf []byte
}

// NewSliceGrowWriter creates an empty writer.
// If you know an upper bound, pass a non-zero capHint to reduce realloc.
func NewSliceGrowWriter(initBuf []byte) *SliceGrowWriter {
	if initBuf == nil {
		initBuf = make([]byte, 0)
	}
	return &SliceGrowWriter{buf: initBuf}
}

// Bytes returns the written bytes view (no copy).
func (w *SliceGrowWriter) Bytes() []byte { return w.buf }

// Reset keeps capacity but clears content.
func (w *SliceGrowWriter) Reset() { w.buf = w.buf[:0] }

// Len returns current length.
func (w *SliceGrowWriter) Len() int { return len(w.buf) }

// Cap returns current capacity.
func (w *SliceGrowWriter) Cap() int { return cap(w.buf) }

func (w *SliceGrowWriter) grow(need int) {
	// need: extra bytes required
	oldLen := len(w.buf)
	newLen := oldLen + need

	if newLen <= cap(w.buf) {
		// just extend length
		w.buf = w.buf[:newLen]
		return
	}

	// exponential growth
	newCap := cap(w.buf)
	if newCap == 0 {
		newCap = 64
	}
	for newCap < newLen {
		newCap *= 2
	}

	nb := make([]byte, newLen, newCap)
	copy(nb, w.buf)
	w.buf = nb
}

func (w *SliceGrowWriter) Write(b []byte) error {
	if len(b) == 0 {
		return nil
	}
	oldLen := len(w.buf)
	w.grow(len(b))
	copy(w.buf[oldLen:], b)
	return nil
}

func (w *SliceGrowWriter) WriteByte(x byte) error {
	oldLen := len(w.buf)
	w.grow(1)
	w.buf[oldLen] = x
	return nil
}
