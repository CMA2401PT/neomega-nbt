package decoder_nbt

import (
	"io"
)

type Reader interface {
	// must fully read p, otherwise return err
	Read(p []byte) (err error)
	ReadByte() (byte, error)
}

// offsetReader is a wrapper around an io.Reader, used to track the offset (amount of bytes read) of the data
// that is being read, so that errors may have offset data.
type offsetReader struct {
	io.Reader

	// ReadByte is a function provided by offsetReader if the io.Reader does not implement io.ByteReader.
	readByte func() (byte, error)
}

// newOffsetReader returns a new offset reader for the io.Reader passed, setting the ReadByte and Next
// functions as appropriate for that particular reader.
func newOffsetReader(r io.Reader) offsetReader {
	reader := offsetReader{Reader: r}
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
func (b offsetReader) Read(p []byte) (err error) {
	_, err = io.ReadAtLeast(b.Reader, p, len(p))
	return
}

func (b offsetReader) ReadByte() (byte, error) {
	return b.readByte()
}
