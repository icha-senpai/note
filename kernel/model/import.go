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

package model

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"io/fs"
	"maps"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"sort"
	"strings"

	"github.com/icha-senpai/note/kernel/av"
	"github.com/icha-senpai/note/kernel/cache"
	"github.com/icha-senpai/note/kernel/conf"
	"github.com/icha-senpai/note/kernel/filesys"
	"github.com/icha-senpai/note/kernel/sql"
	"github.com/icha-senpai/note/kernel/task"
	"github.com/icha-senpai/note/kernel/treenode"
	"github.com/icha-senpai/note/kernel/util"
	"github.com/icha-senpai/note/third_party/forks/dataparser"
	"github.com/icha-senpai/note/third_party/forks/filelock"
	"github.com/icha-senpai/note/third_party/forks/gulu"
	"github.com/icha-senpai/note/third_party/forks/logging"
	"github.com/icha-senpai/note/third_party/forks/lute"
	"github.com/icha-senpai/note/third_party/forks/lute/ast"
	"github.com/icha-senpai/note/third_party/forks/lute/html"
	"github.com/icha-senpai/note/third_party/forks/lute/html/atom"
	"github.com/icha-senpai/note/third_party/forks/lute/parse"
	"github.com/icha-senpai/note/third_party/forks/lute/render"
	util2 "github.com/icha-senpai/note/third_party/forks/lute/util"
	"github.com/icha-senpai/note/third_party/forks/riff"
	shellquote "github.com/kballard/go-shellquote"
)

func HTML2Tree(htmlStr string, luteEngine *lute.Lute, boxID string) (tree *parse.Tree, withMath bool) {
	htmlStr = gulu.Str.RemovePUA(htmlStr)
	assetDirPath := filepath.Join(util.DataDir, "assets")
	if boxID != "" {
		assetDirPath = filepath.Join(util.DataDir, boxID, "assets")
		_ = os.MkdirAll(assetDirPath, 0755)
	}
	tree = luteEngine.HTML2Tree(htmlStr)
	ast.Walk(tree.Root, func(n *ast.Node, entering bool) ast.WalkStatus {
		if !entering {
			return ast.WalkContinue
		}

		switch n.Type {
		case ast.NodeHTMLBlock:
			if bytes.HasPrefix(n.Tokens, []byte("<pre ")) && bytes.HasSuffix(n.Tokens, []byte("</pre>")) {
				if bytes.Contains(n.Tokens, []byte("data:image/svg+xml;base64")) {
					matches := regexp.MustCompile(`(?sU)<pre [^>]*>(.*)</pre>`).FindSubmatch(n.Tokens)
					if len(matches) >= 2 {
						n.Tokens = matches[1]
					}
					subTree := parse.Inline("", n.Tokens, luteEngine.ParseOptions)
					if nil != subTree && nil != subTree.Root && nil != subTree.Root.FirstChild {
						n.Type = ast.NodeParagraph
						var children []*ast.Node
						for c := subTree.Root.FirstChild.FirstChild; nil != c; c = c.Next {
							children = append(children, c)
						}
						for _, c := range children {
							n.AppendChild(c)
						}
					}
				} else if bytes.Contains(n.Tokens, []byte("<svg")) {
					processHTMLBlockSvgImg(n, assetDirPath, boxID)
				}
			}
		case ast.NodeText:
			if n.ParentIs(ast.NodeTableCell) {
				n.Tokens = bytes.ReplaceAll(n.Tokens, []byte("\\|"), []byte("|"))
				n.Tokens = bytes.ReplaceAll(n.Tokens, []byte("|"), []byte("\\|"))
				n.Tokens = bytes.ReplaceAll(n.Tokens, []byte("\\<br /\\>"), []byte("<br />"))
			}
		case ast.NodeInlineMath:
			withMath = true
		case ast.NodeLinkDest:
			dest := n.TokensStr()
			if strings.HasPrefix(dest, "data:image") && strings.Contains(dest, ";base64,") {
				processBase64Img(n, dest, assetDirPath, boxID)
			}
		}
		return ast.WalkContinue
	})
	return
}

func ImportSY(zipPath, boxID, toPath string) (err error) {
	_, err = importSY(zipPath, boxID, toPath, false, false)
	return
}

func ImportSYNotebook(zipPath string) (boxID string, err error) {
	return importSY(zipPath, "", "/", true, false)
}

var ErrSYTargetNotebookRequired = errors.New("target notebook required")

func ImportSYAuto(zipPath, boxID, toPath string) (createdBoxID string, notebook bool, err error) {
	createdBoxID, err = importSY(zipPath, boxID, toPath, false, true)
	notebook = err == nil && createdBoxID != boxID
	return
}

func isSYNotebookExport(hasBoxConf, hasBoxDocMeta bool) bool {
	return hasBoxConf || hasBoxDocMeta
}

func importSY(zipPath, boxID, toPath string, createNotebook, autoDetect bool) (createdBoxID string, err error) {
	util.PushEndlessProgress(Conf.Language(73))
	defer util.ClearPushProgress(100)

	lockSync()
	defer unlockSync()

	baseName := filepath.Base(zipPath)
	ext := filepath.Ext(baseName)
	baseName = strings.TrimSuffix(baseName, ext)
	unzipPath := filepath.Join(filepath.Dir(zipPath), baseName+"-"+gulu.Rand.String(7))
	err = gulu.Zip.Unzip(zipPath, unzipPath)
	if err != nil {
		return
	}
	defer os.RemoveAll(unzipPath)

	var syPaths []string
	filelock.Walk(unzipPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d == nil {
			return nil
		}
		if !d.IsDir() && strings.HasSuffix(d.Name(), ".sy") {
			syPaths = append(syPaths, path)
		}
		return nil
	})

	entries, err := os.ReadDir(unzipPath)
	if err != nil {
		logging.LogErrorf("read unzip dir [%s] failed: %s", unzipPath, err)
		return
	}
	if 1 != len(entries) || !entries[0].IsDir() || len(syPaths) < 1 {
		logging.LogErrorf("invalid .sy.zip [%v]", entries)
		err = errors.New(Conf.Language(199))
		return
	}
	unzipRootPath := filepath.Join(unzipPath, entries[0].Name())
	name := filepath.Base(unzipRootPath)
	if strings.HasPrefix(name, "data-20") && len("data-20230321175442") == len(name) {
		logging.LogErrorf("invalid .sy.zip [unzipRootPath=%s, baseName=%s]", unzipRootPath, name)
		err = errors.New(Conf.Language(199))
		return
	}
	var importedBoxConf *conf.BoxConf
	importedConfPath := filepath.Join(unzipRootPath, ".scribli", "conf.json")
	hasImportedBoxConf := filelock.IsExist(importedConfPath)
	var importedMetadataErr error
	if hasImportedBoxConf {
		confData, readErr := filelock.ReadFile(importedConfPath)
		if readErr == nil {
			importedBoxConf = conf.NewBoxConf()
			if unmarshalErr := gulu.JSON.UnmarshalJSON(confData, importedBoxConf); unmarshalErr != nil {
				logging.LogWarnf("parse imported notebook conf failed: %s", unmarshalErr)
				importedBoxConf = nil
				importedMetadataErr = unmarshalErr
			}
		} else {
			logging.LogWarnf("read imported notebook conf failed: %s", readErr)
			importedMetadataErr = readErr
		}
		if removeErr := filelock.Remove(importedConfPath); removeErr != nil {
			err = removeErr
			return
		}
	}
	var importedBoxDocID string
	importedBoxDocPath := filepath.Join(unzipRootPath, ".scribli", boxDocMetaName)
	hasImportedBoxDocMeta := filelock.IsExist(importedBoxDocPath)
	if hasImportedBoxDocMeta {
		metaData, readErr := filelock.ReadFile(importedBoxDocPath)
		if readErr == nil {
			meta := &boxDocMeta{}
			if unmarshalErr := gulu.JSON.UnmarshalJSON(metaData, meta); unmarshalErr != nil {
				logging.LogWarnf("parse imported notebook document metadata failed: %s", unmarshalErr)
				importedMetadataErr = unmarshalErr
			} else if meta.Spec != boxDocMetaSpec || !ast.IsNodeIDPattern(meta.BoxDocID) {
				logging.LogWarnf("invalid imported notebook document metadata [spec=%d, id=%s]", meta.Spec, meta.BoxDocID)
				importedMetadataErr = errors.New("invalid imported notebook document metadata")
			} else {
				importedBoxDocID = meta.BoxDocID
			}
		} else {
			logging.LogWarnf("read imported notebook document metadata failed: %s", readErr)
			importedMetadataErr = readErr
		}
		if removeErr := filelock.Remove(importedBoxDocPath); removeErr != nil {
			err = removeErr
			return
		}
	}
	if autoDetect {
		if importedMetadataErr != nil {
			err = errors.New(Conf.Language(199))
			return
		}
		createNotebook = isSYNotebookExport(hasImportedBoxConf, hasImportedBoxDocMeta)
	}
	if autoDetect && !createNotebook && boxID == "" {
		err = ErrSYTargetNotebookRequired
		return
	}
	if !createNotebook && nil == Conf.Box(boxID) {
		err = errors.New(Conf.Language(0))
		return
	}
	if createNotebook {
		if importedBoxConf != nil && importedBoxConf.Name != "" {
			name = importedBoxConf.Name
		}
		boxID, err = CreateBox(util.RemoveInvalid(name))
		if err != nil {
			return "", err
		}
		createdBoxID = boxID
		defer func() {
			if err == nil {
				return
			}
			treenode.RemoveBlockTreesByBoxID(boxID)
			sql.DeleteBoxQueue(boxID)
			if removeErr := filelock.Remove(filepath.Join(util.DataDir, boxID)); removeErr != nil {
				logging.LogErrorf("remove notebook [%s] after import failed: %s", boxID, removeErr)
			}
		}()
		if importedBoxConf != nil {
			box := &Box{ID: boxID}
			boxConf := box.GetConf()
			boxConf.Icon = importedBoxConf.Icon
			if strings.Contains(boxConf.Icon, ".") {
				boxConf.Icon = util.FilterUploadEmojiFileName(boxConf.Icon)
			}
			boxConf.RefCreateSavePath = importedBoxConf.RefCreateSavePath
			boxConf.DocCreateSavePath = importedBoxConf.DocCreateSavePath
			boxConf.DailyNoteSavePath = importedBoxConf.DailyNoteSavePath
			boxConf.DailyNoteTemplatePath = importedBoxConf.DailyNoteTemplatePath
			boxConf.SortMode = importedBoxConf.SortMode
			if err = box.SaveConf(boxConf); err != nil {
				return createdBoxID, err
			}
		}
	} else {
		createdBoxID = boxID
	}
	toPath = normalizeBoxDocTarget(boxID, toPath)

	luteEngine := util.NewLute()
	blockIDs := map[string]string{}
	trees := map[string]*parse.Tree{}
	importedBoxDoc := false

	for i, syPath := range syPaths {
		data, readErr := os.ReadFile(syPath)
		if nil != readErr {
			logging.LogErrorf("read .sy [%s] failed: %s", syPath, readErr)
			err = readErr
			return
		}
		tree, _, parseErr := dataparser.ParseJSON(data, luteEngine.ParseOptions)
		if nil != parseErr {
			logging.LogErrorf("parse .sy [%s] failed: %s", syPath, parseErr)
			err = parseErr
			return
		}
		oldRootID := tree.Root.ID
		ast.Walk(tree.Root, func(n *ast.Node, entering bool) ast.WalkStatus {
			if !entering || "" == n.ID {
				return ast.WalkContinue
			}

			// Keep original creation time when importing .sy.zip
			newNodeID := util.TimeFromID(n.ID) + "-" + util.RandString(7)
			if createNotebook && oldRootID == importedBoxDocID && n.ID == importedBoxDocID {
				newNodeID = boxID
			}
			blockIDs[n.ID] = newNodeID
			n.ID = newNodeID
			n.SetIALAttr("id", newNodeID)

			if icon := n.IALAttr("icon"); "" != icon {
				// XSS through emoji name
				icon = util.FilterUploadEmojiFileName(icon)
				n.SetIALAttr("icon", icon)
			}

			return ast.WalkContinue
		})
		tree.ID = tree.Root.ID
		tree.Path = filepath.ToSlash(strings.TrimPrefix(syPath, unzipRootPath))
		if createNotebook && oldRootID == importedBoxDocID {
			importedBoxDoc = true
			tree.Root.SetIALAttr(DocHiddenAttr, "true")
		} else if oldRootID == importedBoxDocID {
			removeBoxDocHiddenAttr(tree)
		}
		trees[tree.ID] = tree
		util.PushEndlessProgress(Conf.language(73) + " " + fmt.Sprintf(Conf.language(70), fmt.Sprintf("%d/%d", i+1, len(syPaths))))
	}
	if importedBoxDoc {
		if err = writeBoxDocID(boxID); err != nil {
			return
		}
	}

	for _, tree := range trees {
		util.PushEndlessProgress(Conf.language(73) + " " + fmt.Sprintf(Conf.language(70), tree.Root.IALAttr("title")))
		ast.Walk(tree.Root, func(n *ast.Node, entering bool) ast.WalkStatus {
			if !entering {
				return ast.WalkContinue
			}

			if treenode.IsBlockRef(n) {
				defID, _, _ := treenode.GetBlockRef(n)
				newDefID := blockIDs[defID]
				if "" != newDefID {
					n.TextMarkBlockRefID = newDefID
				}
			} else if ast.NodeTextMark == n.Type && n.IsTextMarkType("a") {
				// Block hyperlinks do not point to regenerated block IDs when importing .sy.zip
				defID, ok := cutBlockProtocolURL(n.TextMarkAHref)
				if !ok {
					return ast.WalkContinue
				}
				newDefID := blockIDs[defID]
				if "" != newDefID {
					n.TextMarkAHref = makeBlockProtocolURL(newDefID)
				}
			} else if ast.NodeBlockQueryEmbedScript == n.Type {
				for oldID, newID := range blockIDs {

					n.Tokens = bytes.ReplaceAll(n.Tokens, []byte(oldID), []byte(newID))
				}
			}
			return ast.WalkContinue
		})
	}

	var replacements []string
	for oldID, newID := range blockIDs {
		replacements = append(replacements, oldID, newID)
	}
	blockIDReplacer := strings.NewReplacer(replacements...)

	storage := filepath.Join(unzipRootPath, "storage")
	storageAvDir := filepath.Join(storage, "av")
	avIDs := map[string]string{}
	renameAvPaths := map[string]string{}
	if gulu.File.IsExist(storageAvDir) {

		filelock.Walk(storageAvDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d == nil {
				return nil
			}

			if ".json" == d.Name() { //
				if removeErr := os.RemoveAll(path); nil != removeErr {
					logging.LogErrorf("remove empty av file [%s] failed: %s", path, removeErr)
				}
				return nil
			}

			if !strings.HasSuffix(path, ".json") || !ast.IsNodeIDPattern(strings.TrimSuffix(d.Name(), ".json")) {
				return nil
			}

			newAvID := ast.NewNodeID()
			oldAvID := strings.TrimSuffix(d.Name(), ".json")
			newPath := filepath.Join(filepath.Dir(path), newAvID+".json")
			renameAvPaths[path] = newPath
			avIDs[oldAvID] = newAvID
			return nil
		})

		for oldPath, newPath := range renameAvPaths {
			data, readErr := os.ReadFile(oldPath)
			if nil != readErr {
				logging.LogErrorf("read av file [%s] failed: %s", oldPath, readErr)
				err = readErr
				return
			}

			newData := data
			for oldAvID, newAvID := range avIDs {
				newData = bytes.ReplaceAll(newData, []byte(oldAvID), []byte(newAvID))
			}
			newData = []byte(blockIDReplacer.Replace(string(newData)))
			if !bytes.Equal(data, newData) {
				if writeErr := os.WriteFile(oldPath, newData, 0644); nil != writeErr {
					logging.LogErrorf("write av file [%s] failed: %s", oldPath, writeErr)
					err = writeErr
					return
				}
			}

			if err = os.Rename(oldPath, newPath); err != nil {
				logging.LogErrorf("rename av file from [%s] to [%s] failed: %s", oldPath, newPath, err)
				return
			}
		}

		if !IsEncryptedBox(boxID) {
			targetStorageAvDir := filepath.Join(util.DataDir, "storage", "av")
			if copyErr := filelock.Copy(storageAvDir, targetStorageAvDir); nil != copyErr {
				logging.LogErrorf("copy storage av dir from [%s] to [%s] failed: %s", storageAvDir, targetStorageAvDir, copyErr)
			}
		} else {

			if err = encryptBoxAVFiles(boxID, storageAvDir); err != nil {
				return
			}
		}

		for _, tree := range trees {
			util.PushEndlessProgress(Conf.language(73) + " " + fmt.Sprintf(Conf.language(70), tree.Root.IALAttr("title")))
			ast.Walk(tree.Root, func(n *ast.Node, entering bool) ast.WalkStatus {
				if !entering || "" == n.ID {
					return ast.WalkContinue
				}

				ial := parse.IAL2Map(n.KramdownIAL)
				for k, v := range ial {
					if strings.HasPrefix(k, av.NodeAttrNameAvs) {
						newKey, newVal := k, v
						for oldAvID, newAvID := range avIDs {
							newKey = strings.ReplaceAll(newKey, oldAvID, newAvID)
							newVal = strings.ReplaceAll(newVal, oldAvID, newAvID)
						}
						n.RemoveIALAttr(k)
						n.SetIALAttr(newKey, newVal)
					}
				}

				if ast.NodeAttributeView == n.Type {
					n.AttributeViewID = avIDs[n.AttributeViewID]
				}
				return ast.WalkContinue
			})

		}

		var attrViewIDs []string
		for _, avID := range avIDs {
			attrViewIDs = append(attrViewIDs, avID)
		}
		updateBoundBlockAvsAttribute(attrViewIDs)

		relationAvs := map[string]string{}
		for _, avID := range avIDs {
			attrView, _ := av.ParseAttributeView(avID)
			if nil == attrView {
				continue
			}

			for _, keyValues := range attrView.KeyValues {
				if nil != keyValues.Key && av.KeyTypeRelation == keyValues.Key.Type && nil != keyValues.Key.Relation {
					relationAvs[avID] = keyValues.Key.Relation.AvID
				}
			}
		}

		for srcAvID, destAvID := range relationAvs {
			av.UpsertAvBackRel(srcAvID, destAvID)
		}
	}

	storageRiffDir := filepath.Join(storage, "riff")
	if gulu.File.IsExist(storageRiffDir) {
		deckToImport, loadErr := riff.LoadDeck(storageRiffDir, builtinDeckID, Conf.Flashcard.RequestRetention, Conf.Flashcard.MaximumInterval, Conf.Flashcard.Weights)
		if nil != loadErr {
			logging.LogErrorf("load deck [%s] failed: %s", name, loadErr)
		} else {
			deck := Decks[builtinDeckID]
			if nil == deck {
				var createErr error
				deck, createErr = createDeck0("Built-in Deck", builtinDeckID)
				if nil == createErr {
					Decks[deck.ID] = deck
				}
			}

			bIDs := deckToImport.GetBlockIDs()
			cards := deckToImport.GetCardsByBlockIDs(bIDs)
			for _, card := range cards {
				deck.AddCard(ast.NewNodeID(), blockIDs[card.BlockID()])
			}

			if 0 < len(cards) {
				if saveErr := deck.Save(); nil != saveErr {
					logging.LogErrorf("save deck [%s] failed: %s", name, saveErr)
				}
			}
		}
	}

	if removeErr := os.RemoveAll(storage); nil != removeErr {
		logging.LogErrorf("remove temp storage av dir [%s] failed: %s", storage, removeErr)
	}

	if 1 > len(avIDs) {
		for _, tree := range trees {
			ast.Walk(tree.Root, func(n *ast.Node, entering bool) ast.WalkStatus {
				if !entering || !n.IsBlock() {
					return ast.WalkContinue
				}

				n.RemoveIALAttr(av.NodeAttrNameAvs)
				return ast.WalkContinue
			})
		}
	}

	for _, tree := range trees {
		util.PushEndlessProgress(Conf.language(73) + " " + fmt.Sprintf(Conf.language(70), tree.Root.IALAttr("title")))
		syPath := filepath.Join(unzipRootPath, tree.Path)

		finalSyName := tree.ID + ".sy"
		finalRelPath := filepath.ToSlash(filepath.Join(filepath.Dir(tree.Path), finalSyName))
		treenode.UpgradeSpec(tree)
		renderer := render.NewJSONRenderer(tree, luteEngine.RenderOptions, luteEngine.ParseOptions)
		data := renderer.Render()

		if !util.UseSingleLineSave {
			buf := bytes.Buffer{}
			buf.Grow(1024 * 1024 * 2)
			if err = json.Indent(&buf, data, "", "\t"); err != nil {
				return
			}
			data = buf.Bytes()
		}

		newSyPath := filepath.Join(filepath.Dir(syPath), finalSyName)
		if err = writeImportedTree(boxID, syPath, newSyPath, finalRelPath, data); err != nil {
			logging.LogErrorf("write imported .sy [%s] failed: %s", syPath, err)
			return
		}
		tree.Path = finalRelPath
	}

	fullSortIDs := map[string]int{}
	sortIDs := map[string]int{}
	var sortData []byte
	var sortErr error
	sortPath := filepath.Join(unzipRootPath, ".scribli", "sort.json")
	if filelock.IsExist(sortPath) {
		sortData, sortErr = filelock.ReadFile(sortPath)
		if nil != sortErr {
			logging.LogErrorf("read import sort conf failed: %s", sortErr)
		}

		if sortErr = gulu.JSON.UnmarshalJSON(sortData, &sortIDs); nil != sortErr {
			logging.LogErrorf("unmarshal sort conf failed: %s", sortErr)
		}

		boxSortPath := filepath.Join(util.DataDir, boxID, ".scribli", "sort.json")
		if filelock.IsExist(boxSortPath) {
			sortData, sortErr = filelock.ReadFile(boxSortPath)
			if nil != sortErr {
				logging.LogErrorf("read box sort conf failed: %s", sortErr)
			}

			if sortErr = gulu.JSON.UnmarshalJSON(sortData, &fullSortIDs); nil != sortErr {
				logging.LogErrorf("unmarshal box sort conf failed: %s", sortErr)
			}
		}

		for oldID, sort := range sortIDs {
			if newID := blockIDs[oldID]; "" != newID {
				fullSortIDs[newID] = sort
			}
		}

		sortData, sortErr = gulu.JSON.MarshalJSON(fullSortIDs)
		if nil != sortErr {
			logging.LogErrorf("marshal box full sort conf failed: %s", sortErr)
		} else {
			sortErr = filelock.WriteFile(boxSortPath, sortData)
			if nil != sortErr {
				logging.LogErrorf("write box full sort conf failed: %s", sortErr)
			}
		}
		if removeErr := os.RemoveAll(sortPath); nil != removeErr {
			logging.LogErrorf("remove temp sort conf failed: %s", removeErr)
		}
	}

	renamePaths := map[string]string{}
	filelock.Walk(unzipRootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d == nil {
			return nil
		}
		if d.IsDir() && ast.IsNodeIDPattern(d.Name()) {
			renamePaths[path] = path
		}
		return nil
	})
	for p := range renamePaths {
		originalPath := p
		p = strings.TrimPrefix(p, unzipRootPath)
		p = filepath.ToSlash(p)
		parts := strings.Split(p, "/")
		buf := bytes.Buffer{}
		buf.WriteString("/")
		for i, part := range parts {
			if "" == part {
				continue
			}
			newNodeID := blockIDs[part]
			if "" != newNodeID {
				buf.WriteString(newNodeID)
			} else {
				buf.WriteString(part)
			}
			if i < len(parts)-1 {
				buf.WriteString("/")
			}
		}
		newPath := buf.String()
		renamePaths[originalPath] = filepath.Join(unzipRootPath, newPath)
	}

	var oldPaths []string
	for oldPath := range renamePaths {
		oldPaths = append(oldPaths, oldPath)
	}
	sort.Slice(oldPaths, func(i, j int) bool {
		return strings.Count(oldPaths[i], string(os.PathSeparator)) < strings.Count(oldPaths[j], string(os.PathSeparator))
	})
	for i, oldPath := range oldPaths {
		newPath := renamePaths[oldPath]
		if err = filelock.Rename(oldPath, newPath); err != nil {
			logging.LogErrorf("rename path from [%s] to [%s] failed: %s", oldPath, renamePaths[oldPath], err)
			err = errors.New("rename path failed")
			return
		}

		delete(renamePaths, oldPath)
		var toRemoves []string
		newRenamedPaths := map[string]string{}
		for oldP, newP := range renamePaths {
			if strings.HasPrefix(oldP, oldPath) {
				renamedOldP := strings.Replace(oldP, oldPath, newPath, 1)
				newRenamedPaths[renamedOldP] = newP
				toRemoves = append(toRemoves, oldPath)
			}
		}
		for _, toRemove := range toRemoves {
			delete(renamePaths, toRemove)
		}
		maps.Copy(renamePaths, newRenamedPaths)
		for j := i + 1; j < len(oldPaths); j++ {
			if strings.HasPrefix(oldPaths[j], oldPath) {
				renamedOldP := strings.Replace(oldPaths[j], oldPath, newPath, 1)
				oldPaths[j] = renamedOldP
			}
		}
	}

	assetNameMap := map[string]string{}
	var assetsDirs []string
	filelock.Walk(unzipRootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d == nil || unzipRootPath == path {
			return nil
		}
		if d.Name() == "assets" && d.IsDir() {
			assetsDirs = append(assetsDirs, path)
		}
		return nil
	})
	dataAssets := filepath.Join(util.DataDir, "assets")
	if IsEncryptedBox(boxID) {

		boxAssetsDir := filepath.Join(util.DataDir, boxID, "assets")
		if err = os.MkdirAll(boxAssetsDir, 0755); err != nil {
			return
		}
		for _, assets := range assetsDirs {
			if gulu.File.IsDir(assets) {
				filelock.Walk(assets, func(path string, d fs.DirEntry, err error) error {
					if err != nil || d == nil || d.IsDir() {
						return err
					}
					originalName := d.Name()
					ext := filepath.Ext(originalName)
					blockID := ast.NewNodeID()
					diskName := encryptedAssetName(ext, blockID)

					if mapErr := writeAssetNameMapping(boxID, diskName, originalName); mapErr != nil {
						return mapErr
					}
					assetNameMap[originalName] = diskName

					src, readErr := filelock.ReadFile(path)
					if readErr != nil {
						return readErr
					}
					if err = writeAssetFile(filepath.Join(boxAssetsDir, diskName), bytes.NewReader(src), boxID); err != nil {
						return err
					}
					return nil
				})
				if err != nil {
					return
				}
			}
			os.RemoveAll(assets)
		}
	} else {
		for _, assets := range assetsDirs {
			if gulu.File.IsDir(assets) {
				if err = filelock.Copy(assets, dataAssets); err != nil {
					logging.LogErrorf("copy assets from [%s] to [%s] failed: %s", assets, dataAssets, err)
					return
				}
			}
			os.RemoveAll(assets)
		}
	}

	unzipRootEmojisPath := filepath.Join(unzipRootPath, "emojis")
	filelock.Walk(unzipRootEmojisPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d == nil {
			return nil
		}
		if !util.IsValidUploadFileName(d.Name()) {
			emojiFullName := path
			fullPathFilteredName := filepath.Join(filepath.Dir(path), util.FilterUploadEmojiFileName(d.Name()))
			// XSS through emoji name
			logging.LogWarnf("renaming invalid custom emoji file [%s] to [%s]", d.Name(), fullPathFilteredName)
			if removeErr := filelock.Rename(emojiFullName, fullPathFilteredName); nil != removeErr {
				logging.LogErrorf("renaming invalid custom emoji file to [%s] failed: %s", fullPathFilteredName, removeErr)
			}
		}
		return nil
	})
	var emojiDirs []string
	filelock.Walk(unzipRootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d == nil || unzipRootPath == path {
			return nil
		}
		if d.Name() == "emojis" && d.IsDir() {
			emojiDirs = append(emojiDirs, path)
		}
		return nil
	})
	dataEmojis := filepath.Join(util.DataDir, "emojis")
	for _, emojis := range emojiDirs {
		if gulu.File.IsDir(emojis) {
			if err = filelock.Copy(emojis, dataEmojis); err != nil {
				logging.LogErrorf("copy emojis from [%s] to [%s] failed: %s", emojis, dataEmojis, err)
				return
			}
		}
		os.RemoveAll(emojis)
	}

	var baseTargetPath string
	if "/" == toPath {
		baseTargetPath = "/"
	} else {
		block := treenode.GetBlockTreeRootByPath(boxID, toPath)
		if nil == block {
			logging.LogErrorf("not found block by path [%s]", toPath)
			return createdBoxID, nil
		}
		baseTargetPath = strings.TrimSuffix(block.Path, ".sy")
	}

	targetDir := filepath.Join(util.DataDir, boxID, baseTargetPath)
	if err = os.MkdirAll(targetDir, 0755); err != nil {
		return
	}

	var treePaths []string
	filelock.Walk(unzipRootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d == nil {
			return nil
		}
		if d.IsDir() {
			if strings.HasPrefix(d.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}

		if !strings.HasSuffix(d.Name(), ".sy") {
			return nil
		}

		p := strings.TrimPrefix(path, unzipRootPath)
		p = filepath.ToSlash(p)
		treePaths = append(treePaths, p)
		return nil
	})

	if err = filelock.Copy(unzipRootPath, targetDir); err != nil {
		logging.LogErrorf("copy data dir from [%s] to [%s] failed: %s", unzipRootPath, util.DataDir, err)
		err = errors.New("copy data failed")
		return
	}

	boxAbsPath := filepath.Join(util.DataDir, boxID)
	importedAvIDs := map[string]struct{}{}
	for _, importedAvID := range avIDs {
		importedAvIDs[importedAvID] = struct{}{}
	}
	for _, treePath := range treePaths {
		absPath := filepath.Join(targetDir, treePath)
		p := strings.TrimPrefix(absPath, boxAbsPath)
		p = filepath.ToSlash(p)
		cache.RemoveTreeDataInBox(util.GetTreeID(p), boxID)
		cache.RemoveDocIALInBox(p, boxID)
		tree, err := filesys.LoadTree(boxID, p, luteEngine)
		if err != nil {
			logging.LogErrorf("load tree [%s] failed: %s", treePath, err)
			continue
		}

		if IsEncryptedBox(boxID) && 0 < len(assetNameMap) {
			updateImportedAssetRefs(tree, assetNameMap)
			indexWriteTreeIndexQueue(tree)
		}

		treenode.IndexBlockTree(tree)
		cache.PutDocIALInBox(tree.Path, tree.Box, parse.IAL2Map(tree.Root.KramdownIAL))
		var avNodes []*ast.Node
		for _, avNode := range tree.Root.ChildrenByType(ast.NodeAttributeView) {
			if _, ok := importedAvIDs[avNode.AttributeViewID]; ok {
				avNodes = append(avNodes, avNode)
			}
		}
		av.BatchUpsertBlockRel(avNodes)
		sql.IndexTreeQueue(tree)
		util.PushEndlessProgress(Conf.language(73) + " " + fmt.Sprintf(Conf.language(70), tree.Root.IALAttr("title")))
	}

	IncSync()

	task.AppendTask(task.UpdateIDs, util.PushUpdateIDs, blockIDs)
	return
}

func writeImportedTree(boxID, syPath, newSyPath, relPath string, data []byte) error {
	if IsEncryptedBox(boxID) {
		HoldBoxReadLock(boxID)
		defer ReleaseBoxReadLock(boxID)

		dek, err := GetDEKIfUnlocked(boxID)
		if err != nil {
			return errors.New(Conf.Language(314))
		}
		data, err = EncryptFile(boxID, relPath, dek, data)
		if err != nil {
			return err
		}
	}
	if err := os.WriteFile(syPath, data, 0644); err != nil {
		return err
	}
	return filelock.Rename(syPath, newSyPath)
}

func updateImportedAssetRefs(tree *parse.Tree, assetNameMap map[string]string) {
	boxSuffix := ""
	if IsEncryptedBox(tree.Box) {
		boxSuffix = "?box=" + tree.Box
	}
	ast.Walk(tree.Root, func(n *ast.Node, entering bool) ast.WalkStatus {
		if !entering {
			return ast.WalkContinue
		}
		switch n.Type {
		case ast.NodeLink:

			dest := string(n.Tokens)
			if updated := replaceAssetName(dest, assetNameMap, boxSuffix); updated != dest {
				n.Tokens = []byte(updated)
			}
		case ast.NodeImage:

			if dest := n.ChildByType(ast.NodeLinkDest); nil != dest {
				src := string(dest.Tokens)
				if updated := replaceAssetName(src, assetNameMap, boxSuffix); updated != src {
					dest.Tokens = []byte(updated)
				}
			}
		case ast.NodeAudio, ast.NodeVideo:
			src := n.TokensStr()
			if updated := replaceAssetName(src, assetNameMap, boxSuffix); updated != src {
				n.Tokens = []byte(updated)
			}
		case ast.NodeTextMark:

			if "" != n.TextMarkAHref {
				if updated := replaceAssetName(n.TextMarkAHref, assetNameMap, boxSuffix); updated != n.TextMarkAHref {
					n.TextMarkAHref = updated
				}
			}
		}
		return ast.WalkContinue
	})
}

func replaceAssetName(path string, assetNameMap map[string]string, boxSuffix string) string {
	if !strings.Contains(path, "assets/") {
		return path
	}
	for original, diskName := range assetNameMap {

		idx := strings.LastIndex(path, original)
		if idx > 0 && strings.Contains(path[idx:], original) {
			path = path[:idx] + diskName + path[idx+len(original):]

			if boxSuffix != "" && !strings.Contains(path, "?box=") {
				path += boxSuffix
			}
		}
	}
	return path
}

func ImportData(zipPath string) (err error) {
	util.PushEndlessProgress(Conf.Language(73))
	defer util.ClearPushProgress(100)

	lockSync()
	defer unlockSync()

	logging.LogInfof("import data from [%s]", zipPath)
	baseName := filepath.Base(zipPath)
	ext := filepath.Ext(baseName)
	baseName = strings.TrimSuffix(baseName, ext)
	unzipPath := filepath.Join(filepath.Dir(zipPath), baseName)
	err = gulu.Zip.Unzip(zipPath, unzipPath)
	if err != nil {
		return
	}
	defer os.RemoveAll(unzipPath)

	files, err := filepath.Glob(filepath.Join(unzipPath, "*/*.sy"))
	if err != nil {
		logging.LogErrorf("check data.zip failed: %s", err)
		return errors.New("check data.zip failed")
	}
	if 0 < len(files) {
		return errors.New(Conf.Language(198))
	}
	dirs, err := os.ReadDir(unzipPath)
	if err != nil {
		logging.LogErrorf("check data.zip failed: %s", err)
		return errors.New("check data.zip failed")
	}
	if 1 != len(dirs) {
		return errors.New(Conf.Language(198))
	}

	tmpDataPath := filepath.Join(unzipPath, dirs[0].Name())
	tmpDataEmojisPath := filepath.Join(tmpDataPath, "emojis")
	filelock.Walk(tmpDataEmojisPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d == nil {
			return nil
		}
		if !util.IsValidUploadFileName(d.Name()) {
			emojiFullName := path
			fullPathFilteredName := filepath.Join(filepath.Dir(path), util.FilterUploadEmojiFileName(d.Name()))
			// XSS through emoji name
			logging.LogWarnf("renaming invalid custom emoji file [%s] to [%s]", d.Name(), fullPathFilteredName)
			if removeErr := filelock.Rename(emojiFullName, fullPathFilteredName); nil != removeErr {
				logging.LogErrorf("renaming invalid custom emoji file to [%s] failed: %s", fullPathFilteredName, removeErr)
			}
		}
		return nil
	})
	if err = filelock.Copy(tmpDataPath, util.DataDir); err != nil {
		logging.LogErrorf("copy data dir from [%s] to [%s] failed: %s", tmpDataPath, util.DataDir, err)
		err = errors.New("copy data failed")
		return
	}

	restoreNotebookCryptoConfigFromBackup()

	logging.LogInfof("import data from [%s] done", zipPath)
	IncSync()
	FullReindex(false)
	return
}

func ImportFromLocalPath(boxID, localPath string, toPath string) (err error) {
	toPath = normalizeBoxDocTarget(boxID, toPath)
	util.PushEndlessProgress(Conf.Language(73))
	defer func() {
		util.PushClearProgress()

		if e := recover(); nil != e {
			stack := debug.Stack()
			msg := fmt.Sprintf("PANIC RECOVERED: %v\n\t%s\n", e, stack)
			logging.LogErrorf("import from local path failed: %s", msg)
			err = errors.New("import from local path failed, please check kernel log for details")
		}
	}()

	lockSync()
	defer unlockSync()

	FlushTxQueue()

	var baseHPath, baseTargetPath, boxLocalPath string
	if "/" == toPath {
		baseHPath = "/"
		baseTargetPath = "/"
	} else {
		block := treenode.GetBlockTreeRootByPath(boxID, toPath)
		if nil == block {
			logging.LogErrorf("not found block by path [%s]", toPath)
			return nil
		}
		baseHPath = block.HPath
		baseTargetPath = strings.TrimSuffix(block.Path, ".sy")
	}
	boxLocalPath = filepath.Join(util.DataDir, boxID)

	hPathsIDs := map[string]string{}
	idPaths := map[string]string{}
	moveIDs := map[string]string{}
	assetsDone := map[string]string{}
	if gulu.File.IsDir(localPath) {
		targetPaths := map[string]string{}
		count := 0

		filelock.Walk(localPath, func(currentPath string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d == nil {
				return nil
			}
			if strings.HasPrefix(d.Name(), ".") {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}

			if !d.IsDir() && (!strings.HasSuffix(currentPath, ".md") && !strings.HasSuffix(currentPath, ".markdown") ||
				strings.Contains(filepath.ToSlash(currentPath), "/assets/")) {

				existName := assetsDone[currentPath]
				var name string
				if "" == existName {
					baseName := filepath.Base(currentPath)
					baseName = util.FilterUploadFileName(baseName)
					data, readErr := os.ReadFile(currentPath)
					if readErr != nil {
						logging.LogErrorf("read asset [%s] failed: %s", currentPath, readErr)
						return nil
					}
					assetDirForBox := filepath.Join(util.DataDir, "assets")
					if boxID != "" {
						assetDirForBox = filepath.Join(util.DataDir, boxID, "assets")
						_ = os.MkdirAll(assetDirForBox, 0755)
					}
					name, err = storeAssetForBox(boxID, assetDirForBox, baseName, data)
					if err != nil {
						logging.LogErrorf("store asset [%s] for box [%s] failed: %s", currentPath, boxID, err)
						return nil
					}
					assetsDone[currentPath] = name
				}
				return nil
			}

			var tree *parse.Tree
			var ext string
			title := d.Name()
			if !d.IsDir() {
				ext = util.Ext(d.Name())
				title = strings.TrimSuffix(d.Name(), ext)
			}
			id := ast.NewNodeID()

			curRelPath := filepath.ToSlash(strings.TrimPrefix(currentPath, localPath))
			targetPath := path.Join(baseTargetPath, id)
			hPath := path.Join(baseHPath, filepath.Base(localPath), filepath.ToSlash(strings.TrimPrefix(currentPath, localPath)))
			hPath = strings.TrimSuffix(hPath, ext)
			if "" == curRelPath {
				curRelPath = "/"
				hPath = "/" + title
			} else {
				dirPath := targetPaths[path.Dir(curRelPath)]
				targetPath = path.Join(dirPath, id)
			}

			targetPath = strings.ReplaceAll(targetPath, ".sy/", "/")
			targetPath += ".sy"
			if _, ok := targetPaths[curRelPath]; !ok {
				targetPaths[curRelPath] = targetPath
			} else {
				targetPath = targetPaths[curRelPath]
				id = util.GetTreeID(targetPath)
			}

			if d.IsDir() {
				if "assets" == d.Name() {

					return nil
				}

				if subMdFiles := util.GetFilePathsByExts(currentPath, []string{".md", ".markdown"}); 1 > len(subMdFiles) {

					return nil
				}

				if gulu.File.IsExist(currentPath+".md") || gulu.File.IsExist(currentPath+".markdown") {
					targetPaths[curRelPath+".md"] = targetPath
					return nil
				}

				tree = treenode.NewTree(boxID, targetPath, hPath, title)
				importTrees = append(importTrees, tree)
				return nil
			}

			if !strings.HasSuffix(d.Name(), ".md") && !strings.HasSuffix(d.Name(), ".markdown") {
				return nil
			}

			data, readErr := os.ReadFile(currentPath)
			if nil != readErr {
				err = readErr
				return io.EOF
			}

			tree, yfmRootID, yfmTitle, yfmUpdated := parseStdMd(data)
			if nil == tree {
				logging.LogErrorf("parse tree [%s] failed", currentPath)
				return nil
			}

			if "" != yfmRootID {
				moveIDs[id] = yfmRootID
				id = yfmRootID
			}
			if "" != yfmTitle {
				title = yfmTitle
			}
			unescapedTitle, unescapeErr := url.PathUnescape(title)
			if nil == unescapeErr {
				title = unescapedTitle
			}
			hPath = path.Join(path.Dir(hPath), title)
			updated := yfmUpdated
			fname := path.Base(targetPath)
			targetPath = strings.ReplaceAll(targetPath, fname, id+".sy")
			targetPaths[curRelPath] = targetPath

			tree.ID = id
			tree.Root.ID = id
			tree.Root.SetIALAttr("id", tree.Root.ID)
			tree.Root.SetIALAttr("title", title)
			tree.Box = boxID
			targetPath = path.Join(path.Dir(targetPath), tree.Root.ID+".sy")
			tree.Path = targetPath
			targetPaths[curRelPath] = targetPath
			tree.HPath = hPath
			tree.Root.Spec = treenode.CurrentSpec

			docDirLocalPath := filepath.Dir(filepath.Join(boxLocalPath, targetPath))
			assetDirPath := getAssetsDir(boxLocalPath, docDirLocalPath)
			currentDir := filepath.Dir(currentPath)
			ast.Walk(tree.Root, func(n *ast.Node, entering bool) ast.WalkStatus {
				if !entering || (ast.NodeLinkDest != n.Type && !n.IsTextMarkType("a")) {
					return ast.WalkContinue
				}

				var dest string
				if ast.NodeLinkDest == n.Type {
					dest = n.TokensStr()
				} else {
					dest = n.TextMarkAHref
				}

				if strings.HasPrefix(dest, "data:image") && strings.Contains(dest, ";base64,") {
					processBase64Img(n, dest, assetDirPath, boxID)
					return ast.WalkContinue
				}

				decodedDest := string(html.DecodeDestination([]byte(dest)))
				if decodedDest != dest {
					dest = decodedDest
				}
				absolutePath := filepath.Join(currentDir, dest)
				if !gulu.File.IsSubPath(currentDir, absolutePath) {
					return ast.WalkContinue
				}

				if ast.NodeLinkDest == n.Type {
					n.Tokens = []byte(dest)
				} else {
					n.TextMarkAHref = dest
				}
				if !util.IsRelativePath(dest) {
					return ast.WalkContinue
				}
				dest = filepath.ToSlash(dest)
				if "" == dest {
					return ast.WalkContinue
				}

				if !gulu.File.IsExist(absolutePath) {
					return ast.WalkContinue
				}

				if strings.HasSuffix(absolutePath, ".md") || strings.HasSuffix(absolutePath, ".markdown") {
					if !strings.Contains(filepath.ToSlash(absolutePath), "/assets/") {

						// Supports converting relative path hyperlinks into document block references after importing Markdown
						return ast.WalkContinue
					}
				}

				existName := assetsDone[absolutePath]
				var name string
				if "" == existName {
					baseName := filepath.Base(absolutePath)
					if IsEncryptedBox(boxID) {

						ext := filepath.Ext(baseName)
						blockID := ast.NewNodeID()
						name = encryptedAssetName(ext, blockID)

						if mapErr := writeAssetNameMapping(boxID, name, baseName); mapErr != nil {
							logging.LogErrorf("write asset name mapping for [%s] failed: %s", baseName, mapErr)
							return ast.WalkContinue
						}
						assetTargetPath := filepath.Join(assetDirPath, name)
						src, readErr := filelock.ReadFile(absolutePath)
						if readErr != nil {
							logging.LogErrorf("read asset [%s] failed: %s", absolutePath, readErr)
							return ast.WalkContinue
						}
						if err = writeAssetFile(assetTargetPath, bytes.NewReader(src), boxID); err != nil {
							logging.LogErrorf("write encrypted asset [%s] failed: %s", assetTargetPath, err)
							return ast.WalkContinue
						}
					} else {
						name = util.FilterUploadFileName(baseName)
						name = util.AssetName(name, ast.NewNodeID())
						assetTargetPath := filepath.Join(assetDirPath, name)
						if err = filelock.Copy(absolutePath, assetTargetPath); err != nil {
							logging.LogErrorf("copy asset from [%s] to [%s] failed: %s", absolutePath, assetTargetPath, err)
							return ast.WalkContinue
						}
					}
					assetsDone[absolutePath] = name
				} else {
					name = existName
				}
				if ast.NodeLinkDest == n.Type {
					assetURL := "assets/" + name
					if IsEncryptedBox(boxID) {
						assetURL += "?box=" + boxID
					}
					n.Tokens = []byte(assetURL)
				} else {
					assetURL := "assets/" + name
					if IsEncryptedBox(boxID) {
						assetURL += "?box=" + boxID
					}
					n.TextMarkAHref = assetURL
				}
				return ast.WalkContinue
			})

			reassignIDUpdated(tree, id, updated)
			importTrees = append(importTrees, tree)

			hPathsIDs[tree.HPath] = tree.ID
			idPaths[tree.ID] = tree.Path

			count++
			if 0 == count%4 {
				util.PushEndlessProgress(fmt.Sprintf(Conf.language(70), fmt.Sprintf("%s", tree.HPath)))
			}
			return nil
		})
	} else {
		fileName := filepath.Base(localPath)
		if !strings.HasSuffix(fileName, ".md") && !strings.HasSuffix(fileName, ".markdown") {
			return errors.New(Conf.Language(79))
		}

		title := strings.TrimSuffix(fileName, ".markdown")
		title = strings.TrimSuffix(title, ".md")
		targetPath := strings.TrimSuffix(toPath, ".sy")
		id := ast.NewNodeID()
		targetPath = path.Join(targetPath, id+".sy")
		var data []byte
		data, err = os.ReadFile(localPath)
		if err != nil {
			return err
		}
		tree, yfmRootID, yfmTitle, yfmUpdated := parseStdMd(data)
		if nil == tree {
			msg := fmt.Sprintf("parse tree [%s] failed", localPath)
			logging.LogError(msg)
			return errors.New(msg)
		}

		if "" != yfmRootID {
			id = yfmRootID
		}
		if "" != yfmTitle {
			title = yfmTitle
		}
		unescapedTitle, unescapeErr := url.PathUnescape(title)
		if nil == unescapeErr {
			title = unescapedTitle
		}
		updated := yfmUpdated
		fname := path.Base(targetPath)
		targetPath = strings.ReplaceAll(targetPath, fname, id+".sy")

		tree.ID = id
		tree.Root.ID = id
		tree.Root.SetIALAttr("id", tree.Root.ID)
		tree.Root.SetIALAttr("title", title)
		tree.Box = boxID
		tree.Path = targetPath
		tree.HPath = path.Join(baseHPath, title)
		tree.Root.Spec = treenode.CurrentSpec

		localPathParentDir := filepath.Dir(localPath)
		docDirLocalPath := filepath.Dir(filepath.Join(boxLocalPath, targetPath))
		assetDirPath := getAssetsDir(boxLocalPath, docDirLocalPath)
		ast.Walk(tree.Root, func(n *ast.Node, entering bool) ast.WalkStatus {
			if !entering || (ast.NodeLinkDest != n.Type && !n.IsTextMarkType("a")) {
				return ast.WalkContinue
			}

			var dest string
			if ast.NodeLinkDest == n.Type {
				dest = n.TokensStr()
			} else {
				dest = n.TextMarkAHref
			}

			if strings.HasPrefix(dest, "data:image") && strings.Contains(dest, ";base64,") {
				processBase64Img(n, dest, assetDirPath, boxID)
				return ast.WalkContinue
			}

			decodedDest := string(html.DecodeDestination([]byte(dest)))
			if decodedDest != dest {
				dest = decodedDest
			}
			absolutePath := filepath.Join(localPathParentDir, dest)
			if !gulu.File.IsSubPath(localPathParentDir, absolutePath) {
				return ast.WalkContinue
			}

			if ast.NodeLinkDest == n.Type {
				n.Tokens = []byte(dest)
			} else {
				n.TextMarkAHref = dest
			}
			if !util.IsRelativePath(dest) {
				return ast.WalkContinue
			}
			dest = filepath.ToSlash(dest)
			if "" == dest {
				return ast.WalkContinue
			}

			if !gulu.File.IsExist(absolutePath) {
				return ast.WalkContinue
			}

			existName := assetsDone[absolutePath]
			var name string
			if "" == existName {
				baseName := filepath.Base(absolutePath)
				if IsEncryptedBox(boxID) {

					ext := filepath.Ext(baseName)
					blockID := ast.NewNodeID()
					name = encryptedAssetName(ext, blockID)

					if mapErr := writeAssetNameMapping(boxID, name, baseName); mapErr != nil {
						logging.LogErrorf("write asset name mapping for [%s] failed: %s", baseName, mapErr)
						return ast.WalkContinue
					}
					assetTargetPath := filepath.Join(assetDirPath, name)
					src, readErr := filelock.ReadFile(absolutePath)
					if readErr != nil {
						logging.LogErrorf("read asset [%s] failed: %s", absolutePath, readErr)
						return ast.WalkContinue
					}
					if err = writeAssetFile(assetTargetPath, bytes.NewReader(src), boxID); err != nil {
						logging.LogErrorf("write encrypted asset [%s] failed: %s", assetTargetPath, err)
						return ast.WalkContinue
					}
				} else {
					name = util.FilterUploadFileName(baseName)
					name = util.AssetName(name, ast.NewNodeID())
					assetTargetPath := filepath.Join(assetDirPath, name)
					if err = filelock.Copy(absolutePath, assetTargetPath); err != nil {
						logging.LogErrorf("copy asset from [%s] to [%s] failed: %s", absolutePath, assetTargetPath, err)
						return ast.WalkContinue
					}
				}
				assetsDone[absolutePath] = name
			} else {
				name = existName
			}
			if ast.NodeLinkDest == n.Type {
				assetURL := "assets/" + name
				if IsEncryptedBox(boxID) {
					assetURL += "?box=" + boxID
				}
				n.Tokens = []byte(assetURL)
			} else {
				assetURL := "assets/" + name
				if IsEncryptedBox(boxID) {
					assetURL += "?box=" + boxID
				}
				n.TextMarkAHref = assetURL
			}
			return ast.WalkContinue
		})

		reassignIDUpdated(tree, id, updated)

		degradeCrossBoundaryBlockRefs(tree.Root, tree.Box)
		importTrees = append(importTrees, tree)
	}

	if 0 < len(importTrees) {
		for id, newID := range moveIDs {
			for _, importTree := range importTrees {
				importTree.ID = strings.ReplaceAll(importTree.ID, id, newID)
				importTree.Path = strings.ReplaceAll(importTree.Path, id, newID)
			}
		}

		initSearchLinks()
		convertMdHyperlinks2WikiLinks()
		convertWikiLinksAndTags()
		mergeTextAndHandlerNestedInlines()

		box := Conf.Box(boxID)
		for i, tree := range importTrees {
			indexWriteTreeIndexQueue(tree)
			if 0 == i%4 {
				util.PushEndlessProgress(fmt.Sprintf(Conf.Language(66), fmt.Sprintf("%d/%d ", i, len(importTrees))+tree.HPath))
			}
		}
		util.PushClearProgress()

		importTrees = []*parse.Tree{}
		searchLinks = map[string]string{}

		var hPaths []string
		for hPath := range hPathsIDs {
			hPaths = append(hPaths, hPath)
		}
		sort.Strings(hPaths)
		paths := map[string][]string{}
		for _, hPath := range hPaths {
			p := idPaths[hPathsIDs[hPath]]
			parent := path.Dir(p)
			for {
				if baseTargetPath == parent {
					break
				}

				if ps, ok := paths[parent]; !ok {
					paths[parent] = []string{p}
				} else {
					ps = append(ps, p)
					ps = gulu.Str.RemoveDuplicatedElem(ps)
					paths[parent] = ps
				}
				p = parent
				parent = path.Dir(parent)
			}
		}

		sortIDVals := map[string]int{}
		for _, ps := range paths {
			sortVal := 0
			for _, p := range ps {
				sortIDVals[util.GetTreeID(p)] = sortVal
				sortVal++
			}
		}
		box.setSort(sortIDVals)
	}

	IncSync()
	debug.FreeOSMemory()
	return
}

func ImportEbookFromLocalPath(boxID, localPath, toPath string) error {
	ext := strings.ToLower(filepath.Ext(localPath))
	if !isSupportedEbookImportExt(ext) {
		return fmt.Errorf("unsupported ebook format [%s]", ext)
	}
	if !gulu.File.IsExist(localPath) {
		return fmt.Errorf("ebook file [%s] not found", localPath)
	}

	sourcePath := localPath
	cleanup := func() {}
	defer cleanup()

	if ".epub" != ext {
		if !util.IsValidEbookConvertBin(Conf.Export.EbookConvertBin) {
			return errors.New("Please configure [Settings - Export - ebook-convert executable path] first")
		}

		convertDir := filepath.Join(util.TempDir, "import", "ebook-convert", gulu.Rand.String(7))
		if err := os.MkdirAll(convertDir, 0755); err != nil {
			return err
		}
		cleanup = func() { _ = os.RemoveAll(convertDir) }

		sourcePath = filepath.Join(convertDir, strings.TrimSuffix(filepath.Base(localPath), filepath.Ext(localPath))+".epub")
		args, err := parseEbookConvertParams()
		if err != nil {
			return err
		}
		if err = util.EbookConvert(Conf.Export.EbookConvertBin, localPath, sourcePath, args...); err != nil {
			return fmt.Errorf("convert ebook to EPUB failed: %w", err)
		}
	}

	importDir, err := convertEPUBToMarkdownImportDir(sourcePath)
	if err != nil {
		return err
	}
	defer os.RemoveAll(importDir)

	return ImportFromLocalPath(boxID, importDir, toPath)
}

func isSupportedEbookImportExt(ext string) bool {
	switch ext {
	case ".epub", ".mobi", ".azw", ".azw3":
		return true
	default:
		return false
	}
}

func parseEbookConvertParams() ([]string, error) {
	params := util.ReplaceNewline(Conf.Export.EbookConvertParams, " ")
	if "" == strings.TrimSpace(params) {
		return nil, nil
	}
	args, err := shellquote.Split(params)
	if err != nil {
		logging.LogErrorf("parse ebook-convert custom params [%s] failed: %s", params, err)
		return nil, fmt.Errorf("parse ebook-convert custom params failed: %w", err)
	}
	return args, nil
}

func convertEPUBToMarkdownImportDir(epubPath string) (string, error) {
	if !util.IsValidPandocBin(Conf.Export.PandocBin) {
		Conf.Export.PandocBin = util.PandocBinPath
		if !util.IsValidPandocBin(Conf.Export.PandocBin) {
			return "", errors.New(Conf.Language(115))
		}
	}

	importDir := filepath.Join(util.TempDir, "import", "ebook", gulu.Rand.String(7))
	if err := os.MkdirAll(importDir, 0755); err != nil {
		return "", err
	}

	bookName := util.FilterFileName(strings.TrimSuffix(filepath.Base(epubPath), filepath.Ext(epubPath)))
	if "" == bookName {
		bookName = "ebook"
	}
	outputPath := filepath.Join(importDir, bookName+".md")
	args := []string{
		epubPath,
		"--from", "epub",
		"--to", "gfm+footnotes+hard_line_breaks",
		"--extract-media", "assets",
		"-s",
		"-o", outputPath,
	}

	cmd := exec.Command(Conf.Export.PandocBin, args...)
	gulu.CmdAttr(cmd)
	cmd.Dir = importDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		msg := gulu.DecodeCmdOutput(output)
		if msg == "" {
			msg = err.Error()
		}
		_ = os.RemoveAll(importDir)
		logging.LogErrorf("convert EPUB import [%s] failed: %s", epubPath, msg)
		return "", fmt.Errorf("convert EPUB import failed: %s", msg)
	}
	return importDir, nil
}

func parseStdMd(markdown []byte) (ret *parse.Tree, yfmRootID, yfmTitle, yfmUpdated string) {
	luteEngine := util.NewStdLute()
	luteEngine.SetYamlFrontMatter(true)
	ret = parse.Parse("", markdown, luteEngine.ParseOptions)
	if nil == ret {
		return
	}
	yfmRootID, yfmTitle, yfmUpdated = normalizeTree(ret)
	htmlBlock2Media(ret)
	htmlBlock2Inline(ret)
	parse.TextMarks2Inlines(ret)
	parse.NestedInlines2FlattedSpansHybrid(ret, false)
	return
}

// Improve Markdown import to parse audio/video tags
func htmlBlock2Media(tree *parse.Tree) {
	ast.Walk(tree.Root, func(n *ast.Node, entering bool) ast.WalkStatus {
		if !entering || ast.NodeHTMLBlock != n.Type {
			return ast.WalkContinue
		}

		tokens := bytes.TrimSpace(n.Tokens)
		tokens, _ = bytes.CutPrefix(tokens, []byte("<div>"))
		tokens, _ = bytes.CutSuffix(tokens, []byte("</div>"))
		tokens = bytes.TrimSpace(tokens)

		lower := bytes.ToLower(tokens)
		if bytes.HasPrefix(lower, []byte("<audio")) && bytes.HasSuffix(tokens, []byte(">")) {
			n.Type = ast.NodeAudio
			n.Tokens = tokens
		} else if bytes.HasPrefix(lower, []byte("<video")) && bytes.HasSuffix(tokens, []byte(">")) {
			n.Type = ast.NodeVideo
			n.Tokens = tokens
		}
		return ast.WalkContinue
	})
}

func processHTMLBlockSvgImg(n *ast.Node, assetDirPath, boxID string) {
	re := regexp.MustCompile(`(?i)<svg[^>]*>(.*?)</svg>`)
	matches := re.FindStringSubmatch(string(n.Tokens))
	if 1 >= len(matches) {
		return
	}

	svgContent := matches[0]
	name, err := storeAssetForBox(boxID, assetDirPath, "image.svg", []byte(svgContent))
	if err != nil {
		logging.LogErrorf("store svg asset for box [%s] failed: %s", boxID, err)
		return
	}

	assetURL := "assets/" + name
	if boxID != "" && IsEncryptedBox(boxID) {
		assetURL += "?box=" + boxID
	}
	n.Type = ast.NodeParagraph
	img := &ast.Node{Type: ast.NodeImage}
	img.AppendChild(&ast.Node{Type: ast.NodeBang})
	img.AppendChild(&ast.Node{Type: ast.NodeOpenBracket})
	img.AppendChild(&ast.Node{Type: ast.NodeLinkText, Tokens: []byte("image")})
	img.AppendChild(&ast.Node{Type: ast.NodeCloseBracket})
	img.AppendChild(&ast.Node{Type: ast.NodeOpenParen})
	img.AppendChild(&ast.Node{Type: ast.NodeLinkDest, Tokens: []byte(assetURL)})
	img.AppendChild(&ast.Node{Type: ast.NodeCloseParen})
	n.AppendChild(img)
}
func processBase64Img(n *ast.Node, dest string, assetDirPath, boxID string) {
	sep := strings.Index(dest, ";base64,")
	str := strings.TrimSpace(dest[sep+8:])
	re := regexp.MustCompile(`(?i)%0A`)
	str = re.ReplaceAllString(str, "\n")
	var decodeErr error
	unbased, decodeErr := base64.StdEncoding.DecodeString(str)
	if nil != decodeErr {
		logging.LogErrorf("decode base64 image failed: %s", decodeErr)
		return
	}
	dataReader := bytes.NewReader(unbased)
	var img image.Image
	var ext string
	typ := dest[5:sep]
	switch typ {
	case "image/png":
		img, decodeErr = png.Decode(dataReader)
		ext = ".png"
		if nil != decodeErr {
			dataReader.Seek(0, 0)
			img, decodeErr = jpeg.Decode(dataReader)
			ext = ".jpg"
		}
	case "image/jpeg":
		img, decodeErr = jpeg.Decode(dataReader)
		ext = ".jpg"
	case "image/svg+xml":
		ext = ".svg"
	default:
		logging.LogWarnf("unsupported base64 image type [%s]", typ)
		return
	}
	if nil != decodeErr {
		logging.LogErrorf("decode base64 image failed: %s", decodeErr)
		return
	}

	name := "image" + ext
	alt := n.Parent.ChildByType(ast.NodeLinkText)
	if nil != alt {
		name = alt.TokensStr() + ext
	}
	name = util.FilterUploadFileName(name)

	var data []byte
	switch typ {
	case "image/svg+xml":
		data = unbased
	default:
		var buf bytes.Buffer
		switch typ {
		case "image/png":
			encodeErr := png.Encode(&buf, img)
			if nil != encodeErr {
				logging.LogErrorf("encode png image failed: %s", encodeErr)
				return
			}
		case "image/jpeg":
			encodeErr := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 100})
			if nil != encodeErr {
				logging.LogErrorf("encode jpeg image failed: %s", encodeErr)
				return
			}
		}
		data = buf.Bytes()
	}

	diskName, err := storeAssetForBox(boxID, assetDirPath, name, data)
	if err != nil {
		logging.LogErrorf("store base64 image for box [%s] failed: %s", boxID, err)
		return
	}
	assetURL := "assets/" + diskName
	if boxID != "" && IsEncryptedBox(boxID) {
		assetURL += "?box=" + boxID
	}
	n.Tokens = []byte(assetURL)
}

func encryptBoxAVFiles(boxID, storageAvDir string) error {
	if !IsEncryptedBox(boxID) || !gulu.File.IsExist(storageAvDir) {
		return nil
	}
	boxAVDir := filepath.Join(util.DataDir, boxID, "storage", "av")
	if err := os.MkdirAll(boxAVDir, 0755); err != nil {
		return err
	}
	return filelock.Walk(storageAvDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil || d == nil || d.IsDir() {
			return walkErr
		}
		if !strings.HasSuffix(d.Name(), ".json") {
			return nil
		}
		src, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		avID := strings.TrimSuffix(d.Name(), ".json")

		av.SetAVBoxID(avID, boxID)
		enc, encErr := av.EncryptAVData(boxID, avID, src)
		if encErr != nil {
			return encErr
		}
		return os.WriteFile(filepath.Join(boxAVDir, d.Name()), enc, 0644)
	})
}

func htmlBlock2Inline(tree *parse.Tree) {
	imgHtmlBlocks := map[*ast.Node]*html.Node{}
	aHtmlBlocks := map[*ast.Node]*html.Node{}
	var unlinks []*ast.Node
	ast.Walk(tree.Root, func(n *ast.Node, entering bool) ast.WalkStatus {
		if !entering {
			return ast.WalkContinue
		}

		if ast.NodeHTMLBlock == n.Type || (ast.NodeText == n.Type && bytes.HasPrefix(bytes.ToLower(n.Tokens), []byte("<img "))) {
			tokens := bytes.TrimSpace(n.Tokens)
			tokens, _ = bytes.CutPrefix(tokens, []byte("<div>"))
			tokens, _ = bytes.CutSuffix(tokens, []byte("</div>"))
			tokens = bytes.TrimSpace(tokens)

			htmlNodes, pErr := html.ParseFragment(bytes.NewReader(tokens), &html.Node{Type: html.ElementNode})
			if nil != pErr {
				logging.LogErrorf("parse html block [%s] failed: %s", n.Tokens, pErr)
				return ast.WalkContinue
			}
			if 1 > len(htmlNodes) {
				return ast.WalkContinue
			}

			for _, htmlNode := range htmlNodes {
				if atom.Img == htmlNode.DataAtom {
					imgHtmlBlocks[n] = htmlNode
					break
				}
			}
		}
		if ast.NodeHTMLBlock == n.Type || (ast.NodeText == n.Type && bytes.HasPrefix(bytes.ToLower(n.Tokens), []byte("<a "))) {
			tokens := bytes.TrimSpace(n.Tokens)
			tokens, _ = bytes.CutPrefix(tokens, []byte("<div>"))
			tokens, _ = bytes.CutSuffix(tokens, []byte("</div>"))
			tokens = bytes.TrimSpace(tokens)

			if ast.NodeHTMLBlock != n.Type && nil != n.Next && nil != n.Next.Next {
				if ast.NodeText == n.Next.Next.Type && bytes.Equal(n.Next.Next.Tokens, []byte("</a>")) {
					tokens = append(tokens, n.Next.Tokens...)
					tokens = append(tokens, []byte("</a>")...)
					unlinks = append(unlinks, n.Next)
					unlinks = append(unlinks, n.Next.Next)
				}
			}

			htmlNodes, pErr := html.ParseFragment(bytes.NewReader(tokens), &html.Node{Type: html.ElementNode})
			if nil != pErr {
				logging.LogErrorf("parse html block [%s] failed: %s", n.Tokens, pErr)
				return ast.WalkContinue
			}
			if 1 > len(htmlNodes) {
				return ast.WalkContinue
			}

			for _, htmlNode := range htmlNodes {
				if atom.A == htmlNode.DataAtom {
					aHtmlBlocks[n] = htmlNode
					break
				}
			}
		}
		return ast.WalkContinue
	})

	for n, htmlImg := range imgHtmlBlocks {
		src := domAttrValue(htmlImg, "src")
		alt := domAttrValue(htmlImg, "alt")
		title := domAttrValue(htmlImg, "title")

		p := treenode.NewParagraph(n.ID)
		img := &ast.Node{Type: ast.NodeImage}
		p.AppendChild(img)
		img.AppendChild(&ast.Node{Type: ast.NodeBang})
		img.AppendChild(&ast.Node{Type: ast.NodeOpenBracket})
		img.AppendChild(&ast.Node{Type: ast.NodeLinkText, Tokens: []byte(alt)})
		img.AppendChild(&ast.Node{Type: ast.NodeCloseBracket})
		img.AppendChild(&ast.Node{Type: ast.NodeOpenParen})
		img.AppendChild(&ast.Node{Type: ast.NodeLinkDest, Tokens: []byte(src)})
		if "" != title {
			img.AppendChild(&ast.Node{Type: ast.NodeLinkSpace})
			img.AppendChild(&ast.Node{Type: ast.NodeLinkTitle})
		}
		img.AppendChild(&ast.Node{Type: ast.NodeCloseParen})
		if width := domAttrValue(htmlImg, "width"); "" != width {
			if util2.IsDigit(width) {
				width += "px"
			}
			style := "width: " + width + ";"
			ial := &ast.Node{Type: ast.NodeKramdownSpanIAL, Tokens: parse.IAL2Tokens([][]string{{"style", style}})}
			img.SetIALAttr("style", style)
			img.InsertAfter(ial)
		} else if height := domAttrValue(htmlImg, "height"); "" != height {
			if util2.IsDigit(height) {
				height += "px"
			}
			style := "height: " + height + ";"
			ial := &ast.Node{Type: ast.NodeKramdownSpanIAL, Tokens: parse.IAL2Tokens([][]string{{"style", style}})}
			img.SetIALAttr("style", style)
			img.InsertAfter(ial)
		}

		if ast.NodeHTMLBlock == n.Type {
			n.InsertBefore(p)
		} else if ast.NodeText == n.Type {
			if nil != n.Parent {
				if n.Parent.IsContainerBlock() {
					n.InsertBefore(p)
				} else {
					n.InsertBefore(img)
				}
			} else {
				n.InsertBefore(p)
			}
		}
		unlinks = append(unlinks, n)
	}

	for n, htmlA := range aHtmlBlocks {
		href := domAttrValue(htmlA, "href")
		title := domAttrValue(htmlA, "title")
		anchor := strings.TrimSpace(util2.DomText(htmlA))

		if "" == anchor {
			unlinks = append(unlinks, n)
			if nil != n.Next && ast.NodeText == n.Next.Type && "</a>" == n.NextNodeText() {
				unlinks = append(unlinks, n.Next)
			}
			continue
		}

		p := treenode.NewParagraph(n.ID)
		a := &ast.Node{Type: ast.NodeLink}
		p.AppendChild(a)
		a.AppendChild(&ast.Node{Type: ast.NodeOpenBracket})
		a.AppendChild(&ast.Node{Type: ast.NodeLinkText, Tokens: []byte(anchor)})
		a.AppendChild(&ast.Node{Type: ast.NodeCloseBracket})
		a.AppendChild(&ast.Node{Type: ast.NodeOpenParen})
		a.AppendChild(&ast.Node{Type: ast.NodeLinkDest, Tokens: []byte(href)})
		if "" != title {
			a.AppendChild(&ast.Node{Type: ast.NodeLinkSpace})
			a.AppendChild(&ast.Node{Type: ast.NodeLinkTitle, Tokens: []byte(title)})
		}
		a.AppendChild(&ast.Node{Type: ast.NodeCloseParen})

		if ast.NodeHTMLBlock == n.Type || (nil == n.Previous && (nil != n.Next && nil != n.Next.Next && nil == n.Next.Next.Next)) {
			n.InsertBefore(p)
		} else {
			n.InsertBefore(a)
		}
		unlinks = append(unlinks, n)
	}

	for _, n := range unlinks {
		n.Unlink()
	}
}

func reassignIDUpdated(tree *parse.Tree, rootID, updated string) {
	ast.Walk(tree.Root, func(n *ast.Node, entering bool) ast.WalkStatus {
		if !entering || "" == n.ID {
			return ast.WalkContinue
		}

		n.ID = ast.NewNodeID()
		if ast.NodeDocument == n.Type && "" != rootID {
			n.ID = rootID
		}

		n.SetIALAttr("id", n.ID)
		if "" != updated {
			n.SetIALAttr("updated", updated)
			if "" == rootID {
				n.ID = updated + "-" + gulu.Rand.String(7)
				n.SetIALAttr("id", n.ID)
			}
		} else {
			n.SetIALAttr("updated", util.TimeFromID(n.ID))
		}
		return ast.WalkContinue
	})
	tree.ID = tree.Root.ID
	tree.Path = path.Join(path.Dir(tree.Path), tree.ID+".sy")
	tree.Root.SetIALAttr("id", tree.Root.ID)
}

func domAttrValue(n *html.Node, attrName string) string {
	if nil == n {
		return ""
	}

	for _, attr := range n.Attr {
		if attr.Key == attrName {
			return attr.Val
		}
	}
	return ""
}

var importTrees []*parse.Tree
var searchLinks = map[string]string{}

func initSearchLinks() {
	for _, tree := range importTrees {
		ast.Walk(tree.Root, func(n *ast.Node, entering bool) ast.WalkStatus {
			if !entering || (ast.NodeDocument != n.Type && ast.NodeHeading != n.Type) {
				return ast.WalkContinue
			}

			nodePath := tree.HPath + "#"
			if ast.NodeHeading == n.Type {
				nodePath += n.Text()
			}

			searchLinks[nodePath] = n.ID
			return ast.WalkContinue
		})
	}
}

func convertMdHyperlinks2WikiLinks() {
	// Supports converting relative path hyperlinks into document block references after importing Markdown

	var unlinks []*ast.Node
	for _, tree := range importTrees {
		ast.Walk(tree.Root, func(n *ast.Node, entering bool) ast.WalkStatus {
			if !entering || ast.NodeTextMark != n.Type {
				return ast.WalkContinue
			}

			if "a" != n.TextMarkType {
				return ast.WalkContinue
			}

			linkText := n.TextMarkTextContent
			if "" == linkText {
				return ast.WalkContinue
			}
			linkDest := n.TextMarkAHref
			if "" == linkDest {
				return ast.WalkContinue
			}
			if strings.HasPrefix(linkDest, "assets/") {
				return ast.WalkContinue
			}
			if !strings.HasSuffix(linkDest, ".md") && !strings.HasSuffix(linkDest, ".markdown") {
				return ast.WalkContinue
			}
			linkDest = strings.TrimSuffix(linkDest, ".md")
			linkDest = strings.TrimSuffix(linkDest, ".markdown")

			buf := bytes.Buffer{}
			buf.WriteString("[[")
			buf.WriteString(linkDest)
			buf.WriteString("|")
			buf.WriteString(linkText)
			buf.WriteString("]]")

			wikilinkNode := &ast.Node{Type: ast.NodeText, Tokens: buf.Bytes()}
			n.InsertBefore(wikilinkNode)
			unlinks = append(unlinks, n)
			return ast.WalkContinue
		})
	}

	for _, n := range unlinks {
		n.Unlink()
	}
}

func convertWikiLinksAndTags() {
	for _, tree := range importTrees {
		convertWikiLinksAndTags0(tree)
	}
}

func convertWikiLinksAndTags0(tree *parse.Tree) {
	ast.Walk(tree.Root, func(n *ast.Node, entering bool) ast.WalkStatus {
		if !entering || ast.NodeText != n.Type {
			return ast.WalkContinue
		}

		text := n.TokensStr()
		length := len(text)
		start, end := 0, length
		for {
			part := text[start:end]
			if idx := strings.Index(part, "]]"); 0 > idx {
				break
			} else {
				end = start + idx
			}
			if idx := strings.Index(part, "[["); 0 > idx {
				break
			} else {
				start += idx
			}
			if end <= start {
				break
			}

			link := path.Join(path.Dir(tree.HPath), text[start+2:end])
			linkText := path.Base(link)
			dynamicAnchorText := true
			if linkParts := strings.Split(link, "|"); 1 < len(linkParts) {
				link = linkParts[0]
				linkText = linkParts[1]
				dynamicAnchorText = false
			}
			link, linkText = strings.TrimSpace(link), strings.TrimSpace(linkText)
			if !strings.Contains(link, "#") {
				link += "#"
			}

			id := searchLinkID(link)
			if "" == id {
				start, end = end, length
				continue
			}

			linkText = strings.TrimPrefix(linkText, "/")
			repl := "((" + id + " '" + linkText + "'))"
			if !dynamicAnchorText {
				repl = "((" + id + " \"" + linkText + "\"))"
			}
			end += 2
			text = text[:start] + repl + text[end:]
			start, end = start+len(repl), len(text)
			length = end
		}

		text = convertTags(text)
		n.Tokens = gulu.Str.ToBytes(text)
		return ast.WalkContinue
	})
}

func convertTags(text string) (ret string) {
	if !util.MarkdownSettings.InlineTag {
		return text
	}

	pos, i := -1, 0
	tokens := []byte(text)
	for ; i < len(tokens); i++ {
		if '#' == tokens[i] && (0 == i || ' ' == tokens[i-1] || (-1 < pos && '#' == tokens[pos])) {
			if i < len(tokens)-1 && '#' == tokens[i+1] {
				pos = -1
				continue
			}
			pos = i
			continue
		}

		if -1 < pos && ' ' == tokens[i] {
			tokens = append(tokens, 0)
			copy(tokens[i+1:], tokens[i:])
			tokens[i] = '#'
			pos = -1
			i++
		}
	}
	if -1 < pos && pos < i {
		tokens = append(tokens, '#')
	}
	return string(tokens)
}

func searchLinkID(link string) (id string) {
	id = searchLinks[link]
	if "" != id {
		return
	}

	baseName := path.Base(link)
	for searchLink, searchID := range searchLinks {
		if path.Base(searchLink) == baseName {
			return searchID
		}
	}
	return
}

func mergeTextAndHandlerNestedInlines() {
	luteEngine := NewLute()
	luteEngine.SetHTMLTag2TextMark(true)
	for _, tree := range importTrees {
		tree.MergeText()

		var unlinkTextNodes []*ast.Node
		ast.Walk(tree.Root, func(n *ast.Node, entering bool) ast.WalkStatus {
			if !entering || ast.NodeText != n.Type {
				return ast.WalkContinue
			}

			if nil == n.Tokens {
				return ast.WalkContinue
			}

			t := parse.Inline("", n.Tokens, luteEngine.ParseOptions)
			parse.NestedInlines2FlattedSpans(t, false)
			var children []*ast.Node
			for c := t.Root.FirstChild.FirstChild; nil != c; c = c.Next {
				children = append(children, c)
			}
			for _, c := range children {
				n.InsertBefore(c)
			}
			unlinkTextNodes = append(unlinkTextNodes, n)
			return ast.WalkContinue
		})

		for _, node := range unlinkTextNodes {
			node.Unlink()
		}
	}
}
