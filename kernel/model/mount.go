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
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/icha-senpai/note/kernel/sql"
	"github.com/icha-senpai/note/kernel/treenode"
	"github.com/icha-senpai/note/kernel/util"
	"github.com/icha-senpai/note/third_party/forks/filelock"
	"github.com/icha-senpai/note/third_party/forks/gulu"
	"github.com/icha-senpai/note/third_party/forks/logging"
	"github.com/icha-senpai/note/third_party/forks/lute/ast"
)

func GetBoxByName(name string) (ret *Box) {
	for _, box := range Conf.GetOpenedBoxes() {
		if box.Name == name {
			ret = box
			return
		}
	}
	return
}

func CreateBox(name string) (id string, err error) {
	name = normalizeBoxName(name)
	if 512 < utf8.RuneCountInString(name) {

		err = errors.New(Conf.Language(106))
		return
	}
	FlushTxQueue()

	createDocLock.Lock()
	defer createDocLock.Unlock()

	boxes, _ := ListNotebooks()
	for i, b := range boxes {
		c := b.GetConf()
		c.Sort = i + 1
		if err := b.SaveConf(c); err != nil {
			logging.LogErrorf("save box conf [%s] failed: %s", b.ID, err)
		}
	}

	id = ast.NewNodeID()
	boxLocalPath := filepath.Join(util.DataDir, id)
	err = os.MkdirAll(boxLocalPath, 0755)
	if err != nil {
		return
	}

	box := &Box{ID: id, Name: name}
	boxConf := box.GetConf()
	boxConf.Name = name
	if err := box.SaveConf(boxConf); err != nil {
		logging.LogErrorf("save box conf [%s] failed: %s", id, err)
	}
	if _, err = ensureBoxDoc0(id); err != nil {
		treenode.RemoveBlockTreesByBoxID(id)
		sql.DeleteBoxQueue(id)
		if removeErr := filelock.Remove(boxLocalPath); nil != removeErr {
			logging.LogErrorf("remove box [%s] after initializing box document failed: %s", id, removeErr)
		}
		return "", err
	}
	IncSync()
	logging.LogInfof("created box [%s]", id)
	return
}

func RenameBox(boxID, name string) (err error) {
	box := Conf.Box(boxID)
	if nil == box {
		return errors.New(Conf.Language(0))
	}

	name = normalizeBoxName(name)
	if 512 < utf8.RuneCountInString(name) {

		err = errors.New(Conf.Language(106))
		return
	}

	boxConf := box.GetConf()
	boxConf.Name = name
	box.Name = name
	if err = box.SaveConf(boxConf); err != nil {
		logging.LogErrorf("save box conf [%s] failed: %s", boxID, err)
		return
	}
	if err = renameBoxDoc(boxID, name); err != nil {
		logging.LogErrorf("rename box document [box=%s] failed: %s", boxID, err)
		return
	}
	IncSync()
	logging.LogInfof("renamed box [%s] to [%s]", boxID, name)
	return
}

func normalizeBoxName(name string) string {
	name = normalizeDocTitle(name)
	if "" == name {
		name = normalizeDocTitle(Conf.language(105))
	}
	return name
}

var boxLock = sync.Map{}

func RemoveBox(boxID string) (err error) {
	if _, ok := boxLock.Load(boxID); ok {
		err = errors.New(Conf.language(239))
		return
	}

	boxLock.Store(boxID, true)
	defer boxLock.Delete(boxID)

	if util.IsReservedFilename(boxID) {
		return fmt.Errorf("can not remove [%s] caused by it is a reserved file", boxID)
	}

	FlushTxQueue()
	createDocLock.Lock()
	defer createDocLock.Unlock()

	localPath := filepath.Join(util.DataDir, boxID)
	if !filelock.IsExist(localPath) {
		return
	}
	if !gulu.File.IsDir(localPath) {
		return fmt.Errorf("can not remove [%s] caused by it is not a dir", boxID)
	}

	unmount0(boxID)

	isEncrypted := IsEncryptedBox(boxID)

	var historyDir string
	historyDir, err = getHistoryDir(HistoryOpDelete)
	if err != nil {
		logging.LogErrorf("get history dir failed: %s", err)
		return
	}

	p := strings.TrimPrefix(localPath, util.DataDir)
	historyPath := filepath.Join(historyDir, p)
	if err = filelock.Copy(localPath, historyPath); err != nil {
		logging.LogErrorf("gen sync history failed: %s", err)
		return
	}

	if !isEncrypted {
		copyBoxAssetsToDataAssets(boxID)
	}

	if isEncrypted {
		if rmErr := os.RemoveAll(filepath.Join(util.TempDir, "export", boxID)); rmErr != nil {
			logging.LogWarnf("remove export/[%s] dir failed: %s", boxID, rmErr)
		}
		RevokeManagedEncryptedExportsForBox(boxID)
	}

	if err = filelock.Remove(localPath); err != nil {
		return
	}

	if isEncrypted {
		sql.RemoveEncryptedDBFile(boxID)
		treenode.RemoveEncryptedBlockTreeDBFile(boxID)
	}

	IncSync()

	logging.LogInfof("removed box [%s]", boxID)
	return
}

func Unmount(boxID string) {
	FlushTxQueue()

	unmount0(boxID)

	evt := util.NewCmdResult("closeBox", 0, util.PushModeBroadcast)
	evt.Data = map[string]any{
		"box": boxID,
	}
	util.PushEvent(evt)
}

func clearDEKIfUnlockedEncryptedBox(boxID string) {
	if IsEncryptedBox(boxID) && IsBoxUnlocked(boxID) {
		ClearDEK(boxID)
	}
}

func unmount0(boxID string) {
	box := Conf.Box(boxID)
	if nil == box {

		clearDEKIfUnlockedEncryptedBox(boxID)
		return
	}

	boxConf := box.GetConf()
	boxConf.Closed = true
	if err := box.SaveConf(boxConf); err != nil {
		logging.LogErrorf("save box conf [%s] failed: %s", box.ID, err)
	}
	if IsEncryptedBox(box.ID) {

		FlushTxQueue()
		sql.FlushQueue()

		GenerateFileHistoryForBox(box)
		ClearDEK(boxID)
	} else {
		box.Unindex()
	}
}

func Mount(boxID string) (alreadyMount bool, err error) {
	if _, ok := boxLock.Load(boxID); ok {
		err = errors.New(Conf.language(239))
		return
	}

	boxLock.Store(boxID, true)
	defer boxLock.Delete(boxID)

	FlushTxQueue()
	localPath := filepath.Join(util.DataDir, boxID)
	if !gulu.File.IsDir(localPath) {
		return false, errors.New("can not open file, just support open folder only")
	}

	for _, box := range Conf.GetOpenedBoxes() {
		if box.ID == boxID {
			return true, nil
		}
	}

	if IsEncryptedBox(boxID) && !IsBoxUnlocked(boxID) {
		return false, errors.New("encrypted notebook locked, please unlock it first")
	}

	box := &Box{ID: boxID}
	boxConf := box.GetConf()
	boxConf.Closed = false
	if err := box.SaveConf(boxConf); err != nil {
		logging.LogErrorf("save box conf [%s] failed: %s", boxID, err)
	}
	if _, ensureErr := EnsureBoxDoc(boxID); nil != ensureErr {
		logging.LogErrorf("ensure box document [%s] failed: %s", boxID, ensureErr)
	}

	files, _, _ := ListDocTree(box.ID, "/", util.SortModeUnassigned, false, false, Conf.FileTree.MaxListCount)
	box = Conf.Box(boxID)
	if 0 < len(files) || (nil != box && box.Exist(boxDocPath(box.ID))) {
		box.Index()
	}

	return false, nil
}
