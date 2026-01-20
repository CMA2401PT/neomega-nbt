package neomega_nbt

type TagType byte

const (
	TagEnd       TagType = iota
	TagByte              //int8
	TagShort             //int16
	TagInt               // int32
	TagLong              // int64
	TagFloat             // float32
	TagDouble            // float64
	TagByteArray         // array of uint8
	TagString            // string
	TagList              //list
	TagCompound          // map[string]any
	TagIntArray          // array of int32
	TagLongArray         // array of int64
)
const (
	TagInt8    = TagByte
	TagUint8   = TagByte
	TagInt16   = TagShort
	TagUint16  = TagShort
	TagInt32   = TagInt
	TagUint32  = TagInt
	TagInt64   = TagLong
	TagUint64  = TagLong
	TagFloat32 = TagFloat
	TagFloat64 = TagDouble
)

func (t TagType) String() string {
	switch t {
	case TagEnd:
		return "TagEnd"
	case TagByte:
		return "TagByte"
	case TagShort:
		return "TagShort"
	case TagInt:
		return "TagInt"
	case TagLong:
		return "TagLong"
	case TagFloat:
		return "TagFloat"
	case TagDouble:
		return "TagDouble"
	case TagByteArray:
		return "TagByteArray"
	case TagString:
		return "TagString"
	case TagList:
		return "TagList"
	case TagCompound:
		return "TagCompound"
	case TagIntArray:
		return "TagIntArray"
	case TagLongArray:
		return "TagLongArray"
	default:
		return "TagType(Unknown)"
	}
}

type HighLevelValueType uint8

const (
	Invalid          HighLevelValueType = iota
	TrivialByte                         //int8
	TrivialShort                        //int16
	TrivialInt                          // int32
	TrivialLong                         // int64
	TrivialFloat                        // float32
	TrivialDouble                       // float64
	_                                   // array of uint8, tagByteArray
	TrivialString                       // string
	_                                   //list
	_                                   // map[string]any, tagCompound
	_                                   // array of int32, tagIntArray
	_                                   // array of int64, tagLongArray
	Complex          = 16               //表示 Compound，List， Array 这类长度不确定的情况
	ComplexCompound  = 16 | 10          // 10=tagCompound
	ComplexList      = 16 | 9           // 9=tagList
	ComplexByteArray = 16 | 7           // 7=tagByteArray
	ComplexIntArray  = 16 | 11          // 11=tagIntArray
	ComplexLongArray = 16 | 12          // 12=tagLongArray
)

func (t HighLevelValueType) String() string {
	switch t {
	case Invalid:
		return "Invalid"
	case TrivialByte:
		return "Byte"
	case TrivialShort:
		return "Short"
	case TrivialInt:
		return "Int"
	case TrivialLong:
		return "Long"
	case TrivialFloat:
		return "Float"
	case TrivialDouble:
		return "Double"
	case TrivialString:
		return "String"
	case Complex:
		return "Complex"
	case ComplexCompound:
		return "ComplexCompound"
	case ComplexList:
		return "ComplexList"
	case ComplexByteArray:
		return "ComplexByteArray"
	case ComplexIntArray:
		return "ComplexIntArray"
	case ComplexLongArray:
		return "ComplexLongArray"

	default:
		if t&Complex != 0 {
			return "Complex(Unknown)"
		}
		return "HighLevelValueType(Unknown)"
	}
}

func (t HighLevelValueType) IsValid() bool {
	return t != Invalid
}

func (t HighLevelValueType) IsTrivial() bool {
	return (t >= TrivialByte && t <= TrivialDouble) || t == TrivialString
}

func (t HighLevelValueType) IsComplex() bool {
	return t&Complex == Complex
}

func (t HighLevelValueType) IsArray() bool {
	return t == ComplexByteArray || t == ComplexIntArray || t == ComplexLongArray
}

func (t HighLevelValueType) IsList() bool {
	return t == ComplexList
}

func (t HighLevelValueType) IsCompound() bool {
	return t == ComplexCompound
}

type ValuePos uint8

const (
	ValuePosCompoundOrRoot ValuePos = iota
	ValuePosArray
	ValuePosList
)

func (p ValuePos) InArray() bool {
	return p == ValuePosArray
}

func (p ValuePos) InList() bool {
	return p == ValuePosList
}

type Flag string

const (
	// 允许 list 具有不同的类型
	FLAG_SUPPORT_TYPE_ERASE = Flag("FLAG_SUPPORT_TYPE_ERASE")
	// 会忽略实际传入的元素数量
	FLAG_OMIT_LIST_LEN = Flag("FLAG_OMIT_LIST_LEN")
	// 会忽略实际传入的元素数量
	FLAG_OMIT_ARRAY_LEN = Flag("FLAG_OMIT_ARRAY_LEN")
)
