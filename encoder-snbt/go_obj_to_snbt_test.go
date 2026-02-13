package encoder_snbt

import (
	nbt "neomega-nbt"
	decoder_go_object "neomega-nbt/decoder-go-object"
	"testing"
)

func EncodeSNBT(v any, sliceAsArray bool) (string, error) {
	enc, getter := NewEncoderToBytes(nil)
	if err := decoder_go_object.HighLevelDecodeTo(
		v,
		enc,
		decoder_go_object.DecodeOption{
			SliceAsArray: sliceAsArray,
			SortMapKey:   true,
		}); err != nil {
		return "", err
	}
	return string(getter()), nil
}

func TestSnbtEncode(t *testing.T) {
	assertVal := func(v any, out string) {
		t.Helper()
		get, err := EncodeSNBT(v, true)
		if err != nil {
			t.Fatalf("when encode %v, expect: %v, get err: %v", v, out, err)
		} else if string(get) != out {
			t.Fatalf("when encode %v, expect: %v, get: %v", v, out, string(get))
		}
	}
	assertValStrict := func(v any, out string) {
		t.Helper()
		get, err := EncodeSNBT(v, false)
		if err != nil {
			t.Fatalf("when encode %v, expect: %v, get err: %v", v, out, err)
		} else if string(get) != out {
			t.Fatalf("when encode %v, expect: %v, get: %v", v, out, string(get))
		}
	}
	assertVal(`abc\abc"`, `"abc\\abc\""`)
	assertVal(int32(0), `0`)
	assertVal(int32(3), `3`)
	assertVal(int32(-3), `-3`)
	assertVal(int32(2401), `2401`)
	assertVal(int32(-2401), `-2401`)

	assertVal(3, `3`)

	assertVal(int8(0), `0b`)

	assertVal(int8(3), `3b`)
	assertVal(int8(-3), `-3b`)
	assertVal(int8(123), `123b`)
	assertVal(int8(-123), `-123b`)

	assertVal(int16(0), `0s`)
	assertVal(int16(3), `3s`)
	assertVal(int16(-3), `-3s`)
	assertVal(int16(2401), `2401s`)
	assertVal(int16(-2401), `-2401s`)

	assertVal(int64(0), `0l`)
	assertVal(int64(3), `3l`)
	assertVal(int64(-3), `-3l`)
	assertVal(int64(2401), `2401l`)
	assertVal(int64(-2401), `-2401l`)

	assertVal(float32(123.4567e20), `12345670000000000000000f`)
	assertVal(float32(123.4567e-20), `0.000000000000000001234567f`)

	assertVal(float64(123.4567e20), `12345670000000000000000d`)
	assertVal(float64(123.4567e-20), `0.000000000000000001234567d`)

	assertVal([]any{}, `[]`)

	assertVal([]int32{-123, 456, 7, -8, 0, 9}, `[I; -123, 456, 7, -8, 0, 9]`)
	assertVal([]int64{-123, 456, 7, -8, 0, 9}, `[L; -123, 456, 7, -8, 0, 9]`)
	assertVal([]int8{-123, 45, 7, -8, 0, 9}, `[B; -123, 45, 7, -8, 0, 9]`)

	assertValStrict([]int32{-123, 456, 7, -8, 0, 9}, `[-123, 456, 7, -8, 0, 9]`)
	assertValStrict([]int64{-123, 456, 7, -8, 0, 9}, `[-123l, 456l, 7l, -8l, 0l, 9l]`)
	assertValStrict([]int8{-123, 45, 7, -8, 0, 9}, `[-123b, 45b, 7b, -8b, 0b, 9b]`)

	assertVal([]any{int32(123), int8(-4), "abc\\\"", []any{int8(1), int32(-2), "bc"}}, `[123, -4b, "abc\\\"", [1b, -2, "bc"]]`)
	assertValStrict([]any{int32(123), int8(-4), "abc\\\"", []any{int8(1), int32(-2), "bc"}}, `[123, -4b, "abc\\\"", [1b, -2, "bc"]]`)

	assertVal(
		map[string]any{
			"abc":     int32(123),
			"测试":      int8(-4),
			"abc\\\"": "abc\\\"",
			"@{}<>!-=()[]*&^%$#/+~.\";": []any{
				int8(1),
				int32(-2),
				"bc",
				map[string]any{
					"abc": int32(123)},
			},
		},
		`{"@{}<>!-=()[]*&^%$#/+~.\";": [1b, -2, "bc", {"abc": 123}], "abc": 123, "abc\\\"": "abc\\\"", "测试": -4b}`,
	)
	assertValStrict(
		map[string]any{
			"abc":     int32(123),
			"测试":      int8(-4),
			"abc\\\"": "abc\\\"",
			"@{}<>!-=()[]*&^%$#/+~.\";": []any{
				int8(1),
				int32(-2),
				"bc",
				map[string]any{
					"abc": int32(123)},
			},
		},
		`{"@{}<>!-=()[]*&^%$#/+~.\";": [1b, -2, "bc", {"abc": 123}], "abc": 123, "abc\\\"": "abc\\\"", "测试": -4b}`,
	)
}

func TestSnbtArrEncode(t *testing.T) {
	assertVal := func(v any, out string) {
		get, err := EncodeSNBT(v, true)
		if err != nil {
			t.Fatalf("when encode %v, expect: %v, get err: %v", v, out, err)
		} else if string(get) != out {
			t.Fatalf("when encode %v, expect: %v, get: %v", v, out, string(get))
		}
	}
	assertValStrict := func(v any, out string) {
		t.Helper()
		get, err := EncodeSNBT(v, false)
		if err != nil {
			t.Fatalf("when encode %v, expect: %v, get err: %v", v, out, err)
		} else if string(get) != out {
			t.Fatalf("when encode %v, expect: %v, get: %v", v, out, string(get))
		}
	}
	assertVal([]int8{1, 2, 3}, `[B; 1, 2, 3]`)
	assertVal([]int64{1, 2, 3}, `[L; 1, 2, 3]`)
	assertValStrict([]int8{1, 2, 3}, `[1b, 2b, 3b]`)
	assertValStrict([]int64{1, 2, 3}, `[1l, 2l, 3l]`)
}

type MyString string
type MyTypeInt8 int8
type MyTypeInt8Override int8

func (MyTypeInt8Override) GetNBType() nbt.HighLevelValueType {
	return nbt.TrivialByte
}

func (i MyTypeInt8Override) DecodeTo(valuePos nbt.ValuePos, encoder nbt.EncoderHighLevel, getChildDecoder decoder_go_object.GetDecoderFunc) error {
	return encoder.WriteInt8(int8(i)+10, valuePos)
}

type MyTypeInt16 int16
type MyTypeInt32 int32
type MyTypeInt64 int64
type MyTypeFloat32 float32
type MyTypeFloat64 float64
type MyBool bool

type MyArrInt32 [3]MyTypeInt32
type MyArrInt8 [3]MyTypeInt8
type MyArrInt64 [3]MyTypeInt64

type MySliceInt32 []MyTypeInt32
type MySliceInt8 []MyTypeInt8
type MySliceInt64 []MyTypeInt64

type MyAny any
type MySliceAny []MyAny
type MyMapAny map[string]MyAny
type MyStructSub struct {
	SA MyString `nbt:"sa,omitempty"`
}

type MyStruct struct {
	SB MyTypeInt32 `nbt:"sb,omitempty"`
	MyStructSub
}

func TestSnbtEncodeWithCast(t *testing.T) {
	assertVal := func(v any, out string) {
		t.Helper()
		get, err := EncodeSNBT(v, true)
		if err != nil {
			t.Fatalf("when encode %v, expect: %v, get err: %v", v, out, err)
		} else if string(get) != out {
			t.Fatalf("when encode %v, expect: %v, get: %v", v, out, string(get))
		}
	}
	assertValStrict := func(v any, out string) {
		t.Helper()
		get, err := EncodeSNBT(v, false)
		if err != nil {
			t.Fatalf("when encode %v, expect: %v, get err: %v", v, out, err)
		} else if string(get) != out {
			t.Fatalf("when encode %v, expect: %v, get: %v", v, out, string(get))
		}
	}
	assertVal(MyString(`abc\abc"`), `"abc\\abc\""`)
	assertVal(MyTypeInt32(0), `0`)
	assertVal(MyTypeInt32(3), `3`)
	assertVal(MyTypeInt32(-3), `-3`)
	assertVal(MyTypeInt32(2401), `2401`)
	assertVal(MyTypeInt32(-2401), `-2401`)

	assertVal(MyTypeInt8Override(1), `11b`)
	assertValStrict([]MyTypeInt8Override{1, 2, 3}, `[11b, 12b, 13b]`)
	assertVal([]MyTypeInt8Override{1, 2, 3}, `[B; 11, 12, 13]`)
	assertVal(MyTypeInt8(0), `0b`)

	assertVal(MyTypeInt8(3), `3b`)
	assertVal(MyTypeInt8(-3), `-3b`)
	assertVal(MyTypeInt8(123), `123b`)
	assertVal(MyTypeInt8(-123), `-123b`)

	assertVal(MyTypeInt16(0), `0s`)
	assertVal(MyTypeInt16(3), `3s`)
	assertVal(MyTypeInt16(-3), `-3s`)
	assertVal(MyTypeInt16(2401), `2401s`)
	assertVal(MyTypeInt16(-2401), `-2401s`)

	assertVal(MyTypeInt64(0), `0l`)
	assertVal(MyTypeInt64(3), `3l`)
	assertVal(MyTypeInt64(-3), `-3l`)
	assertVal(MyTypeInt64(2401), `2401l`)
	assertVal(MyTypeInt64(-2401), `-2401l`)

	assertVal(MyTypeFloat32(123.4567e20), `12345670000000000000000f`)
	assertVal(MyTypeFloat32(123.4567e-20), `0.000000000000000001234567f`)

	assertVal(MyTypeFloat64(123.4567e20), `12345670000000000000000d`)
	assertVal(MyTypeFloat64(123.4567e-20), `0.000000000000000001234567d`)

	assertVal(MyBool(true), "1b")
	assertVal(MyBool(false), "0b")

	assertVal(MyArrInt32{-123, 456, 7}, `[I; -123, 456, 7]`)
	assertVal(MyArrInt64{-123, 456, 7}, `[L; -123, 456, 7]`)
	assertVal(MyArrInt8{-123, 45, 7}, `[B; -123, 45, 7]`)
	assertValStrict(MyArrInt32{-123, 456, 7}, `[I; -123, 456, 7]`)
	assertValStrict(MyArrInt64{-123, 456, 7}, `[L; -123, 456, 7]`)
	assertValStrict(MyArrInt8{-123, 45, 7}, `[B; -123, 45, 7]`)

	assertVal(MySliceInt32{-123, 456, 7, -8, 0, 9}, `[I; -123, 456, 7, -8, 0, 9]`)
	assertVal(MySliceInt64{-123, 456, 7, -8, 0, 9}, `[L; -123, 456, 7, -8, 0, 9]`)
	assertVal(MySliceInt8{-123, 45, 7, -8, 0, 9}, `[B; -123, 45, 7, -8, 0, 9]`)
	assertValStrict(MySliceInt32{-123, 456, 7, -8, 0, 9}, `[-123, 456, 7, -8, 0, 9]`)
	assertValStrict(MySliceInt64{-123, 456, 7, -8, 0, 9}, `[-123l, 456l, 7l, -8l, 0l, 9l]`)
	assertValStrict(MySliceInt8{-123, 45, 7, -8, 0, 9}, `[-123b, 45b, 7b, -8b, 0b, 9b]`)

	assertVal(MySliceAny{int32(123), int8(-4), "abc\\\"", []any{int8(1), int32(-2), "bc"}}, `[123, -4b, "abc\\\"", [1b, -2, "bc"]]`)

	assertVal(
		MyMapAny{
			"abc":     int32(123),
			"测试":      int8(-4),
			"abc\\\"": "abc\\\"",
			"@{}<>!-=()[]*&^%$#/+~.\";": []any{
				int8(1),
				int32(-2),
				"bc",
				map[string]any{
					"abc": int32(123)},
			},
		},
		`{"@{}<>!-=()[]*&^%$#/+~.\";": [1b, -2, "bc", {"abc": 123}], "abc": 123, "abc\\\"": "abc\\\"", "测试": -4b}`,
	)
	assertValStrict(
		MyMapAny{
			"abc":     int32(123),
			"测试":      int8(-4),
			"abc\\\"": "abc\\\"",
			"@{}<>!-=()[]*&^%$#/+~.\";": []any{
				int8(1),
				int32(-2),
				"bc",
				map[string]any{
					"abc": int32(123)},
			},
		},
		`{"@{}<>!-=()[]*&^%$#/+~.\";": [1b, -2, "bc", {"abc": 123}], "abc": 123, "abc\\\"": "abc\\\"", "测试": -4b}`,
	)
	assertVal(
		MyStruct{
			MyStructSub: MyStructSub{
				SA: "abc",
			},
			SB: MyTypeInt32(34),
		}, `{"sb": 34, "sa": "abc"}`,
	)
	assertValStrict(
		MyStruct{
			MyStructSub: MyStructSub{
				SA: "abc",
			},
			SB: MyTypeInt32(34),
		}, `{"sb": 34, "sa": "abc"}`,
	)
}
