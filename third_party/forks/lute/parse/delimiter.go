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
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/icha-senpai/note/third_party/forks/lute/ast"
	"github.com/icha-senpai/note/third_party/forks/lute/editor"
	"github.com/icha-senpai/note/third_party/forks/lute/lex"
)

type delimiter struct {
	node           *ast.Node
	typ            byte
	num            int
	originalNum    int
	canOpen        bool
	canClose       bool
	previous, next *delimiter

	active            bool
	image             bool
	bracketAfter      bool
	index             int
	previousDelimiter *delimiter
}


func (t *Tree) handleDelim(block *ast.Node, ctx *InlineContext) {
	startPos := ctx.pos
	delim := t.scanDelims(ctx)

	text := ctx.tokens[startPos:ctx.pos]
	node := &ast.Node{Type: ast.NodeText, Tokens: text}
	block.AppendChild(node)

	if delim.canOpen || delim.canClose {
		ctx.delimiters = &delimiter{
			typ:         delim.typ,
			num:         delim.num,
			originalNum: delim.num,
			node:        node,
			previous:    ctx.delimiters,
			next:        nil,
			canOpen:     delim.canOpen,
			canClose:    delim.canClose,
		}
		if nil != ctx.delimiters.previous {
			ctx.delimiters.previous.next = ctx.delimiters
		}
	}
}

func (t *Tree) processEmphasis(stackBottom *delimiter, ctx *InlineContext) {
	if nil == ctx.delimiters {
		return
	}

	var opener, closer, oldCloser *delimiter
	var openerInl, closerInl *ast.Node
	var tempStack *delimiter
	var useDelims int
	var openerFound bool
	var openersBottom = map[byte]*delimiter{}
	var oddMatch = false

	openersBottom[lex.ItemUnderscore] = stackBottom
	openersBottom[lex.ItemAsterisk] = stackBottom
	openersBottom[lex.ItemTilde] = stackBottom
	openersBottom[lex.ItemEqual] = stackBottom
	openersBottom[lex.ItemCrosshatch] = stackBottom
	openersBottom[lex.ItemCaret] = stackBottom

	// find first closer above stack_bottom:
	closer = ctx.delimiters
	for nil != closer && closer.previous != stackBottom {
		closer = closer.previous
	}

	// move forward, looking for closers, and handling each
	for nil != closer {
		var closercc = closer.typ
		if !closer.canClose {
			closer = closer.next
			continue
		}

		// found emphasis closer. now look back for first matching opener:
		opener = closer.previous
		openerFound = false
		for nil != opener && opener != stackBottom && opener != openersBottom[closercc] {
			oddMatch = (closer.canOpen || opener.canClose) && closer.originalNum%3 != 0 && (opener.originalNum+closer.originalNum)%3 == 0
			if opener.typ == closer.typ && opener.canOpen && !oddMatch {
				openerFound = true
				break
			}
			opener = opener.previous
		}
		oldCloser = closer

		if !openerFound {
			closer = closer.next
		} else {
			if lex.ItemCrosshatch == closercc && t.Context.ParseOption.Tag {
				var tagContent strings.Builder
				for n := opener.node.Next; nil != n && n != closer.node; n = n.Next {
					tagContent.WriteString(n.Text())
				}
				content := strings.ReplaceAll(tagContent.String(), editor.Caret, "")
				content = strings.ReplaceAll(content, editor.Zwsp, "")
				if "" == strings.TrimSpace(content) {
					closer = closer.next
					continue
				}
			}

			// calculate actual number of delimiters used from closer
			if closer.num >= 2 && opener.num >= 2 {
				useDelims = 2
			} else {
				useDelims = 1
			}

			openerInl = opener.node
			closerInl = closer.node

			if t.Context.ParseOption.GFMStrikethrough || t.Context.ParseOption.Sub {
				if lex.ItemTilde == closercc && opener.num != closer.num {
					closer = closer.next
					continue
				}
			} else {
				if lex.ItemTilde == closercc {
					closer = closer.next
					continue
				}
			}

			if t.Context.ParseOption.Sup {
				if lex.ItemCaret == closercc && opener.num != closer.num {
					closer = closer.next
					continue
				}
			} else {
				if lex.ItemCaret == closercc {
					closer = closer.next
					continue
				}
			}

			if !t.Context.ParseOption.InlineAsterisk {
				if lex.ItemAsterisk == closercc {
					closer = closer.next
					continue
				}
			}

			if !t.Context.ParseOption.InlineUnderscore {
				if lex.ItemUnderscore == closercc {
					closer = closer.next
					continue
				}
			}

			if t.Context.ParseOption.Mark {
				if lex.ItemEqual == closercc && opener.num != closer.num {
					closer = closer.next
					continue
				}
			} else {
				if lex.ItemEqual == closercc {
					closer = closer.next
					continue
				}
			}

			if t.Context.ParseOption.Tag {
				if lex.ItemCrosshatch == closercc && opener.num != closer.num {
					closer = closer.next
					continue
				}
			} else {
				if lex.ItemCrosshatch == closercc {
					closer = closer.next
					continue
				}
			}

			// remove used delimiters from stack elts and inlines
			opener.num -= useDelims
			closer.num -= useDelims

			openerTokens := openerInl.Tokens[len(openerInl.Tokens)-useDelims:]
			text := openerInl.Tokens[0 : len(openerInl.Tokens)-useDelims]
			openerInl.Tokens = text
			closerTokens := closerInl.Tokens[len(closerInl.Tokens)-useDelims:]
			text = closerInl.Tokens[0 : len(closerInl.Tokens)-useDelims]
			closerInl.Tokens = text

			openMarker := &ast.Node{Tokens: openerTokens, Close: true}
			emStrongDelMark := &ast.Node{Close: true}
			closeMarker := &ast.Node{Tokens: closerTokens, Close: true}
			if 1 == useDelims {
				if lex.ItemAsterisk == closercc {
					emStrongDelMark.Type = ast.NodeEmphasis
					openMarker.Type = ast.NodeEmA6kOpenMarker
					closeMarker.Type = ast.NodeEmA6kCloseMarker
				} else if lex.ItemUnderscore == closercc {
					emStrongDelMark.Type = ast.NodeEmphasis
					openMarker.Type = ast.NodeEmU8eOpenMarker
					closeMarker.Type = ast.NodeEmU8eCloseMarker
				} else if lex.ItemTilde == closercc {
					if t.Context.ParseOption.Sub {
						emStrongDelMark.Type = ast.NodeSub
						openMarker.Type = ast.NodeSubOpenMarker
						closeMarker.Type = ast.NodeSubCloseMarker
					} else if t.Context.ParseOption.GFMStrikethrough && t.Context.ParseOption.GFMStrikethrough1 {
						emStrongDelMark.Type = ast.NodeStrikethrough
						openMarker.Type = ast.NodeStrikethrough1OpenMarker
						closeMarker.Type = ast.NodeStrikethrough1CloseMarker
					}
				} else if lex.ItemEqual == closercc {
					if t.Context.ParseOption.Mark {
						emStrongDelMark.Type = ast.NodeMark
						openMarker.Type = ast.NodeMark1OpenMarker
						closeMarker.Type = ast.NodeMark1CloseMarker
					}
				} else if lex.ItemCrosshatch == closercc {
					if t.Context.ParseOption.Tag {
						emStrongDelMark.Type = ast.NodeTag
						openMarker.Type = ast.NodeTagOpenMarker
						closeMarker.Type = ast.NodeTagCloseMarker
					}
				} else if lex.ItemCaret == closercc {
					if t.Context.ParseOption.Sup {
						emStrongDelMark.Type = ast.NodeSup
						openMarker.Type = ast.NodeSupOpenMarker
						closeMarker.Type = ast.NodeSupCloseMarker
					}
				}
			} else {
				if lex.ItemAsterisk == closercc {
					emStrongDelMark.Type = ast.NodeStrong
					openMarker.Type = ast.NodeStrongA6kOpenMarker
					closeMarker.Type = ast.NodeStrongA6kCloseMarker
				} else if lex.ItemUnderscore == closercc {
					emStrongDelMark.Type = ast.NodeStrong
					openMarker.Type = ast.NodeStrongU8eOpenMarker
					closeMarker.Type = ast.NodeStrongU8eCloseMarker
				} else if lex.ItemTilde == closercc {
					if t.Context.ParseOption.GFMStrikethrough {
						emStrongDelMark.Type = ast.NodeStrikethrough
						openMarker.Type = ast.NodeStrikethrough2OpenMarker
						closeMarker.Type = ast.NodeStrikethrough2CloseMarker
					}
				} else if lex.ItemEqual == closercc {
					if t.Context.ParseOption.Mark {
						emStrongDelMark.Type = ast.NodeMark
						openMarker.Type = ast.NodeMark2OpenMarker
						closeMarker.Type = ast.NodeMark2CloseMarker
					}
				}
			}

			tmp := openerInl.Next
			for nil != tmp && tmp != closerInl {
				next := tmp.Next
				tmp.Unlink()
				emStrongDelMark.AppendChild(tmp)
				tmp = next
			}

			emStrongDelMark.PrependChild(openMarker)
			emStrongDelMark.AppendChild(closeMarker)
			openerInl.InsertAfter(emStrongDelMark)

			// remove elts between opener and closer in delimiters stack
			if opener.next != closer {
				opener.next = closer
				closer.previous = opener
			}

			// if opener has 0 delims, remove it and the inline
			if opener.num == 0 {
				openerInl.Unlink()
				t.removeDelimiter(opener, ctx)
			}

			if closer.num == 0 {
				closerInl.Unlink()
				tempStack = closer.next
				t.removeDelimiter(closer, ctx)
				closer = tempStack
			}
		}

		if !openerFound && !oddMatch {
			// Set lower bound for future searches for openers:
			openersBottom[closercc] = oldCloser.previous
			if !oldCloser.canOpen {
				// We can remove a closer that can't be an opener,
				// once we've seen there's no matching opener:
				t.removeDelimiter(oldCloser, ctx)
			}
		}
	}

	for nil != ctx.delimiters && ctx.delimiters != stackBottom {
		t.removeDelimiter(ctx.delimiters, ctx)
	}
}

func (t *Tree) scanDelims(ctx *InlineContext) *delimiter {
	startPos := ctx.pos
	token := ctx.tokens[startPos]
	delimitersCount := 0
	for i := ctx.pos; i < ctx.tokensLen && token == ctx.tokens[i]; i++ {
		if lex.ItemCrosshatch == token && t.Context.ParseOption.Tag && delimitersCount >= 1 {
			break
		}
		delimitersCount++
		ctx.pos++
	}

	tokenBefore, tokenAfter := rune(lex.ItemNewline), rune(lex.ItemNewline)
	if 0 < startPos {
		c := ctx.tokens[startPos-1]
		if c >= utf8.RuneSelf {
			tokenBefore, _ = utf8.DecodeLastRune(ctx.tokens[:startPos])
		} else {
			tokenBefore = rune(c)
		}

		if (t.Context.ParseOption.VditorWYSIWYG || t.Context.ParseOption.VditorIR || t.Context.ParseOption.VditorSV || t.Context.ParseOption.ProtyleWYSIWYG) && editor.Caret == string(tokenBefore) {
			caretLen := len(editor.Caret)
			if 0 < startPos-caretLen {
				c = ctx.tokens[startPos-caretLen-1]
				if c >= utf8.RuneSelf {
					tokenBefore, _ = utf8.DecodeLastRune(ctx.tokens[:startPos-caretLen])
				} else {
					tokenBefore = rune(c)
				}
			}
		}
	}

	if ctx.tokensLen > ctx.pos {
		t := ctx.tokens[ctx.pos]
		if t >= utf8.RuneSelf {
			tokenAfter, _ = utf8.DecodeRune(ctx.tokens[ctx.pos:])
		} else {
			tokenAfter = rune(t)
		}
	}

	afterIsWhitespace := lex.IsUnicodeWhitespace(tokenAfter)
	afterIsPunct := unicode.IsPunct(tokenAfter) || unicode.IsSymbol(tokenAfter)
	if (lex.ItemAsterisk == token && '~' == tokenAfter) || (lex.ItemTilde == token && '*' == tokenAfter) ||
		(lex.ItemCaret == token && ('+' == tokenAfter || '-' == tokenAfter)) ||
		(lex.ItemTilde == token && ('+' == tokenAfter || '-' == tokenAfter)) {
		afterIsPunct = false
	}
	beforeIsWhitespace := lex.IsUnicodeWhitespace(tokenBefore)
	beforeIsPunct := unicode.IsPunct(tokenBefore) || unicode.IsSymbol(tokenBefore)
	if (lex.ItemAsterisk == token && '~' == tokenBefore) || (lex.ItemTilde == token && '*' == tokenBefore) ||
		(lex.ItemCaret == token && ('+' == tokenBefore || '-' == tokenBefore)) ||
		(lex.ItemTilde == token && ('+' == tokenBefore || '-' == tokenBefore)) {
		beforeIsPunct = false
	}

	if t.Context.ParseOption.ProtyleWYSIWYG {
		afterIsPunct, beforeIsPunct = false, false

		if lex.ItemUnderscore == token && editor.Caret == string(tokenBefore) {
			afterIsWhitespace = true
		}
		if lex.ItemUnderscore == token && editor.Caret == string(tokenAfter) {
			afterIsWhitespace = true
		}
	}

	isLeftFlanking := !afterIsWhitespace && (!afterIsPunct || beforeIsWhitespace || beforeIsPunct)
	isRightFlanking := !beforeIsWhitespace && (!beforeIsPunct || afterIsWhitespace || afterIsPunct)
	var canOpen, canClose bool
	if lex.ItemUnderscore == token {
		canOpen = isLeftFlanking && (!isRightFlanking || beforeIsPunct)
		canClose = isRightFlanking && (!isLeftFlanking || afterIsPunct)
	} else {
		if lex.ItemEqual == token {
			if !t.Context.ParseOption.Mark || 2 != delimitersCount  {
				canOpen, canClose = false, false
			} else {
				canOpen = isLeftFlanking
				canClose = isRightFlanking
			}
		} else if lex.ItemCrosshatch == token {
			if !t.Context.ParseOption.Tag || 1 != delimitersCount  {
				canOpen, canClose = false, false
			} else {
				canOpen = isLeftFlanking
				canClose = isRightFlanking
			}
		} else if lex.ItemCaret == token {
			if !t.Context.ParseOption.Sup || 1 != delimitersCount  {
				canOpen, canClose = false, false
			} else {
				canOpen = isLeftFlanking
				canClose = isRightFlanking
			}
		} else if lex.ItemTilde == token {
			if t.Context.ParseOption.Sub {
				if t.Context.ParseOption.GFMStrikethrough && 3 == delimitersCount {
					canOpen = isLeftFlanking
					canClose = isRightFlanking
				} else if 1 != delimitersCount {
					canOpen, canClose = false, false
					if t.Context.ParseOption.GFMStrikethrough && 2 == delimitersCount {
						canOpen = isLeftFlanking
						canClose = isRightFlanking
					}
				} else {
					canOpen = isLeftFlanking
					canClose = isRightFlanking
				}
			} else if t.Context.ParseOption.GFMStrikethrough {
				if 1 == delimitersCount {
					if !t.Context.ParseOption.GFMStrikethrough1 {
						canOpen, canClose = false, false
					} else {
						canOpen = isLeftFlanking
						canClose = isRightFlanking
					}
				} else {
					canOpen = isLeftFlanking
					canClose = isRightFlanking
				}
			} else {
				canOpen = isLeftFlanking
				canClose = isRightFlanking
			}
		} else {
			canOpen = isLeftFlanking
			canClose = isRightFlanking
		}
	}

	return &delimiter{typ: token, num: delimitersCount, active: true, canOpen: canOpen, canClose: canClose}
}

func (t *Tree) removeDelimiter(delim *delimiter, ctx *InlineContext) (ret *delimiter) {
	if nil != delim.previous {
		delim.previous.next = delim.next
	}
	if nil == delim.next {
		ctx.delimiters = delim.previous
	} else {
		delim.next.previous = delim.previous
	}
	return
}
