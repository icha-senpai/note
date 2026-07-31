// Copyright (c) 2019-present, Scribli


//go:build javascript
// +build javascript

package parse

import (
	"bytes"

	"github.com/icha-senpai/note/third_party/forks/lute/ast"
	"github.com/icha-senpai/note/third_party/forks/lute/editor"
)

func (t *Tree) FindLinkRefDefLink(label []byte) (link *ast.Node) {
	if !t.Context.ParseOption.LinkRef {
		return
	}

	if t.Context.ParseOption.VditorIR || t.Context.ParseOption.VditorSV || t.Context.ParseOption.VditorWYSIWYG || t.Context.ParseOption.ProtyleWYSIWYG {
		label = bytes.ReplaceAll(label, editor.CaretTokens, nil)
	}
	ast.Walk(t.Root, func(n *ast.Node, entering bool) ast.WalkStatus {
		if !entering || ast.NodeLinkRefDef != n.Type {
			return ast.WalkContinue
		}
		if bytes.EqualFold(n.Tokens, label) {
			link = n.FirstChild
			return ast.WalkStop
		}
		return ast.WalkContinue
	})
	return
}
