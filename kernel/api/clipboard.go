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

package api

import (
	"net/http"
	"os"

	"github.com/icha-senpai/note/third_party/forks/clipboard"
	"github.com/icha-senpai/note/third_party/forks/gulu"
	"github.com/icha-senpai/note/third_party/forks/github/gin-gonic/gin"
	"github.com/icha-senpai/note/kernel/model"
	"github.com/icha-senpai/note/kernel/util"
	"github.com/icha-senpai/note/third_party/forks/logging"
)

func readFilePaths(c *gin.Context) {
	ret := gulu.Ret.NewResult()
	defer c.JSON(http.StatusOK, ret)

	var paths []string
	if !gulu.OS.IsLinux() {
		paths, _ = clipboard.ReadFilePaths()
	}

	data := []map[string]any{}
	for _, path := range paths {
		fi, err := os.Stat(path)
		if nil != err {
			logging.LogErrorf("stat file failed: %s", err)
			continue
		}

		data = append(data, map[string]any{
			"name":    fi.Name(),
			"size":    fi.Size(),
			"isDir":   fi.IsDir(),
			"updated": fi.ModTime().UnixMilli(),
			"path":    path,
		})
	}
	ret.Data = data
}

func writeFilePath(c *gin.Context) {
	ret := gulu.Ret.NewResult()
	defer c.JSON(http.StatusOK, ret)

	arg, ok := util.JsonArg(c, ret)
	if !ok {
		return
	}

	var pathArg string
	if !util.ParseJsonArgs(arg, ret, util.BindJsonArg("path", &pathArg, true, true)) {
		return
	}

	absPath, err := model.GetAssetAbsPathInBox(pathArg, "")
	if err != nil {
		logging.LogErrorf("get asset [%s] abs path failed: %s", pathArg, err)
		ret.Code = -1
		ret.Msg = err.Error()
		ret.Data = map[string]any{"closeTimeout": 5000}
		return
	}
	if model.IsEncryptedAssetPath(absPath) {
		ret.Code = -1
		ret.Msg = model.Conf.Language(314)
		ret.Data = map[string]any{"closeTimeout": 5000}
		return
	}

	if err = util.WriteFilePaths([]string{absPath}); err != nil {
		logging.LogErrorf("write file path to clipboard failed: %s", err)
		ret.Code = -1
		ret.Msg = err.Error()
		ret.Data = map[string]any{"closeTimeout": 5000}
		return
	}
}
