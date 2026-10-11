// Scribli contributors. SPDX-License-Identifier: AGPL-3.0-or-later

package model

import (
	"errors"

	"github.com/icha-senpai/note/kernel/av"
	"github.com/icha-senpai/note/third_party/forks/gulu"
)

func (tx *Transaction) doSetAttrViewPresentation(operation *Operation) *TxErr {
	if operation.BlockID != "" {
		if _, err := tx.loadTree(operation.BlockID); err != nil {
			return &TxErr{code: TxErrHandleAttributeView, id: operation.AvID, msg: err.Error()}
		}
	}
	database, err := avParseView(operation.AvID, operation.BlockID)
	if err == nil {
		view := database.GetView(operation.ViewID)
		if view == nil {
			err = av.ErrViewNotFound
		} else {
			payload, marshalErr := gulu.JSON.MarshalJSON(operation.Data)
			err = marshalErr
			if err == nil && operation.Action == "setAttrViewCalendar" {
				config := &av.CalendarConfig{}
				err = gulu.JSON.UnmarshalJSON(payload, config)
				if err == nil && config.DateKeyID != "" {
					var key *av.Key
					key, err = database.GetKey(config.DateKeyID)
					if err == nil && key.Type != av.KeyTypeDate && key.Type != av.KeyTypeCreated && key.Type != av.KeyTypeUpdated {
						err = errors.New("calendar requires a date field")
					}
				}
				if err == nil {
					view.Calendar = config
				}
			} else if err == nil {
				var rules []*av.ConditionalColorRule
				err = gulu.JSON.UnmarshalJSON(payload, &rules)
				if err == nil {
					err = av.ValidateConditionalColors(database, rules)
				}
				if err == nil {
					view.ConditionalColors = rules
				}
			}
			if err == nil {
				err = avSaveView(database, operation.BlockID)
			}
		}
	}
	if err != nil {
		return &TxErr{code: TxErrHandleAttributeView, id: operation.AvID, msg: err.Error()}
	}
	return nil
}
