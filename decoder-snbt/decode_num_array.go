package decoder_snbt

import (
	"fmt"
	nbt "neomega-nbt"
	"neomega-nbt/decoder-snbt/tokens"
	"neomega-nbt/decoder-snbt/tokens/num_arr"
)

func DecodeNumArray(
	snbt []uint8,
	encoder nbt.EncoderHighLevel,
	elemType nbt.HighLevelValueType,
) (
	rest []uint8, // 剩余的不属于此 snbt 部分的内容
	err error,
) {
	arrayLen := 0
	if !encoder.HasFlag(nbt.FLAG_OMIT_ARRAY_LEN) {
		counter := 0
		if _, ok := tokens.MatchToken(snbt, &num_arr.NumArray[int32]{Reciver: func(i int32) error {
			counter += 1
			return nil
		}}); !ok {
			err = fmt.Errorf("not a valid array: %v", string(snbt))
			return snbt, err
		}
	}
	err = encoder.BeginComplexArray(elemType, arrayLen)
	if err != nil {
		return snbt, err
	}
	var ok bool

	switch elemType {
	case nbt.TrivialByte:
		isFirst := true
		if snbt, ok = tokens.MatchToken(snbt, &num_arr.NumArray[int8]{Reciver: func(i int8) (err error) {
			if isFirst {
				isFirst = false
				return encoder.WriteInt8(i, nbt.ValuePosArray)
			} else {
				err := encoder.NextArrayElement()
				if err != nil {
					return err
				}
				return encoder.WriteInt8(i, nbt.ValuePosArray)
			}
		}}); !ok {
			err = fmt.Errorf("not a valid array: %v", string(snbt))
			return snbt, err
		}
	case nbt.TrivialInt:
		isFirst := true
		if snbt, ok = tokens.MatchToken(snbt, &num_arr.NumArray[int32]{Reciver: func(i int32) (err error) {
			if isFirst {
				isFirst = false
				return encoder.WriteInt32(i, nbt.ValuePosArray)
			} else {
				err := encoder.NextArrayElement()
				if err != nil {
					return err
				}
				return encoder.WriteInt32(i, nbt.ValuePosArray)
			}
		}}); !ok {
			err = fmt.Errorf("not a valid array: %v", string(snbt))
			return snbt, err
		}
	case nbt.TrivialLong:
		isFirst := true
		if snbt, ok = tokens.MatchToken(snbt, &num_arr.NumArray[int64]{Reciver: func(i int64) (err error) {
			if isFirst {
				isFirst = false
				return encoder.WriteInt64(i, nbt.ValuePosArray)
			} else {
				err := encoder.NextArrayElement()
				if err != nil {
					return err
				}
				return encoder.WriteInt64(i, nbt.ValuePosArray)
			}
		}}); !ok {
			err = fmt.Errorf("not a valid array: %v", string(snbt))
			return snbt, err
		}
	default:
		panic("array can only be byte/int/long")
	}
	err = encoder.EndComplexArray()
	return snbt, err
}
