package model

import (
	"path/filepath"
	"strings"
	"time"

	"github.com/icha-senpai/note/kernel/cache"
	"github.com/icha-senpai/note/kernel/sql"
	"github.com/icha-senpai/note/kernel/task"
	"github.com/icha-senpai/note/kernel/util"
	"github.com/icha-senpai/note/third_party/forks/logging"
)

func OCRAssetsJob() {
	util.WaitForTesseractInit()

	if !util.TesseractEnabled {
		return
	}

	task.AppendTaskWithTimeout(task.OCRImage, 30*time.Second, autoOCRAssets)
}

func autoOCRAssets() {
	if !util.TesseractEnabled {
		return
	}

	defer logging.Recover()

	assetsPath := util.GetDataAssetsAbsPath()
	assets := getUnOCRAssetsAbsPaths()
	if 0 < len(assets) {
		for i, assetAbsPath := range assets {
			text := util.GetOcrJsonText(util.Tesseract(assetAbsPath))
			p := strings.TrimPrefix(assetAbsPath, assetsPath)
			p = "assets" + filepath.ToSlash(p)
			util.SetAssetText(p, text)
			if 7 <= i {
				break
			}
		}
	}

	util.CleanNotExistAssetsTexts()

	util.NodeOCRQueueLock.Lock()
	defer util.NodeOCRQueueLock.Unlock()
	for _, id := range util.NodeOCRQueue {
		sql.IndexNodeQueue(id)
	}
	util.NodeOCRQueue = nil
}

func getUnOCRAssetsAbsPaths() (ret []string) {

	ocrAssets := cache.FilterAssets(func(path string, asset *cache.Asset) bool {
		return util.IsTesseractExtractable(asset.Path)
	})

	assetsPath := util.GetDataAssetsAbsPath()
	for _, asset := range ocrAssets {

		if util.ExistsAssetText(asset.Path) {
			continue
		}
		absPath := filepath.Join(assetsPath, strings.TrimPrefix(asset.Path, "assets"))

		if IsEncryptedAssetPath(absPath) {
			continue
		}
		ret = append(ret, absPath)
	}
	return
}

func FlushAssetsTextsJob() {
	util.SaveAssetsTexts()
}
