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

package av

import (
	"github.com/icha-senpai/note/third_party/forks/lute/ast"
)

type LayoutKanban struct {
	*BaseLayout

	CoverFrom           CoverFrom       `json:"coverFrom"`
	CoverFromAssetKeyID string          `json:"coverFromAssetKeyID,omitempty"`
	CardAspectRatio     CardAspectRatio `json:"cardAspectRatio"`
	CardSize            CardSize        `json:"cardSize"`
	FitImage            bool            `json:"fitImage"`
	DisplayFieldName    bool            `json:"displayFieldName"`

	FillColBackgroundColor bool `json:"fillColBackgroundColor"`

	Fields []*ViewKanbanField `json:"fields"`
}

func NewLayoutKanban() *LayoutKanban {
	return &LayoutKanban{
		BaseLayout: &BaseLayout{
			Spec:     0,
			ID:       ast.NewNodeID(),
			ShowIcon: true,
		},
		CoverFrom:       CoverFromContentImage,
		CardAspectRatio: CardAspectRatio16_9,
		CardSize:        CardSizeMedium,
	}
}

type ViewKanbanField struct {
	*BaseField
}

type Kanban struct {
	*BaseInstance

	CoverFrom              CoverFrom       `json:"coverFrom"`
	CoverFromAssetKeyID    string          `json:"coverFromAssetKeyID,omitempty"`
	CardAspectRatio        CardAspectRatio `json:"cardAspectRatio"`
	CardSize               CardSize        `json:"cardSize"`
	FitImage               bool            `json:"fitImage"`
	DisplayFieldName       bool            `json:"displayFieldName"`
	FillColBackgroundColor bool            `json:"fillColBackgroundColor"`
	Fields                 []*KanbanField  `json:"fields"`
	Cards                  []*KanbanCard   `json:"cards"`
	CardCount              int             `json:"cardCount"`
}

type KanbanCard struct {
	ID     string              `json:"id"`
	Values []*KanbanFieldValue `json:"values"`

	CoverURL     string `json:"coverURL"`
	CoverContent string `json:"coverContent"`
}

type KanbanField struct {
	*BaseInstanceField
}

type KanbanFieldValue struct {
	*BaseValue
}

func (card *KanbanCard) GetID() string {
	return card.ID
}

func (card *KanbanCard) GetBlockValue() (ret *Value) {
	for _, v := range card.Values {
		if KeyTypeBlock == v.ValueType {
			ret = v.Value
			break
		}
	}
	return
}

func (card *KanbanCard) GetValues() (ret []*Value) {
	ret = []*Value{}
	for _, v := range card.Values {
		ret = append(ret, v.Value)
	}
	return
}

func (card *KanbanCard) GetValue(keyID string) (ret *Value) {
	for _, value := range card.Values {
		if nil != value.Value && keyID == value.Value.KeyID {
			ret = value.Value
			break
		}
	}
	return
}

func (kanban *Kanban) GetItems() (ret []Item) {
	ret = []Item{}
	for _, card := range kanban.Cards {
		ret = append(ret, card)
	}
	return
}

func (kanban *Kanban) SetItems(items []Item) {
	kanban.Cards = []*KanbanCard{}
	for _, item := range items {
		kanban.Cards = append(kanban.Cards, item.(*KanbanCard))
	}
}

func (kanban *Kanban) CountItems() int {
	return len(kanban.Cards)
}

func (kanban *Kanban) GetFields() (ret []Field) {
	ret = []Field{}
	for _, field := range kanban.Fields {
		ret = append(ret, field)
	}
	return ret
}

func (kanban *Kanban) GetField(id string) (ret Field, fieldIndex int) {
	for i, field := range kanban.Fields {
		if field.ID == id {
			return field, i
		}
	}
	return nil, -1
}

func (kanban *Kanban) GetValue(itemID, keyID string) (ret *Value) {
	for _, card := range kanban.Cards {
		if card.ID == itemID {
			return card.GetValue(keyID)
		}
	}
	return nil
}

func (kanban *Kanban) GetType() LayoutType {
	return LayoutTypeKanban
}
