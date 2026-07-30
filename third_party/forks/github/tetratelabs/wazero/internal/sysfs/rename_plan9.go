package sysfs

import (
	"os"

	"github.com/icha-senpai/note/third_party/forks/github/tetratelabs/wazero/experimental/sys"
)

func rename(from, to string) sys.Errno {
	if from == to {
		return 0
	}
	return sys.UnwrapOSError(os.Rename(from, to))
}
