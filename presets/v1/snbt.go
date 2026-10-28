package v1

import (
	"fmt"
	nbt "neomega-nbt"
	decoder_go_object "neomega-nbt/decoder-go-object"
	decoder_shim "neomega-nbt/decoder-shim"
	decoder_snbt "neomega-nbt/decoder-snbt"
	encoder_go_object "neomega-nbt/encoder-go-object"
	encoder_shim "neomega-nbt/encoder-shim"
	encoder_snbt "neomega-nbt/encoder-snbt"
	seq_io "neomega-nbt/seq-io"
	"strings"
)

// Options describe Go values on the source and destination sides respectively.
type OptEncodeToSNBT = decoder_go_object.DecodeOption
type OptDecodeFromSNBT = encoder_go_object.EncodeOption

type OptSNBTToNBT struct {
	Encoding Encoding
	RootTag  string
}

// DecodeFromSNBT requires exactly one value, optionally surrounded by whitespace.
func DecodeFromSNBT(data string, ptr any, opt OptDecodeFromSNBT) error {
	return decodeSNBT(data, encoder_go_object.NewGoObjEncoder(ptr, opt), "")
}

func EncodeToSNBT(data any, opt OptEncodeToSNBT) (string, error) {
	encoder, result := encoder_snbt.NewEncoderToBytes(nil)
	if err := decoder_go_object.HighLevelDecodeTo(data, encoder, opt); err != nil {
		return "", err
	}
	return string(result()), nil
}

// NBTToSNBT omits the root name because SNBT has no root-name field.
func NBTToSNBT(data []byte, encoding Encoding) (string, error) {
	encoder, result := encoder_snbt.NewEncoderToBytes(nil)
	if err := decodeNBT(seq_io.NewByteReaderFromBytes(data), encoder_shim.NewEncoderShim(encoder), encoding); err != nil {
		return "", err
	}
	return string(result()), nil
}

// SNBTToNBT converts directly through events. Binary NBT requires homogeneous lists.
func SNBTToNBT(data string, opt OptSNBTToNBT) ([]byte, error) {
	writer := seq_io.NewSliceGrowWriter(nil)
	encoder, err := newNBTEncoder(writer, opt.Encoding)
	if err != nil {
		return nil, err
	}
	if err := decodeSNBT(data, decoder_shim.NewDecoderShim(encoder), opt.RootTag); err != nil {
		return nil, err
	}
	return writer.Bytes(), nil
}

func decodeSNBT(data string, encoder nbt.EncoderHighLevel, rootTag string) error {
	rest, err := decoder_snbt.DecodeSnbtTo([]byte(data), encoder, decoder_snbt.DecodeOption{RootTag: rootTag})
	if err != nil {
		return err
	}
	if strings.TrimSpace(string(rest)) != "" {
		return fmt.Errorf("unexpected trailing SNBT: %q", rest)
	}
	return nil
}
