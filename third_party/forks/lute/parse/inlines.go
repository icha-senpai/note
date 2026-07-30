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
	"github.com/icha-senpai/note/third_party/forks/lute/ast"
	"github.com/icha-senpai/note/third_party/forks/lute/html"
)

func (t *Tree) parseInlines() {
	t.walkParseInline(t.Root)

	if t.Context.ParseOption.KramdownSpanIAL {
		t.parseKramdownSpanIAL()
	}
}

func (t *Tree) walkParseInline(node *ast.Node) {
	if nil == node {
		return
	}

	typ := node.Type

	if ast.NodeSuperBlock == typ {
		if nil != node.LastChild && ast.NodeSuperBlockLayoutMarker == node.LastChild.Type {
			node.Type = ast.NodeParagraph
			node.Tokens = append([]byte("{{{"), node.LastChild.Tokens...)
			node.FirstChild.Unlink()
			node.LastChild.Unlink()
			typ = ast.NodeParagraph
		}
	}

	if ast.NodeParagraph == typ || ast.NodeHeading == typ || ast.NodeTableCell == typ {
		tokens := node.Tokens
		if ast.NodeParagraph == typ {
			if nil == tokens && nil == node.FirstChild {
				if ast.NodeListItem != node.Parent.Type || t.Context.ParseOption.VditorWYSIWYG || t.Context.ParseOption.VditorIR || t.Context.ParseOption.VditorSV {
					next := node.Next
					node.Unlink()
					node.Next = next
				}
				return
			} else if ial := t.Context.parseKramdownIALInListItem(tokens); 0 < len(ial) {
				if nil != node.Previous {
					for _, kv := range ial {
						node.Previous.SetIALAttr(kv[0], html.UnescapeAttrVal(kv[1]))
					}
					next := node.Next
					node.Unlink()
					node.Next = next
					if nil != node.Next && ast.NodeKramdownBlockIAL == node.Next.Type {
						node.Next.Tokens = IAL2Tokens(mergeIALPreservingOrder(Tokens2IAL(node.Next.Tokens), ial))
					}
					return
				}
			}
		}

		length := len(tokens)
		if 1 > length {
			return
		}

		ctx := &InlineContext{tokens: tokens, tokensLen: length}

		t.parseInline(node, ctx)

		t.processEmphasis(nil, ctx)

		t.mergeText(node)

		editorMode := t.Context.ParseOption.VditorWYSIWYG || t.Context.ParseOption.VditorIR || t.Context.ParseOption.VditorSV || t.Context.ParseOption.ProtyleWYSIWYG
		protyleAutoLink := t.Context.ParseOption.ProtyleWYSIWYG && t.Context.ParseOption.ProtyleWYSIWYGAutoLink
		if (t.Context.ParseOption.GFMAutoLink && !editorMode) || protyleAutoLink {
			t.parseGFMAutoEmailLink(node)
			t.parseGFMAutoLink(node)
		}

		if t.Context.ParseOption.Emoji {
			t.emoji(node)
		}
		return
	} else if ast.NodeCodeBlock == typ {
		if node.IsFencedCodeBlock {
			openMarker := &ast.Node{Type: ast.NodeCodeBlockFenceOpenMarker, Tokens: node.CodeBlockOpenFence, CodeBlockFenceLen: node.CodeBlockFenceLen}
			node.PrependChild(openMarker)
			info := &ast.Node{Type: ast.NodeCodeBlockFenceInfoMarker, CodeBlockInfo: node.CodeBlockInfo}
			node.AppendChild(info)
			code := &ast.Node{Type: ast.NodeCodeBlockCode, Tokens: node.Tokens}
			node.AppendChild(code)
			if nil == node.CodeBlockCloseFence {
				node.CodeBlockCloseFence = node.CodeBlockOpenFence
			}
			closeMarker := &ast.Node{Type: ast.NodeCodeBlockFenceCloseMarker, Tokens: node.CodeBlockCloseFence, CodeBlockFenceLen: node.CodeBlockFenceLen}
			node.AppendChild(closeMarker)
		} else {
			code := &ast.Node{Type: ast.NodeCodeBlockCode, Tokens: node.Tokens}
			node.AppendChild(code)
		}
		node.Tokens = nil
	}

	for child := node.FirstChild; nil != child; child = child.Next {
		t.walkParseInline(child)
	}
}
