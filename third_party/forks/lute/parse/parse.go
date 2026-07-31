// Copyright (c) 2019-present, Scribli


package parse

import (
	"sync"

	"github.com/icha-senpai/note/third_party/forks/lute/ast"
	"github.com/icha-senpai/note/third_party/forks/lute/editor"
	"github.com/icha-senpai/note/third_party/forks/lute/lex"
)

func Parse(name string, markdown []byte, options *Options) (tree *Tree) {
	tree = &Tree{Name: name, Context: &Context{ParseOption: options}}
	tree.Context.Tree = tree
	tree.lexer = lex.NewLexer(markdown)
	tree.Root = &ast.Node{Type: ast.NodeDocument}
	tree.parseBlocks()
	tree.parseInlines()
	tree.finalParseBlockIAL()
	tree.lexer = nil
	return
}

func (t *Tree) finalParseBlockIAL() {
	if !t.Context.ParseOption.KramdownBlockIAL {
		return
	}

	var appends []*ast.Node

	ast.Walk(t.Root, func(n *ast.Node, entering bool) ast.WalkStatus {
		if !entering || !n.IsBlock() || ast.NodeKramdownBlockIAL == n.Type {
			return ast.WalkContinue
		}

		if ast.NodeBlockquote == n.Type && nil != n.FirstChild && nil == n.FirstChild.Next {
			appends = append(appends, n)
		}

		if "" == n.ID {
			id := n.IALAttr("id")
			if "" == id {
				id = ast.NewNodeID()
			}
			n.ID = id

			if t.Context.ParseOption.ProtyleWYSIWYG && t.Context.ParseOption.Spin &&
				ast.NodeDocument != n.Type && nil != n.Next && ast.NodeKramdownBlockIAL != n.Next.Type && "" != n.Next.ID {
				n.ID = n.Next.ID
				n.KramdownIAL = n.Next.KramdownIAL
				if "" == n.IALAttr("updated") {
					n.SetIALAttr("updated", n.ID[:14])
				}
				n.Next.ID = ast.NewNodeID()
				n.Next.KramdownIAL = nil
				n.Next.SetIALAttr("id", n.Next.ID)
				n.Next.SetIALAttr("updated", n.Next.ID[:14])
				if nil != n.Next.Next && ast.NodeKramdownBlockIAL == n.Next.Next.Type {
					n.Next.Next.Tokens = IAL2Tokens(n.Next.KramdownIAL)
				}
				n.InsertAfter(&ast.Node{Type: ast.NodeKramdownBlockIAL, Tokens: IAL2Tokens(n.KramdownIAL)})
				return ast.WalkContinue
			}
		}

		ial := n.Next
		if nil == ial || ast.NodeKramdownBlockIAL != ial.Type {
			if t.Context.ParseOption.ProtyleWYSIWYG {
				n.SetIALAttr("id", n.ID)
				n.SetIALAttr("updated", n.ID[:14])
			}
			return ast.WalkContinue
		}

		n.KramdownIAL = Tokens2IAL(ial.Tokens)
		if "" == n.IALAttr("updated") && t.Context.ParseOption.ProtyleWYSIWYG {
			n.SetIALAttr("updated", n.ID[:14])
			ial.Tokens = IAL2Tokens(n.KramdownIAL)
		}
		return ast.WalkContinue
	})

	for _, n := range appends {
		id := ast.NewNodeID()
		ialTokens := []byte("{: id=\"" + id + "\"}")
		p := &ast.Node{Type: ast.NodeParagraph, ID: id}
		p.KramdownIAL = [][]string{{"id", id}, {"updated", id[:14]}}
		p.ID = id
		p.InsertAfter(&ast.Node{Type: ast.NodeKramdownBlockIAL, Tokens: ialTokens})
		if nil != n.Next && ast.NodeKramdownBlockIAL == n.Next.Type &&
			ast.NodeBlockquote == n.Type && nil != n.FirstChild && ast.NodeBlockquoteMarker == n.FirstChild.Type &&
			nil == n.FirstChild.Next {
			text := &ast.Node{Type: ast.NodeText, Tokens: editor.CaretTokens}
			p.AppendChild(text)
		}
		n.AppendChild(p)
	}

	var docIAL *ast.Node
	var id string
	if nil != t.Context.rootIAL {
		docIAL = t.Context.rootIAL
	} else {
		id = ast.NewNodeID()
		docIAL = &ast.Node{Type: ast.NodeKramdownBlockIAL, Tokens: []byte("{: id=\"" + id + "\" updated=\"" + id[:14] + "\" type=\"doc\"}")}
		t.Root.ID = id
		t.ID = id
	}
	t.Root.AppendChild(docIAL)
}

func Block(name string, markdown []byte, options *Options) (tree *Tree) {
	tree = &Tree{Name: name, Context: &Context{ParseOption: options}}
	tree.Context.Tree = tree
	tree.lexer = lex.NewLexer(markdown)
	tree.Root = &ast.Node{Type: ast.NodeDocument}
	tree.parseBlocks()
	tree.finalParseBlockIAL()
	tree.lexer = nil
	return
}

func Inline(name string, markdown []byte, options *Options) (tree *Tree) {
	tree = &Tree{Name: name, Context: &Context{ParseOption: options}}
	tree.Context.Tree = tree
	tree.Root = &ast.Node{Type: ast.NodeDocument}
	tree.Root.AppendChild(&ast.Node{Type: ast.NodeParagraph, Tokens: markdown})
	tree.parseInlines()
	tree.lexer = nil
	return
}

type Context struct {
	Tree        *Tree
	ParseOption *Options

	Tip                                                      *ast.Node
	oldtip                                                   *ast.Node
	currentLine                                              []byte
	currentLineLen                                           int
	offset, column, nextNonspace, nextNonspaceColumn, indent int
	indented, blank, partiallyConsumedTab, allClosed         bool
	lastMatchedContainer                                     *ast.Node

	rootIAL *ast.Node
}

type InlineContext struct {
	tokens     []byte
	tokensLen  int
	pos        int
	delimiters *delimiter
	brackets   *delimiter
}

func (context *Context) advanceOffset(count int, columns bool) {
	currentLine := context.currentLine
	var charsToTab, charsToAdvance int
	var c byte
	for 0 < count {
		c = currentLine[context.offset]
		if lex.ItemTab == c {
			charsToTab = 4 - (context.column % 4)
			if columns {
				context.partiallyConsumedTab = charsToTab > count
				if context.partiallyConsumedTab {
					charsToAdvance = count
				} else {
					charsToAdvance = charsToTab
					context.offset++
				}
				context.column += charsToAdvance
				count -= charsToAdvance
			} else {
				context.partiallyConsumedTab = false
				context.column += charsToTab
				context.offset++
				count--
			}
		} else {
			context.partiallyConsumedTab = false
			context.offset++
			context.column++
			count--
		}
	}
}

func (context *Context) advanceNextNonspace() {
	context.offset = context.nextNonspace
	context.column = context.nextNonspaceColumn
	context.partiallyConsumedTab = false
}

func (context *Context) findNextNonspace() {
	i := context.offset
	cols := context.column

	var token byte
	for {
		token = context.currentLine[i]
		if lex.ItemSpace == token {
			i++
			cols++
		} else if lex.ItemTab == token {
			i++
			cols += 4 - (cols % 4)
		} else {
			break
		}
	}

	context.blank = lex.ItemNewline == token
	context.nextNonspace = i
	context.nextNonspaceColumn = cols
	context.indent = context.nextNonspaceColumn - context.column
	context.indented = 4 <= context.indent
}

func (context *Context) closeUnmatchedBlocks() {
	if !context.allClosed {
		for context.oldtip != context.lastMatchedContainer {
			parent := context.oldtip.Parent
			context.finalize(context.oldtip)
			context.oldtip = parent
		}
		context.allClosed = true
	}
}

func (context *Context) closeSuperBlockChildren() {
	for n := context.Tip; nil != n && ast.NodeSuperBlock != n.Type; n = n.Parent {
		context.finalize(n)
	}
}

func (context *Context) finalize(block *ast.Node) {
	parent := block.Parent
	block.Close = true

	switch block.Type {
	case ast.NodeCodeBlock:
		context.codeBlockFinalize(block)
	case ast.NodeHTMLBlock, ast.NodeIFrame, ast.NodeVideo, ast.NodeAudio, ast.NodeWidget:
		context.htmlBlockFinalize(block)
	case ast.NodeParagraph:
		insertTable := paragraphFinalize(block, context)
		if insertTable {
			return
		}
	case ast.NodeMathBlock:
		context.mathBlockFinalize(block)
	case ast.NodeYamlFrontMatter:
		context.yamlFrontMatterFinalize(block)
	case ast.NodeList:
		context.listFinalize(block)
	case ast.NodeSuperBlock:
		context.superBlockFinalize(block)
	case ast.NodeGitConflict:
		context.gitConflictFinalize(block)
	case ast.NodeCustomBlock:
		context.customBlockFinalize(block)
	case ast.NodeCallout:
		context.calloutFinalize(block)
	case ast.NodeBlockquote:
		context.blockquoteFinalize(block)
	}

	context.Tip = parent
}

func (context *Context) addChildMarker(nodeType ast.NodeType, tokens []byte) (ret *ast.Node) {
	ret = &ast.Node{Type: nodeType, Tokens: tokens, Close: true}
	context.Tip.AppendChild(ret)
	return
}

func (context *Context) addChild(nodeType ast.NodeType) (ret *ast.Node) {
	for !context.Tip.CanContain(nodeType) {
		context.finalize(context.Tip)
	}

	ret = &ast.Node{Type: nodeType}
	context.Tip.AppendChild(ret)
	context.Tip = ret
	return
}

func (context *Context) listsMatch(listData, itemData *ast.ListData) bool {
	return listData.Typ == itemData.Typ &&
		((0 == listData.Delimiter && 0 == itemData.Delimiter) || listData.Delimiter == itemData.Delimiter) &&
		listData.BulletChar == itemData.BulletChar
}

type Tree struct {
	Root          *ast.Node
	Context       *Context
	lexer         *lex.Lexer
	inlineContext *InlineContext

	Name    string
	ID      string // ID
	Box     string
	Path    string
	HPath   string
	Marks   []string
	Created int64
	Updated int64
	Hash    string
}

type Options struct {
	GFMTable                            bool
	GFMTaskListItem                     bool
	GFMStrikethrough                    bool
	GFMStrikethrough1                   bool
	GFMAutoLink                         bool
	Footnotes                           bool
	HeadingID                           bool
	ToC                                 bool
	Emoji                               bool
	AliasEmoji                          map[string]string
	EmojiAlias                          map[string]string
	EmojiSite                           string
	VditorWYSIWYG                       bool
	VditorIR                            bool
	VditorSV                            bool
	ProtyleWYSIWYG                      bool
	ProtyleWYSIWYGAutoLink              bool
	InlineMath                          bool
	InlineMathAllowDigitAfterOpenMarker bool
	Setext                              bool
	YamlFrontMatter                     bool
	BlockRef                            bool
	FileAnnotationRef                   bool
	Mark                                bool
	KramdownBlockIAL                    bool
	KramdownSpanIAL                     bool
	Tag                                 bool
	ImgPathAllowSpace                   bool
	SuperBlock                          bool
	Sup                                 bool
	Sub                                 bool
	InlineAsterisk                      bool
	InlineUnderscore                    bool
	GitConflict                         bool
	LinkRef                             bool
	IndentCodeBlock                     bool
	ParagraphBeginningSpace             bool
	DataImage                           bool
	TextMark                            bool
	//
	HTMLTag2TextMark bool
	//
	Spin                        bool
	HTML2MarkdownAttrs          []string
	Callout                     bool
	KeepEscaped                 bool
	ArbitraryTaskListItemMarker bool
	EnsureListItemParagraph     bool
}

func (options *Options) IsValidTaskListItemMarker(marker byte) bool {
	return (' ' == marker || 'x' == marker || 'X' == marker) ||
		(options.ArbitraryTaskListItemMarker && '[' != marker && ']' != marker)
}

var EmojiLock = sync.Mutex{}

func NewOptions() *Options {
	return &Options{
		GFMTable:          true,
		GFMTaskListItem:   true,
		GFMStrikethrough:  true,
		GFMStrikethrough1: true,
		GFMAutoLink:       true,
		Footnotes:         true,
		Emoji:             true,
		AliasEmoji:        EmojiAliasUnicode,
		EmojiAlias:        EmojiUnicodeAlias,
		EmojiSite:         "https://cdn.jsdelivr.net/npm/vditor/dist/images/emoji",
		InlineMath:        true,
		Setext:            true,
		YamlFrontMatter:   true,
		BlockRef:          false,
		FileAnnotationRef: false,
		Mark:              false,
		InlineAsterisk:    true,
		InlineUnderscore:  true,
		KramdownBlockIAL:  false,
		HeadingID:         true,
		LinkRef:           true,
		IndentCodeBlock:   true,
		DataImage:         true,
		Callout:           false,
	}
}

func (context *Context) ParentTip() {
	if tip := context.Tip.Parent; nil != tip {
		context.Tip = context.Tip.Parent
	}
}

func (context *Context) TipAppendChild(child *ast.Node) {
	context.Tip.AppendChild(child)
}
