package main

import (
	"testing"
)

func TestCombin(t *testing.T) {
	out := combin("123")
	if len(out) != 6 {
		t.Error(len(out))
	}

	t.Log(out)

	out = combin("1223")
	if len(out) != 12 {
		t.Error(len(out))
	}

	t.Log(out)

	out = combin("1234e")
	if len(out) != 120 {
		t.Error(len(out))
	}

	out = combin("4531be79")
	if len(out) != 40320 {
		t.Error(len(out))
	}
}
