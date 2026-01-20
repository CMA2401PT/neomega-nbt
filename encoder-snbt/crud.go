package encoder_snbt

import (
	nbt "neomega-nbt"
	"strconv"
)

func (e SnbtEncoderHighLevel[W]) WriteInt8(v int8, p nbt.ValuePos) error {
	if err := writeInt(e.w, v); err != nil {
		return err
	}
	if p.InArray() {
		return nil
	}
	return e.w.WriteByte('b')
}
func (e SnbtEncoderHighLevel[W]) WriteInt16(v int16, p nbt.ValuePos) error {
	if err := writeInt(e.w, v); err != nil {
		return err
	}
	if p.InArray() {
		return nil
	}
	return e.w.WriteByte('s')
}
func (e SnbtEncoderHighLevel[W]) WriteInt64(v int64, p nbt.ValuePos) error {
	if err := writeInt(e.w, v); err != nil {
		return err
	}
	if p.InArray() {
		return nil
	}
	return e.w.WriteByte('l')
}

// 无论何时，int后都没有类型标注
func (e SnbtEncoderHighLevel[W]) WriteInt32(v int32, p nbt.ValuePos) error {
	return writeInt(e.w, v)
}

// common value
func (e SnbtEncoderHighLevel[W]) WriteFloat32(v float32, p nbt.ValuePos) error {
	if err := writeString(e.w, strconv.FormatFloat(float64(v), 'f', -1, 32)); err != nil {
		return err
	}
	return e.w.WriteByte('f')
}
func (e SnbtEncoderHighLevel[W]) WriteFloat64(v float64, p nbt.ValuePos) error {
	if err := writeString(e.w, strconv.FormatFloat(float64(v), 'f', -1, 64)); err != nil {
		return err
	}
	return e.w.WriteByte('d')
}
func (e SnbtEncoderHighLevel[W]) WriteString(v string, p nbt.ValuePos) error {
	return writeSNBTString(e.w, v)
}
