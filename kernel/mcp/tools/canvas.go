// Scribli - Refactor your thinking
// Copyright (c) 2020-present, Scribli
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package tools

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/icha-senpai/note/kernel/model"
	"github.com/icha-senpai/note/kernel/treenode"
	"github.com/icha-senpai/note/kernel/util"
	"github.com/icha-senpai/note/third_party/forks/lute/ast"
)

var CanvasTool = &Tool{
	Name:        "canvas",
	Title:       "Canvas",
	Description: "Manage local JSON Canvas-compatible visual workspaces. Actions: create(title?, id?, canvas?), create_and_embed(title?, id?, canvas?, node?/kind?, embedParentID, embedPosition?), list(), get(id), update(id, canvas), delete(id), add_node(id, node), add_scribli_node(id, kind, refID?/label?/text?/query?/assetPath?/databaseID?/viewID?/x?/y?/width?/height?), add_edge(id, edge). Canvas files are stored locally under workspace storage/canvas.",
	InputSchema: ToolSchema{
		Type: "object",
		Properties: map[string]Property{
			"action": {Type: "string", Description: "Operation", Enum: []string{"create", "create_and_embed", "list", "get", "update", "delete", "add_node", "add_scribli_node", "add_edge"}},
			"id":     {Type: "string", Description: "Canvas ID. Defaults to a generated Scribli ID for create."},
			"title":  {Type: "string", Description: "Canvas title for create."},
			"kind":   {Type: "string", Description: "Scribli node kind for add_scribli_node.", Enum: []string{"document", "block", "asset", "database", "query", "executable_output"}},
			"refID":  {Type: "string", Description: "Referenced Scribli document/block/output ID for add_scribli_node."},
			"label":  {Type: "string", Description: "Display label for add_scribli_node."},
			"text":   {Type: "string", Description: "Text body for add_scribli_node."},
			"query":  {Type: "string", Description: "Search/query text for query nodes."},
			"assetPath": {
				Type:        "string",
				Description: "Local workspace asset path for asset nodes, such as assets/image.png.",
			},
			"databaseID": {Type: "string", Description: "Database ID for database nodes."},
			"viewID":     {Type: "string", Description: "Optional database view ID for database nodes."},
			"x":          {Type: "number", Description: "Canvas x coordinate for add_scribli_node."},
			"y":          {Type: "number", Description: "Canvas y coordinate for add_scribli_node."},
			"width":      {Type: "number", Description: "Node width for add_scribli_node."},
			"height":     {Type: "number", Description: "Node height for add_scribli_node."},
			"embedParentID": {
				Type:        "string",
				Description: "Container block/document ID where create_and_embed should append or prepend the scribli-canvas render block.",
			},
			"embedPreviousID": {
				Type:        "string",
				Description: "Sibling block ID to insert the scribli-canvas render block after for create_and_embed.",
			},
			"embedNextID": {
				Type:        "string",
				Description: "Sibling block ID to insert the scribli-canvas render block before for create_and_embed.",
			},
			"embedPosition": {
				Type:        "string",
				Description: "Where to place the scribli-canvas render block for create_and_embed. Defaults to append.",
				Enum:        []string{"append", "prepend", "insert"},
			},
			"canvas": {
				Type:        "object",
				Description: "Full JSON Canvas-compatible payload with nodes and edges arrays.",
				Properties:  map[string]Property{},
			},
			"node": {
				Type:        "object",
				Description: "JSON Canvas node payload. type defaults to text; id/x/y/width/height are filled if omitted.",
				Properties:  map[string]Property{},
			},
			"edge": {
				Type:        "object",
				Description: "JSON Canvas edge payload. id is filled if omitted.",
				Properties:  map[string]Property{},
			},
		},
		Required: []string{"action"},
	},
	OutputSchema: structuredOutputSchema(),
	EffectScope:  EffectScopeLocal,
	ActionEffects: mergeEffectMaps(
		effectMap(ToolEffects{LocalRead: true}, "list", "get"),
		effectMap(ToolEffects{LocalWrite: true}, "create", "create_and_embed", "update", "delete", "add_node", "add_scribli_node", "add_edge"),
	),
	Handler: canvasHandler,
}

func init() {
	register(CanvasTool)
}

func canvasHandler(args map[string]any) (CallToolResult, error) {
	action, _ := args["action"].(string)
	switch action {
	case "create":
		return canvasCreate(args)
	case "create_and_embed":
		return canvasCreateAndEmbed(args)
	case "list":
		return canvasList(args)
	case "get":
		return canvasGet(args)
	case "update":
		return canvasUpdate(args)
	case "delete":
		return canvasDelete(args)
	case "add_node":
		return canvasAddNode(args)
	case "add_scribli_node":
		return canvasAddScribliNode(args)
	case "add_edge":
		return canvasAddEdge(args)
	}
	return errorResult("unknown action '" + action + "', expected one of: [create, create_and_embed, list, get, update, delete, add_node, add_scribli_node, add_edge]"), nil
}

func canvasCreate(args map[string]any) (CallToolResult, error) {
	id, title, canvas, err := buildNewCanvas(args)
	if err != nil {
		return errorResult("create canvas error: " + err.Error()), nil
	}
	if err := writeCanvas(id, canvas); err != nil {
		return errorResult("create canvas error: " + err.Error()), nil
	}
	return structuredTextResult("canvas created: "+id, map[string]any{
		"action": "create",
		"id":     id,
		"title":  title,
		"canvas": canvas,
	}), nil
}

func buildNewCanvas(args map[string]any) (id, title string, canvas map[string]any, err error) {
	id, _ = args["id"].(string)
	id = strings.TrimSpace(id)
	if id == "" {
		id = ast.NewNodeID()
	}
	if err := validateCanvasID(id); err != nil {
		return "", "", nil, err
	}

	title, _ = args["title"].(string)
	title = strings.TrimSpace(title)
	if title == "" {
		title = "Untitled Canvas"
	}

	canvas = map[string]any{}
	if raw, ok := args["canvas"].(map[string]any); ok && raw != nil {
		canvas = cloneJSONMap(raw)
	}
	normalizeCanvas(canvas)
	now := time.Now().UTC().Format(time.RFC3339)
	canvas["scribli"] = mergeScribliCanvasMetadata(canvas["scribli"], map[string]any{
		"id":      id,
		"title":   title,
		"format":  "json-canvas",
		"created": now,
		"updated": now,
	})

	return id, title, canvas, nil
}

func canvasCreateAndEmbed(args map[string]any) (CallToolResult, error) {
	id, title, canvas, err := buildNewCanvas(args)
	if err != nil {
		return errorResult("create_and_embed canvas error: " + err.Error()), nil
	}
	if rawNode, ok := args["node"].(map[string]any); ok && rawNode != nil {
		nodes := canvasArray(canvas["nodes"])
		nodes = append(nodes, normalizeCanvasNode(cloneJSONMap(rawNode)))
		canvas["nodes"] = nodes
	}
	if kind, _ := args["kind"].(string); strings.TrimSpace(kind) != "" {
		node, nodeErr := buildScribliCanvasNode(args)
		if nodeErr != nil {
			return errorResult("create_and_embed canvas error: " + nodeErr.Error()), nil
		}
		nodes := canvasArray(canvas["nodes"])
		nodes = append(nodes, node)
		canvas["nodes"] = nodes
	}

	embed, err := canvasEmbedOperation(args)
	if err != nil {
		return errorResult("create_and_embed canvas error: " + err.Error()), nil
	}
	if err := writeCanvas(id, canvas); err != nil {
		return errorResult("create_and_embed canvas error: " + err.Error()), nil
	}

	embedBlockID, markdown, err := insertCanvasEmbedBlock(id, embed)
	if err != nil {
		if path, pathErr := canvasFilePath(id); pathErr == nil {
			_ = os.Remove(path)
		}
		return errorResult("create_and_embed canvas error: " + err.Error()), nil
	}

	return structuredTextResult("canvas created and embedded: "+id, map[string]any{
		"action":        "create_and_embed",
		"id":            id,
		"title":         title,
		"canvas":        canvas,
		"embedBlockID":  embedBlockID,
		"embedMarkdown": markdown,
		"embed":         embed.toMap(),
	}), nil
}

func canvasList(_ map[string]any) (CallToolResult, error) {
	dir, err := canvasStorageDir()
	if err != nil {
		return errorResult("list canvas error: " + err.Error()), nil
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return structuredTextResult("no canvases found", map[string]any{"action": "list", "count": 0, "items": []map[string]any{}}), nil
		}
		return errorResult("list canvas error: " + err.Error()), nil
	}

	items := []map[string]any{}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".canvas" {
			continue
		}
		id := strings.TrimSuffix(entry.Name(), ".canvas")
		canvas, readErr := readCanvas(id)
		if readErr != nil {
			items = append(items, map[string]any{"id": id, "error": readErr.Error()})
			continue
		}
		items = append(items, canvasSummary(id, canvas))
	}
	sort.Slice(items, func(i, j int) bool {
		left := fmt.Sprintf("%v", items[i]["updated"])
		if left == "" || left == "<nil>" {
			left = fmt.Sprintf("%v", items[i]["id"])
		}
		right := fmt.Sprintf("%v", items[j]["updated"])
		if right == "" || right == "<nil>" {
			right = fmt.Sprintf("%v", items[j]["id"])
		}
		return left > right
	})
	return structuredTextResult(fmt.Sprintf("%d canvas(es) found", len(items)), map[string]any{
		"action": "list",
		"count":  len(items),
		"items":  items,
	}), nil
}

func canvasGet(args map[string]any) (CallToolResult, error) {
	id, err := canvasRequiredID(args)
	if err != nil {
		return errorResult("get canvas error: " + err.Error()), nil
	}
	canvas, err := readCanvas(id)
	if err != nil {
		return errorResult("get canvas error: " + err.Error()), nil
	}
	return structuredTextResult("canvas loaded: "+id, map[string]any{
		"action": "get",
		"id":     id,
		"canvas": canvas,
	}), nil
}

func canvasUpdate(args map[string]any) (CallToolResult, error) {
	id, err := canvasRequiredID(args)
	if err != nil {
		return errorResult("update canvas error: " + err.Error()), nil
	}
	canvas, ok := args["canvas"].(map[string]any)
	if !ok || canvas == nil {
		return errorResult("update canvas error: canvas object is required"), nil
	}
	next := cloneJSONMap(canvas)
	normalizeCanvas(next)
	next["scribli"] = mergeScribliCanvasMetadata(next["scribli"], map[string]any{
		"id":      id,
		"format":  "json-canvas",
		"updated": time.Now().UTC().Format(time.RFC3339),
	})
	if err = writeCanvas(id, next); err != nil {
		return errorResult("update canvas error: " + err.Error()), nil
	}
	return structuredTextResult("canvas updated: "+id, map[string]any{
		"action": "update",
		"id":     id,
		"canvas": next,
	}), nil
}

func canvasDelete(args map[string]any) (CallToolResult, error) {
	id, err := canvasRequiredID(args)
	if err != nil {
		return errorResult("delete canvas error: " + err.Error()), nil
	}
	path, err := canvasFilePath(id)
	if err != nil {
		return errorResult("delete canvas error: " + err.Error()), nil
	}
	if err = os.Remove(path); err != nil {
		return errorResult("delete canvas error: " + err.Error()), nil
	}
	return structuredTextResult("canvas deleted: "+id, map[string]any{
		"action": "delete",
		"id":     id,
	}), nil
}

func canvasAddNode(args map[string]any) (CallToolResult, error) {
	id, err := canvasRequiredID(args)
	if err != nil {
		return errorResult("add_node canvas error: " + err.Error()), nil
	}
	node, ok := args["node"].(map[string]any)
	if !ok || node == nil {
		return errorResult("add_node canvas error: node object is required"), nil
	}
	canvas, err := readCanvas(id)
	if err != nil {
		return errorResult("add_node canvas error: " + err.Error()), nil
	}
	nextNode := normalizeCanvasNode(cloneJSONMap(node))
	nodes := canvasArray(canvas["nodes"])
	nodes = append(nodes, nextNode)
	canvas["nodes"] = nodes
	touchCanvas(canvas, id)
	if err = writeCanvas(id, canvas); err != nil {
		return errorResult("add_node canvas error: " + err.Error()), nil
	}
	return structuredTextResult("canvas node added: "+fmt.Sprintf("%v", nextNode["id"]), map[string]any{
		"action": "add_node",
		"id":     id,
		"node":   nextNode,
		"canvas": canvas,
	}), nil
}

func canvasAddScribliNode(args map[string]any) (CallToolResult, error) {
	id, err := canvasRequiredID(args)
	if err != nil {
		return errorResult("add_scribli_node canvas error: " + err.Error()), nil
	}
	node, err := buildScribliCanvasNode(args)
	if err != nil {
		return errorResult("add_scribli_node canvas error: " + err.Error()), nil
	}
	canvas, err := readCanvas(id)
	if err != nil {
		return errorResult("add_scribli_node canvas error: " + err.Error()), nil
	}
	nodes := canvasArray(canvas["nodes"])
	nodes = append(nodes, node)
	canvas["nodes"] = nodes
	touchCanvas(canvas, id)
	if err = writeCanvas(id, canvas); err != nil {
		return errorResult("add_scribli_node canvas error: " + err.Error()), nil
	}
	return structuredTextResult("scribli canvas node added: "+fmt.Sprintf("%v", node["id"]), map[string]any{
		"action": "add_scribli_node",
		"id":     id,
		"node":   node,
		"canvas": canvas,
	}), nil
}

func canvasAddEdge(args map[string]any) (CallToolResult, error) {
	id, err := canvasRequiredID(args)
	if err != nil {
		return errorResult("add_edge canvas error: " + err.Error()), nil
	}
	edge, ok := args["edge"].(map[string]any)
	if !ok || edge == nil {
		return errorResult("add_edge canvas error: edge object is required"), nil
	}
	canvas, err := readCanvas(id)
	if err != nil {
		return errorResult("add_edge canvas error: " + err.Error()), nil
	}
	nextEdge := normalizeCanvasEdge(cloneJSONMap(edge))
	if nextEdge["fromNode"] == "" || nextEdge["toNode"] == "" {
		return errorResult("add_edge canvas error: edge requires fromNode and toNode"), nil
	}
	edges := canvasArray(canvas["edges"])
	edges = append(edges, nextEdge)
	canvas["edges"] = edges
	touchCanvas(canvas, id)
	if err = writeCanvas(id, canvas); err != nil {
		return errorResult("add_edge canvas error: " + err.Error()), nil
	}
	return structuredTextResult("canvas edge added: "+fmt.Sprintf("%v", nextEdge["id"]), map[string]any{
		"action": "add_edge",
		"id":     id,
		"edge":   nextEdge,
		"canvas": canvas,
	}), nil
}

func canvasRequiredID(args map[string]any) (string, error) {
	id, _ := args["id"].(string)
	id = strings.TrimSpace(id)
	if id == "" {
		return "", errors.New("id is required")
	}
	if err := validateCanvasID(id); err != nil {
		return "", err
	}
	return id, nil
}

type canvasEmbed struct {
	Position   string
	ParentID   string
	PreviousID string
	NextID     string
}

func (embed canvasEmbed) toMap() map[string]any {
	return map[string]any{
		"position":   embed.Position,
		"parentID":   embed.ParentID,
		"previousID": embed.PreviousID,
		"nextID":     embed.NextID,
	}
}

func canvasEmbedOperation(args map[string]any) (canvasEmbed, error) {
	position, _ := args["embedPosition"].(string)
	position = strings.TrimSpace(position)
	if position == "" {
		position = "append"
	}
	if !map[string]bool{"append": true, "prepend": true, "insert": true}[position] {
		return canvasEmbed{}, errors.New("embedPosition must be one of: append, prepend, insert")
	}

	parentID, _ := args["embedParentID"].(string)
	previousID, _ := args["embedPreviousID"].(string)
	nextID, _ := args["embedNextID"].(string)
	embed := canvasEmbed{
		Position:   position,
		ParentID:   strings.TrimSpace(parentID),
		PreviousID: strings.TrimSpace(previousID),
		NextID:     strings.TrimSpace(nextID),
	}

	switch position {
	case "append", "prepend":
		if embed.ParentID == "" {
			return canvasEmbed{}, errors.New("embedParentID is required for append/prepend")
		}
		if err := checkCanvasEmbedContainerParent(embed.ParentID); err != nil {
			return canvasEmbed{}, err
		}
	case "insert":
		if embed.ParentID == "" && embed.PreviousID == "" && embed.NextID == "" {
			return canvasEmbed{}, errors.New("embedParentID, embedPreviousID, or embedNextID is required for insert")
		}
		if embed.ParentID != "" && embed.PreviousID == "" && embed.NextID == "" {
			if err := checkCanvasEmbedContainerParent(embed.ParentID); err != nil {
				return canvasEmbed{}, err
			}
		}
		if embed.PreviousID != "" {
			if err := validateExistingCanvasEmbedSibling(embed.PreviousID, "embedPreviousID"); err != nil {
				return canvasEmbed{}, err
			}
		}
		if embed.NextID != "" {
			if err := validateExistingCanvasEmbedSibling(embed.NextID, "embedNextID"); err != nil {
				return canvasEmbed{}, err
			}
		}
	}
	return embed, nil
}

func checkCanvasEmbedContainerParent(id string) (err error) {
	if !ast.IsNodeIDPattern(id) {
		return errors.New("embedParentID must be a valid Scribli ID")
	}
	defer func() {
		if recover() != nil {
			err = fmt.Errorf("parent block not found: %s", id)
		}
	}()
	return treenode.CheckContainerParent(id)
}

func validateExistingCanvasEmbedSibling(id, field string) error {
	if !ast.IsNodeIDPattern(id) {
		return fmt.Errorf("%s must be a valid Scribli ID", field)
	}
	found := false
	func() {
		defer func() {
			if recover() != nil {
				found = false
			}
		}()
		found = treenode.GetBlockTree(id) != nil
	}()
	if !found {
		return fmt.Errorf("%s block not found: %s", field, id)
	}
	return nil
}

func insertCanvasEmbedBlock(canvasID string, embed canvasEmbed) (blockID, markdown string, err error) {
	blockID = ast.NewNodeID()
	markdown = canvasEmbedMarkdown(canvasID)
	data := pinBlockID(markdown, "markdown", blockID)

	op := &model.Operation{
		Data: data,
	}
	switch embed.Position {
	case "append":
		op.Action = "appendInsert"
		op.ParentID = embed.ParentID
	case "prepend":
		op.Action = "prependInsert"
		op.ParentID = embed.ParentID
	case "insert":
		op.Action = "insert"
		op.ParentID = embed.ParentID
		op.PreviousID = embed.PreviousID
		op.NextID = embed.NextID
	default:
		return "", "", errors.New("embedPosition must be one of: append, prepend, insert")
	}

	transactions := []*model.Transaction{{
		DoOperations: []*model.Operation{op},
	}}
	model.PerformTransactions(&transactions)
	model.FlushTxQueue()

	reloadID := embed.ParentID
	if reloadID == "" {
		reloadID = embed.PreviousID
	}
	if reloadID == "" {
		reloadID = embed.NextID
	}
	if reloadID != "" {
		if bt := treenode.GetBlockTree(reloadID); bt != nil {
			util.PushReloadProtyle(bt.RootID)
		}
	}

	return blockID, markdown, nil
}

func canvasEmbedMarkdown(id string) string {
	return "```scribli-canvas\n" + id + "\n```"
}

func validateCanvasID(id string) error {
	if !ast.IsNodeIDPattern(id) {
		return errors.New("id must be a valid Scribli ID")
	}
	return nil
}

func canvasStorageDir() (string, error) {
	if strings.TrimSpace(util.DataDir) == "" {
		return "", errors.New("workspace data directory is not available")
	}
	dir := filepath.Join(util.DataDir, "storage", "canvas")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return dir, nil
}

func canvasFilePath(id string) (string, error) {
	if err := validateCanvasID(id); err != nil {
		return "", err
	}
	dir, err := canvasStorageDir()
	if err != nil {
		return "", err
	}
	path := filepath.Join(dir, id+".canvas")
	cleanDir := filepath.Clean(dir) + string(os.PathSeparator)
	cleanPath := filepath.Clean(path)
	if !strings.HasPrefix(cleanPath, cleanDir) {
		return "", errors.New("resolved canvas path escapes storage directory")
	}
	return cleanPath, nil
}

func readCanvas(id string) (map[string]any, error) {
	path, err := canvasFilePath(id)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var canvas map[string]any
	if err = json.Unmarshal(data, &canvas); err != nil {
		return nil, err
	}
	normalizeCanvas(canvas)
	return canvas, nil
}

func writeCanvas(id string, canvas map[string]any) error {
	path, err := canvasFilePath(id)
	if err != nil {
		return err
	}
	normalizeCanvas(canvas)
	data, err := json.MarshalIndent(canvas, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0644)
}

func normalizeCanvas(canvas map[string]any) {
	if canvas["nodes"] == nil {
		canvas["nodes"] = []any{}
	}
	if canvas["edges"] == nil {
		canvas["edges"] = []any{}
	}
}

func normalizeCanvasNode(node map[string]any) map[string]any {
	if strings.TrimSpace(fmt.Sprintf("%v", node["id"])) == "" {
		node["id"] = ast.NewNodeID()
	}
	if strings.TrimSpace(fmt.Sprintf("%v", node["type"])) == "" || fmt.Sprintf("%v", node["type"]) == "<nil>" {
		node["type"] = "text"
	}
	defaultNumber(node, "x", 0)
	defaultNumber(node, "y", 0)
	defaultNumber(node, "width", 320)
	defaultNumber(node, "height", 180)
	return node
}

func normalizeCanvasEdge(edge map[string]any) map[string]any {
	if strings.TrimSpace(fmt.Sprintf("%v", edge["id"])) == "" {
		edge["id"] = ast.NewNodeID()
	}
	return edge
}

func buildScribliCanvasNode(args map[string]any) (map[string]any, error) {
	kind, _ := args["kind"].(string)
	kind = strings.TrimSpace(kind)
	if kind == "" {
		return nil, errors.New("kind is required")
	}
	if !map[string]bool{
		"document":          true,
		"block":             true,
		"asset":             true,
		"database":          true,
		"query":             true,
		"executable_output": true,
	}[kind] {
		return nil, errors.New("kind must be one of: document, block, asset, database, query, executable_output")
	}

	label, _ := args["label"].(string)
	label = strings.TrimSpace(label)
	text, _ := args["text"].(string)
	text = strings.TrimSpace(text)
	refID, _ := args["refID"].(string)
	refID = strings.TrimSpace(refID)
	query, _ := args["query"].(string)
	query = strings.TrimSpace(query)
	assetPath, _ := args["assetPath"].(string)
	assetPath = strings.TrimSpace(assetPath)
	databaseID, _ := args["databaseID"].(string)
	databaseID = strings.TrimSpace(databaseID)
	viewID, _ := args["viewID"].(string)
	viewID = strings.TrimSpace(viewID)

	node := map[string]any{
		"id":     ast.NewNodeID(),
		"type":   "text",
		"x":      numberArg(args, "x", 0),
		"y":      numberArg(args, "y", 0),
		"width":  numberArg(args, "width", 320),
		"height": numberArg(args, "height", 180),
	}

	switch kind {
	case "asset":
		if assetPath == "" {
			return nil, errors.New("assetPath is required for asset nodes")
		}
		node["type"] = "file"
		node["file"] = assetPath
	case "database":
		if databaseID == "" {
			return nil, errors.New("databaseID is required for database nodes")
		}
	case "query":
		if query == "" {
			return nil, errors.New("query is required for query nodes")
		}
	case "document", "block", "executable_output":
		if refID == "" {
			return nil, errors.New("refID is required for " + kind + " nodes")
		}
	}

	if label == "" {
		label = kind
		if refID != "" {
			label += ": " + refID
		} else if databaseID != "" {
			label += ": " + databaseID
		} else if query != "" {
			label += ": " + query
		} else if assetPath != "" {
			label += ": " + assetPath
		}
	}
	if node["type"] == "text" {
		if text == "" {
			text = label
		}
		node["text"] = text
	}
	node["scribli"] = map[string]any{
		"kind":       kind,
		"refID":      refID,
		"label":      label,
		"query":      query,
		"assetPath":  assetPath,
		"databaseID": databaseID,
		"viewID":     viewID,
	}
	return node, nil
}

func defaultNumber(m map[string]any, key string, value float64) {
	if _, ok := m[key]; !ok || m[key] == nil {
		m[key] = value
	}
}

func numberArg(args map[string]any, key string, fallback float64) float64 {
	switch v := args[key].(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case json.Number:
		f, err := v.Float64()
		if err == nil {
			return f
		}
	}
	return fallback
}

func canvasArray(raw any) []any {
	if raw == nil {
		return []any{}
	}
	if arr, ok := raw.([]any); ok {
		return arr
	}
	return []any{}
}

func touchCanvas(canvas map[string]any, id string) {
	canvas["scribli"] = mergeScribliCanvasMetadata(canvas["scribli"], map[string]any{
		"id":      id,
		"format":  "json-canvas",
		"updated": time.Now().UTC().Format(time.RFC3339),
	})
}

func mergeScribliCanvasMetadata(raw any, updates map[string]any) map[string]any {
	metadata, _ := raw.(map[string]any)
	next := cloneJSONMap(metadata)
	for k, v := range updates {
		next[k] = v
	}
	return next
}

func canvasSummary(id string, canvas map[string]any) map[string]any {
	metadata, _ := canvas["scribli"].(map[string]any)
	return map[string]any{
		"id":      id,
		"title":   metadata["title"],
		"updated": metadata["updated"],
		"nodes":   len(canvasArray(canvas["nodes"])),
		"edges":   len(canvasArray(canvas["edges"])),
	}
}

func cloneJSONMap(m map[string]any) map[string]any {
	next := map[string]any{}
	for k, v := range m {
		next[k] = v
	}
	return next
}
