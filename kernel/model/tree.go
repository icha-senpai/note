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

package model

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/icha-senpai/note/third_party/forks/lute"
	"github.com/icha-senpai/note/third_party/forks/lute/ast"
	"github.com/icha-senpai/note/third_party/forks/lute/parse"
	"github.com/icha-senpai/note/kernel/av"
	"github.com/icha-senpai/note/kernel/filesys"
	"github.com/icha-senpai/note/kernel/search"
	"github.com/icha-senpai/note/kernel/sql"
	"github.com/icha-senpai/note/kernel/task"
	"github.com/icha-senpai/note/kernel/treenode"
	"github.com/icha-senpai/note/kernel/util"
	"github.com/icha-senpai/note/third_party/forks/dataparser"
	"github.com/icha-senpai/note/third_party/forks/filelock"
	"github.com/icha-senpai/note/third_party/forks/logging"
	"github.com/icha-senpai/note/third_party/forks/external/golang.org/x/time/rate"
)

func resetTree(tree *parse.Tree, titleSuffix string, removeAvBinding bool) {
	tree.ID = ast.NewNodeID()
	tree.Root.ID = tree.ID
	title := tree.Root.IALAttr("title")
	if "" != titleSuffix {
		if t, parseErr := time.Parse("20060102150405", util.TimeFromID(tree.ID)); nil == parseErr {
			titleSuffix += " " + t.Format("2006-01-02 15:04:05")
		} else {
			titleSuffix = "Duplicated " + time.Now().Format("2006-01-02 15:04:05")
		}
		titleSuffix = "(" + titleSuffix + ")"
		titleSuffix = " " + titleSuffix
		if Conf.language(16) == title {
			titleSuffix = ""
		}
	}
	tree.Root.SetIALAttr("id", tree.ID)
	tree.Root.SetIALAttr("title", title+titleSuffix)
	tree.Root.RemoveIALAttr("scroll")
	p := path.Join(path.Dir(tree.Path), tree.ID) + ".sy"
	tree.Path = p
	tree.HPath = tree.HPath + " " + titleSuffix

	refIDs := map[string]string{}
	ast.Walk(tree.Root, func(n *ast.Node, entering bool) ast.WalkStatus {
		if !entering || !treenode.IsBlockRef(n) {
			return ast.WalkContinue
		}
		defID, _, _ := treenode.GetBlockRef(n)
		if "" == defID {
			return ast.WalkContinue
		}
		refIDs[defID] = "1"
		return ast.WalkContinue
	})

	ast.Walk(tree.Root, func(n *ast.Node, entering bool) ast.WalkStatus {
		if !entering || ast.NodeDocument == n.Type {
			return ast.WalkContinue
		}
		if n.IsBlock() && "" != n.ID {
			newID := ast.NewNodeID()
			if "1" == refIDs[n.ID] {

				refIDs[n.ID] = newID
			}
			n.ID = newID
			n.SetIALAttr("id", n.ID)
		}
		return ast.WalkContinue
	})

	ast.Walk(tree.Root, func(n *ast.Node, entering bool) ast.WalkStatus {
		if !entering || !treenode.IsBlockRef(n) {
			return ast.WalkContinue
		}
		defID, _, _ := treenode.GetBlockRef(n)
		if "" == defID {
			return ast.WalkContinue
		}
		if "1" != refIDs[defID] {
			if ast.NodeTextMark == n.Type {
				n.TextMarkBlockRefID = refIDs[defID]
			}
		}
		return ast.WalkContinue
	})

	var attrViewIDs []string

	ast.Walk(tree.Root, func(n *ast.Node, entering bool) ast.WalkStatus {
		if !entering {
			return ast.WalkContinue
		}

		if ast.NodeAttributeView == n.Type {
			av.UpsertBlockRel(n.AttributeViewID, n.ID)
			attrViewIDs = append(attrViewIDs, n.AttributeViewID)
		}
		return ast.WalkContinue
	})

	if removeAvBinding {

		tree.Root.RemoveIALAttr(av.NodeAttrNameAvs)
	}
}

func pagedPaths(localPath string, pageSize int) (ret map[int][]string) {
	ret = map[int][]string{}
	page := 1
	filelock.Walk(localPath, func(path string, d fs.DirEntry, err error) error {
		if nil != err || nil == d {
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

		ret[page] = append(ret[page], path)
		if pageSize <= len(ret[page]) {
			page++
		}
		return nil
	})
	return
}

func loadTree(localPath string, luteEngine *lute.Lute) (ret *parse.Tree, err error) {
	data, err := filelock.ReadFile(localPath)
	if err != nil {
		logging.LogErrorf("get data [path=%s] failed: %s", localPath, err)
		return
	}

	if boxID := extractBoxIDFromPath(localPath); boxID != "" && IsEncryptedBox(boxID) {
		HoldBoxReadLock(boxID)
		defer ReleaseBoxReadLock(boxID)
		dek, dekErr := GetDEKIfUnlocked(boxID)
		if dekErr != nil {
			err = dekErr
			return
		}

		relPath := filepath.ToSlash(strings.TrimPrefix(localPath, filepath.Join(util.DataDir, boxID)+string(os.PathSeparator)))
		if data, err = DecryptFile(boxID, relPath, dek, data); err != nil {
			logging.LogErrorf("decrypt tree [path=%s] failed: %s", localPath, err)
			return
		}
	}

	ret, err = dataparser.ParseJSONWithoutFix(data, luteEngine.ParseOptions)
	if err != nil {
		logging.LogErrorf("parse json to tree [%s] failed: %s", localPath, err)
		return
	}
	return
}

var (
	ErrBoxNotFound   = errors.New("notebook not found")
	ErrBlockNotFound = errors.New("block not found")
	ErrTreeNotFound  = errors.New("tree not found")
	ErrIndexing      = errors.New("indexing")
	ErrBoxUnindexed  = errors.New("notebook unindexed")
	ErrInvalidID     = errors.New("invalid id")
)

func LoadTreeByBlockIDWithReindex(id string) (ret *parse.Tree, err error) {
	return LoadTreeByBlockIDWithReindexInBox(id, "")
}

func LoadTreeByBlockIDWithReindexInBox(id, boxID string) (ret *parse.Tree, err error) {
	if "" == id {
		logging.LogWarnf("block id is empty")
		return nil, ErrTreeNotFound
	}

	bt := treenode.GetBlockTreeInBox(id, boxID)
	if nil == bt && "" == boxID {

		for _, encBoxID := range treenode.GetOpenedEncryptedBoxIDs() {
			if encBT := treenode.GetBlockTreeInBox(id, encBoxID); nil != encBT {
				bt = encBT
				break
			}
		}
	}
	if nil == bt {
		if task.ContainIndexTask() {
			err = ErrIndexing
			return
		}

		err = indexTreeInFilesystem(id)
		bt = treenode.GetBlockTreeInBox(id, boxID)
		if nil == bt {
			if "dev" == util.Mode {
				logging.LogWarnf("block tree not found [id=%s], stack: [%s]", id, logging.ShortStack())
			}
			return
		}
	}

	luteEngine := util.NewLute()
	ret, err = filesys.LoadTree(bt.BoxID, bt.Path, luteEngine)
	return
}

func LoadTreeByBlockID(id string) (ret *parse.Tree, err error) {
	return loadTreeByBlockIDInBox(id, "")
}

func loadTreeByBlockTree(bt *treenode.BlockTree) (ret *parse.Tree, err error) {
	luteEngine := util.NewLute()
	ret, needFix, err := filesys.LoadTreeWithFix(bt.BoxID, bt.Path, luteEngine)
	if nil != err {
		return
	}
	if needFix {
		treenode.UpsertBlockTree(ret)
		sql.IndexTreeQueue(ret)
	}
	return
}

func loadTreeByBlockIDInBox(id, boxID string) (ret *parse.Tree, err error) {
	if !ast.IsNodeIDPattern(id) {
		stack := logging.ShortStack()
		logging.LogErrorf("block id is invalid [id=%s], stack: [%s]", id, stack)
		return nil, ErrTreeNotFound
	}

	bt := treenode.GetBlockTreeInBox(id, boxID)
	if nil == bt && "" == boxID {

		for _, encBoxID := range treenode.GetOpenedEncryptedBoxIDs() {
			if encBT := treenode.GetBlockTreeInBox(id, encBoxID); nil != encBT {
				bt = encBT
				break
			}
		}
	}
	if nil == bt {
		if task.ContainIndexTask() {
			err = ErrIndexing
			return
		}

		stack := logging.ShortStack()
		if !strings.Contains(stack, "BuildBlockBreadcrumb") {
			if "dev" == util.Mode {
				logging.LogWarnf("block tree not found [id=%s], stack: [%s]", id, stack)
			}
		}
		return nil, ErrTreeNotFound
	}

	ret, err = loadTreeByBlockTree(bt)
	return
}

var searchTreeLimiter = rate.NewLimiter(rate.Every(3*time.Second), 1)

func indexTreeInFilesystem(blockID string) error {
	if !searchTreeLimiter.Allow() {
		return ErrIndexing
	}

	msdID := util.PushMsg(Conf.language(45), 7000)
	defer util.PushClearMsg(msdID)

	logging.LogWarnf("searching tree on filesystem [id=%s]", blockID)

	unindexedTreePath := findUnindexedTreePathInAllBoxes(blockID)
	if "" == unindexedTreePath {
		logging.LogInfof("tree not found on filesystem [id=%s]", blockID)
		return ErrTreeNotFound
	}

	boxID := strings.TrimPrefix(unindexedTreePath, util.DataDir)
	boxID = boxID[1:]
	boxID = boxID[:strings.Index(boxID, string(os.PathSeparator))]
	unindexedTreePath = strings.TrimPrefix(unindexedTreePath, util.DataDir)
	unindexedTreePath = strings.TrimPrefix(unindexedTreePath, string(os.PathSeparator))
	unindexedTreePath = strings.TrimPrefix(unindexedTreePath, boxID)
	unindexedTreePath = filepath.ToSlash(unindexedTreePath)
	if nil == Conf.Box(boxID) {
		for _, b := range Conf.GetClosedBoxes() {
			if b.ID == boxID {
				logging.LogInfof("box [%s] is closed", boxID)
				util.PushErrMsg(fmt.Sprintf(Conf.language(197), b.Name), 7000)
				return ErrBoxUnindexed
			}
		}

		logging.LogInfof("box [%s] not found", boxID)

		return ErrTreeNotFound
	}

	tree, err := filesys.LoadTree(boxID, unindexedTreePath, util.NewLute())
	if err != nil {
		logging.LogErrorf("load tree [%s] failed: %s", unindexedTreePath, err)
		return err
	}

	treenode.UpsertBlockTree(tree)
	sql.IndexTreeQueue(tree)
	logging.LogInfof("reindexed tree by filesystem [blockID=%s]", blockID)
	return nil
}

func loadParentTree(tree *parse.Tree) (ret *parse.Tree) {
	if nil == tree {
		return
	}

	boxDir := filepath.Join(util.DataDir, tree.Box)
	parentDir := path.Dir(tree.Path)
	if parentDir == boxDir || parentDir == "/" {
		return
	}

	luteEngine := lute.New()
	parentPath := parentDir + ".sy"
	ret, _ = filesys.LoadTree(tree.Box, parentPath, luteEngine)
	return
}

func findUnindexedTreePathInAllBoxes(id string) (ret string) {
	boxes := Conf.GetBoxes()
	for _, box := range boxes {
		root := filepath.Join(util.DataDir, box.ID)
		paths := search.FindAllMatchedPaths(root, []string{id})
		var rootIDs []string
		rootIDPaths := map[string]string{}
		for _, p := range paths {
			base := filepath.ToSlash(p)
			if !strings.HasSuffix(base, ".sy") {
				continue
			}
			if strings.Contains(base, "/.scribli/") {
				continue
			}
			rootID := util.GetTreeID(p)
			if !ast.IsNodeIDPattern(rootID) {
				continue
			}
			rootIDs = append(rootIDs, rootID)
			rootIDPaths[rootID] = p
		}

		result := treenode.ExistBlockTrees(rootIDs)
		for rootID, exist := range result {
			if !exist {
				return rootIDPaths[rootID]
			}
		}
	}
	return
}
