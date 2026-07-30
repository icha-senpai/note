//go:build tinygo

package sysfs

import (
	"io/fs"

	experimentalsys "github.com/icha-senpai/note/third_party/forks/github/tetratelabs/wazero/experimental/sys"
	"github.com/icha-senpai/note/third_party/forks/github/tetratelabs/wazero/sys"
)

func inoFromFileInfo(_ string, info fs.FileInfo) (sys.Inode, experimentalsys.Errno) {
	return 0, experimentalsys.ENOTSUP
}
