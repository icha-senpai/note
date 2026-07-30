//go:build darwin

package fsnotify

import "github.com/icha-senpai/note/third_party/forks/external/golang.org/x/sys/unix"

// note: this constant is not defined on BSD
const openMode = unix.O_EVTONLY | unix.O_CLOEXEC
