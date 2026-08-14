// Scribli - Refactor your thinking
// Copyright (c) 2020-present, Scribli
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package tools

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/icha-senpai/note/kernel/util"
)

func TestCanvasToolCreateGetAddAndList(t *testing.T) {
	originalDataDir := util.DataDir
	util.DataDir = filepath.Join(t.TempDir(), "data")
	t.Cleanup(func() {
		util.DataDir = originalDataDir
	})

	canvasID := "20260814120000-abcdefg"
	result, err := canvasHandler(map[string]any{
		"action": "create",
		"id":     canvasID,
		"title":  "Research Board",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError || !result.HasStructuredContent() {
		t.Fatalf("unexpected create result: %#v", result)
	}

	result, err = canvasHandler(map[string]any{
		"action": "add_node",
		"id":     canvasID,
		"node": map[string]any{
			"type": "text",
			"text": "Start here",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatalf("unexpected add_node result: %#v", result)
	}
	content := result.StructuredContent.(map[string]any)
	canvas := content["canvas"].(map[string]any)
	nodes := canvas["nodes"].([]any)
	if len(nodes) != 1 {
		t.Fatalf("nodes length = %d, want 1", len(nodes))
	}

	result, err = canvasHandler(map[string]any{
		"action": "get",
		"id":     canvasID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError || !result.HasStructuredContent() {
		t.Fatalf("unexpected get result: %#v", result)
	}

	result, err = canvasHandler(map[string]any{"action": "list"})
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatalf("unexpected list result: %#v", result)
	}
	content = result.StructuredContent.(map[string]any)
	if content["count"] != 1 {
		t.Fatalf("count = %#v, want 1", content["count"])
	}
}

func TestCanvasToolAddsScribliNodeMetadata(t *testing.T) {
	originalDataDir := util.DataDir
	util.DataDir = filepath.Join(t.TempDir(), "data")
	t.Cleanup(func() {
		util.DataDir = originalDataDir
	})

	canvasID := "20260814121000-bcdefgh"
	if result, err := canvasHandler(map[string]any{
		"action": "create",
		"id":     canvasID,
		"title":  "Agent Board",
	}); err != nil || result.IsError {
		t.Fatalf("unexpected create result: %#v, err=%v", result, err)
	}

	result, err := canvasHandler(map[string]any{
		"action":     "add_scribli_node",
		"id":         canvasID,
		"kind":       "database",
		"databaseID": "20260814121100-cdefghi",
		"viewID":     "view-table",
		"label":      "Roadmap DB",
		"x":          float64(64),
		"y":          float64(128),
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError || !result.HasStructuredContent() {
		t.Fatalf("unexpected add_scribli_node result: %#v", result)
	}
	content := result.StructuredContent.(map[string]any)
	node := content["node"].(map[string]any)
	if node["type"] != "text" || node["text"] != "Roadmap DB" {
		t.Fatalf("unexpected node display payload: %#v", node)
	}
	metadata := node["scribli"].(map[string]any)
	if metadata["kind"] != "database" || metadata["databaseID"] != "20260814121100-cdefghi" || metadata["viewID"] != "view-table" {
		t.Fatalf("unexpected scribli metadata: %#v", metadata)
	}
}

func TestCanvasToolCreateAndEmbedExposesWriteAction(t *testing.T) {
	action := CanvasTool.InputSchema.Properties["action"]
	found := false
	for _, allowed := range action.Enum {
		if allowed == "create_and_embed" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("canvas action enum missing create_and_embed: %#v", action.Enum)
	}

	effects, ok := CanvasTool.EffectsFor("create_and_embed")
	if !ok || !effects.LocalWrite {
		t.Fatalf("create_and_embed effects = %#v, ok=%v; want LocalWrite", effects, ok)
	}
}

func TestCanvasToolCreateAndEmbedDoesNotWriteWhenEmbedTargetMissing(t *testing.T) {
	originalDataDir := util.DataDir
	util.DataDir = filepath.Join(t.TempDir(), "data")
	t.Cleanup(func() {
		util.DataDir = originalDataDir
	})

	canvasID := "20260814123000-defghij"
	result, err := canvasHandler(map[string]any{
		"action":        "create_and_embed",
		"id":            canvasID,
		"title":         "Embedded Board",
		"kind":          "query",
		"query":         "canvas planning",
		"embedParentID": "20260814123100-efghijk",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Fatalf("expected missing embed target to fail: %#v", result)
	}
	if len(result.Content) == 0 || !strings.Contains(result.Content[0].Text, "parent block not found") {
		t.Fatalf("unexpected error result: %#v", result.Content)
	}

	path, err := canvasFilePath(canvasID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("canvas file should not be written on embed failure, stat err=%v", err)
	}
}

func TestCanvasEmbedMarkdown(t *testing.T) {
	markdown := canvasEmbedMarkdown("20260814124000-fghijkl")
	if markdown != "```scribli-canvas\n20260814124000-fghijkl\n```" {
		t.Fatalf("unexpected embed markdown: %q", markdown)
	}
}
