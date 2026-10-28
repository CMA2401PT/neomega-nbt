package v1_test

import (
	"bytes"
	"errors"
	"io"
	"math"
	"reflect"
	"strings"
	"testing"

	decoder_go_object "neomega-nbt/decoder-go-object"
	encoder_go_object "neomega-nbt/encoder-go-object"
	nbt "neomega-nbt/presets/v1"
)

func TestConversionPaths(t *testing.T) {
	input := map[string]any{
		"byte": int8(-128), "short": int16(-32768), "int": int32(math.MinInt32),
		"long": int64(math.MinInt64), "float": float32(1.25), "double": float64(2.5),
		"string": "中文\"\\", "list": []int32{1, -2}, "nested": [][]int32{{1}, {3}},
		"bytes": [2]byte{0, 255}, "ints": [2]int32{math.MinInt32, math.MaxInt32},
		"longs": [2]int64{math.MinInt64, math.MaxInt64}, "emptyArray": [0]int32{},
		"emptyMap": map[string]any{},
	}
	text, err := nbt.EncodeToSNBT(input, nbt.OptEncodeToSNBT{SortMapKey: true})
	if err != nil {
		t.Fatal(err)
	}
	var goValue any
	if err := nbt.DecodeFromSNBT(text, &goValue, nbt.OptDecodeFromSNBT{}); err != nil {
		t.Fatal(err)
	}
	again, err := nbt.EncodeToSNBT(goValue, nbt.OptEncodeToSNBT{SortMapKey: true})
	if err != nil {
		t.Fatal(err)
	}
	if again != text {
		t.Fatalf("Go/SNBT round trip:\n%s\n%s", text, again)
	}
	for _, encoding := range encodings {
		opt := nbt.OptEncodeToNBT{Encoding: encoding, DecodeOption: decoder_go_object.DecodeOption{SortMapKey: true, RootTag: "root"}}
		data, err := nbt.EncodeToNBTBytes(input, opt)
		if err != nil {
			t.Fatal(err)
		}
		converted, err := nbt.SNBTToNBT(text, nbt.OptSNBTToNBT{Encoding: encoding, RootTag: "root"})
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(converted, data) {
			t.Fatalf("SNBT/NBT bytes differ for encoding %d", encoding)
		}
		restored, err := nbt.NBTToSNBT(data, encoding)
		if err != nil {
			t.Fatal(err)
		}
		if restored != text {
			t.Fatalf("encoding %d: got %s, want %s", encoding, restored, text)
		}
		for _, target := range encodings {
			converted, err := nbt.ConvertNBT(data, encoding, target)
			if err != nil {
				t.Fatal(err)
			}
			back, err := nbt.ConvertNBT(converted, target, encoding)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(back, data) {
				t.Fatal("binary conversion changed root name, type, order or values")
			}
		}
	}
}

func TestEmptyAndTypedContainers(t *testing.T) {
	cases := []struct {
		text string
		want any
	}{
		{"{}", map[string]any{}}, {"[]", []any{}},
		{"[I;]", [0]int32{}}, {"[B;]", [0]byte{}}, {"[L;]", [0]int64{}},
		{"[I;1,-2]", [2]int32{1, -2}}, {"[B;0,255]", [2]byte{0, 255}},
		{"[L;9223372036854775807]", [1]int64{math.MaxInt64}},
	}
	for _, tc := range cases {
		t.Run(tc.text, func(t *testing.T) {
			var got any
			if err := nbt.DecodeFromSNBT(tc.text, &got, nbt.OptDecodeFromSNBT{}); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("got %#v, want %#v", got, tc.want)
			}
		})
	}
	for _, value := range []any{[]int32{}, []string{}, [][2]int32{}, map[string]int32{}, [0]int32{}} {
		data, err := nbt.EncodeToNBTBytes(value, nbt.OptEncodeToNBT{Encoding: nbt.EncodingLittleEndian})
		if err != nil {
			t.Fatal(err)
		}
		var got any
		if err := nbt.DecodeFromNBTBytes(data, &got, nbt.OptDecodeFromNBT{Encoding: nbt.EncodingLittleEndian}); err != nil {
			t.Fatal(err)
		}
		if reflect.ValueOf(got).Len() != 0 {
			t.Fatalf("nonempty result: %#v", got)
		}
		if _, ok := value.([]int32); ok && reflect.TypeOf(got) != reflect.TypeOf(value) {
			t.Fatalf("lost empty list type: %T", got)
		}
	}
	var slice any
	if err := nbt.DecodeFromSNBT("[I;1,2]", &slice, nbt.OptDecodeFromSNBT{ArrayAsSlice: true}); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(slice, []int32{1, 2}) {
		t.Fatalf("got %#v", slice)
	}
}

func TestSNBTValidation(t *testing.T) {
	for _, text := range []string{"", " ", "{} junk", "{x:}", "{x:[1,\"two\"]}"} {
		if _, err := nbt.SNBTToNBT(text, nbt.OptSNBTToNBT{Encoding: nbt.EncodingLittleEndian}); err == nil {
			t.Errorf("accepted %q", text)
		}
	}
	var mixed []any
	if err := nbt.DecodeFromSNBT(" [1, \"two\"] \n", &mixed, nbt.OptDecodeFromSNBT{}); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(mixed, []any{int32(1), "two"}) {
		t.Fatalf("got %#v", mixed)
	}
	var wrong struct {
		X int32 `nbt:"x"`
	}
	if err := nbt.DecodeFromSNBT("{x:\"wrong type\"}", &wrong, nbt.OptDecodeFromSNBT{}); err == nil {
		t.Fatal("swallowed compound receiver error")
	}
}

func TestIntReceiverUsesFullWidth(t *testing.T) {
	// Preload high bits: writing only four bytes into a Go int/uint leaves them stale on 64-bit hosts.
	var signed int = -1
	var unsigned uint = ^uint(0)
	for _, tc := range []struct {
		ptr   any
		value int32
	}{{&signed, -2}, {&unsigned, 2}} {
		data, err := nbt.EncodeToNBTBytes(tc.value, nbt.OptEncodeToNBT{Encoding: nbt.EncodingBigEndian})
		if err != nil {
			t.Fatal(err)
		}
		if err := nbt.DecodeFromNBTBytes(data, tc.ptr, nbt.OptDecodeFromNBT{Encoding: nbt.EncodingBigEndian}); err != nil {
			t.Fatal(err)
		}
	}
	if signed != -2 || unsigned != 2 {
		t.Fatalf("got %d, %d", signed, unsigned)
	}
}

func TestNBTWireEncoding(t *testing.T) {
	// TagInt, root name "x", payload -2. Independent bytes catch symmetric codec bugs.
	expected := map[nbt.Encoding][]byte{
		nbt.EncodingLittleEndian:        {3, 1, 0, 'x', 254, 255, 255, 255},
		nbt.EncodingBigEndian:           {3, 0, 1, 'x', 255, 255, 255, 254},
		nbt.EncodingNetworkLittleEndian: {3, 1, 'x', 3},
	}
	for encoding, want := range expected {
		data, err := nbt.EncodeToNBTBytes(int32(-2), nbt.OptEncodeToNBT{Encoding: encoding, DecodeOption: decoder_go_object.DecodeOption{RootTag: "x"}})
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(data, want) {
			t.Fatalf("encoding %d: %x != %x", encoding, data, want)
		}
		var got int32
		if err := nbt.DecodeFromNBTBytes(want, &got, nbt.OptDecodeFromNBT{Encoding: encoding}); err != nil || got != -2 {
			t.Fatalf("got %d, %v", got, err)
		}
	}
}

type failWriter struct {
	remaining int
	err       error
}

func (w *failWriter) Write(p []byte) (int, error) {
	if len(p) > w.remaining {
		return 0, w.err
	}
	w.remaining -= len(p)
	return len(p), nil
}

func TestNBTWriterErrors(t *testing.T) {
	failure := errors.New("destination failed")
	// Three bytes for the root array tag/name and four for its length, then fail on an element.
	w := &failWriter{remaining: 7, err: failure}
	if err := nbt.EncodeToNBT([1]int32{42}, w, nbt.OptEncodeToNBT{Encoding: nbt.EncodingLittleEndian}); !errors.Is(err, failure) {
		t.Fatalf("lost write error: %v", err)
	}
	for _, encoding := range []nbt.Encoding{nbt.EncodingLittleEndian, nbt.EncodingBigEndian} {
		if _, err := nbt.EncodeToNBTBytes(strings.Repeat("x", 65536), nbt.OptEncodeToNBT{Encoding: encoding}); err == nil {
			t.Fatal("accepted an overflowing string length")
		}
	}
}

func TestMalformedNBT(t *testing.T) {
	for _, tag := range []byte{7, 9, 11, 12} {
		data := []byte{tag, 0, 0}
		if tag == 9 {
			data = append(data, 3)
		}
		data = append(data, 255, 255, 255, 255)
		var out any
		if err := nbt.DecodeFromNBTBytes(data, &out, nbt.OptDecodeFromNBT{Encoding: nbt.EncodingLittleEndian}); err == nil {
			t.Fatalf("accepted negative length for tag %d", tag)
		}
	}
	var out any
	if err := nbt.DecodeFromNBTBytes([]byte{3, 0, 0}, &out, nbt.OptDecodeFromNBT{Encoding: nbt.EncodingLittleEndian}); err == nil {
		t.Fatal("accepted truncated int")
	}
	if _, err := nbt.EncodeToNBTBytes(int32(1), nbt.OptEncodeToNBT{}); err == nil {
		t.Fatal("accepted unset encoding")
	}
}

func TestStreamReadsOneRoot(t *testing.T) {
	var data bytes.Buffer
	opt := nbt.OptEncodeToNBT{Encoding: nbt.EncodingNetworkLittleEndian}
	for _, v := range []int32{1, 2} {
		if err := nbt.EncodeToNBT(v, &data, opt); err != nil {
			t.Fatal(err)
		}
	}
	for _, want := range []int32{1, 2} {
		var got int32
		if err := nbt.DecodeFromNBT(&data, &got, nbt.OptDecodeFromNBT{Encoding: opt.Encoding, EncodeOption: encoder_go_object.EncodeOption{}}); err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Fatalf("got %d", got)
		}
	}
	if _, err := data.ReadByte(); err != io.EOF {
		t.Fatalf("unexpected trailing data: %v", err)
	}
}

func TestSNBTIntegerMinima(t *testing.T) {
	cases := []struct {
		value any
		want  string
	}{
		{int8(math.MinInt8), "-128b"}, {int16(math.MinInt16), "-32768s"},
		{int32(math.MinInt32), "-2147483648"}, {int64(math.MinInt64), "-9223372036854775808l"},
	}
	for _, tc := range cases {
		got, err := nbt.EncodeToSNBT(tc.value, nbt.OptEncodeToSNBT{})
		if err != nil || got != tc.want {
			t.Fatalf("got %q, %v; want %q", got, err, tc.want)
		}
	}
}

func TestStructTagsAndBool(t *testing.T) {
	type record struct {
		Enabled bool  `nbt:"enabled"`
		Ignored int32 `nbt:"-"`
		hidden  int32 `nbt:"hidden"`
	}
	var got record
	if err := nbt.DecodeFromSNBT("{enabled:true,Ignored:5,hidden:6}", &got, nbt.OptDecodeFromSNBT{}); err != nil {
		t.Fatal(err)
	}
	if got != (record{Enabled: true}) {
		t.Fatalf("got %#v", got)
	}
	value := [2]bool{true, false}
	data, err := nbt.EncodeToNBTBytes(value, nbt.OptEncodeToNBT{Encoding: nbt.EncodingLittleEndian})
	if err != nil {
		t.Fatal(err)
	}
	var restored [2]bool
	if err := nbt.DecodeFromNBTBytes(data, &restored, nbt.OptDecodeFromNBT{Encoding: nbt.EncodingLittleEndian}); err != nil {
		t.Fatal(err)
	}
	if restored != value {
		t.Fatalf("got %v", restored)
	}
}

func TestCustomValueList(t *testing.T) {
	for _, input := range [][]Distance{{1.25, 2.5}, {}} {
		data, err := nbt.EncodeToNBTBytes(input, nbt.OptEncodeToNBT{Encoding: nbt.EncodingLittleEndian})
		if err != nil {
			t.Fatal(err)
		}
		var output []Distance
		if err := nbt.DecodeFromNBTBytes(data, &output, nbt.OptDecodeFromNBT{Encoding: nbt.EncodingLittleEndian}); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(output, input) {
			t.Fatalf("got %v", output)
		}
	}
}
