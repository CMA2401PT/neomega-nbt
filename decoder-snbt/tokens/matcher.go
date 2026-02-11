package tokens

func MatchToken[FT Token](source []uint8, token FT) ([]uint8, bool) {
	if len(source) == 0 {
		return source, token.AcceptZero()
	}
	p, count := 0, 0
	b := source[p]
	isEof := false
	back := 0
	var status Status
	for !isEof {
		p += 1
		count += 1
		back += 1
		status = token.Feed(b) & STATUS_RET
		switch status {
		case RET_TERM_OK:
			return source[p:], true
		case RET_TERM_FAIL:
			// 当前的 char 不属于此token，但是先前的char可能属于
			count -= back
			p -= back
			return source[p:], count != 0 || token.AcceptZero() // 如果count 数量非0，那么说明前面有一截是匹配成功的
		case RET_OK:
			back = 0 // 到目前为止匹配成功，因此无需向后退了
		}
		if p == len(source) {
			b, isEof = 0, true
		} else {
			b, isEof = source[p], false
		}
	}
	// 不会再有更多的 char 了，回退到最后一个成功位置
	count -= back
	p -= back
	return source[p:], count != 0 || token.AcceptZero() // 如果count 数量非0，那么说明前面有一截是匹配成功的
}
