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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/icha-senpai/note/kernel/cache"
	"github.com/icha-senpai/note/kernel/treenode"
	"github.com/icha-senpai/note/kernel/util"
	"github.com/icha-senpai/note/third_party/forks/filelock"
	"github.com/icha-senpai/note/third_party/forks/gulu"
	"github.com/icha-senpai/note/third_party/forks/logging"
	"github.com/icha-senpai/note/third_party/forks/lute/ast"
)

func InsertAssetBytes(id, fileName string, data []byte) (assetPath string, created bool, err error) {
	bt := treenode.GetBlockTree(id)
	if bt == nil {
		return "", false, errors.New(Conf.Language(71))
	}
	if len(data) == 0 {
		return "", false, errors.New("asset data is empty")
	}

	baseName := filepath.Base(fileName)
	fName := util.FilterUploadFileName(baseName)
	ext := strings.ToLower(filepath.Ext(fName))
	fName = strings.TrimSuffix(fName, filepath.Ext(fName)) + ext
	if fName == "" || fName == "." || ext == "" {
		return "", false, errors.New("invalid asset filename")
	}

	docDirLocalPath := filepath.Join(util.DataDir, bt.BoxID, path.Dir(bt.Path))
	assetsDirPath := getAssetsDir(filepath.Join(util.DataDir, bt.BoxID), docDirLocalPath)
	if err = os.MkdirAll(assetsDirPath, 0755); err != nil {
		return "", false, err
	}

	reader := bytes.NewReader(data)
	hash, err := util.GetEtagByHandle(reader, int64(len(data)))
	if err != nil {
		return "", false, err
	}
	if existAssetPath := GetAssetPathByHash(hash, bt.BoxID); existAssetPath != "" {
		originalName := util.RemoveID(filepath.Base(existAssetPath))
		if strings.EqualFold(fName, originalName) {
			return strings.TrimPrefix(existAssetPath, "/"), false, nil
		}
		hash = "random_2_" + gulu.Rand.String(12)
	}

	blockID := ast.NewNodeID()
	if IsEncryptedBox(bt.BoxID) {
		fName = encryptedAssetName(util.Ext(fName), blockID)
		if err = writeAssetNameMapping(bt.BoxID, fName, baseName); err != nil {
			return "", false, err
		}
	} else {
		fName = util.AssetName(fName, blockID)
	}
	writePath := filepath.Join(assetsDirPath, fName)
	if err = writeAssetFile(writePath, bytes.NewReader(data), bt.BoxID); err != nil {
		if IsEncryptedBox(bt.BoxID) {
			_ = removeAssetNameMapping(bt.BoxID, fName)
		}
		return "", false, err
	}

	assetPath = "assets/" + fName
	if IsEncryptedBox(bt.BoxID) {
		assetPath += "?box=" + bt.BoxID
	} else {
		cache.SetAssetHash(hash, assetPath)
	}
	IncSync()
	return assetPath, true, nil
}

func InsertLocalAssets(id string, assetAbsPaths []string, isUpload bool) (succMap map[string]any, err error) {
	succMap = map[string]any{}

	bt := treenode.GetBlockTree(id)
	if nil == bt {
		err = errors.New(Conf.Language(71))
		return
	}

	docDirLocalPath := filepath.Join(util.DataDir, bt.BoxID, path.Dir(bt.Path))
	assetsDirPath := getAssetsDir(filepath.Join(util.DataDir, bt.BoxID), docDirLocalPath)
	if !gulu.File.IsExist(assetsDirPath) {
		if err = os.MkdirAll(assetsDirPath, 0755); err != nil {
			return
		}
	}

	for _, assetAbsPath := range assetAbsPaths {
		baseName := filepath.Base(assetAbsPath)
		fName := baseName
		fName = util.FilterUploadFileName(fName)
		ext := filepath.Ext(fName)
		fName = strings.TrimSuffix(fName, ext)
		ext = strings.ToLower(ext)
		fName += ext
		if gulu.File.IsDir(assetAbsPath) || !isUpload {
			if !strings.HasPrefix(assetAbsPath, "\\\\") {
				assetAbsPath = "file://" + assetAbsPath
			}
			succMap[baseName] = assetAbsPath
			continue
		}

		if gulu.File.IsSubPath(assetsDirPath, assetAbsPath) {

			// Dragging a file from the assets folder into the editor causes the kernel to exit
			succMap[baseName] = "assets/" + baseName
			continue
		}

		fi, statErr := os.Stat(assetAbsPath)
		if nil != statErr {
			err = statErr
			return
		}
		f, openErr := os.Open(assetAbsPath)
		if nil != openErr {
			err = openErr
			return
		}

		hash, hashErr := util.GetEtagByHandle(f, fi.Size())
		if nil != hashErr {
			f.Close()
			return
		}

		if 1 > fi.Size() {
			hash = "random_1_" + gulu.Rand.String(12)
		}

		existAssetPath := GetAssetPathByHash(hash, bt.BoxID)
		if "" != existAssetPath {
			originalName := util.RemoveID(filepath.Base(existAssetPath))
			if strings.ToLower(fName) != strings.ToLower(originalName) {
				hash = "random_2_" + gulu.Rand.String(12)
			}
		}

		if "" != existAssetPath && !strings.HasPrefix(hash, "random_") {
			succMap[baseName] = strings.TrimPrefix(existAssetPath, "/")
			f.Close()
		} else {
			blockID := ast.NewNodeID()
			if IsEncryptedBox(bt.BoxID) {

				fName = encryptedAssetName(util.Ext(fName), blockID)

				if mapErr := writeAssetNameMapping(bt.BoxID, fName, baseName); mapErr != nil {
					err = mapErr
					f.Close()
					return
				}
			} else {
				fName = util.AssetName(fName, blockID)
			}
			writePath := filepath.Join(assetsDirPath, fName)
			if _, err = f.Seek(0, io.SeekStart); err != nil {
				f.Close()
				return
			}
			if err = writeAssetFile(writePath, f, bt.BoxID); err != nil {
				f.Close()
				return
			}
			f.Close()

			p := "assets/" + fName
			if IsEncryptedBox(bt.BoxID) {
				p += "?box=" + bt.BoxID
			}
			succMap[baseName] = p
			if !IsEncryptedBox(bt.BoxID) {
				cache.SetAssetHash(hash, p)
			}
		}
	}
	IncSync()
	return
}

func Upload(c *gin.Context) {
	ret := gulu.Ret.NewResult()
	defer c.JSON(200, ret)

	form, err := c.MultipartForm()
	if err != nil {
		logging.LogErrorf("insert asset failed: %s", err)
		ret.Code = -1
		ret.Msg = err.Error()
		return
	}
	assetsDirPath := filepath.Join(util.DataDir, "assets")
	var uploadBoxID string
	if nil != form.Value["id"] {
		id := form.Value["id"][0]
		bt := treenode.GetBlockTree(id)
		if nil == bt {

			for _, encBoxID := range treenode.GetOpenedEncryptedBoxIDs() {
				if encBT := treenode.GetBlockTreeInBox(id, encBoxID); nil != encBT {
					bt = encBT
					break
				}
			}
		}
		if nil == bt {
			ret.Code = -1
			ret.Msg = Conf.Language(71)
			return
		}
		uploadBoxID = bt.BoxID
		docDirLocalPath := filepath.Join(util.DataDir, bt.BoxID, path.Dir(bt.Path))
		assetsDirPath = getAssetsDir(filepath.Join(util.DataDir, bt.BoxID), docDirLocalPath)
	}

	relAssetsDirPath := "assets"
	if nil != form.Value["assetsDirPath"] {
		relAssetsDirPath = form.Value["assetsDirPath"][0]
		assetsDirPath = filepath.Join(util.DataDir, relAssetsDirPath)
		if !util.IsAbsPathInWorkspace(assetsDirPath) {
			ret.Code = -1
			ret.Msg = "Path [" + assetsDirPath + "] is not in workspace"
			return
		}

		if pathBox := ExtractBoxIDFromAssetsPath(assetsDirPath); pathBox != "" && IsEncryptedBox(pathBox) {
			uploadBoxID = pathBox
		}
	}
	if !gulu.File.IsExist(assetsDirPath) {
		if err = os.MkdirAll(assetsDirPath, 0755); err != nil {
			ret.Code = -1
			ret.Msg = err.Error()
			return
		}
	}

	var errFiles []string
	succMap := map[string]any{}
	files := form.File["file[]"]
	skipIfDuplicated := false
	if nil != form.Value["skipIfDuplicated"] {
		skipIfDuplicated = "true" == form.Value["skipIfDuplicated"][0]
	}

	for _, file := range files {
		baseName := file.Filename
		_, lastID := util.LastID(baseName)
		if !ast.IsNodeIDPattern(lastID) {
			lastID = ""
		}

		needUnzip2Dir := false
		if gulu.OS.IsDarwin() {
			if strings.HasSuffix(baseName, ".rtfd.zip") {
				needUnzip2Dir = true
			}
		}

		fName := baseName
		fName = util.FilterUploadFileName(fName)
		ext := filepath.Ext(fName)
		fName = strings.TrimSuffix(fName, ext)
		ext = strings.ToLower(ext)
		fName += ext
		f, openErr := file.Open()
		if nil != openErr {
			errFiles = append(errFiles, fName)
			ret.Msg = openErr.Error()
			break
		}
		if needUnzip2Dir && IsEncryptedBox(uploadBoxID) {
			errFiles = append(errFiles, fName)
			ret.Msg = "directory assets are not supported in encrypted notebooks"
			f.Close()
			break
		}

		hash, hashErr := util.GetEtagByHandle(f, file.Size)
		if nil != hashErr {
			errFiles = append(errFiles, fName)
			ret.Msg = err.Error()
			f.Close()
			break
		}

		if 1 > file.Size {
			hash = "random_1_" + gulu.Rand.String(12)
		}

		existAssetPath := GetAssetPathByHash(hash, uploadBoxID)
		if "" != existAssetPath {
			originalName := util.RemoveID(filepath.Base(existAssetPath))
			if strings.ToLower(fName) != strings.ToLower(originalName) {
				hash = "random_2_" + gulu.Rand.String(12)
			}
		}

		if "" != existAssetPath && !strings.HasPrefix(hash, "random_") {
			succMap[baseName] = strings.TrimPrefix(existAssetPath, "/")
			f.Close()
		} else {
			if skipIfDuplicated {

				pattern := assetsDirPath + string(os.PathSeparator) + strings.TrimSuffix(fName, ext)
				_, patternLastID := util.LastID(fName)
				if lastID != "" && lastID != patternLastID {

					pattern = assetsDirPath + string(os.PathSeparator) + "*" + lastID + ext
				} else {
					pattern += "*" + ext
				}

				matches, globErr := filepath.Glob(pattern)
				if nil != globErr {
					logging.LogErrorf("glob failed: %s", globErr)
				} else {
					if 0 < len(matches) {
						fName = filepath.Base(matches[0])
						succMap[baseName] = strings.TrimPrefix(path.Join(relAssetsDirPath, fName), "/")
						f.Close()
						break
					}
				}
			}

			if "" == lastID {
				lastID = ast.NewNodeID()
			}
			if IsEncryptedBox(uploadBoxID) {

				fName = encryptedAssetName(util.Ext(fName), lastID)

				if mapErr := writeAssetNameMapping(uploadBoxID, fName, baseName); mapErr != nil {
					errFiles = append(errFiles, fName)
					ret.Msg = mapErr.Error()
					f.Close()
					break
				}
			} else {
				fName = util.AssetName(fName, lastID)
			}
			writePath := filepath.Join(assetsDirPath, fName)
			tmpDir := filepath.Join(util.TempDir, "convert", "zip", gulu.Rand.String(7))
			if needUnzip2Dir {
				if err = os.MkdirAll(tmpDir, 0755); err != nil {
					errFiles = append(errFiles, fName)
					ret.Msg = err.Error()
					f.Close()
					break
				}
				writePath = filepath.Join(tmpDir, fName)
			}

			if _, err = f.Seek(0, io.SeekStart); err != nil {
				logging.LogErrorf("seek failed: %s", err)
				errFiles = append(errFiles, fName)
				ret.Msg = err.Error()
				f.Close()
				break
			}
			if err = writeAssetFile(writePath, f, uploadBoxID); err != nil {
				logging.LogErrorf("write file failed: %s", err)
				errFiles = append(errFiles, fName)
				ret.Msg = err.Error()
				f.Close()
				break
			}
			f.Close()

			if needUnzip2Dir {
				baseName = strings.TrimSuffix(file.Filename, ".rtfd.zip") + ".rtfd"
				fName = baseName
				fName = util.FilterUploadFileName(fName)
				ext = filepath.Ext(fName)
				fName = strings.TrimSuffix(fName, ext)
				ext = strings.ToLower(ext)
				fName += ext
				fName = util.AssetName(fName, ast.NewNodeID())
				tmpDir2 := filepath.Join(util.TempDir, "convert", "zip", gulu.Rand.String(7))
				if err = gulu.Zip.Unzip(writePath, tmpDir2); err != nil {
					errFiles = append(errFiles, fName)
					ret.Msg = err.Error()
					break
				}

				entries, readErr := os.ReadDir(tmpDir2)
				if nil != readErr {
					logging.LogErrorf("read dir [%s] failed: %s", tmpDir2, readErr)
					errFiles = append(errFiles, fName)
					ret.Msg = readErr.Error()
					break
				}
				if 1 > len(entries) {
					logging.LogErrorf("read dir [%s] failed: no entry", tmpDir2)
					errFiles = append(errFiles, fName)
					ret.Msg = "no entry"
					break
				}
				dirName := entries[0].Name()
				srcDir := filepath.Join(tmpDir2, dirName)
				entries, readErr = os.ReadDir(srcDir)
				if nil != readErr {
					logging.LogErrorf("read dir [%s] failed: %s", filepath.Join(tmpDir2, entries[0].Name()), readErr)
					errFiles = append(errFiles, fName)
					ret.Msg = readErr.Error()
					break
				}
				destDir := filepath.Join(assetsDirPath, fName)
				for _, entry := range entries {
					from := filepath.Join(srcDir, entry.Name())
					to := filepath.Join(destDir, entry.Name())
					if copyErr := gulu.File.Copy(from, to); nil != copyErr {
						logging.LogErrorf("copy [%s] to [%s] failed: %s", from, to, copyErr)
						errFiles = append(errFiles, fName)
						ret.Msg = copyErr.Error()
						break
					}
				}
				os.RemoveAll(tmpDir)
				os.RemoveAll(tmpDir2)
			}

			p := strings.TrimPrefix(path.Join(relAssetsDirPath, fName), "/")
			if uploadBoxID != "" && IsEncryptedBox(uploadBoxID) {
				p += "?box=" + uploadBoxID
			}
			succMap[baseName] = p
			if uploadBoxID == "" || !IsEncryptedBox(uploadBoxID) {
				cache.SetAssetHash(hash, p)
			}
		}
	}

	ret.Data = map[string]any{
		"errFiles": errFiles,
		"succMap":  succMap,
	}

	IncSync()
}

func getAssetsDir(boxLocalPath, docDirLocalPath string) (assets string) {
	assets = filepath.Join(docDirLocalPath, "assets")
	if !filelock.IsExist(assets) {
		assets = filepath.Join(boxLocalPath, "assets")
		if !filelock.IsExist(assets) {

			boxID := filepath.Base(boxLocalPath)
			if IsEncryptedBox(boxID) {
				_ = os.MkdirAll(assets, 0755)
				return
			}
			assets = filepath.Join(util.DataDir, "assets")
		}
	}
	return
}

func writeAssetFile(writePath string, src io.Reader, boxID string) (err error) {

	pathBoxID := ExtractBoxIDFromAssetsPath(writePath)

	if boxID != "" && pathBoxID != "" && boxID != pathBoxID {
		return fmt.Errorf("boxID mismatch: param=%s, path=%s", boxID, pathBoxID)
	}

	if pathBoxID == "" && boxID != "" && IsEncryptedBox(boxID) {
		return fmt.Errorf("encrypted box asset must be written inside the box directory, got global path: %s", writePath)
	}
	actualBoxID := pathBoxID
	if actualBoxID == "" {
		actualBoxID = boxID
	}
	if actualBoxID != "" && IsEncryptedBox(actualBoxID) {
		HoldBoxReadLock(actualBoxID)
		defer ReleaseBoxReadLock(actualBoxID)
		dek, dekErr := GetDEKIfUnlocked(actualBoxID)
		if dekErr != nil {

			return dekErr
		}

		raw, readErr := io.ReadAll(src)
		if readErr != nil {
			return readErr
		}
		enc, encErr := EncryptAsset(actualBoxID, filepath.Base(writePath), dek, raw)
		if encErr != nil {
			return encErr
		}
		return filelock.WriteFile(writePath, enc)
	}
	return filelock.WriteFileByReader(writePath, src)
}

func StoreAssetForBox(boxID, assetDirPath, originalName string, data []byte) (diskName string, err error) {
	return storeAssetForBox(boxID, assetDirPath, originalName, data)
}

func storeAssetForBox(boxID, assetDirPath, originalName string, data []byte) (diskName string, err error) {
	if IsEncryptedBox(boxID) {
		HoldBoxReadLock(boxID)
		defer ReleaseBoxReadLock(boxID)

		ext := filepath.Ext(originalName)
		blockID := ast.NewNodeID()
		diskName = encryptedAssetName(ext, blockID)

		if mapErr := writeAssetNameMappingLocked(boxID, diskName, originalName); mapErr != nil {
			return "", mapErr
		}

		dek, dekErr := GetDEKIfUnlocked(boxID)
		if dekErr != nil {
			return "", dekErr
		}
		enc, encErr := EncryptAsset(boxID, diskName, dek, data)
		if encErr != nil {
			return "", encErr
		}
		writePath := filepath.Join(assetDirPath, diskName)
		if err = filelock.WriteFile(writePath, enc); err != nil {
			return "", err
		}
		return diskName, nil
	}

	diskName = util.AssetName(originalName, ast.NewNodeID())
	writePath := filepath.Join(assetDirPath, diskName)
	if err = filelock.WriteFile(writePath, data); err != nil {
		return "", err
	}
	return diskName, nil
}

func encryptedAssetName(ext, blockID string) string {
	return gulu.Rand.String(16) + "-" + blockID + ext
}

func assetNameMappingPath(boxID string) string {
	return filepath.Join(util.DataDir, boxID, "assets", ".names.json")
}

var assetNameMappingLocks sync.Map // map[string]*sync.Mutex

func writeAssetNameMapping(boxID, diskName, originalName string) error {
	if boxID == "" || !IsEncryptedBox(boxID) {
		return nil
	}
	HoldBoxReadLock(boxID)
	defer ReleaseBoxReadLock(boxID)
	return writeAssetNameMappingLocked(boxID, diskName, originalName)
}

func writeAssetNameMappingLocked(boxID, diskName, originalName string) error {
	muI, _ := assetNameMappingLocks.LoadOrStore(boxID, &sync.Mutex{})
	mu := muI.(*sync.Mutex)
	mu.Lock()
	defer mu.Unlock()

	mapping := readAssetNameMappingLocked(boxID)
	mapping[diskName] = originalName
	data, err := json.Marshal(mapping)
	if err != nil {
		return fmt.Errorf("marshal asset name mapping failed: %w", err)
	}
	dek, err := GetDEK(boxID)
	if err != nil || dek == nil {
		return fmt.Errorf("get DEK for asset name mapping failed: %w", err)
	}
	enc, err := EncryptAssetNameMapping(boxID, dek, data)
	if err != nil {
		return fmt.Errorf("encrypt asset name mapping failed: %w", err)
	}

	if err = atomicWriteFile(assetNameMappingPath(boxID), enc); err != nil {
		return fmt.Errorf("write asset name mapping failed: %w", err)
	}
	return nil
}

func removeAssetNameMapping(boxID, diskName string) error {
	if boxID == "" || !IsEncryptedBox(boxID) {
		return nil
	}
	HoldBoxReadLock(boxID)
	defer ReleaseBoxReadLock(boxID)

	muI, _ := assetNameMappingLocks.LoadOrStore(boxID, &sync.Mutex{})
	mu := muI.(*sync.Mutex)
	mu.Lock()
	defer mu.Unlock()
	mapping := readAssetNameMappingLocked(boxID)
	if _, exists := mapping[diskName]; !exists {
		return nil
	}
	delete(mapping, diskName)
	data, err := json.Marshal(mapping)
	if err != nil {
		return fmt.Errorf("marshal asset name mapping failed: %w", err)
	}
	dek, err := GetDEK(boxID)
	if err != nil || dek == nil {
		return fmt.Errorf("get DEK for asset name mapping failed: %w", err)
	}
	enc, err := EncryptAssetNameMapping(boxID, dek, data)
	if err != nil {
		return fmt.Errorf("encrypt asset name mapping failed: %w", err)
	}
	if err = atomicWriteFile(assetNameMappingPath(boxID), enc); err != nil {
		return fmt.Errorf("write asset name mapping failed: %w", err)
	}
	return nil
}

func readAssetNameMapping(boxID string) map[string]string {
	ret := map[string]string{}
	if boxID == "" || !IsEncryptedBox(boxID) {
		return ret
	}
	HoldBoxReadLock(boxID)
	defer ReleaseBoxReadLock(boxID)
	return readAssetNameMappingLocked(boxID)
}

func readAssetNameMappingLocked(boxID string) map[string]string {
	ret := map[string]string{}
	p := assetNameMappingPath(boxID)
	enc, err := filelock.ReadFile(p)
	if err != nil {
		return ret
	}
	dek, err := GetDEK(boxID)
	if err != nil || dek == nil {
		return ret
	}
	data, err := DecryptAssetNameMapping(boxID, dek, enc)
	if err != nil {
		logging.LogErrorf("decrypt asset name mapping failed: %s", err)
		return ret
	}
	if err = json.Unmarshal(data, &ret); err != nil {
		logging.LogErrorf("unmarshal asset name mapping failed: %s", err)
		return map[string]string{}
	}
	return ret
}

func LookupAssetOriginalName(boxID, diskName string) string {
	return readAssetNameMapping(boxID)[diskName]
}

func LookupAssetOriginalNameLocked(boxID, diskName string) string {
	return readAssetNameMappingLocked(boxID)[diskName]
}
