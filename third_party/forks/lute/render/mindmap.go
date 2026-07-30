// Copyright (c) 2019-present, b3log.org
//
// Lute is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//         http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package render

import (
	"bytes"
	"strings"

	"github.com/icha-senpai/note/third_party/forks/lute/editor"
	"github.com/icha-senpai/note/third_party/forks/lute/html"

	"github.com/icha-senpai/note/third_party/forks/lute/ast"
	"github.com/icha-senpai/note/third_party/forks/lute/parse"
	"github.com/icha-senpai/note/third_party/forks/lute/util"
)

func EChartsMindmapStr(listContent string) string {
	return util.BytesToStr(echartsMindmap(util.StrToBytes(listContent)))
}

func EChartsMindmap(listContent []byte) []byte {
	return html.EncodeDestination(echartsMindmap(listContent))
}

func echartsMindmap(listContent []byte) []byte {
	listContent = bytes.ReplaceAll(listContent, editor.CaretTokens, nil)
	tree := parse.Parse("", listContent, parse.NewOptions())
	if nil == tree.Root.FirstChild || ast.NodeList != tree.Root.FirstChild.Type {
		return []byte("{}")
	}

	var toRemoved []*ast.Node
	for c := tree.Root.FirstChild; nil != c; c = c.Next {
		if ast.NodeList != c.Type {
			toRemoved = append(toRemoved, c)
		}
	}
	for _, c := range toRemoved {
		c.Unlink()
	}

	buf := &bytes.Buffer{}
	ast.Walk(tree.Root, func(n *ast.Node, entering bool) ast.WalkStatus {
		switch n.Type {
		case ast.NodeDocument:
			if entering {
				if needRoot(n) {
					buf.WriteString("{\"name\": \"Root\", \"children\": [")
				}
			} else {
				if needRoot(n) {
					buf.WriteString("]}")
				}
			}
			return ast.WalkContinue
		case ast.NodeList:
			return ast.WalkContinue
		case ast.NodeListItem:
			children := nil != n.ChildByType(ast.NodeList)
			if entering {
				buf.WriteString("{\"name\": \"" + text(n.FirstChild) + "\"")
				if children {
					buf.WriteString(", \"children\": [")
				}
			} else {
				if children {
					buf.WriteString("]")
				}
				buf.WriteString("}")
				if nil != n.Next || nil != n.Parent.Next {
					buf.WriteString(", ")
				}
			}
		default:
			return ast.WalkContinue
		}
		return ast.WalkContinue
	})
	return buf.Bytes()
}

func text(listItemFirstChild *ast.Node) (ret string) {
	if nil == listItemFirstChild {
		return ""
	}

	buf := &bytes.Buffer{}
	ast.Walk(listItemFirstChild, func(n *ast.Node, entering bool) ast.WalkStatus {
		if ast.NodeList == n.Type || ast.NodeListItem == n.Type {
			return ast.WalkContinue
		}

		if (ast.NodeText == n.Type || ast.NodeLinkText == n.Type) && entering {
			buf.Write(n.Tokens)
		}
		return ast.WalkContinue
	})

	ret = buf.String()
	ret = strings.ReplaceAll(ret, "\\", "\\\\")
	ret = strings.ReplaceAll(ret, "\"", "\\\"")
	ret = strings.ReplaceAll(ret, editor.Caret, "")
	return
}

func needRoot(root *ast.Node) bool {
	count := 0

	for c := root.FirstChild; nil != c; c = c.Next {
		if ast.NodeList == c.Type {
			count++
		}
	}
	if 1 < count {
		return true
	}
	if 0 == count {
		return true
	}

	count = 0

	for c := root.FirstChild.FirstChild; nil != c; c = c.Next {
		if ast.NodeListItem == c.Type {
			count++
		}
	}
	if 1 < count {
		return true
	}
	return false
}
