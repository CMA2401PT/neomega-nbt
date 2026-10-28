package decoder_snbt_test

import (
	"errors"
	nbt "neomega-nbt"
	decoder "neomega-nbt/decoder-snbt"
	shim "neomega-nbt/encoder-shim"
	"testing"
)

type failingReceiver struct {
	shim.VoidEncoderHighLevel
	err error
}

func (e failingReceiver) WriteInt32(int32, nbt.ValuePos) error { return e.err }

func TestReceiverErrors(t *testing.T) {
	failure := errors.New("receiver failure")
	for _, input := range []string{"{x:1}", "[I;1]", "[1]"} {
		_, err := decoder.DecodeSnbtTo([]byte(input), failingReceiver{err: failure}, decoder.DecodeOption{})
		if !errors.Is(err, failure) {
			t.Errorf("%s: lost error: %v", input, err)
		}
	}
}

func TestLeadingWhitespaceAndRest(t *testing.T) {
	rest, err := decoder.DecodeSnbtTo([]byte(" \n {x:1} suffix"), shim.VoidEncoderHighLevel{}, decoder.DecodeOption{})
	if err != nil {
		t.Fatal(err)
	}
	if string(rest) != " suffix" {
		t.Fatalf("got rest %q", rest)
	}
}
