//go:build freebsd || openbsd || netbsd || dragonfly

package fsnotify

import "github.com/icha-senpai/note/third_party/forks/external/golang.org/x/sys/unix"

const openMode = unix.O_NONBLOCK | unix.O_RDONLY | unix.O_CLOEXEC
