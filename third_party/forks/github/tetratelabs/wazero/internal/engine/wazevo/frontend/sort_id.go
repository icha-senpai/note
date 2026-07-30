package frontend

import (
	"slices"

	"github.com/icha-senpai/note/third_party/forks/github/tetratelabs/wazero/internal/engine/wazevo/ssa"
)

func sortSSAValueIDs(IDs []ssa.ValueID) {
	slices.SortFunc(IDs, func(i, j ssa.ValueID) int {
		return int(i) - int(j)
	})
}
