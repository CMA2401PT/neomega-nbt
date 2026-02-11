package encoder_snbt

import (
	"fmt"
	"io"
	nbt "neomega-nbt"
)

type Writer interface {
	io.Writer
	WriteByte(b byte) (err error)
}

type SnbtEncoderHighLevel[W Writer] struct{ w Writer }

func NewSnbtEncoderHighLevel[W Writer](w W) SnbtEncoderHighLevel[W] {
	return SnbtEncoderHighLevel[W]{w: w}
}

func (e SnbtEncoderHighLevel[W]) WriteRootTag(tag string, valueType nbt.HighLevelValueType) error {
	// snbt 都没有 root tag
	return nil
}
func (e SnbtEncoderHighLevel[W]) BeginComplexArray(valueType nbt.HighLevelValueType, len int) error {
	// e.inArray = true
	switch valueType {
	case nbt.TrivialByte:
		return writeString(e.w, "[B; ")
	case nbt.TrivialInt:
		return writeString(e.w, "[I; ")
	case nbt.TrivialLong:
		return writeString(e.w, "[L; ")
	default:
		panic(fmt.Errorf("unknown array value type: %v", valueType))
	}
}
func (e SnbtEncoderHighLevel[W]) NextArrayElement() error {
	if err := e.w.WriteByte(','); err != nil {
		return err
	}
	return e.w.WriteByte(' ')
}
func (e SnbtEncoderHighLevel[W]) EndComplexArray() error {
	// e.inArray = false
	return e.w.WriteByte(']')
}
func (e SnbtEncoderHighLevel[W]) BeginComplexCompound() error {
	return writeString(e.w, "{")
}
func (e SnbtEncoderHighLevel[W]) NextCompoundElement() error {
	if err := e.w.WriteByte(','); err != nil {
		return err
	}
	return e.w.WriteByte(' ')
}
func (e SnbtEncoderHighLevel[W]) WriteCompoundTag(tag string, valueType nbt.HighLevelValueType) error {
	if err := writeSNBTString(e.w, tag); err != nil {
		return err
	}
	if err := e.w.WriteByte(':'); err != nil {
		return err
	}
	if err := e.w.WriteByte(' '); err != nil {
		return err
	}
	return nil
}
func (e SnbtEncoderHighLevel[W]) EndComplexCompound() error {
	return writeString(e.w, "}")
}
func (e SnbtEncoderHighLevel[W]) BeginAndEndComplexEmptyList() error {
	return writeString(e.w, "[]")
}
func (e SnbtEncoderHighLevel[W]) BeginComplexList(valueType nbt.HighLevelValueType, len int, typeErase bool) error {
	return writeString(e.w, "[")
}
func (e SnbtEncoderHighLevel[W]) NextListElement(valueType nbt.HighLevelValueType) error {
	if err := e.w.WriteByte(','); err != nil {
		return err
	}
	return e.w.WriteByte(' ')
}
func (e SnbtEncoderHighLevel[W]) EndComplexList() error {
	return writeString(e.w, "]")
}

func (e SnbtEncoderHighLevel[W]) HasFlag(flag nbt.Flag) bool {
	switch flag {
	case nbt.FLAG_SUPPORT_TYPE_ERASE:
		return true
	case nbt.FLAG_OMIT_LIST_LEN:
		return true
	default:
		return false
	}
}
