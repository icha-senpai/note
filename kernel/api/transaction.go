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

package api

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/icha-senpai/note/kernel/model"
	"github.com/icha-senpai/note/kernel/util"
	"github.com/icha-senpai/note/third_party/forks/gulu"
)

func performTransactions(c *gin.Context) {
	start := time.Now()
	ret := gulu.Ret.NewResult()
	defer c.JSON(http.StatusOK, ret)

	arg, ok := util.JsonArg(c, ret)
	if !ok {
		return
	}

	var trans []any
	var reqID float64
	var app, session string
	if !util.ParseJsonArgs(arg, ret,
		util.BindJsonArg("transactions", &trans, true, true),
		util.BindJsonArg("reqId", &reqID, true, false),
		util.BindJsonArg("app", &app, false, false),
		util.BindJsonArg("session", &session, false, false),
	) {
		return
	}

	if !util.IsBooted() {
		ret.Code = -1
		ret.Msg = fmt.Sprintf(model.Conf.Language(74), int(util.GetBootProgress()))
		ret.Data = map[string]any{"closeTimeout": 5000}
		return
	}

	data, err := gulu.JSON.MarshalJSON(trans)
	if err != nil {
		ret.Code = -1
		ret.Msg = "parses request failed"
		return
	}

	timestamp := int64(reqID)
	var transactions []*model.Transaction
	if err = gulu.JSON.UnmarshalJSON(data, &transactions); err != nil {
		ret.Code = -1
		ret.Msg = "parses request failed"
		return
	}
	for _, transaction := range transactions {
		transaction.Timestamp = timestamp
		transaction.MarkFromAPI()
	}

	model.PerformTransactions(&transactions)

	ret.Data = transactions

	pushTransactions(app, session, transactions)

	if model.IsMoveOutlineHeading(&transactions) {
		if retData := transactions[0].DoOperations[0].RetData; nil != retData {
			util.PushReloadDoc(retData.(string))
		}
	}

	elapsed := time.Since(start).Milliseconds()
	c.Header("Server-Timing", fmt.Sprintf("total;dur=%d", elapsed))
}

func pushTransactions(app, session string, transactions []*model.Transaction) {
	pushMode := util.PushModeBroadcastExcludeSelf
	if 0 < len(transactions) && 0 < len(transactions[0].DoOperations) {
		model.FlushTxQueue()

		if shouldBroadcastAttrViewTransactions(transactions) {
			pushMode = util.PushModeBroadcast
		}
	}

	evt := util.NewCmdResult("transactions", 0, pushMode)
	evt.AppId = app
	evt.SessionId = session
	evt.Data = transactions

	var rootIDs []string
	for _, tx := range transactions {
		rootIDs = append(rootIDs, tx.GetChangedRootIDs()...)
	}
	rootIDs = gulu.Str.RemoveDuplicatedElem(rootIDs)

	for _, tx := range transactions {
		tx.WaitForCommit()
	}

	undoStates := map[string]map[string]bool{}
	for _, rootID := range rootIDs {
		canUndo, canRedo, _ := model.GlobalUndoLog.State(rootID)
		undoStates[rootID] = map[string]bool{
			"canUndo": canUndo,
			"canRedo": canRedo,
		}
	}
	evt.Context = map[string]any{
		"rootIDs":   rootIDs,
		"undoState": undoStates,
	}

	util.PushEvent(evt)
}

func undoState(c *gin.Context) {
	ret := gulu.Ret.NewResult()
	defer c.JSON(http.StatusOK, ret)

	arg, ok := util.JsonArg(c, ret)
	if !ok {
		return
	}

	var rootID string
	if !util.ParseJsonArgs(arg, ret,
		util.BindJsonArg("rootID", &rootID, true, false),
	) {
		return
	}

	canUndo, canRedo, peekMutatedRootIDs := model.GlobalUndoLog.State(rootID)
	ret.Data = map[string]any{
		"canUndo":            canUndo,
		"canRedo":            canRedo,
		"peekMutatedRootIDs": peekMutatedRootIDs,
	}
}

func performUndo(c *gin.Context) {
	ret := gulu.Ret.NewResult()
	defer c.JSON(http.StatusOK, ret)

	arg, ok := util.JsonArg(c, ret)
	if !ok {
		return
	}

	var rootID, app, session string
	if !util.ParseJsonArgs(arg, ret,
		util.BindJsonArg("rootID", &rootID, true, false),
		util.BindJsonArg("app", &app, false, false),
		util.BindJsonArg("session", &session, false, false),
	) {
		return
	}

	entry := model.GlobalUndoLog.Undo(rootID)
	if nil == entry {

		ret.Data = map[string]any{
			"canUndo": false,
			"canRedo": false,
		}
		return
	}

	tx := &model.Transaction{
		Timestamp:      time.Now().UnixMilli(),
		DoOperations:   entry.UndoOperationsForReplay(),
		UndoOperations: entry.DoOperationsForReplay(),
	}
	tx.MarkReplay()

	model.ResolveReplayDuplicateIds(tx)

	if err := model.PerformTxSync(tx); nil != err {

		model.GlobalUndoLog.UndoRollback(entry, rootID)
		ret.Data = map[string]any{
			"failed": true,
			"msg":    "undo failed: " + err.Error(),
		}
		return
	}

	model.GlobalUndoLog.UndoCommit(entry, rootID)

	crossDoc := len(entry.MutatedRootIDs()) > 1
	pushUndoTransactions(app, session, []*model.Transaction{tx}, true, crossDoc)

	canUndo, canRedo, _ := model.GlobalUndoLog.State(rootID)

	ret.Data = map[string]any{
		"doOperations":   tx.DoOperations,
		"undoOperations": tx.UndoOperations,
		"mutatedRootIDs": entry.MutatedRootIDs(),
		"canUndo":        canUndo,
		"canRedo":        canRedo,
		"isUndo":         true,
	}
}

func performRedo(c *gin.Context) {
	ret := gulu.Ret.NewResult()
	defer c.JSON(http.StatusOK, ret)

	arg, ok := util.JsonArg(c, ret)
	if !ok {
		return
	}

	var rootID, app, session string
	if !util.ParseJsonArgs(arg, ret,
		util.BindJsonArg("rootID", &rootID, true, false),
		util.BindJsonArg("app", &app, false, false),
		util.BindJsonArg("session", &session, false, false),
	) {
		return
	}

	entry := model.GlobalUndoLog.Redo(rootID)
	if nil == entry {
		ret.Data = map[string]any{
			"canUndo": false,
			"canRedo": false,
		}
		return
	}

	tx := &model.Transaction{
		Timestamp:      time.Now().UnixMilli(),
		DoOperations:   entry.DoOperationsForReplay(),
		UndoOperations: entry.UndoOperationsForReplay(),
	}
	tx.MarkReplay()

	model.ResolveReplayDuplicateIds(tx)

	if err := model.PerformTxSync(tx); nil != err {

		model.GlobalUndoLog.RedoRollback(entry, rootID)
		ret.Data = map[string]any{
			"failed": true,
			"msg":    "redo failed: " + err.Error(),
		}
		return
	}

	model.GlobalUndoLog.RedoCommit(entry, rootID)

	crossDoc := len(entry.MutatedRootIDs()) > 1
	pushUndoTransactions(app, session, []*model.Transaction{tx}, true, crossDoc)

	canUndo, canRedo, _ := model.GlobalUndoLog.State(rootID)

	ret.Data = map[string]any{
		"doOperations":   tx.DoOperations,
		"undoOperations": tx.UndoOperations,
		"mutatedRootIDs": entry.MutatedRootIDs(),
		"canUndo":        canUndo,
		"canRedo":        canRedo,
		"isUndo":         false,
	}
}

func clearHistory(c *gin.Context) {
	ret := gulu.Ret.NewResult()
	defer c.JSON(http.StatusOK, ret)

	arg, ok := util.JsonArg(c, ret)
	if !ok {
		return
	}

	var rootID string
	if !util.ParseJsonArgs(arg, ret, util.BindJsonArg("rootID", &rootID, false, false)) {
		return
	}

	model.GlobalUndoLog.Clear(rootID)
}

func pushUndoTransactions(app, session string, transactions []*model.Transaction, isReplay, includeSelf bool) {
	pushMode := util.PushModeBroadcastExcludeSelf
	if includeSelf {
		pushMode = util.PushModeBroadcast
	}
	if !includeSelf && 0 < len(transactions) && 0 < len(transactions[0].DoOperations) {
		if shouldBroadcastAttrViewTransactions(transactions) {
			pushMode = util.PushModeBroadcast
		}
	}

	evt := util.NewCmdResult("transactions", 0, pushMode)
	evt.AppId = app
	evt.SessionId = session
	evt.Data = transactions

	var rootIDs []string
	for _, tx := range transactions {
		rootIDs = append(rootIDs, tx.GetChangedRootIDs()...)
	}
	rootIDs = gulu.Str.RemoveDuplicatedElem(rootIDs)

	undoStates := map[string]map[string]bool{}
	for _, rootID := range rootIDs {
		canUndo, canRedo, _ := model.GlobalUndoLog.State(rootID)
		undoStates[rootID] = map[string]bool{
			"canUndo": canUndo,
			"canRedo": canRedo,
		}
	}
	evt.Context = map[string]any{
		"rootIDs":      rootIDs,
		"undoState":    undoStates,
		"isUndoReplay": isReplay,
	}

	for _, tx := range transactions {
		tx.WaitForCommit()
	}
	util.PushEvent(evt)
}

func shouldBroadcastAttrViewTransactions(transactions []*model.Transaction) bool {
	for _, tx := range transactions {
		for _, operation := range tx.DoOperations {
			if nil != operation && "setAttrViewName" != operation.Action && strings.Contains(strings.ToLower(operation.Action), "attrview") {
				return true
			}
		}
	}
	return false
}
