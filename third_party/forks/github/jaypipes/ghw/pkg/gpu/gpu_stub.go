//go:build !linux && !windows
// +build !linux,!windows

// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.
//

package gpu

import (
	"runtime"

	"github.com/icha-senpai/note/third_party/forks/github/pkg/errors"
)

func (i *Info) load() error {
	return errors.New("gpuFillInfo not implemented on " + runtime.GOOS)
}
