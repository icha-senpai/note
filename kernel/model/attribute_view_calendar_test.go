// Scribli contributors. SPDX-License-Identifier: AGPL-3.0-or-later
package model

import (
	"fmt"
	"github.com/icha-senpai/note/kernel/av"
	"testing"
)

func TestCalendarRenderingIncludesEntriesBeyondTablePage(t *testing.T) {
	database := &av.AttributeView{KeyValues: []*av.KeyValues{{Key: av.NewKey("text", "Title", "", av.KeyTypeText)}}}
	view := &av.View{LayoutType: av.LayoutTypeCalendar, PageSize: 10, Table: av.NewLayoutTable(), Calendar: &av.CalendarConfig{}}
	table := &av.Table{BaseInstance: av.NewViewBaseInstance(view), Columns: []*av.TableColumn{{BaseInstanceField: &av.BaseInstanceField{ID: "text", Type: av.KeyTypeText}}}}
	for i := 0; i < 36; i++ {
		id := fmt.Sprintf("item-%d", i)
		table.Rows = append(table.Rows, &av.TableRow{ID: id, Cells: []*av.TableCell{{BaseValue: &av.BaseValue{ValueType: av.KeyTypeText, Value: &av.Value{KeyID: "text", Type: av.KeyTypeText, Text: &av.ValueText{Content: id}}}}}})
	}
	index, offset, err := renderViewableInstance(table, view, database, 1, 10, false, "item-35")
	if err != nil || len(table.Rows) != 36 || table.RowCount != 36 || index != 35 || offset != 0 {
		t.Fatalf("calendar lost later-page entries or target: rows=%d index=%d offset=%d error=%v", len(table.Rows), index, offset, err)
	}
}
