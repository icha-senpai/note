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
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/icha-senpai/note/third_party/forks/gulu"
	"github.com/icha-senpai/note/third_party/forks/lute/ast"
	"github.com/icha-senpai/note/third_party/forks/github/goccy/go-json"
	"github.com/icha-senpai/note/kernel/cache"
	"github.com/icha-senpai/note/kernel/util"
	"github.com/icha-senpai/note/third_party/forks/filelock"
	"github.com/icha-senpai/note/third_party/forks/logging"
	jsoniter "github.com/icha-senpai/note/third_party/forks/github/json-iterator/go"
)

type AttributeView struct {
	Spec              int                `json:"spec"`
	ID                string             `json:"id"`
	Name              string             `json:"name"`
	KeyValues         []*KeyValues       `json:"keyValues"`
	KeyIDs            []string           `json:"keyIDs"`
	ViewID            string             `json:"viewID"`
	Views             []*View            `json:"views"`
	NewItemTemplates  []*NewItemTemplate `json:"newItemTemplates,omitempty"`
	DefaultTemplateID string             `json:"defaultTemplateID,omitempty"`

	RenderedViewables map[string]Viewable `json:"-"`
}

type NewItemTargetType string

const (
	NewItemTargetDetached NewItemTargetType = "detached"
	NewItemTargetDocument NewItemTargetType = "document"
)

type NewItemSaveLocation struct {
	BoxID        string `json:"boxID,omitempty"`
	PathTemplate string `json:"pathTemplate"`
}

type NewItemFieldValueMode string

const (
	NewItemFieldValueStatic      NewItemFieldValueMode = "static"
	NewItemFieldValueCurrentTime NewItemFieldValueMode = "currentTime"
)

type NewItemFieldValue struct {
	Mode  NewItemFieldValueMode `json:"mode"`
	Value *Value                `json:"value,omitempty"`
}

type NewItemTemplate struct {
	ID                  string                        `json:"id"`
	Name                string                        `json:"name"`
	Icon                string                        `json:"icon,omitempty"`
	TargetType          NewItemTargetType             `json:"targetType"`
	PrimaryKeyTemplate  string                        `json:"primaryKeyTemplate,omitempty"`
	FieldValues         map[string]*NewItemFieldValue `json:"fieldValues,omitempty"`
	SaveLocation        *NewItemSaveLocation          `json:"saveLocation,omitempty"`
	ContentTemplatePath string                        `json:"contentTemplatePath,omitempty"`
}

type NewItemTemplatesConfig struct {
	Templates         []*NewItemTemplate `json:"templates"`
	DefaultTemplateID string             `json:"defaultTemplateID,omitempty"`
}

type KeyValues struct {
	Key    *Key     `json:"key"`
	Values []*Value `json:"values,omitempty"`
}

func (kValues *KeyValues) GetValue(blockID string) (ret *Value) {
	for _, v := range kValues.Values {
		if v.BlockID == blockID {
			ret = v
			return
		}
	}
	return
}

func (kValues *KeyValues) GetBlockValue() (ret *Value) {
	for _, v := range kValues.Values {
		if KeyTypeBlock == v.Type {
			ret = v
			return
		}
	}
	return
}

func GetValue(keyValues []*KeyValues, keyID, itemID string) (ret *Value) {
	for _, kv := range keyValues {
		if kv.Key.ID == keyID {
			for _, v := range kv.Values {
				if v.BlockID == itemID {
					ret = v
					return
				}
			}
		}
	}
	return
}

type KeyType string

const (
	KeyTypeBlock      KeyType = "block"
	KeyTypeText       KeyType = "text"
	KeyTypeNumber     KeyType = "number"
	KeyTypeDate       KeyType = "date"
	KeyTypeSelect     KeyType = "select"
	KeyTypeMSelect    KeyType = "mSelect"
	KeyTypeURL        KeyType = "url"   // URL
	KeyTypeEmail      KeyType = "email" // Email
	KeyTypePhone      KeyType = "phone"
	KeyTypeMAsset     KeyType = "mAsset"
	KeyTypeTemplate   KeyType = "template"
	KeyTypeCreated    KeyType = "created"
	KeyTypeUpdated    KeyType = "updated"
	KeyTypeCheckbox   KeyType = "checkbox"
	KeyTypeRelation   KeyType = "relation"
	KeyTypeRollup     KeyType = "rollup"
	KeyTypeLineNumber KeyType = "lineNumber"
)

type Key struct {
	ID   string  `json:"id"`
	Name string  `json:"name"`
	Type KeyType `json:"type"`
	Icon string  `json:"icon"`
	Desc string  `json:"desc"`

	Options []*SelectOption `json:"options,omitempty"`

	NumberFormat NumberFormat `json:"numberFormat"`

	Template string `json:"template"`

	Relation *Relation `json:"relation,omitempty"`

	Rollup *Rollup `json:"rollup,omitempty"`

	Date *Date `json:"date,omitempty"`

	Created *Created `json:"created,omitempty"`

	Updated *Updated `json:"updated,omitempty"`
}

func NewKey(id, name, icon string, keyType KeyType) *Key {
	return &Key{
		ID:   id,
		Name: name,
		Type: keyType,
		Icon: icon,
	}
}

func (k *Key) GetOption(name string) (ret *SelectOption) {
	for _, option := range k.Options {
		if option.Name == name {
			ret = option
			return
		}
	}
	return
}

type Created struct {
	IncludeTime bool `json:"includeTime"`
}

type Updated struct {
	IncludeTime bool `json:"includeTime"`
}

type Date struct {
	AutoFillNow      bool `json:"autoFillNow"`
	FillSpecificTime bool `json:"fillSpecificTime"`
}

type Rollup struct {
	RelationKeyID string      `json:"relationKeyID"`
	KeyID         string      `json:"keyID"`
	Calc          *RollupCalc `json:"calc"`
}

type RollupCalc struct {
	Operator CalcOperator `json:"operator"`
	Result   *Value       `json:"result"`
}

type Relation struct {
	AvID      string `json:"avID"`
	IsTwoWay  bool   `json:"isTwoWay"`
	BackKeyID string `json:"backKeyID"`
}

type SelectOption struct {
	Name  string `json:"name"`
	Color string `json:"color"`
	Desc  string `json:"desc"`
}

type View struct {
	ID               string         `json:"id"`
	Icon             string         `json:"icon"`
	Name             string         `json:"name"`
	HideAttrViewName bool           `json:"hideAttrViewName"`
	Desc             string         `json:"desc"`
	Filters          []*ViewFilter  `json:"filters,omitempty"`
	Sorts            []*ViewSort    `json:"sorts,omitempty"`
	PageSize         int            `json:"pageSize"`
	LayoutType       LayoutType     `json:"type"`
	Table            *LayoutTable   `json:"table,omitempty"`
	Gallery          *LayoutGallery `json:"gallery,omitempty"`
	Kanban           *LayoutKanban  `json:"kanban,omitempty"`
	ItemIDs          []string       `json:"itemIds,omitempty"`

	Group        *ViewGroup `json:"group,omitempty"`
	GroupCreated int64      `json:"groupCreated"`
	Groups       []*View    `json:"groups,omitempty"`
	GroupItemIDs []string   `json:"groupItemIds"`
	GroupCalc    *GroupCalc `json:"groupCalc,omitempty"`
	GroupKey     *Key       `json:"groupKey,omitempty"`
	GroupVal     *Value     `json:"groupVal,omitempty"`
	GroupFolded  bool       `json:"groupFolded"`
	GroupHidden  int        `json:"groupHidden"`
	GroupSort    int        `json:"groupSort"`
}

type ViewData struct {
	ID               string     `json:"id"`
	Icon             string     `json:"icon"`
	Name             string     `json:"name"`
	Desc             string     `json:"desc"`
	HideAttrViewName bool       `json:"hideAttrViewName"`
	Type             LayoutType `json:"type"`
	PageSize         int        `json:"pageSize"`
}

func (view *View) IsGroupView() bool {
	return nil != view.Group && "" != view.Group.Field
}

func (view *View) GetGroupValue() string {
	if nil == view.GroupVal {
		return ""
	}
	return view.GroupVal.String(false)
}

func (view *View) GetGroupByID(groupID string) *View {
	if nil == view.Groups {
		return nil
	}
	for _, group := range view.Groups {
		if group.ID == groupID {
			return group
		}
	}
	return nil
}

func (view *View) GetGroupByGroupValue(groupVal string) *View {
	if nil == view.Groups {
		return nil
	}
	for _, group := range view.Groups {
		if group.GetGroupValue() == groupVal {
			return group
		}
	}
	return nil
}

func (view *View) RemoveGroupByID(groupID string) {
	if nil == view.Groups {
		return
	}
	for i, group := range view.Groups {
		if group.ID == groupID {
			view.Groups = append(view.Groups[:i], view.Groups[i+1:]...)
			return
		}
	}
}

func (view *View) GetGroupKey(attrView *AttributeView) (ret *Key) {
	if !view.IsGroupView() {
		return
	}

	for _, kv := range attrView.KeyValues {
		if kv.Key.ID == view.Group.Field {
			ret = kv.Key
			return
		}
	}
	return
}

type GroupCalc struct {
	Field     string     `json:"field"`
	FieldCalc *FieldCalc `json:"calc"`
}

type LayoutType string

const (
	LayoutTypeTable   LayoutType = "table"
	LayoutTypeGallery LayoutType = "gallery"
	LayoutTypeKanban  LayoutType = "kanban"
)

const (
	ViewDefaultPageSize = 50
)

func NewTableView() *View {
	return &View{
		ID:         ast.NewNodeID(),
		Name:       GetAttributeViewI18n("table"),
		Filters:    []*ViewFilter{{Combination: FilterCombinationAnd}},
		Sorts:      []*ViewSort{},
		PageSize:   ViewDefaultPageSize,
		LayoutType: LayoutTypeTable,
		Table:      NewLayoutTable(),
	}
}

func NewTableViewWithBlockKey(blockKeyID string) (view *View, blockKey, selectKey *Key) {
	name := GetAttributeViewI18n("table")
	view = &View{
		ID:         ast.NewNodeID(),
		Name:       name,
		Filters:    []*ViewFilter{{Combination: FilterCombinationAnd}},
		Sorts:      []*ViewSort{},
		LayoutType: LayoutTypeTable,
		Table:      NewLayoutTable(),
		PageSize:   ViewDefaultPageSize,
	}
	blockKey = NewKey(blockKeyID, GetAttributeViewI18n("key"), "", KeyTypeBlock)
	view.Table.Columns = []*ViewTableColumn{{BaseField: &BaseField{ID: blockKeyID}}}

	selectKey = NewKey(ast.NewNodeID(), GetAttributeViewI18n("select"), "", KeyTypeSelect)
	view.Table.Columns = append(view.Table.Columns, &ViewTableColumn{BaseField: &BaseField{ID: selectKey.ID}})
	return
}

func NewGalleryView() (ret *View) {
	return &View{
		ID:         ast.NewNodeID(),
		Name:       GetAttributeViewI18n("gallery"),
		Filters:    []*ViewFilter{{Combination: FilterCombinationAnd}},
		Sorts:      []*ViewSort{},
		PageSize:   ViewDefaultPageSize,
		LayoutType: LayoutTypeGallery,
		Gallery:    NewLayoutGallery(),
	}
}

func NewKanbanView() (ret *View) {
	return &View{
		ID:         ast.NewNodeID(),
		Name:       GetAttributeViewI18n("kanban"),
		Filters:    []*ViewFilter{{Combination: FilterCombinationAnd}},
		Sorts:      []*ViewSort{},
		PageSize:   ViewDefaultPageSize,
		LayoutType: LayoutTypeKanban,
		Kanban:     NewLayoutKanban(),
	}
}

type Viewable interface {
	GetType() LayoutType

	GetID() string

	SetGroups(viewables []Viewable)

	SetGroupCalc(group *GroupCalc)

	GetGroupCalc() *GroupCalc

	SetGroupFolded(folded bool)

	GetGroupHidden() int

	SetGroupHidden(hidden int)
}

func NewAttributeView(id string) (ret *AttributeView) {
	view, blockKey, selectKey := NewTableViewWithBlockKey(ast.NewNodeID())
	ret = &AttributeView{
		Spec:              CurrentSpec,
		ID:                id,
		KeyValues:         []*KeyValues{{Key: blockKey}, {Key: selectKey}},
		ViewID:            view.ID,
		Views:             []*View{view},
		RenderedViewables: map[string]Viewable{},
	}
	return
}

func GetAttributeViewName(avID string) (ret string, err error) {

	avJSONPath, boxID := FindAttributeViewPath(avID)
	if avJSONPath == "" {
		avJSONPath = GetAttributeViewDataPath(avID)
		boxID = ""
	}
	if !filelock.IsExist(avJSONPath) {
		return
	}

	return getAttributeViewNameByPathInBox(avJSONPath, boxID)
}

func getAttributeViewNameByPathInBox(avJSONPath, boxID string) (ret string, err error) {
	data, err := filelock.ReadFile(avJSONPath)
	if err != nil {
		logging.LogErrorf("read attribute view [%s] failed: %s", avJSONPath, err)
		return
	}
	if boxID != "" {
		avID := strings.TrimSuffix(filepath.Base(avJSONPath), filepath.Ext(avJSONPath))
		plain, decErr := decryptAVData(boxID, avID, data)
		if decErr != nil {
			logging.LogErrorf("decrypt attribute view [%s] failed: %s", avJSONPath, decErr)
			return "", decErr
		}
		data = plain
	}

	val := jsoniter.Get(data, "name")
	if nil == val || val.ValueType() == jsoniter.InvalidValue {
		return
	}
	ret = val.ToString()
	return
}

func GetAttributeViewNameByPath(avJSONPath string) (ret string, err error) {
	return getAttributeViewNameByPathInBox(avJSONPath, "")
}

func GetAttributeViewNameInBox(avID, boxID string) (ret string, err error) {
	avJSONPath, _ := FindAttributeViewPathInBox(avID, boxID)
	if avJSONPath == "" {
		return
	}
	return getAttributeViewNameByPathInBox(avJSONPath, boxID)
}

func GetAttributeViewContent(avID string) (content string) {
	if "" == avID {
		return
	}

	attrView, err := ParseAttributeView(avID)
	if err != nil {
		logging.LogErrorf("parse attribute view [%s] failed: %s", avID, err)
		return
	}
	return getAttributeViewContent0(attrView)
}

func GetAttributeViewContentByPath(avJSONPath string) (content string) {
	attrView, err := ParseAttributeViewByPath(avJSONPath)
	if err != nil {
		logging.LogErrorf("parse attribute view [%s] failed: %s", avJSONPath, err)
		return
	}
	return getAttributeViewContent0(attrView)
}

func getAttributeViewContent0(attrView *AttributeView) (content string) {
	buf := bytes.Buffer{}
	buf.WriteString(attrView.Name)
	buf.WriteByte(' ')
	for _, v := range attrView.Views {
		buf.WriteString(v.Name)
		buf.WriteByte(' ')
	}

	for _, keyValues := range attrView.KeyValues {
		buf.WriteString(keyValues.Key.Name)
		buf.WriteByte(' ')
		for _, value := range keyValues.Values {
			if nil != value {
				buf.WriteString(value.String(true))
				buf.WriteByte(' ')
			}
		}
	}

	content = strings.TrimSpace(buf.String())
	return
}

func IsAttributeViewExist(avID string) bool {

	avJSONPath, _ := FindAttributeViewPath(avID)
	if avJSONPath == "" {
		avJSONPath = GetAttributeViewDataPath(avID)
	}
	return filelock.IsExist(avJSONPath)
}

func ParseAttributeView(avID string) (ret *AttributeView, err error) {
	if !ast.IsNodeIDPattern(avID) {
		err = ErrInvalidAttributeViewID
		return
	}

	avJSONPath, boxID := FindAttributeViewPath(avID)
	if avJSONPath == "" {

		avJSONPath = GetAttributeViewDataPath(avID)
		return parseAttributeViewByPathInBox(avJSONPath, "")
	}
	if boxID != "" {
		SetAVBoxID(avID, boxID)
	}
	return parseAttributeViewByPathInBox(avJSONPath, boxID)
}

func ParseAttributeViewInBox(avID, boxID string) (ret *AttributeView, err error) {
	if !ast.IsNodeIDPattern(avID) {
		err = ErrInvalidAttributeViewID
		return
	}
	if boxID != "" && !ast.IsNodeIDPattern(boxID) {
		err = ErrInvalidBoxID
		return
	}

	avJSONPath, avBoxID := FindAttributeViewPathInBox(avID, boxID)
	if avJSONPath == "" {
		avJSONPath = attributeViewDataPathByBox(avID, boxID)
		avBoxID = boxID
	} else {

		if boxID != "" {
			SetAVBoxID(avID, boxID)
		}
	}
	return parseAttributeViewByPathInBox(avJSONPath, avBoxID)
}

func ParseAttributeViewByPath(avJSONPath string) (ret *AttributeView, err error) {
	return parseAttributeViewByPathInBox(avJSONPath, avBoxIDFromPath(avJSONPath))
}

func parseAttributeViewByPathInBox(avJSONPath, boxID string) (ret *AttributeView, err error) {
	if !filelock.IsExist(avJSONPath) {
		err = ErrViewNotFound
		return
	}

	avID := filepath.Base(avJSONPath)
	avID = strings.TrimSuffix(avID, filepath.Ext(avID))

	var data []byte
	if cached, ok := cache.GetAVDataInBox(avID, boxID); ok {
		data = cached
	} else {
		var readErr error
		data, readErr = filelock.ReadFile(avJSONPath)
		if nil != readErr {
			logging.LogErrorf("read attribute view [%s] failed: %s", avID, readErr)
			return
		}

		if boxID != "" {
			data, readErr = decryptAVData(boxID, avID, data)
			if readErr != nil {
				logging.LogErrorf("decrypt attribute view [%s] failed: %s", avID, readErr)
				return
			}
		} else if util.IsCiphertext(data) {

			return
		}
		cache.SetAVDataInBox(avID, boxID, data)
	}

	ret = &AttributeView{RenderedViewables: map[string]Viewable{}}
	if err = json.Unmarshal(data, ret); err != nil {
		if strings.Contains(err.Error(), ".relation.contents of type av.Value") {
			mapAv := map[string]any{}
			if err = json.Unmarshal(data, &mapAv); err != nil {
				logging.LogErrorf("unmarshal attribute view [%s] failed: %s", avID, err)
				return
			}

			keyValues := mapAv["keyValues"]
			keyValuesMap := keyValues.([]any)
			for _, kv := range keyValuesMap {
				kvMap := kv.(map[string]any)
				if values := kvMap["values"]; nil != values {
					valuesMap := values.([]any)
					for _, v := range valuesMap {
						if vMap := v.(map[string]any); nil != vMap["relation"] {
							vMap["relation"].(map[string]any)["contents"] = nil
						}
					}
				}
			}

			views := mapAv["views"]
			viewsMap := views.([]any)
			for _, view := range viewsMap {
				if table := view.(map[string]any)["table"]; nil != table {
					tableMap := table.(map[string]any)
					if filters := tableMap["filters"]; nil != filters {
						filtersMap := filters.([]any)
						for _, f := range filtersMap {
							if fMap := f.(map[string]any); nil != fMap["value"] {
								if valueMap := fMap["value"].(map[string]any); nil != valueMap["relation"] {
									valueMap["relation"].(map[string]any)["contents"] = nil
								}
							}
						}
					}
				}
			}

			data, err = json.Marshal(mapAv)
			if err != nil {
				logging.LogErrorf("marshal attribute view [%s] failed: %s", avID, err)
				return
			}

			if err = json.Unmarshal(data, ret); err != nil {
				logging.LogErrorf("unmarshal attribute view [%s] failed: %s", avID, err)
				return
			}
		} else {
			logging.LogErrorf("unmarshal attribute view [%s] failed: %s", avID, err)
			return
		}
	}
	if nil == err {
		err = CheckSpec(ret)
	}
	return
}

func SaveAttributeView(av *AttributeView) (err error) {
	if !ast.IsNodeIDPattern(av.ID) {
		err = ErrInvalidAttributeViewID
		logging.LogErrorf("save attribute view failed: %s", err)
		return
	}

	UpgradeSpec(av)

	blockValues := av.GetBlockKeyValues()
	if nil != blockValues {
		blockIDs := map[string]bool{}
		var duplicatedValueIDs []string
		for _, blockValue := range blockValues.Values {
			if !blockIDs[blockValue.BlockID] {
				blockIDs[blockValue.BlockID] = true
			} else {
				duplicatedValueIDs = append(duplicatedValueIDs, blockValue.ID)
			}
		}
		var tmp []*Value
		for _, blockValue := range blockValues.Values {
			if !gulu.Str.Contains(blockValue.ID, duplicatedValueIDs) {
				tmp = append(tmp, blockValue)
			}
		}
		blockValues.Values = tmp
	}

	for _, view := range av.Views {

		view.ItemIDs = gulu.Str.RemoveDuplicatedElem(view.ItemIDs)

		if 1 > view.PageSize {
			view.PageSize = ViewDefaultPageSize
		}
	}

	for _, kv := range av.KeyValues {
		for i := len(kv.Values) - 1; i >= 0; i-- {
			if kv.Values[i].IsRenderAutoFill {
				kv.Values = append(kv.Values[:i], kv.Values[i+1:]...)
			}
		}
	}

	var data []byte
	if util.UseSingleLineSave {
		data, err = gulu.JSON.MarshalJSON(av)
	} else {
		data, err = gulu.JSON.MarshalIndentJSON(av, "", "\t")
	}
	if err != nil {
		logging.LogErrorf("marshal attribute view [%s] failed: %s", av.ID, err)
		return
	}

	avJSONPath, avBoxID := FindAttributeViewPath(av.ID)
	if avJSONPath == "" {

		avJSONPath = GetAttributeViewDataPath(av.ID)
	}
	if cachedData, ok := cache.GetAVDataInBox(av.ID, avBoxID); ok {
		if len(cachedData) == len(data) && bytes.Equal(cachedData, data) {
			return
		}
	} else {
		if diskData, readErr := filelock.ReadFile(avJSONPath); nil == readErr {

			if avBoxID != "" {
				diskData, _ = decryptAVData(avBoxID, av.ID, diskData)
			}
			if len(diskData) == len(data) && bytes.Equal(diskData, data) {
				cache.SetAVDataInBox(av.ID, avBoxID, data)
				return
			}
		}
	}

	writeData := data
	if avBoxID != "" {
		writeData, err = encryptAVData(avBoxID, av.ID, data)
		if err != nil {
			logging.LogErrorf("encrypt attribute view [%s] failed: %s", av.ID, err)
			return
		}
	}

	if err = os.MkdirAll(filepath.Dir(avJSONPath), 0755); nil != err {
		logging.LogErrorf("create attribute view dir [%s] failed: %s", filepath.Dir(avJSONPath), err)
		return
	}
	if err = util.WriteFileByMmap(avJSONPath, writeData); nil != err {
		if err = filelock.WriteFile(avJSONPath, writeData); nil != err {
			logging.LogErrorf("save attribute view [%s] failed: %s", av.ID, err)
			return
		}
	}

	cache.SetAVDataInBox(av.ID, avBoxID, data)

	if util.ExceedLargeFileWarningSize(len(data)) {
		msg := fmt.Sprintf(util.Langs[util.Lang][268], av.Name+" "+filepath.Base(avJSONPath), util.LargeFileWarningSize)
		util.PushErrMsg(msg, 7000)
	}
	return
}

func (av *AttributeView) GetView(viewID string) (ret *View) {
	for _, v := range av.Views {
		if v.ID == viewID {
			ret = v
			return
		}
	}
	return
}

func (av *AttributeView) GetCurrentView(viewID string) (ret *View, err error) {
	if "" != viewID {
		ret = av.GetView(viewID)
		if nil != ret {
			return
		}
	}

	for _, v := range av.Views {
		if v.ID == av.ViewID {
			ret = v
			return
		}
	}

	if 1 > len(av.Views) {
		err = ErrViewNotFound
		return
	}
	ret = av.Views[0]
	return
}

func (av *AttributeView) ExistBoundBlock(nodeID string) bool {
	for _, blockVal := range av.GetBlockKeyValues().Values {
		if blockVal.Block.ID == nodeID {
			return true
		}
	}
	return false
}

func (av *AttributeView) GetBlockValueByBoundID(nodeID string) *Value {
	for _, kv := range av.KeyValues {
		if KeyTypeBlock == kv.Key.Type {
			for _, v := range kv.Values {
				if v.Block.ID == nodeID {
					return v
				}
			}
		}
	}
	return nil
}

func (av *AttributeView) GetValue(keyID, itemID string) (ret *Value) {
	for _, kv := range av.KeyValues {
		if kv.Key.ID == keyID {
			for _, v := range kv.Values {
				if v.BlockID == itemID {
					ret = v
					return
				}
			}
		}
	}
	return
}

func (av *AttributeView) GetKey(keyID string) (ret *Key, err error) {
	for _, kv := range av.KeyValues {
		if kv.Key.ID == keyID {
			ret = kv.Key
			return
		}
	}
	err = ErrKeyNotFound
	return
}

func (av *AttributeView) GetBlockKeyValues() (ret *KeyValues) {
	for _, kv := range av.KeyValues {
		if KeyTypeBlock == kv.Key.Type {
			ret = kv
			return
		}
	}
	return
}

func (av *AttributeView) GetBlockValue(itemID string) (ret *Value) {
	for _, kv := range av.KeyValues {
		if KeyTypeBlock == kv.Key.Type && 0 < len(kv.Values) {
			for _, v := range kv.Values {
				if v.BlockID == itemID {
					ret = v
					return
				}
			}
		}
	}
	return
}

func (av *AttributeView) GetKeyValues(keyID string) (ret *KeyValues, err error) {
	for _, kv := range av.KeyValues {
		if kv.Key.ID == keyID {
			ret = kv
			return
		}
	}
	err = ErrKeyNotFound
	return
}

func (av *AttributeView) GetBlockKey() (ret *Key) {
	for _, kv := range av.KeyValues {
		if KeyTypeBlock == kv.Key.Type {
			ret = kv.Key
			return
		}
	}
	return
}

func (av *AttributeView) Clone() (ret *AttributeView) {
	ret = &AttributeView{}
	data, err := gulu.JSON.MarshalJSON(av)
	if err != nil {
		logging.LogErrorf("marshal attribute view [%s] failed: %s", av.ID, err)
		return nil
	}
	if err = gulu.JSON.UnmarshalJSON(data, ret); err != nil {
		logging.LogErrorf("unmarshal attribute view [%s] failed: %s", av.ID, err)
		return nil
	}

	ret.ID = ast.NewNodeID()
	templateIDMap := map[string]string{}
	for _, itemTemplate := range ret.NewItemTemplates {
		if nil == itemTemplate {
			continue
		}
		oldID := itemTemplate.ID
		itemTemplate.ID = ast.NewNodeID()
		templateIDMap[oldID] = itemTemplate.ID
	}
	ret.DefaultTemplateID = templateIDMap[ret.DefaultTemplateID]
	if 1 > len(ret.Views) {
		logging.LogErrorf("attribute view [%s] has no views", av.ID)
		return nil
	}

	var oldKeyIDs []string
	keyIDMap := map[string]string{}
	keyTypeMap := map[string]KeyType{}
	for _, kv := range ret.KeyValues {
		newID := ast.NewNodeID()
		keyIDMap[kv.Key.ID] = newID
		keyTypeMap[kv.Key.ID] = kv.Key.Type
		oldKeyIDs = append(oldKeyIDs, kv.Key.ID)
		kv.Key.ID = newID
		kv.Values = []*Value{}

		if KeyTypeRelation == kv.Key.Type {

			kv.Key.Relation.IsTwoWay = false
			kv.Key.Relation.AvID = ""
			kv.Key.Relation.BackKeyID = ""
		}
	}

	for _, itemTemplate := range ret.NewItemTemplates {
		if nil == itemTemplate {
			continue
		}
		fieldValues := map[string]*NewItemFieldValue{}
		for oldKeyID, fieldValue := range itemTemplate.FieldValues {
			newKeyID, ok := keyIDMap[oldKeyID]
			if !ok || KeyTypeRelation == keyTypeMap[oldKeyID] {
				continue
			}
			fieldValues[newKeyID] = fieldValue
		}
		if 0 == len(fieldValues) {
			itemTemplate.FieldValues = nil
		} else {
			itemTemplate.FieldValues = fieldValues
		}
	}

	oldKeyIDs = gulu.Str.RemoveDuplicatedElem(oldKeyIDs)
	sorts := map[string]int{}
	for i, k := range ret.KeyIDs {
		sorts[k] = i
	}
	sort.Slice(oldKeyIDs, func(i, j int) bool {
		return sorts[oldKeyIDs[i]] < sorts[oldKeyIDs[j]]
	})

	for _, view := range ret.Views {
		view.ID = ast.NewNodeID()

		remapFilterColumns(view.Filters, keyIDMap)
		for _, s := range view.Sorts {
			s.Column = keyIDMap[s.Column]
		}

		if nil != view.Group {
			view.Group.Field = keyIDMap[view.Group.Field]
		}

		switch view.LayoutType {
		case LayoutTypeTable:
			view.Table.ID = ast.NewNodeID()
			for _, column := range view.Table.Columns {
				column.ID = keyIDMap[column.ID]
			}
		case LayoutTypeGallery:
			view.Gallery.ID = ast.NewNodeID()
			for _, cardField := range view.Gallery.CardFields {
				cardField.ID = keyIDMap[cardField.ID]
			}
		case LayoutTypeKanban:
			view.Kanban.ID = ast.NewNodeID()
			for _, field := range view.Kanban.Fields {
				field.ID = keyIDMap[field.ID]
			}
		}
		view.ItemIDs = []string{}
	}
	ret.ViewID = ret.Views[0].ID

	ret.KeyIDs = nil
	for _, oldKeyID := range oldKeyIDs {
		newKeyID := keyIDMap[oldKeyID]
		ret.KeyIDs = append(ret.KeyIDs, newKeyID)
	}
	return
}

func GetAttributeViewDataPath(avID string) (ret string) {
	if !ast.IsNodeIDPattern(avID) {
		return
	}

	av := filepath.Join(util.DataDir, "storage", "av")
	ret = filepath.Join(av, avID+".json")
	if !gulu.File.IsDir(av) {
		if err := os.MkdirAll(av, 0755); err != nil {
			logging.LogErrorf("create attribute view dir failed: %s", err)
			return
		}
	}
	return
}

func GetAttributeViewI18n(key string) string {
	return util.AttrViewLangs[util.Lang][key].(string)
}

var (
	ErrAttributeViewNotFound  = errors.New("attribute view not found")
	ErrInvalidAttributeViewID = errors.New("invalid attribute view id")
	ErrInvalidBoxID           = errors.New("invalid box id")
	ErrViewNotFound           = errors.New("view not found")
	ErrKeyNotFound            = errors.New("key not found")
	ErrWrongLayoutType        = errors.New("wrong layout type")
	ErrInvalidColumnAlign     = errors.New("invalid column align")
	ErrSpecTooNew             = errors.New("attribute view spec is too new")
	ErrFilterTooDeep          = errors.New("filter nesting depth exceeds the maximum allowed")
)

const (
	NodeAttrNameAvs        = "custom-avs"
	NodeAttrView           = "custom-sy-av-view"
	NodeAttrViewStaticText = "custom-sy-av-s-text"

	NodeAttrViewNames = "av-names"
)
