//go:build windows

// Scribli contributors. SPDX-License-Identifier: AGPL-3.0-or-later
package model

import (
	"github.com/icha-senpai/note/kernel/util"
	"golang.org/x/sys/windows"
	"os"
	"path/filepath"
	"testing"
)

func TestSharingViolationDoesNotQuarantineDocument(t *testing.T) {
	fixture := setupFileOperationTest(t)
	previous := util.WorkspaceDir
	util.WorkspaceDir = t.TempDir()
	t.Cleanup(func() { util.WorkspaceDir = previous })
	doc := "/20261010000005-abcdefg.sy"
	filename := filepath.Join(util.DataDir, fixture.box.ID, doc)
	if err := os.WriteFile(filename, []byte(`{"ID":"20261010000005-abcdefg","Type":"NodeDocument","Properties":{"id":"20261010000005-abcdefg"}}`), 0600); err != nil {
		t.Fatal(err)
	}
	name, err := windows.UTF16PtrFromString(filename)
	if err != nil {
		t.Fatal(err)
	}
	handle, err := windows.CreateFile(name, windows.GENERIC_READ, 0, nil, windows.OPEN_EXISTING, windows.FILE_ATTRIBUTE_NORMAL, 0)
	if err != nil {
		t.Fatal(err)
	}
	if attrs := fixture.box.docIAL(doc); attrs != nil {
		t.Fatal("locked document unexpectedly read")
	}
	windows.CloseHandle(handle)
	if attrs := fixture.box.docIAL(doc); attrs["id"] != "20261010000005-abcdefg" {
		t.Fatal("document was not readable after unlocking")
	}
	if _, err = os.Stat(filename); err != nil {
		t.Fatal("document removed after sharing violation", err)
	}
	if _, err = os.Stat(filepath.Join(util.WorkspaceDir, "corrupted")); !os.IsNotExist(err) {
		t.Fatal("sharing violation treated as corruption")
	}
}
