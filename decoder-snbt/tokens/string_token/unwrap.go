package string_token

import "neomega-nbt/decoder-snbt/tokens"

type UnwrapString struct {
	Reciver func(byte)
}

func (o *UnwrapString) Feed(b byte) tokens.Status {
	if AcceptUnwrapStringFeed(b) {
		o.Reciver(b)
		return tokens.RET_OK
	} else {
		return tokens.RET_TERM_FAIL
	}
}

func AcceptUnwrapStringFeed(b byte) bool {
	switch b >> 6 {
	case 0:
		if (1<<((b<<2)>>2))&288063250384289792 > 0 {
			return true
		}
	case 1:
		if (1<<((b<<2)>>2))&576460745995190270 > 0 {
			return true
		}
	case 2:
		if (uint64(1)<<((b<<2)>>2))&18446744073709551615 > 0 {
			return true
		}
	case 3:
		if (uint64(1)<<((b<<2)>>2))&18446744073709551615 > 0 {
			return true
		}
	}
	return false
}
func (o *UnwrapString) AcceptZero() bool {
	return false
}
