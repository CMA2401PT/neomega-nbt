package encoder_go_object

import (
	"fmt"
	nbt "neomega-nbt"
	"sort"
	"strings"
	"testing"
)

func Encode(v any, operates func(*GoObjEncoder)) {
	enc := &GoObjEncoder{
		Current: Current{
			PtrToData: v,
		},
	}
	operates(enc)
}

func EncodeWithOpt(v any, operates func(*GoObjEncoder), opt EncodeOption) {
	enc := &GoObjEncoder{
		Current: Current{
			PtrToData: v,
		},
		EncodeOption: opt,
	}
	operates(enc)
}

type MyInt8Override int8

func (i *MyInt8Override) GetReciver(valueType nbt.HighLevelValueType) (
	vr ValueReciver,
	err error,
) {
	if valueType != nbt.TrivialByte {
		return vr, fmt.Errorf("can only accept int8")
	}
	vr.TrivialValReciver = func(v any, i64 int64) error {
		*i = MyInt8Override(int8(i64) + 10)
		return nil
	}
	return vr, nil
}

func (i MyInt8Override) String() string {
	return fmt.Sprintf("<MyInt8:%v>", int8(i))
}

type MyMapOverride map[string]MyInt8Override

func (m *MyMapOverride) GetReciver(valueType nbt.HighLevelValueType) (vr ValueReciver, err error) {
	vr.CompoundConstructor = &MyMapOverrideConstructor{ptr: m}
	return
}

type MyMapOverrideConstructor struct {
	ptr       *MyMapOverride
	data      MyMapOverride // map is already a pointer inside
	lastTag   string
	lastValue MyInt8Override
}

func (c *MyMapOverrideConstructor) OnResetCompound() error {
	c.data = make(MyMapOverride)
	*c.ptr = c.data
	return nil
}

func (c *MyMapOverrideConstructor) GetTagPointer(tag string, valueType nbt.HighLevelValueType) (any, error) {
	c.lastTag = tag
	c.lastValue = 0
	if tag == "my_private" {
		return nil, fmt.Errorf("this key cannot be set!")
	}
	return &c.lastValue, nil
}

func (c *MyMapOverrideConstructor) OnNextElement() error {
	c.data[c.lastTag+"_override"] = c.lastValue
	return nil
}

func (c *MyMapOverrideConstructor) OnCompoundFinish() error {
	c.data[c.lastTag+"_override"] = c.lastValue
	return nil
}

type MyMapToDefaultOverride map[string]any

func (m *MyMapToDefaultOverride) GetReciver(valueType nbt.HighLevelValueType) (vr ValueReciver, err error) {
	vr.CompoundConstructor = &MyMapToDefaultOverrideConstructor{ptr: m}
	return
}

type MyMapToDefaultOverrideConstructor struct {
	ptr       *MyMapToDefaultOverride
	data      MyMapToDefaultOverride // map is already a pointer inside
	lastTag   string
	lastValue any
}

func (c *MyMapToDefaultOverrideConstructor) OnResetCompound() error {
	c.data = make(MyMapToDefaultOverride)
	*c.ptr = c.data
	return nil
}

func (c *MyMapToDefaultOverrideConstructor) GetTagPointer(tag string, valueType nbt.HighLevelValueType) (any, error) {
	c.lastTag = tag
	c.lastValue = nil
	return &c.lastValue, nil
}

func (c *MyMapToDefaultOverrideConstructor) OnNextElement() error {
	c.data[c.lastTag] = c.lastValue
	return nil
}

func (c *MyMapToDefaultOverrideConstructor) OnCompoundFinish() error {
	c.data[c.lastTag] = c.lastValue
	return nil
}

type MyString string
type MyStruct struct {
	A int8 `nbt:"a,omitempty"`
	B uint8
	C MyInt8Override
}

type MyInt8ListOverride []MyInt8Override

func (i *MyInt8ListOverride) GetReciver(valueType nbt.HighLevelValueType) (vr ValueReciver, err error) {
	// valueType==nbt.ComplexList
	vr.ListConstructor = &MyInt8ListConstructor{data: i}
	return
}

type MyInt8ListConstructor struct {
	data      *MyInt8ListOverride
	lastValue MyInt8Override
}

func (c *MyInt8ListConstructor) NonTypedEmptyList() error {
	*c.data = make(MyInt8ListOverride, 0)
	return nil
}

func (c *MyInt8ListConstructor) HasFlag(flag nbt.Flag) bool {
	switch flag {
	case nbt.FLAG_OMIT_LIST_LEN:
		return true // 这个实现支持动态扩容，所以不必准确报告 list 长度
	default:
		return false
	}
}

func (c *MyInt8ListConstructor) GetReusableElementPointer(valueType nbt.HighLevelValueType, len int, typeErase bool) (any, error) {
	if valueType != nbt.TrivialByte || typeErase {
		return nil, fmt.Errorf("can only accept int8")
	}
	*c.data = make(MyInt8ListOverride, 0, len)
	return &c.lastValue, nil
}

func (c *MyInt8ListConstructor) OnListFinish() error {
	(*c.data) = append((*c.data), c.lastValue)
	return nil
}

func (c *MyInt8ListConstructor) OnNextElement() error {
	// 因为不支持类型擦除的标志，所以类型不会发生改变
	// 我们只需把数据压入即可
	(*c.data) = append((*c.data), c.lastValue)
	return nil
}

type MyInt8ArrayOverride []MyInt8Override

func (i *MyInt8ArrayOverride) GetReciver(valueType nbt.HighLevelValueType) (vr ValueReciver, err error) {
	vr.ArrayConstructor = &MyInt8ArrayConstructor{data: i}
	return
}

type MyInt8ArrayConstructor struct {
	data      *MyInt8ArrayOverride
	lastValue MyInt8Override
}

func (c *MyInt8ArrayConstructor) HasFlag(flag nbt.Flag) bool {
	switch flag {
	case nbt.FLAG_OMIT_ARRAY_LEN:
		return true // 这个实现支持动态扩容，所以不必准确报告 array 长度
	default:
		return false
	}
}

func (c *MyInt8ArrayConstructor) GetReusableElementPointer(valueType nbt.HighLevelValueType, len int) (any, error) {
	*c.data = make(MyInt8ArrayOverride, 0, len)
	return &c.lastValue, nil
}

func (c *MyInt8ArrayConstructor) OnArrayFinish() error {
	(*c.data) = append((*c.data), c.lastValue)
	return nil
}

func (c *MyInt8ArrayConstructor) OnNextElement() error {
	(*c.data) = append((*c.data), c.lastValue)
	return nil
}

func TestEncode(t *testing.T) {
	expect := func(v any, s string) {
		t.Helper()
		if fmt.Sprintf("%v", v) != s {
			t.Errorf("expect result %v, get %v", s, fmt.Sprintf("%v", v))
		}
	}

	// 自定义覆写
	var a MyInt8Override
	Encode(&a, func(e *GoObjEncoder) {
		e.WriteRootTag("", nbt.TrivialByte)
		e.WriteInt8(2, 0)
	})
	expect(a, "<MyInt8:12>")

	// any
	var ai any
	Encode(&ai, func(e *GoObjEncoder) {
		e.WriteRootTag("", nbt.TrivialByte)
		e.WriteInt8(2, 0)
	})
	expect(ai, "2")

	// 指针
	var ap int8
	app := &ap
	Encode(&app, func(e *GoObjEncoder) {
		e.WriteRootTag("", nbt.TrivialByte)
		e.WriteInt8(2, 0)
	})
	expect(ap, "2")

	// 自定义覆写
	var b MyInt8ListOverride
	Encode(&b, func(e *GoObjEncoder) {
		e.WriteRootTag("", nbt.ComplexList)
		e.BeginComplexList(nbt.TrivialByte, 3, false)
		e.WriteInt8(3, nbt.ValuePosList)
		e.NextListElement(nbt.TrivialByte)
		e.WriteInt8(2, nbt.ValuePosList)
		e.NextListElement(nbt.TrivialByte)
		e.WriteInt8(1, nbt.ValuePosList)
		e.EndComplexList()
	})
	expect(b, "[<MyInt8:13> <MyInt8:12> <MyInt8:11>]")

	// 自定义覆写
	var c MyMapOverride
	Encode(&c, func(e *GoObjEncoder) {
		e.WriteRootTag("", nbt.ComplexCompound)
		e.BeginComplexCompound()
		e.WriteCompoundTag("a", nbt.TrivialByte)
		e.WriteInt8(3, nbt.ValuePosCompoundOrRoot)
		e.NextCompoundElement()
		e.WriteCompoundTag("b", nbt.TrivialByte)
		e.WriteInt8(2, nbt.ValuePosCompoundOrRoot)
		e.NextCompoundElement()
		e.WriteCompoundTag("c", nbt.TrivialByte)
		e.WriteInt8(1, nbt.ValuePosCompoundOrRoot)
		e.EndComplexCompound()
	})
	expect(c, "[a_override:<MyInt8:13>, b_override:<MyInt8:12>, c_override:<MyInt8:11>]")

	// any
	var ci any
	Encode(&ci, func(e *GoObjEncoder) {
		e.WriteRootTag("", nbt.ComplexCompound)
		e.BeginComplexCompound()
		e.WriteCompoundTag("a", nbt.TrivialByte)
		e.WriteInt8(3, nbt.ValuePosCompoundOrRoot)
		e.NextCompoundElement()
		e.WriteCompoundTag("b", nbt.TrivialByte)
		e.WriteInt8(2, nbt.ValuePosCompoundOrRoot)
		e.NextCompoundElement()
		e.WriteCompoundTag("c", nbt.TrivialByte)
		e.WriteInt8(1, nbt.ValuePosCompoundOrRoot)
		e.EndComplexCompound()
	})
	expect(ci, "map[a:3 b:2 c:1]")

	// any
	ci = nil
	Encode(&ci, func(e *GoObjEncoder) {
		e.WriteRootTag("", nbt.ComplexCompound)
		e.BeginComplexCompound()
		e.EndComplexCompound()
	})
	expect(ci, "map[]")

	// any
	Encode(&ci, func(e *GoObjEncoder) {
		e.WriteRootTag("", nbt.ComplexCompound)
		e.BeginComplexCompound()
		e.WriteCompoundTag("a", nbt.ComplexCompound)
		{
			e.BeginComplexCompound()
			e.WriteCompoundTag("a-1", nbt.TrivialByte)
			e.WriteInt8(3, nbt.ValuePosCompoundOrRoot)
			e.NextCompoundElement()
			e.WriteCompoundTag("b-1", nbt.TrivialByte)
			e.WriteInt8(2, nbt.ValuePosCompoundOrRoot)
			e.NextCompoundElement()
			e.WriteCompoundTag("c-1", nbt.TrivialByte)
			e.WriteInt8(1, nbt.ValuePosCompoundOrRoot)
			e.EndComplexCompound()
		}
		e.NextCompoundElement()
		e.WriteCompoundTag("b", nbt.TrivialByte)
		e.WriteInt8(2, nbt.ValuePosCompoundOrRoot)
		e.NextCompoundElement()
		e.WriteCompoundTag("c", nbt.TrivialByte)
		e.WriteInt8(1, nbt.ValuePosCompoundOrRoot)
		e.EndComplexCompound()
	})
	expect(ci, "map[a:map[a-1:3 b-1:2 c-1:1] b:2 c:1]")

	// 指针,指向 map
	var cp map[MyString]int8
	Encode(&cp, func(e *GoObjEncoder) {
		e.WriteRootTag("", nbt.ComplexCompound)
		e.BeginComplexCompound()
		e.WriteCompoundTag("a", nbt.TrivialByte)
		e.WriteInt8(3, nbt.ValuePosCompoundOrRoot)
		e.NextCompoundElement()
		e.WriteCompoundTag("b", nbt.TrivialByte)
		e.WriteInt8(2, nbt.ValuePosCompoundOrRoot)
		e.NextCompoundElement()
		e.WriteCompoundTag("c", nbt.TrivialByte)
		e.WriteInt8(1, nbt.ValuePosCompoundOrRoot)
		e.EndComplexCompound()
	})
	expect(cp, "map[a:3 b:2 c:1]")

	// 指针，指向结构体
	var cps MyStruct
	Encode(&cps, func(e *GoObjEncoder) {
		e.WriteRootTag("", nbt.ComplexCompound)
		e.BeginComplexCompound()
		e.WriteCompoundTag("a", nbt.TrivialByte)
		e.WriteInt8(3, nbt.ValuePosCompoundOrRoot)
		e.NextCompoundElement()
		e.WriteCompoundTag("B", nbt.TrivialByte)
		e.WriteInt8(2, nbt.ValuePosCompoundOrRoot)
		e.NextCompoundElement()
		e.WriteCompoundTag("C", nbt.TrivialByte)
		e.WriteInt8(1, nbt.ValuePosCompoundOrRoot)
		e.EndComplexCompound()
	})
	expect(cps, "{3 2 <MyInt8:11>}")

	// 父对象是被覆写的对象，但是子对象未被覆写
	var d MyMapToDefaultOverride
	Encode(&d, func(e *GoObjEncoder) {
		e.WriteRootTag("", nbt.ComplexCompound)
		e.BeginComplexCompound()
		e.WriteCompoundTag("a", nbt.TrivialByte)
		e.WriteInt8(3, nbt.ValuePosCompoundOrRoot)
		e.NextCompoundElement()
		e.WriteCompoundTag("b", nbt.TrivialByte)
		e.WriteInt8(2, nbt.ValuePosCompoundOrRoot)
		e.NextCompoundElement()
		e.WriteCompoundTag("c", nbt.TrivialByte)
		e.WriteInt8(1, nbt.ValuePosCompoundOrRoot)
		e.EndComplexCompound()
	})
	expect(d, "MyMapToDefaultOverride[a:3, b:2, c:1]")

	// 子对象是被覆写的对象，但是父对象未被覆写
	var e map[MyString]MyInt8Override
	Encode(&e, func(enc *GoObjEncoder) {
		enc.WriteRootTag("", nbt.ComplexCompound)
		enc.BeginComplexCompound()
		enc.WriteCompoundTag("a", nbt.TrivialByte)
		enc.WriteInt8(3, nbt.ValuePosCompoundOrRoot)
		enc.NextCompoundElement()
		enc.WriteCompoundTag("b", nbt.TrivialByte)
		enc.WriteInt8(2, nbt.ValuePosCompoundOrRoot)
		enc.NextCompoundElement()
		enc.WriteCompoundTag("c", nbt.TrivialByte)
		enc.WriteInt8(1, nbt.ValuePosCompoundOrRoot)
		enc.EndComplexCompound()
	})
	expect(e, "map[a:<MyInt8:13> b:<MyInt8:12> c:<MyInt8:11>]")

	// 空列表
	var be MyInt8ListOverride
	Encode(&be, func(e *GoObjEncoder) {
		e.WriteRootTag("", nbt.ComplexList)
		e.BeginAndEndComplexEmptyList()
	})
	expect(be, "[]")

	// any-空列表
	var ba any
	Encode(&ba, func(e *GoObjEncoder) {
		e.WriteRootTag("", nbt.ComplexList)
		e.BeginAndEndComplexEmptyList()
	})
	_, ok := ba.([]any)
	if !ok {
		t.Errorf("error type, get %T", ba)
	}
	expect(ba, "[]")

	// any-平凡类型
	Encode(&ba, func(e *GoObjEncoder) {
		e.WriteRootTag("", nbt.ComplexList)
		e.BeginComplexList(nbt.TrivialByte, 3, false)
		e.WriteInt8(3, nbt.ValuePosList)
		e.NextListElement(nbt.TrivialByte)
		e.WriteInt8(2, nbt.ValuePosList)
		e.NextListElement(nbt.TrivialByte)
		e.WriteInt8(1, nbt.ValuePosList)
		e.EndComplexList()
	})
	_, ok = ba.([]uint8)
	if !ok {
		t.Errorf("error type, get %T", ba)
	}
	expect(ba, "[3 2 1]")

	// any-类型擦除
	Encode(&ba, func(e *GoObjEncoder) {
		e.WriteRootTag("", nbt.ComplexList)
		e.BeginComplexList(nbt.TrivialByte, 3, true)
		e.WriteInt8(3, nbt.ValuePosList)
		e.NextListElement(nbt.ComplexCompound)
		{
			e.BeginComplexCompound()
			e.WriteCompoundTag("a-1", nbt.TrivialByte)
			e.WriteInt8(5, nbt.ValuePosCompoundOrRoot)
			e.NextCompoundElement()
			e.WriteCompoundTag("b-1", nbt.TrivialByte)
			e.WriteInt8(6, nbt.ValuePosCompoundOrRoot)
			e.NextCompoundElement()
			e.WriteCompoundTag("c-1", nbt.TrivialByte)
			e.WriteInt8(7, nbt.ValuePosCompoundOrRoot)
			e.EndComplexCompound()
		}
		e.NextListElement(nbt.ComplexCompound)
		{
			e.BeginComplexCompound()
			e.WriteCompoundTag("a-2", nbt.TrivialByte)
			e.WriteInt8(15, nbt.ValuePosCompoundOrRoot)
			e.NextCompoundElement()
			e.WriteCompoundTag("b-2", nbt.TrivialByte)
			e.WriteInt8(16, nbt.ValuePosCompoundOrRoot)
			e.NextCompoundElement()
			e.WriteCompoundTag("c-2", nbt.TrivialByte)
			e.WriteInt8(17, nbt.ValuePosCompoundOrRoot)
			e.EndComplexCompound()
		}
		e.EndComplexList()
	})
	_, ok = ba.([]any)
	if !ok {
		t.Errorf("error type, get %T", ba)
	}
	expect(ba, "[3 map[a-1:5 b-1:6 c-1:7] map[a-2:15 b-2:16 c-2:17]]")

	// 指针-平凡类型
	var bp []uint8
	Encode(&bp, func(e *GoObjEncoder) {
		e.WriteRootTag("", nbt.ComplexList)
		e.BeginComplexList(nbt.TrivialByte, 3, false)
		e.WriteInt8(3, nbt.ValuePosList)
		e.NextListElement(nbt.TrivialByte)
		e.WriteInt8(2, nbt.ValuePosList)
		e.NextListElement(nbt.TrivialByte)
		e.WriteInt8(1, nbt.ValuePosList)
		e.EndComplexList()
	})
	expect(bp, "[3 2 1]")

	// 指针-覆写类型
	var bpo []MyInt8Override
	Encode(&bpo, func(e *GoObjEncoder) {
		e.WriteRootTag("", nbt.ComplexList)
		e.BeginComplexList(nbt.TrivialByte, 3, false)
		e.WriteInt8(3, nbt.ValuePosList)
		e.NextListElement(nbt.TrivialByte)
		e.WriteInt8(2, nbt.ValuePosList)
		e.NextListElement(nbt.TrivialByte)
		e.WriteInt8(1, nbt.ValuePosList)
		e.EndComplexList()
	})
	expect(bpo, "[<MyInt8:13> <MyInt8:12> <MyInt8:11>]")

	// 指针-空列表
	Encode(&bpo, func(e *GoObjEncoder) {
		e.WriteRootTag("", nbt.ComplexList)
		e.BeginAndEndComplexEmptyList()
	})
	expect(bpo, "[]")

	// 指针-array-覆写类型
	var bpa [4]MyInt8Override
	Encode(&bpa, func(e *GoObjEncoder) {
		e.WriteRootTag("", nbt.ComplexList)
		e.BeginComplexList(nbt.TrivialByte, 3, false)
		e.WriteInt8(3, nbt.ValuePosList)
		e.NextListElement(nbt.TrivialByte)
		e.WriteInt8(2, nbt.ValuePosList)
		e.NextListElement(nbt.TrivialByte)
		e.WriteInt8(1, nbt.ValuePosList)
		e.EndComplexList()
	})
	expect(bpa, "[<MyInt8:13> <MyInt8:12> <MyInt8:11> <MyInt8:0>]")

	// 指针-覆写类型
	var f MyInt8ArrayOverride
	Encode(&f, func(e *GoObjEncoder) {
		e.WriteRootTag("", nbt.ComplexByteArray)
		e.BeginComplexArray(nbt.TrivialByte, 3)
		e.WriteInt8(3, nbt.ValuePosArray)
		e.NextArrayElement()
		e.WriteInt8(2, nbt.ValuePosArray)
		e.NextArrayElement()
		e.WriteInt8(1, nbt.ValuePosArray)
		e.EndComplexArray()
	})
	expect(f, "[<MyInt8:13> <MyInt8:12> <MyInt8:11>]")

	// 指针-slice 类型
	var fp []MyInt8Override
	Encode(&fp, func(e *GoObjEncoder) {
		e.WriteRootTag("", nbt.ComplexByteArray)
		e.BeginComplexArray(nbt.TrivialByte, 3)
		e.WriteInt8(3, nbt.ValuePosArray)
		e.NextArrayElement()
		e.WriteInt8(2, nbt.ValuePosArray)
		e.NextArrayElement()
		e.WriteInt8(1, nbt.ValuePosArray)
		e.EndComplexArray()
	})
	expect(fp, "[<MyInt8:13> <MyInt8:12> <MyInt8:11>]")

	// 指针-array 类型
	var fa [4]MyInt8Override
	Encode(&fa, func(e *GoObjEncoder) {
		e.WriteRootTag("", nbt.ComplexByteArray)
		e.BeginComplexArray(nbt.TrivialByte, 3)
		e.WriteInt8(3, nbt.ValuePosArray)
		e.NextArrayElement()
		e.WriteInt8(2, nbt.ValuePosArray)
		e.NextArrayElement()
		e.WriteInt8(1, nbt.ValuePosArray)
		e.EndComplexArray()
	})
	expect(fa, "[<MyInt8:13> <MyInt8:12> <MyInt8:11> <MyInt8:0>]")

	// 指针-array 类型
	var fany any
	Encode(&fany, func(e *GoObjEncoder) {
		e.WriteRootTag("", nbt.ComplexByteArray)
		e.BeginComplexArray(nbt.TrivialByte, 3)
		e.WriteInt8(3, nbt.ValuePosArray)
		e.NextArrayElement()
		e.WriteInt8(2, nbt.ValuePosArray)
		e.NextArrayElement()
		e.WriteInt8(1, nbt.ValuePosArray)
		e.EndComplexArray()
	})
	_, ok = fany.([3]uint8)
	if !ok {
		t.Errorf("error type, get %T", fany)
	}
	expect(fany, "[3 2 1]")

	// 指针-array 类型
	EncodeWithOpt(&fany, func(e *GoObjEncoder) {
		e.WriteRootTag("", nbt.ComplexByteArray)
		e.BeginComplexArray(nbt.TrivialByte, 3)
		e.WriteInt8(3, nbt.ValuePosArray)
		e.NextArrayElement()
		e.WriteInt8(2, nbt.ValuePosArray)
		e.NextArrayElement()
		e.WriteInt8(1, nbt.ValuePosArray)
		e.EndComplexArray()
	}, EncodeOption{ArrayAsSlice: true, UseInt8Sequence: true})
	_, ok = fany.([]int8)
	if !ok {
		t.Errorf("error type, get %T", fany)
	}
	expect(fany, "[3 2 1]")

}

func (m MyMapOverride) String() string {
	keys := []string{}
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for i, k := range keys {
		keys[i] = fmt.Sprintf("%v:%v", k, m[k])
	}
	return "[" + strings.Join(keys, ", ") + "]"
}

func (m MyMapToDefaultOverride) String() string {
	keys := []string{}
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for i, k := range keys {
		keys[i] = fmt.Sprintf("%v:%v", k, m[k])
	}
	return "MyMapToDefaultOverride[" + strings.Join(keys, ", ") + "]"
}
