package main

import (
	"testing"
)

func Test_NewId(t *testing.T) {
	for i := int(1); i < 65; i++ {
		s := NewId(i)
		if i != len(s) || s == "" {
			t.Errorf("Id=%s i=%d len=%d\n", s, i, len(s))
		}
	}
}
