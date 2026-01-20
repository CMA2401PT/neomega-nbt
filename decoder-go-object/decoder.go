package decoder_go_object

import (
	nbt "neomega-nbt"
)

type DecodeOption struct {
	// 根标签，但是其实大多数情况下这个会被忽略或者留空
	RootTag string
	// SliceAsArray=false 时，
	// 同理包括 []uint8 []int8 []uint32 []int32 []uint64 []int64
	// SliceAsArray=true 时，
	// []int32 等会被当作对应类型的 array 处理
	// 特别的，对于[]any
	// 当其长度为 0 时，报告为空列表，否则报告为第一个元素的类型
	// 因为 []any 可能具有不同实际类型
	// 所以从第二个元素开始（第一个元素通过 BeginComplexList 报告类型）
	// 通过 NextListElement 报告类型
	SliceAsArray bool
	// SortMapKey=false 时，map 的键值是随机访问的，否则会按字母序先排序
	SortMapKey bool
	// 一般为 "nbt" 即当存在这个标签时，使用标签指定的名字
	TagName string
}

func HighLevelDecodeTo(
	orig any, encoder nbt.EncoderHighLevel, opt DecodeOption,
) error {
	if opt.TagName == "" {
		opt.TagName = "nbt"
	}

	getChildDecoder := WrapGetDecoder(opt)
	decoder, err := getChildDecoder(orig)
	if err != nil {
		return err
	}
	if err := encoder.WriteRootTag(opt.RootTag, decoder.GetNBType()); err != nil {
		return err
	}
	return decoder.DecodeTo(nbt.ValuePosCompoundOrRoot, encoder, getChildDecoder)
}

func WrapGetDecoder(opt DecodeOption) GetDecoderFunc {
	return func(orig any) (o DecoderOverride, err error) {
		if override, ok := orig.(DecoderOverride); ok {
			return override, err
		}
		return GetDefaultDecoder(orig, opt)
	}
}
