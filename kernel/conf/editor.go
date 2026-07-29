// SiYuan - Refactor your thinking
// Copyright (c) 2020-present, b3log.org
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

package conf

import "github.com/siyuan-note/siyuan/kernel/util"

type Editor struct {
	AllowSVGScript                  bool           `json:"allowSVGScript"`
	AllowHTMLBLockScript            bool           `json:"allowHTMLBLockScript"`
	FontSize                        int            `json:"fontSize"`
	FontSizeScrollZoom              bool           `json:"fontSizeScrollZoom"`
	FontFamily                      string         `json:"fontFamily"`
	FontWeight                      int            `json:"fontWeight"`
	FontFamilyDisplay               string         `json:"fontFamilyDisplay"`
	CodeSyntaxHighlightLineNum      bool           `json:"codeSyntaxHighlightLineNum"`
	CodeTabSpaces                   int            `json:"codeTabSpaces"`
	CodeLineWrap                    bool           `json:"codeLineWrap"`
	CodeLigatures                   bool           `json:"codeLigatures"`
	DisplayBookmarkIcon             bool           `json:"displayBookmarkIcon"`
	DisplayNetImgMark               bool           `json:"displayNetImgMark"`
	DatabaseAttrViewMode            int            `json:"databaseAttrViewMode"`
	GenerateHistoryInterval         int            `json:"generateHistoryInterval"`
	HistoryRetentionDays            int            `json:"historyRetentionDays"`
	Emoji                           []string       `json:"emoji"`
	VirtualBlockRef                 bool           `json:"virtualBlockRef"`
	VirtualBlockRefExclude          string         `json:"virtualBlockRefExclude"`
	VirtualBlockRefInclude          string         `json:"virtualBlockRefInclude"`
	BlockRefDynamicAnchorTextMaxLen int            `json:"blockRefDynamicAnchorTextMaxLen"`
	PlantUMLServePath               string         `json:"plantUMLServePath"`
	FullWidth                       bool           `json:"fullWidth"`
	KaTexMacros                     string         `json:"katexMacros"`
	ReadOnly                        bool           `json:"readOnly"`
	EmbedBlockBreadcrumb            bool           `json:"embedBlockBreadcrumb"`
	ListLogicalOutdent              bool           `json:"listLogicalOutdent"`
	ListItemDotNumberClickFocus     bool           `json:"listItemDotNumberClickFocus"`
	FloatWindowMode                 int            `json:"floatWindowMode"`
	FloatWindowDelay                *int           `json:"floatWindowDelay"`
	DynamicLoadBlocks               int            `json:"dynamicLoadBlocks"`
	Justify                         bool           `json:"justify"`
	RTL                             bool           `json:"rtl"`
	Spellcheck                      bool           `json:"spellcheck"`
	SpellcheckLanguages             []string       `json:"spellcheckLanguages"`
	OnlySearchForDoc                bool           `json:"onlySearchForDoc"`
	BacklinkExpandCount             int            `json:"backlinkExpandCount"`
	BackmentionExpandCount          int            `json:"backmentionExpandCount"`
	BacklinkContainChildren         bool           `json:"backlinkContainChildren"`
	BacklinkSort                    *int           `json:"backlinkSort"`
	BackmentionSort                 *int           `json:"backmentionSort"`
	HeadingEmbedMode                int            `json:"headingEmbedMode"`
	PasteURLAutoConvert             bool           `json:"pasteURLAutoConvert"`
	Markdown                        *util.Markdown `json:"markdown"`
}

const (
	MinDynamicLoadBlocks = 48
)

func NewEditor() *Editor {
	return &Editor{
		FontSize:                        16,
		FontSizeScrollZoom:              false,
		CodeSyntaxHighlightLineNum:      false,
		CodeTabSpaces:                   0,
		CodeLineWrap:                    false,
		CodeLigatures:                   false,
		DisplayBookmarkIcon:             true,
		DisplayNetImgMark:               true,
		DatabaseAttrViewMode:            0,
		GenerateHistoryInterval:         10,
		HistoryRetentionDays:            30,
		Emoji:                           []string{},
		VirtualBlockRef:                 false,
		BlockRefDynamicAnchorTextMaxLen: 96,
		PlantUMLServePath:               "https://www.plantuml.com/plantuml/svg/~1",
		FullWidth:                       true,
		KaTexMacros:                     "{}",
		ReadOnly:                        false,
		EmbedBlockBreadcrumb:            false,
		ListLogicalOutdent:              false,
		ListItemDotNumberClickFocus:     true,
		FloatWindowMode:                 0,
		FloatWindowDelay:                func() *int { v := 620; return &v }(),
		DynamicLoadBlocks:               192,
		Justify:                         false,
		RTL:                             false,
		Spellcheck:                      false,
		SpellcheckLanguages:             []string{"en-US"},
		BacklinkExpandCount:             8,
		BackmentionExpandCount:          -1,
		BacklinkContainChildren:         true,
		BacklinkSort:                    func() *int { v := util.SortModeUpdatedDESC; return &v }(),
		BackmentionSort:                 func() *int { v := util.SortModeUpdatedDESC; return &v }(),
		HeadingEmbedMode:                0,
		PasteURLAutoConvert:             false,
		Markdown:                        util.MarkdownSettings,
	}
}
