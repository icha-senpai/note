// Scribli - Refactor your thinking
// Copyright (c) 2020-present, Scribli contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package model

import (
	"reflect"
	"testing"

	"github.com/icha-senpai/note/kernel/av"
)

func TestRestoreAttributeViewPreservesConfigurationAndOrder(t *testing.T) {
	first, last, saved := &av.View{LayoutType: av.LayoutTypeTable}, &av.View{LayoutType: av.LayoutTypeTable}, &av.View{LayoutType: av.LayoutTypeKanban}
	first.ID, last.ID, saved.ID = "first", "last", "saved"
	saved.Name, saved.Desc, saved.PageSize = "My board", "Description", 200
	saved.Group = &av.ViewGroup{Field: "status"}
	saved.Filters = []*av.ViewFilter{{Column: "status", Operator: av.FilterOperatorIsNotEmpty}}
	saved.ItemIDs = []string{"row-1", "row-2"}
	database := &av.AttributeView{Views: []*av.View{first, last}}
	if err := restoreAttributeViewView(database, saved, first.ID); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(database.Views, []*av.View{first, saved, last}) || database.ViewID != saved.ID {
		t.Fatal("restored view should preserve its complete settings and original position")
	}
	if err := restoreAttributeViewView(database, saved, first.ID); err == nil {
		t.Fatal("duplicate restoration must not create a duplicate view")
	}
}

func TestRestoreAttributeViewColumnRestoresCellsFiltersAndDefaults(t *testing.T) {
	key := av.NewKey("status", "Status", "", av.KeyTypeSelect)
	key.Options = []*av.SelectOption{{Name: "Done", Color: "2"}}
	current := &av.Value{ID: "cell", BlockID: "row", Type: av.KeyTypeSelect, MSelect: []*av.ValueSelect{}}
	added := &av.Value{ID: "new-cell", BlockID: "new-row", Type: av.KeyTypeSelect, MSelect: []*av.ValueSelect{{Content: "Done", Color: "2"}}}
	original := &av.Value{ID: "cell", BlockID: "row", Type: av.KeyTypeSelect, MSelect: []*av.ValueSelect{{Content: "To do", Color: "1"}}}
	filter := &av.ViewFilter{Column: key.ID, Value: original}
	originalDefault := &av.NewItemFieldValue{Mode: av.NewItemFieldValueStatic, Value: original}
	unrelatedDefault := &av.NewItemFieldValue{Mode: av.NewItemFieldValueStatic, Value: &av.Value{Type: av.KeyTypeText, Text: &av.ValueText{Content: "keep"}}}
	view := &av.View{LayoutType: av.LayoutTypeTable}
	view.ID = "view"
	template := &av.NewItemTemplate{ID: "template", FieldValues: map[string]*av.NewItemFieldValue{"other": unrelatedDefault}}
	database := &av.AttributeView{KeyValues: []*av.KeyValues{{Key: key, Values: []*av.Value{current, added}}}, Views: []*av.View{view}, NewItemTemplates: []*av.NewItemTemplate{template}}
	snapshot := &attributeViewColumnSnapshot{KeyID: key.ID,
		Options: []*av.SelectOption{{Name: "To do", Color: "1"}, {Name: "Done", Color: "2"}},
		Values:  []*av.Value{original}, Filters: map[string][]*av.ViewFilter{view.ID: {filter}},
		Templates: []*av.NewItemTemplate{{ID: "template", FieldValues: map[string]*av.NewItemFieldValue{key.ID: originalDefault}}},
	}
	if err := restoreAttributeViewColumn(database, snapshot); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(current.MSelect, original.MSelect) || key.Options[0].Name != "To do" || view.Filters[0] != filter || template.FieldValues[key.ID] != originalDefault {
		t.Fatal("undo must restore the category, cells, filters and template defaults together")
	}
	if added.MSelect[0].Content != "Done" || template.FieldValues["other"] != unrelatedDefault {
		t.Fatal("restoration must preserve unrelated cells and template fields")
	}
}
