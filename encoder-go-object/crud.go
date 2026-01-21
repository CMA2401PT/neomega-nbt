package encoder_go_object

import (
	"fmt"
	nbt "neomega-nbt"
	"reflect"
	"unsafe"
)

func (e *GoObjEncoder) WriteInt8(v int8, vp nbt.ValuePos) error {
	if e.Current.TrivialValReciver == nil {
		return fmt.Errorf("%T cannot accept byte value", e.Current.PtrToData)
	}
	return e.Current.TrivialValReciver(v, int64(v))
}

func (e *GoObjEncoder) WriteInt16(v int16, vp nbt.ValuePos) error {
	if e.Current.TrivialValReciver == nil {
		return fmt.Errorf("%T cannot accept short (int16) value", e.Current.PtrToData)
	}
	return e.Current.TrivialValReciver(v, int64(v))
}

func (e *GoObjEncoder) WriteInt32(v int32, vp nbt.ValuePos) error {
	if e.Current.TrivialValReciver == nil {
		return fmt.Errorf("%T cannot accept int (int32) value", e.Current.PtrToData)
	}
	return e.Current.TrivialValReciver(v, int64(v))
}

func (e *GoObjEncoder) WriteInt64(v int64, vp nbt.ValuePos) error {
	if e.Current.TrivialValReciver == nil {
		return fmt.Errorf("%T cannot accept int (int64) value", e.Current.PtrToData)
	}
	return e.Current.TrivialValReciver(v, int64(v))
}

func (e *GoObjEncoder) WriteFloat32(v float32, vp nbt.ValuePos) error {
	cast := int64(v)
	if e.Current.TrivialValReciver == nil {
		return fmt.Errorf("%T cannot accept float (float32) value", e.Current.PtrToData)
	}
	return e.Current.TrivialValReciver(v, cast)
}

func (e *GoObjEncoder) WriteFloat64(v float64, vp nbt.ValuePos) error {
	cast := int64(v)
	if e.Current.TrivialValReciver == nil {
		return fmt.Errorf("%T cannot accept float (float64) value", e.Current.PtrToData)
	}
	return e.Current.TrivialValReciver(v, cast)
}

func (e *GoObjEncoder) WriteString(v string, vp nbt.ValuePos) error {
	if e.Current.TrivialValReciver == nil {
		return fmt.Errorf("%T cannot accept string (float64) value", e.Current.PtrToData)
	}
	return e.Current.TrivialValReciver(v, 0)
}

func GetTrivialReciver(ptrToData any, valueType nbt.HighLevelValueType, val reflect.Value, ptr unsafe.Pointer) (TrivialValReciver, error) {
	ok, rv, err := GetTrivialReciverFastPath(ptrToData, valueType, val, ptr)
	if ok {
		return rv, err
	}
	switch valueType {
	case nbt.TrivialByte, nbt.TrivialShort, nbt.TrivialInt, nbt.TrivialLong:
		switch vk := val.Kind(); vk {
		default:
			return nil, fmt.Errorf("cannot parse %v as %T", valueType, ptrToData)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return func(v any, i64 int64) error {
				val.SetInt(i64)
				return nil
			}, nil

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return func(v any, i64 int64) error {
				val.SetUint(uint64(i64))
				return nil
			}, nil
		}
	case nbt.TrivialFloat:
		switch vk := val.Kind(); vk {
		default:
			return nil, fmt.Errorf("cannot parse %v as %T", valueType, ptrToData)
		case reflect.Float32, reflect.Float64:
			return func(v any, i64 int64) error {
				val.SetFloat(float64(v.(float32)))
				return nil
			}, nil
		}
	case nbt.TrivialDouble:
		switch vk := val.Kind(); vk {
		default:
			return nil, fmt.Errorf("cannot parse %v as %T", valueType, ptrToData)
		case reflect.Float32, reflect.Float64:
			return func(v any, i64 int64) error {
				val.SetFloat(v.(float64))
				return nil
			}, nil
		}
	case nbt.TrivialString:
		if val.Kind() != reflect.String {
			return nil, fmt.Errorf("cannot parse %v as %T", valueType, ptrToData)
		}
		return func(v any, i64 int64) error {
			val.SetString(v.(string))
			return nil
		}, nil
	default:
		return nil, fmt.Errorf("%T cannot receive nbt data type %v", ptrToData, valueType)
	}
}
