// Gulu - Golang common utilities for everyone.
// Copyright (c) 2019-present, Scribli


//go:build !windows

package gulu

import (
	"os/exec"
)

func CmdAttr(cmd *exec.Cmd) {
}

func DecodeCmdOutput(output []byte) string {
	return string(output)
}
