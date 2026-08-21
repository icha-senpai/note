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
	"strings"
	"testing"

	"github.com/icha-senpai/note/kernel/model"
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

func TestSearchToolPageSizeContractMatchesDefault(t *testing.T) {
	pageSize := SearchTool.InputSchema.Properties["pageSize"]
	for _, text := range []string{SearchTool.Description, pageSize.Description} {
		if !strings.Contains(text, "32") {
			t.Fatalf("search pageSize contract does not mention default 32: %q", text)
		}
		if strings.Contains(text, "20") {
			t.Fatalf("search pageSize contract still mentions stale default 20: %q", text)
		}
	}
}

func TestSystemToolMCPToolsDiagnosticAction(t *testing.T) {
	oldConf := model.Conf
	model.Conf = nil
	t.Cleanup(func() {
		model.Conf = oldConf
	})

	action := SystemTool.InputSchema.Properties["action"]
	actionSet := map[string]bool{}
	for _, allowed := range action.Enum {
		actionSet[allowed] = true
	}
	if !actionSet["mcp_tools"] {
		t.Fatalf("system action enum missing mcp_tools: %#v", action.Enum)
	}
	if _, ok := SystemTool.EffectsFor("mcp_tools"); !ok {
		t.Fatal("system mcp_tools action missing effect metadata")
	}

	result, err := systemHandler(map[string]any{"action": "mcp_tools"})
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatalf("mcp_tools returned tool error: %#v", result.Content)
	}
	content, ok := result.StructuredContent.(map[string]any)
	if !ok {
		t.Fatalf("structured content type = %T, want map", result.StructuredContent)
	}
	if content["action"] != "mcp_tools" {
		t.Fatalf("action = %v, want mcp_tools", content["action"])
	}
	if content["registeredCount"].(int) < content["externalEligibleCount"].(int) {
		t.Fatalf("eligible tools exceeded registered tools: %#v", content)
	}

	byName := map[string]map[string]any{}
	for _, item := range content["tools"].([]map[string]any) {
		byName[item["name"].(string)] = item
	}
	if byName["system"]["externalMcpEligible"] != true {
		t.Fatalf("system should be externally eligible: %#v", byName["system"])
	}
	for _, name := range []string{"frontend", "question"} {
		if byName[name]["hiddenReason"] != "agent_only" {
			t.Fatalf("%s hidden reason = %v, want agent_only", name, byName[name]["hiddenReason"])
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
