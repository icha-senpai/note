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

package extensions

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/88250/gulu"
	"github.com/88250/lute"
	"github.com/88250/lute/ast"
	"github.com/88250/lute/parse"
	"github.com/siyuan-note/logging"
	"github.com/siyuan-note/siyuan/kernel/util"
)

func getReadmeFileCandidates(readme LocaleStrings) []string {
	preferred := GetPreferredLocaleString(readme, "README.md")
	defaultName := "README.md"
	if v := strings.TrimSpace(readme["default"]); v != "" {
		defaultName = v
	}
	return gulu.Str.RemoveDuplicatedElem([]string{preferred, defaultName, "README.md"})
}

func getInstalledPackageREADME(installPath, linkBase string, readme LocaleStrings) (ret string) {
	candidates := getReadmeFileCandidates(readme)
	var errMsgs []string
	for _, name := range candidates {
		readmeData, readErr := os.ReadFile(filepath.Join(installPath, name))
		if readErr == nil {
			ret = renderPackageREADME(linkBase, readmeData)
			return
		}
		logging.LogWarnf("read installed %s failed: %s", name, readErr)
		errMsgs = append(errMsgs, fmt.Sprintf("File [%s] not found", name))
	}
	ret = strings.Join(errMsgs, "<br>")
	return
}

func renderPackageREADME(linkBase string, mdData []byte) (ret string) {
	mdData = bytes.TrimPrefix(mdData, []byte("\xef\xbb\xbf"))
	luteEngine := lute.New()
	luteEngine.SetSanitize(true)
	luteEngine.SetSoftBreak2HardBreak(false)
	luteEngine.SetCodeSyntaxHighlight(false)
	luteEngine.SetLinkBase(linkBase)

	tree := parse.Parse("", mdData, luteEngine.ParseOptions)
	normalizeNodesIAL(tree)
	ret = luteEngine.Tree2HTML(tree, luteEngine.RenderOptions, luteEngine.ParseOptions)
	ret = util.ConvertIframeToLink(ret)
	ret = util.LinkTarget(ret, linkBase)
	return
}

func normalizeNodesIAL(tree *parse.Tree) {
	if tree == nil || tree.Root == nil {
		return
	}

	ast.Walk(tree.Root, func(n *ast.Node, entering bool) ast.WalkStatus {
		if !entering {
			return ast.WalkContinue
		}
		if n.Type == ast.NodeCodeBlock {

			n.KramdownIAL = addClassToKramdownIAL(n.KramdownIAL, "code-block")
		}
		return ast.WalkContinue
	})
}

func addClassToKramdownIAL(ial [][]string, class string) [][]string {
	for i, attr := range ial {
		if len(attr) < 2 || attr[0] != "class" {
			continue
		}
		for item := range strings.FieldsSeq(attr[1]) {
			if item == class {
				return ial
			}
		}
		attr[1] = strings.TrimSpace(attr[1] + " " + class)
		ial[i] = attr
		return ial
	}
	return append(ial, []string{"class", class})
}
