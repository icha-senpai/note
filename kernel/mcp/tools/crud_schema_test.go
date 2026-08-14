// Scribli - Refactor your thinking
// Copyright (c) 2020-present, Scribli
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package tools

import (
	"path/filepath"
	"testing"

	"github.com/icha-senpai/note/kernel/util"
)

func TestCoreWorkflowToolsExposeStructuredOutputSchemas(t *testing.T) {
	for _, tool := range []*Tool{
		APICallTool,
		APICatalogTool,
		APIRouteTool,
		BlockTool,
		CanvasTool,
		DatabaseTool,
		DocumentTool,
		ExecutableBlockTool,
		ExportTool,
		ImportTool,
		NotebookTool,
		SearchTool,
		SQLTool,
		SystemTool,
		TodoWriteTool,
		WorkspaceTool,
	} {
		if tool.OutputSchema == nil {
			t.Fatalf("%s missing output schema", tool.Name)
		}
		if tool.OutputSchema.Type != "object" {
			t.Fatalf("%s output schema type = %q, want object", tool.Name, tool.OutputSchema.Type)
		}
	}
}

func TestDatabaseToolExposesCreateAction(t *testing.T) {
	action := DatabaseTool.InputSchema.Properties["action"]
	for _, allowed := range action.Enum {
		if allowed == "create" {
			return
		}
	}
	t.Fatalf("database action enum missing create: %#v", action.Enum)
}

func TestDocumentToolExposesUpdateAction(t *testing.T) {
	action := DocumentTool.InputSchema.Properties["action"]
	for _, allowed := range action.Enum {
		if allowed == "update" {
			return
		}
	}
	t.Fatalf("document action enum missing update: %#v", action.Enum)
}

func TestReadToolsExposeRedactSecretsOption(t *testing.T) {
	for _, tool := range []*Tool{BlockTool, DatabaseTool, DocumentTool, ExportTool} {
		prop, ok := tool.InputSchema.Properties["redactSecrets"]
		if !ok {
			t.Fatalf("%s missing redactSecrets input", tool.Name)
		}
		if prop.Type != "boolean" {
			t.Fatalf("%s redactSecrets type = %q, want boolean", tool.Name, prop.Type)
		}
	}
}

func TestStructuredTextResultSetsStructuredContent(t *testing.T) {
	result := structuredTextResult("done", map[string]any{"action": "test"})
	if !result.HasStructuredContent() {
		t.Fatal("structured result did not mark structured content")
	}
	content, ok := result.StructuredContent.(map[string]any)
	if !ok {
		t.Fatalf("structured content type = %T, want map", result.StructuredContent)
	}
	if content["status"] != "ok" || content["message"] != "done" || content["action"] != "test" {
		t.Fatalf("unexpected structured content: %#v", content)
	}
}

func TestTodoWriteWithoutAgentSessionDoesNotPanic(t *testing.T) {
	originalDataDir := util.DataDir
	util.DataDir = filepath.Join(t.TempDir(), "data")
	t.Cleanup(func() {
		util.DataDir = originalDataDir
	})

	result, err := todoWriteHandler(map[string]any{
		"todos": []any{
			map[string]any{"content": "Check MCP stability", "status": "in_progress"},
		},
	})
	if err != nil {
		t.Fatalf("todo_write returned error: %v", err)
	}
	if result.IsError {
		t.Fatalf("todo_write returned tool error: %#v", result.Content)
	}
	content, ok := result.StructuredContent.(map[string]any)
	if !ok {
		t.Fatalf("structured content type = %T, want map", result.StructuredContent)
	}
	if content["sessionID"] != "mcp-external" {
		t.Fatalf("sessionID = %v, want mcp-external", content["sessionID"])
	}
}
