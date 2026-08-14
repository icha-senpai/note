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

package tools

import (
	"fmt"
	"strings"

	"github.com/icha-senpai/note/kernel/filesys"
	"github.com/icha-senpai/note/kernel/model"
	"github.com/icha-senpai/note/kernel/treenode"
	"github.com/icha-senpai/note/kernel/util"
	"github.com/icha-senpai/note/third_party/forks/lute/ast"
)

var BlockTool = &Tool{
	Name:        "block",
	Description: "Block operations. Actions: get(id), get_kramdown(id), get_children(id), tree_stat(id, by document), dom(id), insert(data, dataType, parentID?, nextID?, previousID?), append(data, dataType, parentID) / prepend(...) add a NEW child — use after block.update when both modifying and adding, update(id, data, dataType) replaces ONE block only (no append), delete(id), move(id, parentID, previousID?), breadcrumb(id), batch_get(ids) / batch_kramdown(ids) where ids is comma-separated.",
	InputSchema: ToolSchema{
		Type: "object",
		Properties: map[string]Property{
			"action":     {Type: "string", Description: "Operation", Enum: []string{"get", "get_kramdown", "get_children", "tree_stat", "dom", "insert", "append", "prepend", "update", "delete", "move", "breadcrumb", "batch_get", "batch_kramdown"}},
			"id":         {Type: "string", Description: "Block ID"},
			"ids":        {Type: "string", Description: "Comma-separated block IDs (for batch_get, batch_kramdown)"},
			"data":       {Type: "string", Description: "Content (markdown or dom)"},
			"dataType":   {Type: "string", Description: "Content type: markdown or dom", Enum: []string{"markdown", "dom"}},
			"parentID":   {Type: "string", Description: "Parent block ID"},
			"nextID":     {Type: "string", Description: "Next sibling block ID (for insert)"},
			"previousID": {Type: "string", Description: "Previous sibling block ID (for insert)"},
			"redactSecrets": {
				Type:        "boolean",
				Description: "Redact likely credentials/tokens from read responses. Defaults to true.",
			},
		},
		Required: []string{"action"},
	},
	OutputSchema: structuredOutputSchema(),
	EffectScope:  EffectScopeLocal,
	ActionEffects: mergeEffectMaps(
		effectMap(ToolEffects{LocalRead: true}, "get", "get_kramdown", "get_children", "tree_stat", "dom", "breadcrumb", "batch_get", "batch_kramdown"),
		effectMap(ToolEffects{LocalWrite: true}, "insert", "append", "prepend", "update", "delete", "move"),
	),
	Handler: blockHandler,
}

func init() {
	register(BlockTool)
}

func blockHandler(args map[string]any) (CallToolResult, error) {
	action, _ := args["action"].(string)
	switch action {
	case "get":
		return blockGet(args)
	case "get_kramdown":
		return blockGetKramdown(args)
	case "get_children":
		return blockGetChildren(args)
	case "tree_stat":
		return blockTreeStat(args)
	case "dom":
		return blockDom(args)
	case "insert":
		return blockInsert(args)
	case "append":
		return blockAppend(args)
	case "prepend":
		return blockPrepend(args)
	case "update":
		return blockUpdate(args)
	case "delete":
		return blockDelete(args)
	case "move":
		return blockMove(args)
	case "breadcrumb":
		return blockBreadcrumb(args)
	case "batch_get":
		return blockBatchGet(args)
	case "batch_kramdown":
		return blockBatchKramdown(args)
	}
	return CallToolResult{
		Content: []ContentItem{{Type: "text", Text: "unknown action '" + action + "', expected one of: [get, get_kramdown, get_children, tree_stat, dom, insert, append, prepend, update, delete, move, breadcrumb, batch_get, batch_kramdown]"}},
		IsError: true,
	}, nil
}

func blockGet(args map[string]any) (CallToolResult, error) {
	id, _ := args["id"].(string)
	if id == "" {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "id is required"}}, IsError: true}, nil
	}

	b, err := model.GetBlock(id, nil)
	if err != nil {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: fmt.Sprintf("get block failed: %s", err)}}, IsError: true}, nil
	}
	if b == nil {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "block not found: " + id}}, IsError: true}, nil
	}

	created := b.Created
	if created == "" {
		created = createdFromID(b.ID)
	}
	updated := b.Updated
	if updated == "" {
		updated = ialUpdated(b.IAL, model.GetBlockKramdown(id, "md"))
	}
	if updated == "" {
		updated = created
	}
	content := maybeRedactSensitiveText(args, b.Content)
	markdown := maybeRedactSensitiveText(args, b.Markdown)
	text := fmt.Sprintf(
		"ID: %s\nType: %s\nHPath: %s\nContent: %s\nMarkdown: %s\nTags: %s\nCreated: %s\nUpdated: %s",
		b.ID, b.Type, b.HPath, content, markdown, b.Tag, created, updated,
	)
	return structuredTextResult(text, map[string]any{
		"action":   "get",
		"id":       b.ID,
		"type":     b.Type,
		"hPath":    b.HPath,
		"content":  content,
		"markdown": markdown,
		"tags":     b.Tag,
		"created":  created,
		"updated":  updated,
	}), nil
}

func blockGetKramdown(args map[string]any) (CallToolResult, error) {
	id, _ := args["id"].(string)
	if id == "" {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "id is required"}}, IsError: true}, nil
	}

	kramdown := maybeRedactSensitiveText(args, model.GetBlockKramdown(id, "md"))
	if kramdown == "" {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "block not found or empty: " + id}}, IsError: true}, nil
	}

	return structuredTextResult(kramdown, map[string]any{
		"action":   "get_kramdown",
		"id":       id,
		"kramdown": kramdown,
	}), nil
}

func blockGetChildren(args map[string]any) (CallToolResult, error) {
	id, _ := args["id"].(string)
	if id == "" {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "id is required"}}, IsError: true}, nil
	}

	children := model.GetChildBlocks(id)
	if len(children) == 0 {
		return structuredTextResult("no child blocks found", map[string]any{
			"action":   "get_children",
			"id":       id,
			"count":    0,
			"children": []map[string]any{},
		}), nil
	}

	var sb strings.Builder
	items := make([]map[string]any, 0, len(children))
	for _, c := range children {
		content := c.Markdown
		if content == "" {
			content = c.Content
		}
		displayContent := maybeRedactSensitiveText(args, c.Content)
		displayMarkdown := maybeRedactSensitiveText(args, c.Markdown)
		content = maybeRedactSensitiveText(args, content)
		items = append(items, map[string]any{
			"id":       c.ID,
			"type":     c.Type,
			"subType":  c.SubType,
			"content":  displayContent,
			"markdown": displayMarkdown,
		})
		if len(content) > 200 {
			content = content[:200] + "..."
		}
		sb.WriteString(fmt.Sprintf("- [%s] %s (%s)\n", c.Type, content, c.ID))
	}
	return structuredTextResult(sb.String(), map[string]any{
		"action":   "get_children",
		"id":       id,
		"count":    len(items),
		"children": items,
	}), nil
}

func blockInsert(args map[string]any) (CallToolResult, error) {
	data, dataType := getBlockData(args)
	if data == "" {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "data is required"}}, IsError: true}, nil
	}

	var parentID, previousID, nextID string
	if v, ok := args["parentID"].(string); ok {
		parentID = v
	}
	if v, ok := args["previousID"].(string); ok {
		previousID = v
	}
	if v, ok := args["nextID"].(string); ok {
		nextID = v
	}

	if parentID != "" && previousID == "" && nextID == "" {
		if err := treenode.CheckContainerParent(parentID); err != nil {
			return CallToolResult{Content: []ContentItem{{Type: "text", Text: err.Error()}}, IsError: true}, nil
		}
	}

	if dataType == "markdown" {
		var err error
		data, err = markdownToBlockDOM(data)
		if err != nil {
			return CallToolResult{Content: []ContentItem{{Type: "text", Text: "convert markdown failed: " + err.Error()}}, IsError: true}, nil
		}
	}

	transactions := []*model.Transaction{{
		DoOperations: []*model.Operation{{
			Action:     "insert",
			Data:       data,
			ParentID:   parentID,
			PreviousID: previousID,
			NextID:     nextID,
		}},
	}}

	model.PerformTransactions(&transactions)
	model.FlushTxQueue()

	reloadID := nextID
	if reloadID == "" {
		reloadID = previousID
	}
	if reloadID == "" {
		reloadID = parentID
	}
	if reloadID != "" {
		if bt := treenode.GetBlockTree(reloadID); bt != nil {
			util.PushReloadProtyle(bt.RootID)
		}
	}

	return structuredTextResult("block inserted", map[string]any{
		"action":     "insert",
		"parentID":   parentID,
		"previousID": previousID,
		"nextID":     nextID,
	}), nil
}

func blockAppend(args map[string]any) (CallToolResult, error) {
	data, dataType := getBlockData(args)
	if data == "" {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "data is required"}}, IsError: true}, nil
	}
	parentID, _ := args["parentID"].(string)
	if parentID == "" {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "parentID is required"}}, IsError: true}, nil
	}

	if err := treenode.CheckContainerParent(parentID); err != nil {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: err.Error()}}, IsError: true}, nil
	}

	if dataType == "markdown" {
		var err error
		data, err = markdownToBlockDOM(data)
		if err != nil {
			return CallToolResult{Content: []ContentItem{{Type: "text", Text: "convert markdown failed: " + err.Error()}}, IsError: true}, nil
		}
	}

	transactions := []*model.Transaction{{
		DoOperations: []*model.Operation{{
			Action:   "appendInsert",
			Data:     data,
			ParentID: parentID,
		}},
	}}

	model.PerformTransactions(&transactions)
	model.FlushTxQueue()

	if bt := treenode.GetBlockTree(parentID); bt != nil {
		util.PushReloadProtyle(bt.RootID)
	}
	return structuredTextResult("block appended", map[string]any{
		"action":   "append",
		"parentID": parentID,
	}), nil
}

func blockPrepend(args map[string]any) (CallToolResult, error) {
	data, dataType := getBlockData(args)
	if data == "" {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "data is required"}}, IsError: true}, nil
	}
	parentID, _ := args["parentID"].(string)
	if parentID == "" {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "parentID is required"}}, IsError: true}, nil
	}

	if err := treenode.CheckContainerParent(parentID); err != nil {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: err.Error()}}, IsError: true}, nil
	}

	if dataType == "markdown" {
		var err error
		data, err = markdownToBlockDOM(data)
		if err != nil {
			return CallToolResult{Content: []ContentItem{{Type: "text", Text: "convert markdown failed: " + err.Error()}}, IsError: true}, nil
		}
	}

	transactions := []*model.Transaction{{
		DoOperations: []*model.Operation{{
			Action:   "prependInsert",
			Data:     data,
			ParentID: parentID,
		}},
	}}

	model.PerformTransactions(&transactions)
	model.FlushTxQueue()

	if bt := treenode.GetBlockTree(parentID); bt != nil {
		util.PushReloadProtyle(bt.RootID)
	}
	return structuredTextResult("block prepended", map[string]any{
		"action":   "prepend",
		"parentID": parentID,
	}), nil
}

func blockUpdate(args map[string]any) (CallToolResult, error) {
	id, _ := args["id"].(string)
	if id == "" {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "id is required"}}, IsError: true}, nil
	}
	data, dataType := getBlockData(args)
	if data == "" {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "data is required"}}, IsError: true}, nil
	}

	operationCount, err := updateBlockContent(id, data, dataType)
	if err != nil {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "update block failed: " + err.Error()}}, IsError: true}, nil
	}
	return structuredTextResult("block updated", map[string]any{
		"action":           "update",
		"id":               id,
		"operationCount":   operationCount,
		"inputContentType": dataType,
	}), nil
}

func updateBlockContent(id, data, dataType string) (operationCount int, err error) {
	luteEngine := util.NewLute()
	if dataType == "markdown" {
		data, err = markdownToBlockDOM(data)
		if err != nil {
			return 0, fmt.Errorf("convert markdown failed: %w", err)
		}
	}
	tree := luteEngine.BlockDOM2Tree(data)
	if nil == tree || nil == tree.Root || nil == tree.Root.FirstChild {
		return 0, fmt.Errorf("parse tree failed")
	}

	block, err := model.GetBlock(id, nil)
	if err != nil {
		return 0, fmt.Errorf("get block failed: %w", err)
	}
	if block == nil {
		return 0, fmt.Errorf("block not found: %s", id)
	}

	var transactions []*model.Transaction
	if "NodeDocument" == block.Type {
		oldTree, loadErr := filesys.LoadTree(block.Box, block.Path, luteEngine)
		if loadErr != nil {
			return 0, fmt.Errorf("load tree failed: %w", loadErr)
		}
		var toRemoves []*ast.Node
		var ops []*model.Operation
		for n := oldTree.Root.FirstChild; nil != n; n = n.Next {
			toRemoves = append(toRemoves, n)
			ops = append(ops, &model.Operation{Action: "delete", ID: n.ID, Data: map[string]any{
				"createEmptyParagraph": false,
			}})
		}
		for _, n := range toRemoves {
			n.Unlink()
		}
		ops = append(ops, &model.Operation{Action: "appendInsert", Data: data, ParentID: id})
		transactions = append(transactions, &model.Transaction{DoOperations: ops})
		operationCount = len(ops)
	} else {
		if "NodeListItem" == block.Type && ast.NodeList == tree.Root.FirstChild.Type {
			tree.Root.AppendChild(tree.Root.FirstChild.FirstChild)
			tree.Root.FirstChild.Unlink()
			if nil != tree.Root.FirstChild && ast.NodeKramdownBlockIAL == tree.Root.FirstChild.Type {
				tree.Root.FirstChild.Unlink()
			}
		}

		if nil != tree.Root.FirstChild {
			tree.Root.FirstChild.SetIALAttr("id", id)
		} else {
			tree.Root.AppendChild(treenode.NewParagraph(id))
		}

		data = luteEngine.Tree2BlockDOM(tree, luteEngine.RenderOptions, luteEngine.ParseOptions)
		transactions = []*model.Transaction{{
			DoOperations: []*model.Operation{{
				Action: "update",
				ID:     id,
				Data:   data,
			}},
		}}
		operationCount = 1
	}

	model.PerformTransactions(&transactions)
	model.FlushTxQueue()

	if bt := treenode.GetBlockTree(id); bt != nil {
		util.PushReloadProtyle(bt.RootID)
	}
	return operationCount, nil
}

func pinBlockID(data, dataType, id string) string {
	luteEngine := util.NewLute()
	if dataType == "markdown" {
		var err error
		data, err = markdownToBlockDOM(data)
		if err != nil {
			return data
		}
	}

	tree := luteEngine.BlockDOM2Tree(data)
	if nil == tree || nil == tree.Root || nil == tree.Root.FirstChild {
		return data
	}

	if ast.NodeList == tree.Root.FirstChild.Type {
		tree.Root.AppendChild(tree.Root.FirstChild.FirstChild)
		tree.Root.FirstChild.Unlink()
		if nil != tree.Root.FirstChild && ast.NodeKramdownBlockIAL == tree.Root.FirstChild.Type {
			tree.Root.FirstChild.Unlink()
		}
	}

	if nil != tree.Root.FirstChild {
		tree.Root.FirstChild.SetIALAttr("id", id)
	} else {
		tree.Root.AppendChild(treenode.NewParagraph(id))
	}

	return luteEngine.Tree2BlockDOM(tree, luteEngine.RenderOptions, luteEngine.ParseOptions)
}

func blockDelete(args map[string]any) (CallToolResult, error) {
	id, _ := args["id"].(string)
	if id == "" {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "id is required"}}, IsError: true}, nil
	}

	bt := treenode.GetBlockTree(id)

	transactions := []*model.Transaction{{
		DoOperations: []*model.Operation{{
			Action: "delete",
			ID:     id,
		}},
	}}

	model.PerformTransactions(&transactions)
	model.FlushTxQueue()

	if bt != nil {
		util.PushReloadProtyle(bt.RootID)
	}
	return structuredTextResult("block deleted: "+id, map[string]any{
		"action": "delete",
		"id":     id,
	}), nil
}

func getBlockData(args map[string]any) (data, dataType string) {
	data, _ = args["data"].(string)
	dataType, _ = args["dataType"].(string)
	if dataType == "" {
		dataType = "markdown"
	}
	return
}

func markdownToBlockDOM(md string) (string, error) {
	luteEngine := util.NewLute()
	luteEngine.SetHTMLTag2TextMark(true)
	result, _ := luteEngine.Md2BlockDOMTree(md, true)
	if result == "" {
		return "", fmt.Errorf("empty result")
	}
	return result, nil
}

func blockMove(args map[string]any) (CallToolResult, error) {
	id, _ := args["id"].(string)
	if id == "" {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "id is required"}}, IsError: true}, nil
	}
	parentID, _ := args["parentID"].(string)
	if parentID == "" {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "parentID is required"}}, IsError: true}, nil
	}
	previousID, _ := args["previousID"].(string)

	if previousID == "" {
		if err := treenode.CheckListItemNesting(parentID, id); err != nil {
			return CallToolResult{Content: []ContentItem{{Type: "text", Text: err.Error()}}, IsError: true}, nil
		}
		if err := treenode.CheckContainerParent(parentID); err != nil {
			return CallToolResult{Content: []ContentItem{{Type: "text", Text: err.Error()}}, IsError: true}, nil
		}
	}

	transactions := []*model.Transaction{{
		DoOperations: []*model.Operation{{
			Action:     "move",
			ID:         id,
			ParentID:   parentID,
			PreviousID: previousID,
		}},
	}}

	model.PerformTransactions(&transactions)
	model.FlushTxQueue()

	if bt := treenode.GetBlockTree(id); bt != nil {
		util.PushReloadProtyle(bt.RootID)
	}
	return structuredTextResult("block moved: "+id, map[string]any{
		"action":     "move",
		"id":         id,
		"parentID":   parentID,
		"previousID": previousID,
	}), nil
}

func blockBreadcrumb(args map[string]any) (CallToolResult, error) {
	id, _ := args["id"].(string)
	if id == "" {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "id is required"}}, IsError: true}, nil
	}

	paths, err := model.BuildBlockBreadcrumb(id, nil)
	if err != nil {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "breadcrumb failed: " + err.Error()}}, IsError: true}, nil
	}

	var sb strings.Builder
	items := make([]map[string]any, 0, len(paths))
	for _, p := range paths {
		sb.WriteString(fmt.Sprintf("%s/%s (%s)\n", p.Type, p.Name, p.ID))
		items = append(items, map[string]any{"id": p.ID, "type": p.Type, "name": p.Name})
	}
	return structuredTextResult(sb.String(), map[string]any{
		"action":     "breadcrumb",
		"id":         id,
		"breadcrumb": items,
	}), nil
}

func blockTreeStat(args map[string]any) (CallToolResult, error) {
	id, _ := args["id"].(string)
	if id == "" {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "id is required"}}, IsError: true}, nil
	}
	stat := filesys.StatTree(id)
	if stat == nil {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "document not found or empty"}}, IsError: true}, nil
	}
	text := fmt.Sprintf("Document statistics:\n- Characters: %d\n- Words: %d\n- Blocks: %d\n- Links: %d\n- Images: %d\n- Refs: %d",
		stat.RuneCount, stat.WordCount, stat.BlockCount, stat.LinkCount, stat.ImageCount, stat.RefCount)
	return structuredTextResult(text, map[string]any{
		"action":     "tree_stat",
		"id":         id,
		"characters": stat.RuneCount,
		"words":      stat.WordCount,
		"blocks":     stat.BlockCount,
		"links":      stat.LinkCount,
		"images":     stat.ImageCount,
		"refs":       stat.RefCount,
	}), nil
}

func blockDom(args map[string]any) (CallToolResult, error) {
	id, _ := args["id"].(string)
	if id == "" {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "id is required"}}, IsError: true}, nil
	}
	dom := maybeRedactSensitiveText(args, model.GetBlockDOM(id))
	if dom == "" {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "block not found or empty: " + id}}, IsError: true}, nil
	}
	return structuredTextResult(dom, map[string]any{
		"action": "dom",
		"id":     id,
		"dom":    dom,
	}), nil
}

func blockBatchGet(args map[string]any) (CallToolResult, error) {
	idsStr, _ := args["ids"].(string)
	if idsStr == "" {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "ids is required (comma-separated)"}}, IsError: true}, nil
	}

	ids := strings.Split(idsStr, ",")
	for i := range ids {
		ids[i] = strings.TrimSpace(ids[i])
	}

	if len(ids) == 0 {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "no valid IDs provided"}}, IsError: true}, nil
	}

	infos := model.GetDocsInfo(ids, false, false)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Batch get %d blocks (found %d):\n\n", len(ids), len(infos)))
	for _, info := range infos {
		sb.WriteString(fmt.Sprintf("- %s: %s (rootID: %s, refCount: %d)\n", info.ID, info.Name, info.RootID, info.RefCount))
	}
	for _, id := range ids {
		found := false
		for _, info := range infos {
			if info.ID == id {
				found = true
				break
			}
		}
		if !found {
			sb.WriteString(fmt.Sprintf("- %s: not found\n", id))
		}
	}

	items := make([]map[string]any, 0, len(ids))
	for _, id := range ids {
		item := map[string]any{"id": id, "found": false}
		for _, info := range infos {
			if info.ID == id {
				item["found"] = true
				item["name"] = info.Name
				item["rootID"] = info.RootID
				item["refCount"] = info.RefCount
				break
			}
		}
		items = append(items, item)
	}
	return structuredTextResult(sb.String(), map[string]any{
		"action": "batch_get",
		"ids":    ids,
		"count":  len(infos),
		"items":  items,
	}), nil
}

func blockBatchKramdown(args map[string]any) (CallToolResult, error) {
	idsStr, _ := args["ids"].(string)
	if idsStr == "" {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "ids is required (comma-separated)"}}, IsError: true}, nil
	}

	ids := strings.Split(idsStr, ",")
	for i := range ids {
		ids[i] = strings.TrimSpace(ids[i])
	}

	if len(ids) == 0 {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "no valid IDs provided"}}, IsError: true}, nil
	}

	kramdowns := model.GetBlockKramdowns(ids, "md")

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Batch kramdown %d blocks (found %d):\n\n", len(ids), len(kramdowns)))
	for _, id := range ids {
		if kd, ok := kramdowns[id]; ok {
			kd = maybeRedactSensitiveText(args, kd)
			sb.WriteString(fmt.Sprintf("--- %s ---\n%s\n\n", id, kd))
		} else {
			sb.WriteString(fmt.Sprintf("--- %s ---\n(not found)\n\n", id))
		}
	}

	items := make([]map[string]any, 0, len(ids))
	for _, id := range ids {
		kramdown, found := kramdowns[id]
		kramdown = maybeRedactSensitiveText(args, kramdown)
		items = append(items, map[string]any{"id": id, "found": found, "kramdown": kramdown})
	}
	return structuredTextResult(sb.String(), map[string]any{
		"action": "batch_kramdown",
		"ids":    ids,
		"count":  len(kramdowns),
		"items":  items,
	}), nil
}
