// Copyright (c) 2019-present, Scribli


package parse

import (
	"github.com/icha-senpai/note/third_party/forks/lute/ast"
	"github.com/icha-senpai/note/third_party/forks/lute/lex"
)

func (t *Tree) parseBang(ctx *InlineContext) (ret *ast.Node) {
	startPos := ctx.pos
	ctx.pos++
	if ctx.pos < ctx.tokensLen && lex.ItemOpenBracket == ctx.tokens[ctx.pos] {
		ctx.pos++
		ret = &ast.Node{Type: ast.NodeText, Tokens: ctx.tokens[startPos:ctx.pos]}
		t.addBracket(ret, startPos+2, true, ctx)
		return
	}

	ret = &ast.Node{Type: ast.NodeText, Tokens: ctx.tokens[startPos:ctx.pos]}
	return
}
