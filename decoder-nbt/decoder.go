package decoder_nbt

import (
	"errors"
	"fmt"
	"io"
	nbt "neomega-nbt"
)

func DecodeFromIOReader[E Encoding[offsetReader]](r io.Reader, e nbt.Encoder) (err error) {
	reader := newOffsetReader(r)
	return Decode[offsetReader, E](reader, e)
}

func DecodeFromBytes[Seq ~string | ~[]uint8, E Encoding[*ByteReader[Seq]]](data Seq, e nbt.Encoder) (err error) {
	reader := NewByteReaderFromBytes(data)
	return Decode[*ByteReader[Seq], E](reader, e)
}

func Decode[R Reader, E Encoding[R]](r R, e nbt.Encoder) (err error) {
	var tagTypeByte uint8
	var tagName string
	var enc E
	tagTypeByte, err = r.ReadByte()
	if err != nil {
		return BufferOverrunError{Op: "ReadTag"}
	}
	t := nbt.TagType(tagTypeByte)
	if t == nbt.TagEnd {
		return e.WriteEnd()
	}

	err = e.WriteType(t)
	if err != nil {
		return BufferOverrunError{Op: "ReadTag"}
	}

	tagName, err = enc.String(r)
	if err != nil {
		return err
	}
	err = e.WriteTag(tagName)
	if err != nil {
		return err
	}
	return decodeValue[R, E](r, e, t)
}

func decodeValue[R Reader, E Encoding[R]](r R, e nbt.Encoder, t nbt.TagType) (err error) {
	var enc E
	switch t {
	default:
		return fmt.Errorf("unexpected nbtType: %v", t)
	case nbt.TagByte:
		value, err := r.ReadByte()
		if err != nil {
			return BufferOverrunError{Op: "Byte"}
		}
		return e.WriteInt8(int8(value))
	case nbt.TagShort:
		value, err := enc.Int16(r)
		if err != nil {
			return err
		}
		return e.WriteInt16(value)
	case nbt.TagInt:
		value, err := enc.Int32(r)
		if err != nil {
			return err
		}
		return e.WriteInt32(value)
	case nbt.TagLong:
		value, err := enc.Int64(r)
		if err != nil {
			return err
		}
		return e.WriteInt64(value)
	case nbt.TagFloat:
		value, err := enc.Float32(r)
		if err != nil {
			return err
		}
		return e.WriteFloat32(value)
	case nbt.TagDouble:
		value, err := enc.Float64(r)
		if err != nil {
			return err
		}
		return e.WriteFloat64(value)
	case nbt.TagString:
		value, err := enc.String(r)
		if err != nil {
			return err
		}
		return e.WriteString(value)
	case nbt.TagByteArray:
		length, err := enc.Int32(r)
		if err != nil {
			return err
		}
		err = e.WriteLen(uint32(length))
		if err != nil {
			return err
		}
		var b uint8
		for _ = range length {
			b, err = r.ReadByte()
			if err != nil {
				return err
			}
			err = e.WriteInt8(int8(b))
			if err != nil {
				return err
			}
		}
	case nbt.TagIntArray:
		length, err := enc.Int32(r)
		if err != nil {
			return err
		}
		err = e.WriteLen(uint32(length))
		if err != nil {
			return err
		}
		var i32 int32
		for _ = range length {
			i32, err = enc.Int32(r)
			if err != nil {
				return err
			}
			err = e.WriteInt32(i32)
			if err != nil {
				return err
			}
		}
	case nbt.TagLongArray:
		length, err := enc.Int32(r)
		if err != nil {
			return err
		}
		err = e.WriteLen(uint32(length))
		if err != nil {
			return err
		}
		var i64 int64
		for range length {
			i64, err = enc.Int64(r)
			if err != nil {
				return err
			}
			err = e.WriteInt64(i64)
			if err != nil {
				return err
			}
		}
	case nbt.TagList:
		var listTypeByte uint8
		var listLength int32
		listTypeByte, err = r.ReadByte()
		if err != nil {
			return BufferOverrunError{Op: "Slice"}
		}
		listType := nbt.TagType(listTypeByte)
		listLength, err = enc.Int32(r)
		if err != nil {
			return err
		}
		if listType == nbt.TagEnd {
			if listLength != 0 {
				return errors.New("for list which cannot infer type, type should be tag end ")
			}
			return e.WriteEmptyListMark()
		}
		err = e.WriteType(listType)
		if err != nil {
			return err
		}
		err = e.WriteLen(uint32(listLength))
		if err != nil {
			return err
		}
		for range listLength {
			err = decodeValue[R, E](r, e, listType)
			if err != nil {
				return err
			}
		}
	case nbt.TagCompound:
		var tagTypeByte uint8
		var tagName string
		var enc E
		for {
			tagTypeByte, err = r.ReadByte()
			if err != nil {
				return BufferOverrunError{Op: "ReadTag"}
			}
			t := nbt.TagType(tagTypeByte)
			if t == nbt.TagEnd {
				return e.WriteEnd()
			}

			err = e.WriteType(t)
			if err != nil {
				return BufferOverrunError{Op: "ReadTag"}
			}

			tagName, err = enc.String(r)
			if err != nil {
				return err
			}
			err = e.WriteTag(tagName)
			if err != nil {
				return err
			}
			err = decodeValue[R, E](r, e, t)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
