// Scribli contributors. SPDX-License-Identifier: AGPL-3.0-or-later
package av

import (
	"github.com/icha-senpai/note/kernel/util"
	"testing"
	"time"
)

func colorFixture() (*AttributeView, *Table) {
	database := &AttributeView{Spec: 5, KeyValues: []*KeyValues{{Key: NewKey("text", "Title", "", KeyTypeText)}, {Key: NewKey("date", "Date", "", KeyTypeDate)}}}
	today := time.Now()
	yesterday := today.AddDate(0, 0, -1).UnixMilli()
	future := today.AddDate(0, 0, 1).UnixMilli()
	table := &Table{BaseInstance: &BaseInstance{}, Columns: []*TableColumn{
		{BaseInstanceField: &BaseInstanceField{ID: "text", Type: KeyTypeText}}, {BaseInstanceField: &BaseInstanceField{ID: "date", Type: KeyTypeDate}},
	}}
	for id, date := range map[string]int64{"past": yesterday, "future": future} {
		table.Rows = append(table.Rows, &TableRow{ID: id, Cells: []*TableCell{
			{BaseValue: &BaseValue{ValueType: KeyTypeText, Value: &Value{KeyID: "text", Type: KeyTypeText, Text: &ValueText{Content: "entry"}}}},
			{BaseValue: &BaseValue{ValueType: KeyTypeDate, Value: &Value{KeyID: "date", Type: KeyTypeDate, Date: &ValueDate{Content: date, IsNotEmpty: true}}}},
		}})
	}
	return database, table
}

func TestConditionalColorsPrecedenceAndRelativeDates(t *testing.T) {
	database, table := colorFixture()
	past := &ConditionalColorRule{ID: "past-rule", Enabled: true, Color: "#ff0000", Filter: &ViewFilter{Column: "date", Operator: FilterOperatorIsLess,
		Value: &Value{Type: KeyTypeDate, Date: &ValueDate{}}, RelativeDate: &RelativeDate{Unit: RelativeDateUnitDay, Direction: RelativeDateDirectionThis}}}
	fallback := &ConditionalColorRule{ID: "fallback", Enabled: true, Color: "#00ff00", Filter: &ViewFilter{Column: "text", Operator: FilterOperatorContains, Value: &Value{Type: KeyTypeText, Text: &ValueText{Content: "entry"}}}}
	colors := ConditionalItemColors(table, database, []*ConditionalColorRule{past, fallback}, nil, nil)
	if colors["past"] != "#ff0000" || colors["future"] != "#00ff00" {
		t.Fatalf("wrong colors: %v", colors)
	}
	past.Enabled = false
	if colors = ConditionalItemColors(table, database, []*ConditionalColorRule{past, fallback}, nil, nil); colors["past"] != "#00ff00" {
		t.Fatal("disabled rule colored entry")
	}
	past.Enabled = true
	past.Filter.Column = "deleted"
	if colors = ConditionalItemColors(table, database, []*ConditionalColorRule{past}, nil, nil); len(colors) != 0 {
		t.Fatal("deleted field colored entries")
	}
	fallback.Filter = group(FilterCombinationAnd, fallback.Filter, &ViewFilter{Column: "date", Operator: FilterOperatorIsGreater, Value: &Value{Type: KeyTypeDate, Date: &ValueDate{}}, RelativeDate: &RelativeDate{Unit: RelativeDateUnitDay, Direction: RelativeDateDirectionThis}})
	if colors = ConditionalItemColors(table, database, []*ConditionalColorRule{fallback}, nil, nil); len(colors) != 1 || colors["future"] != "#00ff00" {
		t.Fatalf("nested conditions: %v", colors)
	}
}

func TestCalendarAndColorSpecCompatibility(t *testing.T) {
	database := &AttributeView{Spec: 5, Views: []*View{{LayoutType: LayoutTypeTable}}}
	UpgradeSpec(database)
	if database.Spec != 5 {
		t.Fatal("ordinary database was upgraded unnecessarily")
	}
	database.Views[0].LayoutType = LayoutTypeCalendar
	UpgradeSpec(database)
	if database.Spec != CurrentSpec {
		t.Fatal("calendar lacks compatibility barrier")
	}
	table := &Table{BaseInstance: NewViewBaseInstance(&View{LayoutType: LayoutTypeCalendar, Table: NewLayoutTable(), Calendar: &CalendarConfig{DateKeyID: "date"}})}
	if table.GetType() != LayoutTypeCalendar || table.Calendar.DateKeyID != "date" {
		t.Fatal("calendar projection lost its configuration")
	}
	database.Spec = 5
	database.Views[0].LayoutType = LayoutTypeTable
	database.Views[0].ConditionalColors = []*ConditionalColorRule{{ID: "rule"}}
	UpgradeSpec(database)
	if database.Spec != CurrentSpec {
		t.Fatal("stored colors lack compatibility barrier")
	}
}

func TestCalendarCloneRemapsPresentationFields(t *testing.T) {
	previousLangs := util.AttrViewLangs
	util.AttrViewLangs = map[string]map[string]any{util.Lang: {"table": "Table", "key": "Title", "select": "Select"}}
	t.Cleanup(func() { util.AttrViewLangs = previousLangs })
	database := NewAttributeView("20261010000000-abcdefg")
	date := NewKey("20261010000001-abcdefg", "Date", "", KeyTypeDate)
	database.KeyValues = append(database.KeyValues, &KeyValues{Key: date})
	view := database.Views[0]
	view.LayoutType = LayoutTypeCalendar
	view.Calendar = &CalendarConfig{DateKeyID: date.ID}
	view.Table.Columns = append(view.Table.Columns, &ViewTableColumn{BaseField: &BaseField{ID: date.ID}})
	view.ConditionalColors = []*ConditionalColorRule{{ID: "rule", Enabled: true, Color: "#123456", Filter: &ViewFilter{Column: date.ID, Operator: FilterOperatorIsNotEmpty, Value: &Value{Type: KeyTypeDate}}}}
	clone := database.Clone()
	if clone == nil || clone.Views[0].Calendar.DateKeyID == date.ID || clone.Views[0].ConditionalColors[0].Filter.Column != clone.Views[0].Calendar.DateKeyID || clone.Views[0].ConditionalColors[0].ID == "rule" {
		t.Fatal("clone retained source field IDs")
	}
	if database.Views[0].ConditionalColors[0].Filter.Column != date.ID {
		t.Fatal("clone changed source rule")
	}
}

func TestConditionalColorValidation(t *testing.T) {
	database, _ := colorFixture()
	rule := &ConditionalColorRule{ID: "rule", Enabled: true, Color: "url(unsafe)", Filter: &ViewFilter{Column: "text", Operator: FilterOperatorIsNotEmpty, Value: &Value{Type: KeyTypeText}}}
	if ValidateConditionalColors(database, []*ConditionalColorRule{rule}) == nil {
		t.Fatal("unsafe color accepted")
	}
	rule.Color = "#abcdef"
	if ValidateConditionalColors(database, []*ConditionalColorRule{rule, rule}) == nil {
		t.Fatal("duplicate rule IDs accepted")
	}
	rule.Filter.Column = "missing"
	if ValidateConditionalColors(database, []*ConditionalColorRule{rule}) == nil {
		t.Fatal("unknown field accepted")
	}
}

func TestCustomCategoryColorSpecCompatibility(t *testing.T) {
	database := &AttributeView{Spec: 5, KeyValues: []*KeyValues{{Key: &Key{Type: KeyTypeSelect, Options: []*SelectOption{{Name: "Old", Color: "14"}}}}}}
	UpgradeSpec(database)
	if database.Spec != 5 {
		t.Fatal("palette categories unnecessarily upgraded")
	}
	database.KeyValues[0].Key.Options[0].Color = "#aabbcc"
	UpgradeSpec(database)
	if database.Spec != CurrentSpec {
		t.Fatal("custom colors lack compatibility barrier")
	}
}

func TestSelectColorValidation(t *testing.T) {
	for _, color := range []string{"", "1", "14", "#abcdef", "#ABCDEF"} {
		if !ValidSelectColor(color) {
			t.Fatalf("valid color rejected: %q", color)
		}
	}
	for _, color := range []string{"0", "15", "#abc", "red", "1);color:red"} {
		if ValidSelectColor(color) {
			t.Fatalf("invalid color accepted: %q", color)
		}
	}
}
