package encoder_snbt

import "unsafe"

func writeSNBTString[W Writer](w W, data string) error {
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

func writeString[W Writer](w W, data string) error {
	bs := unsafe.Slice(unsafe.StringData(data), len(data))
	if _, err := w.Write(bs); err != nil {
		return err
	}
	return nil
}

func writeInt[W Writer, T interface {
	int8 | uint8 | int16 | uint16 | int32 | uint32 | int64 | uint64
}](w W, i T) error {
	buf := [32]byte{}
	if i < 0 {
		if err := w.WriteByte('-'); err != nil {
			return err
		}
		i = -i
	}
	off := 0
	for i >= 10 {
		buf[off] = uint8(i%10) + '0'
		off += 1
		i = i / 10
	}
	if err := w.WriteByte(uint8(i) + '0'); err != nil {
		return err
	}
	for off != 0 {
		if err := w.WriteByte(buf[off-1]); err != nil {
			return err
		}
		off -= 1
	}
	return nil
}
