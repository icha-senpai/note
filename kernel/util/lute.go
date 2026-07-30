// Scribli - Refactor your thinking
// Copyright (c) 2020-present, Scribli
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package util

import (
	"html"
	"strings"

	"github.com/icha-senpai/note/third_party/forks/lute"
	"github.com/icha-senpai/note/third_party/forks/github/PuerkitoBio/goquery"
	"github.com/icha-senpai/note/third_party/forks/logging"
)

var MarkdownSettings = &Markdown{
	InlineAsterisk:      true,
	InlineUnderscore:    true,
	InlineSup:           true,
	InlineSub:           true,
	InlineTag:           true,
	InlineMath:          true,
	InlineStrikethrough: true,
	InlineMark:          true,
}

type Markdown struct {
	InlineAsterisk      bool `json:"inlineAsterisk"`
	InlineUnderscore    bool `json:"inlineUnderscore"`
	InlineSup           bool `json:"inlineSup"`
	InlineSub           bool `json:"inlineSub"`
	InlineTag           bool `json:"inlineTag"`
	InlineMath          bool `json:"inlineMath"`
	InlineStrikethrough bool `json:"inlineStrikethrough"`
	InlineMark          bool `json:"inlineMark"`
}

func NewLute() (ret *lute.Lute) {
	ret = lute.New()
	ret.SetTextMark(true)
	ret.SetProtyleWYSIWYG(true)
	ret.SetBlockRef(true)
	ret.SetFileAnnotationRef(true)
	ret.SetKramdownIAL(true)
	ret.SetTag(true)
	ret.SetSuperBlock(true)
	ret.SetImgPathAllowSpace(true)
	ret.SetGitConflict(true)
	ret.SetInlineAsterisk(MarkdownSettings.InlineAsterisk)
	ret.SetInlineUnderscore(MarkdownSettings.InlineUnderscore)
	ret.SetSup(MarkdownSettings.InlineSup)
	ret.SetSub(MarkdownSettings.InlineSub)
	ret.SetTag(MarkdownSettings.InlineTag)
	ret.SetInlineMath(MarkdownSettings.InlineMath)
	ret.SetGFMStrikethrough(MarkdownSettings.InlineStrikethrough)
	ret.SetMark(MarkdownSettings.InlineMark)
	ret.SetInlineMathAllowDigitAfterOpenMarker(true)
	ret.SetGFMStrikethrough1(false)
	ret.SetFootnotes(false)
	ret.SetToC(false)
	ret.SetIndentCodeBlock(false)
	ret.SetParagraphBeginningSpace(true)
	ret.SetAutoSpace(false)
	ret.SetHeadingID(false)
	ret.SetSetext(false)
	ret.SetYamlFrontMatter(false)
	ret.SetLinkRef(false)
	ret.SetCodeSyntaxHighlight(false)
	ret.SetSanitize(true)
	ret.SetUnorderedListMarker("-")
	ret.SetCallout(true)
	ret.SetDataTask(true)
	ret.SetArbitraryTaskListItemMarker(true)
	ret.SetExportNormalizeTaskListMarker(false)
	ret.SetEnsureListItemParagraph(true)
	return
}

func NewStdLute() (ret *lute.Lute) {
	ret = lute.New()
	ret.SetFootnotes(false)
	ret.SetToC(false)
	ret.SetIndentCodeBlock(true)
	ret.SetAutoSpace(false)
	ret.SetHeadingID(false)
	ret.SetSetext(false)
	ret.SetYamlFrontMatter(false)
	ret.SetLinkRef(false)
	ret.SetGFMAutoLink(false)
	ret.SetImgPathAllowSpace(true)
	ret.SetInlineMathAllowDigitAfterOpenMarker(true) // Formula parsing supports $ followed by numbers when importing Markdown

	// Follow editor Markdown syntax settings when importing Markdown
	ret.SetInlineAsterisk(MarkdownSettings.InlineAsterisk)
	ret.SetInlineUnderscore(MarkdownSettings.InlineUnderscore)
	ret.SetSup(MarkdownSettings.InlineSup)
	ret.SetSub(MarkdownSettings.InlineSub)
	ret.SetTag(MarkdownSettings.InlineTag)
	ret.SetInlineMath(MarkdownSettings.InlineMath)
	ret.SetGFMStrikethrough(MarkdownSettings.InlineStrikethrough)
	ret.SetGFMStrikethrough1(false)
	return
}

func ConvertIframeToLink(htmlStr string) string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlStr))
	if err != nil {
		logging.LogErrorf("parse HTML for iframe conversion failed: %s", err)
		return htmlStr
	}

	doc.Find("iframe").Each(func(i int, s *goquery.Selection) {
		if src, exists := s.Attr("src"); exists && strings.TrimSpace(src) != "" {
			escapedSrc := html.EscapeString(src)
			s.AfterHtml(`<a href="` + escapedSrc + `" target="_blank">` + escapedSrc + `</a>`)
		}
		s.Remove()
	})

	ret, _ := doc.Find("body").Html()
	return ret
}

func LinkTarget(htmlStr, linkBase string) (ret string) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlStr))
	if err != nil {
		logging.LogErrorf("parse HTML failed: %s", err)
		return
	}

	doc.Find("a").Each(func(i int, selection *goquery.Selection) {
		if href, ok := selection.Attr("href"); ok {
			if IsRelativePath(href) {
				selection.SetAttr("href", linkBase+href)
			}

			// The hyperlink in the extension package README fails to jump to the browser to open
			selection.SetAttr("target", "_blank")
		}
	})

	ret, _ = doc.Find("body").Html()
	return
}
