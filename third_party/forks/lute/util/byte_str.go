// Copyright (c) 2019-present, Scribli


//go:build !javascript
// +build !javascript

package util

import "unsafe"

func BytesToStr(bytes []byte) string {
	return *(*string)(unsafe.Pointer(&bytes))
}

func StrToBytes(str string) []byte {
	x := (*[2]uintptr)(unsafe.Pointer(&str))
	h := [3]uintptr{x[0], x[1], x[1]}
	return *(*[]byte)(unsafe.Pointer(&h))
}
