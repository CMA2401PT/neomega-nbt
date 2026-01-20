package decoder_go_object

import (
	"fmt"
	nbt "neomega-nbt"
	"reflect"
)

func ReflectKindToHighLevelValueType(kind reflect.Kind, vt func() reflect.Type, sliceAsArray bool) nbt.HighLevelValueType {
	switch kind {
	case reflect.Int8, reflect.Uint8, reflect.Bool:
		return nbt.TrivialByte
	case reflect.Int16, reflect.Uint16:
		return nbt.TrivialShort
	case reflect.Int32, reflect.Uint32:
		return nbt.TrivialInt
	case reflect.Int64, reflect.Uint64:
		return nbt.TrivialLong
	case reflect.Float32:
		return nbt.TrivialFloat
	case reflect.Float64:
		return nbt.TrivialDouble
	case reflect.String:
		return nbt.TrivialString
	case reflect.Array:
		switch vt().Elem().Kind() {
		case reflect.Int8, reflect.Uint8, reflect.Bool:
			return nbt.ComplexByteArray
		case reflect.Int32, reflect.Uint32:
			return nbt.ComplexIntArray
		case reflect.Int64, reflect.Uint64:
			return nbt.ComplexLongArray
		default:
			return nbt.Invalid
		}
	case reflect.Slice:
		if !sliceAsArray {
			return nbt.ComplexList
		} else {
			switch vt().Elem().Kind() {
			case reflect.Int8, reflect.Uint8, reflect.Bool:
				return nbt.ComplexByteArray
			case reflect.Int32, reflect.Uint32:
				return nbt.ComplexIntArray
			case reflect.Int64, reflect.Uint64:
				return nbt.ComplexLongArray
			default:
				return nbt.ComplexList
			}
		}
	case reflect.Struct, reflect.Map:
		return nbt.ComplexCompound
	default:
		return nbt.Invalid
	}
}

func WriteTrivialReflectValue[E nbt.WriteTrivialWithPos](kind reflect.Kind, val reflect.Value, reciver E, vp nbt.ValuePos) error {
	switch kind {
	case reflect.Int8:
		return reciver.WriteInt8(int8(val.Int()), vp)
	case reflect.Uint8:
		return reciver.WriteInt8(int8(val.Uint()), vp)
	case reflect.Bool:
		if val.Bool() {
			return reciver.WriteInt8(int8(1), vp)
		}
		return reciver.WriteInt8(int8(0), vp)
	case reflect.Int16:
		return reciver.WriteInt16(int16(val.Int()), vp)
	case reflect.Uint16:
		return reciver.WriteInt16(int16(val.Uint()), vp)
	case reflect.Int32:
		return reciver.WriteInt32(int32(val.Int()), vp)
	case reflect.Uint32:
		return reciver.WriteInt32(int32(val.Uint()), vp)
	case reflect.Int64:
		return reciver.WriteInt64(int64(val.Int()), vp)
	case reflect.Uint64:
		return reciver.WriteInt64(int64(val.Uint()), vp)
	case reflect.Float32:
		return reciver.WriteFloat32(float32(val.Float()), vp)
	case reflect.Float64:
		return reciver.WriteFloat64(float64(val.Float()), vp)
	case reflect.String:
		return reciver.WriteString(val.String(), vp)
	}
	return fmt.Errorf("not a trivial nbt value: %v", val)
}

func GetArrayElementWritor[E nbt.WriteTrivialWithPos](val reflect.Value, reciver E) (
	et nbt.HighLevelValueType,
	writeElem func(int) error,
) {
	switch val.Type().Elem().Kind() {
	case reflect.Int8:
		et = nbt.TrivialByte
		writeElem = func(i int) error {
			return reciver.WriteInt8(int8(val.Index(i).Int()), nbt.ValuePosArray)
		}
	case reflect.Uint8:
		et = nbt.TrivialByte
		writeElem = func(i int) error {
			return reciver.WriteInt8(int8(val.Index(i).Uint()), nbt.ValuePosArray)
		}
	case reflect.Int32:
		et = nbt.TrivialInt
		writeElem = func(i int) error {
			return reciver.WriteInt32(int32(val.Index(i).Int()), nbt.ValuePosArray)
		}
	case reflect.Uint32:
		et = nbt.TrivialInt
		writeElem = func(i int) error {
			return reciver.WriteInt32(int32(val.Index(i).Uint()), nbt.ValuePosArray)
		}
	case reflect.Int64:
		et = nbt.TrivialLong
		writeElem = func(i int) error {
			return reciver.WriteInt64(int64(val.Index(i).Int()), nbt.ValuePosArray)
		}
	case reflect.Uint64:
		et = nbt.TrivialLong
		writeElem = func(i int) error {
			return reciver.WriteInt64(int64(val.Index(i).Uint()), nbt.ValuePosArray)
		}
	}
	return
}
