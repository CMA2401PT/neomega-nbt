package left_barcket

import (
	"neomega-nbt/decoder-snbt/tokens"
)

type LeftBrackerFinity struct {
	state uint8
	// left container token specific
	ContainerType uint8 // 1-> "" 2-> '' 3-> {} 4-> [] 5-> [B;] 6-> [I;] 7-> [L;]
}

func (o *LeftBrackerFinity) Feed(b byte) tokens.Status {
	nextMarkTermAccept := LeftBrackerFeed(o.state, b)
	var ContainerType uint8
	o.state = uint8(nextMarkTermAccept >> 8)
	if nextMarkTermAccept&0x80 == 0 {
		m := uint8(nextMarkTermAccept) >> 2
		ContainerType = m + 1
	}
	if termOrAcc := nextMarkTermAccept & 0x3; termOrAcc != 0 {
		switch termOrAcc {
		case 0x3:
			o.ContainerType = ContainerType
			return tokens.RET_TERM_OK
		case 0x2:
			return tokens.RET_TERM_FAIL
		case 0x1:
			o.ContainerType = ContainerType
			return tokens.RET_OK
		}
	}
	return tokens.RET_FEED_MORE
}
func (o LeftBrackerFinity) AcceptZero() bool {
	return false
}
