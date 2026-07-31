// Copyright (c) 2019-present, Scribli
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

func (t *Tree) parseBlocks() {
	t.Context.Tip = t.Root
	lines := 0
	for line := t.lexer.NextLine(); nil != line; line = t.lexer.NextLine() {
		if t.Context.ParseOption.VditorWYSIWYG || t.Context.ParseOption.VditorIR || t.Context.ParseOption.VditorSV || t.Context.ParseOption.ProtyleWYSIWYG {
			if !bytes.Equal(line, editor.CaretNewlineTokens) && t.Context.Tip.ParentIs(ast.NodeListItem) && bytes.HasPrefix(line, editor.CaretTokens) {
				if ast.NodeListItem == t.Context.Tip.Type {
					t.Context.Tip.AppendChild(&ast.Node{Type: ast.NodeText, Tokens: line})
					break
				} else {
					t.Context.Tip.Tokens = bytes.TrimSuffix(t.Context.Tip.Tokens, []byte("\n"))
					t.Context.Tip.Tokens = append(t.Context.Tip.Tokens, editor.CaretNewlineTokens...)
				}
				line = line[len(editor.CaretTokens):]
			}
		}

		t.incorporateLine(line)
		lines++
	}
	for nil != t.Context.Tip {
		t.Context.finalize(t.Context.Tip)
	}
}

func (t *Tree) BlockCount() (ret int) {
	ast.Walk(t.Root, func(n *ast.Node, entering bool) ast.WalkStatus {
		if !entering {
			return ast.WalkContinue
		}

		if "" == n.ID || !n.IsBlock() {
			return ast.WalkContinue
		}

		ret++
		return ast.WalkContinue
	})
	return
}

func (t *Tree) DocBlockCount() (ret int) {
	ast.Walk(t.Root, func(n *ast.Node, entering bool) ast.WalkStatus {
		if !entering {
			return ast.WalkContinue
		}

		if !n.IsChildBlockOf(t.Root, 1) {
			return ast.WalkContinue
		}

		ret++
		return ast.WalkContinue
	})
	return
}

func (t *Tree) incorporateLine(line []byte) {
	t.Context.oldtip = t.Context.Tip
	t.Context.offset = 0
	t.Context.column = 0
	t.Context.blank = false
	t.Context.partiallyConsumedTab = false
	t.Context.currentLine = line
	t.Context.currentLineLen = len(t.Context.currentLine)

	allMatched := true
	var container *ast.Node
	container = t.Root
	lastChild := container.LastChild
	for ; nil != lastChild && !lastChild.Close; lastChild = container.LastChild {
		container = lastChild
		t.Context.findNextNonspace()

		switch _continue(container, t.Context) {
		case 0:
			break
		case 1:
			allMatched = false
			break
		case 2:
			return
		case 3:
			t.Context.closeSuperBlockChildren()
			if ast.NodeSuperBlock != t.Context.Tip.Type {
				sb := t.Context.Tip.Parent
				sb.Close = true
				sb.AppendChild(&ast.Node{Type: ast.NodeSuperBlockCloseMarker})
				t.Context.Tip = sb.Parent
				t.Context.lastMatchedContainer = sb
			} else {
				t.Context.Tip.AppendChild(&ast.Node{Type: ast.NodeSuperBlockCloseMarker})
				t.Context.Tip.Close = true
				t.Context.Tip = t.Context.Tip.Parent
				t.Context.lastMatchedContainer = t.Context.Tip
			}
			return
		}

		if !allMatched {
			container = container.Parent
			break
		}
	}

	t.Context.allClosed = container == t.Context.oldtip
	t.Context.lastMatchedContainer = container

	matchedLeaf := container.Type != ast.NodeParagraph && container.AcceptLines()
	blockParsers := blockStarts()
	startsLen := len(blockParsers)

	for !matchedLeaf {
		t.Context.findNextNonspace()

		maybeMarker := t.Context.currentLine[t.Context.nextNonspace]
		if !t.Context.indented &&
			lex.ItemHyphen != maybeMarker && lex.ItemAsterisk != maybeMarker && lex.ItemPlus != maybeMarker &&
			!lex.IsDigit(maybeMarker) &&
			lex.ItemBacktick != maybeMarker && lex.ItemTilde != maybeMarker &&
			lex.ItemSemicolon != maybeMarker &&
			lex.ItemCrosshatch != maybeMarker &&
			lex.ItemGreater != maybeMarker &&
			lex.ItemLess != maybeMarker &&
			lex.ItemUnderscore != maybeMarker && lex.ItemEqual != maybeMarker &&
			lex.ItemDollar != maybeMarker &&
			lex.ItemOpenBracket != maybeMarker &&
			lex.ItemOpenBrace != maybeMarker &&
			lex.ItemCloseBrace != maybeMarker &&
			lex.ItemBang != maybeMarker && "！"[0] != maybeMarker &&
			editor.Caret[0] != maybeMarker {
			t.Context.advanceNextNonspace()
			break
		}

		i := 0
		for i < startsLen {
			res := blockParsers[i](t, container)
			if res == 1 {
				container = t.Context.Tip
				break
			} else if res == 2 {
				container = t.Context.Tip
				matchedLeaf = true
				break
			} else {
				i++
			}
		}

		if i == startsLen {
			t.Context.advanceNextNonspace()
			break
		}
	}

	if !t.Context.allClosed && !t.Context.blank && t.Context.Tip.Type == ast.NodeParagraph {
		t.addLine()
	} else {
		t.Context.closeUnmatchedBlocks()

		if t.Context.blank && nil != container.LastChild {
			container.LastChild.LastLineBlank = true
		}

		typ := container.Type
		isFenced := ast.NodeCodeBlock == typ && container.IsFencedCodeBlock

		lastLineBlank := t.Context.blank &&
			!(typ == ast.NodeFootnotesDef ||
				typ == ast.NodeBlockquote || typ == ast.NodeCallout ||
				(typ == ast.NodeCodeBlock && isFenced) ||
				(typ == ast.NodeCustomBlock) ||
				(typ == ast.NodeMathBlock) ||
				(typ == ast.NodeGitConflict) ||
				(typ == ast.NodeListItem && nil == container.FirstChild))
		for cont := container; nil != cont; cont = cont.Parent {
			cont.LastLineBlank = lastLineBlank
		}

		if container.AcceptLines() {
			t.addLine()
			switch typ {
			case ast.NodeHTMLBlock:
				html := container
				if html.HtmlBlockType >= 1 && html.HtmlBlockType <= 5 {
					tokens := t.Context.currentLine[t.Context.offset:]
					if t.isHTMLBlockClose(tokens, html.HtmlBlockType) {
						t.Context.finalize(container)
					}
				}
			case ast.NodeMathBlock:
				if 3 > len(container.Tokens) {
					break
				}
				firstMarkerIdx := bytes.Index(container.Tokens, MathBlockMarker)
				if firstMarkerIdx == -1 {
					break
				}
				lastMarkerIdx := bytes.LastIndex(container.Tokens, MathBlockMarker)
				if lastMarkerIdx == -1 {
					break
				}
				if lastMarkerIdx == firstMarkerIdx {
					break
				}

				tails := container.Tokens[lastMarkerIdx:]
				if bytes.HasPrefix(tails, MathBlockMarker) && bytes.HasSuffix(tails, []byte("\n")) {
					if bytes.Equal(MathBlockMarker, bytes.TrimSpace(tails)) {
						t.Context.finalize(container)
					}
				}
			}
		} else if t.Context.offset < t.Context.currentLineLen && !t.Context.blank {
			t.Context.addChild(ast.NodeParagraph)
			t.Context.advanceNextNonspace()
			t.addLine()
		}
	}
}

func (t *Tree) addLine() {
	if t.Context.partiallyConsumedTab {
		t.Context.offset++ // skip over tab
		// add space characters:
		charsToTab := 4 - (t.Context.column % 4)
		t.Context.Tip.AppendTokens(bytes.Repeat(util.StrToBytes(" "), charsToTab))
	}

	startWithSpace := 1 < t.Context.currentLineLen && (' ' == t.Context.currentLine[0] || '\t' == t.Context.currentLine[0])
	docChildPara := ast.NodeDocument == t.Context.Tip.Parent.Type
	if t.Context.ParseOption.ParagraphBeginningSpace && startWithSpace && docChildPara {
		t.Context.Tip.AppendTokens(t.Context.currentLine)
	} else {
		t.Context.Tip.AppendTokens(t.Context.currentLine[t.Context.offset:])
	}
}

func _continue(n *ast.Node, context *Context) int {
	switch n.Type {
	case ast.NodeCodeBlock:
		return CodeBlockContinue(n, context)
	case ast.NodeHTMLBlock:
		return HtmlBlockContinue(n, context)
	case ast.NodeParagraph:
		return ParagraphContinue(n, context)
	case ast.NodeListItem:
		return ListItemContinue(n, context)
	case ast.NodeBlockquote:
		return BlockquoteContinue(n, context)
	case ast.NodeMathBlock:
		return MathBlockContinue(n, context)
	case ast.NodeYamlFrontMatter:
		return YamlFrontMatterContinue(n, context)
	case ast.NodeFootnotesDef:
		return FootnotesContinue(n, context)
	case ast.NodeSuperBlock:
		return SuperBlockContinue(n, context)
	case ast.NodeGitConflict:
		return GitConflictContinue(n, context)
	case ast.NodeCustomBlock:
		return CustomBlockContinue(n, context)
	case ast.NodeCallout:
		return CalloutContinue(n, context)
	case ast.NodeHeading, ast.NodeThematicBreak, ast.NodeKramdownBlockIAL, ast.NodeLinkRefDefBlock, ast.NodeBlockQueryEmbed,
		ast.NodeIFrame, ast.NodeVideo, ast.NodeAudio, ast.NodeWidget, ast.NodeAttributeView:
		return 1
	}
	return 0
}
