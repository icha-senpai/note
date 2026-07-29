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

type BoxConf struct {
	Name                  string         `json:"name"`
	Sort                  int            `json:"sort"`
	Icon                  string         `json:"icon"`
	Closed                bool           `json:"closed"`
	RefCreateSaveBox      string         `json:"refCreateSaveBox"`
	RefCreateSavePath     string         `json:"refCreateSavePath"`
	DocCreateSaveBox      string         `json:"docCreateSaveBox"`
	DocCreateSavePath     string         `json:"docCreateSavePath"`
	DailyNoteSavePath     string         `json:"dailyNoteSavePath"`
	DailyNoteTemplatePath string         `json:"dailyNoteTemplatePath"`
	SortMode              int            `json:"sortMode"`
	Encrypted             bool           `json:"encrypted"`
	BoxCrypt              *BoxEncryption `json:"boxCrypt"`
}

type BoxEncryption struct {
	Spec       int    `json:"spec,omitempty"`
	WrappedDEK []byte `json:"wrappedDEK"`
	WrapNonce  []byte `json:"wrapNonce"`
	CreatedAt  int64  `json:"createdAt"`
}

func NewBoxConf() *BoxConf {
	return &BoxConf{
		Name:                  "Untitled",
		Closed:                true,
		DailyNoteSavePath:     "/daily note/{{now | date \"2006/01\"}}/{{now | date \"2006-01-02\"}}",
		DailyNoteTemplatePath: "",
		SortMode:              util.SortModeFileTree,
		Encrypted:             false,
	}
}
