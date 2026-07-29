// SiYuan - Refactor your thinking
// Copyright (c) 2020-present, b3log.org
//

package av

import (
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/88250/lute/ast"
	"github.com/siyuan-note/filelock"
	"github.com/siyuan-note/logging"
	"github.com/siyuan-note/siyuan/kernel/util"
	"github.com/vmihailenco/msgpack/v5"
)

var AVDEKProvider func(boxID string) ([]byte, error)

var AVLockAcquire func(boxID string)
var AVLockRelease func(boxID string)

var AVEncryptedBoxIDs func() []string

var AVIsEncryptedBox func(boxID string) bool

var AVGetBlockBoxID func(blockID string) string

var pendingAVBox = map[string]string{}
var pendingAVBoxLock = sync.RWMutex{}

func SetAVBoxID(avID, boxID string) {
	pendingAVBoxLock.Lock()
	defer pendingAVBoxLock.Unlock()
	if boxID != "" {
		pendingAVBox[avID] = boxID
	} else {
		delete(pendingAVBox, avID)
	}
}

func GetAVBoxID(avID string) string {
	pendingAVBoxLock.RLock()
	defer pendingAVBoxLock.RUnlock()
	return pendingAVBox[avID]
}

func attributeViewDataPathByBox(avID, boxID string) string {
	if !ast.IsNodeIDPattern(avID) || (boxID != "" && !ast.IsNodeIDPattern(boxID)) {
		return ""
	}

	if boxID != "" {
		return filepath.Join(util.DataDir, boxID, "storage", "av", avID+".json")
	}
	return filepath.Join(util.DataDir, "storage", "av", avID+".json")
}

func FindAttributeViewPath(avID string) (path string, boxID string) {
	if !ast.IsNodeIDPattern(avID) {
		return
	}

	if pendingBoxID := GetAVBoxID(avID); pendingBoxID != "" {
		encPath := attributeViewDataPathByBox(avID, pendingBoxID)

		return encPath, pendingBoxID
	}

	globalPath := attributeViewDataPathByBox(avID, "")
	if filelock.IsExist(globalPath) {
		return globalPath, ""
	}

	if AVEncryptedBoxIDs != nil {
		for _, encBoxID := range AVEncryptedBoxIDs() {
			encPath := attributeViewDataPathByBox(avID, encBoxID)
			if filelock.IsExist(encPath) {
				return encPath, encBoxID
			}
		}
	}
	return "", ""
}

func FindAttributeViewPathInBox(avID, boxID string) (path string, retBoxID string) {
	if !ast.IsNodeIDPattern(avID) || (boxID != "" && !ast.IsNodeIDPattern(boxID)) {
		return
	}

	if pendingBoxID := GetAVBoxID(avID); pendingBoxID != "" {
		if pendingBoxID == boxID {

			return attributeViewDataPathByBox(avID, pendingBoxID), pendingBoxID
		}

	}
	avPath := attributeViewDataPathByBox(avID, boxID)
	if filelock.IsExist(avPath) {
		return avPath, boxID
	}
	return "", boxID
}

func readAttributeViewData(avID string) ([]byte, error) {
	path, boxID := FindAttributeViewPath(avID)
	if path == "" {
		return nil, nil
	}
	data, err := filelock.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if boxID != "" {
		data, err = decryptAVData(boxID, avID, data)
		if err != nil {
			return nil, err
		}
	}
	return data, nil
}

func ReadAttributeViewData(avID string) ([]byte, error) {
	return readAttributeViewData(avID)
}

func ReadAttributeViewDataInBox(avID, boxID string) ([]byte, error) {
	path, retBoxID := FindAttributeViewPathInBox(avID, boxID)
	if path == "" {
		return nil, nil
	}
	data, err := filelock.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if retBoxID != "" {
		data, err = decryptAVData(retBoxID, avID, data)
		if err != nil {
			return nil, err
		}
	}
	return data, nil
}

func writeAttributeViewData(avID, boxID string, data []byte) error {
	if !ast.IsNodeIDPattern(avID) {
		return ErrInvalidAttributeViewID
	}
	if boxID != "" && !ast.IsNodeIDPattern(boxID) {
		return ErrInvalidBoxID
	}

	path := attributeViewDataPathByBox(avID, boxID)

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	if boxID != "" {
		var err error
		data, err = encryptAVData(boxID, avID, data)
		if err != nil {
			return err
		}
	}
	return filelock.WriteFile(path, data)
}

func mirrorBlocksPath(boxID string) string {
	if boxID != "" {
		return filepath.Join(util.DataDir, boxID, "storage", "av", "blocks.msgpack")
	}
	return filepath.Join(util.DataDir, "storage", "av", "blocks.msgpack")
}

func mirrorBlocksPathByAvID(avID string) string {
	_, boxID := FindAttributeViewPath(avID)
	return mirrorBlocksPath(boxID)
}

func readMirrorBlocks(boxID string) (ret map[string][]string) {
	ret = map[string][]string{}
	p := mirrorBlocksPath(boxID)
	if !filelock.IsExist(p) {
		return
	}
	data, err := filelock.ReadFile(p)
	if err != nil {
		logging.LogErrorf("read attribute view blocks failed: %s", err)
		return
	}
	if boxID != "" {

		dec, decErr := decryptAVData(boxID, "mirror", data)
		if decErr != nil {
			logging.LogErrorf("decrypt attribute view blocks failed: %s", decErr)
			return
		}
		data = dec
	}
	if err = msgpack.Unmarshal(data, &ret); err != nil {
		logging.LogErrorf("unmarshal attribute view blocks failed: %s", err)
		return
	}
	return
}

func writeMirrorBlocks(boxID string, data map[string][]string) error {
	p := mirrorBlocksPath(boxID)
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		return err
	}
	raw, err := msgpack.Marshal(data)
	if err != nil {
		return err
	}
	if boxID != "" {

		enc, encErr := encryptAVData(boxID, "mirror", raw)
		if encErr != nil {
			return encErr
		}
		raw = enc
	}
	return filelock.WriteFile(p, raw)
}

func avBoxIDFromPath(absPath string) string {
	rel, err := filepath.Rel(util.DataDir, absPath)
	if err != nil {
		return ""
	}
	rel = filepath.ToSlash(rel)
	parts := strings.SplitN(rel, "/", 2)
	if len(parts) < 2 {
		return ""
	}
	boxID := parts[0]

	if boxID == "storage" {
		return ""
	}
	return boxID
}

func encryptAVData(boxID, avID string, data []byte) ([]byte, error) {
	if AVDEKProvider == nil {
		return data, nil
	}
	if AVLockAcquire != nil {
		AVLockAcquire(boxID)
		defer AVLockRelease(boxID)
	}
	dek, err := AVDEKProvider(boxID)
	if err != nil {
		return nil, err
	}
	if dek == nil {
		return data, nil
	}
	avKey := util.DeriveSubKey(dek, "siyuan/av")
	aad := avAAD(boxID, avID)
	return util.EncryptWithAAD(avKey, data, []byte(aad))
}

func decryptAVData(boxID, avID string, data []byte) ([]byte, error) {
	if AVDEKProvider == nil {
		return data, nil
	}
	if AVLockAcquire != nil {
		AVLockAcquire(boxID)
		defer AVLockRelease(boxID)
	}
	return decryptAVDataLocked(boxID, avID, data)
}

func decryptAVDataLocked(boxID, avID string, data []byte) ([]byte, error) {
	if AVDEKProvider == nil {
		return data, nil
	}
	dek, err := AVDEKProvider(boxID)
	if err != nil {
		return nil, err
	}
	if dek == nil {
		return data, nil
	}
	avKey := util.DeriveSubKey(dek, "siyuan/av")
	aad := avAAD(boxID, avID)
	return util.DecryptWithAAD(avKey, data, []byte(aad))
}

func avAAD(boxID, avID string) string {
	switch avID {
	case "mirror":
		return "siyuan:v1:av-mirror:" + boxID
	case "relation":
		return "siyuan:v1:av-relation:" + boxID
	default:
		return "siyuan:v1:av:" + boxID + ":" + avID
	}
}

func EncryptAVData(boxID, avID string, data []byte) ([]byte, error) {
	return encryptAVData(boxID, avID, data)
}

func DecryptAVData(boxID, avID string, data []byte) ([]byte, error) {
	return decryptAVData(boxID, avID, data)
}

func DecryptAVDataLocked(boxID, avID string, data []byte) ([]byte, error) {
	return decryptAVDataLocked(boxID, avID, data)
}
