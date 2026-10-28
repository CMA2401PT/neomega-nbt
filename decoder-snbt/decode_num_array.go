package decoder_snbt

import (
	"fmt"
	nbt "neomega-nbt"
	"neomega-nbt/decoder-snbt/tokens"
	"neomega-nbt/decoder-snbt/tokens/num_arr"
)

func DecodeNumArray(snbt []byte, encoder nbt.EncoderHighLevel, elemType nbt.HighLevelValueType) ([]byte, error) {
	snbt = consumeWhiteSpace(snbt)
	if len(snbt) > 0 && snbt[0] == ']' {
		return snbt[1:], encoder.BeginAndEndComplexEmptyArray(elemType)
	}
	arrayLen := 0
	if !encoder.HasFlag(nbt.FLAG_OMIT_ARRAY_LEN) {
		_, err := parseNumArray(snbt, func(int64) error { arrayLen++; return nil })
		if err != nil {
			return snbt, err
		}
	}
	if err := encoder.BeginComplexArray(elemType, arrayLen); err != nil {
		return snbt, err
	}
	first := true
	beforeValue := func() error {
		if first {
			first = false
			return nil
		}
		return encoder.NextArrayElement()
	}
	var rest []byte
	var err error
	switch elemType {
	case nbt.TrivialByte:
		rest, err = parseNumArray(snbt, func(v int8) error {
			if err := beforeValue(); err != nil {
				return err
			}
			return encoder.WriteInt8(v, nbt.ValuePosArray)
		})
	case nbt.TrivialInt:
		rest, err = parseNumArray(snbt, func(v int32) error {
			if err := beforeValue(); err != nil {
				return err
			}
			return encoder.WriteInt32(v, nbt.ValuePosArray)
		})
	case nbt.TrivialLong:
		rest, err = parseNumArray(snbt, func(v int64) error {
			if err := beforeValue(); err != nil {
				return err
			}
			return encoder.WriteInt64(v, nbt.ValuePosArray)
		})
	default:
		return snbt, fmt.Errorf("invalid array element type: %v", elemType)
	}
	if err != nil {
		return rest, err
	}
	return rest, encoder.EndComplexArray()
}

// Token matching reports only success/failure; preserve receiver errors separately.
func parseNumArray[I int8 | int32 | int64](snbt []byte, receive func(I) error) ([]byte, error) {
	var receiverErr error
	rest, ok := tokens.MatchToken(snbt, &num_arr.NumArray[I]{Reciver: func(v I) error {
		receiverErr = receive(v)
		return receiverErr
	}})
	if receiverErr != nil {
		return rest, receiverErr
	}
	if !ok {
		return rest, fmt.Errorf("not a valid array: %s", snbt)
	}
	return rest, nil
}
