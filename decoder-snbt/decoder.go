package decoder_snbt

import (
	"errors"
	nbt "neomega-nbt"
)

type DecodeOption struct {
	// 根标签，但是其实大多数情况下这个会被忽略或者留空
	RootTag string
	// 是否允许为空 (无论是否允许，都不会写入root tag)
	AllowEmpty bool
}

func DecodeSnbtTo(
	snbt []uint8,
	encoder nbt.EncoderHighLevel,
	opt DecodeOption,
) (
	rest []uint8, // 剩余的不属于此 snbt 部分的内容
	err error,
) {
	if emptyCheck(snbt) {
		if !opt.AllowEmpty {
			return snbt, errors.New("empty snbt")
		} else {
			return snbt, nil
		}
	}
	return DecodeSnbtValue(snbt, encoder, func(hlvt nbt.HighLevelValueType) error {
		// 当知道元素类型时，写入根标签
		return encoder.WriteRootTag(opt.RootTag, hlvt)
	}, nbt.ValuePosCompoundOrRoot)
}

func DecodeSnbtValue(
	snbt []uint8,
	encoder nbt.EncoderHighLevel,
	// 当知道这个元素的类型时，向上层报告类型
	onKnownValuetype func(nbt.HighLevelValueType) error,
	valuePos nbt.ValuePos,
) (
	rest []uint8, // 剩余的不属于此 snbt 部分的内容
	err error,
) {
	snbt, leftType := leftBracket(snbt)
	switch leftType {
	case ltByteArr: // [B;
		if err := onKnownValuetype(nbt.ComplexByteArray); err != nil {
			return snbt, err
		}
		return DecodeNumArray(snbt, encoder, nbt.TrivialByte)
	case ltIntArr: // [I;
		if err := onKnownValuetype(nbt.ComplexIntArray); err != nil {
			return snbt, err
		}
		return DecodeNumArray(snbt, encoder, nbt.TrivialInt)
	case ltLongArr: // [L;
		if err := onKnownValuetype(nbt.ComplexLongArray); err != nil {
			return snbt, err
		}
		return DecodeNumArray(snbt, encoder, nbt.TrivialLong)
	case ltCompound: // {
		if err := onKnownValuetype(nbt.ComplexCompound); err != nil {
			return snbt, err
		}
		return DecodeCompound(snbt, encoder, valuePos)
	case ltList: // [
		if err := onKnownValuetype(nbt.ComplexList); err != nil {
			return snbt, err
		}
		return DecodeList(snbt, encoder)
	case ltDoubleQString: // "
		if err := onKnownValuetype(nbt.TrivialString); err != nil {
			return snbt, err
		}
		return DecodeDoubleQString(snbt, encoder, valuePos)
	case ltSingleQString: // '
		if err := onKnownValuetype(nbt.TrivialString); err != nil {
			return snbt, err
		}
		return DecodeSingleQString(snbt, encoder, valuePos)
	default:
		// strings or number
		return DecodeNumOrString(snbt, encoder, onKnownValuetype, valuePos)
	}
}
