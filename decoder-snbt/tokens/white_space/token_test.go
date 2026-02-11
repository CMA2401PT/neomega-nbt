package whitespace

import (
	"neomega-nbt/decoder-snbt/tokens"
	"testing"
)

func TestWhiteSpace(t *testing.T) {
	seq := "a \t\n\v\f\rb \t\n\v\f\rcd"
	src := []uint8(seq)
	var ok bool

	if src, ok = tokens.MatchToken(src, FinityWhiteSpace{}); ok {
		t.FailNow()
	}
	src, ok = tokens.MatchToken(src, tokens.FinityAny{})
	if src, ok = tokens.MatchToken(src, FinityWhiteSpace{}); !ok {
		t.FailNow()
	}
	if src, ok = tokens.MatchToken(src, FinityWhiteSpace{}); !ok {
		t.FailNow()
	}
	if src, ok = tokens.MatchToken(src, FinityWhiteSpace{}); !ok {
		t.FailNow()
	}
	if src, ok = tokens.MatchToken(src, FinityWhiteSpace{}); !ok {
		t.FailNow()
	}
	if src, ok = tokens.MatchToken(src, FinityWhiteSpace{}); !ok {
		t.FailNow()
	}
	if src, ok = tokens.MatchToken(src, FinityWhiteSpace{}); !ok {
		t.FailNow()
	}
	if src, ok = tokens.MatchToken(src, FinityWhiteSpace{}); ok {
		t.FailNow()
	}
	src, ok = tokens.MatchToken(src, tokens.FinityAny{})
	if src, ok = tokens.MatchToken(src, FinityVariyLenWhiteSpace{}); !ok {
		t.FailNow()
	}
	if src, ok = tokens.MatchToken(src, FinityWhiteSpace{}); ok {
		t.FailNow()
	}
	if src, ok = tokens.MatchToken(src, tokens.FinityAny{}); !ok {
		t.FailNow()
	}
	if src, ok = tokens.MatchToken(src, FinityVariyLenWhiteSpace{AcceptZeroLength: true}); !ok {
		t.FailNow()
	}
	if src, ok = tokens.MatchToken(src, FinityWhiteSpace{}); ok {
		t.FailNow()
	}
	if src, ok = tokens.MatchToken(src, tokens.FinityAny{}); !ok {
		t.FailNow()
	}
}
