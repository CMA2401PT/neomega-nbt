package num_arr

import (
	"neomega-nbt/decoder-snbt/tokens"
	"testing"
)

// func consumeWhiteSpaceAndComma[S lflb.Source](src S) {
// 	lflb.ReadFinity(src, whitespace.FinityVariyLenWhiteSpace{})
// 	lflb.ReadFinity(src, tokens.Specific(','))
// 	lflb.ReadFinity(src, whitespace.FinityVariyLenWhiteSpace{})
// }

func AssertArray[I ~int32 | ~int8 | ~int64 | ~uint32 | ~uint8 | ~uint64](
	src []uint8, val []I, t *testing.T,
) {
	data := []I{}
	_, ok := tokens.MatchToken(src, &NumArray[I]{
		Reciver: func(i I) error {
			data = append(data, i)
			return nil
		},
	})
	if !ok {
		t.FailNow()
	}
	if len(data) != len(val) {
		t.Logf("length mismatch: %v!=%v", len(data), len(val))
		t.FailNow()
	}
	for i, v := range data {
		if v != val[i] {
			t.Logf("value mismatch: %v!=%v", v, val[i])
			t.FailNow()
		}
	}
}

func AssertFail[I ~int32 | ~int8 | ~int64 | ~uint32 | ~uint8 | ~uint64](
	src []uint8, beginByte uint8, t *testing.T,
) {
	data := []I{}
	rest, ok := tokens.MatchToken(src, &NumArray[I]{
		Reciver: func(i I) error {
			data = append(data, i)
			return nil
		},
	})
	if ok {
		t.FailNow()
	}
	if rest[0] != beginByte {
		t.FailNow()
	}
}

func TestString(t *testing.T) {
	seq := "12, -34 ,-567, 89,10,11 , -12 ,-13 ,14 ,15   ] "
	src := []uint8(seq)
	AssertArray(src, []int32{12, -34, -567, 89, 10, 11, -12, -13, 14, 15}, t)

	seq = "12,3,] "
	src = []uint8(seq)
	AssertArray(src, []int32{12, 3}, t)

	seq = "true,false ]"
	src = []uint8(seq)
	AssertArray(src, []int32{1, 0}, t)

	seq = "12, -34 ,-567, 89,10,11 , -12 ,-13 ,14 ,15   "
	src = []uint8(seq)
	AssertFail[int32](src, '1', t)

	seq = "a,12, -34 ,-567, 89,10,11 , -12 ,-13 ,14 ,15]   "
	src = []uint8(seq)
	AssertFail[int32](src, 'a', t)

	seq = "12, --34] "
	src = []uint8(seq)
	AssertFail[int32](src, '1', t)

	seq = "12,,3] "
	src = []uint8(seq)
	AssertFail[int32](src, '1', t)

}
