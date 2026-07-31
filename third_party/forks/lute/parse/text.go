// Copyright (c) 2019-present, Scribli


package parse

import (
	"bytes"

	"github.com/icha-senpai/note/third_party/forks/lute/ast"
	"github.com/icha-senpai/note/third_party/forks/lute/editor"
	"github.com/icha-senpai/note/third_party/forks/lute/lex"
	"github.com/icha-senpai/note/third_party/forks/lute/util"
)

func (t *Tree) parseText(ctx *InlineContext) *ast.Node {
	start := ctx.pos
	for ; ctx.pos < ctx.tokensLen; ctx.pos++ {
		if t.isMarker(ctx.tokens[ctx.pos]) {
			break
		}
	}
	return &ast.Node{Type: ast.NodeText, Tokens: ctx.tokens[start:ctx.pos]}
}

func (t *Tree) isMarker(token byte) bool {
	if lex.IsMarker(token) {
		return true
	}

	if t.Context.ParseOption.Sup && lex.ItemCaret == token {
		return true
	}
	return false
}

var backslash = util.StrToBytes("\\")

func (t *Tree) parseBackslash(block *ast.Node, ctx *InlineContext) *ast.Node {
	if ctx.pos == ctx.tokensLen-1 {
		ctx.pos++
		return &ast.Node{Type: ast.NodeText, Tokens: backslash}
	}

	ctx.pos++
	token := ctx.tokens[ctx.pos]
	if lex.ItemNewline == token {
		ctx.pos++
		return &ast.Node{Type: ast.NodeHardBreak, Tokens: []byte{token}}
	}
	if lex.IsASCIIPunct(token) {
		if '<' == token && nil != t.Context.oldtip && ast.NodeTable == t.Context.oldtip.Type {
			isBr := ctx.tokens[ctx.pos:]
			if bytes.HasPrefix(isBr, []byte("<br />")) || bytes.HasPrefix(isBr, []byte("<br/>")) || bytes.HasPrefix(isBr, []byte("<br>")) {
				return &ast.Node{Type: ast.NodeText, Tokens: backslash}
			}
		}

		ctx.pos++
		n := &ast.Node{Type: ast.NodeBackslash}
		block.AppendChild(n)
		n.AppendChild(&ast.Node{Type: ast.NodeBackslashContent, Tokens: []byte{token}})
		return nil
	}
	if t.Context.ParseOption.VditorWYSIWYG || t.Context.ParseOption.VditorIR || t.Context.ParseOption.ProtyleWYSIWYG {
		tokens := ctx.tokens[ctx.pos:]
		caret := editor.CaretTokens
		if len(caret) < len(tokens) && bytes.HasPrefix(tokens, caret) {
			token = ctx.tokens[ctx.pos+len(caret)]
			if lex.IsASCIIPunct(token) {
				if '<' == token && nil != t.Context.oldtip && ast.NodeTable == t.Context.oldtip.Type {
					isBr := ctx.tokens[ctx.pos+len(caret):]
					if bytes.HasPrefix(isBr, []byte("<br />")) || bytes.HasPrefix(isBr, []byte("<br/>")) || bytes.HasPrefix(isBr, []byte("<br>")) {
						return &ast.Node{Type: ast.NodeText, Tokens: backslash}
					}
				}

				ctx.pos += len(caret)
				ctx.pos++
				n := &ast.Node{Type: ast.NodeBackslash}
				block.AppendChild(n)
				n.AppendChild(&ast.Node{Type: ast.NodeBackslashContent, Tokens: []byte{token}})
				if t.Context.ParseOption.ProtyleWYSIWYG {
					n.InsertBefore(&ast.Node{Type: ast.NodeText, Tokens: caret})
				} else {
					block.AppendChild(&ast.Node{Type: ast.NodeText, Tokens: caret})
				}
				return nil
			}
		}
	}
	return &ast.Node{Type: ast.NodeText, Tokens: backslash}
}

func (t *Tree) parseNewline(block *ast.Node, ctx *InlineContext) (ret *ast.Node) {
	pos := ctx.pos
	ctx.pos++

	isHardBreak := false
	if lastc := block.LastChild; nil != lastc && ast.NodeText == lastc.Type {
		tokens := lastc.Tokens
		if valueLen := len(tokens); lex.ItemSpace == tokens[valueLen-1] {
			_, lastc.Tokens = lex.TrimRight(tokens)
			if 1 < valueLen {
				isHardBreak = lex.ItemSpace == tokens[len(tokens)-2]
			}
		}
	}

	ret = &ast.Node{Type: ast.NodeSoftBreak, Tokens: []byte{ctx.tokens[pos]}}
	if t.Context.ParseOption.ProtyleWYSIWYG {
		return
	}

	if isHardBreak {
		ret.Type = ast.NodeHardBreak
	}
	return
}

func (t *Tree) MergeText() {
	t.mergeText(t.Root)
}

func (t *Tree) mergeText(node *ast.Node) {
	for child := node.FirstChild; nil != child; {
		next := child.Next
		if ast.NodeText == child.Type {
			for nil != next && ast.NodeText == next.Type {
				child.AppendTokens(next.Tokens)
				next.Unlink()
				next = child.Next
			}
		} else if ast.NodeLinkText == child.Type {
			for nil != next && ast.NodeLinkText == next.Type {
				child.AppendTokens(next.Tokens)
				next.Unlink()
				next = child.Next
			}
		} else {
			t.mergeText(child)
		}
		child = next
	}
}
