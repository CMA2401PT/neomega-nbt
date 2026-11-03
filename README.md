# neomega-nbt

`neomega-nbt` 是一个 Go 数据转换库，用来连接 Minecraft 的二进制 NBT、文本 SNBT 和 Go 对象。NBT 是一种带类型的树形数据：节点可以是数值、字符串、列表、数值数组，也可以是由名字索引子节点的 Compound。SNBT 用文本表达这些数据，便于查看和编辑。

这个库把转换拆成两件事：**读取源数据，发出一串操作；接收这些操作，生成目标数据。** 例如，把 NBT 转成 SNBT 时，二进制读取器可以直接驱动文本生成器，不必先构造一棵 `map[string]any` 数据树。常用的组合已经放在 `presets/v1`，普通调用者可以直接使用。

## 运行第一个例子

模块路径是 `neomega-nbt`，要求 Go 1.25.5。克隆工作区后，在本模块目录运行：

```sh
go run ./examples/roundtrip
```

[`examples/roundtrip/main.go`](examples/roundtrip/main.go) 把带 `nbt` 字段标签的 Go 结构体写成小端 NBT，再读回结构体并显示 SNBT。命令输出：

```text
{"name": "stone", "level": 3}
```

要修改带根名称的未知字段文档，运行第二个例子：

```sh
go run ./examples/document
```

[`examples/document/main.go`](examples/document/main.go) 创建大端文档，再解码、修改并写回，分别演示 Int List 和 IntArray。命令输出：

```text
根名称：Level
{"array": [I; 1, 2, 3], "level": 4, "list": [1, 2, 3]}
```

在工作区以外的项目使用本地源码时，在调用方的 `go.mod` 中添加 `require neomega-nbt v0.0.0` 和指向本模块目录的 `replace neomega-nbt => ../neomega-nbt`，路径按实际位置调整。

传入 Go 接收目标时，要提供指向目标的指针。已有结构体可以限定字段类型；没有预先定义的数据结构时，可以传入 `&value`，其中 `value` 声明为 `any`。

## 选择需要的转换

一次往返通常只是起点。需要查看二进制内容、把文本写回二进制，或者更换编码时，可以直接选择下面的入口。表中的 NBT 数据都是**未压缩的单个根值**；文件压缩层和文件头由调用者处理。

| 需求 | `presets/v1` 入口 |
| --- | --- |
| Go → NBT 字节 | `EncodeToNBTBytes(data, OptEncodeToNBT) ([]byte, error)` |
| NBT 字节 → Go | `DecodeFromNBTBytes(data, ptr, OptDecodeFromNBT) error` |
| Go → SNBT 文本 | `EncodeToSNBT(data, OptEncodeToSNBT) (string, error)` |
| SNBT 文本 → Go | `DecodeFromSNBT(text, ptr, OptDecodeFromSNBT) error` |
| NBT 字节 → SNBT 文本 | `NBTToSNBT(data, encoding) (string, error)` |
| SNBT 文本 → NBT 字节 | `SNBTToNBT(text, OptSNBTToNBT) ([]byte, error)` |
| 更换 NBT 二进制编码 | `ConvertNBT(data, from, to) ([]byte, error)` |

Go 与 NBT 之间另有 `EncodeToNBT(data, writer, opt)` 和 `DecodeFromNBT(reader, ptr, opt)`，用于 `io.Writer`、`io.Reader`。流读取只消耗一个根值，可以连续读取多个根值。`DecodeFromNBTBytes` 也只解析一个根值，不检查其后的内容。

NBT 的编码必须显式指定，零值会返回错误：

- `EncodingLittleEndian`：定长整数使用小端序。
- `EncodingBigEndian`：定长整数使用大端序。
- `EncodingNetworkLittleEndian`：Int、Long 和容器长度使用 ZigZag 变长整数；字符串长度使用无符号变长整数；Short 和浮点数使用小端序。

编码选择应来自数据来源或外层协议。旧拼写 `EncodingLittleEdian`、`EncodingBigEdian`、`EncodingNetworkLittleEdian` 和旧选项名 `OptDecodeToNBT` 保留为别名。

SNBT preset 要求输入恰好包含一个值，允许前后空白。若要解析一段文本的前缀并继续处理剩余内容，可以使用底层 `decoder-snbt.DecodeSnbtTo`；它返回未消费的文本。

## 转换时，哪些信息需要自己决定

格式之间并非所有信息都能一一对应。了解这些差异，才能决定是直接转格式，还是先读入业务对象再修改。

### Go 数字和容器如何对应 NBT

NBT 的整数有明确位宽。需要稳定的数据类型时，优先使用 `int8`、`int16`、`int32`、`int64`；`int`、`uint` 默认作为 Int 写出，设置源端 `IntAsInt64` 后作为 Long。无符号数沿用同位宽整数的位模式，例如 `uint8(255)` 写为 Byte 的 `-1`。`bool` 对应 Byte 的 `0` 和 `1`。

NBT 的 List 与数值数组是不同的类型。List 统一声明元素类型，可以嵌套其他容器；数值数组专门保存 Byte、Int 或 Long。

- Go slice 默认写成 List。设置源端 `SliceAsArray` 后，`int8/uint8`、`int32/uint32`、`int64/uint64` 的 slice 写成对应数值数组，bool slice 写成 ByteArray。
- 对应元素类型的 Go 定长数组直接写成数值数组，例如 `[3]int32` 写成 IntArray。
- map 和结构体写成 Compound。结构体默认使用 `nbt` 标签命名字段，支持 `-` 和 `omitempty`；`omitempty` 按 Go 零值判断。

源端选项属于 `decoder-go-object.DecodeOption`，因为它控制如何读取 Go 对象。`OptEncodeToSNBT` 是它的别名，`OptEncodeToNBT` 则嵌入了它。例如，需要稳定的 map 输出顺序时，设置 `SortMapKey: true`；默认遍历顺序不固定。

接收端选项属于 `encoder-go-object.EncodeOption`。`OptDecodeFromSNBT` 是它的别名，`OptDecodeFromNBT` 嵌入了它。接收为 `any` 时，Compound 构造成 `map[string]any`；NBT 中有明确元素类型的 List 尽量构造成对应 slice；数值数组默认构造成 Go 定长数组。设置 `ArrayAsSlice` 可改为 slice，设置 `UseInt8Sequence` 可让动态构造的字节序列使用 `int8`，而不是默认的 `uint8`。

读取 SNBT 时，接收端若允许列表元素变换类型，列表可以构造成 `[]any`。例如 `[1, "two"]` 可以读入 `[]any`，但不能直接写成二进制 NBT 的 List，因为二进制列表要求所有元素具有同一个 NBT 类型。

### 编辑带根名称的文档

编辑未知字段较多的 NBT 时，需要同时保留根名称和各字段的标签类型。`presets/v1.Document` 将根名称与 Compound 放在一起，`DecodeDocument(data, encoding)` 要求字节中恰好包含一个 Compound 根。`ReadDocument(reader, encoding)` 只消费一个根，适合连续记录。修改 `document.Value` 后，用 `EncodeDocument(document, encoding)` 写回。连续根记录可以使用 `DecodeDocuments` 和 `EncodeDocuments` 批量读写。

Document 使用默认的 Go 类型映射：slice 表示 List，数值定长数组表示 NBT Array。因此，同一个文档中的 Int List 和 IntArray 可以分别保留。`Array(values)` 从数值 slice 构造显式数组，`ArrayValues[T](value)` 将解码后的数组复制为 slice。`Compounds(value)` 读取 Compound 列表，也接受没有元素类型的空列表。

完整的创建、解码、修改和写回流程见 [`examples/document/main.go`](examples/document/main.go)。

重新编码保留根名称和标签类型；Go map 的字段顺序可以变化。需要保持 Compound 字段顺序的纯编码转换，使用 `ConvertNBT`。

### 为什么经过 SNBT 后不一定得到相同字节

SNBT 没有二进制 NBT 的根标签名，所以 `NBTToSNBT` 会丢弃它；`SNBTToNBT` 使用 `OptSNBTToNBT.RootTag` 指定新名字，默认是空字符串。通用 Go 对象接收端忽略根标签名，写入端通过源选项 `RootTag` 指定名字；`Document` 入口会另行保存和恢复根名称。

SNBT 的 `[]` 也不记录空列表的元素类型。因此，有类型的空列表经过 SNBT 后只剩“空列表”。如果需要保留根名、标签类型和 Compound 字段顺序，应直接使用 `ConvertNBT`。它翻译二进制事件，不经过 Go map 或文本表示。

转换函数遇到错误会立即返回。写入流或填充 Go 对象时，失败之前可能已经产生部分结果；如需成功后才替换原对象，应先解码到临时变量。

## 自定义类型与格式

业务类型可以实现 [`decoder-go-object.DecoderOverride`](decoder-go-object/override.go)，决定怎样从 Go 值发出 NBT 操作；目标指针可以实现 [`encoder-go-object.EncoderOverride`](encoder-go-object/override.go)，决定怎样接收操作。名称按数据流命名，因此“把 NBT 读入 Go”会用到 Go 对象的 encoder。可运行的自定义距离类型示例见 [`presets/v1/example_test.go`](presets/v1/example_test.go)。在模块目录执行：

```sh
go test ./presets/v1 -run Example_customValue -v
```

需要接入整个新格式时，根包的 `Encoder` 定义低层顺序操作，`EncoderHighLevel` 定义高层容器事件，两个 shim 负责互相适配。公共协议、空容器约定、源码分层和验收命令见 [AGENTS.md](AGENTS.md)。
