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

package model

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/icha-senpai/note/third_party/forks/gulu"
	"github.com/icha-senpai/note/third_party/forks/lute/ast"
	"github.com/icha-senpai/note/kernel/cache"
	"github.com/icha-senpai/note/kernel/conf"
	"github.com/icha-senpai/note/kernel/filesys"
	"github.com/icha-senpai/note/kernel/sql"
	"github.com/icha-senpai/note/kernel/treenode"
	"github.com/icha-senpai/note/kernel/util"
	"github.com/icha-senpai/note/third_party/forks/filelock"
	"github.com/icha-senpai/note/third_party/forks/logging"
)

var kekVerifierMagic = []byte("siyuan-enc-v1")

const boxEncryptionSpec = 1

var errMasterPasswordMigrationPending = errors.New("master password migration is pending")

var notebookCryptoMu sync.Mutex

var boxLifecycleLocks = sync.Map{} // map[string]*sync.RWMutex

func acquireBoxReadLock(boxID string) {
	muI, _ := boxLifecycleLocks.LoadOrStore(boxID, &sync.RWMutex{})
	muI.(*sync.RWMutex).RLock()
}

func releaseBoxReadLock(boxID string) {
	if muI, ok := boxLifecycleLocks.Load(boxID); ok {
		muI.(*sync.RWMutex).RUnlock()
	}
}

func acquireBoxWriteLock(boxID string) {
	muI, _ := boxLifecycleLocks.LoadOrStore(boxID, &sync.RWMutex{})
	muI.(*sync.RWMutex).Lock()
}

func releaseBoxWriteLock(boxID string) {
	if muI, ok := boxLifecycleLocks.Load(boxID); ok {
		muI.(*sync.RWMutex).Unlock()
	}
}

func NotebookCryptoMuLock() { notebookCryptoMu.Lock() }

func NotebookCryptoMuUnlock() { notebookCryptoMu.Unlock() }

func notebookCryptoBackupPath() string {
	return filepath.Join(util.DataDir, ".scribli", "notebook-crypto-backup.json")
}

func computeBackupChecksum(nc *conf.NotebookCrypto) string {
	tmp := *nc
	tmp.Checksum = ""
	tmp.KEKMAC = nil
	data, _ := json.Marshal(tmp)
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

func computeKEKMAC(nc *conf.NotebookCrypto, kek []byte) []byte {
	tmp := *nc
	tmp.KEKMAC = nil
	data, _ := json.Marshal(tmp)
	mac := hmac.New(sha256.New, kek)
	mac.Write(data)
	return mac.Sum(nil)
}

func verifyKEKMAC(nc *conf.NotebookCrypto, kek []byte) bool {
	if nc == nil || len(nc.KEKMAC) == 0 || len(kek) == 0 {
		return false
	}
	expected := computeKEKMAC(nc, kek)
	return hmac.Equal(expected, nc.KEKMAC)
}

func prepareBackupForWrite(nc *conf.NotebookCrypto) {
	nc.Spec = conf.CurrentNotebookCryptoSpec
	if nc.BackupID == "" {
		nc.BackupID = util.RandString(16)
	}
	nc.CreatedAt = time.Now().Unix()
	nc.Checksum = computeBackupChecksum(nc)
}

func atomicWriteFile(path string, data []byte) error {
	tmpPath := path + "." + gulu.Rand.String(7) + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}

func ExportNotebookCryptoBackup() (downloadPath string, err error) {
	notebookCryptoMu.Lock()
	defer notebookCryptoMu.Unlock()

	backupPath := notebookCryptoBackupPath()
	data, readErr := filelock.ReadFile(backupPath)
	if readErr != nil {
		if os.IsNotExist(readErr) {
			err = errors.New(Conf.Language(315))
			return
		}
		err = readErr
		return
	}
	exportBase := filepath.Join(util.TempDir, "export")
	if mkErr := os.MkdirAll(exportBase, 0755); mkErr != nil {
		err = mkErr
		return
	}

	fileName := "notebook-crypto-backup-" + gulu.Rand.String(7) + ".json"
	downloadPath = "/export/" + url.PathEscape(fileName)
	if writeErr := os.WriteFile(filepath.Join(exportBase, fileName), data, 0644); writeErr != nil {
		err = writeErr
		return
	}
	return
}

func ImportNotebookCryptoBackup(data []byte, password string) error {
	notebookCryptoMu.Lock()
	defer notebookCryptoMu.Unlock()

	Conf.m.RLock()
	enabled := Conf.NotebookCrypto.Enabled
	Conf.m.RUnlock()
	if enabled {
		return errors.New(Conf.Language(324))
	}

	hasHistory, historyErr := scanEncryptedNotebookHistory()
	if historyErr != nil {
		return fmt.Errorf("check encrypted notebook history failed: %w", historyErr)
	}
	if hasHistory {
		return errors.New(Conf.Language(323))
	}

	nc := &conf.NotebookCrypto{}
	if err := json.Unmarshal(data, nc); err != nil {
		return errors.New(Conf.Language(317))
	}
	if len(nc.MasterSalt) == 0 || len(nc.KEKVerifier) == 0 {
		return errors.New(Conf.Language(317))
	}

	params, validErr := util.ValidateArgon2Params(nc.KDFParams)
	if validErr != nil {
		return errors.New(Conf.Language(317))
	}
	if nc.Spec != conf.CurrentNotebookCryptoSpec || nc.Checksum == "" || len(nc.KEKMAC) == 0 {
		return errors.New(Conf.Language(317))
	}
	kek := util.DeriveKey(password, nc.MasterSalt, params)
	defer zeroAndClear(kek)
	if nc.Checksum != computeBackupChecksum(nc) {
		return errors.New(Conf.Language(317))
	}
	if !verifyKEKMAC(nc, kek) {
		return errors.New(Conf.Language(317))
	}
	decrypted, dErr := util.DecryptWithAAD(kek, nc.KEKVerifier, []byte("scribli:v1:kek-verifier"))
	if dErr != nil || string(decrypted) != string(kekVerifierMagic) {
		return errors.New(Conf.Language(311))
	}

	if !verifyKEKAgainstExistingBoxes(kek) {
		return errors.New(Conf.Language(316))
	}

	nc.KDFParams = params
	nc.Enabled = true

	if err := writeNotebookCryptoBackupData(nc, kek); err != nil {
		return fmt.Errorf("failed to persist key backup: %w", err)
	}
	Conf.m.Lock()
	*Conf.NotebookCrypto = *nc
	Conf.m.Unlock()
	Conf.Save()
	return nil
}

func saveNotebookCryptoBackup(kek []byte) error {
	if kek == nil {

		return errors.New("cannot generate notebook crypto backup without KEK")
	}
	Conf.m.Lock()
	nc := *Conf.NotebookCrypto
	prepareBackupForWrite(&nc)
	nc.KEKMAC = computeKEKMAC(&nc, kek)
	Conf.NotebookCrypto.Spec = nc.Spec
	Conf.NotebookCrypto.BackupID = nc.BackupID
	Conf.NotebookCrypto.CreatedAt = nc.CreatedAt
	Conf.NotebookCrypto.Checksum = nc.Checksum
	Conf.NotebookCrypto.KEKMAC = nc.KEKMAC
	Conf.m.Unlock()
	backupPath := notebookCryptoBackupPath()
	if err := os.MkdirAll(filepath.Dir(backupPath), 0755); err != nil {
		return fmt.Errorf("mkdir notebook crypto backup dir failed: %w", err)
	}
	data, err := json.Marshal(nc)
	if err != nil {
		return fmt.Errorf("marshal notebook crypto backup failed: %w", err)
	}
	if err := atomicWriteFile(backupPath, data); err != nil {
		return fmt.Errorf("write notebook crypto backup failed: %w", err)
	}
	return nil
}

func writeNotebookCryptoBackupData(nc *conf.NotebookCrypto, kek []byte) error {
	if kek == nil {
		return errors.New("cannot generate notebook crypto backup without KEK")
	}
	prepareBackupForWrite(nc)
	nc.KEKMAC = computeKEKMAC(nc, kek)
	backupPath := notebookCryptoBackupPath()
	if err := os.MkdirAll(filepath.Dir(backupPath), 0755); err != nil {
		return fmt.Errorf("mkdir notebook crypto backup dir failed: %w", err)
	}
	data, err := json.Marshal(nc)
	if err != nil {
		return fmt.Errorf("marshal notebook crypto backup failed: %w", err)
	}
	if err := atomicWriteFile(backupPath, data); err != nil {
		return fmt.Errorf("write notebook crypto backup failed: %w", err)
	}
	return nil
}

func verifyKEKAgainstExistingBoxes(kek []byte) bool {
	boxIDs, err := listAllEncryptedBoxIDs()
	if err != nil {
		logging.LogErrorf("list encrypted notebooks failed: %s", err)
		return false
	}
	for _, id := range boxIDs {
		boxCrypt, err := GetBoxEncryption(id)
		if err != nil {
			return false
		}
		if boxCrypt == nil || len(boxCrypt.WrappedDEK) == 0 {
			return false
		}
		if _, dErr := decryptWrappedDEK(id, boxCrypt, kek); dErr == nil {
			continue
		}

		backup, bErr := readNotebookCryptBackup(id)
		if bErr == nil && backup != nil && len(backup.WrappedDEK) > 0 &&
			!bytes.Equal(backup.WrappedDEK, boxCrypt.WrappedDEK) {
			if _, err2 := decryptWrappedDEK(id, backup, kek); err2 == nil {
				continue
			}
		}
		return false
	}
	return true
}

func loadNotebookCryptoBackup() (*conf.NotebookCrypto, error) {
	data, err := filelock.ReadFile(notebookCryptoBackupPath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	nc := &conf.NotebookCrypto{}
	if err := json.Unmarshal(data, nc); err != nil {
		return nil, err
	}
	conf.UpgradeSpec(nc)
	if nc.Spec != conf.CurrentNotebookCryptoSpec {
		return nil, fmt.Errorf("unsupported notebook crypto backup spec [%d]", nc.Spec)
	}
	if nc.Checksum == "" {
		return nil, errors.New("notebook crypto backup checksum is missing")
	}
	expected := computeBackupChecksum(nc)
	if nc.Checksum != expected {
		logging.LogWarnf("notebook crypto backup checksum mismatch: expected %s, got %s", expected, nc.Checksum)
		return nil, errors.New("notebook crypto backup is corrupted (checksum mismatch)")
	}
	return nc, nil
}

func removeNotebookCryptoBackup() {
	if err := os.Remove(notebookCryptoBackupPath()); err != nil && !os.IsNotExist(err) {
		logging.LogErrorf("remove notebook crypto backup failed: %s", err)
	}
}

type masterPasswordMigration struct {
	OldVerifier      []byte              `json:"oldVerifier"`
	NewVerifier      []byte              `json:"newVerifier"`
	NewVerifierNonce []byte              `json:"newVerifierNonce"`
	NewKDFParams     json.RawMessage     `json:"newKDFParams"`
	Boxes            []migrationBoxEntry `json:"boxes"`
}

type migrationBoxEntry struct {
	BoxID         string `json:"boxID"`
	NewSpec       int    `json:"newSpec"`
	NewWrappedDEK []byte `json:"newWrappedDEK"`
	NewWrapNonce  []byte `json:"newWrapNonce"`
}

func masterPasswordMigrationPath() string {
	return filepath.Join(util.DataDir, ".scribli", "master-password-migration.json")
}

func writeMasterPasswordMigration(m *masterPasswordMigration) error {
	p := masterPasswordMigrationPath()
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		return fmt.Errorf("mkdir master password migration dir failed: %w", err)
	}
	data, err := gulu.JSON.MarshalIndentJSON(m, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal master password migration failed: %w", err)
	}
	return filelock.WriteFile(p, data)
}

func readMasterPasswordMigration() (*masterPasswordMigration, error) {
	p := masterPasswordMigrationPath()
	if !filelock.IsExist(p) {
		return nil, nil
	}
	data, err := filelock.ReadFile(p)
	if err != nil {
		return nil, fmt.Errorf("read master password migration failed: %w", err)
	}
	var m masterPasswordMigration
	if err = gulu.JSON.UnmarshalJSON(data, &m); err != nil {
		return nil, fmt.Errorf("unmarshal master password migration failed: %w", err)
	}
	return &m, nil
}

func removeMasterPasswordMigration() {
	p := masterPasswordMigrationPath()
	if err := filelock.Remove(p); err != nil && !os.IsNotExist(err) {
		logging.LogErrorf("remove master password migration failed: %s", err)
	}
}

func MasterPasswordMigrationStatus() (pending bool, boxIDs []string) {
	mig, err := readMasterPasswordMigration()
	if err != nil || mig == nil {
		return false, nil
	}
	for _, entry := range mig.Boxes {
		boxIDs = append(boxIDs, entry.BoxID)
	}
	return true, boxIDs
}

func recoverMasterPasswordMigration() {
	mig, err := readMasterPasswordMigration()
	if err != nil {
		logging.LogErrorf("read master password migration failed: %s", err)
		return
	}
	if mig == nil {
		return
	}

	Conf.m.RLock()
	currentVerifier := Conf.NotebookCrypto.KEKVerifier
	Conf.m.RUnlock()

	if bytes.Equal(currentVerifier, mig.NewVerifier) {

		for _, entry := range mig.Boxes {
			box := &Box{ID: entry.BoxID}
			boxConf := box.GetConf()
			if !boxConf.Encrypted || boxConf.BoxCrypt == nil {

				backup, bErr := readNotebookCryptBackup(entry.BoxID)
				if bErr == nil && backup != nil && len(backup.WrappedDEK) > 0 {
					boxConf = box.GetConf()
					boxConf.Encrypted = true
					boxConf.BoxCrypt = backup
					if saveErr := box.SaveConf(boxConf); saveErr != nil {
						logging.LogErrorf("rebuild encrypted conf from backup [%s] failed: %s", entry.BoxID, saveErr)
						return
					}
				} else {

					logging.LogWarnf("rebuild encrypted box [%s] from migration manifest (conf and backup both unavailable)", entry.BoxID)
					boxConf = box.GetConf()
					boxConf.Encrypted = true
					boxConf.BoxCrypt = &conf.BoxEncryption{
						WrappedDEK: entry.NewWrappedDEK,
						WrapNonce:  entry.NewWrapNonce,
						Spec:       entry.NewSpec,
						CreatedAt:  time.Now().UnixMilli(),
					}
					if saveErr := box.SaveConf(boxConf); saveErr != nil {
						logging.LogErrorf("rebuild encrypted conf from manifest [%s] failed: %s", entry.BoxID, saveErr)
						return
					}
				}
			}

			if bytes.Equal(boxConf.BoxCrypt.WrappedDEK, entry.NewWrappedDEK) {
				if writeErr := writeNotebookCryptBackup(entry.BoxID, boxConf.BoxCrypt); writeErr != nil {
					logging.LogErrorf("refresh box crypt backup [%s] failed: %s", entry.BoxID, writeErr)
					return
				}
				continue
			}
			boxConf.BoxCrypt.WrappedDEK = entry.NewWrappedDEK
			boxConf.BoxCrypt.Spec = entry.NewSpec
			boxConf.BoxCrypt.WrapNonce = entry.NewWrapNonce
			if saveErr := box.SaveConf(boxConf); saveErr != nil {
				logging.LogErrorf("recover box conf [%s] failed: %s", entry.BoxID, saveErr)
				return
			}
			if writeErr := writeNotebookCryptBackup(entry.BoxID, boxConf.BoxCrypt); writeErr != nil {
				logging.LogErrorf("recover box crypt backup [%s] failed: %s", entry.BoxID, writeErr)
				return
			}
		}

		Conf.Save()
		logging.LogInfof("master password migration data recovered, waiting for the new password to authenticate the backup")
	} else {

		removeMasterPasswordMigration()
		logging.LogErrorf("master password migration was interrupted, please retry")
	}
}

func hasEncryptedNotebook() (bool, error) {
	ids, err := listAllEncryptedBoxIDs()
	return len(ids) > 0, err
}

//

func scanEncryptedNotebookHistory() (bool, error) {
	entries, err := os.ReadDir(util.HistoryDir)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("read history dir failed: %w", err)
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		snapshotDir := filepath.Join(util.HistoryDir, entry.Name())
		boxEntries, readErr := os.ReadDir(snapshotDir)
		if readErr != nil {
			return false, fmt.Errorf("read history snapshot [%s] failed: %w", entry.Name(), readErr)
		}
		for _, boxEntry := range boxEntries {
			if !boxEntry.IsDir() || !ast.IsNodeIDPattern(boxEntry.Name()) {
				continue
			}
			encrypted, checkErr := isEncryptedHistoryBoxDir(filepath.Join(snapshotDir, boxEntry.Name()))
			if checkErr != nil {
				return false, checkErr
			}
			if encrypted {
				return true, nil
			}
		}
	}
	return false, nil
}

func HasEncryptedNotebookHistory() bool {
	hasHistory, err := scanEncryptedNotebookHistory()
	if err != nil {
		logging.LogErrorf("scan encrypted notebook history failed: %s", err)
		return true
	}
	return hasHistory
}

func isEncryptedHistoryBoxDir(boxDir string) (bool, error) {
	scribliDir := filepath.Join(boxDir, ".scribli")
	backupPath := filepath.Join(scribliDir, "notebook-crypt-backup.json")
	if _, err := os.Stat(backupPath); err == nil {
		return true, nil
	} else if !os.IsNotExist(err) {
		return false, fmt.Errorf("stat encrypted notebook history backup [%s] failed: %w", boxDir, err)
	}
	confPath := filepath.Join(scribliDir, "conf.json")
	if _, err := os.Stat(confPath); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("stat encrypted notebook history conf [%s] failed: %w", boxDir, err)
	}
	data, err := filelock.ReadFile(confPath)
	if err != nil {
		return false, fmt.Errorf("read encrypted notebook history conf [%s] failed: %w", boxDir, err)
	}
	var boxConf conf.BoxConf
	if err = gulu.JSON.UnmarshalJSON(data, &boxConf); err != nil {
		return false, fmt.Errorf("parse encrypted notebook history conf [%s] failed: %w", boxDir, err)
	}
	return boxConf.Encrypted, nil
}

var (
	cachedDEKs     = map[string][]byte{}
	cachedDEKsLock sync.RWMutex
)

var boxLastAccess sync.Map

func EnableEncryptedNotebook(password string) error {
	if len(password) == 0 {
		return errors.New("password must not be empty")
	}

	notebookCryptoMu.Lock()
	defer notebookCryptoMu.Unlock()

	Conf.m.Lock()
	if Conf.NotebookCrypto.Enabled {
		Conf.m.Unlock()
		return errors.New(Conf.Language(312))
	}
	Conf.m.Unlock()

	hasEncrypted, listErr := hasEncryptedNotebook()
	if listErr != nil {
		return fmt.Errorf("list encrypted notebooks failed: %w", listErr)
	}
	if hasEncrypted {
		if _, restoreErr := tryRestoreNotebookCryptoFromBackupLocked(password); restoreErr != nil {

			if strings.Contains(restoreErr.Error(), Conf.Language(311)) {
				return errors.New(Conf.Language(311))
			}
			return errors.New(Conf.Language(315))
		}
		logging.LogInfof("encrypted notebook re-enabled with restored master key material from backup")
		return nil
	}

	hasHistory, historyErr := scanEncryptedNotebookHistory()
	if historyErr != nil {
		return fmt.Errorf("check encrypted notebook history failed: %w", historyErr)
	}
	if hasHistory {
		return errors.New(Conf.Language(323))
	}

	salt, err := util.GenerateSalt()
	if err != nil {
		return err
	}
	Conf.m.RLock()
	kdfParams := Conf.NotebookCrypto.KDFParams
	Conf.m.RUnlock()
	params, validErr := util.ValidateArgon2Params(kdfParams)
	if validErr != nil {
		return validErr
	}
	kek := util.DeriveKey(password, salt, params)
	defer zeroAndClear(kek)

	verifierCT, err := util.EncryptWithAAD(kek, kekVerifierMagic, []byte("scribli:v1:kek-verifier"))
	if err != nil {
		return err
	}
	verifierNonce, nonceErr := util.EncryptionNonce(verifierCT)
	if nonceErr != nil {
		return nonceErr
	}

	Conf.m.Lock()
	previous := *Conf.NotebookCrypto
	Conf.NotebookCrypto.Enabled = true
	Conf.NotebookCrypto.MasterSalt = salt
	Conf.NotebookCrypto.KDFParams = params
	Conf.NotebookCrypto.KEKVerifier = verifierCT
	Conf.NotebookCrypto.VerifierNonce = verifierNonce
	Conf.m.Unlock()

	if err := saveNotebookCryptoBackup(kek); err != nil {

		logging.LogErrorf("save notebook crypto backup failed: %s", err)
		Conf.m.Lock()
		*Conf.NotebookCrypto = previous
		Conf.m.Unlock()
		return fmt.Errorf("enable encrypted notebook failed: failed to persist key backup: %w", err)
	}

	Conf.Save()
	return nil
}

func DisableEncryptedNotebook() error {
	notebookCryptoMu.Lock()
	defer notebookCryptoMu.Unlock()

	ids, listErr := listAllEncryptedBoxIDs()
	if listErr != nil {
		return fmt.Errorf("list encrypted notebooks failed: %w", listErr)
	}
	if len(ids) > 0 {
		return errors.New("cannot disable encrypted notebook feature while encrypted notebooks exist, remove them first")
	}

	hasHistory, historyErr := scanEncryptedNotebookHistory()
	if historyErr != nil {
		return fmt.Errorf("check encrypted notebook history failed: %w", historyErr)
	}
	if hasHistory {
		return errors.New(Conf.Language(323))
	}

	Conf.m.Lock()
	Conf.NotebookCrypto.Enabled = false
	Conf.NotebookCrypto.MasterSalt = nil
	Conf.NotebookCrypto.KEKVerifier = nil
	Conf.NotebookCrypto.VerifierNonce = nil
	Conf.m.Unlock()

	Conf.Save()
	removeNotebookCryptoBackup()
	return nil
}

func restoreNotebookCryptoConfigFromBackup() {
	notebookCryptoMu.Lock()
	defer notebookCryptoMu.Unlock()

	Conf.m.RLock()
	enabled := Conf.NotebookCrypto.Enabled
	Conf.m.RUnlock()
	if enabled {
		return
	}
	backup, err := loadNotebookCryptoBackup()
	if err != nil || backup == nil || len(backup.MasterSalt) == 0 || len(backup.KEKVerifier) == 0 {
		return
	}

	params, validErr := util.ValidateArgon2Params(backup.KDFParams)
	if validErr != nil {
		logging.LogErrorf("skip restore notebook crypto: invalid KDFParams in backup: %s", validErr)
		return
	}
	backup.KDFParams = params

	backup.Enabled = true
	Conf.m.Lock()
	*Conf.NotebookCrypto = *backup
	Conf.m.Unlock()
	Conf.Save()
	logging.LogInfof("notebook crypto config restored from backup (auto-enable after sync/import)")
}

func tryRestoreNotebookCryptoFromBackupLocked(password string) (kek []byte, err error) {
	backup, bErr := loadNotebookCryptoBackup()
	if bErr != nil || backup == nil || len(backup.MasterSalt) == 0 || len(backup.KEKVerifier) == 0 {

		return nil, errors.New(Conf.Language(310))
	}
	params, validErr := util.ValidateArgon2Params(backup.KDFParams)
	if validErr != nil {
		return nil, errors.New(Conf.Language(317))
	}
	kek = util.DeriveKey(password, backup.MasterSalt, params)
	decrypted, dErr := util.DecryptWithAAD(kek, backup.KEKVerifier, []byte("scribli:v1:kek-verifier"))
	if dErr != nil || string(decrypted) != string(kekVerifierMagic) {

		zeroAndClear(kek)
		return nil, errors.New(Conf.Language(311))
	}

	if backup.Spec != conf.CurrentNotebookCryptoSpec || backup.Checksum == "" ||
		len(backup.KEKMAC) == 0 || !verifyKEKMAC(backup, kek) {
		zeroAndClear(kek)
		return nil, errors.New(Conf.Language(316))
	}

	if !verifyKEKAgainstExistingBoxes(kek) {
		zeroAndClear(kek)
		return nil, errors.New(Conf.Language(316))
	}

	backup.KDFParams = params
	backup.Enabled = true
	Conf.m.Lock()
	*Conf.NotebookCrypto = *backup
	Conf.m.Unlock()
	Conf.Save()

	nc := *backup
	if err := writeNotebookCryptoBackupData(&nc, kek); err != nil {
		logging.LogWarnf("rewrite notebook crypto backup after restore failed: %s", err)
	}
	logging.LogInfof("notebook crypto restored from backup (e.g. after sync to a new device)")
	return kek, nil
}

func deriveKEK(password string) ([]byte, error) {
	Conf.m.RLock()
	nc := *Conf.NotebookCrypto
	Conf.m.RUnlock()

	if !nc.Enabled {

		kek, restoreErr := tryRestoreNotebookCryptoFromBackupLocked(password)
		if restoreErr != nil {
			return nil, restoreErr
		}
		return kek, nil
	}
	params, validErr := util.ValidateArgon2Params(nc.KDFParams)
	if validErr != nil {
		return nil, validErr
	}
	kek := util.DeriveKey(password, nc.MasterSalt, params)

	decrypted, err := util.DecryptWithAAD(kek, nc.KEKVerifier, []byte("scribli:v1:kek-verifier"))
	if err != nil {
		zeroAndClear(kek)
		return nil, errors.New(Conf.Language(311))
	}
	if string(decrypted) != string(kekVerifierMagic) {
		zeroAndClear(kek)
		return nil, errors.New(Conf.Language(311))
	}

	mig, migErr := readMasterPasswordMigration()
	migrationPending := migErr == nil && mig != nil && bytes.Equal(nc.KEKVerifier, mig.NewVerifier)
	if !migrationPending {
		backup, backupErr := loadNotebookCryptoBackup()
		backupMatchesConf := backupErr == nil && backup != nil &&
			bytes.Equal(backup.MasterSalt, nc.MasterSalt) &&
			bytes.Equal(backup.KEKVerifier, nc.KEKVerifier) &&
			backup.KDFParams == nc.KDFParams
		if !backupMatchesConf || backup.Spec != conf.CurrentNotebookCryptoSpec || backup.Checksum == "" ||
			len(backup.KEKMAC) == 0 || !verifyKEKMAC(backup, kek) {

			zeroAndClear(kek)
			return nil, errors.New(Conf.Language(315))
		}
	}

	if migrationPending {

		if !verifyKEKAgainstExistingBoxes(kek) {
			zeroAndClear(kek)
			return nil, errMasterPasswordMigrationPending
		}
		if err = saveNotebookCryptoBackup(kek); err != nil {
			zeroAndClear(kek)
			return nil, fmt.Errorf("%w: %v", errMasterPasswordMigrationPending, err)
		}
		removeMasterPasswordMigration()
	}
	return kek, nil
}

func decryptBoxCrypt(boxID string, kek []byte) (dek []byte, boxCrypt *conf.BoxEncryption, err error) {
	boxCrypt, err = GetBoxEncryption(boxID)
	if err != nil || boxCrypt == nil || len(boxCrypt.WrappedDEK) == 0 {
		return nil, nil, fmt.Errorf("no encrypted key material for box [%s]", boxID)
	}

	dek, err = decryptWrappedDEK(boxID, boxCrypt, kek)
	if err == nil {
		return dek, boxCrypt, nil
	}

	backup, bErr := readNotebookCryptBackup(boxID)
	if bErr == nil && backup != nil && len(backup.WrappedDEK) > 0 &&
		!bytes.Equal(backup.WrappedDEK, boxCrypt.WrappedDEK) {
		dek, err = decryptWrappedDEK(boxID, backup, kek)
		if err == nil {

			box := &Box{ID: boxID}
			boxConf := box.GetConf()
			boxConf.Encrypted = true
			boxConf.BoxCrypt = backup
			if saveErr := box.SaveConf(boxConf); saveErr != nil {
				logging.LogWarnf("fix encrypted box conf from backup [%s] failed: %s", boxID, saveErr)
			}
			if needWriteNotebookCryptBackup(boxID, backup) {
				if writeErr := writeNotebookCryptBackup(boxID, backup); writeErr != nil {
					logging.LogWarnf("refresh notebook crypt backup [%s] failed: %s", boxID, writeErr)
				}
			}
			return dek, backup, nil
		}
	}
	return nil, nil, fmt.Errorf("decrypt box [%s] failed: incorrect key or corrupted data", boxID)
}

func UnlockBox(boxID string, password string, boxEnc *conf.BoxEncryption) error {
	if boxEnc == nil || len(boxEnc.WrappedDEK) == 0 {
		return errors.New("no encrypted key material for box")
	}

	notebookCryptoMu.Lock()
	defer notebookCryptoMu.Unlock()

	acquireBoxWriteLock(boxID)
	defer releaseBoxWriteLock(boxID)

	kek, err := deriveKEK(password)
	if err != nil {
		return err
	}
	defer zeroAndClear(kek)

	dek, trustedCrypt, err := decryptBoxCrypt(boxID, kek)
	if err != nil {
		return errors.New(Conf.Language(316))
	}
	boxEnc = trustedCrypt

	cachedDEKsLock.Lock()
	defer cachedDEKsLock.Unlock()
	if err = sql.OpenEncryptedDB(boxID, dek); err != nil {
		return err
	}
	if err = treenode.OpenEncryptedBlockTreeDB(boxID, dek); err != nil {
		sql.RemoveEncryptedDBFile(boxID)
		return err
	}
	cachedDEKs[boxID] = dek

	newVal := &atomic.Int64{}
	newVal.Store(time.Now().UnixNano())
	boxLastAccess.Store(boxID, newVal)

	box := &Box{ID: boxID}
	boxConf := box.GetConf()
	if boxConf == nil || !boxConf.Encrypted || boxConf.BoxCrypt == nil ||
		len(boxConf.BoxCrypt.WrappedDEK) == 0 ||
		!bytes.Equal(boxConf.BoxCrypt.WrappedDEK, boxEnc.WrappedDEK) {
		boxConf.Encrypted = true
		boxConf.BoxCrypt = boxEnc
		if saveErr := box.SaveConf(boxConf); saveErr != nil {
			logging.LogWarnf("fix encrypted box conf [%s] failed: %s", boxID, saveErr)
		}
	}

	if needWriteNotebookCryptBackup(boxID, boxEnc) {
		if err = writeNotebookCryptBackup(boxID, boxEnc); err != nil {
			logging.LogWarnf("write notebook crypt backup [%s] failed: %s", boxID, err)
		}
	}
	return nil
}

func IsBoxUnlocked(boxID string) bool {
	cachedDEKsLock.RLock()
	defer cachedDEKsLock.RUnlock()
	_, ok := cachedDEKs[boxID]
	return ok
}

func LockBox(boxID string) {
	FlushTxQueue()
	acquireBoxWriteLock(boxID)
	lockBoxHeld(boxID)
	releaseBoxWriteLock(boxID)

	cache.ClearTreeCache()
	sql.ClearCache()
	cache.ClearDocsIAL()
	cache.ClearBlocksIAL()
	cache.ClearAVCache()
	ResetVirtualBlockRefCache()
}

func lockBoxHeld(boxID string) {
	RevokeManagedEncryptedExportsForBox(boxID)

	cachedDEKsLock.Lock()
	if dek, ok := cachedDEKs[boxID]; ok {
		zeroAndClear(dek)
		delete(cachedDEKs, boxID)
	}
	cachedDEKsLock.Unlock()

	boxLastAccess.Delete(boxID)

	if !filelock.IsExist(notebookCryptBackupPath(boxID)) {
		box := &Box{ID: boxID}
		boxConf := box.GetConf()
		if boxConf != nil && boxConf.Encrypted && boxConf.BoxCrypt != nil && len(boxConf.BoxCrypt.WrappedDEK) > 0 {
			if err := writeNotebookCryptBackup(boxID, boxConf.BoxCrypt); err != nil {
				logging.LogWarnf("write notebook crypt backup [%s] failed: %s", boxID, err)
			}
		}
	}

	sql.RemoveEncryptedDBFile(boxID)
	treenode.RemoveEncryptedBlockTreeDBFile(boxID)

	repoDirs := []string{
		filepath.Join(util.TempDir, "repo", "diff", boxID),
		filepath.Join(util.TempDir, "repo", "rollback", boxID),
	}
	for _, d := range repoDirs {
		if rmErr := os.RemoveAll(d); rmErr != nil {
			logging.LogWarnf("remove repo dir for box [%s] failed: %s", boxID, rmErr)
		}
	}

	if matches, globErr := filepath.Glob(filepath.Join(util.TempDir, "repo", "sync", "conflicts", "*", boxID)); globErr == nil {
		for _, m := range matches {
			if rmErr := os.RemoveAll(m); rmErr != nil {
				logging.LogWarnf("remove repo sync conflict dir for box [%s] failed: %s", boxID, rmErr)
			}
		}
	}

	if rmErr := os.RemoveAll(filepath.Join(util.TempDir, "export", boxID)); rmErr != nil {
		logging.LogWarnf("remove export/[%s] dir failed: %s", boxID, rmErr)
	}

	treenode.RemoveDynamicRefTexts(boxID)
}

func WrapNewDEK(boxID string, kek []byte) (*conf.BoxEncryption, []byte, error) {
	dek, err := util.GenerateDEK()
	if err != nil {
		return nil, nil, err
	}
	wrapped, err := util.EncryptWithAAD(kek, dek, wrappedDEKAAD(boxID))
	if err != nil {
		return nil, nil, err
	}
	return &conf.BoxEncryption{
		Spec:       boxEncryptionSpec,
		WrappedDEK: wrapped,
		WrapNonce:  mustEncryptionNonce(wrapped),
		CreatedAt:  time.Now().UnixMilli(),
	}, dek, nil
}

func wrappedDEKAAD(boxID string) []byte {
	return []byte("scribli:v1:wrapped-dek:" + boxID)
}

func decryptWrappedDEK(boxID string, enc *conf.BoxEncryption, kek []byte) ([]byte, error) {
	if enc.Spec >= boxEncryptionSpec {
		return util.DecryptWithAAD(kek, enc.WrappedDEK, wrappedDEKAAD(boxID))
	}
	return util.DecryptWithAAD(kek, enc.WrappedDEK, wrappedDEKAAD(boxID))
}

func mustEncryptionNonce(ciphertext []byte) []byte {
	nonce, err := util.EncryptionNonce(ciphertext)
	if err != nil {
		panic("extract encryption nonce failed: " + err.Error())
	}
	return nonce
}

func GetDEK(boxID string) ([]byte, error) {
	cachedDEKsLock.RLock()
	defer cachedDEKsLock.RUnlock()
	dek, ok := cachedDEKs[boxID]
	if !ok {
		return nil, errors.New("no DEK cached for box " + boxID)
	}
	ret := make([]byte, len(dek))
	copy(ret, dek)
	return ret, nil
}

func ClearDEK(boxID string) {
	LockBox(boxID)
}

//

//

//

func ChangeMasterPassword(oldPassword, newPassword string) error {
	if len(newPassword) == 0 {
		return errors.New("new password must not be empty")
	}

	notebookCryptoMu.Lock()
	defer notebookCryptoMu.Unlock()

	cachedDEKsLock.RLock()
	dekCount := len(cachedDEKs)
	cachedDEKsLock.RUnlock()
	if dekCount > 0 {
		return errors.New("cannot change master password while encrypted notebooks are unlocked (DEKs in memory), lock them first")
	}

	oldKEK, err := deriveKEK(oldPassword)
	if err != nil {
		return err
	}
	defer zeroAndClear(oldKEK)

	Conf.m.Lock()
	nc := Conf.NotebookCrypto
	Conf.m.Unlock()

	params, validErr := util.ValidateArgon2Params(nc.KDFParams)
	if validErr != nil {
		return validErr
	}
	newKEK := util.DeriveKey(newPassword, nc.MasterSalt, params)
	defer zeroAndClear(newKEK)
	newVerifier, err := util.EncryptWithAAD(newKEK, kekVerifierMagic, []byte("scribli:v1:kek-verifier"))
	if err != nil {
		return err
	}

	encBoxIDs, listErr := listAllEncryptedBoxIDs()
	if listErr != nil {
		return fmt.Errorf("list encrypted notebooks failed: %w", listErr)
	}
	var entries []migrationBoxEntry
	for _, id := range encBoxIDs {
		dek, _, dErr := decryptBoxCrypt(id, oldKEK)
		if dErr != nil {
			return errors.New(Conf.Language(316) + " [box=" + id + "]")
		}
		newWrapped, nErr := util.EncryptWithAAD(newKEK, dek, wrappedDEKAAD(id))
		if nErr != nil {
			return nErr
		}
		entries = append(entries, migrationBoxEntry{
			BoxID:         id,
			NewSpec:       boxEncryptionSpec,
			NewWrappedDEK: newWrapped,
			NewWrapNonce:  mustEncryptionNonce(newWrapped),
		})
	}

	newParamsJSON, _ := gulu.JSON.MarshalJSON(params)
	mig := &masterPasswordMigration{
		OldVerifier:      nc.KEKVerifier,
		NewVerifier:      newVerifier,
		NewVerifierNonce: mustEncryptionNonce(newVerifier),
		NewKDFParams:     newParamsJSON,
		Boxes:            entries,
	}
	if err = writeMasterPasswordMigration(mig); err != nil {
		return err
	}

	Conf.m.Lock()
	Conf.NotebookCrypto.KEKVerifier = newVerifier
	Conf.NotebookCrypto.VerifierNonce = mustEncryptionNonce(newVerifier)
	Conf.NotebookCrypto.KDFParams = params
	Conf.m.Unlock()

	Conf.Save()

	for _, entry := range entries {
		box := &Box{ID: entry.BoxID}
		boxConf := box.GetConf()
		if !boxConf.Encrypted || boxConf.BoxCrypt == nil {

			backup, bErr := readNotebookCryptBackup(entry.BoxID)
			if bErr == nil && backup != nil && len(backup.WrappedDEK) > 0 {
				boxConf = box.GetConf()
				boxConf.Encrypted = true
				boxConf.BoxCrypt = backup
				if saveErr := box.SaveConf(boxConf); saveErr != nil {
					return fmt.Errorf("%w: %s", errMasterPasswordMigrationPending,
						fmt.Sprintf(Conf.Language(320), entry.BoxID+": rebuild encrypted conf from backup failed: "+saveErr.Error()))
				}
			} else {

				logging.LogWarnf("rebuild encrypted box [%s] from migration entry (conf and backup both unavailable)", entry.BoxID)
				boxConf = box.GetConf()
				boxConf.Encrypted = true
				boxConf.BoxCrypt = &conf.BoxEncryption{
					WrappedDEK: entry.NewWrappedDEK,
					WrapNonce:  entry.NewWrapNonce,
					Spec:       entry.NewSpec,
					CreatedAt:  time.Now().UnixMilli(),
				}
				if saveErr := box.SaveConf(boxConf); saveErr != nil {
					return fmt.Errorf("%w: %s", errMasterPasswordMigrationPending,
						fmt.Sprintf(Conf.Language(320), entry.BoxID+": rebuild encrypted conf from migration entry failed: "+saveErr.Error()))
				}
			}
		}
		boxConf.BoxCrypt.WrappedDEK = entry.NewWrappedDEK
		boxConf.BoxCrypt.Spec = entry.NewSpec
		boxConf.BoxCrypt.WrapNonce = entry.NewWrapNonce
		if err = box.SaveConf(boxConf); err != nil {
			return fmt.Errorf("%w: %s", errMasterPasswordMigrationPending,
				fmt.Sprintf(Conf.Language(320), entry.BoxID+": save conf failed: "+err.Error()))
		}
		if err = writeNotebookCryptBackup(entry.BoxID, boxConf.BoxCrypt); err != nil {
			return fmt.Errorf("%w: %s", errMasterPasswordMigrationPending,
				fmt.Sprintf(Conf.Language(320), entry.BoxID+": update notebook crypt backup failed: "+err.Error()))
		}
	}

	if err = saveNotebookCryptoBackup(newKEK); err != nil {
		return fmt.Errorf("%w: %s", errMasterPasswordMigrationPending,
			fmt.Sprintf(Conf.Language(320), "save notebook crypto backup failed: "+err.Error()))
	}
	removeMasterPasswordMigration()
	return nil
}

func IsEncryptedBox(boxID string) bool {
	box := &Box{ID: boxID}
	boxConf := box.GetConf()
	if boxConf != nil && boxConf.Encrypted {
		return true
	}

	backupPath := notebookCryptBackupPath(boxID)
	if !filelock.IsExist(backupPath) {
		return false
	}
	backup, err := readNotebookCryptBackup(boxID)
	if err != nil {
		logging.LogWarnf("failed to read notebook crypt backup for [%s]: %s", boxID, err)
		return true
	}
	return backup != nil && len(backup.WrappedDEK) > 0
}

func GetBoxEncryption(boxID string) (*conf.BoxEncryption, error) {
	box := &Box{ID: boxID}
	boxConf := box.GetConf()
	confMarkedEncrypted := boxConf != nil && boxConf.Encrypted

	if confMarkedEncrypted && boxConf.BoxCrypt != nil && len(boxConf.BoxCrypt.WrappedDEK) > 0 {
		return boxConf.BoxCrypt, nil
	}

	backup, err := readNotebookCryptBackup(boxID)
	if err != nil {
		return nil, err
	}
	if backup != nil && len(backup.WrappedDEK) > 0 {
		return backup, nil
	}

	if confMarkedEncrypted {

		return nil, errors.New("encrypted notebook has no valid key material")
	}
	return nil, nil
}

func needWriteNotebookCryptBackup(boxID string, crypt *conf.BoxEncryption) bool {
	existing, err := readNotebookCryptBackup(boxID)
	if err != nil || existing == nil {
		return true
	}
	return !bytes.Equal(existing.WrappedDEK, crypt.WrappedDEK) ||
		!bytes.Equal(existing.WrapNonce, crypt.WrapNonce) ||
		existing.CreatedAt != crypt.CreatedAt
}

func DeepCopyBoxEncryption(src *conf.BoxEncryption) *conf.BoxEncryption {
	if src == nil {
		return nil
	}
	return &conf.BoxEncryption{
		Spec:       src.Spec,
		WrappedDEK: append([]byte(nil), src.WrappedDEK...),
		WrapNonce:  append([]byte(nil), src.WrapNonce...),
		CreatedAt:  src.CreatedAt,
	}
}

func listAllEncryptedBoxIDs() ([]string, error) {
	var ids []string
	seen := map[string]bool{}

	boxes, err := ListNotebooks()
	if err != nil {
		return nil, err
	}
	for _, b := range boxes {
		seen[b.ID] = true
		if IsEncryptedBox(b.ID) {
			ids = append(ids, b.ID)
		}
	}

	dirs, err := os.ReadDir(util.DataDir)
	if err != nil {
		return nil, err
	}
	for _, dir := range dirs {
		if !dir.IsDir() || !ast.IsNodeIDPattern(dir.Name()) || seen[dir.Name()] {
			continue
		}
		if IsEncryptedBox(dir.Name()) {
			ids = append(ids, dir.Name())
		}
	}
	return ids, nil
}

func ListAllEncryptedBoxIDs() []string {
	ids, err := listAllEncryptedBoxIDs()
	if err != nil {
		logging.LogErrorf("list encrypted notebooks failed: %s", err)
		return nil
	}
	return ids
}

func IsSameCryptoBoundary(srcBox, dstBox string) bool {
	srcEnc := IsEncryptedBox(srcBox)
	dstEnc := IsEncryptedBox(dstBox)
	if !srcEnc && !dstEnc {
		return true
	}
	return srcEnc && dstEnc && srcBox == dstBox
}

func IsBlockRefCrossingBoundary(srcBoxID, defBlockID string) bool {
	if "" == defBlockID {
		return false
	}
	if IsEncryptedBox(srcBoxID) {

		bt := treenode.GetBlockTreeInBox(defBlockID, srcBoxID)
		return nil == bt || bt.BoxID != srcBoxID
	}

	bt := treenode.GetBlockTree(defBlockID)
	if nil == bt {

		for _, encBoxID := range treenode.GetOpenedEncryptedBoxIDs() {
			if encBT := treenode.GetBlockTreeInBox(defBlockID, encBoxID); nil != encBT {
				bt = encBT
				break
			}
		}
	}
	if nil == bt {

		return normalBoxBlockRefCrossesBoundary(nil)
	}
	return normalBoxBlockRefCrossesBoundary(bt)
}

func normalBoxBlockRefCrossesBoundary(bt *treenode.BlockTree) bool {
	return bt == nil || IsEncryptedBox(bt.BoxID)
}

func IsEncryptedAssetPath(absPath string) bool {
	boxID := ExtractBoxIDFromAssetsPath(absPath)
	return boxID != "" && IsEncryptedBox(boxID)
}

func GetDEKIfUnlocked(boxID string) ([]byte, error) {
	if !IsEncryptedBox(boxID) {
		return nil, nil
	}
	cachedDEKsLock.RLock()
	defer cachedDEKsLock.RUnlock()
	dek, ok := cachedDEKs[boxID]
	if !ok {
		return nil, errors.New("encrypted notebook is locked, please unlock it first")
	}
	ret := make([]byte, len(dek))
	copy(ret, dek)
	return ret, nil
}

func HoldBoxReadLock(boxID string) {
	acquireBoxReadLock(boxID)
}

func ReleaseBoxReadLock(boxID string) {
	releaseBoxReadLock(boxID)
}

func extractBoxIDFromPath(absPath string) string {
	return ExtractBoxIDFromAssetsPath(absPath)
}

func ExtractBoxIDFromAssetsPath(absPath string) string {
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

func ExtractBoxIDFromHistoryPath(absPath string) string {
	absPath = filepath.ToSlash(absPath)
	historyDir := filepath.ToSlash(util.HistoryDir)
	rel, err := filepath.Rel(historyDir, absPath)
	if err != nil {
		return ""
	}
	rel = filepath.ToSlash(rel)
	if strings.HasPrefix(rel, "..") || rel == "." || rel == "" {
		return ""
	}
	parts := strings.SplitN(rel, "/", 3)
	if len(parts) < 2 {
		return ""
	}
	// parts[0] = timestamp-op, parts[1] = boxID
	boxID := parts[1]
	if !ast.IsNodeIDPattern(boxID) {
		return ""
	}
	return boxID
}

func EncryptFile(boxID, relativePath string, dek, plaintext []byte) ([]byte, error) {
	fileKey := util.DeriveSubKey(dek, "scribli/file")
	aad, err := filesys.SyAAD(boxID, relativePath)
	if err != nil {
		return nil, err
	}
	return util.EncryptWithAAD(fileKey, plaintext, []byte(aad))
}

func DecryptFile(boxID, relativePath string, dek, ciphertext []byte) ([]byte, error) {
	fileKey := util.DeriveSubKey(dek, "scribli/file")
	aad, err := filesys.SyAAD(boxID, relativePath)
	if err != nil {
		return nil, err
	}
	return util.DecryptWithAAD(fileKey, ciphertext, []byte(aad))
}

func EncryptAsset(boxID, diskName string, dek, plaintext []byte) ([]byte, error) {
	assetKey := util.DeriveSubKey(dek, "siyuan/asset")
	aad := "scribli:v1:asset:" + boxID + ":assets/" + diskName
	return util.EncryptWithAAD(assetKey, plaintext, []byte(aad))
}

func DecryptAsset(boxID, diskName string, dek, ciphertext []byte) ([]byte, error) {
	assetKey := util.DeriveSubKey(dek, "siyuan/asset")
	aad := "scribli:v1:asset:" + boxID + ":assets/" + diskName
	return util.DecryptWithAAD(assetKey, ciphertext, []byte(aad))
}

func EncryptAssetNameMapping(boxID string, dek, plaintext []byte) ([]byte, error) {
	assetKey := util.DeriveSubKey(dek, "siyuan/asset")
	aad := "scribli:v1:asset-names:" + boxID
	return util.EncryptWithAAD(assetKey, plaintext, []byte(aad))
}

func DecryptAssetNameMapping(boxID string, dek, ciphertext []byte) ([]byte, error) {
	assetKey := util.DeriveSubKey(dek, "siyuan/asset")
	aad := "scribli:v1:asset-names:" + boxID
	return util.DecryptWithAAD(assetKey, ciphertext, []byte(aad))
}

func notebookCryptBackupPath(boxID string) string {
	return filepath.Join(util.DataDir, boxID, ".scribli", "notebook-crypt-backup.json")
}

func writeNotebookCryptBackup(boxID string, crypt *conf.BoxEncryption) error {
	backupPath := notebookCryptBackupPath(boxID)
	if err := os.MkdirAll(filepath.Dir(backupPath), 0755); err != nil {
		return fmt.Errorf("mkdir notebook crypt backup dir failed: %w", err)
	}
	data, err := gulu.JSON.MarshalIndentJSON(crypt, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal notebook crypt backup failed: %w", err)
	}
	if err := filelock.WriteFile(backupPath, data); err != nil {
		return fmt.Errorf("write notebook crypt backup failed: %w", err)
	}
	return nil
}

func readNotebookCryptBackup(boxID string) (*conf.BoxEncryption, error) {
	backupPath := notebookCryptBackupPath(boxID)
	if !filelock.IsExist(backupPath) {
		return nil, nil
	}
	data, err := filelock.ReadFile(backupPath)
	if err != nil {
		return nil, fmt.Errorf("read notebook crypt backup failed: %w", err)
	}
	var crypt conf.BoxEncryption
	if err = gulu.JSON.UnmarshalJSON(data, &crypt); err != nil {
		return nil, fmt.Errorf("unmarshal notebook crypt backup failed: %w", err)
	}
	return &crypt, nil
}

func copyAssetDecryptIfEncrypted(srcPath, destPath string) error {
	boxID := ExtractBoxIDFromAssetsPath(srcPath)
	if boxID != "" && IsEncryptedBox(boxID) {
		HoldBoxReadLock(boxID)
		defer ReleaseBoxReadLock(boxID)
		dek, err := GetDEKIfUnlocked(boxID)
		if err != nil {

			return errors.New(Conf.Language(314))
		}
		raw, readErr := filelock.ReadFile(srcPath)
		if readErr != nil {
			return readErr
		}
		diskName := filepath.Base(srcPath)
		plain, decErr := DecryptAsset(boxID, diskName, dek, raw)
		if decErr != nil {
			return errors.New(Conf.Language(316))
		}
		if err := filelock.WriteFile(destPath, plain); err != nil {
			return err
		}
		return nil
	}
	return filelock.Copy(srcPath, destPath)
}

func CreateEncryptedBox(name, password string) (id string, err error) {
	notebookCryptoMu.Lock()
	defer notebookCryptoMu.Unlock()

	Conf.m.RLock()
	enabled := Conf.NotebookCrypto.Enabled
	Conf.m.RUnlock()
	if !enabled {
		return "", errors.New(Conf.Language(310))
	}

	kek, err := deriveKEK(password)
	if err != nil {
		return "", err
	}
	defer zeroAndClear(kek)

	id, err = CreateBox(name)
	if err != nil {
		return "", err
	}

	boxCreated := true
	defer func() {
		if err != nil && boxCreated {
			sql.RemoveEncryptedDBFile(id)
			treenode.RemoveEncryptedBlockTreeDBFile(id)
			boxDir := filepath.Join(util.DataDir, id)
			if rmErr := filelock.Remove(boxDir); rmErr != nil {
				logging.LogErrorf("cleanup failed encrypted box [%s]: %s", id, rmErr)
			}
			id = ""
		}
	}()

	enc, dek, err := WrapNewDEK(id, kek)
	if err != nil {
		return "", err
	}

	box := &Box{ID: id}
	boxConf := box.GetConf()
	boxConf.Encrypted = true
	boxConf.BoxCrypt = enc
	if err = box.SaveConf(boxConf); err != nil {
		return "", fmt.Errorf("save encrypted notebook conf failed: %w", err)
	}
	if err = writeNotebookCryptBackup(id, enc); err != nil {
		return "", fmt.Errorf("write notebook crypt backup failed: %w", err)
	}

	verifyConf := box.GetConf()
	if verifyConf == nil || !verifyConf.Encrypted || verifyConf.BoxCrypt == nil {
		err = errors.New("encrypted notebook metadata verification failed after write")
		return "", err
	}

	cachedDEKsLock.Lock()
	defer cachedDEKsLock.Unlock()
	if err = sql.OpenEncryptedDB(id, dek); err != nil {
		return "", err
	}
	if err = treenode.OpenEncryptedBlockTreeDB(id, dek); err != nil {
		sql.CloseEncryptedDB(id)
		return "", err
	}
	cachedDEKs[id] = dek

	newVal := &atomic.Int64{}
	newVal.Store(time.Now().UnixNano())
	boxLastAccess.Store(id, newVal)

	IncSync()
	return id, nil
}

func zeroAndClear(key []byte) {
	for i := range key {
		key[i] = 0
	}
}

func TouchUnlockedEncryptedBoxes() {
	now := time.Now().UnixNano()
	cachedDEKsLock.RLock()
	boxIDs := make([]string, 0, len(cachedDEKs))
	for boxID := range cachedDEKs {
		boxIDs = append(boxIDs, boxID)
	}
	cachedDEKsLock.RUnlock()
	for _, boxID := range boxIDs {
		if val, ok := boxLastAccess.Load(boxID); ok {
			val.(*atomic.Int64).Store(now)
		}
	}
}

func AutoLockIdleEncryptedBoxesJob() {
	Conf.m.RLock()
	threshold := Conf.NotebookCrypto.AutoLockMinutes
	Conf.m.RUnlock()
	if threshold <= 0 {
		return
	}

	now := time.Now().UnixNano()
	thresholdNs := int64(time.Duration(threshold) * time.Minute)

	cachedDEKsLock.RLock()
	boxIDs := make([]string, 0, len(cachedDEKs))
	for id := range cachedDEKs {
		boxIDs = append(boxIDs, id)
	}
	cachedDEKsLock.RUnlock()

	for _, boxID := range boxIDs {
		if val, ok := boxLastAccess.Load(boxID); ok {
			lastAccess := val.(*atomic.Int64).Load()
			elapsed := now - lastAccess
			if elapsed >= thresholdNs {
				logging.LogInfof("auto-locking idle encrypted notebook [%s] (elapsed=%ds, threshold=%dm)", boxID, elapsed/1e9, threshold)

				boxName := boxID
				if box := Conf.Box(boxID); nil != box {
					boxName = box.Name
				}
				Unmount(boxID)

				util.PushMsg(fmt.Sprintf(Conf.Language(322), boxName), 0)
			}
		}
	}
}

func SetAutoLockMinutes(minutes int) {
	if minutes < 0 {
		minutes = 0
	}
	Conf.m.Lock()
	Conf.NotebookCrypto.AutoLockMinutes = minutes
	Conf.m.Unlock()
}
