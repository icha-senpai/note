// SPDX-License-Identifier: BSD-3-Clause
//go:build aix && !cgo

package host

import (
	"context"

	"github.com/icha-senpai/note/third_party/forks/github/shirou/gopsutil/v4/internal/common"
)

func numProcs(_ context.Context) (uint64, error) {
	return 0, common.ErrNotImplementedError
}
