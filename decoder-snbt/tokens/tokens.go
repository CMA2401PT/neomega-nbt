package tokens

type Status int

const (
	RET_FEED_MORE = Status(0) //b00, 代表现在尚不知道此 char 是否属于这个 token
	RET_TERM_FAIL = Status(1) //b01，代表这个 char 不属于此 token，但是前面的 char 可能属于这个token
	RET_OK        = Status(2) //b10，代表这个 char 和前面的 char 都属于这个 token，但是后面的 char 也可能属于此 token
	RET_TERM_OK   = Status(3) //b11，代表这个 char 和前面的 char 都属于这个 token，后面的 char 不属于此 token
	STATUS_RET    = Status(3)
)

type Token interface {
	Feed(b byte) Status
	AcceptZero() bool // 接收0长token
}
