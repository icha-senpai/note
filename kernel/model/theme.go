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

package model

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"

	"github.com/icha-senpai/note/third_party/forks/gulu"
	"github.com/icha-senpai/note/third_party/forks/lute/ast"
	"github.com/icha-senpai/note/third_party/forks/lute/parse"
	"github.com/icha-senpai/note/third_party/forks/github/gorilla/css/scanner"
	"github.com/icha-senpai/note/kernel/util"
	"github.com/icha-senpai/note/third_party/forks/logging"
	"github.com/icha-senpai/note/third_party/forks/github/vanng822/css"
)

func fillThemeStyleVar(tree *parse.Tree) {
	if nil == tree || nil == tree.Root {
		return
	}

	var themeStyles map[string]string
	if 1 == Conf.Appearance.Mode {
		themeStyles = getThemeStyleVar(Conf.Appearance.ThemeDark, true)
	} else {
		themeStyles = getThemeStyleVar(Conf.Appearance.ThemeLight, false)
	}
	if 1 > len(themeStyles) {
		return
	}

	ast.Walk(tree.Root, func(n *ast.Node, entering bool) ast.WalkStatus {
		if !entering {
			return ast.WalkContinue
		}

		for _, ial := range n.KramdownIAL {
			if "style" != ial[0] {
				continue
			}

			styleSheet := css.Parse(ial[1])
			buf := bytes.Buffer{}
			for _, r := range styleSheet.GetCSSRuleList() {
				styles := getStyleVarName(r.Style.Selector)
				for style, name := range styles {
					buf.WriteString(style)
					buf.WriteString(": ")

					value := resolveNestedCSSVar(themeStyles, name)

					if "" == value {

						buf.WriteString("var(")
						buf.WriteString(name)
						buf.WriteString(")")
					} else {
						buf.WriteString(value)
					}
					buf.WriteString("; ")
				}
			}
			if 0 < buf.Len() {
				ial[1] = strings.TrimSpace(buf.String())
			}
		}
		return ast.WalkContinue
	})
}

func resolveNestedCSSVar(themeStyles map[string]string, varName string) string {
	visited := make(map[string]bool)
	maxDepth := 10

	currentName := varName
	for range maxDepth {
		if visited[currentName] {
			return ""
		}
		visited[currentName] = true

		value, exists := themeStyles[currentName]
		if !exists {
			return ""
		}

		if !strings.Contains(value, "var(") {
			return value
		}

		nestedVarName := gulu.Str.SubStringBetween(value, "(", ")")
		if "" == nestedVarName {
			return value
		}

		currentName = nestedVarName
	}

	return ""
}

func getStyleVarName(value *css.CSSValue) (ret map[string]string) {
	ret = map[string]string{}

	var start, end int
	var style, name string
	for i, t := range value.Tokens {

		if scanner.TokenIdent == t.Type && 0 == start {
			style = strings.TrimSpace(t.Value)
			continue
		}

		if scanner.TokenFunction == t.Type && "var(" == t.Value {
			start = i
			continue
		}
		if scanner.TokenChar == t.Type && ")" == t.Value {
			end = i

			if 0 < start && 0 < end {
				for _, tt := range value.Tokens[start+1 : end] {
					name += tt.Value
				}
				name = strings.TrimSpace(name)
			}
			start, end = 0, 0
			ret[style] = name
			style, name = "", ""
		}
	}
	return
}

func getThemeStyleVar(theme string, isDarkMode bool) (ret map[string]string) {
	ret = map[string]string{}

	var cssContent string

	defaultTheme := map[bool]string{false: "daylight", true: "midnight"}[isDarkMode]
	if theme != defaultTheme {
		defaultData, err := os.ReadFile(filepath.Join(util.ThemesPath, defaultTheme, "theme.css"))
		if err != nil {
			logging.LogErrorf("read default theme [%s] css file failed: %s", defaultTheme, err)
		} else {
			cssContent = string(defaultData) + "\n"
		}
	}

	userData, err := os.ReadFile(filepath.Join(util.ThemesPath, theme, "theme.css"))
	if err != nil {
		logging.LogErrorf("read theme [%s] css file failed: %s", theme, err)
		return ret
	}
	cssContent += string(userData)

	styleSheet := css.Parse(cssContent)
	stylePriorities := map[string]int{}
	currentMode := map[bool]string{false: "light", true: "dark"}[isDarkMode]
	for _, rule := range styleSheet.GetCSSRuleList() {
		priority := getSelectorPriority(rule.Style.Selector.Text(), currentMode)
		for _, style := range rule.Style.Styles {
			propName := style.Property
			propValue := strings.TrimSpace(style.Value.Text())

			if existingPriority, exists := stylePriorities[propName]; !exists || priority >= existingPriority {
				ret[propName] = propValue
				stylePriorities[propName] = priority
			}

			bugFixPropName := "-" + propName
			bugFixPropValue := strings.TrimSpace(strings.TrimPrefix(propValue, "-"))
			if existingPriority, exists := stylePriorities[bugFixPropName]; !exists || priority >= existingPriority {
				ret[bugFixPropName] = bugFixPropValue
				stylePriorities[bugFixPropName] = priority
			}
		}
	}
	return ret
}

func getSelectorPriority(selector, currentMode string) int {
	selector = strings.TrimSpace(strings.ToLower(selector))

	modeSelectors := []string{
		"[data-theme-mode=\"" + currentMode + "\"]",
		"[data-theme-mode='" + currentMode + "']",
		"[data-theme-mode=" + currentMode + "]",
	}

	for _, modeSelector := range modeSelectors {
		if strings.Contains(selector, modeSelector) {
			if strings.Contains(selector, ":root") || strings.Contains(selector, "html") {
				return 2
			}

			return 1
		}
	}

	return 0
}
