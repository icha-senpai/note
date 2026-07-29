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

package extensions

import (
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/88250/gulu"
	"github.com/siyuan-note/filelock"
	"github.com/siyuan-note/logging"
	"github.com/siyuan-note/siyuan/kernel/util"
	"golang.org/x/mod/semver"
	"golang.org/x/sync/singleflight"
)

func isBelowRequiredAppVersion(pkg *Package) bool {

	if "" == pkg.MinAppVersion {
		return false
	}

	if 0 < semver.Compare("v"+pkg.MinAppVersion, "v"+util.Ver) {
		return true
	}
	return false
}

type ExtensionsInfo struct {
	Packages map[string]map[string]*PackageInfo `json:"packages"`
}

type PackageInfo struct {
	InstallTime int64 `json:"installTime"`
}

var (
	extensionsInfoCache        *ExtensionsInfo
	extensionsInfoModTime      time.Time
	extensionsInfoCacheLock    = sync.RWMutex{}
	extensionsInfoSingleFlight singleflight.Group
)

func getExtensionsInfo() {
	infoPath := filepath.Join(util.DataDir, "storage", "extensions.json")
	info, err := os.Stat(infoPath)

	extensionsInfoCacheLock.RLock()
	cache := extensionsInfoCache
	modTime := extensionsInfoModTime
	extensionsInfoCacheLock.RUnlock()

	if cache != nil && err == nil && info.ModTime().Equal(modTime) {
		return
	}

	_, _, _ = extensionsInfoSingleFlight.Do("loadExtensionsInfo", func() (any, error) {

		newRet := loadExtensionsInfo()

		extensionsInfoCacheLock.Lock()
		extensionsInfoCache = newRet
		if err == nil {
			extensionsInfoModTime = info.ModTime()
		}
		extensionsInfoCacheLock.Unlock()
		return newRet, nil
	})
}

func loadExtensionsInfo() (ret *ExtensionsInfo) {

	ret = &ExtensionsInfo{
		Packages: make(map[string]map[string]*PackageInfo),
	}

	infoDir := filepath.Join(util.DataDir, "storage")
	if err := os.MkdirAll(infoDir, 0755); err != nil {
		logging.LogErrorf("create extensions info dir [%s] failed: %s", infoDir, err)
		return
	}

	infoPath := filepath.Join(infoDir, "extensions.json")
	if !filelock.IsExist(infoPath) {
		return
	}

	data, err := filelock.ReadFile(infoPath)
	if err != nil {
		logging.LogErrorf("read extensions info [%s] failed: %s", infoPath, err)
		return
	}

	if err = gulu.JSON.UnmarshalJSON(data, &ret); err != nil {
		logging.LogErrorf("unmarshal extensions info [%s] failed: %s", infoPath, err)
		ret = &ExtensionsInfo{
			Packages: make(map[string]map[string]*PackageInfo),
		}
	}

	return
}

func saveExtensionsInfo() {
	infoPath := filepath.Join(util.DataDir, "storage", "extensions.json")

	data, err := gulu.JSON.MarshalIndentJSON(extensionsInfoCache, "", "\t")
	if err != nil {
		logging.LogErrorf("marshal extensions info [%s] failed: %s", infoPath, err)
		return
	}
	if err = filelock.WriteFile(infoPath, data); err != nil {
		logging.LogErrorf("write extensions info [%s] failed: %s", infoPath, err)
		return
	}

	if fi, statErr := os.Stat(infoPath); statErr == nil {
		extensionsInfoModTime = fi.ModTime()
	}
}

func setPackageInstallTime(pkgType, pkgName string, installTime time.Time) {
	getExtensionsInfo()

	extensionsInfoCacheLock.Lock()
	defer extensionsInfoCacheLock.Unlock()

	if extensionsInfoCache == nil {
		return
	}
	if extensionsInfoCache.Packages[pkgType] == nil {
		extensionsInfoCache.Packages[pkgType] = make(map[string]*PackageInfo)
	}
	p := extensionsInfoCache.Packages[pkgType][pkgName]
	if p == nil {
		p = &PackageInfo{}
		extensionsInfoCache.Packages[pkgType][pkgName] = p
	}
	p.InstallTime = installTime.UnixMilli()
	saveExtensionsInfo()
}

func getPackageHInstallDate(pkgType, pkgName, installPath string) string {
	getExtensionsInfo()
	extensionsInfoCacheLock.RLock()
	var installTime int64
	if extensionsInfoCache != nil && extensionsInfoCache.Packages[pkgType] != nil {
		if p := extensionsInfoCache.Packages[pkgType][pkgName]; p != nil {
			installTime = p.InstallTime
		}
	}
	extensionsInfoCacheLock.RUnlock()

	if installTime > 0 {
		return time.UnixMilli(installTime).Format("2006-01-02")
	}

	fi, err := os.Stat(installPath)
	if err != nil {
		logging.LogWarnf("stat install package folder [%s] failed: %s", installPath, err)
		return time.Now().Format("2006-01-02")
	}
	setPackageInstallTime(pkgType, pkgName, fi.ModTime())

	return fi.ModTime().Format("2006-01-02")
}

func RemovePackageInfo(pkgType, pkgName string) {
	getExtensionsInfo()

	extensionsInfoCacheLock.Lock()
	defer extensionsInfoCacheLock.Unlock()

	if extensionsInfoCache != nil && extensionsInfoCache.Packages[pkgType] != nil {
		delete(extensionsInfoCache.Packages[pkgType], pkgName)
	}

	saveExtensionsInfo()
}
