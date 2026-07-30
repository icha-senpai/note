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

func ParagraphContinue(p *ast.Node, context *Context) int {
	if context.blank {
		return 1
	}
	return 0
}

func paragraphFinalize(p *ast.Node, context *Context) (insertTable bool) {
	if context.ParseOption.ParagraphBeginningSpace {
		_, p.Tokens = lex.TrimRight(p.Tokens)
	} else {
		p.Tokens = lex.TrimWhitespace(p.Tokens)
	}

	hasReferenceDefs := false
	for tokens := p.Tokens; 0 < len(tokens) && lex.ItemOpenBracket == tokens[0]; tokens = p.Tokens {
		if tokens = context.parseLinkRefDef(tokens); nil != tokens {
			p.Tokens = tokens
			hasReferenceDefs = true
			continue
		}
		break
	}
	if hasReferenceDefs && lex.IsBlankLine(p.Tokens) {
		p.Unlink()
	}

	if context.ParseOption.KramdownBlockIAL && nil != context.Tip.Parent && ast.NodeListItem == context.Tip.Parent.Type && p == context.Tip.Parent.FirstChild {
		if ial := Tokens2IAL(p.Tokens); nil != ial {
			emptyP := &ast.Node{Type: ast.NodeParagraph, KramdownIAL: ial}
			m := IAL2Map(ial)
			emptyP.ID = m["id"]
			context.Tip.Parent.AppendChild(emptyP)
			emptyP.InsertAfter(&ast.Node{Type: ast.NodeKramdownBlockIAL, Tokens: IAL2Tokens(ial)})
			p.Unlink()
			return
		}
	}

	if context.ParseOption.GFMTaskListItem {
		if listItem := p.Parent; nil != listItem && ast.NodeListItem == listItem.Type && listItem.FirstChild == p {
			if 3 == listItem.ListData.Typ {
				isEditor := context.ParseOption.VditorWYSIWYG || context.ParseOption.VditorIR || context.ParseOption.VditorSV || context.ParseOption.ProtyleWYSIWYG
				isTaskListItem := 3 < len(p.Tokens)
				if context.ParseOption.ProtyleWYSIWYG {
					isTaskListItem = 3 <= len(p.Tokens)
				}
				if isTaskListItem {

					tokens := p.Tokens
					if context.ParseOption.KramdownBlockIAL {
						if ial := context.parseKramdownIALInListItem(tokens); 0 < len(ial) {
							tokens = tokens[bytes.Index(tokens, []byte("}"))+1:]
							p.KramdownIAL = ial
						}
					}

					if (3 == len(tokens) && tokens[0] == lex.ItemOpenBracket && tokens[2] == lex.ItemCloseBracket && tokens[1] != lex.ItemCloseBracket) ||
						(3 < len(tokens) && (lex.IsWhitespace(tokens[3]) || editor.CaretTokens[0] == tokens[3] || editor.CaretTokens[0] == tokens[2])) {
						var caretStartText, caretAfterCloseBracket, caretInBracket bool
						if context.ParseOption.VditorWYSIWYG || context.ParseOption.VditorIR || context.ParseOption.VditorSV || context.ParseOption.ProtyleWYSIWYG {
							closeBracket := bytes.IndexByte(tokens, lex.ItemCloseBracket)
							if bytes.HasPrefix(tokens, editor.CaretTokens) {
								tokens = bytes.ReplaceAll(tokens, editor.CaretTokens, nil)
								caretStartText = true
							} else if bytes.HasPrefix(tokens[closeBracket+1:], editor.CaretTokens) {
								tokens = bytes.ReplaceAll(tokens, editor.CaretTokens, nil)
								caretAfterCloseBracket = true
							} else if bytes.Contains(tokens[1:closeBracket], editor.CaretTokens) {
								tokens = bytes.ReplaceAll(tokens, editor.CaretTokens, nil)
								caretInBracket = true
							}
						}
						marker := tokens[1]
						if context.ParseOption.IsValidTaskListItemMarker(marker) {
							taskListItemMarker := &ast.Node{Type: ast.NodeTaskListItemMarker, Tokens: tokens[:3]}
							taskListItemMarker.ReviveFromMarker(marker)
							if context.ParseOption.ProtyleWYSIWYG {
								p.InsertBefore(taskListItemMarker)
							} else {
								p.PrependChild(taskListItemMarker)
							}
							p.Tokens = tokens[3:]
							if isEditor {
								p.Tokens = bytes.TrimSpace(p.Tokens)
								if caretStartText || caretAfterCloseBracket || caretInBracket {
									p.Tokens = append([]byte(" "+editor.Caret), p.Tokens...)
								} else {
									if !context.ParseOption.ProtyleWYSIWYG {
										p.Tokens = append([]byte(" "), p.Tokens...)
									}
								}
							}

							if 0 < len(p.Tokens) {
								subTree := Parse("", p.Tokens, context.ParseOption)
								subBlock := subTree.Root.FirstChild
								if ast.NodeParagraph != subBlock.Type {
									if !context.ParseOption.ProtyleWYSIWYG {
										// Protyle `Optimize typography` exception in case of task list and heading
										listItem.PrependChild(&ast.Node{Type: ast.NodeText, Tokens: []byte(" ")})
									}
									if nil != p.FirstChild {
										listItem.PrependChild(p.FirstChild)
									}
									subBlock.ID = p.ID
									subBlock.KramdownIAL = p.KramdownIAL

									// Incomplete data when pasting task list nested list
									var last *ast.Node
									var blocks []*ast.Node
									for b := subBlock; nil != b && ast.NodeDocument != b.Type; b = b.Next {
										if ast.NodeKramdownBlockIAL == b.Type {
											if util.IsDocIAL(b.Tokens) {
												break
											}
										}

										last = b
									}
									for b := last; nil != b; b = b.Previous {
										blocks = append(blocks, b)
									}
									for _, b := range blocks {
										p.InsertAfter(b)
									}

									p.Unlink()
								}
							}
						}
					}
				}
			}
		}
	}

	if context.ParseOption.GFMTable {
		if paragraph, table := context.parseTable(p); nil != table {
			if nil != paragraph {
				p.Tokens = paragraph.Tokens
				p.InsertAfter(table)
				table.Close = true
				context.Tip = table
				return true
			} else {
				p.Type = ast.NodeTable
				p.TableAligns = table.TableAligns
				for tr := table.FirstChild; nil != tr; {
					nextTr := tr.Next
					p.AppendChild(tr)
					tr = nextTr
				}
			}
			return
		}
	}

	if context.ParseOption.ToC {
		if toc := context.parseToC(p); nil != toc {
			p.Type = ast.NodeToC
			return
		}
	}
	return
}
