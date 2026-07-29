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

type Appearance struct {
	Mode                int                 `json:"mode"`
	ModeOS              bool                `json:"modeOS"`
	DarkThemes          []*AppearanceTheme  `json:"darkThemes"`
	LightThemes         []*AppearanceTheme  `json:"lightThemes"`
	ThemeDark           string              `json:"themeDark"`
	ThemeLight          string              `json:"themeLight"`
	ThemeVer            string              `json:"themeVer"`
	Icons               []*AppearanceIcon   `json:"icons"`
	Icon                string              `json:"icon"`
	IconVer             string              `json:"iconVer"`
	CodeBlockThemeLight string              `json:"codeBlockThemeLight"`
	CodeBlockThemeDark  string              `json:"codeBlockThemeDark"`
	Lang                string              `json:"lang"`
	ThemeJS             bool                `json:"themeJS"`
	CloseButtonBehavior int                 `json:"closeButtonBehavior"`
	HideToolbar         bool                `json:"hideToolbar"`
	HideStatusBar       bool                `json:"hideStatusBar"`
	StatusBar           *util.StatusBar     `json:"statusBar"`
	Notifications       *util.Notifications `json:"notifications"`
}

func NewAppearance() *Appearance {
	return &Appearance{
		Mode:                0,
		ModeOS:              true,
		ThemeDark:           "midnight",
		ThemeLight:          "daylight",
		Icon:                "litheness",
		CodeBlockThemeLight: "github",
		CodeBlockThemeDark:  "base16/dracula",
		Lang:                "en",
		CloseButtonBehavior: 0,
		HideToolbar:         true,
		HideStatusBar:       false,
		StatusBar:           &util.StatusBar{},
		Notifications:       util.NewNotifications(),
	}
}

type AppearanceTheme struct {
	Name  string `json:"name"`  // daylight
	Label string `json:"label"` // i18n display name
}

type AppearanceIcon struct {
	Name  string `json:"name"`  // litheness
	Label string `json:"label"` // i18n display name
}
