package encoder_nbt

import (
	"io"
	nbt "neomega-nbt"
	seq_io "neomega-nbt/seq-io"
)

var _ nbt.Encoder = NBTEncoder[seq_io.OffsetWriter, LittleEndian[seq_io.OffsetWriter]]{}

type NBTEncoder[W seq_io.Writer, E Encoding[W]] struct {
	w W
}

func NewEncoderFromIOWriter[E Encoding[seq_io.OffsetWriter]](w io.Writer) NBTEncoder[seq_io.OffsetWriter, E] {
	return NBTEncoder[seq_io.OffsetWriter, E]{w: seq_io.NewOffsetWriter(w)}
}

func NewEncoderToBytes[E Encoding[*seq_io.SliceGrowWriter]](initBuf []uint8) (
	enc NBTEncoder[*seq_io.SliceGrowWriter, E],
	getResult func() []uint8,
) {
	writer := seq_io.NewSliceGrowWriter(initBuf)
	return NBTEncoder[*seq_io.SliceGrowWriter, E]{w: writer}, func() []uint8 {
		return writer.Bytes()
	}
}

func NewEncoder[W seq_io.Writer, E Encoding[W]](w W) NBTEncoder[W, E] {
	return NBTEncoder[W, E]{w: w}
}

func (e NBTEncoder[W, E]) WriteType(t nbt.TagType) error {
	return e.w.WriteByte(byte(t))
}

func (e NBTEncoder[W, E]) WriteEmptyListMark() (err error) {
	err = e.w.WriteByte(byte(nbt.TagEnd))
	if err != nil {
		return err
	}
	var enc E
	return enc.WriteInt32(e.w, 0)
}

func (e NBTEncoder[W, E]) WriteEnd() error {
	return e.w.WriteByte(byte(nbt.TagEnd))
}

func (e NBTEncoder[W, E]) WriteTag(tag string) error {
	var enc E
	return enc.WriteString(e.w, tag)
}

func (e NBTEncoder[W, E]) WriteLen(len uint32) error {
	var enc E
	return enc.WriteInt32(e.w, int32(len))
}

func (e NBTEncoder[W, E]) WriteInt8(v int8) error {
	return e.w.WriteByte(byte(v))
}

func (e NBTEncoder[W, E]) WriteInt16(v int16) error {
	var enc E
	return enc.WriteInt16(e.w, v)
}

func (e NBTEncoder[W, E]) WriteInt32(v int32) error {
	var enc E
	return enc.WriteInt32(e.w, v)
}

func (e NBTEncoder[W, E]) WriteInt64(v int64) error {
	var enc E
	return enc.WriteInt64(e.w, v)
}

func (e NBTEncoder[W, E]) WriteFloat32(v float32) error {
	var enc E
	return enc.WriteFloat32(e.w, v)
}

func (e NBTEncoder[W, E]) WriteFloat64(v float64) error {
	var enc E
	return enc.WriteFloat64(e.w, v)
}

func (e NBTEncoder[W, E]) WriteString(v string) error {
	var enc E
	return enc.WriteString(e.w, v)
}
