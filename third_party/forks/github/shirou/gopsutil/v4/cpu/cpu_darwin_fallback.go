// SPDX-License-Identifier: BSD-3-Clause
//go:build darwin && !arm64

package cpu

import "github.com/icha-senpai/note/third_party/forks/external/golang.org/x/sys/unix"

func getFrequency() (float64, error) {
	// Use the rated frequency of the CPU. This is a static value and does not
	// account for low power or Turbo Boost modes.
	cpuFrequency, err := unix.SysctlUint64("hw.cpufrequency")
	return float64(cpuFrequency) / 1000000.0, err
}
