package encoder_go_object

import (
	nbt "neomega-nbt"
	"reflect"
	"unsafe"
)

func makePtrReciver[T any](ptr unsafe.Pointer, cast func(int64, any) T) TrivialValReciver {
	p := (*T)(ptr)
	return func(v any, i64 int64) error {
		*p = cast(i64, v)
		return nil
	}
}

func GetTrivialReciverFastPath(ptrToData any, valueType nbt.HighLevelValueType, elem reflect.Value, ptr unsafe.Pointer) (bool, TrivialValReciver, error) {
	switch valueType {
	case nbt.TrivialByte:
		switch elem.Kind() {
		case reflect.Int8:
			return true, makePtrReciver[int8](ptr,
				func(i64 int64, _ any) int8 { return int8(i64) }), nil
		case reflect.Uint8:
			return true, makePtrReciver[uint8](ptr,
				func(i64 int64, _ any) uint8 { return uint8(i64) }), nil
		}

	case nbt.TrivialShort:
		switch elem.Kind() {
		case reflect.Int16:
			return true, makePtrReciver[int16](ptr,
				func(i64 int64, _ any) int16 { return int16(i64) }), nil
		case reflect.Uint16:
			return true, makePtrReciver[uint16](ptr,
				func(i64 int64, _ any) uint16 { return uint16(i64) }), nil
		}

	case nbt.TrivialInt:
		switch elem.Kind() {
		case reflect.Int32, reflect.Int:
			return true, makePtrReciver[int32](ptr,
				func(i64 int64, _ any) int32 { return int32(i64) }), nil
		case reflect.Uint32, reflect.Uint:
			return true, makePtrReciver[uint32](ptr,
				func(i64 int64, _ any) uint32 { return uint32(i64) }), nil
		}

	case nbt.TrivialLong:
		switch elem.Kind() {
		case reflect.Int64:
			return true, makePtrReciver[int64](ptr,
				func(i64 int64, _ any) int64 { return i64 }), nil
		case reflect.Uint64:
			return true, makePtrReciver[uint64](ptr,
				func(i64 int64, _ any) uint64 { return uint64(i64) }), nil
		}

	case nbt.TrivialFloat:
		if elem.Kind() == reflect.Float32 {
			return true, makePtrReciver[float32](ptr,
				func(_ int64, v any) float32 { return v.(float32) }), nil
		}

	case nbt.TrivialDouble:
		if elem.Kind() == reflect.Float64 {
			return true, makePtrReciver[float64](ptr,
				func(_ int64, v any) float64 { return v.(float64) }), nil
		}

	case nbt.TrivialString:
		if elem.Kind() == reflect.String {
			return true, makePtrReciver[string](ptr,
				func(_ int64, v any) string { return v.(string) }), nil
		}
	}
	return false, nil, nil
}
