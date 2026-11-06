# 维护 neomega-nbt

本模块负责在 Go 对象、二进制 NBT 和文本 SNBT 之间传递带类型的数据。读取端产生操作，生成端消费操作，preset 负责把它们组合成使用者入口。使用方法见 [README.md](README.md)，工作区命令与依赖规则继承 [根 AGENTS.md](../AGENTS.md)。

修改应落在负责该语义的层：格式解析在读取端，目标表示在生成端，组合入口在 `presets/v1`。本模块是基础库，不应依赖方块、转换应用或渲染模块。

## 结构与边界

| 位置 | 维护职责 |
| --- | --- |
| `nbt.go`、`consts.go` | 公共操作协议、类型与能力标志；改动会影响所有读取端和生成端。 |
| `decoder-*`、`encoder-*` | 各种表示的读取与生成；Go 对象方向按数据流命名。 |
| `decoder-shim`、`encoder-shim` | 连接高层容器事件与低层顺序操作。 |
| `seq-io` | 二进制读取、写入适配。 |
| `presets/v1` | 公开的现成转换、文档编辑入口与组合回归。 |
| `examples` | 可直接运行的调用教程；与 README 的命令和输出同步。 |

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

## 自定义容器为什么还需要构造器

简单值可以一次交付，但 map、列表或业务容器通常要逐个填充。接收端因此通过 `ValueReciver` 提供 `CompoundConstructor`、`ListConstructor` 或 `ArrayConstructor`。构造器决定子值写到哪里，以及什么时候把暂存值放进容器。

普通非空容器的约定是：

1. Compound 通过 `GetTagPointer` 提供当前字段的目标。List 和 Array 通过 `GetReusableElementPointer` 提供可以重复使用的元素目标。
2. 当下一个元素到来时，`OnNextElement` 提交前一个元素。接收端随后把下一个值写入目标。
3. `OnCompoundFinish`、`OnListFinish` 或 `OnArrayFinish` 提交最后一个元素，并完成容器。

对于空 Compound，仍会调用开始和结束；构造器需要知道没有取过任何字段目标。空列表与空数组则走下面的专用路径，不会提交暂存元素。

## 大数组的可选批量接收

`ByteArrayConsumer` 让二进制 ByteArray 接收器自行分段消费原始字节，避免逐字节操作的调用成本。decoder-nbt 在长度事件之前查询它，encoder-shim 和 encoder-go-object 将请求转发到当前目标。handled 为 true 时，替代长度、元素与结束事件；接收器必须恰好读取声明的字节数，并完成自己的结果。shim 随后恢复父容器。handled 为 false 时不得读取或改变状态，普通事件路径继续执行。

这是可选接口，不扩大所有 Encoder 的必实现方法集合。修改转发时要覆盖根数组、Compound 字段、List 内数组、空数组、回退、三种字节序、截断和接收器错误；回归测试位于 `presets/v1/stream_array_test.go`。

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

## 变更约束与验收

保留已有编码常量和选项别名，包括旧拼写。改变类型映射、空容器事件或适配器时，同时检查三种二进制编码，以及对象、文本和二进制之间受影响的组合。往返测试不能替代固定字节样例，因为读写两端可能同时犯错。

`Document` 保留根名称与数值数组类型；普通 Go 对象入口忽略根名称。`DecodeDocument` 拒绝尾随字节，流式 `ReadDocument` 只消费一个根。修改这些入口时保持两种边界的区别。

在工作区根目录运行：

```sh
./check.sh neomega-nbt
```

在模块目录运行使用者示例：

```sh
go run ./examples/roundtrip
go run ./examples/document
```

检查入口执行 `go vet`、`go test` 和脱离工作区编译。普通测试包含 `presets/v1/example_test.go` 中的基础转换和自定义类型示例；文档中的固定输出必须来自对应的实际运行。
