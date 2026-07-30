// Copyright (c) 2019-present, b3log.org
//
// Lute is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//         http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package parse

import (
	"bytes"

	"github.com/icha-senpai/note/third_party/forks/lute/ast"
	"github.com/icha-senpai/note/third_party/forks/lute/editor"
	"github.com/icha-senpai/note/third_party/forks/lute/lex"
	"github.com/icha-senpai/note/third_party/forks/lute/util"
)

func MathBlockStart(t *Tree, container *ast.Node) int {
	if t.Context.indented {
		return 0
	}

	if ok, mathBlockDollarOffset := t.parseMathBlock(); ok {
		t.Context.closeUnmatchedBlocks()
		block := t.Context.addChild(ast.NodeMathBlock)
		block.MathBlockDollarOffset = mathBlockDollarOffset
		t.Context.advanceNextNonspace()
		t.Context.advanceOffset(mathBlockDollarOffset, false)
		return 2
	}
	return 0
}

func MathBlockContinue(mathBlock *ast.Node, context *Context) int {
	ln := context.currentLine
	indent := context.indent
	if 3 >= indent && context.isMathBlockClose(ln[context.nextNonspace:]) {
		context.finalize(mathBlock)
		return 2
	} else {
		i := mathBlock.MathBlockDollarOffset
		var token byte
		for i > 0 {
			token = lex.Peek(ln, context.offset)
			if lex.ItemSpace != token && lex.ItemTab != token {
				break
			}
			context.advanceOffset(1, true)
			i--
		}
	}
	return 0
}

var MathBlockMarker = util.StrToBytes("$$")
var MathBlockMarkerCaret = util.StrToBytes("$$" + editor.Caret)

func (context *Context) mathBlockFinalize(mathBlock *ast.Node) {
	if 2 > len(mathBlock.Tokens) {
		/*
			- foo

			    $$
			bar
			$$
		*/
		mathBlock.AppendChild(&ast.Node{Type: ast.NodeMathBlockOpenMarker})
		mathBlock.AppendChild(&ast.Node{Type: ast.NodeMathBlockContent})
		mathBlock.AppendChild(&ast.Node{Type: ast.NodeMathBlockCloseMarker})
		return
	}
	tokens := mathBlock.Tokens[2:]
	tokens = lex.TrimWhitespace(tokens)
	if context.ParseOption.VditorWYSIWYG || context.ParseOption.VditorIR || context.ParseOption.VditorSV || context.ParseOption.ProtyleWYSIWYG {
		if bytes.HasSuffix(tokens, MathBlockMarkerCaret) {
			tokens = bytes.TrimSuffix(tokens, MathBlockMarkerCaret)
			tokens = append(tokens, editor.CaretTokens...)
		}
	}
	if bytes.HasSuffix(tokens, MathBlockMarker) {
		tokens = tokens[:len(tokens)-2]
	}
	if bytes.Contains(tokens, []byte("<span data-type=")) {
		inlineTree := Inline("", tokens, context.ParseOption)
		if nil != inlineTree {
			tokens = []byte(inlineTree.Root.Content())
		}
	}

	mathBlock.Tokens = nil
	mathBlock.AppendChild(&ast.Node{Type: ast.NodeMathBlockOpenMarker})
	mathBlock.AppendChild(&ast.Node{Type: ast.NodeMathBlockContent, Tokens: tokens})
	mathBlock.AppendChild(&ast.Node{Type: ast.NodeMathBlockCloseMarker})
}

func (t *Tree) parseMathBlock() (ok bool, mathBlockDollarOffset int) {
	marker := t.Context.currentLine[t.Context.nextNonspace]
	if lex.ItemDollar != marker {
		return
	}

	fenceChar := marker
	fenceLength := 0
	for i := t.Context.nextNonspace; i < t.Context.currentLineLen && fenceChar == t.Context.currentLine[i]; i++ {
		fenceLength++
	}

	if 2 > fenceLength {
		return
	}
	return true, t.Context.indent
}

func (context *Context) isMathBlockClose(tokens []byte) bool {
	if context.ParseOption.KramdownBlockIAL && simpleCheckIsBlockIAL(tokens) {
		if ial := context.parseKramdownBlockIAL(tokens); 0 < len(ial) {
			context.Tip.ID = IAL2Map(ial)["id"]
			context.Tip.KramdownIAL = ial
			context.Tip.InsertAfter(&ast.Node{Type: ast.NodeKramdownBlockIAL, Tokens: tokens})
			return true
		}
	}

	closeMarker := tokens[0]
	if closeMarker != lex.ItemDollar {
		return false
	}
	if 2 > lex.Accept(tokens, closeMarker) {
		return false
	}
	tokens = lex.TrimWhitespace(tokens)
	for _, token := range tokens {
		if token != lex.ItemDollar {
			return false
		}
	}
	return true
}
