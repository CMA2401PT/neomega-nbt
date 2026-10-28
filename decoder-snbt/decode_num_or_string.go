package decoder_snbt

import (
	"fmt"
	nbt "neomega-nbt"
	"neomega-nbt/decoder-snbt/tokens"
	"neomega-nbt/decoder-snbt/tokens/number"
	"neomega-nbt/decoder-snbt/tokens/string_token"
)

func DecodeNumOrString(
	snbt []uint8, encoder nbt.EncoderHighLevel,
	onKnownValuetype func(nbt.HighLevelValueType) error,
	valuePos nbt.ValuePos,
) (
	rest []uint8, // 剩余的不属于此 snbt 部分的内容
	err error,
) {
	origSnbt := snbt
	number := &number.NumberFinity{}
	var ok bool
	snbt, ok = tokens.MatchToken(snbt, number)
	if ok {
		// 尝试继续作为一个 string 读取
		_, ok = tokens.MatchToken(snbt, &string_token.AnyString{
			Reciver: func(b byte) {},
		})
		if !ok {
			// 是数字
			switch number.DataType {
			case 0:
				if err := onKnownValuetype(nbt.TrivialInt); err != nil {
					return snbt, err
				}
				if err := encoder.WriteInt32(number.Val().(int32), valuePos); err != nil {
					return snbt, err
				}
			case 1:
				if err := onKnownValuetype(nbt.TrivialByte); err != nil {
					return snbt, err
				}
				if err := encoder.WriteInt8(number.Val().(int8), valuePos); err != nil {
					return snbt, err
				}
			case 2:
				if err := onKnownValuetype(nbt.TrivialShort); err != nil {
					return snbt, err
				}
				if err := encoder.WriteInt16(number.Val().(int16), valuePos); err != nil {
					return snbt, err
				}
			case 3:
				if err := onKnownValuetype(nbt.TrivialLong); err != nil {
					return snbt, err
				}
				if err := encoder.WriteInt64(number.Val().(int64), valuePos); err != nil {
					return snbt, err
				}
			case 4:
				if err := onKnownValuetype(nbt.TrivialFloat); err != nil {
					return snbt, err
				}
				if err := encoder.WriteFloat32(number.Val().(float32), valuePos); err != nil {
					return snbt, err
				}
			case 5:
				if err := onKnownValuetype(nbt.TrivialDouble); err != nil {
					return snbt, err
				}
				if err := encoder.WriteFloat64(number.Val().(float64), valuePos); err != nil {
					return snbt, err
				}
			default:
				panic(fmt.Errorf("should be a know type, but get: %v", number.Val()))
			}
			return snbt, nil
		}
	}

	// 当确定是一个没有包裹的 string 时:
	// 先恢复原样
	snbt = origSnbt
	var out []uint8
	if err := onKnownValuetype(nbt.TrivialString); err != nil {
		return snbt, err
	}
	snbt, ok = tokens.MatchToken(snbt, &string_token.AnyString{
		Reciver: func(b byte) {
			out = append(out, b)
		},
	})
	if ok {
		err = encoder.WriteString(string(out), valuePos)
		return snbt, err
	} else {
		err = fmt.Errorf("not a valid string or number: %v", string(snbt))
		return snbt, err
	}
}
