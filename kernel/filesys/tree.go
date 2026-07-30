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

package filesys

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/icha-senpai/note/third_party/forks/gulu"
	"github.com/icha-senpai/note/third_party/forks/lute"
	"github.com/icha-senpai/note/third_party/forks/lute/ast"
	"github.com/icha-senpai/note/third_party/forks/lute/html"
	"github.com/icha-senpai/note/third_party/forks/lute/parse"
	"github.com/icha-senpai/note/third_party/forks/lute/render"
	"github.com/icha-senpai/note/kernel/cache"
	"github.com/icha-senpai/note/kernel/treenode"
	"github.com/icha-senpai/note/kernel/util"
	"github.com/icha-senpai/note/third_party/forks/dataparser"
	"github.com/icha-senpai/note/third_party/forks/filelock"
	"github.com/icha-senpai/note/third_party/forks/logging"
	jsoniter "github.com/icha-senpai/note/third_party/forks/github/json-iterator/go"
	"github.com/icha-senpai/note/third_party/forks/github/panjf2000/ants/v2"
)

func LoadTrees(ids []string) (ret map[string]*parse.Tree) {
	ret = map[string]*parse.Tree{}
	if 1 > len(ids) {
		return ret
	}

	bts := treenode.GetBlockTrees(ids)

	foundSet := map[string]bool{}
	for id := range bts {
		foundSet[id] = true
	}
	var missing []string
	for _, id := range ids {
		if !foundSet[id] {
			missing = append(missing, id)
		}
	}
	if len(missing) > 0 {
		for _, encBoxID := range treenode.GetOpenedEncryptedBoxIDs() {
			if len(missing) == 0 {
				break
			}
			encBTs := treenode.GetBlockTreesInBox(missing, encBoxID)
			maps.Copy(bts, encBTs)
			var stillMissing []string
			for _, id := range missing {
				if _, found := encBTs[id]; !found {
					stillMissing = append(stillMissing, id)
				}
			}
			missing = stillMissing
		}
	}

	luteEngine := util.NewLute()
	var boxIDs []string
	var paths []string
	blockIDs := map[string][]string{}
	for _, bt := range bts {
		boxIDs = append(boxIDs, bt.BoxID)
		paths = append(paths, bt.Path)
		if _, ok := blockIDs[bt.RootID]; !ok {
			blockIDs[bt.RootID] = []string{}
		}
		blockIDs[bt.RootID] = append(blockIDs[bt.RootID], bt.ID)
	}

	trees, errs := batchLoadTrees(boxIDs, paths, luteEngine)
	for i := range trees {
		tree := trees[i]
		err := errs[i]
		if err != nil || tree == nil {
			logging.LogErrorf("load tree failed: %s", err)
			continue
		}

		bIDs := blockIDs[tree.Root.ID]
		for _, bID := range bIDs {
			ret[bID] = tree
		}
	}
	return
}

func batchLoadTrees(boxIDs, paths []string, luteEngine *lute.Lute) (ret []*parse.Tree, errs []error) {
	waitGroup := sync.WaitGroup{}
	lock := sync.Mutex{}
	poolSize := min(runtime.NumCPU(), 8)

	p, _ := ants.NewPoolWithFunc(poolSize, func(arg any) {
		defer waitGroup.Done()

		i := arg.(int)
		boxID := boxIDs[i]
		path := paths[i]
		tree, err := LoadTree(boxID, path, luteEngine)
		lock.Lock()
		ret = append(ret, tree)
		errs = append(errs, err)
		lock.Unlock()
	})
	loaded := map[string]bool{}
	for i := range paths {
		if loaded[boxIDs[i]+paths[i]] {
			continue
		}

		loaded[boxIDs[i]+paths[i]] = true

		waitGroup.Add(1)
		p.Invoke(i)
	}
	waitGroup.Wait()
	p.Release()
	return
}

func ValidateBoxRelativePath(boxID, p string) (string, error) {
	p = filepath.ToSlash(p)

	origP := p

	p = strings.TrimPrefix(p, "/")

	if p == "" {
		return p, nil
	}
	if strings.HasPrefix(p, "..") || strings.Contains(p, "/../") || strings.HasSuffix(p, "/..") || p == ".." || p == "." {
		return "", fmt.Errorf("path [%s] must not contain '..'", origP)
	}
	resolved := filepath.Join(util.DataDir, boxID, origP)
	boxRoot := filepath.Join(util.DataDir, boxID)
	if !gulu.File.IsSubPath(boxRoot, resolved) {
		return "", fmt.Errorf("path [%s] escapes box directory", origP)
	}
	return p, nil
}

func LoadTreeWithFix(boxID, p string, luteEngine *lute.Lute) (ret *parse.Tree, needFix bool, err error) {
	if _, err = ValidateBoxRelativePath(boxID, p); err != nil {
		logging.LogErrorf("invalid tree path [%s] for box [%s]: %s", p, boxID, err)
		return
	}
	rootID := util.GetTreeID(p)
	if raw, ok := cache.GetTreeDataInBox(rootID, boxID); ok {
		ret, err = LoadTreeByData(raw, boxID, p, luteEngine)
		return
	}

	filePath := filepath.Join(util.DataDir, boxID, p)
	data, err := filelock.ReadFile(filePath)
	if nil != err {
		logging.LogErrorf("load tree [%s] failed: %s", p, err)
		return
	}

	if data, err = decryptData(boxID, p, data); nil != err {
		logging.LogErrorf("decrypt tree [%s] failed: %s", p, err)
		return
	}

	data, needFix, err = fixTreeJSONData(boxID, p, data, luteEngine)
	if nil != err {
		return
	}

	ret, err = LoadTreeByData(data, boxID, p, luteEngine)
	if nil == err {
		cache.SetTreeDataInBox(rootID, boxID, data)
	}
	return
}

func LoadTree(boxID, p string, luteEngine *lute.Lute) (ret *parse.Tree, err error) {
	ret, _, err = LoadTreeWithFix(boxID, p, luteEngine)
	return
}

func LoadTreeByData(data []byte, boxID, p string, luteEngine *lute.Lute) (ret *parse.Tree, err error) {
	ret, err = parseJSON2Tree(boxID, p, data, luteEngine)
	if nil != err {
		logging.LogErrorf("parse tree [%s] failed: %s", p, err)
		return
	}
	ret.Path = p
	ret.Root.Path = p

	hPath := "/" + strings.TrimPrefix(filepath.ToSlash(p), "/")
	parts := strings.Split(hPath, "/")
	if len(parts) < 2 {
		logging.LogErrorf("parse tree [%s] failed: invalid path", p)
		err = errors.New("invalid path")
		return
	}

	parts = parts[1 : len(parts)-1]
	if 1 > len(parts) {
		ret.HPath = "/" + ret.Root.IALAttr("title")
		ret.Hash = treenode.NodeHash(ret.Root, ret, luteEngine)
		return
	}

	hPathBuilder := bytes.Buffer{}
	hPathBuilder.WriteString("/")
	for i := range parts {
		var parentAbsPath string
		if 0 < i {
			parentAbsPath = strings.Join(parts[:i+1], "/")
		} else {
			parentAbsPath = parts[0]
		}
		parentAbsPath += ".sy"
		parentPath := parentAbsPath
		parentAbsPath = filepath.Join(util.DataDir, boxID, parentAbsPath)

		parentDocIAL := DocIAL(parentAbsPath)
		if 1 > len(parentDocIAL) {

			parentTree := treenode.NewTree(boxID, parentPath, hPathBuilder.String()+"Untitled", "Untitled")
			if _, writeErr := WriteTree(parentTree); nil != writeErr {
				logging.LogErrorf("rebuild parent tree [%s] failed: %s", parentAbsPath, writeErr)
			} else {
				logging.LogInfof("rebuilt parent tree [%s]", parentAbsPath)
				treenode.UpsertBlockTree(parentTree)
			}
			hPathBuilder.WriteString("Untitled/")
			continue
		}

		title := parentDocIAL["title"]
		if "" == title {
			title = "Untitled"
		}
		hPathBuilder.WriteString(title)
		hPathBuilder.WriteString("/")
	}
	hPathBuilder.WriteString(ret.Root.IALAttr("title"))
	ret.HPath = hPathBuilder.String()
	ret.Hash = treenode.NodeHash(ret.Root, ret, luteEngine)
	return
}

func DocIAL(absPath string) (ret map[string]string) {

	boxID := docIALBoxID(absPath)
	if boxID != "" && DEKProvider != nil {
		if dek, err := DEKProvider(boxID); err == nil && dek != nil {

			raw, readErr := filelock.ReadFile(absPath)
			if readErr != nil {
				logging.LogErrorf("read file [%s] failed: %s", absPath, readErr)
				return nil
			}
			relPath := filepath.ToSlash(strings.TrimPrefix(absPath, filepath.Join(util.DataDir, boxID)+string(os.PathSeparator)))
			plain, decErr := decryptData(boxID, relPath, raw)
			if decErr != nil {

				logging.LogErrorf("decrypt doc [%s] for IAL failed: %s", absPath, decErr)
				return map[string]string{}
			}
			iter := jsoniter.Parse(jsoniter.ConfigCompatibleWithStandardLibrary, bytes.NewReader(plain), 512)
			for field := iter.ReadObject(); field != ""; field = iter.ReadObject() {
				if field == "Properties" {
					iter.ReadVal(&ret)
					break
				} else {
					iter.Skip()
				}
			}
			for k, v := range ret {
				ret[k] = html.UnescapeAttrVal(v)
			}
			return
		}
	}

	filelock.Lock(absPath)
	file, err := os.Open(absPath)
	if err != nil {
		logging.LogErrorf("open file [%s] failed: %s", absPath, err)
		filelock.Unlock(absPath)
		return nil
	}

	iter := jsoniter.Parse(jsoniter.ConfigCompatibleWithStandardLibrary, file, 512)
	for field := iter.ReadObject(); field != ""; field = iter.ReadObject() {
		if field == "Properties" {
			iter.ReadVal(&ret)
			break
		} else {
			iter.Skip()
		}
	}
	file.Close()
	filelock.Unlock(absPath)

	for k, v := range ret {
		ret[k] = html.UnescapeAttrVal(v)
	}
	return
}

func TreeSize(tree *parse.Tree) (size uint64) {
	luteEngine := util.NewLute()
	renderer := render.NewJSONRenderer(tree, luteEngine.RenderOptions, luteEngine.ParseOptions)
	return uint64(len(renderer.Render()))
}

func WriteTree(tree *parse.Tree) (size uint64, err error) {
	data, filePath, err := prepareWriteTree(tree)
	if err != nil {
		return
	}

	encData, encErr := encryptData(tree.Box, tree.Path, data)
	if encErr != nil {
		err = encErr
		return
	}

	if cachedData, ok := cache.GetTreeDataInBox(tree.ID, tree.Box); ok {
		if len(cachedData) == len(data) && bytes.Equal(cachedData, data) {
			return
		}
	} else {

		if diskData, readErr := filelock.ReadFile(filePath); nil == readErr {
			decDisk, decErr := decryptData(tree.Box, tree.Path, diskData)
			if decErr == nil && len(decDisk) == len(data) && bytes.Equal(decDisk, data) {
				cache.SetTreeDataInBox(tree.ID, tree.Box, data)
				return
			}
		}
	}

	if err = util.WriteFileByMmap(filePath, encData); nil != err {
		if err = writeTreeByWriteFile(filePath, encData); nil != err {
			return
		}
	}

	if util.ExceedLargeFileWarningSize(len(data)) {
		msg := fmt.Sprintf(util.Langs[util.Lang][268], tree.Root.IALAttr("title")+" "+filepath.Base(filePath), util.LargeFileWarningSize)
		util.PushErrMsg(msg, 7000)
	}

	cache.SetTreeDataInBox(tree.ID, tree.Box, data)
	afterWriteTree(tree)
	size = uint64(len(data))
	return
}

func writeTreeByWriteFile(filePath string, data []byte) (err error) {
	if err = filelock.WriteFile(filePath, data); err != nil {
		msg := fmt.Sprintf("write data [%s] failed: %s", filePath, err)
		logging.LogErrorf("%s", msg)
		err = errors.New(msg)
		return
	}
	return
}

func prepareWriteTree(tree *parse.Tree) (data []byte, filePath string, err error) {
	luteEngine := util.NewLute()

	if nil == tree.Root.FirstChild {
		newP := treenode.NewParagraph("")
		tree.Root.AppendChild(newP)
		tree.Root.SetIALAttr("updated", util.TimeFromID(newP.ID))
		treenode.UpsertBlockTree(tree)
	}

	treenode.UpgradeSpec(tree)

	if _, err = ValidateBoxRelativePath(tree.Box, tree.Path); err != nil {
		return
	}

	filePath = filepath.Join(util.DataDir, tree.Box, tree.Path)
	tree.Root.SetIALAttr("type", "doc")
	renderer := render.NewJSONRenderer(tree, luteEngine.RenderOptions, luteEngine.ParseOptions)
	data = renderer.Render()
	data, _ = removeUnescapedUnicodeNull(data)
	if !util.UseSingleLineSave {
		buf := bytes.Buffer{}
		buf.Grow(1024 * 1024 * 2)
		if err = json.Indent(&buf, data, "", "\t"); err != nil {
			logging.LogErrorf("json indent failed: %s", err)
			return
		}
		data = buf.Bytes()
	}

	if err = os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return
	}
	return
}

func removeUnescapedUnicodeNull(data []byte) (ret []byte, needFix bool) {
	patLen := 6 // len(`\u0000`)
	n := len(data)
	if n < patLen {
		return data, false
	}
	if !bytes.Contains(data, []byte(`\u0000`)) {
		return data, false
	}

	dst := make([]byte, 0, n)
	i := 0
	for i < n {
		from := i
		j := bytes.IndexByte(data[i:], '\\')
		if j < 0 {
			dst = append(dst, data[from:]...)
			break
		}
		i += j
		dst = append(dst, data[from:i]...)

		if i+patLen <= n &&
			data[i+1] == 'u' &&
			data[i+2] == '0' &&
			data[i+3] == '0' &&
			data[i+4] == '0' &&
			data[i+5] == '0' {

			backslashes := 0
			for k := i - 1; k >= 0 && data[k] == '\\'; k-- {
				backslashes++
			}

			if backslashes%2 == 0 {
				i += patLen
				continue
			}
		}

		dst = append(dst, data[i])
		i++
	}
	return dst, len(dst) != n
}

func afterWriteTree(tree *parse.Tree) {
	docIAL := parse.IAL2Map(tree.Root.KramdownIAL)
	cache.PutDocIALInBox(tree.Path, tree.Box, docIAL)
}

func fixTreeJSONData(boxID, p string, jsonData []byte, luteEngine *lute.Lute) (data []byte, needFix bool, err error) {
	jsonData, needFix = removeUnescapedUnicodeNull(jsonData)
	ret, parseNeedFix, err := dataparser.ParseJSON(jsonData, luteEngine.ParseOptions)
	if parseNeedFix {
		needFix = true
	}
	if err != nil {
		logging.LogErrorf("parse json [%s] to tree failed: %s", boxID+p, err)
		err = fmt.Errorf("parse json [%s] to tree failed: %w", boxID+p, err)
		return
	}

	ret.Box = boxID
	ret.Path = p

	if err = treenode.CheckSpec(ret); errors.Is(err, treenode.ErrSpecTooNew) {
		return
	}

	if treenode.UpgradeSpec(ret) {
		needFix = true
	}

	if escapeAttributeValues(ret) {
		needFix = true
	}

	if pathID := util.GetTreeID(p); pathID != ret.Root.ID {
		if encryptedBox(boxID) {

			err = fmt.Errorf("encrypted .sy [%s]: base id [%s] != root id [%s]", p, pathID, ret.Root.ID)
			logging.LogErrorf("%s", err)
			return
		}
		needFix = true
		logging.LogInfof("reset tree id from [%s] to [%s]", ret.Root.ID, pathID)
		ret.Root.ID = pathID
		ret.ID = pathID
		ret.Root.SetIALAttr("id", ret.ID)
	}

	if !needFix {
		return jsonData, false, nil
	}

	renderer := render.NewJSONRenderer(ret, luteEngine.RenderOptions, luteEngine.ParseOptions)
	data = renderer.Render()

	if !util.UseSingleLineSave {
		buf := bytes.Buffer{}
		buf.Grow(1024 * 1024 * 2)
		if err = json.Indent(&buf, data, "", "\t"); err != nil {
			return
		}
		data = buf.Bytes()
	}

	filePath := filepath.Join(util.DataDir, ret.Box, ret.Path)
	if err = os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return
	}

	encData, encErr := encryptData(ret.Box, ret.Path, data)
	if encErr != nil {
		err = encErr
		return
	}
	if err = filelock.WriteFile(filePath, encData); err != nil {
		logging.LogErrorf("write data [%s] failed: %s", filePath, err)
	}
	return
}

func parseJSON2Tree(boxID, p string, jsonData []byte, luteEngine *lute.Lute) (ret *parse.Tree, err error) {
	ret, _, err = dataparser.ParseJSON(jsonData, luteEngine.ParseOptions)
	if err != nil {
		logging.LogErrorf("parse json [%s] to tree failed: %s", boxID+p, err)
		err = fmt.Errorf("parse json [%s] to tree failed: %w", boxID+p, err)
		return
	}

	ret.Box = boxID
	ret.Path = p

	if err = treenode.CheckSpec(ret); errors.Is(err, treenode.ErrSpecTooNew) {
		return
	}
	return
}

func escapeAttributeValues(tree *parse.Tree) (hasEscaped bool) {
	if nil == tree || nil == tree.Root {
		return false
	}

	ast.Walk(tree.Root, func(n *ast.Node, entering bool) ast.WalkStatus {
		if !entering || !n.IsBlock() || "" == n.ID || 0 == len(n.KramdownIAL) {
			return ast.WalkContinue
		}

		if escaped := escapeNodeAttributeValues(n); escaped {
			hasEscaped = true
		}
		return ast.WalkContinue
	})
	return hasEscaped
}

func escapeNodeAttributeValues(node *ast.Node) (escaped bool) {
	if nil == node || 0 == len(node.KramdownIAL) {
		return false
	}

	for _, kv := range node.KramdownIAL {

		canonical := html.EscapeAttrVal(html.UnescapeAttrVal(kv[1]))
		if canonical != kv[1] {
			kv[1] = canonical
			escaped = true
		}
	}
	return
}
