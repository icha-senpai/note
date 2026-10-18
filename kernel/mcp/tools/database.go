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
	"encoding/json"
	"fmt"
	"strings"

	"github.com/icha-senpai/note/kernel/av"
	"github.com/icha-senpai/note/kernel/model"
	"github.com/icha-senpai/note/third_party/forks/lute/ast"
)

var DatabaseTool = &Tool{
	Name:        "database",
	Description: "Attribute view (database) operations. Actions: create(name, id?, blockID?), search(keyword), get(id), render(id, viewID?, query?, page=1, pageSize=50), keys(id), key_add(id, name, type, icon?, prev?), key_remove(id, keyID, removeRelationDest?), item_add(id, blockID?, content?, viewID?, groupID?, previousID?, detached?, ignoreDefaultFill?), item_remove(id, itemIDs comma-separated), item_update(id, keyID, itemID, value as JSON string), view_list(id), view_add(id, name?, layout?), view_remove(id, viewID), view_update(id, viewID, name?, desc?, icon?, pageSize?, hideName?), view_layout(id, blockID, layout) changes a bound view's layout, view_group(id, blockID, keyID?, order?, hideEmpty?) sets grouping/group-by on a bound view (empty keyID clears), unused(), clean(id?).",
	InputSchema: ToolSchema{
		Type: "object",
		Properties: map[string]Property{
			"action":             {Type: "string", Description: "Operation", Enum: []string{"create", "search", "get", "render", "keys", "key_add", "key_remove", "item_add", "item_remove", "item_update", "view_list", "view_add", "view_remove", "view_update", "view_layout", "view_group", "unused", "clean"}},
			"layout":             {Type: "string", Description: "View layout: table/kanban/gallery/calendar (for view_add, view_layout)", Enum: []string{"table", "kanban", "gallery", "calendar"}},
			"viewID":             {Type: "string", Description: "View ID (for render, item_add, view_remove, view_update)"},
			"desc":               {Type: "string", Description: "View description (for view_update)"},
			"order":              {Type: "string", Description: "Group order for view_group: asc/desc (default asc)", Enum: []string{"asc", "desc"}},
			"hideEmpty":          {Type: "boolean", Description: "Hide empty groups (for view_group) or hide the database name (for view_update)"},
			"hideName":           {Type: "boolean", Description: "Hide the attribute view name on the view (for view_update)"},
			"keyword":            {Type: "string", Description: "Search keyword (for search)"},
			"id":                 {Type: "string", Description: "Attribute view ID (optional for create, required for get, render, keys, key_add, key_remove, item_add, item_remove, item_update, view_list, view_add, view_remove, view_update, view_layout, view_group, clean)"},
			"query":              {Type: "string", Description: "Filter query (for render)"},
			"page":               {Type: "number", Description: "Page number (default 1)"},
			"pageSize":           {Type: "number", Description: "Results per page (default 50 for render; also sets page size for view_update)"},
			"name":               {Type: "string", Description: "Database name (for create) or key name (for key_add)"},
			"type":               {Type: "string", Description: "Key type: block/text/number/date/select/mSelect/url/email/phone/mAsset/template/created/updated/checkbox/relation/rollup/lineNumber (for key_add)"},
			"icon":               {Type: "string", Description: "Key icon (for key_add, optional)"},
			"prev":               {Type: "string", Description: "Previous key ID for ordering (for key_add, optional)"},
			"keyID":              {Type: "string", Description: "Key ID (for key_remove, item_update)"},
			"removeRelationDest": {Type: "boolean", Description: "Also remove related data in linked databases (for key_remove, optional)"},
			"blockID":            {Type: "string", Description: "Block ID — for create = block to bind; item_add = row block to bind; view_layout/view_group = bound block owning the target view"},
			"content":            {Type: "string", Description: "Block column text content (for item_add, optional)"},
			"groupID":            {Type: "string", Description: "Group ID for positioning (for item_add, optional)"},
			"previousID":         {Type: "string", Description: "Previous item ID for positioning (for item_add, optional)"},
			"detached":           {Type: "boolean", Description: "Create detached row (for item_add, optional)"},
			"ignoreDefaultFill":  {Type: "boolean", Description: "Skip filling default values (for item_add, optional)"},
			"itemID":             {Type: "string", Description: "Item ID (for item_update)"},
			"itemIDs":            {Type: "string", Description: "Comma-separated item IDs (for item_remove)"},
			"value":              {Type: "string", Description: "JSON value for the cell (for item_update)"},
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
		effectMap(ToolEffects{LocalRead: true}, "search", "get", "render", "keys", "view_list", "unused"),
		effectMap(ToolEffects{LocalWrite: true}, "create", "key_add", "key_remove", "item_add", "item_remove", "item_update", "view_add", "view_remove", "view_update", "view_layout", "view_group", "clean"),
	),
	Handler: databaseHandler,
}

func init() {
	register(DatabaseTool)
}

func databaseHandler(args map[string]any) (CallToolResult, error) {
	action, _ := args["action"].(string)
	switch action {
	case "create":
		return databaseCreate(args)
	case "search":
		return databaseSearch(args)
	case "get":
		return databaseGet(args)
	case "render":
		return databaseRender(args)
	case "keys":
		return databaseKeys(args)
	case "key_add":
		return databaseKeyAdd(args)
	case "key_remove":
		return databaseKeyRemove(args)
	case "item_add":
		return databaseItemAdd(args)
	case "item_remove":
		return databaseItemRemove(args)
	case "item_update":
		return databaseItemUpdate(args)
	case "view_list":
		return databaseViewList(args)
	case "view_add":
		return databaseViewAdd(args)
	case "view_remove":
		return databaseViewRemove(args)
	case "view_update":
		return databaseViewUpdate(args)
	case "view_layout":
		return databaseViewLayout(args)
	case "view_group":
		return databaseViewGroup(args)
	case "unused":
		return databaseUnused(args)
	case "clean":
		return databaseClean(args)
	}
	return CallToolResult{
		Content: []ContentItem{{Type: "text", Text: "unknown action '" + action + "', expected one of: [create, search, get, render, keys, key_add, key_remove, item_add, item_remove, item_update, view_list, view_add, view_remove, view_update, view_layout, view_group, unused, clean]"}},
		IsError: true,
	}, nil
}

func databaseViewList(args map[string]any) (CallToolResult, error) {
	attrView, result, done := parseAttrViewArg(args)
	if done {
		return result, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Views on '%s' (av: %s, current: %s):\n\n", attrView.Name, attrView.ID, attrView.ViewID))
	items := make([]map[string]any, 0, len(attrView.Views))
	for _, v := range attrView.Views {
		item := map[string]any{
			"id":        v.ID,
			"name":      v.Name,
			"layout":    string(v.LayoutType),
			"pageSize":  v.PageSize,
			"itemCount": len(v.ItemIDs),
		}
		if v.HideAttrViewName {
			item["hideAttrViewName"] = true
		}
		if v.Icon != "" {
			item["icon"] = v.Icon
		}
		if v.Desc != "" {
			item["desc"] = v.Desc
		}
		if v.Group != nil {
			item["group"] = map[string]any{
				"field":     v.Group.Field,
				"method":    int(v.Group.Method),
				"order":     int(v.Group.Order),
				"hideEmpty": v.Group.HideEmpty,
			}
		}
		sb.WriteString(fmt.Sprintf("- %s [%s] (viewID: %s, items: %d)\n", v.Name, v.LayoutType, v.ID, len(v.ItemIDs)))
		items = append(items, item)
	}
	return structuredTextResult(sb.String(), map[string]any{
		"action":        "view_list",
		"id":            attrView.ID,
		"name":          attrView.Name,
		"currentViewID": attrView.ViewID,
		"count":         len(items),
		"items":         items,
	}), nil
}

func databaseViewAdd(args map[string]any) (CallToolResult, error) {
	attrView, result, done := parseAttrViewArg(args)
	if done {
		return result, nil
	}

	layout, _ := args["layout"].(string)
	var view *av.View
	switch av.LayoutType(layout) {
	case av.LayoutTypeKanban:
		view = av.NewKanbanView()
	case av.LayoutTypeGallery:
		view = av.NewGalleryView()
	case av.LayoutTypeCalendar:
		view = av.NewTableView()
		view.LayoutType = av.LayoutTypeCalendar
		view.Table = nil
		view.Calendar = &av.CalendarConfig{}
		view.Name = av.GetAttributeViewI18n("calendar")
	default:
		view = av.NewTableView()
	}
	if name, _ := args["name"].(string); strings.TrimSpace(name) != "" {
		view.Name = strings.TrimSpace(name)
	}

	attrView.Views = append(attrView.Views, view)
	if attrView.ViewID == "" {
		attrView.ViewID = view.ID
	}
	if err := av.SaveAttributeView(attrView); err != nil {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "save database failed: " + err.Error()}}, IsError: true}, nil
	}
	model.ReloadAttrView(attrView.ID)

	return structuredTextResult(fmt.Sprintf("view added: %s [%s] (viewID: %s)", view.Name, view.LayoutType, view.ID), map[string]any{
		"action": "view_add",
		"id":     attrView.ID,
		"viewID": view.ID,
		"name":   view.Name,
		"layout": string(view.LayoutType),
	}), nil
}

func databaseViewRemove(args map[string]any) (CallToolResult, error) {
	attrView, result, done := parseAttrViewArg(args)
	if done {
		return result, nil
	}
	viewID, _ := args["viewID"].(string)
	if viewID == "" {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "viewID is required"}}, IsError: true}, nil
	}
	if len(attrView.Views) <= 1 {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "cannot remove the only view"}}, IsError: true}, nil
	}

	index := -1
	for i, v := range attrView.Views {
		if v.ID == viewID {
			index = i
			break
		}
	}
	if index < 0 {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "view not found: " + viewID}}, IsError: true}, nil
	}
	attrView.Views = append(attrView.Views[:index], attrView.Views[index+1:]...)
	if attrView.ViewID == viewID {
		attrView.ViewID = attrView.Views[0].ID
	}
	if err := av.SaveAttributeView(attrView); err != nil {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "save database failed: " + err.Error()}}, IsError: true}, nil
	}
	model.ReloadAttrView(attrView.ID)

	return structuredTextResult("view removed: "+viewID, map[string]any{
		"action":        "view_remove",
		"id":            attrView.ID,
		"viewID":        viewID,
		"currentViewID": attrView.ViewID,
	}), nil
}

func databaseViewUpdate(args map[string]any) (CallToolResult, error) {
	attrView, result, done := parseAttrViewArg(args)
	if done {
		return result, nil
	}
	viewID, _ := args["viewID"].(string)
	if viewID == "" {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "viewID is required"}}, IsError: true}, nil
	}
	view := attrView.GetView(viewID)
	if view == nil {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "view not found: " + viewID}}, IsError: true}, nil
	}

	changed := map[string]any{}
	if name, exists := args["name"]; exists {
		if s, ok := name.(string); ok && strings.TrimSpace(s) != "" {
			view.Name = strings.TrimSpace(s)
			changed["name"] = view.Name
		}
	}
	if desc, exists := args["desc"]; exists {
		if s, ok := desc.(string); ok {
			view.Desc = s
			changed["desc"] = view.Desc
		}
	}
	if icon, exists := args["icon"]; exists {
		if s, ok := icon.(string); ok {
			view.Icon = s
			changed["icon"] = view.Icon
		}
	}
	if hideName, exists := args["hideName"]; exists {
		if b, ok := hideName.(bool); ok {
			view.HideAttrViewName = b
			changed["hideAttrViewName"] = b
		}
	}
	if pageSize := intArg(args, "pageSize", 0); pageSize > 0 {
		view.PageSize = pageSize
		changed["pageSize"] = pageSize
	}
	if len(changed) == 0 {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "nothing to update; pass name, desc, icon, hideName, or pageSize"}}, IsError: true}, nil
	}

	if err := av.SaveAttributeView(attrView); err != nil {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "save database failed: " + err.Error()}}, IsError: true}, nil
	}
	model.ReloadAttrView(attrView.ID)

	changed["action"] = "view_update"
	changed["id"] = attrView.ID
	changed["viewID"] = viewID
	return structuredTextResult("view updated: "+viewID, changed), nil
}

func databaseViewLayout(args map[string]any) (CallToolResult, error) {
	id, _ := args["id"].(string)
	if id == "" {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "id is required"}}, IsError: true}, nil
	}
	layout, _ := args["layout"].(string)
	if layout == "" {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "layout is required"}}, IsError: true}, nil
	}
	blockID, _ := args["blockID"].(string)

	if err := model.ChangeAttrViewLayout(blockID, id, av.LayoutType(layout)); err != nil {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "change layout failed: " + err.Error()}}, IsError: true}, nil
	}
	return structuredTextResult(fmt.Sprintf("layout changed to %s", layout), map[string]any{
		"action":  "view_layout",
		"id":      id,
		"blockID": blockID,
		"layout":  layout,
	}), nil
}

func databaseViewGroup(args map[string]any) (CallToolResult, error) {
	id, _ := args["id"].(string)
	if id == "" {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "id is required"}}, IsError: true}, nil
	}
	blockID, _ := args["blockID"].(string)
	keyID, _ := args["keyID"].(string)
	hideEmpty, _ := args["hideEmpty"].(bool)
	order, _ := args["order"].(string)

	group := &av.ViewGroup{
		Field:     keyID,
		Method:    av.GroupMethodValue,
		Order:     av.GroupOrderAsc,
		HideEmpty: hideEmpty,
	}
	if order == "desc" {
		group.Order = av.GroupOrderDesc
	}
	if keyID == "" {
		group = &av.ViewGroup{}
	}

	if err := model.SetAttributeViewGroup(id, blockID, group); err != nil {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "set group failed: " + err.Error()}}, IsError: true}, nil
	}
	message := fmt.Sprintf("group set on key %s", keyID)
	if keyID == "" {
		message = "grouping cleared"
	}
	return structuredTextResult(message, map[string]any{
		"action":    "view_group",
		"id":        id,
		"blockID":   blockID,
		"keyID":     keyID,
		"order":     group.Order,
		"hideEmpty": group.HideEmpty,
	}), nil
}

func parseAttrViewArg(args map[string]any) (attrView *av.AttributeView, result CallToolResult, done bool) {
	id, _ := args["id"].(string)
	if id == "" {
		return nil, CallToolResult{Content: []ContentItem{{Type: "text", Text: "id is required"}}, IsError: true}, true
	}
	attrView, err := av.ParseAttributeView(id)
	if err != nil {
		return nil, CallToolResult{Content: []ContentItem{{Type: "text", Text: "parse database failed: " + err.Error()}}, IsError: true}, true
	}
	return attrView, CallToolResult{}, false
}

func databaseCreate(args map[string]any) (CallToolResult, error) {
	name, _ := args["name"].(string)
	name = strings.TrimSpace(name)
	if name == "" {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "name is required"}}, IsError: true}, nil
	}

	id, _ := args["id"].(string)
	id = strings.TrimSpace(id)
	if id == "" {
		id = ast.NewNodeID()
	} else if !ast.IsNodeIDPattern(id) {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "id must be a valid Scribli block/database ID"}}, IsError: true}, nil
	}
	if av.IsAttributeViewExist(id) {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "database already exists: " + id}}, IsError: true}, nil
	}

	attrView := av.NewAttributeView(id)
	attrView.Name = name
	if err := av.SaveAttributeView(attrView); err != nil {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "create database failed: " + err.Error()}}, IsError: true}, nil
	}

	blockID, _ := args["blockID"].(string)
	blockID = strings.TrimSpace(blockID)
	if blockID != "" {
		av.UpsertBlockRel(id, blockID)
	}
	model.ReloadAttrView(id)

	message := fmt.Sprintf("database created: %s (%s)", id, name)
	return structuredTextResult(message, map[string]any{
		"action":  "create",
		"id":      id,
		"name":    name,
		"viewID":  attrView.ViewID,
		"blockID": blockID,
	}), nil
}

func databaseSearch(args map[string]any) (CallToolResult, error) {
	keyword, _ := args["keyword"].(string)
	if keyword == "" {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "keyword is required"}}, IsError: true}, nil
	}

	results := model.SearchAttributeView(keyword, nil, "", "")
	items := make([]map[string]any, 0, len(results))
	seen := map[string]bool{}
	for _, r := range results {
		items = append(items, map[string]any{"id": r.AvID, "name": r.AvName, "hPath": r.HPath, "unused": false})
		seen[r.AvID] = true
	}
	for _, item := range databaseUnusedSearchMatches(keyword, seen) {
		items = append(items, item)
	}
	if len(items) == 0 {
		return structuredTextResult("no attribute views found", map[string]any{
			"action":  "search",
			"keyword": keyword,
			"count":   0,
			"items":   []map[string]any{},
		}), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Attribute views matching '%s' (%d):\n\n", keyword, len(items)))
	for _, item := range items {
		extra := ""
		if unused, _ := item["unused"].(bool); unused {
			extra = " [unused]"
		}
		sb.WriteString(fmt.Sprintf("- %s (id: %s, hPath: %s)%s\n", item["name"], item["id"], item["hPath"], extra))
	}
	return structuredTextResult(sb.String(), map[string]any{
		"action":  "search",
		"keyword": keyword,
		"count":   len(items),
		"items":   items,
	}), nil
}

func databaseGet(args map[string]any) (CallToolResult, error) {
	id, _ := args["id"].(string)
	if id == "" {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "id is required"}}, IsError: true}, nil
	}

	attrView := model.GetAttributeView(id)
	if attrView == nil {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "attribute view not found: " + id}}, IsError: true}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Attribute View: %s\n\n", id))
	sb.WriteString(fmt.Sprintf("Name: %s\n", attrView.Name))
	sb.WriteString(fmt.Sprintf("Keys (%d):\n", len(attrView.KeyValues)))
	for _, kv := range attrView.KeyValues {
		sb.WriteString(fmt.Sprintf("- %s (%s): %s\n", kv.Key.Name, kv.Key.Type, kv.Key.Icon))
	}
	sb.WriteString(fmt.Sprintf("\nViews (%d):\n", len(attrView.Views)))
	for _, v := range attrView.Views {
		sb.WriteString(fmt.Sprintf("- %s (%s, pageSize: %d)\n", v.Name, v.LayoutType, v.PageSize))
	}
	return structuredTextResult(sb.String(), map[string]any{
		"action": "get",
		"id":     id,
		"name":   attrView.Name,
		"keys":   databaseKeySummaries(attrView),
		"views":  databaseViewSummaries(attrView),
	}), nil
}

func databaseRender(args map[string]any) (CallToolResult, error) {
	id, _ := args["id"].(string)
	if id == "" {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "id is required"}}, IsError: true}, nil
	}
	viewID, _ := args["viewID"].(string)
	query, _ := args["query"].(string)
	page := 1
	if v, ok := args["page"].(float64); ok {
		page = int(v)
	}
	pageSize := 50
	if v, ok := args["pageSize"].(float64); ok {
		pageSize = int(v)
	}

	viewable, _, err := model.RenderAttributeView("", id, viewID, query, page, pageSize, nil, false, false)
	if err != nil {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "render failed: " + err.Error()}}, IsError: true}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Attribute View Render (page %d):\n\n", page))
	if viewable != nil {
		if table, ok := viewable.(*av.Table); ok {
			for _, row := range table.Rows {
				vals := make([]string, 0, len(row.Cells))
				for _, cell := range row.Cells {
					vals = append(vals, maybeRedactSensitiveText(args, databaseCellDisplay(cell.Value)))
				}
				sb.WriteString(strings.Join(vals, " | ") + "\n")
			}
		}
	}
	return structuredTextResult(sb.String(), map[string]any{
		"action":   "render",
		"id":       id,
		"viewID":   viewID,
		"query":    query,
		"page":     page,
		"pageSize": pageSize,
		"rows":     databaseRenderedRows(args, viewable),
	}), nil
}

func databaseKeys(args map[string]any) (CallToolResult, error) {
	id, _ := args["id"].(string)
	if id == "" {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "id is required"}}, IsError: true}, nil
	}
	attrView := model.GetAttributeView(id)
	if attrView == nil {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "attribute view not found: " + id}}, IsError: true}, nil
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Keys for %s (%s):\n\n", id, attrView.Name))
	for _, kv := range attrView.KeyValues {
		sb.WriteString(fmt.Sprintf("- %s (%s) [%s]\n", kv.Key.ID, kv.Key.Name, kv.Key.Type))
	}
	return structuredTextResult(sb.String(), map[string]any{
		"action": "keys",
		"id":     id,
		"name":   attrView.Name,
		"keys":   databaseKeySummaries(attrView),
	}), nil
}

func databaseKeyAdd(args map[string]any) (CallToolResult, error) {
	id, _ := args["id"].(string)
	name, _ := args["name"].(string)
	keyType, _ := args["type"].(string)
	if id == "" || name == "" || keyType == "" {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "id, name and type are required"}}, IsError: true}, nil
	}
	icon, _ := args["icon"].(string)
	prev, _ := args["prev"].(string)
	keyID := ast.NewNodeID()
	if err := model.AddAttributeViewKey(id, keyID, name, keyType, icon, prev); err != nil {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "add key failed: " + err.Error()}}, IsError: true}, nil
	}
	model.ReloadAttrView(id)
	return structuredTextResult(fmt.Sprintf("key added: %s (%s)", keyID, name), map[string]any{
		"action": "key_add",
		"id":     id,
		"keyID":  keyID,
		"name":   name,
		"type":   keyType,
	}), nil
}

func databaseKeyRemove(args map[string]any) (CallToolResult, error) {
	id, _ := args["id"].(string)
	keyID, _ := args["keyID"].(string)
	if id == "" || keyID == "" {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "id and keyID are required"}}, IsError: true}, nil
	}
	removeRelation := false
	if v, ok := args["removeRelationDest"].(bool); ok {
		removeRelation = v
	}
	if err := model.RemoveAttributeViewKey(id, keyID, removeRelation); err != nil {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "remove key failed: " + err.Error()}}, IsError: true}, nil
	}
	model.ReloadAttrView(id)
	return structuredTextResult("key removed: "+keyID, map[string]any{
		"action": "key_remove",
		"id":     id,
		"keyID":  keyID,
	}), nil
}

func databaseItemAdd(args map[string]any) (CallToolResult, error) {
	id, _ := args["id"].(string)
	if id == "" {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "id is required"}}, IsError: true}, nil
	}
	isDetached := false
	if v, ok := args["detached"].(bool); ok {
		isDetached = v
	}
	blockID, _ := args["blockID"].(string)
	if !isDetached && blockID == "" {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "blockID is required for non-detached rows"}}, IsError: true}, nil
	}
	content, _ := args["content"].(string)
	viewID, _ := args["viewID"].(string)
	groupID, _ := args["groupID"].(string)
	previousID, _ := args["previousID"].(string)
	ignoreFill := false
	if v, ok := args["ignoreDefaultFill"].(bool); ok {
		ignoreFill = v
	}
	src := map[string]any{"isDetached": isDetached}
	if blockID != "" {
		src["id"] = blockID
	}
	if content != "" {
		src["content"] = content
	}
	srcs := []map[string]any{src}
	if err := model.AddAttributeViewBlock(nil, srcs, id, blockID, viewID, groupID, previousID, ignoreFill); err != nil {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "add item failed: " + err.Error()}}, IsError: true}, nil
	}
	model.ReloadAttrView(id)
	return structuredTextResult("item added", map[string]any{
		"action":   "item_add",
		"id":       id,
		"blockID":  blockID,
		"detached": isDetached,
		"viewID":   viewID,
		"groupID":  groupID,
	}), nil
}

func databaseItemRemove(args map[string]any) (CallToolResult, error) {
	id, _ := args["id"].(string)
	itemIDsStr, _ := args["itemIDs"].(string)
	if id == "" || itemIDsStr == "" {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "id and itemIDs are required"}}, IsError: true}, nil
	}
	itemIDs := strings.Split(itemIDsStr, ",")
	for i := range itemIDs {
		itemIDs[i] = strings.TrimSpace(itemIDs[i])
	}
	if err := model.RemoveAttributeViewBlock(itemIDs, id); err != nil {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "remove items failed: " + err.Error()}}, IsError: true}, nil
	}
	model.ReloadAttrView(id)
	return structuredTextResult(fmt.Sprintf("%d item(s) removed", len(itemIDs)), map[string]any{
		"action":  "item_remove",
		"id":      id,
		"itemIDs": itemIDs,
		"count":   len(itemIDs),
	}), nil
}

func databaseItemUpdate(args map[string]any) (CallToolResult, error) {
	id, _ := args["id"].(string)
	keyID, _ := args["keyID"].(string)
	itemID, _ := args["itemID"].(string)
	valueStr, _ := args["value"].(string)
	if id == "" || keyID == "" || itemID == "" || valueStr == "" {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "id, keyID, itemID and value are required"}}, IsError: true}, nil
	}
	var valueData map[string]any
	if err := json.Unmarshal([]byte(valueStr), &valueData); err != nil {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "invalid JSON value: " + err.Error()}}, IsError: true}, nil
	}
	if _, err := model.UpdateAttributeViewCell(nil, id, keyID, itemID, valueData); err != nil {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "update cell failed: " + err.Error()}}, IsError: true}, nil
	}
	model.ReloadAttrView(id)
	return structuredTextResult("cell updated", map[string]any{
		"action": "item_update",
		"id":     id,
		"keyID":  keyID,
		"itemID": itemID,
	}), nil
}

func databaseUnused(args map[string]any) (CallToolResult, error) {
	items, err := model.UnusedAttributeViews(true)
	if err != nil {
		return CallToolResult{}, err
	}
	if len(items) == 0 {
		return structuredTextResult("no unused databases found", map[string]any{
			"action": "unused",
			"count":  0,
			"items":  []map[string]any{},
		}), nil
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Unused databases (%d):\n\n", len(items)))
	for _, item := range items {
		sb.WriteString(fmt.Sprintf("- %s (%s)\n", item.Item, item.Name))
	}
	structuredItems := make([]map[string]any, 0, len(items))
	for _, item := range items {
		structuredItems = append(structuredItems, map[string]any{"id": item.Item, "name": item.Name})
	}
	return structuredTextResult(sb.String(), map[string]any{
		"action": "unused",
		"count":  len(structuredItems),
		"items":  structuredItems,
	}), nil
}

func databaseClean(args map[string]any) (CallToolResult, error) {
	id, _ := args["id"].(string)
	if id != "" {
		if err := model.RemoveUnusedAttributeView(id); err != nil {
			return CallToolResult{}, err
		}
		return structuredTextResult("unused database cleaned: "+id, map[string]any{
			"action": "clean",
			"id":     id,
			"count":  1,
		}), nil
	}
	removed, err := model.RemoveUnusedAttributeViews()
	if err != nil {
		return CallToolResult{}, err
	}
	return structuredTextResult(fmt.Sprintf("%d unused database(s) cleaned", len(removed)), map[string]any{
		"action": "clean",
		"ids":    removed,
		"count":  len(removed),
	}), nil
}

func databaseKeySummaries(attrView *av.AttributeView) []map[string]any {
	keys := make([]map[string]any, 0, len(attrView.KeyValues))
	for _, kv := range attrView.KeyValues {
		if kv == nil || kv.Key == nil {
			continue
		}
		keys = append(keys, map[string]any{
			"id":   kv.Key.ID,
			"name": kv.Key.Name,
			"type": string(kv.Key.Type),
			"icon": kv.Key.Icon,
		})
	}
	return keys
}

func databaseViewSummaries(attrView *av.AttributeView) []map[string]any {
	views := make([]map[string]any, 0, len(attrView.Views))
	for _, view := range attrView.Views {
		if view == nil {
			continue
		}
		views = append(views, map[string]any{
			"id":         view.ID,
			"name":       view.Name,
			"layoutType": string(view.LayoutType),
			"pageSize":   view.PageSize,
		})
	}
	return views
}

func databaseRenderedRows(args map[string]any, viewable av.Viewable) []map[string]any {
	table, ok := viewable.(*av.Table)
	if !ok || table == nil {
		return []map[string]any{}
	}
	rows := make([]map[string]any, 0, len(table.Rows))
	for _, row := range table.Rows {
		if row == nil {
			continue
		}
		cells := make([]map[string]any, 0, len(row.Cells))
		for _, cell := range row.Cells {
			if cell == nil {
				continue
			}
			cells = append(cells, map[string]any{
				"keyID":     cell.ID,
				"valueType": string(cell.ValueType),
				"value":     maybeRedactSensitiveText(args, databaseCellDisplay(cell.Value)),
			})
		}
		rows = append(rows, map[string]any{"id": row.ID, "cells": cells})
	}
	return rows
}

func databaseUnusedSearchMatches(keyword string, seen map[string]bool) []map[string]any {
	keyword = strings.ToLower(strings.TrimSpace(keyword))
	if keyword == "" {
		return nil
	}
	matches := []map[string]any{}
	items, err := model.UnusedAttributeViews(true)
	if err != nil {
		return nil
	}
	for _, item := range items {
		if item == nil || seen[item.Item] {
			continue
		}
		haystack := strings.ToLower(item.Item + " " + item.Name)
		if !strings.Contains(haystack, keyword) {
			continue
		}
		seen[item.Item] = true
		matches = append(matches, map[string]any{
			"id":     item.Item,
			"name":   item.Name,
			"hPath":  "",
			"unused": true,
		})
	}
	return matches
}

func databaseCellDisplay(value *av.Value) string {
	if value == nil {
		return ""
	}
	return value.String(true)
}
