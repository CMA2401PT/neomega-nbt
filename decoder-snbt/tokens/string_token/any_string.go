package string_token

import (
	"neomega-nbt/decoder-snbt/tokens"
)

type AnyString struct {
	state   byte //0 -> unknown 1-> stringD 2-> stringS 3-> unwrap 5-> stringD+slash 6-> stringS+slash
	Reciver func(byte)
}

func (o *AnyString) Feed(b byte) tokens.Status {
	switch o.state {
	case 0:
		// unknow
		switch b {
		case '"':
			o.state = 1
			return tokens.RET_FEED_MORE
		case '\'':
			o.state = 2
			return tokens.RET_FEED_MORE
		default:
			if AcceptUnwrapStringFeed(b) {
				o.state = 3
				o.Reciver(b)
				return tokens.RET_OK
			}
			return tokens.RET_TERM_FAIL
		}
	case 1:
		// stringD
		if b == '\\' {
			o.state = 5 // stringD + slash
			return tokens.RET_FEED_MORE
		}
		if b == '"' {
			return tokens.RET_TERM_OK
		}
		o.Reciver(b)
		return tokens.RET_FEED_MORE
	case 5:
		o.state = 1
		switch b {
		case 'b':
			o.Reciver('\b')
		case 'f':
			o.Reciver('\f')
		case 'n':
			o.Reciver('\n')
		case 'r':
			o.Reciver('\r')
		case 's':
			o.Reciver(' ')
		case 't':
			o.Reciver('\t')
		default:
			o.Reciver(b)
		}
		return tokens.RET_FEED_MORE
	case 2:
		// stringS
		if b == '\\' {
			o.state = 6 // stringS + slash
			return tokens.RET_FEED_MORE
		}
		if b == '\'' {
			return tokens.RET_TERM_OK
		}
		o.Reciver(b)
		return tokens.RET_FEED_MORE
	case 6:
		o.state = 2
		switch b {
		case 'b':
			o.Reciver('\b')
		case 'f':
			o.Reciver('\f')
		case 'n':
			o.Reciver('\n')
		case 'r':
			o.Reciver('\r')
		case 's':
			o.Reciver(' ')
		case 't':
			o.Reciver('\t')
		default:
			o.Reciver(b)
		}
		return tokens.RET_FEED_MORE
	case 3:
		if AcceptUnwrapStringFeed(b) {
			o.Reciver(b)
			return tokens.RET_OK
		} else {
			return tokens.RET_TERM_FAIL
		}
	default:
		return tokens.RET_TERM_FAIL
	}
}

func (o *AnyString) AcceptZero() bool {
	return false
}
