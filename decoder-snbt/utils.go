package decoder_snbt

import (
	"neomega-nbt/decoder-snbt/tokens"
	"neomega-nbt/decoder-snbt/tokens/left_barcket"
	whitespace "neomega-nbt/decoder-snbt/tokens/white_space"
)

func emptyCheck(snbt []uint8) bool {
	return len(consumeWhiteSpace(snbt)) == 0
}

func consumeWhiteSpace(snbt []uint8) []uint8 {
	snbt, _ = tokens.MatchToken(snbt, whitespace.FinityVariyLenWhiteSpace{})
	return snbt
}

type leftType uint8

const (
	ltUnknown       = leftType(0)
	ltByteArr       = leftType(5)
	ltIntArr        = leftType(6)
	ltLongArr       = leftType(7)
	ltCompound      = leftType(3)
	ltList          = leftType(4)
	ltDoubleQString = leftType(1)
	ltSingleQString = leftType(2)
)

func leftBracket(snbt []uint8) (
	rest []uint8,
	lt leftType,
) {
	lb := &left_barcket.LeftBrackerFinity{}
	var ok bool
	rest, ok = tokens.MatchToken(snbt, lb)
	if !ok {
		return snbt, ltUnknown
	}
	return rest, leftType(lb.ContainerType)
}
