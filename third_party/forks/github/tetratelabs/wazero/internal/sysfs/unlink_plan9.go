package sysfs

import (
	"syscall"

	"github.com/icha-senpai/note/third_party/forks/github/tetratelabs/wazero/experimental/sys"
)

func unlink(name string) sys.Errno {
	err := syscall.Remove(name)
	return sys.UnwrapOSError(err)
}
