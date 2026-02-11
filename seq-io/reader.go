package seq_io

import (
	"errors"
	"io"
)

type Reader interface {
	// must fully read p, otherwise return err
	Read(p []byte) (err error)
	ReadByte() (byte, error)
}

// OffsetReader is a wrapper around an io.Reader, used to track the offset (amount of bytes read) of the data
// that is being read, so that errors may have offset data.
type OffsetReader struct {
	io.Reader
	// ReadByte is a function provided by offsetReader if the io.Reader does not implement io.ByteReader.
	readByte func() (byte, error)
}

// newOffsetReader returns a new offset reader for the io.Reader passed, setting the ReadByte and Next
// functions as appropriate for that particular reader.
func NewOffsetReader(r io.Reader) OffsetReader {
	reader := OffsetReader{Reader: r}
	if byteReader, ok := r.(io.ByteReader); ok {
		reader.readByte = func() (byte, error) {
			return byteReader.ReadByte()
		}
	} else {
		reader.readByte = func() (byte, error) {
			var buf [1]byte
			_, err := io.ReadAtLeast(r, buf[:], 1)
			return buf[0], err
		}
	}
	return reader
}

// Read reads from the io.Reader and increases the reader's offset by exactly n.
func (b OffsetReader) Read(p []byte) (err error) {
	_, err = io.ReadAtLeast(b.Reader, p, len(p))
	return
}

func (b OffsetReader) ReadByte() (byte, error) {
	return b.readByte()
}

var ErrUnexpectedEOF = errors.New("unexpected EOF")

type ByteReader[Seq ~string | ~[]uint8] struct {
	b Seq
	i int
}

func NewByteReaderFromBytes[Seq ~string | ~[]uint8](data Seq) *ByteReader[Seq] {
	return &ByteReader[Seq]{b: data}
}

func (r *ByteReader[Seq]) Remaining() int {
	return len(r.b) - r.i
}

func (r *ByteReader[Seq]) Read(p []byte) error {
	n := len(p)
	if n == 0 {
		return nil
	}
	if r.i+n > len(r.b) {
		// 必须读满，否则报错
		return ErrUnexpectedEOF
	}
	copy(p, r.b[r.i:r.i+n])
	r.i += n
	return nil
}

func (r *ByteReader[Seq]) ReadByte() (byte, error) {
	if r.i >= len(r.b) {
		return 0, ErrUnexpectedEOF
	}
	v := r.b[r.i]
	r.i++
	return v, nil
}
