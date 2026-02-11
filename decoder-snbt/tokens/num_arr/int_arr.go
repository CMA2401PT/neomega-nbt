package num_arr

import (
	"neomega-nbt/decoder-snbt/tokens"
)

type NumArray[I ~int32 | ~int8 | ~int64 | ~uint32 | ~uint8 | ~uint64] struct {
	state byte
	// 0x1 activate -> is reading a number
	// 0x2 upcoming -> should follow a number
	// 0x4 neg -> current number is negative
	// 0x8 true
	// 0x10 (16) false
	// 1,2 << 5 rue, 1,2,3 <<5 alse
	currD   I
	Reciver func(I) error // 存储解析后的数字
}

func (f *NumArray[I]) Feed(b byte) tokens.Status {
	if b == ']' {
		// 最后一个数字后面允许有多余的逗号
		// if f.state&0x2 != 0 {
		// 	// should have number append, but not
		// 	return tokens.RET_TERM_FAIL
		// }
		if f.state&(0x8+0x10) != 0 {
			// in reading true/false
			return tokens.RET_TERM_FAIL
		}
		if f.state&0x1 != 0 {
			// has a val
			if f.state&0x4 == 0 {
				if err := f.Reciver(f.currD); err != nil {
					return tokens.RET_TERM_FAIL
				}
			} else {
				if err := f.Reciver(-f.currD); err != nil {
					return tokens.RET_TERM_FAIL
				}
			}
		}
		return tokens.RET_TERM_OK
	} else if b == 't' {
		if f.state&0x1 != 0 {
			return tokens.RET_TERM_FAIL
		}
		f.currD = 1
		f.state = (0x8 + 0x1)
		return tokens.RET_FEED_MORE
	} else if b == 'f' {
		if f.state&0x1 != 0 {
			return tokens.RET_TERM_FAIL
		}
		f.currD = 0
		f.state = (0x10 + 0x1)
		return tokens.RET_FEED_MORE
	} else if f.state&(0x8+0x10) != 0 {
		// in reading true/false
		switch b {
		case 'r':
			if f.state != (0x8 + 0x1) {
				return tokens.RET_TERM_FAIL
			}
			f.state = (0x8 + 0x1) + (0x1 << 5)
			return tokens.RET_FEED_MORE
		case 'u':
			if f.state != (0x8+0x1)+(0x1<<5) {
				return tokens.RET_TERM_FAIL
			}
			f.state = (0x8 + 0x1) + (0x2 << 5)
			return tokens.RET_FEED_MORE
		case 'a':
			if f.state != (0x10 + 0x1) {
				return tokens.RET_TERM_FAIL
			}
			f.state = (0x10 + 0x1) + (0x1 << 5)
			return tokens.RET_FEED_MORE
		case 'l':
			if f.state != (0x10+0x1)+(0x1<<5) {
				return tokens.RET_TERM_FAIL
			}
			f.state = (0x10 + 0x1) + (0x2 << 5)
			return tokens.RET_FEED_MORE
		case 's':
			if f.state != (0x10+0x1)+(0x2<<5) {
				return tokens.RET_TERM_FAIL
			}
			f.state = (0x10 + 0x1) + (0x3 << 5)
			return tokens.RET_FEED_MORE
		case 'e':
			if f.state != (0x8+0x1)+(0x2<<5) && f.state != (0x10+0x1)+(0x3<<5) {
				return tokens.RET_TERM_FAIL
			}
			f.state = 1
			return tokens.RET_FEED_MORE
		default:
			return tokens.RET_TERM_FAIL
		}
	}
	if b >= '0' && b <= '9' {
		f.state |= 0x1
		f.state &= ^uint8(0x2)
		f.currD = f.currD*10 + I(b-'0')
		return tokens.RET_FEED_MORE
	} else if b == '-' {
		if f.state&0x5 != 0 {
			return tokens.RET_TERM_FAIL
		}
		// neg
		f.state |= 0x4
		f.state &= ^uint8(0x2)
		return tokens.RET_FEED_MORE
	} else if b == ',' {
		if f.state&0x2 != 0 || f.state&0x1 == 0 {
			// should have number before, but not
			return tokens.RET_TERM_FAIL
		}
		// handle value
		if f.state&0x4 == 0 {
			if err := f.Reciver(f.currD); err != nil {
				return tokens.RET_TERM_FAIL
			}
		} else {
			if err := f.Reciver(-f.currD); err != nil {
				return tokens.RET_TERM_FAIL
			}
		}
		f.currD = 0
		f.state = 0x2
		return tokens.RET_FEED_MORE
	} else if b == ' ' || b == '\n' || b == '\t' || b == '\r' || b == 'B' || b == 'b' || b == 'L' || b == 'l' || b == 'I' || b == 'i' {
		return tokens.RET_FEED_MORE
	}
	return tokens.RET_TERM_FAIL
}

func (o NumArray[I]) AcceptZero() bool {
	return false
}

func (o *NumArray[I]) Reset() {
	o.Reciver = nil
	o.state = 0
	o.currD = 0
}
