package v1_test

import (
	"fmt"
	nbt "neomega-nbt"
	decoder_go "neomega-nbt/decoder-go-object"
	encoder_go "neomega-nbt/encoder-go-object"
	preset "neomega-nbt/presets/v1"
)

func Example() {
	type Block struct {
		Name  string `nbt:"name"`
		Level int32  `nbt:"level"`
	}
	input := Block{Name: "stone", Level: 3}
	data, err := preset.EncodeToNBTBytes(input, preset.OptEncodeToNBT{Encoding: preset.EncodingLittleEndian})
	if err != nil {
		panic(err)
	}
	var output Block
	if err := preset.DecodeFromNBTBytes(data, &output, preset.OptDecodeFromNBT{Encoding: preset.EncodingLittleEndian}); err != nil {
		panic(err)
	}
	text, err := preset.EncodeToSNBT(output, preset.OptEncodeToSNBT{})
	if err != nil {
		panic(err)
	}
	fmt.Println(text)
	// Output: {"name": "stone", "level": 3}
}

// Distance is measured in blocks in Go, and stored as integer milliblocks in NBT.
type Distance float64

func (Distance) GetNBType() nbt.HighLevelValueType { return nbt.TrivialInt }
func (d Distance) DecodeTo(pos nbt.ValuePos, encoder nbt.EncoderHighLevel, _ decoder_go.GetDecoderFunc) error {
	return encoder.WriteInt32(int32(d*1000), pos)
}
func (d *Distance) GetReciver(valueType nbt.HighLevelValueType) (encoder_go.ValueReciver, error) {
	if valueType != nbt.TrivialInt {
		return encoder_go.ValueReciver{}, fmt.Errorf("Distance expects Int, got %v", valueType)
	}
	return encoder_go.ValueReciver{TrivialValReciver: func(_ any, value int64) error {
		*d = Distance(value) / 1000
		return nil
	}}, nil
}

func Example_customValue() {
	text, err := preset.EncodeToSNBT(Distance(1.25), preset.OptEncodeToSNBT{})
	if err != nil {
		panic(err)
	}
	var distance Distance
	if err := preset.DecodeFromSNBT(text, &distance, preset.OptDecodeFromSNBT{}); err != nil {
		panic(err)
	}
	fmt.Println(text, distance)
	// Output: 1250 1.25
}
