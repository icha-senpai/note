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
	"sort"
	"strings"

	"github.com/icha-senpai/note/kernel/model"
	"github.com/icha-senpai/note/kernel/treenode"
	"github.com/icha-senpai/note/kernel/util"
	"github.com/icha-senpai/note/third_party/forks/lute/ast"
)

var DocumentTool = &Tool{
	Name:        "document",
	Description: "Document operations. Actions: get(id), create(notebook, path=hPath, title, markdown?), update(id, markdown), list(notebook, path=hPath default /), get_tree(id or notebook+path, depth?, includeContent?, offset?, maxChars?, maxDocs?) bulk-fetches a document and its descendant documents' kramdown in one call, recent(notebook?, since?, limit?, order?) lists documents by last-updated time across open notebooks, delete(id), rename(id, title), move(id, notebook, path=target hPath), duplicate(id), search_docs(keyword), info(id).",
	InputSchema: ToolSchema{
		Type: "object",
		Properties: map[string]Property{
			"action":   {Type: "string", Description: "Operation", Enum: []string{"get", "create", "update", "list", "get_tree", "recent", "delete", "rename", "move", "duplicate", "search_docs", "info"}},
			"since":    {Type: "string", Description: "Only include documents updated at or after this timestamp (YYYYMMDDHHMMSS; separators allowed)"},
			"limit":    {Type: "integer", Description: "Max documents returned for recent (default 20)"},
			"order":    {Type: "string", Description: "Sort order for recent: desc (newest first, default) or asc", Enum: []string{"desc", "asc"}},
			"id":       {Type: "string", Description: "Document block ID"},
			"title":    {Type: "string", Description: "Document title (for create, rename)"},
			"path":     {Type: "string", Description: "Document hPath, the human-readable path shown in the document tree (e.g. /folder/doc). Used for create, list, get_tree, move."},
			"markdown": {Type: "string", Description: "Markdown content (for create, update)"},
			"keyword":  {Type: "string", Description: "Search keyword (for search_docs)"},
			"notebook": {Type: "string", Description: "Notebook ID (required for create, list, move; with path for get_tree)"},
			"depth":    {Type: "integer", Description: "Max descendant depth for get_tree: 0/unset = unlimited, 1 = direct children only"},
			"maxDocs":  {Type: "integer", Description: "Max documents included in get_tree"},
			"offset":   {Type: "integer", Description: "Rune offset into the bundled output for paging (get_tree)"},
			"maxChars": {Type: "integer", Description: "Maximum runes of bundled output to return (get_tree)"},
			"includeContent": {
				Type:        "boolean",
				Description: "Include each document's kramdown in get_tree. Default true; false returns a recursive listing only.",
			},
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
		effectMap(ToolEffects{LocalRead: true}, "get", "list", "get_tree", "recent", "search_docs", "info"),
		effectMap(ToolEffects{LocalWrite: true}, "create", "update", "delete", "rename", "move", "duplicate"),
	),
	Handler: documentHandler,
}

func init() {
	register(DocumentTool)
}

func documentHandler(args map[string]any) (CallToolResult, error) {
	action, _ := args["action"].(string)
	switch action {
	case "get":
		return documentGet(args)
	case "create":
		return documentCreate(args)
	case "update":
		return documentUpdate(args)
	case "list":
		return documentList(args)
	case "get_tree":
		return documentGetTree(args)
	case "recent":
		return documentRecent(args)
	case "delete":
		return documentDelete(args)
	case "rename":
		return documentRename(args)
	case "move":
		return documentMove(args)
	case "duplicate":
		return documentDuplicate(args)
	case "search_docs":
		return documentSearchDocs(args)
	case "info":
		return documentInfo(args)
	}
	return CallToolResult{
		Content: []ContentItem{{Type: "text", Text: "unknown action '" + action + "', expected one of: [get, create, update, list, get_tree, recent, delete, rename, move, duplicate, search_docs, info]"}},
		IsError: true,
	}, nil
}

func documentRecent(args map[string]any) (CallToolResult, error) {
	notebook, _ := args["notebook"].(string)
	since := normalizeSince(args["since"])
	limit := intArg(args, "limit", 20)
	if limit < 1 {
		limit = 20
	}
	order, _ := args["order"].(string)
	desc := order != "asc"

	type recentDoc struct {
		id, hPath, updated, box string
	}
	var docs []recentDoc
	for _, box := range model.Conf.GetOpenedBoxes() {
		if box == nil || (notebook != "" && box.ID != notebook) {
			continue
		}
		for _, bt := range treenode.GetBlockTreesByBoxID(box.ID) {
			if bt.Type != "d" || bt.RootID == "" {
				continue
			}
			if since != "" && bt.Updated < since {
				continue
			}
			docs = append(docs, recentDoc{id: bt.RootID, hPath: bt.HPath, updated: bt.Updated, box: box.ID})
		}
	}

	sort.Slice(docs, func(i, j int) bool {
		if desc {
			return docs[i].updated > docs[j].updated
		}
		return docs[i].updated < docs[j].updated
	})
	if len(docs) > limit {
		docs = docs[:limit]
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Recently updated documents (%d):\n\n", len(docs)))
	items := make([]map[string]any, 0, len(docs))
	for _, d := range docs {
		sb.WriteString(fmt.Sprintf("- %s (id: %s, updated: %s)\n", d.hPath, d.id, d.updated))
		items = append(items, map[string]any{
			"id": d.id, "hPath": d.hPath, "updated": d.updated, "notebook": d.box,
		})
	}
	return structuredTextResult(sb.String(), map[string]any{
		"action":   "recent",
		"notebook": notebook,
		"since":    since,
		"order":    order,
		"count":    len(items),
		"items":    items,
	}), nil
}

func normalizeSince(raw any) string {
	s, _ := raw.(string)
	if s == "" {
		return ""
	}
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	digits := b.String()
	if len(digits) > 14 {
		digits = digits[:14]
	}
	return digits
}

func documentGetTree(args map[string]any) (CallToolResult, error) {
	id, _ := args["id"].(string)
	notebook, _ := args["notebook"].(string)
	hPath, _ := args["path"].(string)

	var boxID, rootID, rootPath, rootHPath string
	if id != "" {
		tree, err := model.LoadTreeByBlockID(id)
		if err != nil {
			return CallToolResult{Content: []ContentItem{{Type: "text", Text: fmt.Sprintf("load doc failed: %s", err)}}, IsError: true}, nil
		}
		if tree == nil || tree.Root == nil {
			return CallToolResult{Content: []ContentItem{{Type: "text", Text: "document not found: " + id}}, IsError: true}, nil
		}
		boxID, rootID, rootPath, rootHPath = tree.Box, tree.Root.ID, tree.Path, tree.HPath
	} else if notebook != "" {
		boxID = notebook
		if hPath == "" || hPath == "/" {
			rootPath, rootHPath = "/", "/"
		} else {
			bt := treenode.GetBlockTreeRootByHPath(notebook, hPath)
			if bt == nil {
				return CallToolResult{Content: []ContentItem{{Type: "text", Text: "target path not found: " + hPath}}, IsError: true}, nil
			}
			rootID, rootPath, rootHPath = bt.RootID, bt.Path, bt.HPath
		}
	} else {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "id or notebook is required"}}, IsError: true}, nil
	}

	depth := intArg(args, "depth", 0)
	maxDocs := intArg(args, "maxDocs", 0)
	offset := intArg(args, "offset", 0)
	maxChars := intArg(args, "maxChars", 0)
	includeContent := true
	if v, ok := args["includeContent"].(bool); ok {
		includeContent = v
	}

	// Descendant documents live under the root's directory path prefix.
	dirPrefix := strings.TrimSuffix(rootPath, ".sy")
	if rootPath == "/" {
		dirPrefix = "/"
	}

	type docEntry struct {
		id, hPath, docPath string
		depth              int
	}
	var docs []docEntry
	seenIDs := map[string]bool{}
	if rootID != "" {
		docs = append(docs, docEntry{id: rootID, hPath: rootHPath, docPath: rootPath})
		seenIDs[rootID] = true
	}
	for _, bt := range treenode.GetBlockTreesByPathPrefix(boxID, dirPrefix) {
		if bt.Type != "d" || bt.Path == rootPath || seenIDs[bt.RootID] {
			continue
		}
		seenIDs[bt.RootID] = true
		d := strings.Count(strings.TrimPrefix(bt.Path, dirPrefix), "/")
		if depth > 0 && d > depth {
			continue
		}
		docs = append(docs, docEntry{id: bt.RootID, hPath: bt.HPath, docPath: bt.Path, depth: d})
	}
	sort.Slice(docs, func(i, j int) bool { return docs[i].docPath < docs[j].docPath })

	totalDocs := len(docs)
	if maxDocs > 0 && len(docs) > maxDocs {
		docs = docs[:maxDocs]
	}

	var kramdowns map[string]string
	if includeContent {
		ids := make([]string, 0, len(docs))
		for _, d := range docs {
			ids = append(ids, d.id)
		}
		kramdowns = model.GetBlockKramdownsInBox(ids, "md", boxID)
	}

	var sb strings.Builder
	items := make([]map[string]any, 0, len(docs))
	for _, d := range docs {
		kd := maybeRedactSensitiveText(args, kramdowns[d.id])
		sb.WriteString(fmt.Sprintf("===== %s (id: %s) =====\n\n", d.hPath, d.id))
		if includeContent {
			sb.WriteString(kd)
			sb.WriteString("\n\n")
		}
		items = append(items, map[string]any{
			"id":    d.id,
			"hPath": d.hPath,
			"depth": d.depth,
			"chars": len([]rune(kd)),
		})
	}

	bundle := sb.String()
	paged, totalChars, nextOffset := sliceRunes(bundle, offset, maxChars)
	text := paged
	if nextOffset >= 0 {
		text += fmt.Sprintf("\n[truncated: showing runes %d-%d of %d, continue with offset %d]", offset, nextOffset, totalChars, nextOffset)
	}

	return structuredTextResult(text, map[string]any{
		"action":      "get_tree",
		"rootID":      rootID,
		"rootHPath":   rootHPath,
		"notebook":    boxID,
		"documents":   items,
		"docCount":    len(items),
		"totalDocs":   totalDocs,
		"skippedDocs": totalDocs - len(items),
		"totalChars":  totalChars,
		"offset":      offset,
		"truncated":   nextOffset >= 0,
		"nextOffset":  nextOffset,
	}), nil
}

func documentGet(args map[string]any) (CallToolResult, error) {
	id, _ := args["id"].(string)
	if id == "" {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "id is required"}}, IsError: true}, nil
	}

	tree, err := model.LoadTreeByBlockID(id)
	if err != nil {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: fmt.Sprintf("load doc failed: %s", err)}}, IsError: true}, nil
	}

	b, _ := model.GetBlock(id, tree)
	if b == nil {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "document not found: " + id}}, IsError: true}, nil
	}

	info, _ := model.GetDocInfo(id)
	title := b.Name
	if info != nil && info.Name != "" {
		title = info.Name
	}
	if title == "" {
		title = tree.Root.IALAttr("title")
	}
	markdown := b.Markdown
	if markdown == "" {
		markdown = model.GetBlockKramdown(id, "md")
	}
	created := b.Created
	if created == "" {
		created = createdFromID(tree.Root.ID)
	}
	updated := b.Updated
	if updated == "" && info != nil {
		updated = ialUpdated(info.IAL, markdown)
	}
	if updated == "" {
		updated = ialUpdated(b.IAL, markdown)
	}
	if updated == "" {
		updated = created
	}
	content := b.Content
	if content == "" {
		content = title
	}
	displayContent := maybeRedactSensitiveText(args, content)
	displayMarkdown := maybeRedactSensitiveText(args, markdown)
	hPath := b.HPath
	if hPath == "" {
		hPath = tree.HPath
	}
	text := fmt.Sprintf(
		"ID: %s\nTitle: %s\nHPath: %s\nBox: %s\nContent: %s\nMarkdown: %s\nType: %s\nCreated: %s\nUpdated: %s",
		b.ID, title, hPath, b.Box, displayContent, displayMarkdown, b.Type, created, updated,
	)
	return structuredTextResult(text, map[string]any{
		"action":   "get",
		"id":       b.ID,
		"title":    title,
		"hPath":    hPath,
		"notebook": b.Box,
		"path":     b.Path,
		"content":  displayContent,
		"markdown": displayMarkdown,
		"type":     b.Type,
		"created":  created,
		"updated":  updated,
	}), nil
}

func documentCreate(args map[string]any) (CallToolResult, error) {
	notebook, _ := args["notebook"].(string)
	if notebook == "" {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "notebook is required"}}, IsError: true}, nil
	}
	hPath, _ := args["path"].(string)
	if hPath == "" {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "path is required"}}, IsError: true}, nil
	}
	markdown, _ := args["markdown"].(string)
	title, _ := args["title"].(string)
	if title == "" {
		title = hPath
		if strings.Contains(title, "/") {
			parts := strings.Split(strings.TrimRight(title, "/"), "/")
			title = parts[len(parts)-1]
		}
	}

	parentPath := "/"
	parentDir := parentDir(hPath)
	if parentDir != "/" {
		bt := treenode.GetBlockTreeRootByHPath(notebook, parentDir)
		if bt == nil {
			return CallToolResult{Content: []ContentItem{{Type: "text", Text: "parent path not found: " + parentDir}}, IsError: true}, nil
		}
		parentPath = strings.TrimSuffix(bt.Path, ".sy")
	}

	id := ast.NewNodeID()
	docPath := strings.TrimRight(parentPath, "/") + "/" + id + ".sy"
	tree, err := model.CreateDocByMd(notebook, docPath, title, markdown, nil, nil)
	if err != nil {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: fmt.Sprintf("create doc failed: %s", err)}}, IsError: true}, nil
	}

	message := fmt.Sprintf("document created: %s (hPath: %s)", tree.Root.ID, hPath)
	return structuredTextResult(message, map[string]any{
		"action":   "create",
		"id":       tree.Root.ID,
		"title":    title,
		"hPath":    hPath,
		"notebook": notebook,
		"path":     tree.Path,
	}), nil
}

func documentUpdate(args map[string]any) (CallToolResult, error) {
	id, _ := args["id"].(string)
	if id == "" {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "id is required"}}, IsError: true}, nil
	}
	markdown, _ := args["markdown"].(string)
	if markdown == "" {
		markdown = "\n"
	}
	if _, err := updateBlockContent(id, markdown, "markdown"); err != nil {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "update document failed: " + err.Error()}}, IsError: true}, nil
	}
	return structuredTextResult("document updated: "+id, map[string]any{
		"action": "update",
		"id":     id,
	}), nil
}

func parentDir(p string) string {
	i := strings.LastIndex(p, "/")
	if i <= 0 {
		return "/"
	}
	return p[:i]
}

func documentList(args map[string]any) (CallToolResult, error) {
	notebook, _ := args["notebook"].(string)
	if notebook == "" {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "notebook is required"}}, IsError: true}, nil
	}
	hPath, _ := args["path"].(string)
	if hPath == "" {
		hPath = "/"
	}

	fsPath := hPath
	if hPath != "/" {
		bt := treenode.GetBlockTreeRootByHPath(notebook, hPath)
		if bt == nil {
			return CallToolResult{Content: []ContentItem{{Type: "text", Text: "target path not found: " + hPath}}, IsError: true}, nil
		}
		fsPath = bt.Path
	}

	files, _, err := model.ListDocTree(notebook, fsPath, 0, false, false, 128)
	if err != nil {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: fmt.Sprintf("list docs failed: %s", err)}}, IsError: true}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Documents in %s (hPath: %s):\n\n", notebook, hPath))
	for _, f := range files {
		sb.WriteString(fmt.Sprintf("- %s (id: %s, hPath: %s)\n", f.Name, f.ID, strings.TrimRight(hPath, "/")+"/"+f.Name))
	}
	items := make([]map[string]any, 0, len(files))
	for _, f := range files {
		items = append(items, map[string]any{
			"id":    f.ID,
			"name":  f.Name,
			"hPath": strings.TrimRight(hPath, "/") + "/" + f.Name,
		})
	}
	return structuredTextResult(sb.String(), map[string]any{
		"action":   "list",
		"notebook": notebook,
		"hPath":    hPath,
		"count":    len(items),
		"items":    items,
	}), nil
}

func documentDelete(args map[string]any) (CallToolResult, error) {
	id, _ := args["id"].(string)
	if id == "" {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "id is required"}}, IsError: true}, nil
	}

	tree, err := model.LoadTreeByBlockID(id)
	if err != nil {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: fmt.Sprintf("load doc failed: %s", err)}}, IsError: true}, nil
	}

	if err = model.RemoveDoc(tree.Box, tree.Path); err != nil {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: fmt.Sprintf("delete doc failed: %s", err)}}, IsError: true}, nil
	}
	return structuredTextResult("document deleted: "+id, map[string]any{
		"action": "delete",
		"id":     id,
	}), nil
}

func documentRename(args map[string]any) (CallToolResult, error) {
	id, _ := args["id"].(string)
	title, _ := args["title"].(string)
	if id == "" || title == "" {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "id and title are required"}}, IsError: true}, nil
	}

	tree, err := model.LoadTreeByBlockID(id)
	if err != nil {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: fmt.Sprintf("load doc failed: %s", err)}}, IsError: true}, nil
	}

	if err := model.RenameDoc(tree.Box, tree.Path, title); err != nil {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: fmt.Sprintf("rename doc failed: %s", err)}}, IsError: true}, nil
	}

	return structuredTextResult(fmt.Sprintf("document renamed: %s -> %s", id, title), map[string]any{
		"action": "rename",
		"id":     id,
		"title":  title,
	}), nil
}

func documentMove(args map[string]any) (CallToolResult, error) {
	id, _ := args["id"].(string)
	notebook, _ := args["notebook"].(string)
	hPath, _ := args["path"].(string)
	if id == "" || notebook == "" || hPath == "" {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "id, notebook and path are required"}}, IsError: true}, nil
	}

	tree, err := model.LoadTreeByBlockID(id)
	if err != nil {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: fmt.Sprintf("load doc failed: %s", err)}}, IsError: true}, nil
	}

	fsPath := hPath
	if hPath != "/" {
		bt := treenode.GetBlockTreeRootByHPath(notebook, hPath)
		if bt == nil {
			return CallToolResult{Content: []ContentItem{{Type: "text", Text: "target path not found: " + hPath}}, IsError: true}, nil
		}
		fsPath = bt.Path
	}

	if err := model.MoveDocs([]string{tree.Path}, notebook, fsPath, nil); err != nil {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: fmt.Sprintf("move doc failed: %s", err)}}, IsError: true}, nil
	}

	return structuredTextResult(fmt.Sprintf("document moved: %s -> %s (hPath: %s)", id, notebook, hPath), map[string]any{
		"action":   "move",
		"id":       id,
		"notebook": notebook,
		"hPath":    hPath,
	}), nil
}

func documentDuplicate(args map[string]any) (CallToolResult, error) {
	id, _ := args["id"].(string)
	if id == "" {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "id is required"}}, IsError: true}, nil
	}

	tree, err := model.LoadTreeByBlockID(id)
	if err != nil {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: fmt.Sprintf("load doc failed: %s", err)}}, IsError: true}, nil
	}

	originalID := id
	model.DuplicateDoc(tree)
	util.PushReloadFiletree()
	return structuredTextResult("document duplicated: "+originalID+" -> "+tree.Root.ID, map[string]any{
		"action":     "duplicate",
		"id":         tree.Root.ID,
		"originalID": originalID,
		"hPath":      tree.HPath,
		"path":       tree.Path,
		"notebook":   tree.Box,
	}), nil
}

func documentSearchDocs(args map[string]any) (CallToolResult, error) {
	keyword, _ := args["keyword"].(string)
	if keyword == "" {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "keyword is required"}}, IsError: true}, nil
	}

	docs := model.SearchDocs(keyword, false, nil)
	if len(docs) == 0 {
		return structuredTextResult("no documents found", map[string]any{
			"action":  "search_docs",
			"keyword": keyword,
			"count":   0,
			"items":   []map[string]any{},
		}), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Documents matching '%s' (%d):\n\n", keyword, len(docs)))
	for _, d := range docs {
		item := documentSearchDocItem(d)
		sb.WriteString(fmt.Sprintf("- %s (id: %s, hPath: %s)\n", item["name"], item["id"], item["hPath"]))
	}
	items := make([]map[string]any, 0, len(docs))
	for _, d := range docs {
		items = append(items, documentSearchDocItem(d))
	}
	return structuredTextResult(sb.String(), map[string]any{
		"action":  "search_docs",
		"keyword": keyword,
		"count":   len(items),
		"items":   items,
	}), nil
}

func documentSearchDocItem(d map[string]string) map[string]any {
	item := map[string]any{
		"id":       d["id"],
		"name":     d["name"],
		"hPath":    d["hPath"],
		"path":     d["path"],
		"notebook": d["box"],
		"boxIcon":  d["boxIcon"],
	}
	if item["id"] == "" && d["path"] != "" && d["path"] != "/" {
		item["id"] = util.GetTreeID(d["path"])
	}
	if item["name"] == "" {
		hPath := strings.TrimRight(d["hPath"], "/")
		if hPath != "" {
			parts := strings.Split(hPath, "/")
			item["name"] = parts[len(parts)-1]
		}
	}
	if item["id"] == "" {
		item["id"] = d["box"]
	}
	return item
}

func documentInfo(args map[string]any) (CallToolResult, error) {
	id, _ := args["id"].(string)
	if id == "" {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "id is required"}}, IsError: true}, nil
	}

	info, err := model.GetDocInfo(id)
	if err != nil {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: fmt.Sprintf("get doc info failed: %s", err)}}, IsError: true}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(
		"ID: %s\nRootID: %s\nName: %s\nRefCount: %d\nSubFileCount: %d\nIcon: %s",
		info.ID, info.RootID, info.Name, info.RefCount, info.SubFileCount, info.Icon,
	))
	if len(info.RefIDs) > 0 {
		sb.WriteString(fmt.Sprintf("\nRefIDs: %s", strings.Join(info.RefIDs, ", ")))
	}
	if len(info.AttrViews) > 0 {
		sb.WriteString("\nAttrViews:")
		for _, av := range info.AttrViews {
			sb.WriteString(fmt.Sprintf("\n  - %s: %s", av.ID, av.Name))
		}
	}
	if len(info.IAL) > 0 {
		sb.WriteString("\nIAL:")
		for k, v := range info.IAL {
			v = maybeRedactSensitiveText(args, v)
			if len(v) > 100 {
				v = v[:100] + "..."
			}
			sb.WriteString(fmt.Sprintf("\n  %s: %s", k, v))
		}
	}

	attrViews := make([]map[string]any, 0, len(info.AttrViews))
	for _, item := range info.AttrViews {
		attrViews = append(attrViews, map[string]any{"id": item.ID, "name": item.Name})
	}
	return structuredTextResult(sb.String(), map[string]any{
		"action":       "info",
		"id":           info.ID,
		"rootID":       info.RootID,
		"name":         info.Name,
		"refCount":     info.RefCount,
		"subFileCount": info.SubFileCount,
		"icon":         info.Icon,
		"refIDs":       info.RefIDs,
		"attrViews":    attrViews,
		"ial":          maybeRedactStringMap(args, info.IAL),
	}), nil
}
