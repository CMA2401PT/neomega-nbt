package v1_test

import (
	"bytes"
	"reflect"
	"testing"

	decoder_go_object "neomega-nbt/decoder-go-object"
	nbt "neomega-nbt/presets/v1"
)

var encodings = []nbt.Encoding{
	nbt.EncodingLittleEdian, nbt.EncodingBigEdian, nbt.EncodingNetworkLittleEdian,
}

func TestNBTRoundTrip(t *testing.T) {
	cases := []struct {
		name  string
		value any
	}{
		{"int", int32(-12345)}, {"long", int64(-1234567890123)},
		{"float", float32(1.25)}, {"double", float64(-2.5)}, {"string", "你好"},
		{"int list", []int32{1, -2}}, {"string list", []string{"a", "b"}},
		{"empty typed list", []int32{}}, {"empty untyped list", []any{}},
		{"empty compound", map[string]any{}},
		{"nested list", [][]int32{{1, 2}, {}, {3}}},
		{"array list", [][2]int32{{1, 2}, {3, 4}}},
		{"compound list", []map[string]any{{"x": int32(1)}, {}, {"x": int32(2)}}},
		{"byte array", [2]byte{0, 255}}, {"empty array", [0]int32{}},
	}
	for _, encoding := range encodings {
		for _, tc := range cases {
			t.Run(tc.name+"/"+string(rune('0'+encoding)), func(t *testing.T) {
				// A failing case should not hide regressions in other conversion paths.
				defer func() {
					if p := recover(); p != nil {
						t.Errorf("panic: %v", p)
					}
				}()
				opt := nbt.OptDecodeToNBT{Encoding: encoding, DecodeOption: decoder_go_object.DecodeOption{SortMapKey: true}}
				data, err := nbt.EncodeToNBTBytes(tc.value, opt)
				if err != nil {
					t.Fatal(err)
				}
				var buf bytes.Buffer
				if err := nbt.EncodeToNBT(tc.value, &buf, opt); err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(data, buf.Bytes()) {
					t.Fatal("stream and byte encoders disagree")
				}
				for _, stream := range []bool{false, true} {
					out := reflect.New(reflect.TypeOf(tc.value))
					decOpt := nbt.OptDecodeFromNBT{Encoding: encoding}
					if stream {
						err = nbt.DecodeFromNBT(bytes.NewReader(data), out.Interface(), decOpt)
					} else {
						err = nbt.DecodeFromNBTBytes(data, out.Interface(), decOpt)
					}
					if err != nil {
						t.Fatalf("stream=%v: %v", stream, err)
					}
					if !reflect.DeepEqual(out.Elem().Interface(), tc.value) {
						t.Fatalf("stream=%v: got %#v, want %#v", stream, out.Elem().Interface(), tc.value)
					}
				}
			})
		}
	}
}

func TestNBTToAnyStringList(t *testing.T) {
	data, err := nbt.EncodeToNBTBytes([]string{"a", "b"}, nbt.OptDecodeToNBT{Encoding: nbt.EncodingLittleEdian})
	if err != nil {
		t.Fatal(err)
	}
	var out any
	if err := nbt.DecodeFromNBTBytes(data, &out, nbt.OptDecodeFromNBT{Encoding: nbt.EncodingLittleEdian}); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(out, []string{"a", "b"}) {
		t.Fatalf("got %#v", out)
	}
}
