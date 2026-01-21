package encoder_go_object

import nbt "neomega-nbt"

type EncoderOverride interface {
	GetReciver(valueType nbt.HighLevelValueType) (ValueReciver, error)
}

// 接收一个平凡的数据 v，包括整数，浮点和字符串
// array，list，compound 不属于此类
// 如果 v 是一个朴素的整数，
// 那么其会被转到 int64 并设置 i64
// 如果是 float32 或 float64，同样会被转到 int64
// 否则 i64 是0
type TrivialValReciver func(
	v any,
	i64 int64,
) error

type ValueReciver struct {
	// 若能接收一个朴素值，应该设置这个
	// 若不能，则为 nil
	TrivialValReciver
	// 若能接收 compound 类型，应该设置这个
	// 若不能，则为nil
	CompoundConstructor
	// 若能接收 list 类型，应该设置这个
	// 若不能，则为nil
	ListConstructor
	ArrayConstructor
}

type CompoundConstructor interface {
	OnResetCompound() error
	GetTagPointer(tag string, valueType nbt.HighLevelValueType) (any, error)
	OnNextElement() error
	OnCompoundFinish() error
}

type ListConstructor interface {
	// 如果 decoder 报告这个 list 类型未知，且长度为 0，这个函数会被调用
	NonTypedEmptyList() error
	// 开始时被调用一次，这个接收器会被反复使用
	// len 是列表总长度
	// 若支持非固定长度列表，例如实际长度未知，则可设置 nbt.FLAG_OMIT_LIST_LEN
	// 此时传入的 len 长度可能为 0
	// 由于类型可能会发生变化，若支持类型变化，则应该设置 nbt.FLAG_SUPPORT_TYPE_ERASE
	// 此时将传入第一个元素的类型，但是typeErase 将被置为 true
	// 若不支持类型变化，则不要设置 nbt.FLAG_SUPPORT_TYPE_ERASE
	GetReusableElementPointer(valueType nbt.HighLevelValueType, len int, typeErase bool) (any, error)
	OnNextElement() error
	OnListFinish() error
	HasFlag(flag nbt.Flag) bool
}

type ArrayConstructor interface {
	GetReusableElementPointer(valueType nbt.HighLevelValueType, len int) (any, error)
	OnNextElement() error
	OnArrayFinish() error
	HasFlag(flag nbt.Flag) bool
}
