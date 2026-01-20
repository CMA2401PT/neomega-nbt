package encoder_shim

import (
	"fmt"
	nbt "neomega-nbt"
)

type EncodeState uint8

type EncodeStateWithLen struct {
	EncodeState
	CoumpoundElementCount int
	Aux                   uint32
}

type EncoderShim[HL nbt.EncoderHighLevel] struct {
	EncodeStateWithLen
	Stack            []EncodeStateWithLen
	HighLevelEncoder HL
}

func (fa *EncoderShim[HL]) Done() bool {
	if len(fa.Stack) == 0 && fa.EncodeState == EncodeStateEnd {
		return true
	}
	return false
}

func (fa *EncoderShim[HL]) Finish() error {
	if !fa.Done() {
		return fa.UnexpectError()
	}
	return fa.HighLevelEncoder.Finish()
}

func (fa *EncoderShim[HL]) IsRoot() bool {
	return len(fa.Stack) == 0
}

func (fa *EncoderShim[HL]) Pop() error {
	if len(fa.Stack) == 0 {
		fa.EncodeState = EncodeStateEnd
		return fa.HighLevelEncoder.Finish()
	}

	fa.EncodeStateWithLen = fa.Stack[len(fa.Stack)-1]
	fa.Stack = fa.Stack[:len(fa.Stack)-1]

	if fa.EncodeState&EncodeStateList != EncodeStateList {
		return nil
	}
	// is list
	fa.Aux -= 1
	if fa.Aux == 0 {
		// list finish
		if err := fa.HighLevelEncoder.EndComplexList(); err != nil {
			return err
		}
		if err := fa.Pop(); err != nil {
			return err
		}
	} else {
		// next list element
		elementStateType := fa.UnBoxListElementType()
		if err := fa.HighLevelEncoder.NextListElement(UnmapToBasicValueType(elementStateType)); err != nil {
			return err
		}

		fa.Push()
		fa.EncodeState = elementStateType
	}
	return nil
}

func (fa *EncoderShim[HL]) Push() {
	fa.Stack = append(fa.Stack, fa.EncodeStateWithLen)
	fa.EncodeStateWithLen = EncodeStateWithLen{0, 0, 0}
}

func (fa *EncoderShim[HL]) Expect() string {
	if fa.Done() {
		return "done"
	}
	return StateExpectCommon(fa.EncodeState, fa.UnBoxListElementType(), fa.Aux)
}

func (fa *EncoderShim[HL]) UnexpectError() error {
	return fmt.Errorf("expect value: %v", fa.Expect())
}

func (fa *EncoderShim[HL]) UnBoxListElementType() EncodeState {
	if fa.EncodeState == EncodeStateListTypeForList|EncodeStateList {
		return EncodeStateList
	}
	return fa.EncodeState & EncodeState(^uint8(EncodeStateList))
}

func (fa *EncoderShim[HL]) WriteType(tagType nbt.TagType) error {
	if fa.EncodeState == EncodeStateList {
		// is list (list element type)
		boxedState := BoxListElementTagType(tagType)
		if boxedState == EncodeStateList {
			return fmt.Errorf("%v is not a valid tag type for list element", tagType)
		}
		fa.EncodeState = boxedState
		return nil
	}
	if fa.EncodeState == EncodeStateCompound {
		if fa.CoumpoundElementCount != 0 {
			if err := fa.HighLevelEncoder.NextCompoundElement(); err != nil {
				return err
			}
		}
		fa.CoumpoundElementCount += 1
		fa.Push()
	}
	if fa.EncodeState != EncodeStateBegin {
		return fa.UnexpectError()
	}
	fa.EncodeState = EncodeStateTag
	fa.Aux = StashTagType(tagType)
	if fa.Aux == 0 {
		return fmt.Errorf("not valid tag type: %v", tagType)
	}
	return nil
}

func (fa *EncoderShim[HL]) WriteTag(tag string) error {
	if fa.EncodeState != EncodeStateTag {
		return fa.UnexpectError()
	}
	fa.EncodeState = EncodeState(fa.Aux)
	fa.Aux = 0

	if fa.EncodeState&EncodeStateList == EncodeStateList {
		// list, postpone until length is told
	}
	if fa.EncodeState&EncodeStateNumberArray == EncodeStateNumberArray {
		// array, postpone until length is told
	}
	if fa.IsRoot() {
		if err := fa.HighLevelEncoder.WriteRootTag(tag, UnmapToBasicValueType(fa.EncodeState)); err != nil {
			return err
		}
	} else {

		if err := fa.HighLevelEncoder.WriteCompoundTag(tag, UnmapToBasicValueType(fa.EncodeState)); err != nil {
			return err
		}
	}
	if fa.EncodeState == EncodeStateCompound {
		if err := fa.HighLevelEncoder.BeginComplexCompound(); err != nil {
			return err
		}
	}
	return nil
}

func (fa *EncoderShim[HL]) WriteEmptyListMark() error {
	if fa.EncodeState&EncodeStateList == EncodeStateList {
		if fa.Aux != 0 || fa.EncodeState != EncodeStateList {
			return fa.UnexpectError()
		}
		if err := fa.HighLevelEncoder.BeginAndEndComplexEmptyList(); err != nil {
			return err
		}
		if err := fa.Pop(); err != nil {
			return err
		}
		return nil
	}
	return fa.UnexpectError()
}

func (fa *EncoderShim[HL]) WriteLen(len uint32) error {
	if fa.EncodeState&EncodeStateList == EncodeStateList {
		if fa.Aux != 0 || fa.EncodeState == EncodeStateList {
			return fa.UnexpectError()
		}
		fa.Aux = len
		elementStateType := fa.UnBoxListElementType()
		if err := fa.HighLevelEncoder.BeginComplexList(UnmapToBasicValueType(elementStateType), int(len), false); err != nil {
			return err
		}
		if len == 0 {
			if err := fa.HighLevelEncoder.EndComplexList(); err != nil {
				return err
			}
			if err := fa.Pop(); err != nil {
				return err
			}

		} else {
			fa.Push()
			fa.EncodeState = elementStateType
		}
		return nil
	}
	if fa.EncodeState&EncodeStateNumberArray == EncodeStateNumberArray {
		if fa.Aux != 0 {
			return fa.UnexpectError()
		}
		switch fa.EncodeState {
		case EncodeStateByteArray:
			if err := fa.HighLevelEncoder.BeginComplexArray(nbt.TrivialByte, int(len)); err != nil {
				return err
			}
		case EncodeStateIntArray:
			if err := fa.HighLevelEncoder.BeginComplexArray(nbt.TrivialInt, int(len)); err != nil {
				return err
			}
		case EncodeStateLongArray:
			if err := fa.HighLevelEncoder.BeginComplexArray(nbt.TrivialLong, int(len)); err != nil {
				return err
			}
		}
		fa.Aux = len
		if len == 0 {
			if err := fa.HighLevelEncoder.EndComplexArray(); err != nil {
				return err
			}
			if err := fa.Pop(); err != nil {
				return err
			}
		}
		return nil
	}

	return fa.UnexpectError()
}

func (fa *EncoderShim[HL]) WriteEnd() error {
	if fa.EncodeState != EncodeStateCompound {
		return fa.UnexpectError()
	}
	if err := fa.HighLevelEncoder.EndComplexCompound(); err != nil {
		return err
	}
	if err := fa.Pop(); err != nil {
		return err
	}
	return nil
}

func (fa *EncoderShim[HL]) CanWriteNumberValue(expect EncodeState) bool {
	return fa.EncodeState^EncodeStateNumberArray == expect || fa.EncodeState == expect
}

func (fa *EncoderShim[HL]) ValuePos() nbt.ValuePos {
	if fa.Aux > 0 {
		if fa.EncodeState&EncodeStateNumberArray == EncodeStateNumberArray {
			return nbt.ValuePosArray
		} else {
			return nbt.ValuePosList
		}
	}
	return nbt.ValuePosCompoundOrRoot
}

func (fa *EncoderShim[HL]) WriteNumberValue(expect EncodeState) error {
	if fa.EncodeState^EncodeStateNumberArray == expect {
		fa.Aux -= 1
		if fa.Aux == 0 {
			if err := fa.HighLevelEncoder.EndComplexArray(); err != nil {
				return err
			}
			if err := fa.Pop(); err != nil {
				return err
			}
		} else {
			if err := fa.HighLevelEncoder.NextArrayElement(); err != nil {
				return err
			}
		}
		return nil
	}
	if fa.EncodeState == expect {
		if err := fa.Pop(); err != nil {
			return err
		}
		return nil
	}
	return fa.UnexpectError()
}

func (fa *EncoderShim[HL]) WriteNormalValue(expect EncodeState) error {
	if fa.EncodeState == expect {
		if err := fa.Pop(); err != nil {
			return err
		}
		return nil
	}
	return fa.UnexpectError()
}
