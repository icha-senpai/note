// DejaVu - Data snapshot and sync.
// Copyright (c) 2022-present, b3log.org
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

package entity

import (
	"encoding/base64"
	"fmt"
	"time"

	"github.com/icha-senpai/note/third_party/forks/go-humanize"
	"github.com/icha-senpai/note/third_party/forks/encryption"
	"github.com/icha-senpai/note/third_party/forks/logging"
)

type Index struct {
	ID              string   `json:"id"`              // Hash
	Memo            string   `json:"memo"`
	Created         int64    `json:"created"`
	Files           []string `json:"files"`
	Count           int      `json:"count"`
	Size            int64    `json:"size"`
	SystemID        string   `json:"systemID"`
	SystemName      string   `json:"systemName"`
	SystemOS        string   `json:"systemOS"`
	CheckIndexID    string   `json:"checkIndexID"`    // Check Index ID
	AesKeyVerifyVal string   `json:"aesKeyVerifyVal"`
}

func (index *Index) String() string {
	return fmt.Sprintf("device=%s/%s, id=%s, files=%d, size=%s, created=%s",
		index.SystemID, index.SystemOS, index.ID, len(index.Files), humanize.BytesCustomCeil(uint64(index.Size), 2), time.UnixMilli(index.Created).Format("2006-01-02 15:04:05"))
}

func (index *Index) InitAESKeyVerifyVal(aesKey []byte) {
	data, err := encryption.AesEncrypt([]byte("scribli"), aesKey)
	if nil != err {
		logging.LogErrorf("init aes key verify val failed: %s", err)
		return
	}

	index.AesKeyVerifyVal = base64.StdEncoding.EncodeToString(data)
}

func (index *Index) VerifyAESKey(aesKey []byte) bool {
	if "" == index.AesKeyVerifyVal { // Compatibility with indexes that do not store a verification value.
		return true
	}

	data, err := base64.StdEncoding.DecodeString(index.AesKeyVerifyVal)
	if nil != err {
		logging.LogErrorf("decode aes key verify val failed: %s", err)
		return false
	}

	plainData, err := encryption.AesDecrypt(data, aesKey)
	if nil != err {
		logging.LogErrorf("decrypt aes key verify val failed: %s", err)
		return false
	}
	return "scribli" == string(plainData)
}

//
//
//
type CheckIndex struct {
	ID      string            `json:"id"`      // Hash
	IndexID string            `json:"indexID"` // Index ID
	Files   []*CheckIndexFile `json:"files"`   // File IDs
}

type CheckIndexFile struct {
	ID     string   `json:"id"`     // File ID
	Chunks []string `json:"chunks"` // Chunk IDs
}

type CheckReport struct {
	CheckTime      int64    `json:"checkTime"`
	CheckCount     int      `json:"checkCount"`
	FixCount       int      `json:"fixCount"`
	MissingObjects []string `json:"missingObjects"`
}
