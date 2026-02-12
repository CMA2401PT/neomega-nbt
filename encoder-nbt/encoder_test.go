package encoder_nbt

import (
	"fmt"
	nbt "neomega-nbt"
	seq_io "neomega-nbt/seq-io"
	"testing"
)

func TestEncoderLittleEdian(t *testing.T) {
	enc, getter := NewEncoderToBytes[LittleEndian[*seq_io.SliceGrowWriter]](nil)
	enc.WriteType(nbt.TagInt32)
	enc.WriteTag("12")
	enc.WriteInt32(7)
	ret := getter()
	fmt.Println("ret", ret)
}
