// Copyright (c) 2019-present, Scribli


//go:build javascript
// +build javascript

package render

import (
	"github.com/icha-senpai/note/third_party/forks/lute/ast"
	"github.com/icha-senpai/note/third_party/forks/lute/html"
	"github.com/icha-senpai/note/third_party/forks/lute/lex"
)

func (r *HtmlRenderer) renderCodeBlock(node *ast.Node, entering bool) ast.WalkStatus {
	r.Newline()

	if !node.IsFencedCodeBlock {
		if entering {
			r.WriteString("<pre><code>")
			r.Write(html.EscapeHTML(node.FirstChild.Tokens))
			r.WriteString("</code></pre>")
			r.Newline()
			return ast.WalkSkipChildren
		} else {
			return ast.WalkContinue
		}
	}
	return ast.WalkContinue
}

func (r *HtmlRenderer) renderCodeBlockCode(node *ast.Node, entering bool) ast.WalkStatus {
	var language string
	if 0 < len(node.Previous.CodeBlockInfo) {
		infoWords := lex.Split(node.Previous.CodeBlockInfo, lex.ItemSpace)
		language = string(infoWords[0])
	}
	preDiv := NoHighlight(language)

	if entering {
		r.Newline()
		var attrs [][]string
		r.handleKramdownBlockIAL(node)
		attrs = append(attrs, node.KramdownIAL...)
		if !preDiv {
			r.Tag("pre", attrs, false)
		}
		tokens := node.Tokens
		if 0 < len(node.Previous.CodeBlockInfo) {
			if "mindmap" == language {
				json := EChartsMindmap(tokens)
				r.WriteString("<div data-code=\"")
				r.Write(json)
				r.WriteString("\" class=\"language-mindmap\">")
			} else {
				if preDiv {
					r.WriteString("<div class=\"language-" + language + "\">")
				} else {
					r.WriteString("<code class=\"language-" + language + "\">")
				}
			}
			tokens = html.EscapeHTML(tokens)
			r.Write(tokens)
		} else {
			r.WriteString("<code>")
			tokens = html.EscapeHTML(tokens)
			r.Write(tokens)
		}
	} else {
		if preDiv {
			r.WriteString("</div>")
		} else {
			r.WriteString("</code></pre>")
		}
		r.Newline()
	}
	return ast.WalkContinue
}
