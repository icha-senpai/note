package api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/icha-senpai/note/kernel/model"
	"github.com/icha-senpai/note/kernel/util"
)

func TestGetFileAllowsWorkspaceTemp(t *testing.T) {
	originalWorkspaceDir := util.WorkspaceDir
	originalTempDir := util.TempDir
	workspaceDir := t.TempDir()
	util.WorkspaceDir = workspaceDir
	util.TempDir = filepath.Join(workspaceDir, "temp")
	defer func() {
		util.WorkspaceDir = originalWorkspaceDir
		util.TempDir = originalTempDir
	}()

	artifact := filepath.Join(util.TempDir, "export", "plugin-package.zip")
	if err := os.MkdirAll(filepath.Dir(artifact), 0755); err != nil {
		t.Fatal(err)
	}
	content := []byte("plugin package")
	if err := os.WriteFile(artifact, content, 0644); err != nil {
		t.Fatal(err)
	}

	request := httptest.NewRequest(http.MethodPost, "/api/file/getFile", strings.NewReader(`{"path":"temp/export/plugin-package.zip"}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router := gin.New()
	router.POST("/api/file/getFile", func(context *gin.Context) {
		context.Set(model.RoleContextKey, model.RoleAdministrator)
		getFile(context)
	})
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("workspace temp file should be accessible, got status %d: %s", recorder.Code, recorder.Body.String())
	}
	if recorder.Body.String() != string(content) {
		t.Fatalf("unexpected workspace temp file content: %q", recorder.Body.String())
	}
}
