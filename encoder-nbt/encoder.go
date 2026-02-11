package encoder_nbt

import (
	"io"
	nbt "neomega-nbt"
)

var _ nbt.Encoder = NBTEncoder[offsetWriter, LittleEndian[offsetWriter]]{}

type NBTEncoder[W Writer, E Encoding[W]] struct {
	Writer W
}

func NewEncoderFromIOWriter[E Encoding[offsetWriter]](w io.Writer) NBTEncoder[offsetWriter, E] {
	var writer offsetWriter
	if byteWriter, ok := w.(io.ByteWriter); ok {
		writer = offsetWriter{Writer: w, writeByte: byteWriter.WriteByte}
	} else {
		writer = offsetWriter{Writer: w, writeByte: func(b byte) error {
			var buf [1]byte
			buf[0] = b
			_, err := w.Write(buf[:])
			return err
		}}
	}
	return NBTEncoder[offsetWriter, E]{Writer: writer}
}

func NewEncoderToBytes[E Encoding[*SliceGrowWriter]](initBuf []uint8) (
	enc NBTEncoder[*SliceGrowWriter, E],
	getResult func() []uint8,
) {
	writer := NewSliceGrowWriter(initBuf)
	return NBTEncoder[*SliceGrowWriter, E]{Writer: writer}, func() []uint8 {
		return writer.Bytes()
	}
}

func NewEncoder[W Writer, E Encoding[W]](w W) NBTEncoder[W, E] {
	return NBTEncoder[W, E]{Writer: w}
}

func (e NBTEncoder[W, E]) WriteType(t nbt.TagType) error {
	return e.Writer.WriteByte(byte(t))
}

func (e NBTEncoder[W, E]) WriteEmptyListMark() (err error) {
	err = e.Writer.WriteByte(byte(nbt.TagEnd))
	if err != nil {
		return err
	}
	var enc E
	return enc.WriteInt32(e.Writer, 0)
}

func (e NBTEncoder[W, E]) WriteEnd() error {
	return e.Writer.WriteByte(byte(nbt.TagEnd))
}

func (e NBTEncoder[W, E]) WriteTag(tag string) error {
	var enc E
	return enc.WriteString(e.Writer, tag)
}

func (e NBTEncoder[W, E]) WriteLen(len uint32) error {
	var enc E
	return enc.WriteInt32(e.Writer, int32(len))
}

func (e NBTEncoder[W, E]) WriteInt8(v int8) error {
	return e.Writer.WriteByte(byte(v))
}

func (e NBTEncoder[W, E]) WriteInt16(v int16) error {
	var enc E
	return enc.WriteInt16(e.Writer, v)
}

func (e NBTEncoder[W, E]) WriteInt32(v int32) error {
	var enc E
	return enc.WriteInt32(e.Writer, v)
}

func (e NBTEncoder[W, E]) WriteInt64(v int64) error {
	var enc E
	return enc.WriteInt64(e.Writer, v)
}

func (e NBTEncoder[W, E]) WriteFloat32(v float32) error {
	var enc E
	return enc.WriteFloat32(e.Writer, v)
}

func (e NBTEncoder[W, E]) WriteFloat64(v float64) error {
	var enc E
	return enc.WriteFloat64(e.Writer, v)
}

func (e NBTEncoder[W, E]) WriteString(v string) error {
	var enc E
	return enc.WriteString(e.Writer, v)
}
