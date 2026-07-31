// Copyright (c) 2019-present, Scribli


package parse

import (
	"bytes"

	"github.com/icha-senpai/note/third_party/forks/lute/ast"
	"github.com/icha-senpai/note/third_party/forks/lute/lex"
	"github.com/icha-senpai/note/third_party/forks/lute/util"
)

func (t *Tree) emoji(node *ast.Node) {
	for child := node.FirstChild; nil != child; {
		next := child.Next
		if ast.NodeText == child.Type || ast.NodeLinkText == child.Type {
			t.emoji0(child)
		} else {
			t.emoji(child)
		}
		child = next
	}
}

var EmojiSitePlaceholder = util.StrToBytes("${emojiSite}")
var emojiDot = util.StrToBytes(".")

func (t *Tree) emoji0(node *ast.Node) {
	first := node
	tokens := node.Tokens
	node.Tokens = []byte{}
	length := len(tokens)
	var token byte
	var maybeEmoji []byte
	var pos int
	for i := 0; i < length; {
		token = tokens[i]
		if i == length-1 {
			node.Tokens = append(node.Tokens, tokens[pos:]...)
			break
		}

		if lex.ItemColon != token {
			i++
			continue
		}

		node.Tokens = append(node.Tokens, tokens[pos:i]...)

		matchCloseColon := false
		for pos = i + 1; pos < length; pos++ {
			token = tokens[pos]
			if lex.IsWhitespace(token) {
				break
			}
			if lex.ItemColon == token {
				matchCloseColon = true
				break
			}
		}
		if !matchCloseColon {
			node.Tokens = append(node.Tokens, tokens[i:pos]...)
			i++
			continue
		}

		maybeEmoji = tokens[i+1 : pos]
		if 1 > len(maybeEmoji) {
			node.Tokens = append(node.Tokens, tokens[pos])
			i++
			continue
		}

		EmojiLock.Lock()
		emoji, ok := t.Context.ParseOption.AliasEmoji[util.BytesToStr(maybeEmoji)]
		EmojiLock.Unlock()
		if ok {
			emojiNode := &ast.Node{Type: ast.NodeEmoji}
			emojiUnicodeOrImg := &ast.Node{Type: ast.NodeEmojiUnicode}
			emojiNode.AppendChild(emojiUnicodeOrImg)
			emojiTokens := util.StrToBytes(emoji)
			if bytes.Contains(emojiTokens, EmojiSitePlaceholder) {
				alias := util.BytesToStr(maybeEmoji)
				suffix := ".png"
				if "huaji" == alias {
					suffix = ".gif"
				}
				src := t.Context.ParseOption.EmojiSite + "/" + alias + suffix
				emojiUnicodeOrImg.Type = ast.NodeEmojiImg
				emojiUnicodeOrImg.Tokens = t.EmojiImgTokens(alias, src)
			} else if bytes.Contains(emojiTokens, emojiDot) {
				alias := util.BytesToStr(maybeEmoji)
				emojiUnicodeOrImg.Type = ast.NodeEmojiImg
				emojiUnicodeOrImg.Tokens = t.EmojiImgTokens(alias, emoji)
			} else {
				emojiUnicodeOrImg.Tokens = emojiTokens
			}

			emojiUnicodeOrImg.AppendChild(&ast.Node{Type: ast.NodeEmojiAlias, Tokens: tokens[i : pos+1]})
			node.InsertAfter(emojiNode)

			if pos+1 < length {
				text := &ast.Node{Type: ast.NodeText, Tokens: []byte{}}
				emojiNode.InsertAfter(text)
				node = text
			}
		} else {
			node.Tokens = append(node.Tokens, tokens[i:pos+1]...)
		}

		pos++
		i = pos
	}

	if 1 > len(first.Tokens) {
		first.Unlink()
	}
	if 1 > len(node.Tokens) {
		node.Unlink()
	}
}

func (t *Tree) EmojiImgTokens(alias, src string) []byte {
	return util.StrToBytes("<img alt=\"" + alias + "\" class=\"emoji\" src=\"" + src + "\" title=\"" + alias + "\" />")
}
