package neomega_nbt

// 我们的目标是将 Decode 和 Encoder 描述为一系列操作序列
// 相比 EncoderHighLevel， Encoder 更加的序列化，语义也更为简单
// 其在构造复杂类型中不足，但是很适合序列化操作，例如大端转小端或者 snbt 操作
// 例如，对于一个普通的值，描述为
// WriteType(tagInt), WriteTag("tag"), WriteInt32(int32)/WriteUint32(uint32)
// 对于 List，Compound 则可以消除循环结构(构造不确定长度的中间数据或结果数据被移动到外部)
// ...,WriteType(tagList),WriteEmptyListMark()/ WriteType(tagInt),WriteLen(len),WriteInt(int32),...,WriteInt(int32)
// ...,WriteType(tagCompound),WriteType(tagInt),WriteTag("tag"),...,WriteEnd()
// 显然，并非所有的操作序列都能接受
// encoder-shim 用状态机检查操作顺序并生成高层事件；配合 VoidEncoderHighLevel 可只检查而不生成数据。
type Encoder interface {
	WriteType(TagType) error
	WriteTag(tag string) error
	// 当 WriteType(tagList) 时，特殊情况为空列表且不知类型，应该紧接着调用
	WriteEmptyListMark() error
	// 非特殊情况应该先写入 WriteType 再写入 WirteLen，
	// WriteType(tagByteArray, tagIntArray, tagLongArray) 同样需要 WriteLen
	// 标准为 int32，但是考虑到部分情况会超过 int32，且负值实际无意义
	WriteLen(len uint32) error
	WriteEnd() error
	// 写入常规值
	WriteTrivial
}

// 显然， Decode 就是这个的逆过程
// 即可以将一个东西变为对 Encoder 的一系列操作，故有：
type Decoder interface {
	DecodeTo(Encoder)
}

// 或者直接定义 Decode 函数
// func DecodeTo(Encoder)
// 例如 DecodeGoObjectTo(SNBTEncoder) 完成 go 对象到 snbt 的序列化

type WriteTrivial interface {
	// 写入常规值
	WriteInt8(int8) error
	WriteInt16(int16) error
	WriteInt32(int32) error
	WriteInt64(int64) error
	WriteFloat32(float32) error
	WriteFloat64(float64) error
	WriteString(string) error
}

// 相比 Encoder，EncoderHighLevel 消除了更多琐碎的情况， 并更加语义化(节点化)
// 并把 ValueType 分为了简单值（数字，字符串） 和复杂值 （Compound，List， Array） 两种
// 每次写入 BasicValueType 之后，若为简单值，则紧接着写入值
// 若为复杂值则 .BeginComplex (复杂值类型和说明) 此时需要压栈
// 接着 .Write,.Write,... 等一串操作
// 最后 .EndComplex 弹栈
type EncoderHighLevel interface {
	// 和WriteCompoundTag区别不大，但是因为根tag普遍会被特殊处理，所以单独拎出来
	WriteRootTag(tag string, valueType HighLevelValueType) error

	// 接下来会是一个 Array，其类型为 valueType in [TagByte,TagInt,TagLong]
	// 第一个 array 元素前被调用；已知为空时使用专用空数组算子。
	// 紧接其后的是一个 WriteTrivialWithPos，类型不会被报告
	BeginComplexArray(valueType HighLevelValueType, len int) error
	// 在每两个 array 元素间被调用，即逗号位置
	// 紧接其后的是一个 WriteTrivialWithPos，类型不会被报告
	NextArrayElement() error
	// 最后一个 array 元素后被调用
	EndComplexArray() error

	// 接下来会是一个 Compound
	// 紧接其后的是一个 WriteCompoundTag
	BeginComplexCompound() error
	// 当前 Compound 的tag 和其值类型
	WriteCompoundTag(tag string, valueType HighLevelValueType) error
	// 在每两个 compount 键值间被调用，即逗号位置
	// 紧接其后的是一个 WriteCompoundTag
	NextCompoundElement() error
	// 最后一个 compount 键值后被调用
	EndComplexCompound() error

	// 特殊情况，是一个空的列表，类型也未知
	BeginAndEndComplexEmptyList() error
	// 已知元素类型的空列表，保留其类型；不再调用 Begin/EndComplexList。
	BeginAndEndComplexEmptyTypedList(valueType HighLevelValueType) error
	// 空数值数组，valueType 为 Byte、Int 或 Long；不再调用 Begin/EndComplexArray。
	BeginAndEndComplexEmptyArray(valueType HighLevelValueType) error
	// 接下来会是一个 List，其中元素类型定义与 valueType 一致
	// 类型会通过 NextListElement 报告
	// 紧接其后的是一个 WriteTrivialWithPos， 但是
	// 如果 valueType 为 Compound 或者 Array 或者 List，则会是
	// BeginComplexCompound , BeginComplexArray, BeginComplexList(BeginAndEndComplexEmptyList)
	// 另外，由于 snbt 的存在以及 java 事实上的类型擦除，所以接下来的数据可能与报告类型不同
	// 若可以接受此类情况，应当在 HasFlag 中标记 FLAG_SUPPORT_TYPE_ERASE
	// 此时，传入的类型为第一个 element 的类型，并设置 type erase 为 true
	// 在一些情况下，传入元素数量可能与开始时报告的不同，
	// 若支持这种情况（比如 go 的 slice 可以动态扩容）
	// 应当在 HasFlag 中标记 FLAG_OMIT_LIST_LEN
	// 已知为空时使用专用空列表算子；未知长度仍可在接收端允许时报告 0。
	BeginComplexList(valueType HighLevelValueType, len int, typeErase bool) error
	// 在每两个 list 元素间被调用，即逗号位置
	// 下一个元素的类型被报告，因为可能类型发生变化
	// 若无法接受类型变化则应当回 error
	NextListElement(valueType HighLevelValueType) error
	// 最后一个 list 元素后被调用
	EndComplexList() error
	// 一些情况下，Encoder 可以支持一些特殊情况或者放松对操作的约束
	// 例如，当变为 go 或 snbt 或一些其它对象时，不一定需要提前知道 list/array 的长度
	// 或者，允许 list 实际具有不同的类型
	// 我们将这些情况通过 Flag 接口标识其对特性的支持
	HasFlag(flag Flag) bool

	// 写入常规值
	WriteTrivialWithPos
}

// 类似的， DecoderHighLevel 就是 EncoderHighLevel 的逆过程
// 即可以将一个东西变为对 EncoderHighLevel 的一系列操作，故有：
type DecoderHighLevel interface {
	HighLevelDecodeTo(EncoderHighLevel)
}

// 为了抹平两者之间的差异，我们定义了
// Decoder -> Encoder
// DecoderHL -> EncoderHL
// Decoder -> EncoderShim -> EncoderHL (语义提升，节点化，从单个操作语义较少的序列提升为语义较多的序列)
// DecoderHL -> DecoderShim -> Encoder (语义下降，序列化，从单个操作语义较多的序列下降为语义较少的序列)

type WriteTrivialWithPos interface {
	// 写入常规值，但是附加了是否在 list，array 中的信息
	// 这些信息在很多时候很有用，因为对于 list，array 针对性的优化很好做
	WriteInt8(int8, ValuePos) error
	WriteInt16(int16, ValuePos) error
	WriteInt32(int32, ValuePos) error
	WriteInt64(int64, ValuePos) error
	WriteFloat32(float32, ValuePos) error
	WriteFloat64(float64, ValuePos) error
	WriteString(string, ValuePos) error
}
