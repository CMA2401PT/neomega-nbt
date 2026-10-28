// Package v1 provides ready-to-use conversions between Go values, NBT and SNBT.
package v1

import (
	"fmt"
	"io"
	nbt "neomega-nbt"
	decoder_go_object "neomega-nbt/decoder-go-object"
	decoder_nbt "neomega-nbt/decoder-nbt"
	decoder_shim "neomega-nbt/decoder-shim"
	encoder_go_object "neomega-nbt/encoder-go-object"
	encoder_nbt "neomega-nbt/encoder-nbt"
	encoder_shim "neomega-nbt/encoder-shim"
	seq_io "neomega-nbt/seq-io"
)

type Encoding uint8

const (
	EncodingLittleEndian Encoding = iota + 1
	EncodingBigEndian
	EncodingNetworkLittleEndian
)

// Retain the original spellings for existing callers.
const (
	EncodingLittleEdian        = EncodingLittleEndian
	EncodingBigEdian           = EncodingBigEndian
	EncodingNetworkLittleEdian = EncodingNetworkLittleEndian
)

type OptDecodeFromNBT struct {
	Encoding Encoding
	encoder_go_object.EncodeOption
}

type OptEncodeToNBT struct {
	Encoding Encoding
	decoder_go_object.DecodeOption
}

// OptDecodeToNBT is the original name of OptEncodeToNBT.
type OptDecodeToNBT = OptEncodeToNBT

func decodeNBT[R seq_io.Reader](reader R, encoder nbt.Encoder, encoding Encoding) error {
	switch encoding {
	case EncodingLittleEndian:
		return decoder_nbt.Decode[R, decoder_nbt.LittleEndian[R]](reader, encoder)
	case EncodingBigEndian:
		return decoder_nbt.Decode[R, decoder_nbt.BigEndian[R]](reader, encoder)
	case EncodingNetworkLittleEndian:
		return decoder_nbt.Decode[R, decoder_nbt.NetworklittleEndian[R]](reader, encoder)
	default:
		return fmt.Errorf("unsupported NBT encoding: %d", encoding)
	}
}

func newNBTEncoder[W seq_io.Writer](writer W, encoding Encoding) (nbt.Encoder, error) {
	switch encoding {
	case EncodingLittleEndian:
		return encoder_nbt.NewEncoder[W, encoder_nbt.LittleEndian[W]](writer), nil
	case EncodingBigEndian:
		return encoder_nbt.NewEncoder[W, encoder_nbt.BigEndian[W]](writer), nil
	case EncodingNetworkLittleEndian:
		return encoder_nbt.NewEncoder[W, encoder_nbt.NetworkLittleEndian[W]](writer), nil
	default:
		return nil, fmt.Errorf("unsupported NBT encoding: %d", encoding)
	}
}

// DecodeFromNBTBytes reads one NBT root into ptr. The root name is ignored.
func DecodeFromNBTBytes(data []byte, ptr any, opt OptDecodeFromNBT) error {
	return decodeNBT(seq_io.NewByteReaderFromBytes(data), encoder_shim.NewEncoderShim(
		encoder_go_object.NewGoObjEncoder(ptr, opt.EncodeOption),
	), opt.Encoding)
}

func EncodeToNBTBytes(data any, opt OptEncodeToNBT) ([]byte, error) {
	writer := seq_io.NewSliceGrowWriter(nil)
	if err := encodeToNBT(data, writer, opt); err != nil {
		return nil, err
	}
	return writer.Bytes(), nil
}

// DecodeFromNBT reads one NBT root, leaving following data in reader.
func DecodeFromNBT(reader io.Reader, ptr any, opt OptDecodeFromNBT) error {
	return decodeNBT(seq_io.NewOffsetReader(reader), encoder_shim.NewEncoderShim(
		encoder_go_object.NewGoObjEncoder(ptr, opt.EncodeOption),
	), opt.Encoding)
}

func EncodeToNBT(data any, writer io.Writer, opt OptEncodeToNBT) error {
	return encodeToNBT(data, seq_io.NewOffsetWriter(writer), opt)
}

func encodeToNBT[W seq_io.Writer](data any, writer W, opt OptEncodeToNBT) error {
	encoder, err := newNBTEncoder(writer, opt.Encoding)
	if err != nil {
		return err
	}
	return decoder_go_object.HighLevelDecodeTo(data, decoder_shim.NewDecoderShim(encoder), opt.DecodeOption)
}

// ConvertNBT changes the binary encoding without constructing an intermediate Go object.
// It preserves root names, tag types and compound order.
func ConvertNBT(data []byte, from, to Encoding) ([]byte, error) {
	writer := seq_io.NewSliceGrowWriter(nil)
	encoder, err := newNBTEncoder(writer, to)
	if err != nil {
		return nil, err
	}
	if err := decodeNBT(seq_io.NewByteReaderFromBytes(data), encoder, from); err != nil {
		return nil, err
	}
	return writer.Bytes(), nil
}
