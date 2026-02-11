package decoder_snbt

import (
	"fmt"
	nbt "neomega-nbt"
	"neomega-nbt/decoder-snbt/tokens"
	"neomega-nbt/decoder-snbt/tokens/string_token"
)

func DecodeDoubleQString(snbt []uint8, encoder nbt.EncoderHighLevel, valuePos nbt.ValuePos) (
	rest []uint8, // 剩余的不属于此 snbt 部分的内容
	err error,
) {
	var out []uint8
	var ok bool
	snbt, ok = tokens.MatchToken(snbt, &string_token.StringD{
		Reciver: func(b byte) {
			out = append(out, b)
		},
	})
	if ok {
		err = encoder.WriteString(string(out), valuePos)
		return snbt, err
	} else {
		err = fmt.Errorf("not a valid string: %v", string(snbt))
		return snbt, err
	}
}

func DecodeSingleQString(snbt []uint8, encoder nbt.EncoderHighLevel, valuePos nbt.ValuePos) (
	rest []uint8, // 剩余的不属于此 snbt 部分的内容
	err error,
) {
	var out []uint8
	var ok bool
	snbt, ok = tokens.MatchToken(snbt, &string_token.StringS{
		Reciver: func(b byte) {
			out = append(out, b)
		},
	})
	if ok {
		err = encoder.WriteString(string(out), valuePos)
		return snbt, err
	} else {
		err = fmt.Errorf("not a valid string: %v", string(snbt))
		return snbt, err
	}
}
