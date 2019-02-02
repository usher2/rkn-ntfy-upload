package main

import (
	"crypto/rand"
	"encoding/base32"
	"strings"
)

func NewId(x int) string {
	if x < 1 {
		x = 1
	}
	y := (5*x + 4) / 8
	u := make([]byte, y)
	if _, err := rand.Read(u[:]); err != nil {
		panic(err)
	}
	id := make([]byte, base32.StdEncoding.WithPadding(base32.NoPadding).EncodedLen(y))
	base32.StdEncoding.WithPadding(base32.NoPadding).Encode(id, u)
	return strings.ToLower(string(id[0:x]))
}
