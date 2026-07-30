// SPDX-License-Identifier: BSD-3-Clause
//go:build freebsd

package sensors

import (
	"context"

	"github.com/icha-senpai/note/third_party/forks/github/shirou/gopsutil/v4/internal/common"
)

func TemperaturesWithContext(_ context.Context) ([]TemperatureStat, error) {
	return []TemperatureStat{}, common.ErrNotImplementedError
}
