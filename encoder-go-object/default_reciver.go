package encoder_go_object

import (
	"fmt"
	nbt "neomega-nbt"
	"reflect"
	"strings"
	"unsafe"
)

func ExpandPointer(ptrToData any, valueType nbt.HighLevelValueType, val reflect.Value, opt EncodeOption, ptr unsafe.Pointer) (vr ValueReciver, err error) {
	if valueType.IsTrivial() {
		vr.TrivialValReciver, err = GetTrivialReciver(ptrToData, valueType, val, ptr)
		return
	}
	switch valueType {
	case nbt.ComplexCompound:
		switch vk := val.Kind(); vk {
		case reflect.Struct:
			vr.CompoundConstructor = NewPtrStructCompoundConstructor(val, opt)
			return
		case reflect.Map:
			vr.CompoundConstructor = NewPtrMapCompoundConstructor(val, opt)
			return
		}
		return
	case nbt.ComplexList:
		vr.ListConstructor = NewPtrListConstructor(val, opt)
		return
	case nbt.ComplexByteArray, nbt.ComplexLongArray, nbt.ComplexIntArray:
		vr.ArrayConstructor = NewPtrArrayConstructor(val, opt)
		return
	}
	return vr, fmt.Errorf("cannot encode %v to %T", valueType, ptrToData)
}

func ExpandAny(ptrToData any, valueType nbt.HighLevelValueType, val reflect.Value, opt EncodeOption) (vr ValueReciver, err error) {
	if valueType.IsTrivial() {
		vr.TrivialValReciver = func(v any, i64 int64) error {
			val.Set(reflect.ValueOf(v))
			return nil
		}
		return
	}
	switch valueType {
	case nbt.ComplexCompound:
		vr.CompoundConstructor = NewAnyCompoundConstructor(val, opt)
		return
	case nbt.ComplexList:
		vr.ListConstructor = NewAnyListConstructor(val, opt)
		return
	case nbt.ComplexByteArray, nbt.ComplexLongArray, nbt.ComplexIntArray:
		vr.ArrayConstructor = NewAnyArrayConstructor(val, opt)
		return
	}
	return vr, fmt.Errorf("cannot encode %v to %T", valueType, ptrToData)
}

type PtrMapCompoundConstructor struct {
	value       reflect.Value
	lastTag     string
	keyType     reflect.Type
	elementType reflect.Type
	lastValue   reflect.Value
	opt         EncodeOption
}

func NewPtrMapCompoundConstructor(val reflect.Value, opt EncodeOption) CompoundConstructor {
	return &PtrMapCompoundConstructor{value: val, opt: opt}
}

type AnyCompoundConstructor struct {
	PtrMapCompoundConstructor
	anyRV reflect.Value
}

func NewAnyCompoundConstructor(rv reflect.Value, opt EncodeOption) CompoundConstructor {
	return &AnyCompoundConstructor{
		PtrMapCompoundConstructor: PtrMapCompoundConstructor{opt: opt},
		anyRV:                     rv,
	}
}
func (c *PtrMapCompoundConstructor) OnResetCompound() error {
	valType := c.value.Type()
	m := reflect.MakeMap(valType)
	c.value.Set(m)
	c.elementType = valType.Elem()
	c.keyType = valType.Key()
	c.lastValue = reflect.New(c.elementType).Elem()
	return nil
}

func (c *AnyCompoundConstructor) OnResetCompound() error {
	value := new(map[string]any)
	rm := reflect.ValueOf(value).Elem()
	c.PtrMapCompoundConstructor.value = rm
	return c.PtrMapCompoundConstructor.OnResetCompound()
}
func (c *AnyCompoundConstructor) OnCompoundFinish() error {
	if err := c.PtrMapCompoundConstructor.OnCompoundFinish(); err != nil {
		return err
	}
	c.anyRV.Set(c.PtrMapCompoundConstructor.value)
	return nil
}

func (c *PtrMapCompoundConstructor) GetTagPointer(tag string, valueType nbt.HighLevelValueType) (any, error) {
	c.lastTag = tag
	c.lastValue = reflect.New(c.elementType)
	return c.lastValue.Interface(), nil
}

func (c *PtrMapCompoundConstructor) OnNextElement() error {
	key := reflect.New(c.keyType).Elem()
	key.SetString(c.lastTag)
	c.value.SetMapIndex(key, c.lastValue.Elem())
	return nil
}

func (c *PtrMapCompoundConstructor) OnCompoundFinish() error {
	key := reflect.New(c.keyType).Elem()
	key.SetString(c.lastTag)
	c.value.SetMapIndex(key, c.lastValue.Elem())
	return nil
}

type PtrStructCompoundConstructor struct {
	value     reflect.Value
	lastField reflect.Value
	opt       EncodeOption
}

func NewPtrStructCompoundConstructor(val reflect.Value, opt EncodeOption) CompoundConstructor {
	// val 必须是 struct
	return &PtrStructCompoundConstructor{
		value: val,
		opt:   opt,
	}
}
func (c *PtrStructCompoundConstructor) OnResetCompound() error {
	c.value.SetZero()
	return nil
}
func (c *PtrStructCompoundConstructor) GetTagPointer(
	tag string,
	valueType nbt.HighLevelValueType,
) (any, error) {

	t := c.value.Type()

	// 1. 优先：字段名匹配
	if f, ok := t.FieldByName(tag); ok {
		v := c.value.FieldByIndex(f.Index)
		if !v.CanSet() {
			return nil, fmt.Errorf("field %s cannot be set", tag)
		}
		c.lastField = v
		return v.Addr().Interface(), nil
	}

	found := false
	// 2. struct tag 匹配（假设使用 `nbt` tag）
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)

		raw := f.Tag.Get(c.opt.TagName)
		if raw == "-" {
			continue
		}
		name := raw
		if i := strings.IndexByte(raw, ','); i >= 0 {
			name = raw[:i]
		}
		if name == "" {
			name = f.Name
		}
		if name == tag {
			v := c.value.Field(i)
			c.lastField = v
			found = true
			return v.Addr().Interface(), nil
		}
	}
	if !found {
		var omit any
		return &omit, nil
	}

	return nil, fmt.Errorf("no such field for tag %q", tag)
}

func (c *PtrStructCompoundConstructor) OnNextElement() error    { return nil }
func (c *PtrStructCompoundConstructor) OnCompoundFinish() error { return nil }

type PtrListConstructor struct {
	postponeReciver reflect.Value
	opt             EncodeOption
	append          func() error
	finish          func() error
}

func NewPtrListConstructor(rv reflect.Value, opt EncodeOption) ListConstructor {
	return &PtrListConstructor{
		postponeReciver: rv,
		opt:             opt,
	}
}

func (c *PtrListConstructor) NonTypedEmptyList() error {
	sliceType := c.postponeReciver.Type()
	c.postponeReciver.Set(reflect.MakeSlice(sliceType, 0, 0))
	return nil
}

func (c *PtrListConstructor) HasFlag(flag nbt.Flag) bool {
	switch flag {
	case nbt.FLAG_OMIT_LIST_LEN:
		return c.postponeReciver.Type().Kind() == reflect.Slice // 这个实现支持动态扩容，所以不必准确报告 list 长度
	case nbt.FLAG_SUPPORT_TYPE_ERASE:
		return c.postponeReciver.Type().Elem().Kind() == reflect.Interface // 若传入为 invalid 类型，则 判断是否为 []any
	default:
		return false
	}
}

func (c *PtrListConstructor) GetReusableElementPointer(valueType nbt.HighLevelValueType, len int, typeErase bool) (any, error) {
	rvT := c.postponeReciver.Type()
	et := rvT.Elem()
	if typeErase && et.Kind() != reflect.Interface {
		return nil, fmt.Errorf("%v not support type erase", et)
	}
	tmpElemPtr := reflect.New(et)
	elem := tmpElemPtr.Elem()
	switch rvT.Kind() {
	case reflect.Slice:
		s := reflect.MakeSlice(rvT, 0, len)
		c.append = func() error {
			s = reflect.Append(s, elem)
			return nil
		}
		c.finish = func() error {
			c.postponeReciver.Set(s)
			return nil
		}
	case reflect.Array:
		c.postponeReciver.SetZero()
		i := 0
		c.append = func() error {
			c.postponeReciver.Index(i).Set(elem)
			i += 1
			return nil
		}
		c.finish = func() error {
			return nil
		}
	default:
		return nil, fmt.Errorf("%v cannot receive list", rvT)
	}

	return tmpElemPtr.Interface(), nil
}

func (c *PtrListConstructor) OnListFinish() error {
	if err := c.append(); err != nil {
		return err
	}
	return c.finish()
}

func (c *PtrListConstructor) OnNextElement() error {
	return c.append()
}

type AnyListConstructor struct {
	PtrListConstructor
	anyRV reflect.Value
}

func NewAnyListConstructor(rv reflect.Value, opt EncodeOption) ListConstructor {
	return &AnyListConstructor{
		PtrListConstructor: PtrListConstructor{
			postponeReciver: rv,
			opt:             opt,
		},
		anyRV: rv,
	}
}

func (c *AnyListConstructor) NonTypedEmptyList() error {
	c.postponeReciver.Set(reflect.ValueOf([]any{}))
	return nil
}

func (c *AnyListConstructor) HasFlag(flag nbt.Flag) bool {
	switch flag {
	case nbt.FLAG_OMIT_LIST_LEN:
		return true // 这个实现支持动态扩容，所以不必准确报告 list 长度
	case nbt.FLAG_SUPPORT_TYPE_ERASE:
		return true
	default:
		return false
	}
}

func (c *AnyListConstructor) GetReusableElementPointer(valueType nbt.HighLevelValueType, len int, typeErase bool) (any, error) {
	var listPtr any
	if !typeErase {
		if valueType.IsTrivial() {
			switch valueType {
			case nbt.TrivialByte:
				if c.opt.UseInt8Sequence {
					listPtr = new([]int8)
				} else {
					listPtr = new([]uint8)
				}
			case nbt.TrivialShort:
				listPtr = new([]int16)
			case nbt.TrivialInt:
				listPtr = new([]int32)
			case nbt.TrivialLong:
				listPtr = new([]int64)
			case nbt.TrivialFloat:
				listPtr = new([]float32)
			case nbt.TrivialDouble:
				listPtr = new([]float64)
			default:
				panic(fmt.Errorf("%v is not trivial type", valueType))
			}
		} else if valueType.IsCompound() {
			listPtr = new([]map[string]any)
		} else if valueType.IsList() || valueType.IsArray() {
			listPtr = new([]any)
		} else {
			listPtr = new([]any)
		}
	} else {
		listPtr = new([]any)
	}
	c.postponeReciver = reflect.ValueOf(listPtr).Elem()
	return c.PtrListConstructor.GetReusableElementPointer(valueType, len, typeErase)
}

func (c *AnyListConstructor) OnListFinish() error {
	if err := c.PtrListConstructor.OnListFinish(); err != nil {
		return err
	}
	c.anyRV.Set(c.postponeReciver)
	return nil
}

type PtrArrayConstructor struct {
	postponeReciver reflect.Value
	opt             EncodeOption
	append          func() error
	finish          func() error
}

func NewPtrArrayConstructor(rv reflect.Value, opt EncodeOption) ArrayConstructor {
	return &PtrArrayConstructor{
		postponeReciver: rv,
		opt:             opt,
	}
}

func (c *PtrArrayConstructor) HasFlag(flag nbt.Flag) bool {
	switch flag {
	case nbt.FLAG_OMIT_ARRAY_LEN:
		return c.postponeReciver.Type().Kind() == reflect.Slice // 这个实现支持动态扩容，所以不必准确报告 list 长度
	default:
		return false
	}
}

func (c *PtrArrayConstructor) GetReusableElementPointer(valueType nbt.HighLevelValueType, len int) (any, error) {
	rvT := c.postponeReciver.Type()
	et := rvT.Elem()
	tmpElemPtr := reflect.New(et)
	elem := tmpElemPtr.Elem()
	switch rvT.Kind() {
	case reflect.Slice:
		s := reflect.MakeSlice(rvT, 0, len)
		c.append = func() error {
			s = reflect.Append(s, elem)
			return nil
		}
		c.finish = func() error {
			c.postponeReciver.Set(s)
			return nil
		}
	case reflect.Array:
		c.postponeReciver.SetZero()
		i := 0
		c.append = func() error {
			c.postponeReciver.Index(i).Set(elem)
			i += 1
			return nil
		}
		c.finish = func() error {
			return nil
		}
	default:
		return nil, fmt.Errorf("%v cannot receive list", rvT)
	}

	return tmpElemPtr.Interface(), nil
}

func (c *PtrArrayConstructor) OnArrayFinish() error {
	if err := c.append(); err != nil {
		return err
	}
	return c.finish()
}

func (c *PtrArrayConstructor) OnNextElement() error {
	return c.append()
}

type AnyArrayConstructor struct {
	PtrArrayConstructor
	anyRV   reflect.Value
	isSlice bool
}

func NewAnyArrayConstructor(rv reflect.Value, opt EncodeOption) ArrayConstructor {
	return &AnyArrayConstructor{
		PtrArrayConstructor: PtrArrayConstructor{
			postponeReciver: rv,
			opt:             opt,
		},
		anyRV:   rv,
		isSlice: opt.ArrayAsSlice,
	}
}

func (c *AnyArrayConstructor) GetReusableElementPointer(valueType nbt.HighLevelValueType, len int) (any, error) {
	if c.isSlice {
		var listPtr any
		switch valueType {
		case nbt.TrivialByte:
			if c.opt.UseInt8Sequence {

				listPtr = new([]int8)
			} else {
				listPtr = new([]uint8)
			}
		case nbt.TrivialInt:
			listPtr = new([]int32)
		case nbt.TrivialLong:
			listPtr = new([]int64)
		default:
			panic(fmt.Errorf("%v is not trivial type", valueType))
		}
		c.postponeReciver = reflect.ValueOf(listPtr).Elem()
		return c.PtrArrayConstructor.GetReusableElementPointer(valueType, len)
	} else {
		var elemType reflect.Type

		switch valueType {
		case nbt.TrivialByte:
			if c.opt.UseInt8Sequence {
				elemType = reflect.TypeOf(int8(0))
			} else {
				elemType = reflect.TypeOf(uint8(0))
			}

		case nbt.TrivialInt:
			elemType = reflect.TypeOf(int32(0))

		case nbt.TrivialLong:
			elemType = reflect.TypeOf(int64(0))

		default:
			panic(fmt.Errorf("%v is not trivial type", valueType))
		}

		// runtime 构造 [len]T
		arrType := reflect.ArrayOf(len, elemType)

		// *[len]T
		arrPtr := reflect.New(arrType)

		// 保存为 array，而不是 slice
		c.postponeReciver = arrPtr.Elem()

		return c.PtrArrayConstructor.GetReusableElementPointer(valueType, len)
	}
}

func (c *AnyArrayConstructor) HasFlag(flag nbt.Flag) bool {
	switch flag {
	case nbt.FLAG_OMIT_LIST_LEN:
		return c.isSlice // 这个实现支持动态扩容，所以不必准确报告 array 长度
	default:
		return false
	}
}

func (c *AnyArrayConstructor) OnArrayFinish() error {
	if err := c.PtrArrayConstructor.OnArrayFinish(); err != nil {
		return err
	}
	c.anyRV.Set(c.postponeReciver)
	return nil
}
