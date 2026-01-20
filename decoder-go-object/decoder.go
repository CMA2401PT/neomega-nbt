package decoder_go_object

import (
	"fmt"
	nbt "neomega-nbt"
	"reflect"
	"sort"
	"strings"
)

type DecodeOption struct {
	// 根标签，但是其实大多数情况下这个会被忽略或者留空
	RootTag string
	// SliceAsArray=false 时，
	// 同理包括 []uint8 []int8 []uint32 []int32 []uint64 []int64
	// SliceAsArray=true 时，
	// []int32 等会被当作对应类型的 array 处理
	// 特别的，对于[]any
	// 当其长度为 0 时，报告为空列表，否则报告为第一个元素的类型
	// 因为 []any 可能具有不同实际类型
	// 所以从第二个元素开始（第一个元素通过 BeginComplexList 报告类型）
	// 通过 NextListElement 报告类型
	SliceAsArray bool
	// SortMapKey=false 时，map 的键值是随机访问的，否则会按字母序先排序
	SortMapKey bool
	// 一般为 "nbt" 即当存在这个标签时，使用标签指定的名字
	TagName string
}

func HighLevelDecodeTo[R nbt.EncoderHighLevel](
	orig any, reciver R, opt DecodeOption,
) error {
	if opt.TagName == "" {
		opt.TagName = "nbt"
	}
	return HighLevelDecodeElemTo(orig, reciver, nbt.ValuePosCompoundOrRoot, func(hlvt nbt.HighLevelValueType) error {
		return reciver.WriteRootTag(opt.RootTag, hlvt)
	}, opt)
}

func HighLevelDecodeElemTo[R nbt.EncoderHighLevel](
	orig any, reciver R,
	vp nbt.ValuePos,
	writeType func(nbt.HighLevelValueType) error,
	opt DecodeOption,
) error {
	orig, val, kind, vt := UnpackAny(orig, opt.SliceAsArray)
	return HighLevelDecodeUpackedElemTo(orig, val, kind, vt, reciver, vp, writeType, opt)
}

func HighLevelDecodeUpackedElemTo[R nbt.EncoderHighLevel](
	orig any, val reflect.Value, kind reflect.Kind, vt nbt.HighLevelValueType,
	reciver R,
	vp nbt.ValuePos,
	writeType func(nbt.HighLevelValueType) error,
	opt DecodeOption,
) error {
	if !vt.IsValid() {
		return fmt.Errorf("unaccapt value %v to nbt value type", orig)
	}
	if writeType != nil {
		if err := writeType(vt); err != nil {
			return err
		}
	}
	if vt.IsTrivial() {
		return WriteTrivialReflectValue(kind, val, reciver, vp)
	}
	if vt.IsArray() {
		return HighLevelDecodeArrayTo(orig, val, kind, reciver, opt)
	}
	if vt.IsList() {
		HighLevelDecodeListTo(orig, val, kind, reciver, opt)
	}
	if vt.IsCompound() {
		return HighLevelDecodeCompoundTo(val, kind, reciver, opt)
	}
	return nil
}

// 将一个 any 类型去除指针化，并获得 go 的反射值，类型和 nbt 类型
func UnpackAny(orig any, sliceAsArray bool) (any, reflect.Value, reflect.Kind, nbt.HighLevelValueType) {
	val := reflect.ValueOf(orig)
	kind := val.Kind()
	for val.Kind() == reflect.Pointer {
		orig = val.Elem().Interface()
		val = reflect.ValueOf(orig)
		kind = val.Kind()
	}
	vt := ReflectKindToHighLevelValueType(kind, func() reflect.Type { return val.Type() }, sliceAsArray)
	return orig, val, kind, vt
}

func HighLevelDecodeListTo[R nbt.EncoderHighLevel](
	orig any,
	val reflect.Value, kind reflect.Kind, reciver R,
	opt DecodeOption,
) error {
	// fast path ([]int8,[]uint8,[]int32,...)
	ok, err := HighLevelDecodeListToFastPath(orig, val, kind, reciver, opt)
	if ok {
		return err
	}
	// end fast path
	elem := val.Type().Elem()
	n := val.Len()
	et := ReflectKindToHighLevelValueType(elem.Kind(), func() reflect.Type { return elem.Elem() }, opt.SliceAsArray)
	typeCheck := false
	typeErase := false
	if et == nbt.Invalid {
		if n == 0 { // 当元素长度为 0 时，类型推断也无从提起
			return reciver.BeginAndEndComplexEmptyList()
		}
		// 否则，就从第一个元素推断实际类型
		fisrtElem := val.Index(0).Interface()
		_, _, _, et = UnpackAny(fisrtElem, opt.SliceAsArray)
		// 如果第一个元素类型也不知道，就万策尽了
		if et == nbt.Invalid {
			return fmt.Errorf("unaccaptable type %v to nbt list element", elem)
		}
		if reciver.HasFlag(nbt.FLAG_SUPPORT_TYPE_ERASE) {
			typeErase = true
			// 如果支持类型擦除，就无需检查每个的类型
		} else {
			typeCheck = true
		}
	}
	if err := reciver.BeginComplexList(et, n, typeErase); err != nil {
		return err
	}
	for i := range n {
		subV := val.Index(i).Interface()
		subV, val, kind, vt := UnpackAny(subV, opt.SliceAsArray)
		if i != 0 {
			if typeCheck && vt != et {
				return fmt.Errorf("mismatch list element type %v(%v) with declared type %v", vt, val.Type(), et)
			}
			if err := reciver.NextListElement(vt); err != nil {
				return err
			}
		}
		if err := HighLevelDecodeUpackedElemTo(subV, val, kind, vt, reciver, nbt.ValuePosList, nil, opt); err != nil {
			return err
		}
	}
	return reciver.EndComplexList()
}

func HighLevelDecodeArrayTo[R nbt.EncoderHighLevel](
	orig any, val reflect.Value, kind reflect.Kind, reciver R,
	opt DecodeOption,
) error {
	n := val.Len()
	et, write := GetArrayElementWritor(val, reciver)
	if et == nbt.Invalid {
		return fmt.Errorf("unaccaptable array element type  %v", val.Type().Elem())
	}
	if err := reciver.BeginComplexArray(et, n); err != nil {
		return err
	}
	for i := range n {
		if i != 0 {
			if err := reciver.NextArrayElement(); err != nil {
				return err
			}
		}
		write(i)
	}
	return reciver.EndComplexArray()
}

func HighLevelDecodeCompoundTo[R nbt.EncoderHighLevel](
	val reflect.Value, kind reflect.Kind, reciver R,
	opt DecodeOption,
) error {
	if kind == reflect.Map {
		if opt.SortMapKey {
			return HighLevelDecodeMapSort(val, kind, reciver, opt)
		} else {
			return HighLevelDecodeMapUnsort(val, kind, reciver, opt)
		}
	}
	return HighLevelDecodeStruct(val, reciver, true, true, opt)
}

func HighLevelDecodeMapUnsort[R nbt.EncoderHighLevel](
	val reflect.Value, kind reflect.Kind, reciver R,
	opt DecodeOption,
) error {
	if err := reciver.BeginComplexCompound(); err != nil {
		return err
	}
	stringK := true
	if val.Type().Key().Kind() != reflect.String {
		stringK = false
	}
	iter := val.MapRange()
	// take data
	first := true
	for iter.Next() {
		if first {
			first = false
		} else {
			if err := reciver.NextCompoundElement(); err != nil {
				return err
			}
		}
		key := ""
		if stringK {
			key = iter.Key().String()
		} else {
			key = fmt.Sprintf("%v", iter.Key().Interface())
		}
		ev := iter.Value().Interface()
		if err := HighLevelDecodeElemTo(ev, reciver, nbt.ValuePosCompoundOrRoot, func(hlvt nbt.HighLevelValueType) error {
			return reciver.WriteCompoundTag(key, hlvt)
		}, opt); err != nil {
			return err
		}
	}
	if err := reciver.EndComplexCompound(); err != nil {
		return err
	}
	return nil
}

func HighLevelDecodeMapSort[R nbt.EncoderHighLevel](
	val reflect.Value, kind reflect.Kind, reciver R,
	opt DecodeOption,
) error {
	if err := reciver.BeginComplexCompound(); err != nil {
		return err
	}
	stringK := true
	if val.Type().Key().Kind() != reflect.String {
		stringK = false
	}

	// 1. 收集所有 key（以字符串形式作为排序依据）
	keys := make([]string, 0, val.Len())
	keyMap := make(map[string]reflect.Value, val.Len())

	iter := val.MapRange()
	for iter.Next() {
		var k string
		if stringK {
			k = iter.Key().String()
		} else {
			k = fmt.Sprintf("%v", iter.Key().Interface())
		}
		keys = append(keys, k)
		keyMap[k] = iter.Key()
	}

	// 2. 按字母序排序
	sort.Strings(keys)

	// 3. 按排序后的顺序访问
	for i, k := range keys {
		if i > 0 {
			if err := reciver.NextCompoundElement(); err != nil {
				return err
			}
		}

		origKey := keyMap[k]
		ev := val.MapIndex(origKey).Interface()

		if err := HighLevelDecodeElemTo(
			ev,
			reciver,
			nbt.ValuePosCompoundOrRoot,
			func(hlvt nbt.HighLevelValueType) error {
				return reciver.WriteCompoundTag(k, hlvt)
			}, opt); err != nil {
			return err
		}
	}
	if err := reciver.EndComplexCompound(); err != nil {
		return err
	}
	return nil
}

func HighLevelDecodeStruct[R nbt.EncoderHighLevel](
	val reflect.Value,
	reciver R, first bool, root bool, opt DecodeOption,
) error {
	if root {
		if err := reciver.BeginComplexCompound(); err != nil {
			return err
		}
	}

	for i := 0; i < val.NumField(); i++ {
		fieldType := val.Type().Field(i)
		if fieldType.PkgPath != "" {
			continue
		}
		fieldValue := val.Field(i)

		// 1. 匿名字段优先处理（语义不变）
		if fieldType.Anonymous {
			if err := HighLevelDecodeStruct(fieldValue, reciver, first, false, opt); err != nil {
				return err
			}
			continue
		}
		rawTag := fieldType.Tag.Get(opt.TagName)
		if rawTag == "-" {
			continue
		}
		// 2. 解析主名和选项
		name := rawTag
		opts := ""

		if i := strings.IndexByte(rawTag, ','); i >= 0 {
			name = rawTag[:i]
			opts = rawTag[i+1:]
		}

		// 3. omitempty 处理
		if strings.Contains(opts, "omitempty") {
			if reflect.DeepEqual(
				fieldValue.Interface(),
				reflect.Zero(fieldValue.Type()).Interface(),
			) {
				continue
			}
		}

		// 4. 确定最终 tag 名
		tagName := name
		if tagName == "" {
			tagName = fieldType.Name
		}
		if first {
			first = false
		} else {
			if err := reciver.NextCompoundElement(); err != nil {
				return err
			}
		}
		if err := HighLevelDecodeElemTo(fieldValue.Interface(), reciver, nbt.ValuePosCompoundOrRoot, func(hlvt nbt.HighLevelValueType) error {
			return reciver.WriteCompoundTag(tagName, hlvt)
		}, opt); err != nil {
			return err
		}
	}
	if root {
		if err := reciver.EndComplexCompound(); err != nil {
			return err
		}
	}
	return nil
}
