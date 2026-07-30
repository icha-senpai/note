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

package filesys

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/icha-senpai/note/third_party/forks/lute/ast"
	"github.com/icha-senpai/note/kernel/util"
)

var DEKProvider func(boxID string) ([]byte, error)

var DEKLockAcquire func(boxID string)
var DEKLockRelease func(boxID string)

func SyObjectBase(relativePath string) (string, error) {
	p := filepath.ToSlash(relativePath)
	p = strings.TrimPrefix(p, "/")
	base := p
	if idx := strings.LastIndex(p, "/"); idx >= 0 {
		base = p[idx+1:]
	}
	if !strings.HasSuffix(base, ".sy") {
		return "", fmt.Errorf("invalid .sy base name [%s]: must end with .sy", base)
	}
	stem := strings.TrimSuffix(base, ".sy")
	if !ast.IsNodeIDPattern(stem) {
		return "", fmt.Errorf("invalid .sy base name [%s]: stem is not a node ID", base)
	}
	return base, nil
}

func SyAAD(boxID, relativePath string) (string, error) {
	base, err := SyObjectBase(relativePath)
	if err != nil {
		return "", err
	}
	return "scribli:v1:file:" + boxID + ":" + base, nil
}

func encryptedBox(boxID string) bool {
	if DEKProvider == nil {
		return false
	}
	if DEKLockAcquire != nil {
		DEKLockAcquire(boxID)
		defer DEKLockRelease(boxID)
	}
	dek, err := DEKProvider(boxID)
	return err == nil && dek != nil
}

func encryptData(boxID, relativePath string, data []byte) ([]byte, error) {
	if DEKProvider == nil {
		return data, nil
	}
	if DEKLockAcquire != nil {
		DEKLockAcquire(boxID)
		defer DEKLockRelease(boxID)
	}
	dek, err := DEKProvider(boxID)
	if err != nil {
		return nil, err
	}
	if dek == nil {
		return data, nil
	}
	fileKey := util.DeriveSubKey(dek, "scribli/file")
	aad, err := SyAAD(boxID, relativePath)
	if err != nil {
		return nil, err
	}
	return util.EncryptWithAAD(fileKey, data, []byte(aad))
}

func decryptData(boxID, relativePath string, data []byte) ([]byte, error) {
	if DEKProvider == nil {
		return data, nil
	}
	if DEKLockAcquire != nil {
		DEKLockAcquire(boxID)
		defer DEKLockRelease(boxID)
	}
	dek, err := DEKProvider(boxID)
	if err != nil {
		return nil, err
	}
	if dek == nil {
		return data, nil
	}
	fileKey := util.DeriveSubKey(dek, "scribli/file")
	aad, err := SyAAD(boxID, relativePath)
	if err != nil {
		return nil, err
	}
	return util.DecryptWithAAD(fileKey, data, []byte(aad))
}

func docIALBoxID(absPath string) string {
	absPath = filepath.ToSlash(absPath)
	dataDir := filepath.ToSlash(util.DataDir)
	rel, err := filepath.Rel(dataDir, absPath)
	if err != nil {
		return ""
	}
	rel = filepath.ToSlash(rel)
	if strings.HasPrefix(rel, "..") || rel == "." || rel == "" {
		return ""
	}
	parts := strings.SplitN(rel, "/", 2)
	boxID := parts[0]
	if !ast.IsNodeIDPattern(boxID) {
		return ""
	}
	return boxID
}
