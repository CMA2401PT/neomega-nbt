# neomega-nbt

`neomega-nbt` 是一个 Go 数据转换库，用来连接 Minecraft 的二进制 NBT、文本 SNBT 和 Go 对象。NBT 是一种带类型的树形数据：节点可以是数值、字符串、列表、数值数组，也可以是由名字索引子节点的 Compound。SNBT 用文本表达这些数据，便于查看和编辑。

这个库把转换拆成两件事：**读取源数据，发出一串操作；接收这些操作，生成目标数据。** 例如，把 NBT 转成 SNBT 时，二进制读取器可以直接驱动文本生成器，不必先构造一棵 `map[string]any` 数据树。常用的组合已经放在 `presets/v1`，普通调用者可以直接使用。

## 为什么要拆开读取与生成

假设我们先实现了 Go 对象与 NBT 的互转，接着又需要 SNBT、不同字节序，以及业务自己的对象类型。如果每一种转换都单独实现，遍历树、识别类型、处理容器等逻辑就会重复出现。

这里选择让各模块共享一套操作约定。读取端只描述“读到了什么”，生成端只负责“怎样保存它”。因此，新写一个读取端，就可以接到已有的生成端；新写一个生成端，也可以接收已有读取端的数据。

以 `{level: 3}` 为例，先忽略类型和错误处理，过程可以理解为：

```text
读取源数据                  接收端生成目标数据
开始 Compound       ────→   建立对象，或者写入左花括号
字段 level，值为 3    ────→   保存字段，或者写入文本
结束 Compound       ────→   完成对象，或者写入右花括号
```

代码中的 `decoder-*` 是读取端，`encoder-*` 是生成端。这里的名称是相对于数据流而言的：`decoder-go-object` 读取 Go 对象，`encoder-go-object` 构造 Go 对象。这也解释了为什么“把 NBT 解码进 Go 对象”会用到 Go 对象的 **encoder**。

## 为什么有两套操作接口

共享操作不意味着所有格式都适合相同粒度的接口。二进制格式关心类型、长度和字节顺序；构造 Go 对象或输出文本时，更希望直接知道“容器开始了”“下一个元素到了”“容器结束了”。

因此，根包定义了两个接收接口：

- `Encoder` 是低层接口，接收 `WriteType`、`WriteTag`、`WriteLen`、`WriteInt32` 等顺序操作。它接近二进制 NBT 的布局。
- `EncoderHighLevel` 是高层接口，接收容器的 Begin / Next / End，以及带位置信息的值。它适合构造对象和生成文本。`ValuePos` 表示当前值位于根或 Compound、List、还是数值数组中。

例如，低层接口写一个有名字的整数，需要按顺序调用 `WriteType(TagInt)`、`WriteTag("level")`、`WriteInt32(3)`。高层接口则先报告字段的名字与类型，再交付整数值。

为了让两种接口可以互接，库提供两个适配器。适配器只翻译操作，不建立完整的中间数据树：

| 数据路径 | 连接方式 |
| --- | --- |
| Go 对象 → SNBT | `decoder-go-object` → `encoder-snbt`，直接使用高层操作 |
| NBT → Go 对象或 SNBT | `decoder-nbt` → `encoder-shim` → 对应高层生成端 |
| Go 对象或 SNBT → NBT | 对应高层读取端 → `decoder-shim` → `encoder-nbt` |
| 一种 NBT 编码 → 另一种 NBT 编码 | `decoder-nbt` → `encoder-nbt`，直接使用低层操作 |

`encoder-shim` 用状态机和栈补出容器边界。比如，低层读完列表声明的元素数量后，它会发出高层的列表结束操作。`decoder-shim` 做相反的工作，把高层容器事件展开为低层操作。

## 先完成一次转换

理解这些模块有助于扩展，但使用现成转换时无需手工连接它们。`presets/v1` 已经负责选择编码器和适配器。

项目的 `go.mod` 声明 Go 1.25.5，模块路径为 `neomega-nbt`。在旁边新建一个示例项目时，可以用本地替换接入：

```sh
# 在存放 neomega-nbt 的父目录执行。
mkdir nbt-demo
cd nbt-demo
go mod init example.com/nbt-demo
go mod edit -require=neomega-nbt@v0.0.0
go mod edit -replace=neomega-nbt=../neomega-nbt
```

把下面的程序保存为 `main.go`，运行 `GOWORK=off go run .`。这个程序把 Go 结构体写成小端 NBT，再读回结构体，最后输出便于查看的 SNBT。

```go
package main

import (
    "fmt"

    preset "neomega-nbt/presets/v1"
)

func main() {
    type Block struct {
        Name  string `nbt:"name"`
        Level int32  `nbt:"level"`
    }
    input := Block{Name: "stone", Level: 3}

    data, err := preset.EncodeToNBTBytes(input, preset.OptEncodeToNBT{
        Encoding: preset.EncodingLittleEndian,
    })
    if err != nil {
        panic(err)
    }

    var output Block
    err = preset.DecodeFromNBTBytes(data, &output, preset.OptDecodeFromNBT{
        Encoding: preset.EncodingLittleEndian,
    })
    if err != nil {
        panic(err)
    }

    text, err := preset.EncodeToSNBT(output, preset.OptEncodeToSNBT{})
    if err != nil {
        panic(err)
    }
    fmt.Println(text)
}
```

输出：

```text
{"name": "stone", "level": 3}
```

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

Go 与 NBT 之间另有 `EncodeToNBT(data, writer, opt)` 和 `DecodeFromNBT(reader, ptr, opt)`，用于 `io.Writer`、`io.Reader`。流读取只消耗一个根值，可以连续读取多个根值。字节入口也只解析一个根值，不检查其后的内容。

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

### 为什么经过 SNBT 后不一定得到相同字节

SNBT 没有二进制 NBT 的根标签名，所以 `NBTToSNBT` 会丢弃它；`SNBTToNBT` 使用 `OptSNBTToNBT.RootTag` 指定新名字，默认是空字符串。Go 对象接收端同样忽略根标签名，写入端通过源选项 `RootTag` 指定名字。

SNBT 的 `[]` 也不记录空列表的元素类型。因此，有类型的空列表经过 SNBT 后只剩“空列表”。如果需要保留根名、标签类型和 Compound 字段顺序，应直接使用 `ConvertNBT`。它翻译二进制事件，不经过 Go map 或文本表示。

转换函数遇到错误会立即返回。写入流或填充 Go 对象时，失败之前可能已经产生部分结果；如需成功后才替换原对象，应先解码到临时变量。

## 扩展自己的 Go 类型

现成的类型映射不一定符合业务含义。例如，Go 中的距离可能用浮点数表示，而保存格式要求以千分之一格为单位的整数。这时只需覆盖这个类型的转换方式，无需重写 NBT 或 SNBT 编码器。

源值实现 `decoder-go-object.DecoderOverride`，报告自己的 NBT 类型并向接收端交付数据。目标指针实现 `encoder-go-object.EncoderOverride`，提供接收这个 NBT 类型的方法。下面的 `Distance` 在 Go 中是浮点距离，在 NBT/SNBT 中是 Int：

```go
package main

import (
    "fmt"

    nbt "neomega-nbt"
    decoder_go "neomega-nbt/decoder-go-object"
    encoder_go "neomega-nbt/encoder-go-object"
    preset "neomega-nbt/presets/v1"
)

type Distance float64

func (Distance) GetNBType() nbt.HighLevelValueType {
    return nbt.TrivialInt
}

func (d Distance) DecodeTo(
    pos nbt.ValuePos,
    encoder nbt.EncoderHighLevel,
    _ decoder_go.GetDecoderFunc,
) error {
    return encoder.WriteInt32(int32(d*1000), pos)
}

func (d *Distance) GetReciver(
    valueType nbt.HighLevelValueType,
) (encoder_go.ValueReciver, error) {
    if valueType != nbt.TrivialInt {
        return encoder_go.ValueReciver{}, fmt.Errorf("Distance expects Int, got %v", valueType)
    }
    return encoder_go.ValueReciver{
        TrivialValReciver: func(_ any, value int64) error {
            *d = Distance(value) / 1000
            return nil
        },
    }, nil
}

func main() {
    text, err := preset.EncodeToSNBT(Distance(1.25), preset.OptEncodeToSNBT{})
    if err != nil {
        panic(err)
    }
    var distance Distance
    if err := preset.DecodeFromSNBT(text, &distance, preset.OptDecodeFromSNBT{}); err != nil {
        panic(err)
    }
    fmt.Println(text, distance) // 1250 1.25
}
```

这里保留了接口中的 `GetReciver`、`ValueReciver` 拼写。示例里的精度和范围由 `Distance` 的业务约定负责；库只负责传递它声明的 Int。

### 自定义容器为什么还需要构造器

简单值可以一次交付，但 map、列表或业务容器通常要逐个填充。接收端因此通过 `ValueReciver` 提供 `CompoundConstructor`、`ListConstructor` 或 `ArrayConstructor`。构造器决定子值写到哪里，以及什么时候把暂存值放进容器。

普通非空容器的约定是：

1. Compound 通过 `GetTagPointer` 提供当前字段的目标。List 和 Array 通过 `GetReusableElementPointer` 提供可以重复使用的元素目标。
2. 当下一个元素到来时，`OnNextElement` 提交前一个元素。接收端随后把下一个值写入目标。
3. `OnCompoundFinish`、`OnListFinish` 或 `OnArrayFinish` 提交最后一个元素，并完成容器。

对于空 Compound，仍会调用开始和结束；构造器需要知道没有取过任何字段目标。空列表与空数组则走下面的专用路径，不会提交暂存元素。

## 扩展操作流或新的格式

当变化涉及整个格式，而不只是一个 Go 类型时，就应该直接接入操作接口。一个新读取端遍历自己的数据并调用接收接口；一个新生成端实现接收接口。若两端的层级不同，再选择前文的适配器。

高层容器通常按以下顺序工作。`Next` 表示两个元素之间的分隔，而不是一次值写入：

```text
BeginComplexList(元素类型, 长度, 是否允许类型变化)
    第一个元素
NextListElement(下一个元素类型)
    第二个元素
EndComplexList()
```

读取端必须传播接收端返回的错误，并报告与实际数据一致的类型。对于自定义 Go 容器的读取，`DecodeTo` 收到的 `getChildDecoder` 可以继续处理子值，包括子值自己的 override。

### 长度未知时如何继续

读取文本时，通常还没走到 `]` 就要开始输出列表。但二进制 NBT 要求先写长度。接收端通过 `HasFlag` 表达自己是否能接受这些情况：

- `FLAG_OMIT_LIST_LEN`：允许列表长度只是占位值，实际数量由收到的元素决定。
- `FLAG_OMIT_ARRAY_LEN`：对数值数组提供同样的能力。
- `FLAG_SUPPORT_TYPE_ERASE`：允许列表元素的实际类型变化，后续类型由 `NextListElement` 报告。

SNBT 读取端会根据接收端能力决定是否先扫描以计算长度。因此，操作流能避免完整的中间对象树，但并不保证所有转换都只扫描一次。

### 已知为空时为什么要用专用算子

如果允许未知长度，`BeginComplexList(type, 0, ...)` 的 `0` 就可能只是占位值，不能同时用它断言容器为空。专用空值算子让接收端明确知道“没有任何元素要提交”，并保留已有的非空容器提交约定：

| 高层操作 | 含义 | Go 容器构造器收到的操作 |
| --- | --- | --- |
| `BeginAndEndComplexEmptyList()` | 空列表，元素类型未知 | `NonTypedEmptyList()` |
| `BeginAndEndComplexEmptyTypedList(type)` | 空列表，元素类型已知 | `TypedEmptyList(type)` |
| `BeginAndEndComplexEmptyArray(type)` | 空数值数组，元素类型为 Byte、Int 或 Long | `EmptyArray(type)` |

专用算子一次完成容器，后面不再追加该容器的 Begin / Next / End。`encoder-shim` 会把二进制的零长度列表和数组转换成对应空值算子。自定义高层生成端和 Go 容器构造器也需要实现相应方法。

## 从哪里继续阅读代码

理解数据流后，可以按需要进入对应实现，而不必从所有接口开始读：

- 想组合现有转换：从 [`presets/v1`](presets/v1) 看各路径的连接方式。
- 想添加 Go 类型映射：从 [`decoder-go-object/override.go`](decoder-go-object/override.go) 和 [`encoder-go-object/override.go`](encoder-go-object/override.go) 看扩展接口。
- 想添加新格式：从 [`nbt.go`](nbt.go) 看操作约定，再对照 [`decoder-shim`](decoder-shim) 与 [`encoder-shim`](encoder-shim)。
- 想添加二进制编码：分别实现 [`decoder-nbt/encoding.go`](decoder-nbt/encoding.go) 和 [`encoder-nbt/encoding.go`](encoder-nbt/encoding.go) 的 `Encoding` 接口。它们把整数、浮点数、字符串的字节布局与树的遍历分开。`seq-io` 提供完整读写一段字节的接口与适配器。

## 验证修改

转换修改最容易影响其他组合，因此测试不仅检查单个值，还覆盖三种二进制编码、文本与对象的往返、嵌套容器、空值算子、自定义类型和错误传播。固定字节样例用于避免读写两端犯同一个错误而让往返测试误通过。

在仓库根目录运行：

```sh
GOWORK=off go test ./...
GOWORK=off go vet ./...
```

`GOWORK=off` 让命令只使用本模块，避免父目录工作区中的其他模块影响测试。README 的基础用法与自定义类型示例也保存在 [`presets/v1/example_test.go`](presets/v1/example_test.go)，作为可运行的示例测试。
