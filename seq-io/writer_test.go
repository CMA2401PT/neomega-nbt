package seq_io_test

import (
	"errors"
	"io"
	seq "neomega-nbt/seq-io"
	"testing"
)

type shortWriter struct{}

func (shortWriter) Write(p []byte) (int, error) { return len(p) - 1, nil }

func TestOffsetWriterReportsShortWrites(t *testing.T) {
	w := seq.NewOffsetWriter(shortWriter{})
	if err := w.Write([]byte{1, 2}); !errors.Is(err, io.ErrShortWrite) {
		t.Fatalf("Write: %v", err)
	}
	if err := w.WriteByte(1); !errors.Is(err, io.ErrShortWrite) {
		t.Fatalf("WriteByte: %v", err)
	}
}
