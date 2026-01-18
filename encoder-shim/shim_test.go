package encoder_shim

import (
	"fmt"
	nbt "neomega-nbt"
	"testing"
)

func _TestEncoderRestriction(t *testing.T, makeEncoder func() interface {
	nbt.Encoder
	Done() bool
}) {
	ok := func(err error) {
		t.Helper()
		if err != nil {
			t.Error(err)
		}
	}
	done := func(b bool) {
		t.Helper()
		if !b {
			t.Error("encoder should finished")
		}
	}
	// normal value

	// uint8, int8
	e := makeEncoder()
	ok(e.WriteType(nbt.TagByte))
	ok(e.WriteTag("tag"))
	ok(e.WriteUint8(0))
	done(e.Done())

	e = makeEncoder()
	ok(e.WriteType(nbt.TagByte))
	ok(e.WriteTag("tag"))
	ok(e.WriteInt8(0))
	done(e.Done())

	// string
	e = makeEncoder()
	ok(e.WriteType(nbt.TagString))
	ok(e.WriteTag("tag"))
	ok(e.WriteString("string"))
	done(e.Done())

	// array
	e = makeEncoder()
	ok(e.WriteType(nbt.TagByteArray))
	ok(e.WriteTag("tag"))
	ok(e.WriteLen(0))
	done(e.Done())

	e = makeEncoder()
	ok(e.WriteType(nbt.TagByteArray))
	ok(e.WriteTag("tag"))
	ok(e.WriteLen(3))
	ok(e.WriteInt8(1))
	ok(e.WriteUint8(2))
	ok(e.WriteInt8(3))
	done(e.Done())

	// int array
	e = makeEncoder()
	ok(e.WriteType(nbt.TagIntArray))
	ok(e.WriteTag("tag"))
	ok(e.WriteLen(0))
	done(e.Done())

	e = makeEncoder()
	ok(e.WriteType(nbt.TagIntArray))
	ok(e.WriteTag("tag"))
	ok(e.WriteLen(3))
	ok(e.WriteInt32(1))
	ok(e.WriteUint32(2))
	ok(e.WriteUint32(3))
	done(e.Done())

	// long array
	e = makeEncoder()
	ok(e.WriteType(nbt.TagLongArray))
	ok(e.WriteTag("tag"))
	ok(e.WriteLen(0))
	done(e.Done())

	e = makeEncoder()
	ok(e.WriteType(nbt.TagLongArray))
	ok(e.WriteTag("tag"))
	ok(e.WriteLen(3))
	ok(e.WriteInt64(1))
	ok(e.WriteUint64(2))
	ok(e.WriteUint64(3))
	done(e.Done())

	// list
	e = makeEncoder()
	ok(e.WriteType(nbt.TagList))
	ok(e.WriteTag("tag"))
	ok(e.WriteEmptyListMark())
	done(e.Done())

	e = makeEncoder()
	ok(e.WriteType(nbt.TagList))
	ok(e.WriteTag("tag"))
	ok(e.WriteType(nbt.TagInt32))
	ok(e.WriteLen(0))
	done(e.Done())

	e = makeEncoder()
	ok(e.WriteType(nbt.TagList))
	ok(e.WriteTag("tag"))
	ok(e.WriteType(nbt.TagInt32))
	ok(e.WriteLen(3))
	ok(e.WriteInt32(1))
	ok(e.WriteInt32(2))
	ok(e.WriteInt32(3))
	done(e.Done())

	e = makeEncoder()
	ok(e.WriteType(nbt.TagList))
	ok(e.WriteTag("tag"))
	ok(e.WriteType(nbt.TagList))
	ok(e.WriteLen(5))
	for i := range 5 {
		switch i {
		case 0:
			ok(e.WriteEmptyListMark())
		case 1:
			ok(e.WriteType(nbt.TagInt32))
			ok(e.WriteLen(0))
		case 2:
			ok(e.WriteType(nbt.TagByteArray))
			ok(e.WriteLen(2))
			ok(e.WriteLen(0)) // 0 of first byte array
			ok(e.WriteLen(2))
			ok(e.WriteInt8(1))
			ok(e.WriteInt8(2))
		default:
			ok(e.WriteType(nbt.TagInt32))
			ok(e.WriteLen(uint32(i)))
			for j := range i {
				ok(e.WriteInt32(int32(j)))
			}
		}
	}
	done(e.Done())

	// compound
	e = makeEncoder()
	ok(e.WriteType(nbt.TagCompound))
	ok(e.WriteTag("tag"))
	ok(e.WriteEnd())
	done(e.Done())

	e = makeEncoder()
	ok(e.WriteType(nbt.TagCompound))
	ok(e.WriteTag("tag"))
	for i := range 5 {
		switch i {
		case 0:
			ok(e.WriteType(nbt.TagList))
			ok(e.WriteTag("0"))
			ok(e.WriteEmptyListMark())
		case 1:
			ok(e.WriteType(nbt.TagList))
			ok(e.WriteTag("1"))
			ok(e.WriteType(nbt.TagInt32))
			ok(e.WriteLen(0))

		case 2:
			ok(e.WriteType(nbt.TagList))
			ok(e.WriteTag("2"))
			ok(e.WriteType(nbt.TagInt32))
			ok(e.WriteLen(uint32(2)))
			for j := range 2 {
				ok(e.WriteInt32(int32(j)))
			}
		case 3:
			ok(e.WriteType(nbt.TagCompound))
			ok(e.WriteTag("3"))
			ok(e.WriteEnd())
		case 4:
			ok(e.WriteType(nbt.TagString))
			ok(e.WriteTag("tag"))
			ok(e.WriteString("string"))
		default:
			ok(e.WriteType(nbt.TagInt32))
			ok(e.WriteTag(fmt.Sprintf("value-%v", i)))
			ok(e.WriteInt32(int32(i)))
		}
	}
	ok(e.WriteEnd())
	done(e.Done())

}

func TestEncoderShim(t *testing.T) {
	_TestEncoderRestriction(t, func() interface {
		nbt.Encoder
		Done() bool
	} {
		return &EncoderShim[VoidEncoderHighLevel]{}
	})
}
