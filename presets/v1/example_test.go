package v1_test

import (
	"fmt"
	nbt "neomega-nbt"
	decoder_go "neomega-nbt/decoder-go-object"
	encoder_go "neomega-nbt/encoder-go-object"
	preset "neomega-nbt/presets/v1"
)

func Example() {
	// 1. 用结构体约束字段类型；int32 对应 NBT Int，nbt 标签决定字段名。
	type Block struct {
		Name  string `nbt:"name"`
		Level int32  `nbt:"level"`
	}
	var input Block = Block{Name: "stone", Level: 3}

	// 2. 本例约定输入输出都是小端；实际使用时应按文件或协议选择编码。
	// 返回的是未压缩的单个 NBT 根，不包含文件头。
	var data []byte
	var err error
	data, err = preset.EncodeToNBTBytes(input, preset.OptEncodeToNBT{
		Encoding: preset.EncodingLittleEndian,
	})
	if err != nil {
		panic(err)
	}

	// 3. 接收目标必须传指针，库会填充它指向的字段。
	var output Block
	err = preset.DecodeFromNBTBytes(data, &output, preset.OptDecodeFromNBT{
		Encoding: preset.EncodingLittleEndian,
	})
	if err != nil {
		panic(err)
	}

	// 4. 默认选项足够显示这个结构体；SNBT 是便于阅读的文本格式。
	var text string
	text, err = preset.EncodeToSNBT(output, preset.OptEncodeToSNBT{})
	if err != nil {
		panic(err)
	}
	fmt.Println(text)
	// Output: {"name": "stone", "level": 3}
}

// Distance 在 Go 中以格为单位，在 NBT 中以千分之一格的 Int 保存。
type Distance float64

const milliblocksPerBlock = 1000

// GetNBType 声明此业务值传递给生成端时使用 NBT Int。
func (Distance) GetNBType() nbt.HighLevelValueType { return nbt.TrivialInt }

// DecodeTo 从 Go 距离读出千分之一格数值；decoder 的方向是读取 Go 源值。
func (d Distance) DecodeTo(pos nbt.ValuePos, encoder nbt.EncoderHighLevel, _ decoder_go.GetDecoderFunc) error {
	return encoder.WriteInt32(int32(d*milliblocksPerBlock), pos)
}

// GetReciver 为目标指针提供接收方法；沿用公共接口的 Reciver 拼写。
func (d *Distance) GetReciver(valueType nbt.HighLevelValueType) (encoder_go.ValueReciver, error) {
	if valueType != nbt.TrivialInt {
		return encoder_go.ValueReciver{}, fmt.Errorf("距离要求 Int，实际为 %v", valueType)
	}
	// 整数由回调的第二个参数以 int64 交付，再换回格数。
	return encoder_go.ValueReciver{TrivialValReciver: func(_ any, value int64) error {
		*d = Distance(value) / milliblocksPerBlock
		return nil
	}}, nil
}

func Example_customValue() {
	// 1. 源值的 override 把 1.25 格输出成 Int 1250。
	var text string
	var err error
	text, err = preset.EncodeToSNBT(Distance(1.25), preset.OptEncodeToSNBT{})
	if err != nil {
		panic(err)
	}
	// 2. 目标指针的 override 接收 Int，并恢复成格。
	var distance Distance
	err = preset.DecodeFromSNBT(text, &distance, preset.OptDecodeFromSNBT{})
	if err != nil {
		panic(err)
	}
	fmt.Println(text, distance)
	// Output: 1250 1.25
}
