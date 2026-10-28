package encoder_snbt

import (
	seq_io "neomega-nbt/seq-io"
	"unsafe"
)

func writeSNBTString[W seq_io.Writer](w W, data string) error {
	if err := w.WriteByte('"'); err != nil {
		return err
	}
	bs := unsafe.Slice(unsafe.StringData(data), len(data))
	for _, b := range bs {
		if b == '\\' || b == '"' {
			if err := w.WriteByte('\\'); err != nil {
				return err
			}
		}
		if err := w.WriteByte(b); err != nil {
			return err
		}
	}
	if err := w.WriteByte('"'); err != nil {
		return err
	}
	return nil
}

func writeString[W seq_io.Writer](w W, data string) error {
	bs := unsafe.Slice(unsafe.StringData(data), len(data))
	if err := w.Write(bs); err != nil {
		return err
	}
	return nil
}

func writeInt[W seq_io.Writer, T interface {
	int8 | uint8 | int16 | uint16 | int32 | uint32 | int64 | uint64
}](w W, i T) error {
	// Convert the magnitude in uint64 so negating the smallest signed value cannot overflow.
	var magnitude uint64
	if i < 0 {
		if err := w.WriteByte('-'); err != nil {
			return err
		}
		magnitude = uint64(-(i + 1)) + 1
	} else {
		magnitude = uint64(i)
	}
	var buf [20]byte
	off := len(buf)
	for {
		off--
		buf[off] = byte(magnitude%10) + '0'
		magnitude /= 10
		if magnitude == 0 {
			break
		}
	}
	return w.Write(buf[off:])
}
