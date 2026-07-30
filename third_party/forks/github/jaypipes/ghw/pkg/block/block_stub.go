//go:build !linux && !darwin && !windows
// +build !linux,!darwin,!windows

// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.
//

package block

import (
	"runtime"

	"github.com/icha-senpai/note/third_party/forks/github/pkg/errors"
)

func (i *Info) load() error {
	return errors.New("blockFillInfo not implemented on " + runtime.GOOS)
}
