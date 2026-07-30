//go:build tinygo

package sysfs

import (
	"os"

	"github.com/icha-senpai/note/third_party/forks/github/tetratelabs/wazero/experimental/sys"
)

func datasync(f *os.File) sys.Errno {
	return sys.ENOSYS
}
