package decoder_snbt

import (
	"errors"
	"fmt"
	nbt "neomega-nbt"
	"neomega-nbt/decoder-snbt/tokens"
	"neomega-nbt/decoder-snbt/tokens/string_token"
	encoder_shim "neomega-nbt/encoder-shim"
)

func DecodeCompound(
	snbt []uint8, encoder nbt.EncoderHighLevel,
	valuePos nbt.ValuePos,
) (
	rest []uint8, // 剩余的不属于此 snbt 部分的内容
	err error,
) {
	if err := encoder.BeginComplexCompound(); err != nil {
		return snbt, err
	}
	var out []uint8
	readKey := func() (string, error) {
		out = out[:0]
		var ok bool
		snbt, ok = tokens.MatchToken(snbt, &string_token.AnyString{
			Reciver: func(b byte) {
				out = append(out, b)
			},
		})
		if !ok {
			return "", fmt.Errorf("missing compound key in: %v", string(snbt))
		}
		return string(out), nil
	}
	var key string
	isFirst := true
	for {
		snbt = consumeWhiteSpace(snbt)
		if len(snbt) == 0 {
			return snbt, errors.New("missing ending }")
		}
		if snbt[0] == '}' {
			snbt = snbt[1:]
			break
		}
		// read key
		key, err = readKey()
		if err != nil {
			return snbt, err
		}
		snbt = consumeWhiteSpace(snbt)
		if len(snbt) == 0 {
			return snbt, fmt.Errorf("missing compound : in: %v", string(snbt))
		}
		if snbt[0] != ':' {
			return snbt, fmt.Errorf("missing compound : in: %v", string(snbt))
		}
		snbt = snbt[1:]
		snbt = consumeWhiteSpace(snbt)
		snbt, err = DecodeSnbtValue(snbt, encoder, func(hlvt nbt.HighLevelValueType) error {
			if !isFirst {
				err = encoder.NextCompoundElement()
				if err != nil {
					return err
				}
			} else {
				isFirst = false
			}
			return encoder.WriteCompoundTag(key, hlvt)
		}, nbt.ValuePosCompoundOrRoot)
		snbt = consumeWhiteSpace(snbt)
		if len(snbt) == 0 {
			return snbt, errors.New("missing ending }")
		}
		if snbt[0] != ',' {
			if snbt[0] == '}' {
				snbt = snbt[1:]
				break
			}
			return snbt, fmt.Errorf("expect a , or } for compound in: %v", string(snbt))
		}
		snbt = snbt[1:]
	}
	return snbt, encoder.EndComplexCompound()
}

type ElementCountEncoderHighLevel struct {
	encoder_shim.VoidEncoderHighLevel
	elementCount int
	level        int
}

func (e *ElementCountEncoderHighLevel) NextListElement(nbt.HighLevelValueType) error {
	if e.level == 1 {
		e.elementCount += 1
	}
	return nil
}

func (e *ElementCountEncoderHighLevel) EndComplexList() error {
	if e.level == 1 {
		e.elementCount += 1
	}
	e.level -= 1
	return nil
}

func (e *ElementCountEncoderHighLevel) BeginComplexList(valueType nbt.HighLevelValueType, len int, typeErase bool) error {
	e.level += 1
	return nil
}

func (e *ElementCountEncoderHighLevel) HasFlag(flag nbt.Flag) bool {
	switch flag {
	case nbt.FLAG_OMIT_LIST_LEN:
		return true
	case nbt.FLAG_SUPPORT_TYPE_ERASE:
		return true
	}
	return false
}

func DecodeList(
	snbt []uint8, encoder nbt.EncoderHighLevel,
) (
	rest []uint8, // 剩余的不属于此 snbt 部分的内容
	err error,
) {
	snbt = consumeWhiteSpace(snbt)
	if len(snbt) == 0 {
		return snbt, errors.New("missing ending ]")
	}
	// special case
	if snbt[0] == ']' {
		snbt = snbt[1:]
		return snbt, encoder.BeginAndEndComplexEmptyList()
	}

	tokenNumber := 0
	if !encoder.HasFlag(nbt.FLAG_OMIT_LIST_LEN) {
		origSnbt := snbt
		counter := &ElementCountEncoderHighLevel{}
		_, err = DecodeList(snbt, counter)
		if err != nil {
			return origSnbt, fmt.Errorf("an error occour when counting elements in list: %v", err)
		}
		tokenNumber = counter.elementCount
		snbt = origSnbt
	}
	isFirst := true
	firstType := nbt.TrivialByte
	typeErase := encoder.HasFlag(nbt.FLAG_SUPPORT_TYPE_ERASE)
	for {
		snbt = consumeWhiteSpace(snbt)
		if len(snbt) == 0 {
			return snbt, errors.New("missing ending ]")
		}
		if snbt[0] == ']' {
			snbt = snbt[1:]
			break
		}
		snbt, err = DecodeSnbtValue(snbt, encoder, func(hlvt nbt.HighLevelValueType) error {
			if !isFirst {
				if !typeErase && hlvt != firstType {
					return fmt.Errorf("type changes from %v -> %v when decoding %v", firstType, hlvt, string(snbt))
				}
				err = encoder.NextListElement(hlvt)
				if err != nil {
					return err
				}
			} else {
				err = encoder.BeginComplexList(hlvt, tokenNumber, typeErase)
				firstType = hlvt
				if err != nil {
					return err
				}
				isFirst = false
			}
			return nil
		}, nbt.ValuePosList)
		if err != nil {
			return snbt, err
		}
		snbt = consumeWhiteSpace(snbt)
		if len(snbt) == 0 {
			return snbt, errors.New("missing ending ]")
		}
		if snbt[0] != ',' {
			if snbt[0] == ']' {
				snbt = snbt[1:]
				break
			}
			return snbt, fmt.Errorf("expect a , or ] for list in: %v", string(snbt))
		}
		snbt = snbt[1:]
	}
	return snbt, encoder.EndComplexList()
}
