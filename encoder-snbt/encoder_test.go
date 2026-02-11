package encoder_snbt

import (
	"bufio"
	"bytes"
	"fmt"
	neomega_nbt "neomega-nbt"
	shim "neomega-nbt/encoder-shim"
	"testing"
)

type SimpleSnbtEncoderHighLevel struct {
	shim.EncoderShim[SnbtEncoderHighLevel[*bufio.Writer]]
	buf    *bytes.Buffer
	writer *bufio.Writer
}

func (e *SimpleSnbtEncoderHighLevel) Finish() error {
	e.writer.Flush()
	fmt.Println("out:", string(e.buf.String()))
	return nil
}

func NewSimpleSnbtEncoderHighLevel() *SimpleSnbtEncoderHighLevel {
	buf := bytes.NewBuffer(nil)
	writer := bufio.NewWriter(buf)
	enc := shim.EncoderShim[SnbtEncoderHighLevel[*bufio.Writer]]{
		HighLevelEncoder: NewSnbtEncoderHighLevel(writer),
	}
	return &SimpleSnbtEncoderHighLevel{
		EncoderShim: enc,
		buf:         buf,
		writer:      writer,
	}
}

type EncoderWithFinishCheck interface {
	neomega_nbt.Encoder
	Finish() error
}

func _TestEncoderRestriction(t *testing.T, makeEncoder func() EncoderWithFinishCheck) {
	ok := func(err error) {
		t.Helper()
		if err != nil {
			t.Error(err)
		}
	}
	done := func(err error) {
		t.Helper()
		if err != nil {
			t.Error(err)
		}
	}
	// normal value

	// uint8, int8
	e := makeEncoder()
	ok(e.WriteType(neomega_nbt.TagByte))
	ok(e.WriteTag("tag"))
	ok(e.WriteInt8(0))
	done(e.Finish())

	e = makeEncoder()
	ok(e.WriteType(neomega_nbt.TagByte))
	ok(e.WriteTag("tag"))
	ok(e.WriteInt8(0))
	done(e.Finish())

	// string
	e = makeEncoder()
	ok(e.WriteType(neomega_nbt.TagString))
	ok(e.WriteTag("tag"))
	ok(e.WriteString("string"))
	done(e.Finish())

	// array
	e = makeEncoder()
	ok(e.WriteType(neomega_nbt.TagByteArray))
	ok(e.WriteTag("tag"))
	ok(e.WriteLen(0))
	done(e.Finish())

	e = makeEncoder()
	ok(e.WriteType(neomega_nbt.TagByteArray))
	ok(e.WriteTag("tag"))
	ok(e.WriteLen(3))
	ok(e.WriteInt8(4))
	ok(e.WriteInt8(2))
	ok(e.WriteInt8(3))
	done(e.Finish())

	// int array
	e = makeEncoder()
	ok(e.WriteType(neomega_nbt.TagIntArray))
	ok(e.WriteTag("tag"))
	ok(e.WriteLen(0))
	done(e.Finish())

	e = makeEncoder()
	ok(e.WriteType(neomega_nbt.TagIntArray))
	ok(e.WriteTag("tag"))
	ok(e.WriteLen(3))
	ok(e.WriteInt32(1))
	ok(e.WriteInt32(2))
	ok(e.WriteInt32(3))
	done(e.Finish())

	// long array
	e = makeEncoder()
	ok(e.WriteType(neomega_nbt.TagLongArray))
	ok(e.WriteTag("tag"))
	ok(e.WriteLen(0))
	done(e.Finish())

	e = makeEncoder()
	ok(e.WriteType(neomega_nbt.TagLongArray))
	ok(e.WriteTag("tag"))
	ok(e.WriteLen(3))
	ok(e.WriteInt64(1))
	ok(e.WriteInt64(2))
	ok(e.WriteInt64(3))
	done(e.Finish())

	// list
	e = makeEncoder()
	ok(e.WriteType(neomega_nbt.TagList))
	ok(e.WriteTag("tag"))
	ok(e.WriteEmptyListMark())
	done(e.Finish())

	e = makeEncoder()
	ok(e.WriteType(neomega_nbt.TagList))
	ok(e.WriteTag("tag"))
	ok(e.WriteType(neomega_nbt.TagInt32))
	ok(e.WriteLen(0))
	done(e.Finish())

	e = makeEncoder()
	ok(e.WriteType(neomega_nbt.TagList))
	ok(e.WriteTag("tag"))
	ok(e.WriteType(neomega_nbt.TagInt32))
	ok(e.WriteLen(3))
	ok(e.WriteInt32(1))
	ok(e.WriteInt32(2))
	ok(e.WriteInt32(3))
	done(e.Finish())

	e = makeEncoder()
	ok(e.WriteType(neomega_nbt.TagList))
	ok(e.WriteTag("tag"))
	ok(e.WriteType(neomega_nbt.TagInt64))
	ok(e.WriteLen(3))
	ok(e.WriteInt64(1))
	ok(e.WriteInt64(2))
	ok(e.WriteInt64(3))
	done(e.Finish())

	e = makeEncoder()
	ok(e.WriteType(neomega_nbt.TagList))
	ok(e.WriteTag("tag"))
	ok(e.WriteType(neomega_nbt.TagList))
	ok(e.WriteLen(5))
	for i := range 5 {
		switch i {
		case 0:
			ok(e.WriteEmptyListMark())
		case 1:
			ok(e.WriteType(neomega_nbt.TagInt32))
			ok(e.WriteLen(0))
		case 2:
			ok(e.WriteType(neomega_nbt.TagByteArray))
			ok(e.WriteLen(2))
			ok(e.WriteLen(0)) // 0 of first byte array
			ok(e.WriteLen(2))
			ok(e.WriteInt8(1))
			ok(e.WriteInt8(2))
		case 3:
			ok(e.WriteType(neomega_nbt.TagList))
			ok(e.WriteLen(1))
			ok(e.WriteType(neomega_nbt.TagLong))
			ok(e.WriteLen(3))
			ok(e.WriteInt64(3))
			ok(e.WriteInt64(4))
			ok(e.WriteInt64(3))
		default:
			ok(e.WriteType(neomega_nbt.TagInt32))
			ok(e.WriteLen(uint32(i)))
			for j := range i {
				ok(e.WriteInt32(int32(j)))
			}
		}
	}
	done(e.Finish())

	// compound
	e = makeEncoder()
	ok(e.WriteType(neomega_nbt.TagCompound))
	ok(e.WriteTag("tag"))
	ok(e.WriteEnd())
	done(e.Finish())

	e = makeEncoder()
	ok(e.WriteType(neomega_nbt.TagCompound))
	ok(e.WriteTag("tag"))
	for i := range 6 {
		switch i {
		case 0:
			ok(e.WriteType(neomega_nbt.TagList))
			ok(e.WriteTag("0"))
			ok(e.WriteEmptyListMark())
		case 1:
			ok(e.WriteType(neomega_nbt.TagList))
			ok(e.WriteTag("1"))
			ok(e.WriteType(neomega_nbt.TagInt32))
			ok(e.WriteLen(0))

		case 2:
			ok(e.WriteType(neomega_nbt.TagList))
			ok(e.WriteTag("2"))
			ok(e.WriteType(neomega_nbt.TagInt32))
			ok(e.WriteLen(uint32(2)))
			for j := range 2 {
				ok(e.WriteInt32(int32(j)))
			}
		case 3:
			ok(e.WriteType(neomega_nbt.TagCompound))
			ok(e.WriteTag("3"))
			ok(e.WriteType(neomega_nbt.TagList))
			ok(e.WriteTag("2"))
			ok(e.WriteType(neomega_nbt.TagInt32))
			ok(e.WriteLen(uint32(2)))
			for j := range 2 {
				ok(e.WriteInt32(int32(j)))
			}
			ok(e.WriteType(neomega_nbt.TagList))
			ok(e.WriteTag("4"))
			ok(e.WriteType(neomega_nbt.TagInt32))
			ok(e.WriteLen(uint32(2)))
			for j := range 2 {
				ok(e.WriteInt32(int32(j + 1)))
			}
			ok(e.WriteEnd())
		case 4:
			ok(e.WriteType(neomega_nbt.TagString))
			ok(e.WriteTag("tag"))
			ok(e.WriteString("string"))
		default:
			ok(e.WriteType(neomega_nbt.TagInt32))
			ok(e.WriteTag(fmt.Sprintf("value-%v", i)))
			ok(e.WriteInt32(int32(i)))
		}
	}
	ok(e.WriteEnd())
	done(e.Finish())

}

func TestEncoderShim(t *testing.T) {
	_TestEncoderRestriction(t, func() EncoderWithFinishCheck { return NewSimpleSnbtEncoderHighLevel() })
}
