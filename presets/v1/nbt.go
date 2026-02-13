package v1

import (
	"errors"
	"io"
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
	EncodingLittleEdian = Encoding(iota + 1)
	EncodingBigEdian
	EncodingNetworkLittleEdian
)

type OptDecodeFromNBT struct {
	Encoding Encoding
	encoder_go_object.EncodeOption
}

type OptDecodeToNBT struct {
	Encoding Encoding
	decoder_go_object.DecodeOption
}

func DecodeFromNBTBytes(data []uint8, ptr any, opt OptDecodeFromNBT) error {
	switch opt.Encoding {
	default:
		return errors.New("must specific encoding")
	case EncodingLittleEdian:
		return decoder_nbt.DecodeFromBytes[[]uint8, decoder_nbt.LittleEndian[*seq_io.ByteReader[[]uint8]]](
			data,
			encoder_shim.NewEncoderShim(
				encoder_go_object.NewGoObjEncoder(ptr,
					opt.EncodeOption,
				),
			),
		)
	case EncodingBigEdian:
		return decoder_nbt.DecodeFromBytes[[]uint8, decoder_nbt.BigEndian[*seq_io.ByteReader[[]uint8]]](
			data,
			encoder_shim.NewEncoderShim(
				encoder_go_object.NewGoObjEncoder(ptr,
					opt.EncodeOption,
				),
			),
		)
	case EncodingNetworkLittleEdian:
		return decoder_nbt.DecodeFromBytes[[]uint8, decoder_nbt.BigEndian[*seq_io.ByteReader[[]uint8]]](
			data,
			encoder_shim.NewEncoderShim(
				encoder_go_object.NewGoObjEncoder(ptr,
					opt.EncodeOption,
				),
			),
		)
	}
}

func EncodeToNBTBytes(data any, opt OptDecodeToNBT) ([]uint8, error) {
	writer := seq_io.NewSliceGrowWriter(nil)
	switch opt.Encoding {
	default:
		return nil, errors.New("must specific encoding")
	case EncodingLittleEdian:
		enc := encoder_nbt.NewEncoder[*seq_io.SliceGrowWriter, encoder_nbt.LittleEndian[*seq_io.SliceGrowWriter]](writer)
		err := decoder_go_object.HighLevelDecodeTo(data, decoder_shim.NewDecoderShim(enc), opt.DecodeOption)
		if err != nil {
			return nil, err
		}
	case EncodingBigEdian:
		enc := encoder_nbt.NewEncoder[*seq_io.SliceGrowWriter, encoder_nbt.BigEndian[*seq_io.SliceGrowWriter]](writer)
		err := decoder_go_object.HighLevelDecodeTo(data, decoder_shim.NewDecoderShim(enc), opt.DecodeOption)
		if err != nil {
			return nil, err
		}
	case EncodingNetworkLittleEdian:
		enc := encoder_nbt.NewEncoder[*seq_io.SliceGrowWriter, encoder_nbt.NetworkLittleEndian[*seq_io.SliceGrowWriter]](writer)
		err := decoder_go_object.HighLevelDecodeTo(data, decoder_shim.NewDecoderShim(enc), opt.DecodeOption)
		if err != nil {
			return nil, err
		}
	}
	return writer.Bytes(), nil
}

func DecodeFromNBT(reader io.Reader, ptr any, opt OptDecodeFromNBT) error {
	switch opt.Encoding {
	default:
		return errors.New("must specific encoding")
	case EncodingLittleEdian:
		return decoder_nbt.Decode[seq_io.OffsetReader, decoder_nbt.LittleEndian[seq_io.OffsetReader]](
			seq_io.NewOffsetReader(reader),
			encoder_shim.NewEncoderShim(
				encoder_go_object.NewGoObjEncoder(ptr,
					opt.EncodeOption,
				),
			),
		)
	case EncodingBigEdian:
		return decoder_nbt.Decode[seq_io.OffsetReader, decoder_nbt.BigEndian[seq_io.OffsetReader]](
			seq_io.NewOffsetReader(reader),
			encoder_shim.NewEncoderShim(
				encoder_go_object.NewGoObjEncoder(ptr,
					opt.EncodeOption,
				),
			),
		)
	case EncodingNetworkLittleEdian:
		return decoder_nbt.Decode[seq_io.OffsetReader, decoder_nbt.NetworklittleEndian[seq_io.OffsetReader]](
			seq_io.NewOffsetReader(reader),
			encoder_shim.NewEncoderShim(
				encoder_go_object.NewGoObjEncoder(ptr,
					opt.EncodeOption,
				),
			),
		)
	}
}

func EncodeToNBT(data any, w io.Writer, opt OptDecodeToNBT) error {
	writer := seq_io.NewOffsetWriter(w)
	switch opt.Encoding {
	default:
		return errors.New("must specific encoding")
	case EncodingLittleEdian:
		enc := encoder_nbt.NewEncoder[seq_io.OffsetWriter, encoder_nbt.LittleEndian[seq_io.OffsetWriter]](writer)
		return decoder_go_object.HighLevelDecodeTo(data, decoder_shim.NewDecoderShim(enc), opt.DecodeOption)
	case EncodingBigEdian:
		enc := encoder_nbt.NewEncoder[seq_io.OffsetWriter, encoder_nbt.BigEndian[seq_io.OffsetWriter]](writer)
		return decoder_go_object.HighLevelDecodeTo(data, decoder_shim.NewDecoderShim(enc), opt.DecodeOption)
	case EncodingNetworkLittleEdian:
		enc := encoder_nbt.NewEncoder[seq_io.OffsetWriter, encoder_nbt.NetworkLittleEndian[seq_io.OffsetWriter]](writer)
		return decoder_go_object.HighLevelDecodeTo(data, decoder_shim.NewDecoderShim(enc), opt.DecodeOption)
	}
}
