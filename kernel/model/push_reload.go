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
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/emirpasic/gods/sets/hashset"
	"github.com/icha-senpai/note/kernel/av"
	"github.com/icha-senpai/note/kernel/conf"
	"github.com/icha-senpai/note/kernel/filesys"
	"github.com/icha-senpai/note/kernel/sql"
	"github.com/icha-senpai/note/kernel/task"
	"github.com/icha-senpai/note/kernel/treenode"
	"github.com/icha-senpai/note/kernel/util"
	"github.com/icha-senpai/note/third_party/forks/go-humanize"
	"github.com/icha-senpai/note/third_party/forks/gulu"
	"github.com/icha-senpai/note/third_party/forks/logging"
	"github.com/icha-senpai/note/third_party/forks/lute"
	"github.com/icha-senpai/note/third_party/forks/lute/ast"
	"github.com/icha-senpai/note/third_party/forks/lute/parse"
	"github.com/icha-senpai/note/third_party/forks/lute/render"
)

func PushReloadSnippet(snippet *conf.Snpt) {
	util.BroadcastByType("main", "setSnippet", 0, "", snippet)
}

func PushReloadPlugin(uninstallPluginNameSet, unloadPluginNameSet, reloadPluginSet, dataChangePluginSet *hashset.Set, excludeApp string) {

	orderedSets := []*hashset.Set{uninstallPluginNameSet, unloadPluginNameSet, reloadPluginSet, dataChangePluginSet}
	slices := make([][]string, len(orderedSets))

	for i, set := range orderedSets {
		if nil != set {

			for _, n := range set.Values() {
				name := n.(string)

				for _, lowerSet := range orderedSets[i+1:] {
					if nil != lowerSet {
						lowerSet.Remove(name)
					}
				}
			}
		}

		if nil == set {
			slices[i] = []string{}
		} else {
			strs := make([]string, 0, set.Size())
			for _, n := range set.Values() {
				strs = append(strs, n.(string))
			}
			slices[i] = strs
		}
	}

	logging.LogInfof("reload plugins, uninstalls=%v, unloads=%v, reloads=%v, dataChanges=%v", slices[0], slices[1], slices[2], slices[3])
	payload := map[string]any{
		"uninstallPlugins":  slices[0],
		"unloadPlugins":     slices[1],
		"reloadPlugins":     slices[2],
		"dataChangePlugins": slices[3],
	}

	if "" == excludeApp {
		util.BroadcastByType("main", "reloadPlugin", 0, "", payload)
		return
	}
	util.BroadcastByTypeAndExcludeApp(excludeApp, "main", "reloadPlugin", 0, "", payload)
}

func refreshDocInfo(tree *parse.Tree) {
	if nil == tree {
		return
	}

	refreshDocInfoWithSize(tree, filesys.TreeSize(tree))
}

func refreshDocInfoWithSize(tree *parse.Tree, size uint64) {
	if nil == tree {
		return
	}

	refreshDocInfo0(tree, size)
	go func() {
		time.Sleep(128 * time.Millisecond)
		refreshParentDocInfo(tree)
	}()
}

func refreshParentDocInfo(tree *parse.Tree) {
	if nil == tree {
		return
	}

	parentTree := loadParentTree(tree)
	if nil == parentTree {
		return
	}

	luteEngine := lute.New()
	renderer := render.NewJSONRenderer(parentTree, luteEngine.RenderOptions, luteEngine.ParseOptions)
	data := renderer.Render()
	refreshDocInfo0(parentTree, uint64(len(data)))
}

func refreshBoxDocInfo(tree *parse.Tree) {
	if nil == tree || path.Dir(tree.Path) != "/" || IsBoxDoc(tree.Box, tree.ID) {
		return
	}
	refreshBoxDocInfoByBoxID(tree.Box)
}

func refreshBoxDocInfoByBoxID(boxID string) {
	if !IsBoxDocEnabled() {
		return
	}
	box := Conf.Box(boxID)
	if nil == box {
		return
	}
	util.BroadcastByType("filetree", "reloadNotebookInfo", 0, "", boxID)
}

func refreshDocInfo0(tree *parse.Tree, size uint64) {
	cTime, _ := time.ParseInLocation("20060102150405", tree.ID[:14], time.Local)
	mTime := cTime
	if updated := tree.Root.IALAttr("updated"); "" != updated {
		if updatedTime, err := time.ParseInLocation("20060102150405", updated, time.Local); err == nil {
			mTime = updatedTime
		}
	}

	subFileCount := 0
	if IsBoxDoc(tree.Box, tree.ID) {
		subFileCount = BoxDocSubFileCount(tree.Box)
	} else if "true" != tree.Root.IALAttr(DocHiddenAttr) {
		subDir := filepath.Join(util.DataDir, tree.Box, strings.TrimSuffix(tree.Path, ".sy"))
		subFiles, err := os.ReadDir(subDir)
		if err == nil {
			for _, subFile := range subFiles {
				if !strings.HasSuffix(subFile.Name(), ".sy") {
					continue
				}

				subDocIAL := filesys.DocIAL(filepath.Join(subDir, subFile.Name()))
				if "true" == subDocIAL[DocHiddenAttr] {
					continue
				}
				subFileCount++
			}
		}
	}

	docInfo := map[string]any{
		"box":          tree.Box,
		"rootID":       tree.ID,
		"name":         tree.Root.IALAttr("title"),
		"alias":        tree.Root.IALAttr("alias"),
		"name1":        tree.Root.IALAttr("name"),
		"memo":         tree.Root.IALAttr("memo"),
		"bookmark":     tree.Root.IALAttr("bookmark"),
		"size":         size,
		"hSize":        humanize.BytesCustomCeil(size, 2),
		"mtime":        mTime.Unix(),
		"ctime":        cTime.Unix(),
		"hMtime":       mTime.Format("2006-01-02 15:04:05") + ", " + util.HumanizeTime(mTime, Conf.Lang),
		"hCtime":       cTime.Format("2006-01-02 15:04:05") + ", " + util.HumanizeTime(cTime, Conf.Lang),
		"subFileCount": subFileCount,
	}

	task.AppendAsyncTaskWithDelay(task.ReloadProtyle, 500*time.Millisecond, util.PushReloadDocInfo, docInfo)
}

func ReloadFiletree() {
	task.AppendAsyncTaskWithDelay(task.ReloadFiletree, 200*time.Millisecond, util.PushReloadFiletree)
}

func ReloadTag() {
	task.AppendAsyncTaskWithDelay(task.ReloadTag, 200*time.Millisecond, util.PushReloadTag)
}

func ReloadProtyle(rootID string) {

	defTree, _ := LoadTreeByBlockID(rootID)
	if nil != defTree {
		defIDs := sql.QueryChildDefIDsByRootDefID(rootID)

		var defNodes []*ast.Node
		ast.Walk(defTree.Root, func(n *ast.Node, entering bool) ast.WalkStatus {
			if !entering || !n.IsBlock() {
				return ast.WalkContinue
			}

			if gulu.Str.Contains(n.ID, defIDs) {
				defNodes = append(defNodes, n)
			}
			return ast.WalkContinue
		})

		for _, def := range defNodes {
			refreshDynamicRefText(def, defTree)
		}
	}

	refIDs := sql.QueryRefIDsByDefID(rootID, true)
	var rootIDs []string
	bts := treenode.GetBlockTrees(refIDs)
	for _, bt := range bts {
		rootIDs = append(rootIDs, bt.RootID)
	}
	rootIDs = gulu.Str.RemoveDuplicatedElem(rootIDs)
	for _, id := range rootIDs {
		task.AppendAsyncTaskWithDelay(task.ReloadProtyle, 200*time.Millisecond, util.PushReloadProtyle, id)
	}

	task.AppendAsyncTaskWithDelay(task.ReloadProtyle, 200*time.Millisecond, util.PushReloadProtyle, rootID)
}

func refreshRefCount(blockID string) {
	sql.FlushQueue()

	bt := treenode.GetBlockTree(blockID)
	if nil == bt {
		return
	}

	isDoc := bt.ID == bt.RootID
	var rootRefIDs []string
	var refCount, rootRefCount int
	refIDs := sql.QueryRefIDsByDefID(bt.ID, isDoc)
	if isDoc {
		rootRefIDs = refIDs
	} else {
		rootRefIDs = sql.QueryRefIDsByDefID(bt.RootID, true)
	}
	refCount = len(refIDs)
	rootRefCount = len(rootRefIDs)
	var defIDs []string
	if isDoc {
		defIDs = sql.QueryChildDefIDsByRootDefID(bt.ID)
	} else {
		defIDs = append(defIDs, bt.ID)
	}

	util.PushSetDefRefCount(bt.RootID, blockID, defIDs, refCount, rootRefCount)
}

func refreshDynamicRefText(updatedDefNode *ast.Node, updatedTree *parse.Tree) {
	changedDefs := map[string]*ast.Node{updatedDefNode.ID: updatedDefNode}
	changedTrees := map[string]*parse.Tree{updatedTree.ID: updatedTree}
	refreshDynamicRefTexts(changedDefs, changedTrees)
}

func refreshDynamicRefTexts(updatedDefNodes map[string]*ast.Node, updatedTrees map[string]*parse.Tree) (changedRootIDs []string) {
	for t := range updatedTrees {
		changedRootIDs = append(changedRootIDs, t)
	}

	for range 7 {
		updatedRefNodes, updatedRefTrees := refreshDynamicRefTexts0(updatedDefNodes, updatedTrees)
		if 1 > len(updatedRefNodes) {
			break
		}
		updatedDefNodes, updatedTrees = updatedRefNodes, updatedRefTrees

		for t := range updatedTrees {
			changedRootIDs = append(changedRootIDs, t)
		}
	}

	changedRootIDs = gulu.Str.RemoveDuplicatedElem(changedRootIDs)
	return
}

func refreshDynamicRefTexts0(updatedDefNodes map[string]*ast.Node, updatedTrees map[string]*parse.Tree) (updatedRefNodes map[string]*ast.Node, updatedRefTrees map[string]*parse.Tree) {
	updatedRefNodes = map[string]*ast.Node{}
	updatedRefTrees = map[string]*parse.Tree{}

	treeRefNodeIDs := map[string]*hashset.Set{}
	var changedNodes []*ast.Node
	var refs []*sql.Ref
	for _, updateNode := range updatedDefNodes {
		refs, changedNodes = getRefsCacheByDefNode(updateNode)
		for _, ref := range refs {
			if refIDs, ok := treeRefNodeIDs[ref.RootID]; !ok {
				refIDs = hashset.New()
				refIDs.Add(ref.BlockID)
				treeRefNodeIDs[ref.RootID] = refIDs
			} else {
				refIDs.Add(ref.BlockID)
			}
		}
	}
	for _, n := range changedNodes {
		updatedDefNodes[n.ID] = n
	}

	changedRefTree := map[string]*parse.Tree{}

	for refTreeID, refNodeIDs := range treeRefNodeIDs {
		refTree, ok := updatedTrees[refTreeID]
		if !ok {
			var err error
			refTree, err = LoadTreeByBlockID(refTreeID)
			if err != nil {
				continue
			}
		}

		var refTreeChanged bool
		ast.Walk(refTree.Root, func(n *ast.Node, entering bool) ast.WalkStatus {
			if !entering {
				return ast.WalkContinue
			}

			if n.IsBlock() && refNodeIDs.Contains(n.ID) {
				changed, changedDefNodes := updateRefText(n, updatedDefNodes)
				if !refTreeChanged && changed {
					refTreeChanged = true
					updatedRefNodes[n.ID] = n
					updatedRefTrees[refTreeID] = refTree
				}

				for _, defNode := range changedDefNodes {
					switch defNode.refType {
					case "ref-d":
						task.AppendAsyncTaskWithDelay(task.SetRefDynamicText, 200*time.Millisecond, util.PushSetRefDynamicText, refTreeID, n.ID, defNode.id, defNode.refText, refTree.Box)
					}
				}
				return ast.WalkContinue
			}
			return ast.WalkContinue
		})

		if refTreeChanged {
			changedRefTree[refTreeID] = refTree
			sql.UpdateRefsTreeQueue(refTree)
		}
	}

	updateAttributeViewBlockText(updatedDefNodes)

	for _, tree := range changedRefTree {
		indexWriteTreeUpsertQueue(tree)
	}
	return
}

func updateAttributeViewBlockText(updatedDefNodes map[string]*ast.Node) {
	var parents []*ast.Node
	for _, updatedDefNode := range updatedDefNodes {
		for parent := updatedDefNode.Parent; nil != parent && ast.NodeDocument != parent.Type; parent = parent.Parent {
			parents = append(parents, parent)
		}
	}
	for _, parent := range parents {
		updatedDefNodes[parent.ID] = parent
	}

	for _, updatedDefNode := range updatedDefNodes {
		avs := updatedDefNode.IALAttr(av.NodeAttrNameAvs)
		if "" == avs {
			continue
		}

		avIDs := strings.SplitSeq(avs, ",")
		for avID := range avIDs {
			attrView, parseErr := av.ParseAttributeView(avID)
			if nil != parseErr {
				continue
			}

			changedAv := false
			blockValues := attrView.GetBlockKeyValues()
			if nil == blockValues {
				continue
			}

			for _, blockValue := range blockValues.Values {
				if blockValue.Block.ID == updatedDefNode.ID {
					newIcon, newContent := getNodeAvBlockText(updatedDefNode, avID)
					if newIcon != blockValue.Block.Icon {
						blockValue.Block.Icon = newIcon
						changedAv = true
					}
					if newContent != blockValue.Block.Content {
						blockValue.Block.Content = util.UnescapeHTML(newContent)
						changedAv = true
					}
					break
				}
			}
			if changedAv {
				av.SaveAttributeView(attrView)
				ReloadAttrView(avID)

				refreshRelatedSrcAvs(avID, nil)
			}
		}
	}
}

func ReloadAttrView(avID string) {
	task.AppendAsyncTaskWithDelay(task.ReloadAttributeView, 200*time.Millisecond, pushReloadAttrView, avID)
}

func pushReloadAttrView(avID string) {
	util.BroadcastByType("protyle", "refreshAttributeView", 0, "", map[string]any{"id": avID})
}

func PushCreate(box *Box, p string, arg map[string]any) {
	evt := util.NewCmdResult("create", 0, util.PushModeBroadcast)
	listDocTree := false
	if nil == arg {
		arg = map[string]any{
			"listDocTree": true,
		}
	}

	listDocTreeArg := arg["listDocTree"]
	if nil != listDocTreeArg {
		listDocTree = listDocTreeArg.(bool)
	}

	evt.Data = map[string]any{
		"box":         box,
		"path":        p,
		"listDocTree": listDocTree,
	}
	util.PushEvent(evt)
}
