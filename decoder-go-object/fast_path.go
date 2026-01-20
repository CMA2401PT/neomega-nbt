package decoder_go_object

import (
	nbt "neomega-nbt"
	"reflect"
	"unsafe"
)

func castSliceView[T any](v reflect.Value) []T {
	if v.Kind() != reflect.Slice {
		panic("castSliceView: not slice")
	}
	if v.Len() == 0 {
		return nil
	}
	ptr := unsafe.Pointer(v.Pointer()) // &a[0]
	return unsafe.Slice((*T)(ptr), v.Len())
}
func fastPathNumericSlice[T any, R nbt.EncoderHighLevel](
	v reflect.Value,
	reciver R,
	et nbt.HighLevelValueType,
	write func(R, T) error,
	array bool,
) error {
	list := castSliceView[T](v)
	n := len(list)

	if array {
		if err := reciver.BeginComplexArray(et, n); err != nil {
			return err
		}
	} else {
		if err := reciver.BeginComplexList(et, n, false); err != nil {
			return err
		}
	}

	for i, x := range list {
		if i != 0 {
			if array {
				if err := reciver.NextArrayElement(); err != nil {
					return err
				}
			} else {
				if err := reciver.NextListElement(et); err != nil {
					return err
				}
			}
		}
		if err := write(reciver, x); err != nil {
			return err
		}
	}
	if array {
		if err := reciver.EndComplexArray(); err != nil {
			return err
		}
	} else {
		if err := reciver.EndComplexList(); err != nil {
			return err
		}
	}
	return nil
}

func HighLevelDecodeListToFastPath[R nbt.EncoderHighLevel](
	orig any,
	rv reflect.Value, kind reflect.Kind, reciver R,
	opt DecodeOption,
) (bool, error) {
	pos := nbt.ValuePosList
	if opt.SliceAsArray {
		pos = nbt.ValuePosArray
	}
	if kind != reflect.Slice {
		return false, nil
	}
	elem := rv.Type().Elem()
	if elem.Implements(reflect.TypeOf((*DecoderOverride)(nil)).Elem()) {
		return false, nil
	}

	switch elem.Kind() {
	default:
		return false, nil
	case reflect.Int8:
		return true, fastPathNumericSlice[int8](
			rv, reciver, nbt.TrivialByte,
			func(r R, x int8) error {
				return r.WriteInt8(x, pos)
			}, opt.SliceAsArray)

	case reflect.Uint8:
		return true, fastPathNumericSlice[uint8](
			rv, reciver, nbt.TrivialByte,
			func(r R, x uint8) error {
				return r.WriteInt8(int8(x), pos)
			}, opt.SliceAsArray)

	case reflect.Int16:
		return true, fastPathNumericSlice[int16](
			rv, reciver, nbt.TrivialShort,
			func(r R, x int16) error {
				return r.WriteInt16(x, nbt.ValuePosList)
			}, false)

	case reflect.Uint16:
		return true, fastPathNumericSlice[uint16](
			rv, reciver, nbt.TrivialShort,
			func(r R, x uint16) error {
				return r.WriteInt16(int16(x), nbt.ValuePosList)
			}, false)

	case reflect.Int32:
		return true, fastPathNumericSlice[int32](
			rv, reciver, nbt.TrivialInt,
			func(r R, x int32) error {
				return r.WriteInt32(x, pos)
			}, opt.SliceAsArray)

	case reflect.Uint32:
		return true, fastPathNumericSlice[uint32](
			rv, reciver, nbt.TrivialInt,
			func(r R, x uint32) error {
				return r.WriteInt32(int32(x), pos)
			}, opt.SliceAsArray)

	case reflect.Int64:
		return true, fastPathNumericSlice[int64](
			rv, reciver, nbt.TrivialLong,
			func(r R, x int64) error {
				return r.WriteInt64(x, pos)
			}, opt.SliceAsArray)

	case reflect.Uint64:
		return true, fastPathNumericSlice[uint64](
			rv, reciver, nbt.TrivialLong,
			func(r R, x uint64) error {
				return r.WriteInt64(int64(x), pos)
			}, opt.SliceAsArray)

	case reflect.Float32:
		return true, fastPathNumericSlice[float32](
			rv, reciver, nbt.TrivialFloat,
			func(r R, x float32) error {
				return r.WriteFloat32(x, nbt.ValuePosList)
			}, false)

	case reflect.Float64:
		return true, fastPathNumericSlice[float64](
			rv, reciver, nbt.TrivialDouble,
			func(r R, x float64) error {
				return r.WriteFloat64(x, nbt.ValuePosList)
			}, false)

	case reflect.String:
		list := castSliceView[string](rv)
		n := len(list)
		if err := reciver.BeginComplexList(nbt.TrivialString, n, false); err != nil {
			return true, err
		}
		for i, s := range list {
			if i != 0 {
				if err := reciver.NextListElement(nbt.TrivialString); err != nil {
					return true, err
				}
			}
			if err := reciver.WriteString(s, nbt.ValuePosList); err != nil {
				return true, err
			}
		}
		return true, reciver.EndComplexList()
	}
}
