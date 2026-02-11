package encoder_nbt

import "io"

type Writer interface {
	io.Writer
	WriteByte(byte) error
}

// offsetWriter is a wrapper around an io.Writer which keeps track of the amount of bytes written, so that it
// may be used in errors.
type offsetWriter struct {
	io.Writer
	writeByte func(byte) error
}

// Write writes a byte slice to the underlying io.Writer. It increases the byte offset by exactly n.
func (w offsetWriter) Write(b []byte) (n int, err error) {
	n, err = w.Writer.Write(b)
	return
}

func (w offsetWriter) WriteByte(b byte) (err error) {
	return w.writeByte(b)
}
