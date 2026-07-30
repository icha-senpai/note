// Scribli - Refactor your thinking
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

package mobile

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/icha-senpai/note/third_party/forks/gulu"
	"github.com/icha-senpai/note/third_party/forks/lute/ast"
	"github.com/icha-senpai/note/kernel/cache"
	"github.com/icha-senpai/note/kernel/job"
	"github.com/icha-senpai/note/kernel/model"
	"github.com/icha-senpai/note/kernel/plugin"
	"github.com/icha-senpai/note/kernel/server"
	"github.com/icha-senpai/note/kernel/sql"
	"github.com/icha-senpai/note/kernel/util"
	"github.com/icha-senpai/note/third_party/forks/filelock"
	"github.com/icha-senpai/note/third_party/forks/logging"
	_ "github.com/icha-senpai/note/third_party/forks/external/golang.org/x/mobile/bind"
)

func StartKernelFast(container, appDir, workspaceBaseDir, localIPs string) {
	go server.Serve(true, model.Conf.CookieKey)
}

func StartKernel(container, appDir, workspaceBaseDir, timezoneID, localIPs, lang, osVer string) {
	SetTimezone(container, appDir, timezoneID)
	util.Mode = "prod"
	util.MobileOSVer = osVer
	util.LocalIPs = strings.Split(localIPs, ",")
	util.BootMobile(container, appDir, workspaceBaseDir, lang)

	model.InitConf()
	go server.Serve(false, model.Conf.CookieKey)
	go func() {
		model.InitAppearance()
		sql.InitDatabase(false)
		sql.InitHistoryDatabase(false)
		sql.InitAssetContentDatabase(false)
		sql.SetCaseSensitive(model.Conf.Search.CaseSensitive)
		sql.SetIndexAssetPath(model.Conf.Search.IndexAssetPath)

		model.BootSyncData()
		model.InitBoxes()
		model.LoadFlashcards()
		util.LoadAssetsTexts()

		util.SetBooted()
		util.PushClearAllMsg()

		job.StartCron()
		go model.AutoGenerateFileHistory()
		go cache.LoadAssets()
		go plugin.InitManager()
		go model.StartEmbeddingIndexer()
	}()
}

func Language(num int) string {
	return model.Conf.Language(num)
}

func ShowMsg(msg string, timeout int) {
	util.PushMsg(msg, timeout)
}

func IsHttpServing() bool {
	return util.HttpServing
}

func SetHttpServerPort(port int) {
	filelock.AndroidServerPort = port
}

func GetCurrentWorkspacePath() string {
	return util.WorkspaceDir
}

func GetAssetAbsPath(asset string) (ret string) {
	ret, err := model.GetAssetAbsPath(asset)
	if err != nil {
		logging.LogErrorf("get asset [%s] abs path failed: %s", asset, err)
		ret = asset
	}
	return
}

func GetMimeTypeByExt(ext string) string {
	return util.GetMimeTypeByExt(ext)
}

func SetTimezone(container, appDir, timezoneID string) {
	if "ios" == container {
		os.Setenv("ZONEINFO", filepath.Join(appDir, "app", "zoneinfo.zip"))
	}
	z, err := time.LoadLocation(strings.TrimSpace(timezoneID))
	if err != nil {
		fmt.Printf("load location failed: %s\n", err)
		time.Local = time.FixedZone("CST", 8*3600)
		return
	}
	time.Local = z
}

func DisableFeature(feature string) {
	util.DisableFeature(feature)
}

func FilepathBase(path string) string {
	return filepath.Base(path)
}

func FilterUploadFileName(name string) string {
	return util.FilterUploadFileName(name)
}

func AssetName(name string) string {
	return util.AssetName(name, ast.NewNodeID())
}

func HTML2Markdown(html string) string {
	return util.NewLute().HTML2Md(html)
}

func Unzip(zipFilePath, destination string) {
	if err := gulu.Zip.Unzip(zipFilePath, destination); nil != err {
		logging.LogErrorf("unzip [%s] failed: %s", zipFilePath, err)
		panic(err)
	}
}

func GetExportFilePath(exportPath string) (ret string) {
	var absPath string
	if after, ok := strings.CutPrefix(exportPath, "/export/"); ok {
		fileName := after
		if decoded, err := url.PathUnescape(fileName); err == nil {
			fileName = decoded
		}
		fileName = filepath.Clean(fileName)
		if strings.HasPrefix(fileName, "..") {
			logging.LogWarnf("get export file path [%s] blocked: path traversal attempt [%s]", exportPath, fileName)
			return
		}

		if model.IsManagedEncryptedExportPath(fileName) {
			artifact, ok := model.ResolveManagedExportForMobile(fileName)
			if !ok {
				logging.LogWarnf("get export file path [%s] blocked: managed export not available or box locked", exportPath)
				return
			}
			return artifact
		}
		absPath = filepath.Join(util.TempDir, "export", fileName)
		exportBaseDir := filepath.Join(util.TempDir, "export")
		if !gulu.File.IsSubPath(exportBaseDir, absPath) {
			logging.LogWarnf("get export file path [%s] blocked: path [%s] is outside export base dir [%s]", exportPath, absPath, exportBaseDir)
			return
		}
	} else if strings.HasPrefix(exportPath, "assets/") {
		var err error
		absPath, err = model.GetAssetAbsPath(exportPath)
		if nil != err {
			logging.LogErrorf("get asset abs path [%s] failed: %s", exportPath, err)
			return
		}
	} else {
		logging.LogWarnf("get export file path [%s] failed: unsupported path prefix", exportPath)
		return
	}

	if "" == absPath {
		logging.LogWarnf("get export file path [%s] failed: resolved to empty abs path", exportPath)
		return
	}
	return absPath
}

func Exit() {
	os.Exit(logging.ExitCodeOk)
}
