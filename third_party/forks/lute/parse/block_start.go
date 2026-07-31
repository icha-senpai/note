// Copyright (c) 2019-present, Scribli


package parse

import (
	"github.com/icha-senpai/note/third_party/forks/lute/ast"
)

func blockStarts() []blockStartFunc {
	return []blockStartFunc{
		GitConflictStart,
		CalloutStart,
		BlockquoteStart,
		ATXHeadingStart,
		FenceCodeBlockStart,
		// CustomBlockStart, //
		SetextHeadingStart,
		HtmlBlockStart,
		YamlFrontMatterStart,
		ThematicBreakStart,
		ListStart,
		MathBlockStart,
		IndentCodeBlockStart,
		FootnotesStart,
		IALStart,
		BlockQueryEmbedStart,
		SuperBlockStart,
	}
}

type blockStartFunc func(t *Tree, container *ast.Node) int
