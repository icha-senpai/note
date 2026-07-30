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
	"strconv"

	"github.com/icha-senpai/note/third_party/forks/lute/ast"
	"github.com/icha-senpai/note/third_party/forks/lute/editor"
	"github.com/icha-senpai/note/third_party/forks/lute/lex"
	"github.com/icha-senpai/note/third_party/forks/lute/util"
)

func ListStart(t *Tree, container *ast.Node) int {
	if ast.NodeList != container.Type && t.Context.indented {
		return 0
	}

	data, ial := t.parseListMarker(container)
	if nil == data {
		return 0
	}

	var emptyListItem *ast.Node
	if t.Context.ParseOption.EnsureListItemParagraph && ast.NodeListItem == t.Context.Tip.Type {
		fc := t.Context.Tip.FirstChild
		if nil == fc || ast.NodeKramdownBlockIAL == fc.Type {
			emptyListItem = t.Context.Tip
		}
	}

	t.Context.closeUnmatchedBlocks()

	//
	if nil != emptyListItem && nil == emptyListItem.FirstChild {
		p := &ast.Node{Type: ast.NodeParagraph}
		if t.Context.ParseOption.KramdownBlockIAL {
			id := ast.NewNodeID()
			ialTokens := []byte("{: id=\"" + id + "\"}")
			p.ID = id
			p.KramdownIAL = [][]string{{"id", id}}
			p.AppendChild(&ast.Node{Type: ast.NodeKramdownBlockIAL, Tokens: ialTokens})
		}
		emptyListItem.AppendChild(p)
	}

	listsMatch := container.Type == ast.NodeList && t.Context.listsMatch(container.ListData, data)
	if t.Context.Tip.Type != ast.NodeList || !listsMatch {
		list := t.Context.addChild(ast.NodeList)
		list.ListData = data
	}
	listItem := t.Context.addChild(ast.NodeListItem)
	listItem.ListData = data
	if t.Context.ParseOption.KramdownBlockIAL && nil != ial {
		listItem.KramdownIAL = ial
		listItem.ID = listItem.IALAttr("id")
		t.Context.offset += len(IAL2Tokens(ial))
	}

	listItem.Tokens = data.Marker
	if 1 == listItem.ListData.Typ || (3 == listItem.ListData.Typ && 0 == listItem.ListData.BulletChar) {
		prev := listItem.Previous
		if nil != prev {
			listItem.ListData.Num = prev.ListData.Num + 1
		} else {
			listItem.ListData.Num = data.Start
		}
	}
	return 1
}

func ListItemContinue(listItem *ast.Node, context *Context) int {
	if context.blank {
		if nil == listItem.FirstChild {
			return 1
		}

		context.advanceNextNonspace()
	} else if context.indent >= listItem.ListData.MarkerOffset+listItem.ListData.Padding {
		context.advanceOffset(listItem.ListData.MarkerOffset+listItem.ListData.Padding, true)
	} else {
		return 1
	}
	return 0
}

func (context *Context) listFinalize(list *ast.Node) {
	item := list.FirstChild

	for nil != item {
		if endsWithBlankLine(item) && nil != item.Next {
			list.ListData.Tight = false
			break
		}

		subItem := item.FirstChild
		for nil != subItem {
			if endsWithBlankLine(subItem) && (nil != item.Next || nil != subItem.Next) {
				list.ListData.Tight = false
				break
			}
			subItem = subItem.Next
		}
		item = item.Next
	}

	if context.ParseOption.KramdownBlockIAL {
		for li := list.FirstChild; nil != li; li = li.Next {
			if nil == li.FirstChild {
				if ast.NodeKramdownBlockIAL != li.Type {
					id := ast.NewNodeID()
					ialTokens := []byte("{: id=\"" + id + "\"}")
					li.KramdownIAL = [][]string{{"id", id}}
					li.ID = id
					li.InsertAfter(&ast.Node{Type: ast.NodeKramdownBlockIAL, Tokens: ialTokens})

					id = ast.NewNodeID()
					ialTokens = []byte("{: id=\"" + id + "\"}")
					p := &ast.Node{Type: ast.NodeParagraph, ID: id}
					p.KramdownIAL = [][]string{{"id", id}}
					p.ID = id
					p.InsertAfter(&ast.Node{Type: ast.NodeKramdownBlockIAL, Tokens: ialTokens})
					li.AppendChild(p)
					li = li.Next
				}

				continue
			}

			if 7 < len(li.FirstChild.Tokens) && '{' == li.FirstChild.Tokens[0] {
				if ial := context.parseKramdownIALInListItem(li.FirstChild.Tokens); 0 < len(ial) {
					li.KramdownIAL = ial
					li.ID = ial[0][1]
					ialTokens := IAL2Tokens(ial)
					li.InsertAfter(&ast.Node{Type: ast.NodeKramdownBlockIAL, Tokens: ialTokens})
					tokens := li.FirstChild.Tokens[bytes.Index(li.FirstChild.Tokens, []byte("}"))+1:]
					tokens = lex.TrimWhitespace(tokens)
					li.FirstChild.Tokens = tokens
					li = li.Next
				}
			} else {
				var ialTokens []byte
				if nil == li.KramdownIAL {
					id := ast.NewNodeID()
					ialTokens = []byte("{: id=\"" + id + "\"}")
					li.KramdownIAL = [][]string{{"id", id}}
					li.ID = id
				} else {
					ialTokens = IAL2Tokens(li.KramdownIAL)
				}
				li.InsertAfter(&ast.Node{Type: ast.NodeKramdownBlockIAL, Tokens: ialTokens})
			}
		}
	}
}

var items1 = util.StrToBytes("1")

func (t *Tree) parseListMarker(container *ast.Node) (data *ast.ListData, ial [][]string) {
	if 4 <= t.Context.indent {
		return nil, nil
	}

	ln := t.Context.currentLine
	if t.Context.ParseOption.ProtyleWYSIWYG {
		if beforeDotTokensIdx := bytes.Index(ln, []byte(". ")); -1 < beforeDotTokensIdx {
			beforeCaretDotTokens := ln[:beforeDotTokensIdx]
			if -1 < bytes.Index(beforeCaretDotTokens, editor.CaretTokens) {
				beforeCaretDotTokens = bytes.ReplaceAll(beforeCaretDotTokens, editor.CaretTokens, nil)
				if util.IsDigit(util.BytesToStr(beforeCaretDotTokens)) {
					ln = bytes.ReplaceAll(ln, editor.CaretTokens, nil)
					ln = bytes.Replace(ln, []byte(". "), []byte(". "+editor.Caret), 1)
					t.Context.currentLine = ln
				}
			}
		}
	}

	tokens := ln[t.Context.nextNonspace:]
	data = &ast.ListData{
		Typ:          0,
		Tight:        true,
		MarkerOffset: t.Context.indent,
		Num:          -1,
	}

	markerLength := 1
	marker := []byte{tokens[0]}
	var delim byte
	if lex.ItemPlus == marker[0] || lex.ItemHyphen == marker[0] || lex.ItemAsterisk == marker[0] {
		data.BulletChar = marker[0]
	} else if marker, delim = t.parseOrderedListMarker(tokens); nil != marker {
		if container.Type != ast.NodeParagraph || bytes.Equal(items1, marker) {
			data.Typ = 1
			data.Start, _ = strconv.Atoi(util.BytesToStr(marker))
			markerLength = len(marker) + 1
			data.Delimiter = delim
		} else {
			return nil, nil
		}
	} else {
		return nil, nil
	}

	data.Marker = marker
	if 1 == data.Typ {
		data.Marker = append(data.Marker, []byte(".")...)
	}

	token := ln[t.Context.nextNonspace+markerLength]

	if !lex.IsWhitespace(token) {
		return nil, nil
	}

	if container.Type == ast.NodeParagraph && lex.ItemNewline == token {
		return nil, nil
	}

	t.Context.advanceNextNonspace()
	t.Context.advanceOffset(markerLength, true)
	spacesStartCol := t.Context.column
	spacesStartOffset := t.Context.offset
	for {
		t.Context.advanceOffset(1, true)
		token = lex.Peek(ln, t.Context.offset)
		if t.Context.column-spacesStartCol >= 5 || 0 == (token) || (lex.ItemSpace != token && lex.ItemTab != token) {
			break
		}
	}

	token = lex.Peek(ln, t.Context.offset)
	isBlankItem := 0 == token || lex.ItemNewline == token
	spacesAfterMarker := t.Context.column - spacesStartCol
	if spacesAfterMarker >= 5 || spacesAfterMarker < 1 || isBlankItem {
		data.Padding = markerLength + 1
		t.Context.column = spacesStartCol
		t.Context.offset = spacesStartOffset
		if token = lex.Peek(ln, t.Context.offset); lex.ItemSpace == token || lex.ItemTab == token {
			t.Context.advanceOffset(1, true)
		}
	} else {
		data.Padding = markerLength + spacesAfterMarker
	}

	if !isBlankItem {

		tokens := ln[t.Context.offset:]
		if t.Context.ParseOption.KramdownBlockIAL {
			if ial = t.Context.parseKramdownIALInListItem(tokens); 0 < len(ial) {
				tokens = tokens[bytes.Index(tokens, []byte("}"))+1:]
			}
		}

		if t.Context.ParseOption.VditorWYSIWYG || t.Context.ParseOption.VditorIR || t.Context.ParseOption.VditorSV {
			tokens = bytes.ReplaceAll(tokens, editor.CaretTokens, nil)
		}

		if 3 <= len(tokens) {
			if lex.ItemOpenBracket == tokens[0] && lex.ItemCloseBracket != tokens[1] && lex.ItemCloseBracket == tokens[2] {
				marker := tokens[1]
				if t.Context.ParseOption.IsValidTaskListItemMarker(marker) {
					data.Typ = 3
					data.Checked = marker != ' '
				}
			}
		}
	}
	return
}

func (t *Tree) parseOrderedListMarker(tokens []byte) (marker []byte, delimiter byte) {
	length := len(tokens)
	var i int
	var token byte
	for ; i < length; i++ {
		token = tokens[i]
		if !lex.IsDigit(token) || 8 < i {
			delimiter = token
			break
		}
		marker = append(marker, token)
	}

	if 1 > len(marker) || (lex.ItemDot != delimiter && lex.ItemCloseParen != delimiter) {
		return nil, 0
	}

	return
}

func endsWithBlankLine(block *ast.Node) bool {
	for nil != block {
		if block.LastLineBlank {
			return true
		}
		t := block.Type
		if !block.LastLineChecked && (t == ast.NodeList || t == ast.NodeListItem) {
			block.LastLineChecked = true
			block = block.LastChild
		} else {
			block.LastLineChecked = true
			break
		}
	}

	return false
}
