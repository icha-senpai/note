// SiYuan - Refactor your thinking
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

type BaseLayout struct {
	Spec int    `json:"spec"`
	ID   string `json:"id"`

	ShowIcon  bool `json:"showIcon"`
	WrapField bool `json:"wrapField"`

	//Deprecated
	Filters []*ViewFilter `json:"filters,omitempty"`
	//Deprecated
	Sorts []*ViewSort `json:"sorts,omitempty"`
	//Deprecated
	PageSize int `json:"pageSize,omitempty"`
}

type BaseField struct {
	ID     string     `json:"id"`
	Wrap   bool       `json:"wrap"`
	Hidden bool       `json:"hidden"`
	Desc   string     `json:"desc,omitempty"`
	Calc   *FieldCalc `json:"calc,omitempty"`
}

type BaseValue struct {
	ID        string  `json:"id"`
	Value     *Value  `json:"value"`
	ValueType KeyType `json:"valueType"`
}

type BaseInstance struct {
	ID               string        `json:"id"` // ID
	Icon             string        `json:"icon"`
	Name             string        `json:"name"`
	Desc             string        `json:"desc"`
	HideAttrViewName bool          `json:"hideAttrViewName"`
	Filters          []*ViewFilter `json:"filters"`
	Sorts            []*ViewSort   `json:"sorts"`
	Group            *ViewGroup    `json:"group"`
	PageSize         int           `json:"pageSize"`
	ShowIcon         bool          `json:"showIcon"`
	WrapField        bool          `json:"wrapField"`

	GroupKey    *Key       `json:"groupKey,omitempty"`
	GroupValue  *Value     `json:"groupValue,omitempty"`
	Groups      []Viewable `json:"groups,omitempty"`
	GroupCalc   *GroupCalc `json:"groupCalc,omitempty"`
	GroupFolded bool       `json:"groupFolded"`
	GroupHidden int        `json:"groupHidden"`
}

func NewViewBaseInstance(view *View) *BaseInstance {
	showIcon, wrapField := true, false
	switch view.LayoutType {
	case LayoutTypeTable:
		showIcon = view.Table.ShowIcon
		wrapField = view.Table.WrapField
	case LayoutTypeGallery:
		showIcon = view.Gallery.ShowIcon
		wrapField = view.Gallery.WrapField
	case LayoutTypeKanban:
		showIcon = view.Kanban.ShowIcon
		wrapField = view.Kanban.WrapField
	}
	return &BaseInstance{
		ID:               view.ID,
		Icon:             view.Icon,
		Name:             view.Name,
		Desc:             view.Desc,
		HideAttrViewName: view.HideAttrViewName,
		Filters:          view.Filters,
		Sorts:            view.Sorts,
		Group:            view.Group,
		GroupKey:         view.GroupKey,
		GroupValue:       view.GroupVal,
		GroupCalc:        view.GroupCalc,
		GroupFolded:      view.GroupFolded,
		GroupHidden:      view.GroupHidden,
		PageSize:         view.PageSize,
		ShowIcon:         showIcon,
		WrapField:        wrapField,
	}
}

func (baseInstance *BaseInstance) GetSorts() []*ViewSort {
	return baseInstance.Sorts
}

func (baseInstance *BaseInstance) GetFilters() []*ViewFilter {
	return baseInstance.Filters
}

func (baseInstance *BaseInstance) SetGroups(viewables []Viewable) {
	baseInstance.Groups = viewables
}

func (baseInstance *BaseInstance) SetGroupCalc(group *GroupCalc) {
	baseInstance.GroupCalc = group
}

func (baseInstance *BaseInstance) GetGroupCalc() *GroupCalc {
	return baseInstance.GroupCalc
}

func (baseInstance *BaseInstance) SetGroupFolded(folded bool) {
	baseInstance.GroupFolded = folded
}

func (baseInstance *BaseInstance) GetGroupHidden() int {
	return baseInstance.GroupHidden
}

func (baseInstance *BaseInstance) SetGroupHidden(hidden int) {
	baseInstance.GroupHidden = hidden
}

func (baseInstance *BaseInstance) GetID() string {
	return baseInstance.ID
}

type BaseInstanceField struct {
	ID     string     `json:"id"` // ID
	Name   string     `json:"name"`
	Type   KeyType    `json:"type"`
	Icon   string     `json:"icon"`
	Wrap   bool       `json:"wrap"`
	Hidden bool       `json:"hidden"`
	Desc   string     `json:"desc"`
	Calc   *FieldCalc `json:"calc"`

	Options      []*SelectOption `json:"options,omitempty"`
	NumberFormat NumberFormat    `json:"numberFormat"`
	Template     string          `json:"template"`
	Relation     *Relation       `json:"relation,omitempty"`
	Rollup       *Rollup         `json:"rollup,omitempty"`
	Date         *Date           `json:"date,omitempty"`
	Created      *Created        `json:"created,omitempty"`
	Updated      *Updated        `json:"updated,omitempty"`
}

func (baseInstanceField *BaseInstanceField) GetID() string {
	return baseInstanceField.ID
}

func (baseInstanceField *BaseInstanceField) GetCalc() *FieldCalc {
	return baseInstanceField.Calc
}

func (baseInstanceField *BaseInstanceField) SetCalc(calc *FieldCalc) {
	baseInstanceField.Calc = calc
}

func (baseInstanceField *BaseInstanceField) GetType() KeyType {
	return baseInstanceField.Type
}

func (baseInstanceField *BaseInstanceField) GetNumberFormat() NumberFormat {
	return baseInstanceField.NumberFormat
}

type Collection interface {
	GetItems() (ret []Item)

	SetItems(items []Item)

	CountItems() int

	GetFields() []Field

	GetField(id string) (ret Field, fieldIndex int)

	GetValue(itemID, keyID string) (ret *Value)

	GetSorts() []*ViewSort

	GetFilters() []*ViewFilter
}

type Field interface {
	GetID() string

	GetType() KeyType

	GetCalc() *FieldCalc

	SetCalc(*FieldCalc)

	GetNumberFormat() NumberFormat
}

type Item interface {
	GetBlockValue() *Value

	GetValues() []*Value

	GetValue(keyID string) (ret *Value)

	GetID() string
}
