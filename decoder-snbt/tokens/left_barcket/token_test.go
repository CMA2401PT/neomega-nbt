package left_barcket

import (
	"neomega-nbt/decoder-snbt/tokens"
	whitespace "neomega-nbt/decoder-snbt/tokens/white_space"
	"testing"
)

func TestLeftContainer(t *testing.T) {
	seq := []uint8(`" ' { [ [I; [B; [L; [I `)
	assertType := func(tp uint8) {
		lb := &LeftBrackerFinity{}
		var ok bool
		seq, ok = tokens.MatchToken(seq, lb)
		if !ok || lb.ContainerType != tp {
			t.Logf("want: %v, get: %v \n", tp, lb.ContainerType)
			t.FailNow()
		}
		seq, ok = tokens.MatchToken(seq, whitespace.FinityVariyLenWhiteSpace{})
	}
	assertType(1)
	assertType(2)
	assertType(3)
	assertType(4)
	assertType(6)
	assertType(5)
	assertType(7)
	assertType(4)
}
