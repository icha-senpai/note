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

//go:build darwin

package model

import (
	"os"
	"path/filepath"
	"time"

	"github.com/icha-senpai/note/third_party/forks/gulu"
	"github.com/icha-senpai/note/kernel/cache"
	"github.com/icha-senpai/note/kernel/util"
	"github.com/icha-senpai/note/third_party/forks/logging"
	"github.com/icha-senpai/note/third_party/forks/github/radovskyb/watcher"
)

var assetsWatcher *watcher.Watcher

func WatchAssets() {
	if util.IsMobileContainer() {
		return
	}

	go watchAssets()
}

func watchAssets() {
	CloseWatchAssets()
	assetsDir := filepath.Join(util.DataDir, "assets")

	assetsWatcher = watcher.New()

	if !gulu.File.IsDir(assetsDir) {
		os.MkdirAll(assetsDir, 0755)
	}

	if err := assetsWatcher.Add(assetsDir); err != nil {
		logging.LogErrorf("add assets watcher for folder [%s] failed: %s", assetsDir, err)
		return
	}

	go func() {
		defer logging.Recover()

		for {
			select {
			case event, ok := <-assetsWatcher.Event:
				if !ok {
					return
				}

				if watcher.Write == event.Op {
					IncSync()
				}

				go cache.LoadAssets()

				if watcher.Remove == event.Op {
					HandleAssetsRemoveEvent(event.Path)
				} else {
					HandleAssetsChangeEvent(event.Path)
				}
			case err, ok := <-assetsWatcher.Error:
				if !ok {
					return
				}
				logging.LogErrorf("watch assets failed: %s", err)
			case <-assetsWatcher.Closed:
				return
			}
		}
	}()

	if err := assetsWatcher.Start(10 * time.Second); err != nil {
		logging.LogErrorf("start assets watcher for folder [%s] failed: %s", assetsDir, err)
		return
	}
}

func CloseWatchAssets() {
	if nil != assetsWatcher {
		assetsWatcher.Close()
		assetsWatcher = nil
	}
}
