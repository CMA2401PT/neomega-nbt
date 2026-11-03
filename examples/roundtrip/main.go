// 在 neomega-nbt 目录运行：go run ./examples/roundtrip
package main

import (
	"fmt"

	preset "neomega-nbt/presets/v1"
)

func main() {
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
}
