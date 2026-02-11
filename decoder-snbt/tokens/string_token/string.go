package string_token

import "neomega-nbt/decoder-snbt/tokens"

// abc"
type StringD struct {
	slash   bool
	Reciver func(byte)
}

func (o *StringD) Feed(b byte) tokens.Status {
	if o.slash {
		o.slash = false
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
	}
	if b == '\\' {
		o.slash = true
		return tokens.RET_FEED_MORE
	}
	if b == '"' {
		return tokens.RET_TERM_OK
	}
	o.Reciver(b)
	return tokens.RET_FEED_MORE
}

func (o *StringD) AcceptZero() bool {
	return false
}

// abc'
type StringS struct {
	slash   bool
	Reciver func(byte)
}

func (o *StringS) Feed(b byte) tokens.Status {
	if o.slash {
		o.slash = false
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
	}
	if b == '\\' {
		o.slash = true
		return tokens.RET_FEED_MORE
	}
	if b == '\'' {
		return tokens.RET_TERM_OK
	}
	o.Reciver(b)
	return tokens.RET_FEED_MORE
}

func (o *StringS) AcceptZero() bool {
	return false
}
