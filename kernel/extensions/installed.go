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
	// 如果包没有指定 minAppVersion，则允许安装
	if "" == pkg.MinAppVersion {
		return false
	}

	// 如果包要求的 minAppVersion 大于当前版本，则不允许安装
	if 0 < semver.Compare("v"+pkg.MinAppVersion, "v"+util.Ver) {
		return true
	}
	return false
}

// ExtensionsInfo 本地扩展包的持久化信息
type ExtensionsInfo struct {
	Packages map[string]map[string]*PackageInfo `json:"packages"`
}

// PackageInfo 本地扩展包的持久化信息
type PackageInfo struct {
	InstallTime int64 `json:"installTime"` // 安装时间戳（毫秒）
}

var (
	extensionsInfoCache        *ExtensionsInfo
	extensionsInfoModTime      time.Time
	extensionsInfoCacheLock    = sync.RWMutex{}
	extensionsInfoSingleFlight singleflight.Group
)

// getExtensionsInfo 确保本地扩展包持久化信息已加载到内存缓存
func getExtensionsInfo() {
	infoPath := filepath.Join(util.DataDir, "storage", "extensions.json")
	info, err := os.Stat(infoPath)

	extensionsInfoCacheLock.RLock()
	cache := extensionsInfoCache
	modTime := extensionsInfoModTime
	extensionsInfoCacheLock.RUnlock()
	// 文件修改时间没变则认为缓存有效
	if cache != nil && err == nil && info.ModTime().Equal(modTime) {
		return
	}

	_, _, _ = extensionsInfoSingleFlight.Do("loadExtensionsInfo", func() (any, error) {
		// 缓存失效时从磁盘加载
		newRet := loadExtensionsInfo()
		// 更新缓存和修改时间
		extensionsInfoCacheLock.Lock()
		extensionsInfoCache = newRet
		if err == nil {
			extensionsInfoModTime = info.ModTime()
		}
		extensionsInfoCacheLock.Unlock()
		return newRet, nil
	})
}

// loadExtensionsInfo 从磁盘加载本地扩展包持久化信息
func loadExtensionsInfo() (ret *ExtensionsInfo) {
	// 初始化一个空的 ExtensionsInfo，后续使用时无需判断 nil
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

// saveExtensionsInfo 保存本地扩展包持久化信息（调用者需持有 extensionsInfoCacheLock 写锁）
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

// setPackageInstallTime 设置本地扩展包的安装时间
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

// getPackageHInstallDate 获取本地扩展包的安装日期
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

	// 如果 extensions.json 中没有记录，使用文件夹修改时间并记录到 extensions.json 中
	fi, err := os.Stat(installPath)
	if err != nil {
		logging.LogWarnf("stat install package folder [%s] failed: %s", installPath, err)
		return time.Now().Format("2006-01-02")
	}
	setPackageInstallTime(pkgType, pkgName, fi.ModTime())

	return fi.ModTime().Format("2006-01-02")
}

// RemovePackageInfo 删除本地扩展包的持久化信息
func RemovePackageInfo(pkgType, pkgName string) {
	getExtensionsInfo()

	extensionsInfoCacheLock.Lock()
	defer extensionsInfoCacheLock.Unlock()

	if extensionsInfoCache != nil && extensionsInfoCache.Packages[pkgType] != nil {
		delete(extensionsInfoCache.Packages[pkgType], pkgName)
	}

	saveExtensionsInfo()
}
