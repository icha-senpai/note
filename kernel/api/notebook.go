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
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/icha-senpai/note/kernel/model"
	"github.com/icha-senpai/note/kernel/util"
	"github.com/icha-senpai/note/third_party/forks/gulu"
)

func getNotebookInfo(c *gin.Context) {
	ret := gulu.Ret.NewResult()
	defer c.JSON(http.StatusOK, ret)

	arg, ok := util.JsonArg(c, ret)
	if !ok {
		return
	}

	var boxID string
	if !util.ParseJsonArgs(arg, ret, util.BindJsonArg("notebook", &boxID, true, true)) {
		return
	}
	if util.InvalidIDPattern(boxID, ret) {
		return
	}

	box := model.Conf.Box(boxID)
	if nil == box {
		ret.Code = -1
		ret.Msg = "notebook [" + boxID + "] not found"
		return
	}

	boxInfo := box.GetInfo()
	ret.Data = map[string]any{
		"boxInfo": boxInfo,
	}
}

func setNotebookIcon(c *gin.Context) {
	ret := gulu.Ret.NewResult()
	defer c.JSON(http.StatusOK, ret)

	arg, ok := util.JsonArg(c, ret)
	if !ok {
		return
	}

	var boxID, icon string
	if !util.ParseJsonArgs(arg, ret,
		util.BindJsonArg("notebook", &boxID, true, true),
		util.BindJsonArg("icon", &icon, true, false),
	) {
		return
	}
	model.SetBoxIcon(boxID, icon)
}

func changeSortNotebook(c *gin.Context) {
	ret := gulu.Ret.NewResult()
	defer c.JSON(http.StatusOK, ret)

	arg, ok := util.JsonArg(c, ret)
	if !ok {
		return
	}

	idsArg := arg["notebooks"].([]any)
	var ids []string
	for _, p := range idsArg {
		ids = append(ids, p.(string))
	}
	model.ChangeBoxSort(ids)
}

func renameNotebook(c *gin.Context) {
	ret := gulu.Ret.NewResult()
	defer c.JSON(http.StatusOK, ret)

	arg, ok := util.JsonArg(c, ret)
	if !ok {
		return
	}

	var notebook, name string
	if !util.ParseJsonArgs(arg, ret,
		util.BindJsonArg("notebook", &notebook, true, true),
		util.BindJsonArg("name", &name, true, false),
	) {
		return
	}
	if util.InvalidIDPattern(notebook, ret) {
		return
	}
	err := model.RenameBox(notebook, name)
	if err != nil {
		ret.Code = -1
		ret.Msg = err.Error()
		ret.Data = map[string]any{"closeTimeout": 5000}
		return
	}

	evt := util.NewCmdResult("renamenotebook", 0, util.PushModeBroadcast)
	evt.Data = map[string]any{
		"box":  notebook,
		"name": name,
	}
	util.PushEvent(evt)
}

func removeNotebook(c *gin.Context) {
	ret := gulu.Ret.NewResult()
	defer c.JSON(http.StatusOK, ret)

	arg, ok := util.JsonArg(c, ret)
	if !ok {
		return
	}

	var notebook string
	if !util.ParseJsonArgs(arg, ret, util.BindJsonArg("notebook", &notebook, true, true)) {
		return
	}
	if util.InvalidIDPattern(notebook, ret) {
		return
	}

	if util.ReadOnly {
		ret.Code = -1
		ret.Msg = model.Conf.Language(34)
		ret.Data = map[string]any{"closeTimeout": 5000}
		return
	}

	err := model.RemoveBox(notebook)
	if err != nil {
		ret.Code = -1
		ret.Msg = err.Error()
		return
	}

	evt := util.NewCmdResult("removeBox", 0, util.PushModeBroadcast)
	evt.Data = map[string]any{
		"box": notebook,
	}
	util.PushEvent(evt)
	model.TriggerOnboardingIfEmpty()
}

func createNotebook(c *gin.Context) {
	ret := gulu.Ret.NewResult()
	defer c.JSON(http.StatusOK, ret)

	arg, ok := util.JsonArg(c, ret)
	if !ok {
		return
	}

	var name string
	if !util.ParseJsonArgs(arg, ret, util.BindJsonArg("name", &name, true, false)) {
		return
	}
	id, err := model.CreateBox(name)
	if err != nil {
		ret.Code = -1
		ret.Msg = err.Error()
		return
	}

	existed, err := model.Mount(id)
	if err != nil {
		ret.Code = -1
		ret.Msg = err.Error()
		return
	}

	box := model.Conf.Box(id)
	if nil == box {
		ret.Code = -1
		ret.Msg = "opened notebook [" + id + "] not found"
		return
	}

	ret.Data = map[string]any{
		"notebook": box,
	}

	evt := util.NewCmdResult("createnotebook", 0, util.PushModeBroadcast)
	evt.Data = map[string]any{
		"box":     box,
		"existed": existed,
	}
	util.PushEvent(evt)
}

func openNotebook(c *gin.Context) {
	ret := gulu.Ret.NewResult()
	defer c.JSON(http.StatusOK, ret)

	arg, ok := util.JsonArg(c, ret)
	if !ok {
		return
	}

	var notebook string
	if !util.ParseJsonArgs(arg, ret, util.BindJsonArg("notebook", &notebook, true, true)) {
		return
	}
	if util.InvalidIDPattern(notebook, ret) {
		return
	}

	if util.ReadOnly {
		ret.Code = -1
		ret.Msg = model.Conf.Language(34)
		ret.Data = map[string]any{"closeTimeout": 5000}
		return
	}

	msgId := util.PushMsg(model.Conf.Language(45), 1000*60*15)
	defer util.PushClearMsg(msgId)
	existed, err := model.Mount(notebook)
	if err != nil {
		ret.Code = -1
		ret.Msg = err.Error()
		return
	}

	box := model.Conf.Box(notebook)
	if nil == box {
		ret.Code = -1
		ret.Msg = "opened notebook [" + notebook + "] not found"
		return
	}

	evt := util.NewCmdResult("mount", 0, util.PushModeBroadcast)
	evt.Data = map[string]any{
		"box":     box,
		"existed": existed,
	}
	util.PushEvent(evt)
}

func closeNotebook(c *gin.Context) {
	ret := gulu.Ret.NewResult()
	defer c.JSON(http.StatusOK, ret)

	arg, ok := util.JsonArg(c, ret)
	if !ok {
		return
	}

	notebook := arg["notebook"].(string)
	if util.InvalidIDPattern(notebook, ret) {
		return
	}
	model.Unmount(notebook)
}

func getNotebookConf(c *gin.Context) {
	ret := gulu.Ret.NewResult()
	defer c.JSON(http.StatusOK, ret)

	arg, ok := util.JsonArg(c, ret)
	if !ok {
		return
	}

	notebook := arg["notebook"].(string)
	if util.InvalidIDPattern(notebook, ret) {
		return
	}

	box := model.Conf.GetBox(notebook)
	if nil == box {
		ret.Code = -1
		ret.Msg = "notebook [" + notebook + "] not found"
		return
	}

	ret.Data = map[string]any{
		"box":  box.ID,
		"name": box.Name,
		"conf": box.GetConf(),
	}
}

func setNotebookConf(c *gin.Context) {
	ret := gulu.Ret.NewResult()
	defer c.JSON(http.StatusOK, ret)

	arg, ok := util.JsonArg(c, ret)
	if !ok {
		return
	}

	notebook := arg["notebook"].(string)
	if util.InvalidIDPattern(notebook, ret) {
		return
	}

	box := model.Conf.GetBox(notebook)
	if nil == box {
		ret.Code = -1
		ret.Msg = "notebook [" + notebook + "] not found"
		return
	}

	param, err := gulu.JSON.MarshalJSON(arg["conf"])
	if err != nil {
		ret.Code = -1
		ret.Msg = err.Error()
		return
	}

	boxConf := box.GetConf()

	savedBoxCrypt := model.DeepCopyBoxEncryption(boxConf.BoxCrypt)
	savedEncrypted := boxConf.Encrypted
	if err = gulu.JSON.UnmarshalJSON(param, boxConf); err != nil {
		ret.Code = -1
		ret.Msg = err.Error()
		return
	}
	boxConf.Encrypted = savedEncrypted
	boxConf.BoxCrypt = savedBoxCrypt

	boxConf.DocCreateSavePath = util.TrimSpaceInPath(boxConf.DocCreateSavePath)

	boxConf.RefCreateSavePath = util.TrimSpaceInPath(boxConf.RefCreateSavePath)

	boxConf.DailyNoteSavePath = util.TrimSpaceInPath(boxConf.DailyNoteSavePath)
	if "" != boxConf.DailyNoteSavePath {
		if !strings.HasPrefix(boxConf.DailyNoteSavePath, "/") {
			boxConf.DailyNoteSavePath = "/" + boxConf.DailyNoteSavePath
		}
	}
	if "/" == boxConf.DailyNoteSavePath {
		ret.Code = -1
		ret.Msg = model.Conf.Language(49)
		return
	}

	boxConf.DailyNoteTemplatePath = util.TrimSpaceInPath(boxConf.DailyNoteTemplatePath)
	if "" != boxConf.DailyNoteTemplatePath {
		if !strings.HasSuffix(boxConf.DailyNoteTemplatePath, ".md") {
			boxConf.DailyNoteTemplatePath += ".md"
		}
		if !strings.HasPrefix(boxConf.DailyNoteTemplatePath, "/") {
			boxConf.DailyNoteTemplatePath = "/" + boxConf.DailyNoteTemplatePath
		}
	}

	if err := box.SaveConf(boxConf); err != nil {
		ret.Code = -1
		ret.Msg = err.Error()
		return
	}
	ret.Data = boxConf
}

func lsNotebooks(c *gin.Context) {
	ret := gulu.Ret.NewResult()
	defer c.JSON(http.StatusOK, ret)

	flashcard := false

	arg := map[string]any{}
	if err := c.ShouldBindJSON(&arg); err == nil {
		if arg["flashcard"] != nil {
			flashcard = arg["flashcard"].(bool)
		}
	}

	var notebooks []*model.Box
	var publishAccess model.PublishAccess
	isReadOnlyRole := model.IsReadOnlyRoleContext(c)
	if flashcard {
		notebooks = model.GetFlashcardNotebooks()
	} else {
		var err error
		notebooks, err = model.ListNotebooks()
		if err != nil {
			return
		}
		if isReadOnlyRole {
			publishAccess = model.GetPublishAccess()
			tempNotebooks := []*model.Box{}
			for _, notebook := range notebooks {

				if notebook.Closed {
					continue
				}

				invisible := false
				for _, item := range publishAccess {
					if item.ID == notebook.ID {
						if !item.Visible {
							invisible = true
						}
						break
					}
				}
				if invisible {
					continue
				}
				tempNotebooks = append(tempNotebooks, notebook)
			}
			notebooks = tempNotebooks
		}
	}

	boxDocEnabled := model.IsBoxDocEnabled()
	if !flashcard && boxDocEnabled {
		for _, notebook := range notebooks {
			if !notebook.Closed {
				if isReadOnlyRole {
					notebook.SubFileCount = model.BoxDocSubFileCountForPublish(notebook.ID, publishAccess)
				} else {
					notebook.SubFileCount = model.BoxDocSubFileCount(notebook.ID)
				}
			}
		}
	}

	ret.Data = map[string]any{
		"notebooks":     notebooks,
		"boxDocEnabled": boxDocEnabled,
	}
}

func enableEncryptedNotebooks(c *gin.Context) {
	ret := gulu.Ret.NewResult()
	defer c.JSON(http.StatusOK, ret)

	arg, ok := util.JsonArg(c, ret)
	if !ok {
		return
	}

	var password string
	if !util.ParseJsonArgs(arg, ret, util.BindJsonArg("password", &password, true, true)) {
		return
	}

	if err := model.EnableEncryptedNotebook(password); err != nil {
		ret.Code = -1
		ret.Msg = err.Error()
		return
	}
}

func disableEncryptedNotebooks(c *gin.Context) {
	ret := gulu.Ret.NewResult()
	defer c.JSON(http.StatusOK, ret)

	if err := model.DisableEncryptedNotebook(); err != nil {
		ret.Code = -1
		ret.Msg = err.Error()
		return
	}
}

func createEncryptedNotebook(c *gin.Context) {
	ret := gulu.Ret.NewResult()
	defer c.JSON(http.StatusOK, ret)

	arg, ok := util.JsonArg(c, ret)
	if !ok {
		return
	}

	var name, password string
	if !util.ParseJsonArgs(arg, ret,
		util.BindJsonArg("name", &name, true, false),
		util.BindJsonArg("password", &password, true, true),
	) {
		return
	}

	id, err := model.CreateEncryptedBox(name, password)
	if err != nil {
		ret.Code = -1
		ret.Msg = err.Error()
		return
	}

	existed, err := model.Mount(id)
	if err != nil {
		model.LockBox(id)
		ret.Code = -1
		ret.Msg = err.Error()
		return
	}

	box := model.Conf.GetBox(id)
	evt := util.NewCmdResult("mount", 0, util.PushModeBroadcast)
	evt.Data = map[string]any{
		"box":     box,
		"existed": existed,
	}
	util.PushEvent(evt)

	ret.Data = map[string]any{
		"notebook": box,
	}
}

func unlockNotebook(c *gin.Context) {
	ret := gulu.Ret.NewResult()
	defer c.JSON(http.StatusOK, ret)

	arg, ok := util.JsonArg(c, ret)
	if !ok {
		return
	}

	var notebook, password string
	if !util.ParseJsonArgs(arg, ret,
		util.BindJsonArg("notebook", &notebook, true, true),
		util.BindJsonArg("password", &password, true, true),
	) {
		return
	}

	if util.InvalidIDPattern(notebook, ret) {
		return
	}

	boxCrypt, err := model.GetBoxEncryption(notebook)
	if err != nil {
		ret.Code = -1
		ret.Msg = model.Conf.Language(318)
		return
	}
	if boxCrypt == nil || len(boxCrypt.WrappedDEK) == 0 {
		ret.Code = -1
		ret.Msg = model.Conf.Language(319)
		return
	}

	if err := model.UnlockBox(notebook, password, boxCrypt); err != nil {
		ret.Code = -1
		ret.Msg = err.Error()
		return
	}
}

func unlockAndOpenNotebook(c *gin.Context) {
	ret := gulu.Ret.NewResult()
	defer c.JSON(http.StatusOK, ret)

	arg, ok := util.JsonArg(c, ret)
	if !ok {
		return
	}

	var notebook, password string
	if !util.ParseJsonArgs(arg, ret,
		util.BindJsonArg("notebook", &notebook, true, true),
		util.BindJsonArg("password", &password, true, true),
	) {
		return
	}

	if util.InvalidIDPattern(notebook, ret) {
		return
	}

	boxCrypt, err := model.GetBoxEncryption(notebook)
	if err != nil {
		ret.Code = -1
		ret.Msg = model.Conf.Language(318)
		return
	}
	if boxCrypt == nil || len(boxCrypt.WrappedDEK) == 0 {
		ret.Code = -1
		ret.Msg = model.Conf.Language(319)
		return
	}

	if err := model.UnlockBox(notebook, password, boxCrypt); err != nil {
		ret.Code = -1
		ret.Msg = err.Error()
		return
	}

	msgId := util.PushMsg(model.Conf.Language(45), 1000*60*15)
	defer util.PushClearMsg(msgId)
	existed, err := model.Mount(notebook)
	if err != nil {
		model.LockBox(notebook)
		ret.Code = -1
		ret.Msg = err.Error()
		return
	}

	box := model.Conf.Box(notebook)
	if nil == box {
		model.LockBox(notebook)
		ret.Code = -1
		ret.Msg = "opened notebook [" + notebook + "] not found"
		return
	}

	evt := util.NewCmdResult("mount", 0, util.PushModeBroadcast)
	evt.Data = map[string]any{
		"box":     box,
		"existed": existed,
	}
	util.PushEvent(evt)
}

func lockNotebook(c *gin.Context) {
	ret := gulu.Ret.NewResult()
	defer c.JSON(http.StatusOK, ret)

	arg, ok := util.JsonArg(c, ret)
	if !ok {
		return
	}

	var notebook string
	if !util.ParseJsonArgs(arg, ret, util.BindJsonArg("notebook", &notebook, true, true)) {
		return
	}

	if util.InvalidIDPattern(notebook, ret) {
		return
	}

	if !model.IsEncryptedBox(notebook) {
		ret.Code = -1
		ret.Msg = model.Conf.Language(319)
		return
	}

	model.Unmount(notebook)
}

func setNotebookCryptoAutoLock(c *gin.Context) {
	ret := gulu.Ret.NewResult()
	defer c.JSON(http.StatusOK, ret)

	arg, ok := util.JsonArg(c, ret)
	if !ok {
		return
	}

	var autoLockMinutes float64
	if !util.ParseJsonArgs(arg, ret, util.BindJsonArg("autoLockMinutes", &autoLockMinutes, true, false)) {
		return
	}

	minutes := max(int(autoLockMinutes), 0)

	model.SetAutoLockMinutes(minutes)
	model.Conf.Save()
}

func touchEncryptedNotebooks(c *gin.Context) {
	model.TouchUnlockedEncryptedBoxes()
	c.JSON(http.StatusOK, gulu.Ret.NewResult())
}

func changeMasterPassword(c *gin.Context) {
	ret := gulu.Ret.NewResult()
	defer c.JSON(http.StatusOK, ret)

	arg, ok := util.JsonArg(c, ret)
	if !ok {
		return
	}

	var oldPassword, newPassword string
	if !util.ParseJsonArgs(arg, ret,
		util.BindJsonArg("oldPassword", &oldPassword, true, true),
		util.BindJsonArg("newPassword", &newPassword, true, true),
	) {
		return
	}

	if err := model.ChangeMasterPassword(oldPassword, newPassword); err != nil {
		ret.Code = -1
		ret.Msg = err.Error()
		return
	}
}

func getEncryptedNotebookStatus(c *gin.Context) {
	ret := gulu.Ret.NewResult()
	defer c.JSON(http.StatusOK, ret)

	model.NotebookCryptoMuLock()
	boxIDs := model.ListAllEncryptedBoxIDs()
	enabled := model.NotebookCryptoEnabled()
	model.NotebookCryptoMuUnlock()
	pendingMigration, migrationBoxes := model.MasterPasswordMigrationStatus()

	hasHistoryDependency := model.HasEncryptedNotebookHistory()

	boxes := make([]map[string]any, 0, len(boxIDs))
	for _, id := range boxIDs {
		box := model.Conf.Box(id)
		name := ""
		if box != nil {
			name = box.Name
		}
		boxes = append(boxes, map[string]any{
			"id":       id,
			"name":     name,
			"unlocked": model.IsBoxUnlocked(id),
		})
	}

	ret.Data = map[string]any{
		"enabled":              enabled,
		"count":                len(boxIDs),
		"boxes":                boxes,
		"migrationPending":     pendingMigration,
		"migrationBoxes":       migrationBoxes,
		"hasHistoryDependency": hasHistoryDependency,
	}
}

func exportNotebookCryptoBackup(c *gin.Context) {
	ret := gulu.Ret.NewResult()
	defer c.JSON(http.StatusOK, ret)

	downloadPath, err := model.ExportNotebookCryptoBackup()
	if err != nil {
		ret.Code = -1
		ret.Msg = err.Error()
		return
	}
	ret.Data = map[string]any{
		"file": downloadPath,
	}
}

func importNotebookCryptoBackup(c *gin.Context) {
	ret := gulu.Ret.NewResult()
	defer c.JSON(http.StatusOK, ret)

	form, err := c.MultipartForm()
	if err != nil {
		ret.Code = -1
		ret.Msg = err.Error()
		return
	}
	if 1 > len(form.File["file"]) {
		ret.Code = -1
		ret.Msg = "file not found"
		return
	}
	fh := form.File["file"][0]
	f, err := fh.Open()
	if err != nil {
		ret.Code = -1
		ret.Msg = err.Error()
		return
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		ret.Code = -1
		ret.Msg = err.Error()
		return
	}
	password := ""
	if vals := form.Value["password"]; len(vals) > 0 {
		password = vals[0]
	}
	if err := model.ImportNotebookCryptoBackup(data, password); err != nil {
		ret.Code = -1
		ret.Msg = err.Error()
		return
	}
}
