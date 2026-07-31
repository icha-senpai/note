// Copyright (c) 2019-present, Scribli
//
// Lute is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//         http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package ast

import (
	"bytes"
	"math/rand"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/icha-senpai/note/third_party/forks/lute/editor"
	"github.com/icha-senpai/note/third_party/forks/lute/html"
	"github.com/icha-senpai/note/third_party/forks/lute/lex"
	"github.com/icha-senpai/note/third_party/forks/lute/util"
)

type Node struct {
	ID   string `json:",omitempty"`
	Box  string `json:"-"`
	Path string `json:"-"`
	Spec string `json:",omitempty"`

	Type       NodeType `json:"-"`
	Parent     *Node    `json:"-"`
	Previous   *Node    `json:"-"`
	Next       *Node    `json:"-"`
	FirstChild *Node    `json:"-"`
	LastChild  *Node    `json:"-"`
	Children   []*Node  `json:",omitempty"`
	Tokens     []byte   `json:"-"`
	TypeStr    string   `json:"Type"`
	Data       string   `json:"Data,omitempty"`

	Close           bool `json:"-"`
	LastLineBlank   bool `json:"-"`
	LastLineChecked bool `json:"-"`

	CodeMarkerLen int `json:",omitempty"`

	IsFencedCodeBlock  bool `json:",omitempty"`
	CodeBlockFenceChar byte `json:",omitempty"`

	CodeBlockFenceLen    int    `json:",omitempty"`
	CodeBlockFenceOffset int    `json:",omitempty"`
	CodeBlockOpenFence   []byte `json:",omitempty"`
	CodeBlockInfo        []byte `json:",omitempty"`
	CodeBlockCloseFence  []byte `json:",omitempty"`

	HtmlBlockType int `json:",omitempty"`

	ListData *ListData `json:",omitempty"`

	TaskListItemChecked bool `json:",omitempty"`
	TaskListItemMarker  byte `json:",omitempty"`

	TableAligns              []int `json:",omitempty"`
	TableCellAlign           int   `json:",omitempty"`
	TableCellContentWidth    int   `json:",omitempty"`
	TableCellContentMaxWidth int   `json:",omitempty"`

	LinkType     int    `json:",omitempty"`
	LinkRefLabel []byte `json:",omitempty"`

	HeadingLevel        int    `json:",omitempty"` // 1~6
	HeadingSetext       bool   `json:",omitempty"`
	HeadingNormalizedID string `json:",omitempty"`

	MathBlockDollarOffset int `json:",omitempty"`

	FootnotesRefLabel []byte  `json:",omitempty"`
	FootnotesRefId    string  `json:",omitempty"`
	FootnotesRefs     []*Node `json:",omitempty"`

	HtmlEntityTokens []byte `json:",omitempty"`

	KramdownIAL [][]string        `json:"-"`
	Properties  map[string]string `json:",omitempty"`

	TextMarkType                string `json:",omitempty"`
	TextMarkAHref               string `json:",omitempty"`
	TextMarkATitle              string `json:",omitempty"`
	TextMarkInlineMathContent   string `json:",omitempty"`
	TextMarkInlineMemoContent   string `json:",omitempty"`
	TextMarkBlockRefID          string `json:",omitempty"`
	TextMarkBlockRefSubtype     string `json:",omitempty"`
	TextMarkFileAnnotationRefID string `json:",omitempty"`
	TextMarkTextContent         string `json:",omitempty"`

	AttributeViewID   string `json:",omitempty"`
	AttributeViewType string `json:",omitempty"`

	CustomBlockFenceOffset int    `json:",omitempty"`
	CustomBlockInfo        string `json:",omitempty"`

	CalloutType     string `json:",omitempty"`
	CalloutTitle    string `json:",omitempty"`
	CalloutIcon     string `json:",omitempty"`
	CalloutIconType int    `json:",omitempty"`
}

func (n *Node) EffectiveTaskListItemMarker() string {
	if n.TaskListItemMarker != 0 {
		if n.TaskListItemMarker == 'x' {
			return "X"
		}
		return html.EscapeHTMLStr(string(n.TaskListItemMarker))
	}
	if n.TaskListItemChecked {
		return "X"
	}
	return " "
}

func (n *Node) ReviveFromMarker(marker byte) {
	n.TaskListItemMarker = marker
	n.TaskListItemChecked = marker != ' ' && marker != 0
}

func (n *Node) ReviveFromDataTask(dataTask string, checked bool) {
	var marker byte
	if 1 == len(dataTask) {
		marker = dataTask[0]
	} else if checked {
		marker = 'X'
	} else {
		marker = ' '
	}
	n.ReviveFromMarker(marker)
}

const (
	CalloutTypeNote      = "NOTE"
	CalloutTypeTip       = "TIP"
	CalloutTypeImportant = "IMPORTANT"
	CalloutTypeWarning   = "WARNING"
	CalloutTypeCaution   = "CAUTION"
)

func IsBuiltInCalloutType(typ string) bool {
	switch typ {
	case CalloutTypeNote, CalloutTypeTip, CalloutTypeImportant, CalloutTypeWarning, CalloutTypeCaution:
		return true
	}
	return false
}

func GetCalloutIcon(typ string) string {
	switch typ {
	case CalloutTypeNote:
		return "✏️"
	case CalloutTypeTip:
		return "💡"
	case CalloutTypeImportant:
		return "❗"
	case CalloutTypeWarning:
		return "⚠️"
	case CalloutTypeCaution:
		return "🚨"
	}
	return ""
}

func GetCalloutTitle(typ string) string {
	switch typ {
	case CalloutTypeNote:
		return "Note"
	case CalloutTypeTip:
		return "Tip"
	case CalloutTypeImportant:
		return "Important"
	case CalloutTypeWarning:
		return "Warning"
	case CalloutTypeCaution:
		return "Caution"
	}
	return ""
}

type ListData struct {
	Typ          int    `json:",omitempty"`
	Tight        bool   `json:",omitempty"`
	BulletChar   byte   `json:",omitempty"`
	Start        int    `json:",omitempty"`
	Delimiter    byte   `json:",omitempty"`
	Padding      int    `json:",omitempty"`
	MarkerOffset int    `json:",omitempty"`
	Checked      bool   `json:",omitempty"`
	Marker       []byte `json:",omitempty"`
	Num          int    `json:",omitempty"`
}

var Testing bool

func NewNodeID() string {
	if Testing {
		return "20060102150405-1a2b3c4"
	}
	now := time.Now()
	return now.Format("20060102150405") + "-" + randStr(7)
}

func IsNodeIDPattern(str string) bool {
	if len("20060102150405-1a2b3c4") != len(str) {
		return false
	}

	if 1 != strings.Count(str, "-") {
		return false
	}

	parts := strings.Split(str, "-")
	idPart := parts[0]
	if 14 != len(idPart) {
		return false
	}

	for _, c := range idPart {
		if !('0' <= c && '9' >= c) {
			return false
		}
	}

	randPart := parts[1]
	if 7 != len(randPart) {
		return false
	}

	for _, c := range randPart {
		if !('a' <= c && 'z' >= c) && !('0' <= c && '9' >= c) {
			return false
		}
	}
	return true
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())

	for t := NodeDocument; t < NodeTypeMaxVal; t++ {
		strNodeTypeMap[t.String()] = t
	}
}

func randStr(length int) string {
	letter := []rune("abcdefghijklmnopqrstuvwxyz0123456789")
	b := make([]rune, length)
	for i := range b {
		b[i] = letter[rand.Intn(len(letter))]
	}
	return string(b)
}

func (n *Node) Marker(entering bool) (ret string) {
	switch n.Type {
	case NodeTagOpenMarker, NodeTagCloseMarker:
		if entering {
			return "#"
		}
	case NodeEmA6kOpenMarker, NodeEmA6kCloseMarker:
		if entering {
			return "*"
		}
	case NodeEmU8eOpenMarker, NodeEmU8eCloseMarker:
		if entering {
			return "_"
		}
	case NodeStrongA6kOpenMarker, NodeStrongA6kCloseMarker:
		if entering {
			return "**"
		}
	case NodeStrongU8eOpenMarker, NodeStrongU8eCloseMarker:
		if entering {
			return "__"
		}
	case NodeStrikethrough2OpenMarker, NodeStrikethrough2CloseMarker:
		if entering {
			return "~~"
		}
	case NodeSupOpenMarker, NodeSupCloseMarker:
		if entering {
			return "^"
		}
	case NodeSubOpenMarker, NodeSubCloseMarker:
		if entering {
			return "~"
		}
	case NodeInlineMathOpenMarker, NodeInlineMathCloseMarker:
		if entering {
			return "$"
		}
	case NodeKbdOpenMarker:
		if entering {
			return "<kbd>"
		}
	case NodeKbdCloseMarker:
		if entering {
			return "</kbd>"
		}
	case NodeUnderlineOpenMarker:
		if entering {
			return "<u>"
		}
	case NodeUnderlineCloseMarker:
		if entering {
			return "</u>"
		}
	case NodeMark2OpenMarker, NodeMark2CloseMarker:
		if entering {
			return "=="
		}
	case NodeBang:
		if entering {
			return "!"
		}
	case NodeOpenBracket:
		if entering {
			return "["
		}
	case NodeCloseBracket:
		if entering {
			return "]"
		}
	case NodeOpenParen:
		if entering {
			return "("
		}
	case NodeCloseParen:
		if entering {
			return ")"
		}
	}

	return ""
}

func (n *Node) ContainTextMarkTypes(types ...string) bool {
	nodeTypes := strings.Split(n.TextMarkType, " ")
	for _, typ := range types {
		for _, nodeType := range nodeTypes {
			if typ == nodeType {
				return true
			}
		}
	}
	return false
}

func (n *Node) IsTextMarkType(typ string) bool {
	types := strings.Split(n.TextMarkType, " ")
	for _, t := range types {
		if typ == t {
			return true
		}
	}
	return false
}

func (n *Node) IsNextSameInlineMemo() bool {
	if nil == n {
		return false
	}

	var nextInlineMemo *Node
	for node := n.Next; nil != node; node = node.Next {
		if nil == n.Next || NodeKramdownSpanIAL == node.Type || nil == node.Next || NodeKramdownSpanIAL == node.Next.Type {
			continue
		}

		if NodeTextMark == node.Type && node.IsTextMarkType("inline-memo") {
			nextInlineMemo = node
			break
		}
	}

	if nil != nextInlineMemo && n.TextMarkInlineMemoContent == nextInlineMemo.TextMarkInlineMemoContent {
		return true
	}
	return false
}

func (n *Node) IsSameTextMarkType(node *Node) bool {
	if "" == n.TextMarkType || "" == node.TextMarkType {
		return false
	}

	a := strings.Split(n.TextMarkType, " ")
	b := strings.Split(node.TextMarkType, " ")
	if len(a) != len(b) {
		return false
	}
	sort.Strings(a)
	sort.Strings(b)
	for i := range a {
		if a[i] != b[i] {
			return false
		}

		switch a[i] {
		case "block-ref":
			if n.TextMarkBlockRefID != node.TextMarkBlockRefID {
				return false
			}
		case "a":
			if n.TextMarkAHref != node.TextMarkAHref || node.TextMarkATitle != node.TextMarkATitle {
				return false
			}
		case "inline-memo":
			if n.TextMarkInlineMemoContent != node.TextMarkInlineMemoContent {
				return false
			}
		}
	}
	return true
}

func (n *Node) SortTextMarkDataTypes() {
	if "" == n.TextMarkTextContent {
		return
	}

	dataTypes := strings.Split(n.TextMarkType, " ")
	sort.Strings(dataTypes)
	n.TextMarkType = strings.Join(dataTypes, " ")
}

func (n *Node) RemoveIALAttr(name string) {
	tmp := n.KramdownIAL[:0]
	for _, kv := range n.KramdownIAL {
		if name != kv[0] {
			tmp = append(tmp, kv)
		}
	}
	n.KramdownIAL = tmp
}

func (n *Node) RemoveIALAttrsByPrefix(prefix string) {
	tmp := n.KramdownIAL[:0]
	for _, kv := range n.KramdownIAL {
		if !strings.HasPrefix(kv[0], prefix) {
			tmp = append(tmp, kv)
		}
	}
	n.KramdownIAL = tmp
}

func (n *Node) SetIALAttr(name, value string) {
	value = html.EscapeAttrVal(value)
	for _, kv := range n.KramdownIAL {
		if name == kv[0] {
			kv[1] = value
			return
		}
	}
	n.KramdownIAL = append(n.KramdownIAL, []string{name, value})
}

func (n *Node) IALAttr(name string) string {
	for _, kv := range n.KramdownIAL {
		if name == kv[0] {
			return html.UnescapeAttrVal(kv[1])
		}
	}
	return ""
}

func (n *Node) IsEmptyBlockIAL() bool {
	if NodeKramdownBlockIAL != n.Type {
		return false
	}

	if util.IsDocIAL(n.Tokens) {
		return false
	}

	if nil != n.Previous {
		if NodeKramdownBlockIAL == n.Previous.Type {
			return true
		}
		return false
	}

	//
	if nil != n.Parent && !n.Parent.IsContainerBlock() {
		return false
	}
	return true
}

func (n *Node) TokensStr() string {
	return util.BytesToStr(n.Tokens)
}

func (n *Node) LastDeepestChild() (ret *Node) {
	if nil == n.LastChild {
		return n
	}
	return n.LastChild.LastDeepestChild()
}

func (n *Node) FirstDeepestChild() (ret *Node) {
	if nil == n.FirstChild {
		return n
	}
	return n.FirstChild.FirstDeepestChild()
}

func (n *Node) ChildByType(childType NodeType) *Node {
	for c := n.FirstChild; nil != c; c = c.Next {
		if c.Type == childType {
			return c
		}
	}
	return nil
}

func (n *Node) ChildrenByType(childType NodeType) (ret []*Node) {
	ret = []*Node{}
	Walk(n, func(n *Node, entering bool) WalkStatus {
		if (childType == n.Type) && entering {
			ret = append(ret, n)
		}
		return WalkContinue
	})
	return
}

func (n *Node) Text() (ret string) {
	buf := &bytes.Buffer{}
	Walk(n, func(n *Node, entering bool) WalkStatus {
		if !entering {
			return WalkContinue
		}
		switch n.Type {
		case NodeText, NodeLinkText, NodeBlockRefText, NodeBlockRefDynamicText, NodeFileAnnotationRefText, NodeFootnotesRef:
			buf.Write(n.Tokens)
		case NodeTextMark:
			buf.WriteString(n.TextMarkTextContent)
		}
		return WalkContinue
	})
	return buf.String()
}

func (n *Node) TextLen() (ret int) {
	buf := make([]byte, 0, 4096)
	Walk(n, func(n *Node, entering bool) WalkStatus {
		if !entering {
			return WalkContinue
		}
		switch n.Type {
		case NodeText, NodeLinkText, NodeBlockRefText, NodeBlockRefDynamicText, NodeFileAnnotationRefText, NodeFootnotesRef:
			buf = append(buf, n.Tokens...)
		case NodeTextMark:
			buf = append(buf, n.TextMarkTextContent...)
		}
		return WalkContinue
	})
	return utf8.RuneCount(buf)
}

func (n *Node) Content() (ret string) {
	buf := &bytes.Buffer{}
	Walk(n, func(n *Node, entering bool) WalkStatus {
		if !entering {
			if nil != n.Next && nil != n.Next.Next && 1 < buf.Len() && n.IsBlock() && buf.Bytes()[buf.Len()-1] != '\n' {
				buf.WriteByte('\n')
			}
			return WalkContinue
		}

		switch n.Type {
		case NodeText, NodeLinkText, NodeBlockRefText, NodeBlockRefDynamicText, NodeFileAnnotationRefText, NodeFootnotesRef,
			NodeCodeSpanContent, NodeCodeBlockCode, NodeInlineMathContent, NodeMathBlockContent,
			NodeHTMLEntity, NodeEmojiAlias, NodeEmojiUnicode, NodeBackslashContent, NodeYamlFrontMatterContent,
			NodeGitConflictContent:
			buf.Write(n.Tokens)
		case NodeTextMark:
			if "" != n.TextMarkTextContent {
				if n.IsTextMarkType("code") || n.IsTextMarkType("tag") || n.IsTextMarkType("strong") || n.IsTextMarkType("em") || n.IsTextMarkType("a") {
					buf.WriteString(html.UnescapeString(n.TextMarkTextContent))
				} else {
					buf.WriteString(n.TextMarkTextContent)
				}
			} else if "" != n.TextMarkInlineMathContent {
				content := n.TextMarkInlineMathContent
				content = strings.ReplaceAll(content, editor.IALValEscNewLine, " ")
				buf.WriteString(content)
			}
			if "" != n.TextMarkInlineMemoContent {
				content := n.TextMarkInlineMemoContent
				content = strings.ReplaceAll(content, editor.IALValEscNewLine, " ")
				buf.WriteString(content)
			}
		}
		return WalkContinue
	})

	return buf.String()
}

func (n *Node) EscapeMarkerContent() (ret string) {
	ret = n.Content()
	ret = string(lex.EscapeProtyleMarkers([]byte(ret)))
	return
}

func (n *Node) Stat() (runeCnt, wordCnt, linkCnt, imgCnt, refCnt int) {
	buf := make([]byte, 0, 8192)
	Walk(n, func(n *Node, entering bool) WalkStatus {
		if !entering {
			return WalkContinue
		}

		switch n.Type {
		case NodeText, NodeLinkText, NodeBlockRefText, NodeBlockRefDynamicText, NodeFileAnnotationRefText, NodeFootnotesRef,
			NodeCodeSpanContent, NodeCodeBlockCode, NodeInlineMathContent, NodeMathBlockContent,
			NodeHTMLEntity, NodeEmojiAlias, NodeEmojiUnicode, NodeBackslashContent, NodeYamlFrontMatterContent,
			NodeGitConflictContent:
			buf = append(buf, n.Tokens...)
		case NodeTextMark:
			if 0 < len(n.TextMarkTextContent) {
				buf = append(buf, n.TextMarkTextContent...)
			} else if 0 < len(n.TextMarkInlineMathContent) {
				content := n.TextMarkInlineMathContent
				content = strings.ReplaceAll(content, editor.IALValEscNewLine, " ")
				buf = append(buf, content...)
			} else if "" != n.TextMarkInlineMemoContent {
				content := n.TextMarkInlineMemoContent
				content = strings.ReplaceAll(content, editor.IALValEscNewLine, " ")
				buf = append(buf, content...)
			}

			if n.IsTextMarkType("a") {
				linkCnt++
			}
			if n.IsTextMarkType("block-ref") || n.IsTextMarkType("file-annotation-ref") {
				refCnt++
			}
		case NodeLink:
			linkCnt++
		case NodeImage:
			imgCnt++
		case NodeBlockRef:
			refCnt++
		}
		if n.IsBlock() {
			buf = append(buf, ' ')
		}
		return WalkContinue
	})

	buf = bytes.TrimSpace(buf)
	runeCnt, wordCnt = util.WordCount(util.BytesToStr(buf))
	return
}

func (n *Node) TokenLen() (ret int) {
	Walk(n, func(n *Node, entering bool) WalkStatus {
		if !entering {
			return WalkContinue
		}
		ret += lex.BytesShowLength(n.Tokens)
		return WalkContinue
	})
	return
}

func (n *Node) DocChild() (ret *Node) {
	ret = n
	for p := n; nil != p; p = p.Parent {
		if NodeDocument == p.Type {
			return
		}
		ret = p
	}
	return
}

func (n *Node) IsChildBlockOf(parent *Node, depth int) bool {
	if "" == n.ID || !n.IsBlock() {
		return false
	}

	if depth == 0 {
		for p := n.Parent; nil != p; p = p.Parent {
			if p == parent {
				return true
			}
		}
		return false
	}

	nodeParent := n.Parent
	for i := 1; i < depth; i++ {
		if nil == nodeParent {
			break
		}
		nodeParent = nodeParent.Parent
	}
	if parent != nodeParent {
		return false
	}
	return true
}

func (n *Node) NextNodeText() string {
	if nil == n.Next {
		return ""
	}
	return n.Next.Text()
}

func (n *Node) PreviousNodeText() string {
	prev := n.Previous
	if nil == prev {
		return ""
	}
	if NodeKramdownSpanIAL == prev.Type {
		prev = prev.Previous
	}
	if nil == prev {
		return ""
	}
	return prev.Text()
}

func (n *Node) Unlink() {
	if nil != n.Previous {
		n.Previous.Next = n.Next
	} else if nil != n.Parent {
		n.Parent.FirstChild = n.Next
	}
	if nil != n.Next {
		n.Next.Previous = n.Previous
	} else if nil != n.Parent {
		n.Parent.LastChild = n.Previous
	}
	n.Parent = nil
	n.Next = nil
	n.Previous = nil
}

func (n *Node) AppendTokens(tokens []byte) {
	n.Tokens = append(n.Tokens, string(tokens)...)
}

func (n *Node) PrependTokens(tokens []byte) {
	n.Tokens = append(tokens, n.Tokens...)
}

func (n *Node) InsertAfter(sibling *Node) {
	sibling.Unlink()
	sibling.Next = n.Next
	if nil != sibling.Next {
		sibling.Next.Previous = sibling
	}
	sibling.Previous = n
	n.Next = sibling
	sibling.Parent = n.Parent
	if nil != sibling.Parent && nil == sibling.Next && nil != sibling.Parent.LastChild {
		sibling.Parent.LastChild = sibling
	}
}

func (n *Node) InsertBefore(sibling *Node) {
	sibling.Unlink()
	sibling.Previous = n.Previous
	if nil != sibling.Previous {
		sibling.Previous.Next = sibling
	}
	sibling.Next = n
	n.Previous = sibling
	sibling.Parent = n.Parent
	if nil != sibling.Parent && nil == sibling.Previous {
		sibling.Parent.FirstChild = sibling
	}
}

func (n *Node) AppendChild(child *Node) {
	child.Unlink()
	child.Parent = n
	if nil != n.LastChild {
		n.LastChild.Next = child
		child.Previous = n.LastChild
		n.LastChild = child
	} else {
		n.FirstChild = child
		n.LastChild = child
	}
}

func (n *Node) PrependChild(child *Node) {
	child.Unlink()
	child.Parent = n
	if nil != n.FirstChild {
		n.FirstChild.Previous = child
		child.Next = n.FirstChild
		n.FirstChild = child
	} else {
		n.FirstChild = child
		n.LastChild = child
	}
}

func (n *Node) List() (ret []*Node) {
	ret = make([]*Node, 0, 512)
	Walk(n, func(n *Node, entering bool) WalkStatus {
		if entering {
			ret = append(ret, n)
		}
		return WalkContinue
	})
	return
}

func (n *Node) BlockIDs() []string {
	ret := make([]string, 0, 512)
	Walk(n, func(n *Node, entering bool) WalkStatus {
		if entering && n.IsBlock() && "" != n.ID {
			ret = append(ret, n.ID)
		}
		return WalkContinue
	})
	return ret
}

func (n *Node) ParentIs(nodeType NodeType, nodeTypes ...NodeType) bool {
	types := append(nodeTypes, nodeType)
	deep := 0
	for p := n.Parent; nil != p; p = p.Parent {
		for _, pt := range types {
			if pt == p.Type {
				return true
			}
		}
		deep++
		if 128 < deep {
			break
		}
	}
	return false
}

func (n *Node) IsBlock() bool {
	switch n.Type {
	case NodeDocument, NodeParagraph, NodeHeading, NodeThematicBreak, NodeBlockquote, NodeList, NodeListItem, NodeHTMLBlock,
		NodeCodeBlock, NodeTable, NodeMathBlock, NodeFootnotesDefBlock, NodeFootnotesDef, NodeToC, NodeYamlFrontMatter,
		NodeBlockQueryEmbed, NodeKramdownBlockIAL, NodeSuperBlock, NodeGitConflict, NodeAudio, NodeVideo, NodeIFrame, NodeWidget,
		NodeAttributeView, NodeCustomBlock, NodeCallout:
		return true
	}
	return false
}

func (n *Node) IsContainerBlock() bool {
	switch n.Type {
	case NodeDocument, NodeBlockquote, NodeList, NodeListItem, NodeFootnotesDefBlock, NodeFootnotesDef, NodeSuperBlock, NodeCallout:
		return true
	}
	return false
}

func (n *Node) IsMarker() bool {
	switch n.Type {
	case NodeHeadingC8hMarker, NodeBlockquoteMarker, NodeCodeBlockFenceOpenMarker, NodeCodeBlockFenceCloseMarker, NodeCodeBlockFenceInfoMarker,
		NodeEmA6kOpenMarker, NodeEmA6kCloseMarker, NodeEmU8eOpenMarker, NodeEmU8eCloseMarker, NodeStrongA6kOpenMarker, NodeStrongA6kCloseMarker,
		NodeStrongU8eOpenMarker, NodeStrongU8eCloseMarker, NodeCodeSpanOpenMarker, NodeCodeSpanCloseMarker, NodeTaskListItemMarker,
		NodeStrikethrough1OpenMarker, NodeStrikethrough1CloseMarker, NodeStrikethrough2OpenMarker, NodeStrikethrough2CloseMarker,
		NodeMathBlockOpenMarker, NodeMathBlockCloseMarker, NodeInlineMathOpenMarker, NodeInlineMathCloseMarker, NodeYamlFrontMatterOpenMarker, NodeYamlFrontMatterCloseMarker,
		NodeMark1OpenMarker, NodeMark1CloseMarker, NodeMark2OpenMarker, NodeMark2CloseMarker, NodeTagOpenMarker, NodeTagCloseMarker,
		NodeSuperBlockOpenMarker, NodeSuperBlockLayoutMarker, NodeSuperBlockCloseMarker, NodeSupOpenMarker, NodeSupCloseMarker, NodeSubOpenMarker, NodeSubCloseMarker:
		return true
	}
	return false
}

func (n *Node) IsCloseMarker() bool {
	switch n.Type {
	case NodeHeadingC8hMarker, NodeBlockquoteMarker, NodeCodeBlockFenceCloseMarker, NodeEmA6kCloseMarker, NodeEmU8eCloseMarker,
		NodeStrongA6kCloseMarker, NodeStrongU8eCloseMarker, NodeCodeSpanCloseMarker, NodeStrikethrough1CloseMarker, NodeStrikethrough2CloseMarker,
		NodeMathBlockCloseMarker, NodeInlineMathCloseMarker, NodeYamlFrontMatterCloseMarker, NodeMark1CloseMarker, NodeMark2CloseMarker,
		NodeTagCloseMarker, NodeSuperBlockCloseMarker, NodeSupCloseMarker, NodeSubCloseMarker:
		return true
	}
	return false
}

func (n *Node) AcceptLines() bool {
	switch n.Type {
	case NodeParagraph, NodeCodeBlock, NodeHTMLBlock, NodeMathBlock, NodeYamlFrontMatter, NodeBlockQueryEmbed,
		NodeGitConflict, NodeIFrame, NodeWidget, NodeVideo, NodeAudio, NodeAttributeView, NodeCustomBlock:
		return true
	}
	return false
}

func (n *Node) CanContain(nodeType NodeType) bool {
	switch n.Type {
	case NodeCodeBlock, NodeHTMLBlock, NodeParagraph, NodeThematicBreak, NodeTable, NodeMathBlock, NodeYamlFrontMatter,
		NodeGitConflict, NodeIFrame, NodeWidget, NodeVideo, NodeAudio, NodeAttributeView, NodeCustomBlock:
		return false
	case NodeList:
		return NodeListItem == nodeType
	case NodeFootnotesDefBlock:
		return NodeFootnotesDef == nodeType
	case NodeFootnotesDef:
		return NodeFootnotesDef != nodeType
	case NodeSuperBlock:
		if nil != n.LastChild && NodeSuperBlockCloseMarker == n.LastChild.Type {
			return false
		}
		return true
	}
	return NodeListItem != nodeType
}

//go:generate stringer -type=NodeType
type NodeType int

var strNodeTypeMap = map[string]NodeType{}
var strNodeTypeMapLock = sync.RWMutex{}

func Str2NodeType(nodeTypeStr string) NodeType {
	strNodeTypeMapLock.RLock()
	defer strNodeTypeMapLock.RUnlock()
	if ret, ok := strNodeTypeMap[nodeTypeStr]; !ok {
		return -1
	} else {
		return ret
	}
}

const (
	// CommonMark

	NodeDocument                  NodeType = 0
	NodeParagraph                 NodeType = 1
	NodeHeading                   NodeType = 2
	NodeHeadingC8hMarker          NodeType = 3
	NodeThematicBreak             NodeType = 4
	NodeBlockquote                NodeType = 5
	NodeBlockquoteMarker          NodeType = 6
	NodeList                      NodeType = 7
	NodeListItem                  NodeType = 8
	NodeHTMLBlock                 NodeType = 9
	NodeInlineHTML                NodeType = 10
	NodeCodeBlock                 NodeType = 11
	NodeCodeBlockFenceOpenMarker  NodeType = 12
	NodeCodeBlockFenceCloseMarker NodeType = 13
	NodeCodeBlockFenceInfoMarker  NodeType = 14
	NodeCodeBlockCode             NodeType = 15
	NodeText                      NodeType = 16
	NodeEmphasis                  NodeType = 17
	NodeEmA6kOpenMarker           NodeType = 18
	NodeEmA6kCloseMarker          NodeType = 19
	NodeEmU8eOpenMarker           NodeType = 20
	NodeEmU8eCloseMarker          NodeType = 21
	NodeStrong                    NodeType = 22
	NodeStrongA6kOpenMarker       NodeType = 23
	NodeStrongA6kCloseMarker      NodeType = 24
	NodeStrongU8eOpenMarker       NodeType = 25
	NodeStrongU8eCloseMarker      NodeType = 26
	NodeCodeSpan                  NodeType = 27
	NodeCodeSpanOpenMarker        NodeType = 28
	NodeCodeSpanContent           NodeType = 29
	NodeCodeSpanCloseMarker       NodeType = 30
	NodeHardBreak                 NodeType = 31
	NodeSoftBreak                 NodeType = 32
	NodeLink                      NodeType = 33
	NodeImage                     NodeType = 34
	NodeBang                      NodeType = 35 // !
	NodeOpenBracket               NodeType = 36 // [
	NodeCloseBracket              NodeType = 37 // ]
	NodeOpenParen                 NodeType = 38 // (
	NodeCloseParen                NodeType = 39 // )
	NodeLinkText                  NodeType = 40
	NodeLinkDest                  NodeType = 41
	NodeLinkTitle                 NodeType = 42
	NodeLinkSpace                 NodeType = 43
	NodeHTMLEntity                NodeType = 44
	NodeLinkRefDefBlock           NodeType = 45
	NodeLinkRefDef                NodeType = 46
	NodeLess                      NodeType = 47 // <
	NodeGreater                   NodeType = 48 // >

	// GFM

	NodeTaskListItemMarker        NodeType = 100
	NodeStrikethrough             NodeType = 101
	NodeStrikethrough1OpenMarker  NodeType = 102
	NodeStrikethrough1CloseMarker NodeType = 103
	NodeStrikethrough2OpenMarker  NodeType = 104
	NodeStrikethrough2CloseMarker NodeType = 105
	NodeTable                     NodeType = 106
	NodeTableHead                 NodeType = 107
	NodeTableRow                  NodeType = 108
	NodeTableCell                 NodeType = 109

	// Emoji

	NodeEmoji        NodeType = 200 // Emoji
	NodeEmojiUnicode NodeType = 201 // Emoji Unicode
	NodeEmojiImg     NodeType = 202
	NodeEmojiAlias   NodeType = 203 // Emoji ASCII

	NodeMathBlock             NodeType = 300
	NodeMathBlockOpenMarker   NodeType = 301
	NodeMathBlockContent      NodeType = 302
	NodeMathBlockCloseMarker  NodeType = 303
	NodeInlineMath            NodeType = 304
	NodeInlineMathOpenMarker  NodeType = 305
	NodeInlineMathContent     NodeType = 306
	NodeInlineMathCloseMarker NodeType = 307

	NodeBackslash        NodeType = 400
	NodeBackslashContent NodeType = 401

	NodeVditorCaret NodeType = 405

	NodeFootnotesDefBlock NodeType = 410
	NodeFootnotesDef      NodeType = 411
	NodeFootnotesRef      NodeType = 412

	NodeToC NodeType = 415

	NodeHeadingID NodeType = 420

	// YAML Front Matter

	NodeYamlFrontMatter            NodeType = 425 // https://jekyllrb.com/docs/front-matter/
	NodeYamlFrontMatterOpenMarker  NodeType = 426
	NodeYamlFrontMatterContent     NodeType = 427
	NodeYamlFrontMatterCloseMarker NodeType = 428

	NodeBlockRef            NodeType = 430
	NodeBlockRefID          NodeType = 431
	NodeBlockRefSpace       NodeType = 432
	NodeBlockRefText        NodeType = 433
	NodeBlockRefDynamicText NodeType = 434

	NodeMark             NodeType = 450
	NodeMark1OpenMarker  NodeType = 451
	NodeMark1CloseMarker NodeType = 452
	NodeMark2OpenMarker  NodeType = 453
	NodeMark2CloseMarker NodeType = 454

	NodeKramdownBlockIAL NodeType = 455
	NodeKramdownSpanIAL  NodeType = 456

	NodeTag            NodeType = 460
	NodeTagOpenMarker  NodeType = 461
	NodeTagCloseMarker NodeType = 462

	NodeBlockQueryEmbed       NodeType = 465
	NodeOpenBrace             NodeType = 466 // {
	NodeCloseBrace            NodeType = 467 // }
	NodeBlockQueryEmbedScript NodeType = 468

	NodeSuperBlock             NodeType = 475
	NodeSuperBlockOpenMarker   NodeType = 476
	NodeSuperBlockLayoutMarker NodeType = 477
	NodeSuperBlockCloseMarker  NodeType = 478

	NodeSup            NodeType = 485
	NodeSupOpenMarker  NodeType = 486
	NodeSupCloseMarker NodeType = 487
	NodeSub            NodeType = 490
	NodeSubOpenMarker  NodeType = 491
	NodeSubCloseMarker NodeType = 492

	NodeGitConflict            NodeType = 495
	NodeGitConflictOpenMarker  NodeType = 496
	NodeGitConflictContent     NodeType = 497
	NodeGitConflictCloseMarker NodeType = 498

	NodeIFrame NodeType = 500

	NodeAudio NodeType = 505

	NodeVideo NodeType = 510

	NodeKbd            NodeType = 515
	NodeKbdOpenMarker  NodeType = 516
	NodeKbdCloseMarker NodeType = 517

	NodeUnderline            NodeType = 520
	NodeUnderlineOpenMarker  NodeType = 521
	NodeUnderlineCloseMarker NodeType = 522

	NodeBr NodeType = 525

	NodeTextMark NodeType = 530

	NodeWidget NodeType = 535 // <iframe data-type="NodeWidget" data-subtype="widget"></iframe>

	NodeFileAnnotationRef      NodeType = 540
	NodeFileAnnotationRefID    NodeType = 541
	NodeFileAnnotationRefSpace NodeType = 542
	NodeFileAnnotationRefText  NodeType = 543

	NodeAttributeView NodeType = 550

	NodeCustomBlock NodeType = 560

	NodeHTMLTag      NodeType = 570
	NodeHTMLTagOpen  NodeType = 571
	NodeHTMLTagClose NodeType = 572

	NodeCallout NodeType = 580

	NodeTypeMaxVal NodeType = 1024
)
