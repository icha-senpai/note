//go:build linux && !tinygo

package sysfs

import (
	"os"
	"syscall"

	"github.com/icha-senpai/note/third_party/forks/github/tetratelabs/wazero/experimental/sys"
)

func datasync(f *os.File) sys.Errno {
	return sys.UnwrapOSError(syscall.Fdatasync(int(f.Fd())))
}
