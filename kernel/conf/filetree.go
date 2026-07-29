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

import (
	"github.com/siyuan-note/siyuan/kernel/util"
)

type FileTree struct {
	AlwaysSelectOpenedFile   bool   `json:"alwaysSelectOpenedFile"`
	OpenFilesUseCurrentTab   bool   `json:"openFilesUseCurrentTab"`
	DocIconClickExpand       bool   `json:"docIconClickExpand"`
	ParentDocClickExpand     bool   `json:"parentDocClickExpand"`
	BoxDocEnabled            *bool  `json:"boxDocEnabled"`
	RefCreateSaveBox         string `json:"refCreateSaveBox"`
	RefCreateSavePath        string `json:"refCreateSavePath"`
	DocCreateSaveBox         string `json:"docCreateSaveBox"`
	DocCreateSavePath        string `json:"docCreateSavePath"`
	ShorthandSaveBox         string `json:"shorthandSaveBox"`
	ShorthandSavePath        string `json:"shorthandSavePath"`
	MaxListCount             int    `json:"maxListCount"`
	MaxOpenTabCount          int    `json:"maxOpenTabCount"`
	AllowCreateDeeper        bool   `json:"allowCreateDeeper"`
	RemoveDocWithoutConfirm  bool   `json:"removeDocWithoutConfirm"`
	CloseTabsOnStart         bool   `json:"closeTabsOnStart"`
	UseSingleLineSave        bool   `json:"useSingleLineSave"`
	LargeFileWarningSize     int    `json:"largeFileWarningSize"`
	CreateDocAtTop           *bool  `json:"createDocAtTop"`
	Sort                     int    `json:"sort"`
	RecentDocsMaxListCount   int    `json:"recentDocsMaxListCount"`
	NoSplitScreenWhenOpenTab bool   `json:"noSplitScreenWhenOpenTab"`
}

func NewFileTree() *FileTree {
	return &FileTree{
		AlwaysSelectOpenedFile:   false,
		OpenFilesUseCurrentTab:   false,
		DocIconClickExpand:       false,
		ParentDocClickExpand:     false,
		BoxDocEnabled:            func() *bool { b := true; return &b }(),
		Sort:                     util.SortModeCustom,
		MaxListCount:             512,
		MaxOpenTabCount:          8,
		AllowCreateDeeper:        false,
		CloseTabsOnStart:         false,
		UseSingleLineSave:        util.UseSingleLineSave,
		LargeFileWarningSize:     util.LargeFileWarningSize,
		CreateDocAtTop:           func() *bool { b := false; return &b }(),
		NoSplitScreenWhenOpenTab: false,
	}
}

const (
	MinFileTreeRecentDocsListCount = 32
	MaxFileTreeRecentDocsListCount = 256
)
