package encoder_shim_test

import (
	nbt "neomega-nbt"
	decoder_go "neomega-nbt/decoder-go-object"
	decoder_snbt "neomega-nbt/decoder-snbt"
	shim "neomega-nbt/encoder-shim"
	"reflect"
	"testing"
)

type emptyRecorder struct {
	shim.VoidEncoderHighLevel
	calls []string
	types []nbt.HighLevelValueType
}

func (r *emptyRecorder) BeginComplexList(nbt.HighLevelValueType, int, bool) error {
	r.calls = append(r.calls, "begin list")
	return nil
}
func (r *emptyRecorder) EndComplexList() error { r.calls = append(r.calls, "end list"); return nil }
func (r *emptyRecorder) BeginComplexArray(nbt.HighLevelValueType, int) error {
	r.calls = append(r.calls, "begin array")
	return nil
}
func (r *emptyRecorder) EndComplexArray() error { r.calls = append(r.calls, "end array"); return nil }
func (r *emptyRecorder) BeginAndEndComplexEmptyTypedList(v nbt.HighLevelValueType) error {
	r.calls = append(r.calls, "empty typed list")
	r.types = append(r.types, v)
	return nil
}
func (r *emptyRecorder) BeginAndEndComplexEmptyArray(v nbt.HighLevelValueType) error {
	r.calls = append(r.calls, "empty array")
	r.types = append(r.types, v)
	return nil
}

func TestEmptyOperators(t *testing.T) {
	for _, array := range []bool{false, true} {
		want := "empty typed list"
		var source any = []int32{}
		tag := nbt.TagList
		if array {
			want = "empty array"
			source = [0]int32{}
			tag = nbt.TagIntArray
		}
		rec := &emptyRecorder{}
		if err := decoder_go.HighLevelDecodeTo(source, rec, decoder_go.DecodeOption{}); err != nil {
			t.Fatal(err)
		}
		enc := shim.NewEncoderShim(rec)
		for _, err := range []error{enc.WriteType(tag), enc.WriteTag("")} {
			if err != nil {
				t.Fatal(err)
			}
		}
		if !array {
			if err := enc.WriteType(nbt.TagInt); err != nil {
				t.Fatal(err)
			}
		}
		if err := enc.WriteLen(0); err != nil {
			t.Fatal(err)
		}
		if !enc.Done() {
			t.Fatal("empty root not completed")
		}
		if !reflect.DeepEqual(rec.calls, []string{want, want}) {
			t.Fatalf("unexpected lifecycle: %v", rec.calls)
		}
		if !reflect.DeepEqual(rec.types, []nbt.HighLevelValueType{nbt.TrivialInt, nbt.TrivialInt}) {
			t.Fatalf("lost element type: %v", rec.types)
		}
	}
	rec := &emptyRecorder{}
	if _, err := decoder_snbt.DecodeSnbtTo([]byte("[I; ]"), rec, decoder_snbt.DecodeOption{}); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rec.calls, []string{"empty array"}) {
		t.Fatalf("unexpected SNBT lifecycle: %v", rec.calls)
	}
}
