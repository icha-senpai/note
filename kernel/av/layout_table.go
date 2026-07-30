// Scribli - Refactor your thinking
// Copyright (c) 2020-present, b3log.org
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

package av

import (
	"github.com/88250/lute/ast"
)

type TableColumnAlign string

const (
	TableColumnAlignDefault TableColumnAlign = ""
	TableColumnAlignLeft    TableColumnAlign = "left"
	TableColumnAlignCenter  TableColumnAlign = "center"
	TableColumnAlignRight   TableColumnAlign = "right"
)

func (align TableColumnAlign) IsValid() bool {
	return TableColumnAlignDefault == align || TableColumnAlignLeft == align || TableColumnAlignCenter == align ||
		TableColumnAlignRight == align
}

type LayoutTable struct {
	*BaseLayout

	Columns []*ViewTableColumn `json:"columns"`

	//Deprecated
	RowIDs []string `json:"rowIds"`
}

func NewLayoutTable() *LayoutTable {
	return &LayoutTable{
		BaseLayout: &BaseLayout{
			Spec:     0,
			ID:       ast.NewNodeID(),
			ShowIcon: true,
		},
	}
}

type ViewTableColumn struct {
	*BaseField

	Pin   bool             `json:"pin"`
	Width string           `json:"width"`
	Align TableColumnAlign `json:"align,omitempty"`
	Calc  *FieldCalc       `json:"calc,omitempty"`
}

type Table struct {
	*BaseInstance

	Columns  []*TableColumn `json:"columns"`
	Rows     []*TableRow    `json:"rows"`
	RowCount int            `json:"rowCount"`
}

type TableColumn struct {
	*BaseInstanceField

	Pin   bool             `json:"pin"`
	Width string           `json:"width"`
	Align TableColumnAlign `json:"align"`
}

type TableRow struct {
	ID    string       `json:"id"`
	Cells []*TableCell `json:"cells"`
}

type TableCell struct {
	*BaseValue

	Color   string `json:"color"`
	BgColor string `json:"bgColor"`
}

func (table *Table) GetColumn(id string) *TableColumn {
	for _, column := range table.Columns {
		if column.ID == id {
			return column
		}
	}
	return nil
}

func (row *TableRow) GetID() string {
	return row.ID
}

func (row *TableRow) GetBlockValue() (ret *Value) {
	for _, cell := range row.Cells {
		if KeyTypeBlock == cell.ValueType {
			ret = cell.Value
			break
		}
	}
	return
}

func (row *TableRow) GetValues() (ret []*Value) {
	ret = []*Value{}
	for _, cell := range row.Cells {
		if nil != cell.Value {
			ret = append(ret, cell.Value)
		}
	}
	return
}

func (row *TableRow) GetValue(keyID string) (ret *Value) {
	for _, cell := range row.Cells {
		if nil != cell.Value && keyID == cell.Value.KeyID {
			ret = cell.Value
			break
		}
	}
	return
}

func (table *Table) GetItems() (ret []Item) {
	ret = []Item{}
	for _, row := range table.Rows {
		if nil != row {
			ret = append(ret, row)
		}
	}
	return
}

func (table *Table) SetItems(items []Item) {
	table.Rows = []*TableRow{}
	for _, item := range items {
		table.Rows = append(table.Rows, item.(*TableRow))
	}
}

func (table *Table) CountItems() int {
	return len(table.Rows)
}

func (table *Table) GetFields() (ret []Field) {
	ret = []Field{}
	for _, column := range table.Columns {
		ret = append(ret, column)
	}
	return ret
}

func (table *Table) GetField(id string) (ret Field, fieldIndex int) {
	for _, column := range table.Columns {
		if column.ID == id {
			return column, fieldIndex
		}
	}
	return nil, -1
}

func (table *Table) GetValue(itemID, keyID string) (ret *Value) {
	for _, row := range table.Rows {
		if row.ID == itemID {
			return row.GetValue(keyID)
		}
	}
	return nil
}

func (*Table) GetType() LayoutType {
	return LayoutTypeTable
}
