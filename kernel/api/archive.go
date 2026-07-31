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
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/icha-senpai/note/kernel/model"
	"github.com/icha-senpai/note/kernel/util"
	"github.com/icha-senpai/note/third_party/forks/gulu"
	"github.com/icha-senpai/note/third_party/forks/logging"
)

func rejectEncryptedArchivePath(absPath string) error {
	boxID := model.ExtractBoxIDFromAssetsPath(absPath)
	if boxID != "" && model.IsEncryptedBox(boxID) {
		return fmt.Errorf("path belongs to encrypted notebook [%s]", boxID)
	}
	if resolved := util.ResolveLongestExistingParent(absPath); resolved != absPath {
		boxID = model.ExtractBoxIDFromAssetsPath(resolved)
		if boxID != "" && model.IsEncryptedBox(boxID) {
			return fmt.Errorf("symlink resolves into encrypted notebook [%s]", boxID)
		}
	}
	return nil
}

func zip(c *gin.Context) {
	ret := gulu.Ret.NewResult()
	defer c.JSON(http.StatusOK, ret)

	arg, ok := util.JsonArg(c, ret)
	if !ok {
		return
	}

	var entryPath, zipFilePath string
	if !util.ParseJsonArgs(arg, ret,
		util.BindJsonArg("path", &entryPath, true, true),
		util.BindJsonArg("zipPath", &zipFilePath, true, true),
	) {
		return
	}
	entryAbsPath, err := util.GetAbsPathInWorkspace(entryPath)
	if err != nil {
		ret.Code = -1
		ret.Msg = err.Error()
		return
	}
	if err = rejectEncryptedArchivePath(entryAbsPath); err != nil {
		ret.Code = -1
		ret.Msg = err.Error()
		return
	}
	zipAbsFilePath, err := util.GetAbsPathInWorkspace(zipFilePath)
	if err != nil {
		ret.Code = -1
		ret.Msg = err.Error()
		return
	}
	if err = rejectEncryptedArchivePath(zipAbsFilePath); err != nil {
		ret.Code = -1
		ret.Msg = err.Error()
		return
	}

	zipFile, err := gulu.Zip.Create(zipAbsFilePath)
	if err != nil {
		logging.LogErrorf("create zip [%s] failed: %s", zipAbsFilePath, err)
		ret.Code = -1
		ret.Msg = "create zip file failed" + errMsgSeeKernelLog
		return
	}

	base := filepath.Base(entryAbsPath)
	if gulu.File.IsDir(entryAbsPath) {
		err = zipFile.AddDirectory(base, entryAbsPath)
	} else {
		err = zipFile.AddEntry(base, entryAbsPath)
	}
	if err != nil {
		logging.LogErrorf("zip add entry [%s] failed: %s", entryAbsPath, err)
		ret.Code = -1
		ret.Msg = "zip failed" + errMsgSeeKernelLog
		return
	}

	if err = zipFile.Close(); err != nil {
		logging.LogErrorf("close zip [%s] failed: %s", zipAbsFilePath, err)
		ret.Code = -1
		ret.Msg = "close zip file failed" + errMsgSeeKernelLog
		return
	}
}

func unzip(c *gin.Context) {
	ret := gulu.Ret.NewResult()
	defer c.JSON(http.StatusOK, ret)

	arg, ok := util.JsonArg(c, ret)
	if !ok {
		return
	}

	var zipFilePath, entryPath string
	if !util.ParseJsonArgs(arg, ret,
		util.BindJsonArg("zipPath", &zipFilePath, true, true),
		util.BindJsonArg("path", &entryPath, true, false),
	) {
		return
	}
	zipAbsFilePath, err := util.GetAbsPathInWorkspace(zipFilePath)
	if err != nil {
		ret.Code = -1
		ret.Msg = err.Error()
		return
	}
	if err = rejectEncryptedArchivePath(zipAbsFilePath); err != nil {
		ret.Code = -1
		ret.Msg = err.Error()
		return
	}
	entryAbsPath, err := util.GetAbsPathInWorkspace(entryPath)
	if err != nil {
		ret.Code = -1
		ret.Msg = err.Error()
		return
	}
	if err = rejectEncryptedArchivePath(entryAbsPath); err != nil {
		ret.Code = -1
		ret.Msg = err.Error()
		return
	}

	if !gulu.File.IsExist(zipAbsFilePath) {
		ret.Code = -1
		ret.Msg = "zip file does not exist"
		return
	}

	if err := gulu.Zip.Unzip(zipAbsFilePath, entryAbsPath); err != nil {
		logging.LogErrorf("unzip [%s] -> [%s] failed: %s", zipAbsFilePath, entryAbsPath, err)
		ret.Code = -1
		ret.Msg = "unzip failed" + errMsgSeeKernelLog
		return
	}
}
