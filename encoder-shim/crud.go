package encoder_shim

import (
	"fmt"
	nbt "neomega-nbt"
)

const (
	EncodeStateBegin = EncodeState(0) // -> writeTag
	EncodeStateEnd   = EncodeState(1) // -> done (if stack = 0)
	EncodeStateTag   = EncodeState(iota)
	EncodeStateValueInt8
	EncodeStateValueInt16
	EncodeStateValueInt32
	EncodeStateValueInt64
	EncodeStateValueFloat32
	EncodeStateValueFloat64
	EncodeStateValueString
	EncodeStateCompound
	EncodeStateNumberArray     = 16
	EncodeStateByteArray       = EncodeStateValueInt8 | EncodeStateNumberArray
	EncodeStateIntArray        = EncodeStateValueInt32 | EncodeStateNumberArray
	EncodeStateLongArray       = EncodeStateValueInt64 | EncodeStateNumberArray
	EncodeStateList            = 32
	EncodeStateListTypeForList = 64
)

func ListElementTypeName(s EncodeState) string {
	switch s {
	case EncodeStateValueInt8:
		return "int8/uint8 value"
	case EncodeStateValueInt16:
		return "int16/uint16 value"
	case EncodeStateValueInt32:
		return "int32/uint32 value"
	case EncodeStateValueInt64:
		return "int64/uint64 value"
	case EncodeStateValueFloat32:
		return "float32/float32 value"
	case EncodeStateValueFloat64:
		return "float64/float64 value"
	case EncodeStateValueString:
		return "string value"
	case EncodeStateByteArray:
		return "byte array"
	case EncodeStateIntArray:
		return "int array"
	case EncodeStateLongArray:
		return "long array"
	case EncodeStateCompound:
		return "compound"
	case EncodeStateList:
		return "list"
	default:
		return "unknown list element type"
	}
}

func StateExpectCommon(EncodeState, listElementType EncodeState, aux uint32) string {
	switch EncodeState {
	case EncodeStateBegin:
		return "data type"
	case EncodeStateCompound:
		return "data type for compound or end"
	case EncodeStateEnd:
		panic("should always pop at this state")
	case EncodeStateTag:
		return "tag string"
	case EncodeStateValueInt8:
		return "int8/uint8 value"
	case EncodeStateValueInt16:
		return "int16/uint16 value"
	case EncodeStateValueInt32:
		return "int32/uint32 value"
	case EncodeStateValueInt64:
		return "int64/uint64 value"
	case EncodeStateValueFloat32:
		return "float32/float32 value"
	case EncodeStateValueFloat64:
		return "float64/float64 value"
	case EncodeStateValueString:
		return "string value"
	case EncodeStateByteArray:
		if aux == 0 {
			return "length(int32) of byte array"
		} else {
			return fmt.Sprintf("uint8/int8 value for byte array(%v elements remaining)", aux)
		}
	case EncodeStateIntArray:
		if aux == 0 {
			return "length(int32) of int array"
		} else {
			return fmt.Sprintf("uint32/int32 value for int array(%v elements remaining)", aux)
		}
	case EncodeStateLongArray:
		if aux == 0 {
			return "length(int64) of int array"
		} else {
			return fmt.Sprintf("uint64/int64 for long array(%v elements remaining)", aux)
		}
	default:
		if EncodeState == EncodeStateList {
			return "type for list or empty mark"
		}
		if EncodeState&EncodeStateList == EncodeStateList {
			if aux == 0 {
				return "length(int64) of list"
			} else {
				return fmt.Sprintf("%v for long array(%v elements remaining)", ListElementTypeName(listElementType), aux)
			}
		}
	}
	return "unknown"
}

func BoxListElementTagType(tagType nbt.TagType) EncodeState {
	switch tagType {
	case nbt.TagInt8:
		return EncodeStateList | EncodeStateValueInt8
	case nbt.TagInt16:
		return EncodeStateList | EncodeStateValueInt16
	case nbt.TagInt32:
		return EncodeStateList | EncodeStateValueInt32
	case nbt.TagInt64:
		return EncodeStateList | EncodeStateValueInt64
	case nbt.TagFloat:
		return EncodeStateList | EncodeStateValueFloat32
	case nbt.TagDouble:
		return EncodeStateList | EncodeStateValueFloat64
	case nbt.TagString:
		return EncodeStateList | EncodeStateValueString
	case nbt.TagByteArray:
		return EncodeStateList | EncodeStateByteArray
	case nbt.TagIntArray:
		return EncodeStateList | EncodeStateIntArray
	case nbt.TagLongArray:
		return EncodeStateList | EncodeStateLongArray
	case nbt.TagCompound:
		return EncodeStateList | EncodeStateCompound
	case nbt.TagList:
		return EncodeStateList | EncodeStateListTypeForList
	default:
		return EncodeStateList
	}
}

func StashTagType(tagType nbt.TagType) uint32 {
	switch tagType {
	case nbt.TagByte:
		return uint32(EncodeStateValueInt8)
	case nbt.TagShort:
		return uint32(EncodeStateValueInt16)
	case nbt.TagInt:
		return uint32(EncodeStateValueInt32)
	case nbt.TagLong:
		return uint32(EncodeStateValueInt64)
	case nbt.TagFloat:
		return uint32(EncodeStateValueFloat32)
	case nbt.TagDouble:
		return uint32(EncodeStateValueFloat64)
	case nbt.TagByteArray:
		return uint32(EncodeStateByteArray)
	case nbt.TagString:
		return uint32(EncodeStateValueString)
	case nbt.TagList:
		return uint32(EncodeStateList)
	case nbt.TagCompound:
		return uint32(EncodeStateCompound)
	case nbt.TagIntArray:
		return uint32(EncodeStateIntArray)
	case nbt.TagLongArray:
		return uint32(EncodeStateLongArray)
	default:
		return 0
	}
}

func UnmapToBasicValueType(tagType EncodeState) nbt.HighLevelValueType {
	switch tagType {
	case (EncodeStateValueInt8):
		return nbt.TrivialByte
	case (EncodeStateValueInt16):
		return nbt.TrivialShort
	case (EncodeStateValueInt32):
		return nbt.TrivialInt
	case (EncodeStateValueInt64):
		return nbt.TrivialLong
	case (EncodeStateValueFloat32):
		return nbt.TrivialFloat
	case (EncodeStateValueFloat64):
		return nbt.TrivialDouble
	case (EncodeStateByteArray):
		return nbt.ComplexByteArray
	case (EncodeStateValueString):
		return nbt.TrivialString
	case (EncodeStateList):
		return nbt.ComplexList
	case (EncodeStateCompound):
		return nbt.ComplexCompound
	case (EncodeStateIntArray):
		return nbt.ComplexIntArray
	case (EncodeStateLongArray):
		return nbt.ComplexLongArray
	default:
		return 0
	}
}

func (fa *EncoderShim[HL]) WriteInt16(v int16) error {
	if err := fa.HighLevelEncoder.WriteInt16(v, fa.ValuePos()); err != nil {
		return err
	}
	if err := fa.WriteNormalValue(EncodeStateValueInt16); err != nil {
		return err
	}
	return nil
}
func (fa *EncoderShim[HL]) WriteUint16(v uint16) error {
	if err := fa.HighLevelEncoder.WriteUint16(v, fa.ValuePos()); err != nil {
		return err
	}
	if err := fa.WriteNormalValue(EncodeStateValueInt16); err != nil {
		return err
	}

	return nil
}
func (fa *EncoderShim[HL]) WriteFloat32(v float32) error {
	if err := fa.HighLevelEncoder.WriteFloat32(v, fa.ValuePos()); err != nil {
		return err
	}
	if err := fa.WriteNormalValue(EncodeStateValueFloat32); err != nil {
		return err
	}
	return nil
}
func (fa *EncoderShim[HL]) WriteFloat64(v float64) error {
	if err := fa.HighLevelEncoder.WriteFloat64(v, fa.ValuePos()); err != nil {
		return err
	}
	if err := fa.WriteNormalValue(EncodeStateValueFloat64); err != nil {
		return err
	}
	return nil
}
func (fa *EncoderShim[HL]) WriteString(v string) error {
	if err := fa.HighLevelEncoder.WriteString(v, fa.ValuePos()); err != nil {
		return err
	}
	if err := fa.WriteNormalValue(EncodeStateValueString); err != nil {
		return err
	}
	return nil
}

func (fa *EncoderShim[HL]) WriteInt8(v int8) error {
	if fa.CanWriteNumberValue(EncodeStateValueInt8) {
		if err := fa.HighLevelEncoder.WriteInt8(v, fa.ValuePos()); err != nil {
			return err
		}
	}
	return fa.WriteNumberValue(EncodeStateValueInt8)
}
func (fa *EncoderShim[HL]) WriteUint8(v uint8) error {
	if fa.CanWriteNumberValue(EncodeStateValueInt8) {
		if err := fa.HighLevelEncoder.WriteUint8(v, fa.ValuePos()); err != nil {
			return err
		}
	}
	return fa.WriteNumberValue(EncodeStateValueInt8)
}
func (fa *EncoderShim[HL]) WriteInt32(v int32) error {
	if fa.CanWriteNumberValue(EncodeStateValueInt32) {
		if err := fa.HighLevelEncoder.WriteInt32(v, fa.ValuePos()); err != nil {
			return err
		}
	}
	return fa.WriteNumberValue(EncodeStateValueInt32)
}
func (fa *EncoderShim[HL]) WriteUint32(v uint32) error {
	if fa.CanWriteNumberValue(EncodeStateValueInt32) {
		if err := fa.HighLevelEncoder.WriteUint32(v, fa.ValuePos()); err != nil {
			return err
		}
	}
	return fa.WriteNumberValue(EncodeStateValueInt32)
}
func (fa *EncoderShim[HL]) WriteInt64(v int64) error {
	if fa.CanWriteNumberValue(EncodeStateValueInt64) {
		if err := fa.HighLevelEncoder.WriteInt64(v, fa.ValuePos()); err != nil {
			return err
		}
	}
	return fa.WriteNumberValue(EncodeStateValueInt64)
}
func (fa *EncoderShim[HL]) WriteUint64(v uint64) error {
	if fa.CanWriteNumberValue(EncodeStateValueInt64) {
		if err := fa.HighLevelEncoder.WriteUint64(v, fa.ValuePos()); err != nil {
			return err
		}
	}
	return fa.WriteNumberValue(EncodeStateValueInt64)
}
