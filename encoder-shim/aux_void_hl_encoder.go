package encoder_shim

import nbt "neomega-nbt"

type VoidEncoderHighLevel struct{}

func (VoidEncoderHighLevel) WriteRootTag(tag string, valueType nbt.HighLevelValueType) error {
	return nil
}
func (VoidEncoderHighLevel) BeginComplexArray(valueType nbt.HighLevelValueType, len int) error {
	return nil
}

func (VoidEncoderHighLevel) EndComplexArray() error      { return nil }
func (VoidEncoderHighLevel) NextArrayElement() error     { return nil }
func (VoidEncoderHighLevel) BeginComplexCompound() error { return nil }

func (VoidEncoderHighLevel) WriteCompoundTag(tag string, valueType nbt.HighLevelValueType) error {
	return nil
}
func (VoidEncoderHighLevel) NextCompoundElement() error { return nil }
func (VoidEncoderHighLevel) EndComplexCompound() error  { return nil }

func (VoidEncoderHighLevel) BeginAndEndComplexEmptyList() error { return nil }
func (VoidEncoderHighLevel) BeginComplexList(valueType nbt.HighLevelValueType, len int, typeErase bool) error {
	return nil
}
func (VoidEncoderHighLevel) NextListElement(nbt.HighLevelValueType) error { return nil }
func (VoidEncoderHighLevel) EndComplexList() error                        { return nil }

func (VoidEncoderHighLevel) WriteInt8(int8, nbt.ValuePos) error       { return nil }
func (VoidEncoderHighLevel) WriteInt16(int16, nbt.ValuePos) error     { return nil }
func (VoidEncoderHighLevel) WriteInt32(int32, nbt.ValuePos) error     { return nil }
func (VoidEncoderHighLevel) WriteInt64(int64, nbt.ValuePos) error     { return nil }
func (VoidEncoderHighLevel) WriteFloat32(float32, nbt.ValuePos) error { return nil }
func (VoidEncoderHighLevel) WriteFloat64(float64, nbt.ValuePos) error { return nil }
func (VoidEncoderHighLevel) WriteString(string, nbt.ValuePos) error   { return nil }
func (VoidEncoderHighLevel) HasFlag(flag nbt.Flag) bool               { return false }
