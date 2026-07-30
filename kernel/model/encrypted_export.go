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
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

package model

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/88250/lute/ast"
	"github.com/siyuan-note/logging"
	"github.com/siyuan-note/siyuan/kernel/util"
)

type managedEncryptedExport struct {
	boxID     string
	artifact  string
	expiresAt time.Time
}

var managedEncryptedExports = struct {
	sync.Mutex
	jobs map[string]managedEncryptedExport
}{jobs: map[string]managedEncryptedExport{}}

func newManagedEncryptedExportID() (string, error) {
	random := make([]byte, 16)
	if _, err := rand.Read(random); err != nil {
		return "", err
	}
	return hex.EncodeToString(random), nil
}

func registerManagedEncryptedExport(boxID, kind, artifact string) string {
	relativePath := path.Join(boxID, kind, filepath.Base(artifact))
	managedEncryptedExports.Lock()
	managedEncryptedExports.jobs[relativePath] = managedEncryptedExport{
		boxID:     boxID,
		artifact:  artifact,
		expiresAt: time.Now().Add(time.Hour),
	}
	managedEncryptedExports.Unlock()
	return relativePath
}

func RegisterManagedEncryptedExport(boxID, kind, artifact string) string {
	return registerManagedEncryptedExport(boxID, kind, artifact)
}

func ResolveManagedEncryptedExport(relativePath string) (boxID, artifact string, ok bool) {
	relativePath = path.Clean("/" + relativePath)
	relativePath = relativePath[1:]

	managedEncryptedExports.Lock()
	job, ok := managedEncryptedExports.jobs[relativePath]
	if !ok {
		managedEncryptedExports.Unlock()
		return "", "", false
	}
	if time.Now().After(job.expiresAt) {
		delete(managedEncryptedExports.jobs, relativePath)
		managedEncryptedExports.Unlock()
		_ = os.Remove(job.artifact)
		return "", "", false
	}
	managedEncryptedExports.Unlock()
	return job.boxID, job.artifact, true
}

func RevokeManagedEncryptedExportsForBox(boxID string) {
	managedEncryptedExports.Lock()
	defer managedEncryptedExports.Unlock()
	for relativePath, job := range managedEncryptedExports.jobs {
		if job.boxID == boxID {
			delete(managedEncryptedExports.jobs, relativePath)
		}
	}
}

func clearEncryptedExportTempOnBoot() {
	if strings.TrimSpace(util.TempDir) == "" {
		logging.LogWarnf("skip clearing stale encrypted export temp: temp dir is not initialized")
		return
	}
	exportDir := filepath.Join(util.TempDir, "export")
	entries, err := os.ReadDir(exportDir)
	if os.IsNotExist(err) {
		return
	}
	if err != nil {
		logging.LogWarnf("read export temp dir [%s] failed: %s", exportDir, err)
		return
	}
	for _, entry := range entries {
		if !ast.IsNodeIDPattern(entry.Name()) {
			continue
		}
		entryPath := filepath.Join(exportDir, entry.Name())
		if err = os.RemoveAll(entryPath); err != nil {
			logging.LogWarnf("remove stale encrypted export temp [%s] failed: %s", entryPath, err)
		}
	}
}

func IsManagedEncryptedExportPath(relativePath string) bool {
	relativePath = path.Clean("/" + relativePath)
	parts := strings.SplitN(strings.TrimPrefix(relativePath, "/"), "/", 3)
	return len(parts) >= 1 && ast.IsNodeIDPattern(parts[0])
}

func ResolveManagedExportForMobile(relativePath string) (absPath string, ok bool) {
	boxID, artifact, resolved := ResolveManagedEncryptedExport(relativePath)
	if !resolved {
		return "", false
	}
	if _, dekErr := GetDEKIfUnlocked(boxID); dekErr != nil {
		return "", false
	}
	return artifact, true
}
