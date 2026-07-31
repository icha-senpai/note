// Gulu - Golang common utilities for everyone.
// Copyright (c) 2019-present, Scribli


package gulu

// IsNumOrLetter checks the specified rune is number or letter.
func (*GuluRune) IsNumOrLetter(r rune) bool {
	return ('0' <= r && '9' >= r) || Rune.IsLetter(r)
}

// IsLetter checks the specified rune is letter.
func (*GuluRune) IsLetter(r rune) bool {
	return 'a' <= r && 'z' >= r || 'A' <= r && 'Z' >= r
}
