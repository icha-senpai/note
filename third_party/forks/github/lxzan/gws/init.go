package gws

import "github.com/icha-senpai/note/third_party/forks/github/lxzan/gws/internal"

var (
	framePadding    = frameHeader{}
	defaultLogger   = new(stdLogger)
	bufferThreshold = uint32(256 * 1024)
	binaryPool      = new(internal.BufferPool)
)

func init() {
	SetBufferThreshold(bufferThreshold)
}

// Set the buffer threshold, x=pow(2,n), that buffers larger than x bytes are not reclaimed.
func SetBufferThreshold(x uint32) {
	bufferThreshold = internal.ToBinaryNumber(x)
	binaryPool = internal.NewBufferPool(128, bufferThreshold)
}
