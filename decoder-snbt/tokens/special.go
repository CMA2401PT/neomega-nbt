package tokens

import "fmt"

type FinityAny struct{}

func (o FinityAny) AcceptZero() bool   { return false }
func (o FinityAny) Feed(b byte) Status { return RET_TERM_OK }

type Specific byte

func (o Specific) Feed(b byte) Status {
	if b == byte(o) {
		return RET_TERM_OK
	} else {
		return RET_TERM_FAIL
	}
}
func (o Specific) AcceptZero() bool { return false }
func (o Specific) String() string   { return fmt.Sprintf("[%v]", string([]byte{byte(o)})) }

type Merger[TightToken, LooseToken Token] struct {
	Tight   TightToken
	Loose   LooseToken
	isLoose bool
}

func (m *Merger[TA, TB]) AcceptZero() bool {
	if m.Tight.AcceptZero() {
		m.isLoose = false
		return true
	}
	if m.Loose.AcceptZero() {
		m.isLoose = true
		return true
	}
	return false
}

func (m *Merger[TA, TB]) Feed(b byte) Status {
	rp := RET_TERM_FAIL
	rs := m.Loose.Feed(b) & STATUS_RET
	if !m.isLoose {
		rp = m.Tight.Feed(b) & STATUS_RET
		if rp == RET_TERM_FAIL {
			// 代表 primary 已经失败，此时，是否成功看的是 loose 了
			m.isLoose = (rs != RET_TERM_FAIL)
		} else if rs == RET_TERM_FAIL {
			panic(fmt.Errorf("%T is not loose than %T", m.Loose, m.Tight))
		}
	}
	if m.isLoose {
		return rs
	}
	// tight 还没 fail
	if rp == rs {
		return rs
	}
	if rs == RET_FEED_MORE {
		if rp == RET_OK || rp == RET_TERM_OK {
			return RET_OK
		}
		return RET_FEED_MORE
	}
	return rs
}
