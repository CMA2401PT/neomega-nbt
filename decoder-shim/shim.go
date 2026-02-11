package decoder_shim

import (
	"errors"
	nbt "neomega-nbt"
)

type DecoderShim[LL nbt.Encoder] struct {
	LowLevelEncoder LL
}

func NewDecoderShim[LL nbt.Encoder](lowlevelEncoder LL) DecoderShim[LL] {
	return DecoderShim[LL]{LowLevelEncoder: lowlevelEncoder}
}

func (s DecoderShim[LL]) WriteRootTag(tag string, valueType nbt.HighLevelValueType) (err error) {
	err = s.LowLevelEncoder.WriteType(valueType.LowLevelType())
	if err != nil {
		return err
	}
	return s.LowLevelEncoder.WriteTag(tag)
}

func (s DecoderShim[LL]) WriteCompoundTag(tag string, valueType nbt.HighLevelValueType) (err error) {
	err = s.LowLevelEncoder.WriteType(valueType.LowLevelType())
	if err != nil {
		return err
	}
	return s.LowLevelEncoder.WriteTag(tag)
}

func (s DecoderShim[LL]) BeginComplexArray(valueType nbt.HighLevelValueType, len int) (err error) {
	return s.LowLevelEncoder.WriteLen(uint32(len))
}

func (s DecoderShim[LL]) EndComplexCompound() error {
	return s.LowLevelEncoder.WriteEnd()
}

func (s DecoderShim[LL]) BeginAndEndComplexEmptyList() error {
	return s.LowLevelEncoder.WriteEmptyListMark()
}

func (s DecoderShim[LL]) BeginComplexList(valueType nbt.HighLevelValueType, len int, typeErase bool) (err error) {
	if typeErase {
		return errors.New("type erase is not supported")
	}
	err = s.LowLevelEncoder.WriteType(valueType.LowLevelType())
	if err != nil {
		return err
	}
	return s.LowLevelEncoder.WriteLen(uint32(len))
}

func (s DecoderShim[LL]) NextArrayElement() error                                { return nil }
func (s DecoderShim[LL]) EndComplexArray() error                                 { return nil }
func (s DecoderShim[LL]) BeginComplexCompound() error                            { return nil }
func (s DecoderShim[LL]) NextCompoundElement() error                             { return nil }
func (s DecoderShim[LL]) NextListElement(valueType nbt.HighLevelValueType) error { return nil }
func (s DecoderShim[LL]) EndComplexList() error                                  { return nil }
func (s DecoderShim[LL]) HasFlag(flag nbt.Flag) bool                             { return false }

func (s DecoderShim[LL]) WriteInt8(i int8, _ nbt.ValuePos) error {
	return s.LowLevelEncoder.WriteInt8(i)
}
func (s DecoderShim[LL]) WriteInt16(i int16, _ nbt.ValuePos) error {
	return s.LowLevelEncoder.WriteInt16(i)
}
func (s DecoderShim[LL]) WriteInt32(i int32, _ nbt.ValuePos) error {
	return s.LowLevelEncoder.WriteInt32(i)
}
func (s DecoderShim[LL]) WriteInt64(i int64, _ nbt.ValuePos) error {
	return s.LowLevelEncoder.WriteInt64(i)
}
func (s DecoderShim[LL]) WriteFloat32(i float32, _ nbt.ValuePos) error {
	return s.LowLevelEncoder.WriteFloat32(i)
}
func (s DecoderShim[LL]) WriteFloat64(i float64, _ nbt.ValuePos) error {
	return s.LowLevelEncoder.WriteFloat64(i)
}
func (s DecoderShim[LL]) WriteString(i string, _ nbt.ValuePos) error {
	return s.LowLevelEncoder.WriteString(i)
}
