package string_token

import (
	"neomega-nbt/decoder-snbt/tokens"
	whitespace "neomega-nbt/decoder-snbt/tokens/white_space"
	"testing"
)

func consumeWhiteSpaceAndComma(src []uint8) []uint8 {
	src, _ = tokens.MatchToken(src, whitespace.FinityVariyLenWhiteSpace{})
	src, _ = tokens.MatchToken(src, tokens.Specific(','))
	src, _ = tokens.MatchToken(src, whitespace.FinityVariyLenWhiteSpace{})
	return src
}

func TestString(t *testing.T) {
	seq := []uint8{}

	assertFailUnwrap := func() {
		var ok bool
		seq, ok = tokens.MatchToken(seq, &UnwrapString{
			Reciver: func(b byte) {

			},
		})
		if ok {
			t.FailNow()
		}
	}
	assertValUnwrap := func(v string) {
		var out []uint8
		var ok bool
		seq, ok = tokens.MatchToken(seq, &UnwrapString{
			Reciver: func(b byte) {
				out = append(out, b)
			},
		})
		if !ok {
			t.FailNow()
		}
		if string(out) != v {
			t.Logf("want: %v get: %v", v, string(out))
			t.FailNow()
		}
	}

	assertFailD := func() {
		var ok bool
		seq, ok = tokens.MatchToken(seq, &StringD{
			Reciver: func(b byte) {

			},
		})
		if ok {
			t.FailNow()
		}
	}
	assertValD := func(v string) {
		var out []uint8
		var ok bool
		seq, ok = tokens.MatchToken(seq, &StringD{
			Reciver: func(b byte) {
				out = append(out, b)
			},
		})
		if !ok {
			t.FailNow()
		}
		if string(out) != v {
			t.Logf("want: %v get: %v", v, string(out))
			t.FailNow()
		}
	}
	assertFailS := func() {
		var ok bool
		seq, ok = tokens.MatchToken(seq, &StringS{
			Reciver: func(b byte) {

			},
		})
		if ok {
			t.FailNow()
		}
	}
	assertValS := func(v string) {
		var out []uint8
		var ok bool
		seq, ok = tokens.MatchToken(seq, &StringS{
			Reciver: func(b byte) {
				out = append(out, b)
			},
		})
		if !ok {
			t.FailNow()
		}
		if string(out) != v {
			t.Logf("want: %v get: %v", v, string(out))
			t.FailNow()
		}
	}
	assertValAny := func(v string) {
		var out []uint8
		var ok bool
		seq, ok = tokens.MatchToken(seq, &AnyString{
			Reciver: func(b byte) {
				out = append(out, b)
			},
		})
		if !ok {
			t.FailNow()
		}
		if string(out) != v {
			t.Logf("want: %v get: %v", v, string(out))
			t.FailNow()
		}
	}
	assertFailAny := func() {
		var ok bool
		var out []uint8
		seq, ok = tokens.MatchToken(seq, &AnyString{
			Reciver: func(b byte) {
				out = append(out, b)
			},
		})
		if ok {
			t.Fatalf("expect fail, get: %v", string(out))
		}
	}
	seq = []uint8("]")
	assertFailUnwrap()
	seq = []uint8(" ")
	assertFailUnwrap()
	seq = []uint8(",")
	assertFailUnwrap()
	seq = []uint8("}")
	assertFailUnwrap()
	seq = []uint8("  abc_123.+-ABC, a_c+123.测试")
	assertFailUnwrap()
	seq = consumeWhiteSpaceAndComma(seq)
	assertValUnwrap("abc_123.+-ABC")
	seq = consumeWhiteSpaceAndComma(seq)
	assertValUnwrap("a_c+123.测试")

	seq = []uint8(`abc()  _+:'\"\\测试1" "abc()  _+:'\"\\测试2`)
	assertValD(`abc()  _+:'"\测试1`)
	seq = consumeWhiteSpaceAndComma(seq)
	seq = seq[1:]
	assertFailD()

	seq = []uint8(`abc()  _+:"\'\\测试' 'abc()  _+:"\'\\测试`)
	assertValS(`abc()  _+:"'\测试`)
	seq = consumeWhiteSpaceAndComma(seq)
	seq = seq[1:]
	assertFailS()
	if seq[0] != 'a' {
		t.FailNow()
	}

	seq = []uint8(`'abc()  _+:"\'\\测试' "abc()  _+:'\"\\测试" abc_123.+-ABC, `)
	assertValAny(`abc()  _+:"'\测试`)
	seq = consumeWhiteSpaceAndComma(seq)
	assertValAny(`abc()  _+:'"\测试`)
	seq = consumeWhiteSpaceAndComma(seq)
	assertValAny(`abc_123.+-ABC`)

	seq = []uint8(`'abc()  _+:"\'\\测试 `)
	assertFailAny()
	if seq[0] != '\'' {
		t.FailNow()
	}

	seq = []uint8(`"abc()  _+:'\"\\测试 `)
	assertFailAny()
	if seq[0] != '"' {
		t.FailNow()
	}
	seq = []uint8(`abc_123.+-ABC`)
	assertValAny(`abc_123.+-ABC`)
}
