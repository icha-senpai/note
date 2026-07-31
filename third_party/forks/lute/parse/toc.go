// Copyright (c) 2019-present, Scribli


package parse

import (
	"bytes"

	"github.com/icha-senpai/note/third_party/forks/lute/ast"
	"github.com/icha-senpai/note/third_party/forks/lute/editor"
	"github.com/icha-senpai/note/third_party/forks/lute/lex"
)

func (context *Context) parseToC(paragraph *ast.Node) *ast.Node {
	lines := lex.Split(paragraph.Tokens, lex.ItemNewline)
	if 1 != len(lines) {
		return nil
	}

	content := bytes.TrimSpace(lines[0])
	if context.ParseOption.VditorWYSIWYG || context.ParseOption.VditorIR || context.ParseOption.VditorSV {
		content = bytes.ReplaceAll(content, editor.CaretTokens, nil)
	}
	if !bytes.EqualFold(content, []byte("[toc]")) {
		return nil
	}
	return &ast.Node{Type: ast.NodeToC}
}
