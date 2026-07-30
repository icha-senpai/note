//go:build amd64 && !purego

package bswap

import "github.com/icha-senpai/note/third_party/forks/github/segmentio/asm/cpu"

var cpuX86 = cpu.X86
