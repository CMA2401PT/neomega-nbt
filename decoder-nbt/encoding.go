package decoder_nbt

import (
	"encoding/binary"
	"fmt"
	"math"
	seq_io "neomega-nbt/seq-io"
	"unsafe"
)

type Encoding[R seq_io.Reader] interface {
	Int16(r R) (int16, error)
	Int32(r R) (int32, error)
	Int64(r R) (int64, error)
	Float32(r R) (float32, error)
	Float64(r R) (float64, error)
	String(r R) (string, error)
}

// BufferOverrunError is returned when the data buffer passed in when reading is overrun, meaning one of the
// reading operations extended beyond the end of the slice.
type BufferOverrunError struct {
	Op string
}

// Error ...
func (err BufferOverrunError) Error() string {
	return fmt.Sprintf("nbt: unexpected buffer end during op: '%v'", err.Op)
}

// InvalidVarintError is returned if a varint(32/64) is encountered that does
// not end after 5 or 10 bytes respectively.
type InvalidVarintError struct {
	Off int64
	N   int
}

// Error ...
func (err InvalidVarintError) Error() string {
	return fmt.Sprintf("nbt: varint did not terminate after %v bytes at offset %v", err.N, err.Off)
}

type LittleEndian[R seq_io.Reader] struct{}

// Int16 ...
func (LittleEndian[R]) Int16(r R) (int16, error) {
	var b [2]byte
	if err := r.Read(b[:]); err != nil {
		return 0, BufferOverrunError{Op: "Int16"}
	}
	u := uint16(b[0]) | uint16(b[1])<<8
	return int16(u), nil
}

// Int32 ...
func (LittleEndian[R]) Int32(r R) (int32, error) {
	var b [4]byte
	if err := r.Read(b[:]); err != nil {
		return 0, BufferOverrunError{Op: "Int32"}
	}
	u := uint32(b[0]) |
		uint32(b[1])<<8 |
		uint32(b[2])<<16 |
		uint32(b[3])<<24
	return int32(u), nil
}

// Int64 ...
func (LittleEndian[R]) Int64(r R) (int64, error) {
	var b [8]byte
	if err := r.Read(b[:]); err != nil {
		return 0, BufferOverrunError{Op: "Int64"}
	}
	u := uint64(b[0]) |
		uint64(b[1])<<8 |
		uint64(b[2])<<16 |
		uint64(b[3])<<24 |
		uint64(b[4])<<32 |
		uint64(b[5])<<40 |
		uint64(b[6])<<48 |
		uint64(b[7])<<56
	return int64(u), nil
}

// Float32 ...
func (LittleEndian[R]) Float32(r R) (float32, error) {
	var b [4]byte
	if err := r.Read(b[:]); err != nil {
		return 0, BufferOverrunError{Op: "Float32"}
	}
	u := uint32(b[0]) |
		uint32(b[1])<<8 |
		uint32(b[2])<<16 |
		uint32(b[3])<<24
	return math.Float32frombits(u), nil
}

// Float64 ...
func (LittleEndian[R]) Float64(r R) (float64, error) {
	var b [8]byte
	if err := r.Read(b[:]); err != nil {
		return 0, BufferOverrunError{Op: "Float64"}
	}
	u := uint64(b[0]) |
		uint64(b[1])<<8 |
		uint64(b[2])<<16 |
		uint64(b[3])<<24 |
		uint64(b[4])<<32 |
		uint64(b[5])<<40 |
		uint64(b[6])<<48 |
		uint64(b[7])<<56
	return math.Float64frombits(u), nil
}

// String ...
func (e LittleEndian[R]) String(r R) (string, error) {
	strLen, err := e.Int16(r)
	if err != nil {
		return "", BufferOverrunError{Op: "String"}
	}
	b := make([]byte, uint16(strLen))
	if err := r.Read(b); err != nil {
		return "", BufferOverrunError{Op: "String"}
	}
	return *(*string)(unsafe.Pointer(&b)), nil
}

type BigEndian[R seq_io.Reader] struct{}

// Int16 ...
func (BigEndian[R]) Int16(r R) (int16, error) {
	b := make([]byte, 2)
	if err := r.Read(b); err != nil {
		return 0, BufferOverrunError{Op: "Int16"}
	}
	return int16(binary.BigEndian.Uint16(b)), nil
}

// Int32 ...
func (BigEndian[R]) Int32(r R) (int32, error) {
	b := make([]byte, 4)
	if err := r.Read(b); err != nil {
		return 0, BufferOverrunError{Op: "Int32"}
	}
	return int32(binary.BigEndian.Uint32(b)), nil
}

// Int64 ...
func (BigEndian[R]) Int64(r R) (int64, error) {
	b := make([]byte, 8)
	if err := r.Read(b); err != nil {
		return 0, BufferOverrunError{Op: "Float64"}
	}
	return int64(binary.BigEndian.Uint64(b)), nil
}

// Float32 ...
func (BigEndian[R]) Float32(r R) (float32, error) {
	b := make([]byte, 4)
	if err := r.Read(b); err != nil {
		return 0, BufferOverrunError{Op: "Float32"}
	}
	return math.Float32frombits(binary.BigEndian.Uint32(b)), nil
}

// Float64 ...
func (BigEndian[R]) Float64(r R) (float64, error) {
	b := make([]byte, 8)
	if err := r.Read(b); err != nil {
		return 0, BufferOverrunError{Op: "Float64"}
	}
	return math.Float64frombits(binary.BigEndian.Uint64(b)), nil
}

// String ...
func (e BigEndian[R]) String(r R) (string, error) {
	strLen, err := e.Int16(r)
	if err != nil {
		return "", BufferOverrunError{Op: "String"}
	}
	b := make([]byte, uint16(strLen))
	if err := r.Read(b); err != nil {
		return "", BufferOverrunError{Op: "String"}
	}
	return *(*string)(unsafe.Pointer(&b)), nil
}

type NetworklittleEndian[R seq_io.Reader] struct{ LittleEndian[R] }

// Int32 ...
func (NetworklittleEndian[R]) Int32(r R) (int32, error) {
	var ux uint32
	for i := uint(0); i < 35; i += 7 {
		b, err := r.ReadByte()
		if err != nil {
			return 0, BufferOverrunError{Op: "Int32"}
		}
		ux |= uint32(b&0x7f) << i
		if b&0x80 == 0 {
			x := int32(ux >> 1)
			if ux&1 != 0 {
				x = ^x
			}
			return x, nil
		}
	}
	return 0, InvalidVarintError{N: 5}
}

// Int64 ...
func (NetworklittleEndian[R]) Int64(r R) (int64, error) {
	var ux uint64
	for i := uint(0); i < 70; i += 7 {
		b, err := r.ReadByte()
		if err != nil {
			return 0, BufferOverrunError{Op: "Int64"}
		}
		ux |= uint64(b&0x7f) << i
		if b&0x80 == 0 {
			x := int64(ux >> 1)
			if ux&1 != 0 {
				x = ^x
			}
			return x, nil
		}
	}
	return 0, InvalidVarintError{N: 10}
}

// String ...
func (e NetworklittleEndian[R]) String(r R) (string, error) {
	length, err := e.stringLength(r)
	if err != nil {
		return "", err
	}
	data := make([]byte, length)
	if err := r.Read(data); err != nil {
		return "", BufferOverrunError{Op: "String"}
	}
	return *(*string)(unsafe.Pointer(&data)), nil
}

// stringLength reads the length of a string as a varuint32.
func (NetworklittleEndian[R]) stringLength(r R) (uint32, error) {
	var ux uint32
	for i := uint(0); i < 35; i += 7 {
		b, err := r.ReadByte()
		if err != nil {
			return 0, BufferOverrunError{Op: "StringLength"}
		}
		ux |= uint32(b&0x7f) << i
		if b&0x80 == 0 {
			return ux, nil
		}
	}
	return 0, InvalidVarintError{N: 5}
}
