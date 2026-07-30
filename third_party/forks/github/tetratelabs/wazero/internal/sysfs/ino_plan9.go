package sysfs

import (
	"io/fs"

	experimentalsys "github.com/icha-senpai/note/third_party/forks/github/tetratelabs/wazero/experimental/sys"
	"github.com/icha-senpai/note/third_party/forks/github/tetratelabs/wazero/sys"
)

func inoFromFileInfo(_ string, info fs.FileInfo) (sys.Inode, experimentalsys.Errno) {
	if v, ok := info.Sys().(*sys.Stat_t); ok {
		return v.Ino, 0
	}
	return 0, 0
}
