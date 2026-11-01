package v1

import (
	"bytes"
	"reflect"
	"testing"
)

func TestDocumentRetainsRootAndContainerTypes(t *testing.T) {
	doc := Document{Name: "named root", Value: Compound{
		"list": []int32{1, 2}, "array": Array([]int32{1, 2}),
		"emptyList": []int64{}, "emptyArray": Array([]int64{}),
		"emptyCompounds": []Compound{}, "nested": []Compound{{"byteArray": Array([]byte{0, 255})}},
	}}
	for _, encoding := range []Encoding{EncodingLittleEndian, EncodingBigEndian, EncodingNetworkLittleEndian} {
		raw, err := EncodeDocument(doc, encoding)
		if err != nil {
			t.Fatal(err)
		}
		decoded, err := DecodeDocument(raw, encoding)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(decoded, doc) {
			t.Fatalf("encoding %d: %#v != %#v", encoding, decoded, doc)
		}
		if _, err = DecodeDocument(append(raw, 0), encoding); err == nil {
			t.Fatal("accepted trailing bytes")
		}
		stream := bytes.NewReader(append(bytes.Clone(raw), raw...))
		for range 2 {
			if _, err = ReadDocument(stream, encoding); err != nil {
				t.Fatal(err)
			}
		}
		if stream.Len() != 0 {
			t.Fatal("did not consume roots")
		}
		values, ok := ArrayValues[int32](decoded.Value["array"])
		if !ok || !reflect.DeepEqual(values, []int32{1, 2}) {
			t.Fatal(values, ok)
		}
		if _, ok = ArrayValues[int32](decoded.Value["list"]); ok {
			t.Fatal("list accepted as array")
		}
	}
}

func TestCompoundsAcceptsEmptyListsButNotArrays(t *testing.T) {
	for _, value := range []any{[]byte{}, []int32{}, []string{}, []any{}, []Compound{}, []Compound{{"id": "test"}}} {
		raw, err := EncodeDocument(Document{Value: Compound{"value": value}}, EncodingBigEndian)
		if err != nil {
			t.Fatal(err)
		}
		doc, err := DecodeDocument(raw, EncodingBigEndian)
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := Compounds(doc.Value["value"]); !ok {
			t.Fatalf("rejected list %T", value)
		}
	}
	for _, value := range []any{Array([]byte{}), Array([]int32{}), []byte{1}, []string{"x"}, []any{Compound{}}, Compound{}, "", nil} {
		if _, ok := Compounds(value); ok {
			t.Fatalf("accepted non-compound list or non-list %T", value)
		}
	}
}
