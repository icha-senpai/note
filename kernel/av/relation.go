package av

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/icha-senpai/note/third_party/forks/gulu"
	"github.com/icha-senpai/note/kernel/util"
	"github.com/icha-senpai/note/third_party/forks/filelock"
	"github.com/icha-senpai/note/third_party/forks/logging"
	"github.com/icha-senpai/note/third_party/forks/github/vmihailenco/msgpack/v5"
)

var (
	attributeViewRelationsLock = sync.Mutex{}
)

func relationsPath(boxID string) string {
	if boxID != "" {
		return filepath.Join(util.DataDir, boxID, "storage", "av", "relations.msgpack")
	}
	return filepath.Join(util.DataDir, "storage", "av", "relations.msgpack")
}

func readRelations(boxID string) (avRels map[string][]string) {
	avRels = map[string][]string{}
	p := relationsPath(boxID)
	if !filelock.IsExist(p) {
		return
	}
	data, err := filelock.ReadFile(p)
	if err != nil {
		logging.LogErrorf("read attribute view relations failed: %s", err)
		return
	}
	if boxID != "" {
		dec, decErr := decryptAVData(boxID, "relation", data)
		if decErr != nil {
			logging.LogErrorf("decrypt attribute view relations failed: %s", decErr)
			return
		}
		data = dec
	}
	if err = msgpack.Unmarshal(data, &avRels); err != nil {
		logging.LogErrorf("unmarshal attribute view relations failed: %s", err)
		return
	}
	return
}

func writeRelations(boxID string, avRels map[string][]string) {
	p := relationsPath(boxID)
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		logging.LogErrorf("create attribute view dir failed: %s", err)
		return
	}
	data, err := msgpack.Marshal(avRels)
	if err != nil {
		logging.LogErrorf("marshal attribute view relations failed: %s", err)
		return
	}
	if boxID != "" {
		enc, encErr := encryptAVData(boxID, "relation", data)
		if encErr != nil {
			logging.LogErrorf("encrypt attribute view relations failed: %s", encErr)
			return
		}
		data = enc
	}
	if err = filelock.WriteFile(p, data); err != nil {
		logging.LogErrorf("write attribute view relations failed: %s", err)
		return
	}
}

func relationsBoxIDByAvID(avID string) string {
	_, boxID := FindAttributeViewPath(avID)
	return boxID
}

func GetSrcAvIDs(destAvID string) []string {
	attributeViewRelationsLock.Lock()
	defer attributeViewRelationsLock.Unlock()

	boxID := relationsBoxIDByAvID(destAvID)
	avRels := readRelations(boxID)
	srcAvIDs := avRels[destAvID]
	if nil == srcAvIDs {
		return nil
	}
	return srcAvIDs
}

func RemoveAvRel(srcAvID, destAvID string) {
	attributeViewRelationsLock.Lock()
	defer attributeViewRelationsLock.Unlock()

	boxID := relationsBoxIDByAvID(destAvID)
	avRels := readRelations(boxID)

	srcAvIDs := avRels[destAvID]
	if nil == srcAvIDs {
		return
	}

	var newAvIDs []string
	for _, v := range srcAvIDs {
		if v != srcAvID {
			newAvIDs = append(newAvIDs, v)
		}
	}
	avRels[destAvID] = newAvIDs
	writeRelations(boxID, avRels)
}

func UpsertAvBackRel(srcAvID, destAvID string) {
	attributeViewRelationsLock.Lock()
	defer attributeViewRelationsLock.Unlock()

	_, srcBox := FindAttributeViewPath(srcAvID)
	_, destBox := FindAttributeViewPath(destAvID)
	if AVIsEncryptedBox != nil {
		srcEnc := srcBox != "" && AVIsEncryptedBox(srcBox)
		destEnc := destBox != "" && AVIsEncryptedBox(destBox)
		if srcEnc != destEnc || (srcEnc && destEnc && srcBox != destBox) {
			logging.LogWarnf("skip cross-boundary AV relation: src=%s(box=%s) dest=%s(box=%s)", srcAvID, srcBox, destAvID, destBox)
			return
		}
	}

	boxID := destBox
	avRels := readRelations(boxID)

	srcAvIDs := avRels[destAvID]
	srcAvIDs = append(srcAvIDs, srcAvID)
	srcAvIDs = gulu.Str.RemoveDuplicatedElem(srcAvIDs)
	avRels[destAvID] = srcAvIDs
	writeRelations(boxID, avRels)
}
