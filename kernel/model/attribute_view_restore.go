// Scribli - Refactor your thinking
// Copyright (c) 2020-present, Scribli contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package model

import (
	"errors"

	"github.com/icha-senpai/note/kernel/av"
	"github.com/icha-senpai/note/third_party/forks/gulu"
	"github.com/icha-senpai/note/third_party/forks/lute/parse"
)

func (tx *Transaction) doRestoreAttrViewView(operation *Operation) *TxErr {
	attrView, err := avParseView(operation.AvID, operation.BlockID)
	if err == nil {
		var data []byte
		data, err = gulu.JSON.MarshalJSON(operation.Data)
		if err == nil {
			view := &av.View{}
			err = gulu.JSON.UnmarshalJSON(data, view)
			if err == nil && view.ID != operation.ID {
				err = errors.New("database view snapshot ID mismatch")
			}
			if err == nil {
				err = restoreAttributeViewView(attrView, view, operation.PreviousID)
			}
			if err == nil {
				err = avSaveView(attrView, operation.BlockID)
			}
			if err == nil {
				node, tree, _ := getNodeByBlockID(nil, operation.BlockID)
				if node == nil {
					err = errors.New("database block not found")
				} else {
					attrs := parse.IAL2Map(node.KramdownIAL)
					attrs[av.NodeAttrView] = view.ID
					node.AttributeViewType = string(view.LayoutType)
					err = setNodeAttrs(node, tree, attrs)
				}
			}
		}
	}
	if err != nil {
		return &TxErr{code: TxErrHandleAttributeView, id: operation.AvID, msg: err.Error()}
	}
	operation.RetData = attrView.GetView(operation.ID).LayoutType
	return nil
}

func restoreAttributeViewView(attrView *av.AttributeView, view *av.View, previousID string) error {
	if view.ID == "" || (view.LayoutType != av.LayoutTypeTable && view.LayoutType != av.LayoutTypeGallery && view.LayoutType != av.LayoutTypeKanban) {
		return errors.New("invalid database view snapshot")
	}
	if attrView.GetView(view.ID) != nil {
		return errors.New("database view already exists")
	}
	index := 0
	if previousID != "" {
		index = len(attrView.Views)
		for i, item := range attrView.Views {
			if item.ID == previousID {
				index = i + 1
				break
			}
		}
	}
	attrView.Views = append(attrView.Views, nil)
	copy(attrView.Views[index+1:], attrView.Views[index:])
	attrView.Views[index] = view
	attrView.ViewID = view.ID
	return nil
}

type attributeViewColumnSnapshot struct {
	KeyID     string                      `json:"keyID"`
	Options   []*av.SelectOption          `json:"options"`
	Values    []*av.Value                 `json:"values"`
	Filters   map[string][]*av.ViewFilter `json:"filters"`
	Templates []*av.NewItemTemplate       `json:"templates"`
}

func (tx *Transaction) doRestoreAttrViewColOption(operation *Operation) *TxErr {
	attrView, err := avParseView(operation.AvID, operation.BlockID)
	if err == nil {
		var data []byte
		data, err = gulu.JSON.MarshalJSON(operation.Data)
		if err == nil {
			snapshot := &attributeViewColumnSnapshot{}
			err = gulu.JSON.UnmarshalJSON(data, snapshot)
			if err == nil {
				err = restoreAttributeViewColumn(attrView, snapshot)
			}
			if err == nil {
				regenAttrViewGroups(attrView)
				err = avSaveView(attrView, operation.BlockID)
			}
		}
	}
	if err != nil {
		return &TxErr{code: TxErrHandleAttributeView, id: operation.AvID, msg: err.Error()}
	}
	return nil
}

func restoreAttributeViewColumn(attrView *av.AttributeView, snapshot *attributeViewColumnSnapshot) error {
	key, err := attrView.GetKey(snapshot.KeyID)
	if err != nil {
		return err
	}
	if key.Type != av.KeyTypeSelect && key.Type != av.KeyTypeMSelect {
		return errors.New("invalid select column snapshot")
	}
	key.Options = snapshot.Options
	values := map[string]*av.Value{}
	for _, value := range snapshot.Values {
		if value != nil {
			values[value.ID] = value
		}
	}
	for _, kv := range attrView.KeyValues {
		if kv.Key.ID != key.ID {
			continue
		}
		for _, value := range kv.Values {
			if value != nil && values[value.ID] != nil {
				value.MSelect = values[value.ID].MSelect
			}
		}
	}
	for _, view := range attrView.Views {
		if filters, ok := snapshot.Filters[view.ID]; ok {
			view.Filters = filters
		}
	}
	for _, current := range attrView.NewItemTemplates {
		for _, previous := range snapshot.Templates {
			if current.ID != previous.ID {
				continue
			}
			if field, ok := previous.FieldValues[key.ID]; ok {
				if current.FieldValues == nil {
					current.FieldValues = map[string]*av.NewItemFieldValue{}
				}
				current.FieldValues[key.ID] = field
			}
		}
	}
	return nil
}
