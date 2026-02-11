package whitespace

import (
	"neomega-nbt/decoder-snbt/tokens"
)

type FinityWhiteSpace struct{}

func (o FinityWhiteSpace) Feed(b byte) tokens.Status {
	if b>>6 == 0 && ((1<<((b<<2)>>2))&4294983168 > 0) {
		return tokens.RET_TERM_OK
	}
	return tokens.RET_TERM_FAIL
}
func (o FinityWhiteSpace) AcceptZero() bool { return false }

func (o FinityWhiteSpace) String() string { return "[white space]" }

type FinityVariyLenWhiteSpace struct {
	AcceptZeroLength bool
}

func (o FinityVariyLenWhiteSpace) Feed(b byte) tokens.Status {
	if b>>6 == 0 && ((1<<((b<<2)>>2))&4294983168 > 0) {
		return tokens.RET_OK
	}
	return tokens.RET_TERM_FAIL
}
func (o FinityVariyLenWhiteSpace) AcceptZero() bool { return o.AcceptZeroLength }

// // ok when white space count==0
// func ConsumeWhiteSpace[ST char_sequence.CharSequence](source ST) {
// 	b, isEof := source.Current()
// 	if isEof {
// 		return
// 	}
// 	for {
// 		if (finityVariyLenWhiteSpace{}).Feed(b) == tokens.RET_TERM_FAIL {
// 			return
// 		}
// 		source.MoveToNext()
// 		b, isEof = source.Current()
// 	}
// }
