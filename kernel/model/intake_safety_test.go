package model

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/icha-senpai/note/kernel/cache"
	"github.com/icha-senpai/note/kernel/util"
)

func TestIncompleteDatabaseScanNeverRemovesDefinitions(t *testing.T) {
	fixture := setupFileOperationTest(t)
	dir := filepath.Join(util.DataDir, "storage", "av")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	id := "20261010000000-abcdefg"
	filename := filepath.Join(dir, id+".json")
	if err := os.WriteFile(filename, []byte(`{"id":"`+id+`","name":"Retained"}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(util.DataDir, fixture.box.ID, "20261010000001-abcdefg.sy"), []byte("incomplete document"), 0600); err != nil {
		t.Fatal(err)
	}
	if items, err := UnusedAttributeViews(false); err == nil || items != nil {
		t.Fatalf("incomplete scan succeeded: %v, %v", items, err)
	}
	if _, err := RemoveUnusedAttributeViews(); err == nil {
		t.Fatal("bulk cleanup ignored scan failure")
	}
	if err := RemoveUnusedAttributeView(id); err == nil {
		t.Fatal("single cleanup ignored scan failure")
	}
	if _, err := os.Stat(filename); err != nil {
		t.Fatal("database definition was removed", err)
	}
}

func TestDocumentReadFailureDoesNotMoveData(t *testing.T) {
	fixture := setupFileOperationTest(t)
	previousWorkspace := util.WorkspaceDir
	util.WorkspaceDir = t.TempDir()
	t.Cleanup(func() { util.WorkspaceDir = previousWorkspace })
	doc := "/20261010000002-abcdefg.sy"
	filename := filepath.Join(util.DataDir, fixture.box.ID, doc)
	if err := os.Mkdir(filename, 0755); err != nil {
		t.Fatal(err)
	}
	if attrs := fixture.box.docIAL(doc); attrs != nil {
		t.Fatal("unreadable document returned properties")
	}
	if _, err := os.Stat(filename); err != nil {
		t.Fatal("unreadable source was moved", err)
	}
	if _, err := os.Stat(filepath.Join(util.WorkspaceDir, "corrupted")); !os.IsNotExist(err) {
		t.Fatal("temporary read failure classified as corruption")
	}
	cache.RemoveDocIAL(doc)
}

func TestCorruptRelationsAbortDatabaseCleanup(t *testing.T) {
	setupFileOperationTest(t)
	dir := filepath.Join(util.DataDir, "storage", "av")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	id := "20261010000003-abcdefg"
	filename := filepath.Join(dir, id+".json")
	if err := os.WriteFile(filename, []byte(`{"id":"`+id+`","name":"Retained"}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "relations.msgpack"), []byte("not msgpack"), 0600); err != nil {
		t.Fatal(err)
	}
	if items, err := UnusedAttributeViews(false); err == nil || items != nil {
		t.Fatal("corrupt relation scan succeeded")
	}
	if err := RemoveUnusedAttributeView(id); err == nil {
		t.Fatal("cleanup ignored corrupt relations")
	}
	if _, err := os.Stat(filename); err != nil {
		t.Fatal("definition was removed", err)
	}
}

func TestTemplateReferenceScanPreservesBoundaryMatchesAndRejectsErrors(t *testing.T) {
	root := t.TempDir()
	id := "20261010000004-abcdefg"
	content := append(make([]byte, 65530), []byte(id)...)
	if err := os.WriteFile(filepath.Join(root, "template.md"), content, 0600); err != nil {
		t.Fatal(err)
	}
	ids, err := scanTemplateDatabaseReferences(root, []string{id})
	if err != nil || len(ids) != 1 || ids[0] != id {
		t.Fatalf("lost boundary reference: %v %v", ids, err)
	}
	if _, err = scanTemplateDatabaseReferences(filepath.Join(root, "template.md"), []string{id}); err == nil {
		t.Fatal("invalid template root accepted")
	}
}
