package decoder_go_object

import (
	"fmt"
	nbt "neomega-nbt"
	"reflect"
	"sort"
	"strings"
)

func GetDefaultDecoder(orig any, opt DecodeOption) (o DecoderOverride, err error) {
	orig, val, kind, valueType := UnpackAny(orig, opt.SliceAsArray)
	if !valueType.IsValid() {
		err = fmt.Errorf("unaccapt value %v to nbt value type", orig)
		return
	}
	if valueType.IsTrivial() {
		return TrivialValueDecoder{orig, val, kind, valueType}, nil
	}
	if valueType.IsArray() {
		return ArrayValueDecoder{orig, val, kind, valueType}, nil
	}
	if valueType.IsList() {
		return ListValueDecoder{orig, val, kind, opt}, nil
	}
	if valueType.IsCompound() {
		if kind == reflect.Map {
			return MapValueDecoder{val: val, opt: opt}, nil
		}
		return StructValueDecoder{val: val, opt: opt, root: true}, nil
	}
	return nil, fmt.Errorf("unaccapt value %v to nbt value type", orig)
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

type TrivialValueDecoder struct {
	orig      any
	val       reflect.Value
	kind      reflect.Kind
	valueType nbt.HighLevelValueType
}

func (d TrivialValueDecoder) GetNBType() nbt.HighLevelValueType {
	return d.valueType
}

func (d TrivialValueDecoder) DecodeTo(valuePos nbt.ValuePos, encoder nbt.EncoderHighLevel, _ GetDecoderFunc) error {
	return WriteTrivialReflectValue(d.kind, d.val, encoder, valuePos)
}

type ListValueDecoder struct {
	orig any
	val  reflect.Value
	kind reflect.Kind
	opt  DecodeOption
}

func (ListValueDecoder) GetNBType() nbt.HighLevelValueType {
	return nbt.ComplexList
}

func (d ListValueDecoder) DecodeTo(_ nbt.ValuePos, encoder nbt.EncoderHighLevel, decodeChild GetDecoderFunc) error {
	// fast path ([]int8,[]uint8,[]int32,...)
	ok, err := HighLevelDecodeListToFastPath(d.orig, d.val, d.kind, encoder, d.opt)
	if ok {
		return err
	}
	// end fast path
	val := d.val
	elem := val.Type().Elem()
	n := val.Len()
	et := ReflectKindToHighLevelValueType(elem.Kind(), func() reflect.Type { return elem.Elem() }, d.opt.SliceAsArray)
	typeCheck := false
	typeErase := false
	if et == nbt.Invalid {
		if n == 0 { // 当元素长度为 0 时，类型推断也无从提起
			return encoder.BeginAndEndComplexEmptyList()
		}
		// 否则，就从第一个元素推断实际类型
		fisrtElem := val.Index(0).Interface()
		_, _, _, et = UnpackAny(fisrtElem, d.opt.SliceAsArray)
		// 如果第一个元素类型也不知道，就万策尽了
		if et == nbt.Invalid {
			return fmt.Errorf("unaccaptable type %v to nbt list element", elem)
		}
		if encoder.HasFlag(nbt.FLAG_SUPPORT_TYPE_ERASE) {
			typeErase = true
			// 如果支持类型擦除，就无需检查每个的类型
		} else {
			typeCheck = true
		}
	}
	if err := encoder.BeginComplexList(et, n, typeErase); err != nil {
		return err
	}
	for i := range n {
		elem := val.Index(i).Interface()
		decodedElem, err := decodeChild(elem)
		if err != nil {
			return err
		}
		vt := decodedElem.GetNBType()
		if i != 0 {
			if typeCheck && vt != et {
				return fmt.Errorf("mismatch list element type %v(%v) with declared type %v", vt, val.Type(), et)
			}
			if err := encoder.NextListElement(vt); err != nil {
				return err
			}
		}
		if err := decodedElem.DecodeTo(nbt.ValuePosList, encoder, decodeChild); err != nil {
			return err
		}
	}
	return encoder.EndComplexList()
}

type ArrayValueDecoder struct {
	orig           any
	val            reflect.Value
	kind           reflect.Kind
	arrayValueType nbt.HighLevelValueType
}

func (d ArrayValueDecoder) GetNBType() nbt.HighLevelValueType {
	return d.arrayValueType
}

func (d ArrayValueDecoder) DecodeTo(_ nbt.ValuePos, encoder nbt.EncoderHighLevel, _ GetDecoderFunc) error {
	val := d.val
	n := val.Len()
	et, write := GetArrayElementWritor(val, encoder)
	if et == nbt.Invalid {
		return fmt.Errorf("unaccaptable array element type  %v", val.Type().Elem())
	}
	if err := encoder.BeginComplexArray(et, n); err != nil {
		return err
	}
	for i := range n {
		if i != 0 {
			if err := encoder.NextArrayElement(); err != nil {
				return err
			}
		}
		write(i)
	}
	return encoder.EndComplexArray()
}

type StructValueDecoder struct {
	val    reflect.Value
	opt    DecodeOption
	root   bool
	shared *StructValueDecoderShared
}

type StructValueDecoderShared struct {
	isFirst bool
}

func (StructValueDecoder) GetNBType() nbt.HighLevelValueType {
	return nbt.ComplexCompound
}

func (d StructValueDecoder) DecodeTo(_ nbt.ValuePos, encoder nbt.EncoderHighLevel, getDecoder GetDecoderFunc) error {
	if d.shared == nil {
		d.shared = &StructValueDecoderShared{
			isFirst: true,
		}
	}
	root := d.root
	val := d.val
	if root {
		if err := encoder.BeginComplexCompound(); err != nil {
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
			if err := (StructValueDecoder{val: fieldValue, opt: d.opt, root: false, shared: d.shared}).DecodeTo(
				nbt.ValuePosCompoundOrRoot, encoder, getDecoder,
			); err != nil {
				return err
			}
			continue
		}
		rawTag := fieldType.Tag.Get(d.opt.TagName)
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
		if d.shared.isFirst {
			d.shared.isFirst = false
		} else {
			if err := encoder.NextCompoundElement(); err != nil {
				return err
			}
		}
		decodedElem, err := getDecoder(fieldValue.Interface())
		if err != nil {
			return err
		}
		if err := encoder.WriteCompoundTag(tagName, decodedElem.GetNBType()); err != nil {
			return err
		}
		if err := decodedElem.DecodeTo(nbt.ValuePosCompoundOrRoot, encoder, getDecoder); err != nil {
			return err
		}
	}
	if root {
		if err := encoder.EndComplexCompound(); err != nil {
			return err
		}
	}
	return nil
}

type MapValueDecoder struct {
	val reflect.Value
	opt DecodeOption
}

func (MapValueDecoder) GetNBType() nbt.HighLevelValueType {
	return nbt.ComplexCompound
}

func (d MapValueDecoder) DecodeTo(_ nbt.ValuePos, encoder nbt.EncoderHighLevel, getDecoder GetDecoderFunc) error {
	if d.opt.SortMapKey {
		return d.DecodeToSort(encoder, getDecoder)
	}
	return d.DecodeToUnsort(encoder, getDecoder)
}

func (d MapValueDecoder) WriteElem(key string, elem any, encoder nbt.EncoderHighLevel, getDecoder GetDecoderFunc) error {
	decodedElem, err := getDecoder(elem)
	if err != nil {
		return err
	}
	if err := encoder.WriteCompoundTag(key, decodedElem.GetNBType()); err != nil {
		return err
	}
	if err := decodedElem.DecodeTo(nbt.ValuePosCompoundOrRoot, encoder, getDecoder); err != nil {
		return err
	}
	return nil
}

func (d MapValueDecoder) DecodeToUnsort(encoder nbt.EncoderHighLevel, getDecoder GetDecoderFunc) error {
	if err := encoder.BeginComplexCompound(); err != nil {
		return err
	}
	stringK := true
	if d.val.Type().Key().Kind() != reflect.String {
		stringK = false
	}
	iter := d.val.MapRange()
	// take data
	first := true
	for iter.Next() {
		if first {
			first = false
		} else {
			if err := encoder.NextCompoundElement(); err != nil {
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
		if err := d.WriteElem(key, ev, encoder, getDecoder); err != nil {
			return err
		}
	}
	if err := encoder.EndComplexCompound(); err != nil {
		return err
	}
	return nil
}

func (d MapValueDecoder) DecodeToSort(encoder nbt.EncoderHighLevel, getDecoder GetDecoderFunc) error {
	if err := encoder.BeginComplexCompound(); err != nil {
		return err
	}
	stringK := true
	if d.val.Type().Key().Kind() != reflect.String {
		stringK = false
	}

	// 1. 收集所有 key（以字符串形式作为排序依据）
	keys := make([]string, 0, d.val.Len())
	keyMap := make(map[string]reflect.Value, d.val.Len())

	iter := d.val.MapRange()
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
			if err := encoder.NextCompoundElement(); err != nil {
				return err
			}
		}

		origKey := keyMap[k]
		ev := d.val.MapIndex(origKey).Interface()
		if err := d.WriteElem(k, ev, encoder, getDecoder); err != nil {
			return err
		}
	}
	if err := encoder.EndComplexCompound(); err != nil {
		return err
	}
	return nil
}
