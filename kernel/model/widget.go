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
	"os"
	"path/filepath"
	"strings"

	"github.com/icha-senpai/note/kernel/extensions"
	"github.com/icha-senpai/note/kernel/util"
	"github.com/icha-senpai/note/third_party/forks/logging"
)

type WidgetSearchResult struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

func SearchWidget(keyword string) (ret []*WidgetSearchResult) {
	ret = []*WidgetSearchResult{}
	widgetsDirPath := filepath.Join(util.DataDir, "widgets")
	widgetsDir, err := os.ReadDir(widgetsDirPath)
	if err != nil {
		logging.LogErrorf("read dir [%s] failed: %s", widgetsDirPath, err)
		return
	}

	var widgets []*extensions.Package
	for _, dir := range widgetsDir {
		if !util.IsDirRegularOrSymlink(dir) {
			continue
		}
		dirName := dir.Name()
		if strings.HasPrefix(dirName, ".") {
			continue
		}

		widget, _ := extensions.ParsePackageJSON(filepath.Join(widgetsDirPath, dirName, "widget.json"))
		if nil == widget {
			continue
		}

		widgets = append(widgets, widget)
	}

	widgets = extensions.FilterPackages(widgets, keyword)
	for _, widget := range widgets {
		b := &WidgetSearchResult{
			Name:    extensions.GetPreferredLocaleString(widget.DisplayName, widget.Name),
			Content: widget.Name,
		}
		ret = append(ret, b)
	}

	return
}
