package encoder_nbt

import (
	"encoding/binary"
	"fmt"
	"math"
	"unsafe"
)

type Encoding[W Writer] interface {
	WriteInt16(w W, x int16) error
	WriteInt32(w W, x int32) error
	WriteInt64(w W, x int64) error
	WriteFloat32(w W, x float32) error
	WriteFloat64(w W, x float64) error
	WriteString(w W, x string) error
}
type LittleEndian[W Writer] struct{}

// FailedWriteError is returned if a Write operation failed on an W, meaning some of the data could
// not be written to the io.Writer.
type FailedWriteError struct {
	Op  string
	Err error
}

// Error ...
func (err FailedWriteError) Error() string {
	return fmt.Sprintf("nbt: failed write during op '%v' : %v", err.Op, err.Err)
}

// WriteInt16 ...
func (LittleEndian[W]) WriteInt16(w W, x int16) error {
	b := *(*[2]byte)(unsafe.Pointer(&x))
	if _, err := w.Write(b[:]); err != nil {
		return FailedWriteError{Op: "WriteInt16"}
	}
	return nil
}

// WriteInt32 ...
func (LittleEndian[W]) WriteInt32(w W, x int32) error {
	b := *(*[4]byte)(unsafe.Pointer(&x))
	if _, err := w.Write(b[:]); err != nil {
		return FailedWriteError{Op: "WriteInt32"}
	}
	return nil
}

// WriteInt64 ...
func (LittleEndian[W]) WriteInt64(w W, x int64) error {
	b := *(*[8]byte)(unsafe.Pointer(&x))
	if _, err := w.Write(b[:]); err != nil {
		return FailedWriteError{Op: "WriteInt64"}
	}
	return nil
}

// WriteFloat32 ...
func (LittleEndian[W]) WriteFloat32(w W, x float32) error {
	b := *(*[4]byte)(unsafe.Pointer(&x))
	if _, err := w.Write(b[:]); err != nil {
		return FailedWriteError{Op: "WriteFloat32"}
	}
	return nil
}

// WriteFloat64 ...
func (LittleEndian[W]) WriteFloat64(w W, x float64) error {
	b := *(*[8]byte)(unsafe.Pointer(&x))
	if _, err := w.Write(b[:]); err != nil {
		return FailedWriteError{Op: "WriteFloat64"}
	}
	return nil
}

// WriteString ...
func (e LittleEndian[W]) WriteString(w W, x string) error {
	if err := e.WriteInt16(w, int16(uint16(len(x)))); err != nil {
		return FailedWriteError{Op: "WriteString"}
	}
	// Use unsafe conversion from a string to a byte slice to prevent copying.
	b := *(*[]byte)(unsafe.Pointer(&x))
	if _, err := w.Write(b); err != nil {
		return FailedWriteError{Op: "WriteString"}
	}
	return nil
}

type BigEndian[W Writer] struct{}

// WriteInt16 ...
func (BigEndian[W]) WriteInt16(w W, x int16) error {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, uint16(x))
	if _, err := w.Write(b); err != nil {
		return FailedWriteError{Op: "WriteInt16"}
	}
	return nil
}

// WriteInt32 ...
func (BigEndian[W]) WriteInt32(w W, x int32) error {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(x))
	if _, err := w.Write(b[:]); err != nil {
		return FailedWriteError{Op: "WriteInt32"}
	}
	return nil
}

// WriteInt64 ...
func (BigEndian[W]) WriteInt64(w W, x int64) error {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(x))
	if _, err := w.Write(b[:]); err != nil {
		return FailedWriteError{Op: "WriteInt64"}
	}
	return nil
}

// WriteFloat32 ...
func (BigEndian[W]) WriteFloat32(w W, x float32) error {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, math.Float32bits(x))
	if _, err := w.Write(b[:]); err != nil {
		return FailedWriteError{Op: "WriteFloat32"}
	}
	return nil
}

// WriteFloat64 ...
func (BigEndian[W]) WriteFloat64(w W, x float64) error {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, math.Float64bits(x))
	if _, err := w.Write(b[:]); err != nil {
		return FailedWriteError{Op: "WriteFloat64"}
	}
	return nil
}

// WriteString ...
func (e BigEndian[W]) WriteString(w W, x string) error {
	if err := e.WriteInt16(w, int16(uint16(len(x)))); err != nil {
		return FailedWriteError{Op: "WriteString"}
	}
	// Use unsafe conversion from a string to a byte slice to prevent copying.
	b := *(*[]byte)(unsafe.Pointer(&x))
	if _, err := w.Write(b); err != nil {
		return FailedWriteError{Op: "WriteString"}
	}
	return nil
}

type NetworkLittleEndian[W Writer] struct{ LittleEndian[W] }

// WriteInt32 ...
func (NetworkLittleEndian[W]) WriteInt32(w W, x int32) error {
	ux := uint32(x) << 1
	if x < 0 {
		ux = ^ux
	}
	for ux >= 0x80 {
		if err := w.WriteByte(byte(ux) | 0x80); err != nil {
			return FailedWriteError{Op: "WriteInt32"}
		}
		ux >>= 7
	}
	if err := w.WriteByte(byte(ux)); err != nil {
		return FailedWriteError{Op: "WriteInt32"}
	}
	return nil
}

// WriteInt64 ...
func (NetworkLittleEndian[W]) WriteInt64(w W, x int64) error {
	ux := uint64(x) << 1
	if x < 0 {
		ux = ^ux
	}
	for ux >= 0x80 {
		if err := w.WriteByte(byte(ux) | 0x80); err != nil {
			return FailedWriteError{Op: "WriteInt64"}
		}
		ux >>= 7
	}
	if err := w.WriteByte(byte(ux)); err != nil {
		return FailedWriteError{Op: "WriteInt64"}
	}
	return nil
}

// WriteString ...
func (NetworkLittleEndian[W]) WriteString(w W, x string) error {
	// Netease
	// if len(x) > maxStringSize {
	// 	return InvalidStringError{Off: w.off, N: uint(len(x)), Err: errStringTooLong}
	// }
	ux := uint32(len(x))
	for ux >= 0x80 {
		if err := w.WriteByte(byte(ux) | 0x80); err != nil {
			return FailedWriteError{Op: "WriteString"}
		}
		ux >>= 7
	}
	if err := w.WriteByte(byte(ux)); err != nil {
		return FailedWriteError{Op: "WriteString"}
	}
	// Use unsafe conversion from a string to a byte slice to prevent copying.
	if _, err := w.Write(*(*[]byte)(unsafe.Pointer(&x))); err != nil {
		return FailedWriteError{Op: "WriteString"}
	}
	return nil
}
