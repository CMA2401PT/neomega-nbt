// 在 neomega-nbt 目录运行：go run ./examples/document
package main

import (
	"fmt"

	preset "neomega-nbt/presets/v1"
)

func main() {
	// 1. Document 同时保存根名称和 Compound（按字段名索引的值集合）。
	// slice 默认是 NBT List；Array 把数值 slice 复制为 Go 定长数组，选择 NBT IntArray。
	var original preset.Document = preset.Document{
		Name: "Level",
		Value: preset.Compound{
			"level": int32(3),
			"list":  []int32{1, 2, 3},
			"array": preset.Array([]int32{1, 2, 3}),
		},
	}
	// 本例自己创建输入；实际读取文件时，从文件取得未压缩的 NBT 字节。
	var data []byte
	var err error
	data, err = preset.EncodeDocument(original, preset.EncodingBigEndian)
	if err != nil {
		panic(err)
	}

	// 2. DecodeDocument 要求恰好一个 Compound 根，拒绝尾随字节。
	// 解码后数值 List 与 Array 的 Go 类型仍然不同，重新编码会保留标签区别。
	var document preset.Document
	document, err = preset.DecodeDocument(data, preset.EncodingBigEndian)
	if err != nil {
		panic(err)
	}
	document.Value["level"] = int32(4)

	// 3. 把编辑后的文档写回大端 NBT；根名称由 Document 恢复。
	var updated []byte
	updated, err = preset.EncodeDocument(document, preset.EncodingBigEndian)
	if err != nil {
		panic(err)
	}

	// 4. 再读回写出的字节，展示根名称和内容；SNBT 本身不保存根名称。
	var restored preset.Document
	restored, err = preset.DecodeDocument(updated, preset.EncodingBigEndian)
	if err != nil {
		panic(err)
	}
	fmt.Printf("根名称：%s\n", restored.Name)
	// map 顺序不固定；只在展示时排序，使运行输出便于核对。
	var text string
	text, err = preset.EncodeToSNBT(restored.Value, preset.OptEncodeToSNBT{
		SortMapKey: true,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(text)
}
