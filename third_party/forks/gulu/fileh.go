// Gulu - Golang common utilities for everyone.
// Copyright (c) 2019-present, Scribli


//go:build !windows
// +build !windows

package gulu

import "path/filepath"

// IsHidden checks whether the file specified by the given path is hidden.
func (*GuluFile) IsHidden(path string) bool {
	path = filepath.Base(path)
	if 1 > len(path) {
		return false
	}
	return "." == path[:1]
}
