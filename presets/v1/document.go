package v1

import (
	"bytes"
	"fmt"
	"io"
	"reflect"

	nbt "neomega-nbt"
	encoder_go_object "neomega-nbt/encoder-go-object"
	encoder_shim "neomega-nbt/encoder-shim"
	seq_io "neomega-nbt/seq-io"
)

// Compound uses typed Go values. Slices represent lists; numeric Go arrays
// represent NBT arrays, so their tag types survive an edit and re-encoding.
type Compound = map[string]any

// Document retains the name and compound of one NBT root.
type Document struct {
	Name  string
	Value Compound
}

type documentEncoder struct {
	nbt.EncoderHighLevel
	document *Document
}

func (e *documentEncoder) WriteRootTag(name string, kind nbt.HighLevelValueType) error {
	if kind != nbt.ComplexCompound {
		return fmt.Errorf("NBT root is %s, expected compound", kind)
	}
	e.document.Name = name
	return e.EncoderHighLevel.WriteRootTag(name, kind)
}

// ReadDocument consumes exactly one root, leaving following records unread.
func ReadDocument(r io.Reader, encoding Encoding) (Document, error) {
	var d Document
	enc := &documentEncoder{
		EncoderHighLevel: encoder_go_object.NewGoObjEncoder(&d.Value, OptDecodeFromNBT{}.EncodeOption),
		document:         &d,
	}
	err := decodeNBT(seq_io.NewOffsetReader(r), encoder_shim.NewEncoderShim(enc), encoding)
	return d, err
}

// DecodeDocument requires one complete root without trailing bytes.
func DecodeDocument(data []byte, encoding Encoding) (Document, error) {
	r := bytes.NewReader(data)
	d, err := ReadDocument(r, encoding)
	if err == nil && r.Len() != 0 {
		err = fmt.Errorf("%d trailing NBT bytes", r.Len())
	}
	return d, err
}

func EncodeDocument(d Document, encoding Encoding) ([]byte, error) {
	opt := OptEncodeToNBT{Encoding: encoding}
	opt.RootTag = d.Name
	return EncodeToNBTBytes(d.Value, opt)
}

// Compounds accepts a compound list or any empty List, whose element type has
// no values to constrain. NBT Arrays remain distinct, including empty Arrays.
func Compounds(value any) ([]Compound, bool) {
	if values, ok := value.([]Compound); ok {
		return values, true
	}
	if values := reflect.ValueOf(value); values.IsValid() && values.Kind() == reflect.Slice && values.Len() == 0 {
		return nil, true
	}
	return nil, false
}

type arrayElement interface {
	~uint8 | ~int8 | ~int32 | ~int64
}

// Array copies values into a Go array, explicitly selecting an NBT array tag.
// Other numeric slices in the same compound remain NBT lists.
func Array[T arrayElement](values []T) any {
	a := reflect.New(reflect.ArrayOf(len(values), reflect.TypeFor[T]())).Elem()
	reflect.Copy(a, reflect.ValueOf(values))
	return a.Interface()
}

// ArrayValues copies a numeric NBT array decoded with the default options.
func ArrayValues[T arrayElement](value any) ([]T, bool) {
	a := reflect.ValueOf(value)
	if !a.IsValid() || a.Kind() != reflect.Array || a.Type().Elem() != reflect.TypeFor[T]() {
		return nil, false
	}
	out := make([]T, a.Len())
	reflect.Copy(reflect.ValueOf(out), a)
	return out, true
}

// DecodeDocuments reads consecutive compound roots, as used by record streams.
func DecodeDocuments(data []byte, encoding Encoding) ([]Document, error) {
	r := bytes.NewReader(data)
	var documents []Document
	for r.Len() > 0 {
		document, err := ReadDocument(r, encoding)
		if err != nil {
			return nil, err
		}
		documents = append(documents, document)
	}
	return documents, nil
}

// EncodeDocuments concatenates complete roots without adding a list wrapper.
func EncodeDocuments(documents []Document, encoding Encoding) ([]byte, error) {
	var result []byte
	for _, document := range documents {
		data, err := EncodeDocument(document, encoding)
		if err != nil {
			return nil, err
		}
		result = append(result, data...)
	}
	return result, nil
}
