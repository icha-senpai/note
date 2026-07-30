package frontend

import (
	"github.com/icha-senpai/note/third_party/forks/github/tetratelabs/wazero/internal/engine/wazevo/ssa"
	"github.com/icha-senpai/note/third_party/forks/github/tetratelabs/wazero/internal/wasm"
)

func FunctionIndexToFuncRef(idx wasm.Index) ssa.FuncRef {
	return ssa.FuncRef(idx)
}
