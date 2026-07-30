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
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/icha-senpai/note/third_party/forks/gulu"
	"github.com/icha-senpai/note/kernel/conf"
	"github.com/icha-senpai/note/kernel/util"
)

func setDEKForTest(boxID string, dek []byte) {
	cachedDEKsLock.Lock()
	defer cachedDEKsLock.Unlock()
	cachedDEKs[boxID] = dek
}

func TestIsBoxUnlockedLifecycle(t *testing.T) {
	LockBox("lifecycle-test-box")
	boxID := "lifecycle-test-box"
	if IsBoxUnlocked(boxID) {
		t.Fatalf("box should not be unlocked after LockBox")
	}
	dek, _ := util.GenerateDEK()
	setDEKForTest(boxID, dek)
	if !IsBoxUnlocked(boxID) {
		t.Fatalf("box should be unlocked after setDEKForTest")
	}
	LockBox(boxID)
	if IsBoxUnlocked(boxID) {
		t.Fatalf("box should be locked after LockBox")
	}
}

func TestGetDEKReturnsErrorAfterLock(t *testing.T) {
	dek, _ := util.GenerateDEK()
	boxID := "get-dek-test-box"
	setDEKForTest(boxID, dek)

	got, err := GetDEK(boxID)
	if err != nil {
		t.Fatalf("GetDEK before lock failed: %v", err)
	}
	if !bytes.Equal(dek, got) {
		t.Fatalf("GetDEK returned wrong DEK")
	}

	LockBox(boxID)
	if _, err := GetDEK(boxID); err == nil {
		t.Fatalf("GetDEK should fail after LockBox")
	}
}

func TestWrapNewDEKRoundTrip(t *testing.T) {
	kek, _ := util.GenerateDEK()
	defer LockBox("wrap-roundtrip-box")

	boxEnc, _, err := WrapNewDEK("wrap-roundtrip-box", kek)
	if err != nil {
		t.Fatalf("WrapNewDEK failed: %v", err)
	}
	if len(boxEnc.WrappedDEK) == 0 || len(boxEnc.WrapNonce) != 12 {
		t.Fatalf("BoxEncryption fields malformed: wrappedDEK=%d wrapNonce=%d", len(boxEnc.WrappedDEK), len(boxEnc.WrapNonce))
	}

	dek, err := decryptWrappedDEK("wrap-roundtrip-box", boxEnc, kek)
	if err != nil {
		t.Fatalf("decryptWrappedDEK failed: %v", err)
	}
	if len(dek) != 32 {
		t.Fatalf("expected 32-byte DEK, got %d", len(dek))
	}
}

func TestDecryptWrappedDEKWithWrongKEK(t *testing.T) {
	kek1, _ := util.GenerateDEK()
	boxEnc, _, _ := WrapNewDEK("wrong-kek-box", kek1)

	kek2, _ := util.GenerateDEK()
	defer LockBox("wrong-kek-box")

	if _, err := decryptWrappedDEK("wrong-kek-box", boxEnc, kek2); err == nil {
		t.Fatalf("decryptWrappedDEK with wrong KEK should fail")
	}
}

func TestWrapNewDEKProducesUniqueDEKs(t *testing.T) {
	kek, _ := util.GenerateDEK()
	defer LockBox("uniq-box-1")
	defer LockBox("uniq-box-2")

	_, dek1, _ := WrapNewDEK("uniq-box-1", kek)
	_, dek2, _ := WrapNewDEK("uniq-box-2", kek)

	if bytes.Equal(dek1, dek2) {
		t.Fatalf("two WrapNewDEK calls produced identical DEKs (not random?)")
	}
}

func TestBoxEncryptionRoundTripViaUtil(t *testing.T) {
	kek, _ := util.GenerateDEK()
	originalDEK, _ := util.GenerateDEK()

	wrapped, _ := util.EncryptWithAAD(kek, originalDEK, wrappedDEKAAD(""))
	boxEnc := &conf.BoxEncryption{
		WrappedDEK: wrapped,
		WrapNonce:  wrapped[:12],
	}

	recoveredDEK, err := decryptWrappedDEK("", boxEnc, kek)
	if err != nil {
		t.Fatalf("Decrypt wrapped DEK failed: %v", err)
	}
	if !bytes.Equal(originalDEK, recoveredDEK) {
		t.Fatalf("DEK round-trip mismatch")
	}
}

func TestUnmount0ClearsDEKForUnmountedEncryptedBox(t *testing.T) {
	boxID := "unmount-unlocked-test-box"

	origDataDir := util.DataDir
	tempDir := t.TempDir()
	util.DataDir = tempDir
	defer func() {
		util.DataDir = origDataDir
		LockBox("unmount-unlocked-test-box")
	}()

	confDir := filepath.Join(tempDir, boxID, ".scribli")
	if err := os.MkdirAll(confDir, 0755); err != nil {
		t.Fatalf("mkdir conf dir failed: %v", err)
	}
	boxConf := conf.NewBoxConf()
	boxConf.Encrypted = true
	boxConf.Closed = true
	confData, _ := gulu.JSON.MarshalIndentJSON(boxConf, "", "  ")
	if err := os.WriteFile(filepath.Join(confDir, "conf.json"), confData, 0644); err != nil {
		t.Fatalf("write conf.json failed: %v", err)
	}

	if !IsEncryptedBox(boxID) {
		t.Fatalf("precondition failed: IsEncryptedBox should return true")
	}

	dek, _ := util.GenerateDEK()
	setDEKForTest(boxID, dek)
	if !IsBoxUnlocked(boxID) {
		t.Fatalf("precondition failed: box should be unlocked after setDEKForTest")
	}

	clearDEKIfUnlockedEncryptedBox(boxID)

	if IsBoxUnlocked(boxID) {
		t.Fatalf("DEK should be cleared for unlocked encrypted box")
	}
}

func TestBackupRejectsUnsupportedSpec(t *testing.T) {
	for _, spec := range []int{0, conf.CurrentNotebookCryptoSpec + 1} {
		t.Run(fmt.Sprintf("spec-%d", spec), func(t *testing.T) {
			origDataDir := util.DataDir
			util.DataDir = t.TempDir()
			defer func() { util.DataDir = origDataDir }()

			backup := &conf.NotebookCrypto{Spec: spec}
			backupPath := filepath.Join(util.DataDir, ".scribli", "notebook-crypto-backup.json")
			if err := os.MkdirAll(filepath.Dir(backupPath), 0755); err != nil {
				t.Fatal(err)
			}
			data, _ := json.Marshal(backup)
			if err := os.WriteFile(backupPath, data, 0644); err != nil {
				t.Fatal(err)
			}

			if _, err := loadNotebookCryptoBackup(); err == nil {
				t.Fatalf("unsupported notebook crypto spec [%d] should be rejected", spec)
			}
		})
	}
}

func TestBackupChecksumCorruption(t *testing.T) {
	origDataDir := util.DataDir
	tempDir := t.TempDir()
	util.DataDir = tempDir
	defer func() { util.DataDir = origDataDir }()

	nc := &conf.NotebookCrypto{
		Enabled:    true,
		MasterSalt: []byte("corrupt-test-salt12"),
	}
	prepareBackupForWrite(nc)
	backupPath := filepath.Join(tempDir, ".scribli", "notebook-crypto-backup.json")
	os.MkdirAll(filepath.Dir(backupPath), 0755)
	data, _ := json.Marshal(nc)
	os.WriteFile(backupPath, data, 0644)

	nc1, err := loadNotebookCryptoBackup()
	if err != nil {
		t.Fatalf("loadNotebookCryptoBackup should succeed with valid backup: %v", err)
	}
	if nc1.Spec != 1 {
		t.Fatalf("expected Spec=1, got %d", nc1.Spec)
	}

	nc.MasterSalt[0] ^= 0xFF

	data, _ = json.Marshal(nc)
	os.WriteFile(backupPath, data, 0644)

	_, err = loadNotebookCryptoBackup()
	if err == nil {
		t.Fatalf("loadNotebookCryptoBackup should fail with corrupted backup")
	}
}

func TestBackupKEKMACVerification(t *testing.T) {
	nc := &conf.NotebookCrypto{
		Enabled:    true,
		MasterSalt: []byte("kekmac-test-salt12"),
	}
	prepareBackupForWrite(nc)

	correctKek, _ := util.GenerateDEK()
	nc.KEKMAC = computeKEKMAC(nc, correctKek)

	if !verifyKEKMAC(nc, correctKek) {
		t.Fatalf("verifyKEKMAC should pass with correct KEK")
	}

	wrongKek, _ := util.GenerateDEK()
	if verifyKEKMAC(nc, wrongKek) {
		t.Fatalf("verifyKEKMAC should fail with wrong KEK")
	}
}

func TestDeriveKEKRejectsTamperedBackupMAC(t *testing.T) {
	origDataDir := util.DataDir
	util.DataDir = t.TempDir()
	defer func() { util.DataDir = origDataDir }()

	password := "authenticated-backup-test"
	salt, _ := util.GenerateSalt()
	params := util.DefaultArgon2Params()
	kek := util.DeriveKey(password, salt, params)
	verifier, _ := util.EncryptWithAAD(kek, kekVerifierMagic, []byte("scribli:v1:kek-verifier"))
	nc := conf.NotebookCrypto{
		Enabled:     true,
		MasterSalt:  salt,
		KDFParams:   params,
		KEKVerifier: verifier,
		Spec:        1,
		KEKMAC:      []byte("tampered"),
	}
	prepareBackupForWrite(&nc)
	nc.KEKMAC = []byte("tampered")
	backupPath := filepath.Join(util.DataDir, ".scribli", "notebook-crypto-backup.json")
	if err := os.MkdirAll(filepath.Dir(backupPath), 0755); err != nil {
		t.Fatal(err)
	}
	backupData, _ := json.Marshal(&nc)
	if err := os.WriteFile(backupPath, backupData, 0644); err != nil {
		t.Fatal(err)
	}

	originalConf := Conf
	Conf = NewAppConf()
	Conf.NotebookCrypto = &nc
	defer func() { Conf = originalConf }()

	if derived, err := deriveKEK(password); err == nil {
		zeroAndClear(derived)
		t.Fatalf("deriveKEK should reject a backup with an invalid KEKMAC")
	}
}

func TestDeriveKEKAllowsLocalAutoLockChange(t *testing.T) {
	origDataDir := util.DataDir
	util.DataDir = t.TempDir()
	defer func() { util.DataDir = origDataDir }()

	password := "local-auto-lock-test"
	salt, _ := util.GenerateSalt()
	params := util.DefaultArgon2Params()
	kek := util.DeriveKey(password, salt, params)
	defer zeroAndClear(kek)
	verifier, _ := util.EncryptWithAAD(kek, kekVerifierMagic, []byte("scribli:v1:kek-verifier"))
	backup := &conf.NotebookCrypto{
		Enabled:         true,
		MasterSalt:      salt,
		KDFParams:       params,
		KEKVerifier:     verifier,
		AutoLockMinutes: 5,
	}
	if err := writeNotebookCryptoBackupData(backup, kek); err != nil {
		t.Fatal(err)
	}

	local := *backup
	local.AutoLockMinutes = 30
	originalConf := Conf
	Conf = NewAppConf()
	Conf.NotebookCrypto = &local
	defer func() { Conf = originalConf }()

	derived, err := deriveKEK(password)
	if err != nil {
		t.Fatalf("deriveKEK rejected a local AutoLockMinutes change: %v", err)
	}
	zeroAndClear(derived)
}

func TestDeepCopyBoxEncryptionPreservesSpec(t *testing.T) {
	src := &conf.BoxEncryption{Spec: 1, WrappedDEK: []byte{1, 2}, WrapNonce: []byte{3, 4}, CreatedAt: 5}
	got := DeepCopyBoxEncryption(src)
	if got.Spec != src.Spec {
		t.Fatalf("BoxEncryption.Spec changed during deep copy: got %d want %d", got.Spec, src.Spec)
	}
}

func TestUnknownBlockRefFailsClosed(t *testing.T) {
	if !normalBoxBlockRefCrossesBoundary(nil) {
		t.Fatalf("an unresolved block reference should fail closed")
	}
}

func TestBackupMACRoundTrip(t *testing.T) {
	origDataDir := util.DataDir
	tempDir := t.TempDir()
	util.DataDir = tempDir
	defer func() { util.DataDir = origDataDir }()

	password := "round-trip-test"
	salt, _ := util.GenerateSalt()
	params := util.DefaultArgon2Params()
	kek := util.DeriveKey(password, salt, params)
	defer zeroAndClear(kek)

	verifierCT, _ := util.EncryptWithAAD(kek, kekVerifierMagic, []byte("scribli:v1:kek-verifier"))
	nc := &conf.NotebookCrypto{
		Enabled:     true,
		MasterSalt:  salt,
		KDFParams:   params,
		KEKVerifier: verifierCT,
	}

	if err := writeNotebookCryptoBackupData(nc, kek); err != nil {
		t.Fatalf("writeNotebookCryptoBackupData failed: %v", err)
	}

	loaded, err := loadNotebookCryptoBackup()
	if err != nil {
		t.Fatalf("loadNotebookCryptoBackup failed: %v", err)
	}
	if loaded.Spec >= 1 && len(loaded.KEKMAC) > 0 && !verifyKEKMAC(loaded, kek) {
		t.Fatalf("verifyKEKMAC failed on round-trip backup (MAC was computed in wrong order)")
	}
}

func TestLockBoxConcurrentReads(t *testing.T) {
	LockBox("concurrent-single-box")
	boxID := "concurrent-single-box"
	dek, _ := util.GenerateDEK()
	setDEKForTest(boxID, dek)

	stop := make(chan struct{})
	var wg sync.WaitGroup
	for range 5 {
		wg.Go(func() {
			for {
				select {
				case <-stop:
					return
				default:
					HoldBoxReadLock(boxID)
					time.Sleep(time.Microsecond)
					ReleaseBoxReadLock(boxID)
				}
			}
		})
	}

	time.Sleep(10 * time.Millisecond)
	LockBox(boxID)
	close(stop)
	wg.Wait()

	if IsBoxUnlocked(boxID) {
		t.Fatalf("box should be locked after concurrent LockBox")
	}
}

func TestLockBoxClearsTempDirs(t *testing.T) {
	boxID := "temp-cleanup-box"
	dek, _ := util.GenerateDEK()
	setDEKForTest(boxID, dek)

	origTempDir := util.TempDir
	tempDir := t.TempDir()
	util.TempDir = tempDir
	defer func() {
		util.TempDir = origTempDir
		LockBox("temp-cleanup-box")
	}()

	exportDir := filepath.Join(tempDir, "export", boxID)
	repoDiffDir := filepath.Join(tempDir, "repo", "diff", boxID)
	repoRollbackDir := filepath.Join(tempDir, "repo", "rollback", boxID)
	for _, d := range []string{exportDir, repoDiffDir, repoRollbackDir} {
		if err := os.MkdirAll(d, 0755); err != nil {
			t.Fatalf("mkdir %s failed: %v", d, err)
		}
		if err := os.WriteFile(filepath.Join(d, "test.txt"), []byte("test"), 0644); err != nil {
			t.Fatalf("write test file failed: %v", err)
		}
	}

	LockBox(boxID)

	for _, d := range []string{exportDir, repoDiffDir, repoRollbackDir} {
		if _, err := os.Stat(d); !os.IsNotExist(err) {
			t.Fatalf("temp dir %s should be removed by LockBox", d)
		}
	}
}

func TestEncryptFileAADBoundToBaseNameNotPath(t *testing.T) {
	boxID := "20240101120000-boxaaaaa"
	dek, _ := util.GenerateDEK()
	base := "20240101120000-1a2b3c4.sy"
	plain := []byte(`{"ID":"20240101120000-1a2b3c4","Properties":{"title":"doc"}}`)

	ct, err := EncryptFile(boxID, "/20240101120000-parentA/"+base, dek, plain)
	if err != nil {
		t.Fatalf("EncryptFile failed: %v", err)
	}

	if got, err := DecryptFile(boxID, "/20240101120000-parentB/"+base, dek, ct); err != nil {
		t.Fatalf("decrypt with different parent dir should succeed: %v", err)
	} else if string(got) != string(plain) {
		t.Fatalf("decrypted content mismatch")
	}
	if got, err := DecryptFile(boxID, base, dek, ct); err != nil {
		t.Fatalf("decrypt with bare base name should succeed: %v", err)
	} else if string(got) != string(plain) {
		t.Fatalf("decrypted content mismatch")
	}

	otherBase := "20240101120000-zzzzzzz.sy"
	if _, err := DecryptFile(boxID, otherBase, dek, ct); err == nil {
		t.Fatal("decrypt with different base name must fail")
	}

	otherBox := "20240101120000-otherbox"
	if _, err := DecryptFile(otherBox, base, dek, ct); err == nil {
		t.Fatal("decrypt with different boxID must fail")
	}
}

func TestEncryptFileRejectsInvalidBaseName(t *testing.T) {
	boxID := "20240101120000-boxaaaaa"
	dek, _ := util.GenerateDEK()
	plain := []byte("test")

	if _, err := EncryptFile(boxID, "/dir/random.txt", dek, plain); err == nil {
		t.Fatal("should reject non-.sy extension")
	}
	if _, err := EncryptFile(boxID, "/dir/notanid.sy", dek, plain); err == nil {
		t.Fatal("should reject non-node-id stem")
	}
}

func TestDecryptFileRejectsInvalidBaseName(t *testing.T) {
	boxID := "20240101120000-boxaaaaa"
	dek, _ := util.GenerateDEK()
	ct := []byte("ciphertext-bytes")

	if _, err := DecryptFile(boxID, "/dir/random.txt", dek, ct); err == nil {
		t.Fatal("should reject non-.sy extension on decrypt")
	}
	if _, err := DecryptFile(boxID, "/dir/notanid.sy", dek, ct); err == nil {
		t.Fatal("should reject non-node-id stem on decrypt")
	}
}

func TestEnabledWithoutBackupReturnsRecoveryError(t *testing.T) {
	origDataDir := util.DataDir
	tempDir := t.TempDir()
	util.DataDir = tempDir
	defer func() { util.DataDir = origDataDir }()

	password := "recovery-test-pw"
	salt, _ := util.GenerateSalt()
	params := util.DefaultArgon2Params()
	kek := util.DeriveKey(password, salt, params)
	defer zeroAndClear(kek)

	verifierCT, _ := util.EncryptWithAAD(kek, kekVerifierMagic, []byte("scribli:v1:kek-verifier"))
	nc := &conf.NotebookCrypto{
		Enabled:     true,
		MasterSalt:  salt,
		KDFParams:   params,
		KEKVerifier: verifierCT,
	}
	originalConf := Conf
	Conf = NewAppConf()
	Conf.NotebookCrypto = nc
	defer func() { Conf = originalConf }()

	_, err := deriveKEK(password)
	if err == nil {
		t.Fatal("deriveKEK should fail when enabled but backup is missing")
	}

	if err.Error() != Conf.Language(315) {
		t.Fatalf("expected recovery hint (Language 315), got: %v", err)
	}

	if _, statErr := os.Stat(notebookCryptoBackupPath()); !os.IsNotExist(statErr) {
		t.Fatalf("backup file should not be generated during deriveKEK; stat err=%v", statErr)
	}
}

func TestSaveNotebookCryptoBackupRejectsNilKEK(t *testing.T) {
	origDataDir := util.DataDir
	util.DataDir = t.TempDir()
	defer func() { util.DataDir = origDataDir }()

	originalConf := Conf
	Conf = NewAppConf()
	Conf.NotebookCrypto = conf.NewNotebookCrypto()
	defer func() { Conf = originalConf }()

	if err := saveNotebookCryptoBackup(nil); err == nil {
		t.Fatal("saveNotebookCryptoBackup(nil) should be rejected")
	}
	if err := writeNotebookCryptoBackupData(Conf.NotebookCrypto, nil); err == nil {
		t.Fatal("writeNotebookCryptoBackupData(nc, nil) should be rejected")
	}
}

func TestEncryptedNotebookHistoryScanFailsClosed(t *testing.T) {
	originalHistoryDir := util.HistoryDir
	historyPath := filepath.Join(t.TempDir(), "history") + "\x00"
	util.HistoryDir = historyPath
	defer func() { util.HistoryDir = originalHistoryDir }()

	if _, err := scanEncryptedNotebookHistory(); err == nil {
		t.Fatal("unreadable history structure should return an error")
	}
	if !HasEncryptedNotebookHistory() {
		t.Fatal("public history dependency check should fail closed on scan errors")
	}
}

func TestListEncryptedNotebooksReturnsScanError(t *testing.T) {
	originalDataDir := util.DataDir
	dataPath := filepath.Join(t.TempDir(), "data") + "\x00"
	util.DataDir = dataPath
	defer func() { util.DataDir = originalDataDir }()

	if _, err := listAllEncryptedBoxIDs(); err == nil {
		t.Fatal("invalid data directory should return a scan error")
	}
}

func TestEnableEncryptedNotebookRestoresConfigWhenBackupWriteFails(t *testing.T) {
	originalConf := Conf
	originalDataDir := util.DataDir
	originalHistoryDir := util.HistoryDir
	dataDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dataDir, ".scribli"), []byte("not a directory"), 0600); err != nil {
		t.Fatal(err)
	}
	Conf = NewAppConf()
	Conf.NotebookCrypto = conf.NewNotebookCrypto()
	Conf.FileTree = conf.NewFileTree()
	util.DataDir = dataDir
	util.HistoryDir = filepath.Join(dataDir, "history")
	defer func() {
		Conf = originalConf
		util.DataDir = originalDataDir
		util.HistoryDir = originalHistoryDir
	}()

	before, err := json.Marshal(Conf.NotebookCrypto)
	if err != nil {
		t.Fatal(err)
	}
	if err = EnableEncryptedNotebook("backup-write-failure"); err == nil {
		t.Fatal("enable encrypted notebook should fail when the recovery backup cannot be written")
	}
	after, err := json.Marshal(Conf.NotebookCrypto)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("failed enable should restore the complete in-memory notebook crypto configuration")
	}
}
