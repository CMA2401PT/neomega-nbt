package encoder_go_object

import (
	"fmt"
	nbt "neomega-nbt"
	"reflect"
)

type Current struct {
	PtrToData any
	ValueReciver
	aux Aux
}

type Aux struct {
	emptyCompound          bool
	lastListElemType       nbt.HighLevelValueType
	lastListValueReciver   ValueReciver
	parentListConstructor  ListConstructor
	parentArrayConstructor ArrayConstructor
}

type EncodeOption struct {
	// nbt 键标签名，一般为 "nbt"，留空自动设置为 "nbt"
	TagName string
	// 当没有明确指定类型，例如 &any 时，遇到 List[Byte] 和 ByteArray 时
	// 构造为 []int8, [...]int8 而不是  []uint8, [...]uint8
	UseInt8Sequence bool
	// ArrayAsSlice
	// 当没有明确指定类型时，例如 &any 时
	// 如果遇到 ByteArray，IntArray 和 LongArray 时
	// 构造为 []int8/[]uint8 []int32 []int64 而不是 [...]int8/[...]uint8 [...]int32 [...]int64
	ArrayAsSlice bool
}

type GoObjEncoder struct {
	Stack   []Current
	Current Current
	EncodeOption
}

func NewGoObjEncoder(ptrToReciver any, opt EncodeOption) nbt.EncoderHighLevel {
	return &GoObjEncoder{
		Current: Current{
			PtrToData: ptrToReciver,
		},
		EncodeOption: opt,
	}
}

func (e *GoObjEncoder) WriteRootTag(tag string, valueType nbt.HighLevelValueType) (err error) {
	if e.TagName == "" {
		e.TagName = "nbt"
	}
	reciver, err := ExpandReciver(e.Current.PtrToData, valueType, e.EncodeOption)
	if err != nil {
		return err
	}
	e.Current.ValueReciver = reciver
	return nil
}

func ExpandReciver(ptrToData any, valueType nbt.HighLevelValueType, opt EncodeOption) (ValueReciver, error) {
	overrider, ok := ptrToData.(EncoderOverride)
	if ok {
		return overrider.GetReciver(valueType)
	}

	val := reflect.ValueOf(ptrToData)
	elem := val.Elem()
	if elem.Kind() == reflect.Pointer {
		return ExpandReciver(elem.Interface(), valueType, opt)
	}
	if elem.Kind() == reflect.Interface {
		// ptrToData=&any
		return ExpandAny(ptrToData, valueType, elem, opt)
	}
	// ptrToData=&int8,&int16,...
	return ExpandPointer(ptrToData, valueType, elem, opt, val.UnsafePointer())
}

func (e *GoObjEncoder) Pop() (err error) {
	if len(e.Stack) == 0 {
		panic("pop on empty stack")
	}
	e.Current = e.Stack[len(e.Stack)-1]
	e.Stack = e.Stack[:len(e.Stack)-1]
	return err
}

func (e *GoObjEncoder) Push(ptrToData any, reciver ValueReciver) {
	e.Stack = append(e.Stack, e.Current)
	e.Current = Current{PtrToData: ptrToData, ValueReciver: reciver}
}

func (e *GoObjEncoder) BeginComplexCompound() error {
	e.Current.aux.emptyCompound = true
	if e.Current.CompoundConstructor == nil {
		return fmt.Errorf("%T cannot accept compound value", e.Current.PtrToData)
	}
	return e.Current.CompoundConstructor.OnResetCompound()
}

func (e *GoObjEncoder) NextCompoundElement() error {
	// 弹出最后一个元素
	if err := e.Pop(); err != nil {
		return err
	}
	return e.Current.CompoundConstructor.OnNextElement()
}

func (e *GoObjEncoder) WriteCompoundTag(tag string, valueType nbt.HighLevelValueType) error {
	e.Current.aux.emptyCompound = false
	ptr, err := e.Current.CompoundConstructor.GetTagPointer(tag, valueType)
	if err != nil {
		return err
	}
	reciver, err := ExpandReciver(ptr, valueType, e.EncodeOption)
	if err != nil {
		return err
	}
	e.Push(ptr, reciver)
	return nil
}

func (e *GoObjEncoder) EndComplexCompound() error {
	// 弹出最后一个元素
	if !e.Current.aux.emptyCompound {
		if err := e.Pop(); err != nil {
			return err
		}
	} else {
		e.Current.aux.emptyCompound = false
	}
	return e.Current.CompoundConstructor.OnCompoundFinish()
}

func (e *GoObjEncoder) BeginAndEndComplexEmptyList() error {
	if e.Current.ListConstructor == nil {
		return fmt.Errorf("%T cannot accept non typed empty list", e.Current.PtrToData)
	}
	return e.Current.ListConstructor.NonTypedEmptyList()
}

func (e *GoObjEncoder) BeginComplexList(valueType nbt.HighLevelValueType, len int, typeErase bool) error {
	if e.Current.ListConstructor == nil {
		return fmt.Errorf("%T cannot accept list", e.Current.PtrToData)
	}
	c := e.Current.ListConstructor
	ptr, err := e.Current.ListConstructor.GetReusableElementPointer(valueType, len, typeErase)
	if err != nil {
		return err
	}
	reciver, err := ExpandReciver(ptr, valueType, e.EncodeOption)
	if err != nil {
		return err
	}
	e.Push(ptr, reciver)
	e.Current.aux.lastListElemType = valueType
	e.Current.aux.parentListConstructor = c
	e.Current.aux.lastListValueReciver = e.Current.ValueReciver
	return nil
}

func (e *GoObjEncoder) NextListElement(valueType nbt.HighLevelValueType) error {
	if e.Current.aux.lastListElemType == valueType {
		if err := e.Current.aux.parentListConstructor.OnNextElement(); err != nil {
			return err
		}
		e.Current.ValueReciver = e.Current.aux.lastListValueReciver
		return nil
	}
	ptr := e.Current.PtrToData
	e.Pop()
	if err := e.Current.ListConstructor.OnNextElement(); err != nil {
		return err
	}

	reciver, err := ExpandReciver(ptr, valueType, e.EncodeOption)
	if err != nil {
		return err
	}
	c := e.Current.ListConstructor
	e.Push(ptr, reciver)
	e.Current.aux.lastListElemType = valueType
	e.Current.aux.parentListConstructor = c
	e.Current.aux.lastListValueReciver = e.Current.ValueReciver
	return nil
}

func (e *GoObjEncoder) EndComplexList() error {
	// 弹出最后一个元素
	if err := e.Pop(); err != nil {
		return err
	}
	return e.Current.ListConstructor.OnListFinish()
}

func (e *GoObjEncoder) BeginComplexArray(valueType nbt.HighLevelValueType, len int) error {
	if e.Current.ArrayConstructor == nil {
		return fmt.Errorf("%T cannot accept array", e.Current.PtrToData)
	}
	ptr, err := e.Current.ArrayConstructor.GetReusableElementPointer(valueType, len)
	if err != nil {
		return err
	}
	reciver, err := ExpandReciver(ptr, valueType, e.EncodeOption)
	if err != nil {
		return err
	}
	c := e.Current.ArrayConstructor
	e.Push(ptr, reciver)
	e.Current.aux.parentArrayConstructor = c
	return nil
}

func (e *GoObjEncoder) NextArrayElement() error {
	if err := e.Current.aux.parentArrayConstructor.OnNextElement(); err != nil {
		return err
	}
	return nil
}

func (e *GoObjEncoder) EndComplexArray() error {
	if err := e.Pop(); err != nil {
		return err
	}
	return e.Current.ArrayConstructor.OnArrayFinish()
}

func (e *GoObjEncoder) HasFlag(flag nbt.Flag) bool {
	switch flag {
	case nbt.FLAG_SUPPORT_TYPE_ERASE:
		return e.Current.ListConstructor.HasFlag(flag)
	case nbt.FLAG_OMIT_LIST_LEN:
		return e.Current.ListConstructor.HasFlag(flag)
	case nbt.FLAG_OMIT_ARRAY_LEN:
		return e.Current.ArrayConstructor.HasFlag(flag)
	}
	return false
}
