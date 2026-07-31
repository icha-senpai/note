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
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/icha-senpai/note/kernel/search"
	"github.com/icha-senpai/note/kernel/sql"
	"github.com/icha-senpai/note/kernel/task"
	"github.com/icha-senpai/note/kernel/util"
	"github.com/icha-senpai/note/third_party/forks/epub"
	"github.com/icha-senpai/note/third_party/forks/eventbus"
	"github.com/icha-senpai/note/third_party/forks/filelock"
	"github.com/icha-senpai/note/third_party/forks/go-humanize"
	"github.com/icha-senpai/note/third_party/forks/gulu"
	"github.com/icha-senpai/note/third_party/forks/logging"
	"github.com/icha-senpai/note/third_party/forks/lute/ast"
)

type AssetContent struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Ext     string `json:"ext"`
	Path    string `json:"path"`
	Size    int64  `json:"size"`
	HSize   string `json:"hSize"`
	Updated int64  `json:"updated"`
	Content string `json:"content"`
}

func GetAssetContent(id, query string, queryMethod int) (ret *AssetContent) {
	if "" != query && (0 == queryMethod || 1 == queryMethod) {
		if 0 == queryMethod {
			query = stringQuery(query)
		}
	}
	if !ast.IsNodeIDPattern(id) {
		return
	}

	table := "asset_contents_fts_case_insensitive"
	filter := "id = ?"
	args := []any{id}
	if "" != query {
		filter += " AND `" + table + "` MATCH ?"
		args = append(args, buildAssetContentColumnFilter()+":("+query+")")
	}

	projections := "id, name, ext, path, size, updated, " +
		"highlight(" + table + ", 6, '" + search.SearchMarkLeft + "', '" + search.SearchMarkRight + "') AS content"
	stmt := "SELECT " + projections + " FROM " + table + " WHERE " + filter
	assetContents := sql.SelectAssetContentsRawStmtNoParseArgs(stmt, args, 1)
	results := fromSQLAssetContents(&assetContents)
	if 1 > len(results) {
		return
	}
	ret = results[0]
	ret.Content = strings.ReplaceAll(ret.Content, "\n", "<br>")
	return
}

//

func GetAssetContentByPath(path string) (ret *AssetContent) {
	path = strings.TrimSpace(path)
	if "" == path {
		return
	}

	table := "asset_contents_fts_case_insensitive"
	stmt := "SELECT id, name, ext, path, size, updated, content FROM " + table + " WHERE path = ? LIMIT 1"
	assetContents := sql.SelectAssetContentsRawStmtNoParseArgs(stmt, []any{path}, 1)
	results := fromSQLAssetContents(&assetContents)
	if 1 > len(results) {
		return
	}
	ret = results[0]
	return
}

//

func FullTextSearchAssetContent(query string, types map[string]bool, method, orderBy, page, pageSize int) (ret []*AssetContent, matchedAssetCount, pageCount int, err error) {
	query = strings.TrimSpace(query)
	orderByClause := buildAssetContentOrderBy(orderBy)
	switch method {
	case 1:
		filter, filterArgs := buildAssetContentTypeFilter(types)
		ret, matchedAssetCount = fullTextSearchAssetContentByQuerySyntax(query, filter, filterArgs, orderByClause, page, pageSize)
	case 2: // SQL
		ret, matchedAssetCount, err = searchAssetContentBySQL(query, page, pageSize)
		if err != nil {
			return
		}
	case 3:
		typeFilter, typeArgs := buildAssetContentTypeFilter(types)
		ret, matchedAssetCount = fullTextSearchAssetContentByRegexp(query, typeFilter, typeArgs, orderByClause, page, pageSize)
	default:
		filter, filterArgs := buildAssetContentTypeFilter(types)
		ret, matchedAssetCount = fullTextSearchAssetContentByKeyword(query, filter, filterArgs, orderByClause, page, pageSize)
	}
	pageCount = (matchedAssetCount + pageSize - 1) / pageSize

	if 1 > len(ret) {
		ret = []*AssetContent{}
	}
	return
}

func fullTextSearchAssetContentByQuerySyntax(query, typeFilter string, typeArgs []any, orderBy string, page, pageSize int) (ret []*AssetContent, matchedAssetCount int) {
	query = filterQueryInvisibleChars(query)
	return fullTextSearchAssetContentByFTS(query, typeFilter, typeArgs, orderBy, page, pageSize)
}

func fullTextSearchAssetContentByKeyword(query, typeFilter string, typeArgs []any, orderBy string, page, pageSize int) (ret []*AssetContent, matchedAssetCount int) {
	query = filterQueryInvisibleChars(query)
	query = stringQuery(query)
	return fullTextSearchAssetContentByFTS(query, typeFilter, typeArgs, orderBy, page, pageSize)
}

func fullTextSearchAssetContentByRegexp(exp, typeFilter string, typeArgs []any, orderBy string, page, pageSize int) (ret []*AssetContent, matchedAssetCount int) {
	exp = filterQueryInvisibleChars(exp)
	fieldFilter, args := assetContentFieldRegexp(exp)
	args = append(args, typeArgs...)
	stmt := "SELECT * FROM `asset_contents_fts_case_insensitive` WHERE " + fieldFilter + typeFilter
	stmt += " " + orderBy
	stmt += " LIMIT " + strconv.Itoa(pageSize) + " OFFSET " + strconv.Itoa((page-1)*pageSize)
	assetContents := sql.SelectAssetContentsRawStmtNoParseArgs(stmt, args, Conf.Search.Limit)
	ret = fromSQLAssetContents(&assetContents)
	if 1 > len(ret) {
		ret = []*AssetContent{}
	}

	matchedAssetCount = fullTextSearchAssetContentCountByRegexp(exp, typeFilter, typeArgs)
	return
}

func assetContentFieldRegexp(exp string) (clause string, args []any) {
	clause = "(name REGEXP ? OR content REGEXP ?)"
	args = []any{exp, exp}
	return
}

func fullTextSearchAssetContentCountByRegexp(exp, typeFilter string, typeArgs []any) (matchedAssetCount int) {
	table := "asset_contents_fts_case_insensitive"
	fieldFilter, args := assetContentFieldRegexp(exp)
	args = append(args, typeArgs...)
	stmt := "SELECT COUNT(path) AS `assets` FROM `" + table + "` WHERE " + fieldFilter + typeFilter
	result, _ := sql.QueryAssetContentNoLimitArgs(stmt, args...)
	if 1 > len(result) {
		return
	}
	matchedAssetCount = int(result[0]["assets"].(int64))
	return
}

func fullTextSearchAssetContentByFTS(query, typeFilter string, typeArgs []any, orderBy string, page, pageSize int) (ret []*AssetContent, matchedAssetCount int) {
	table := "asset_contents_fts_case_insensitive"
	projections := "id, name, ext, path, size, updated, " +
		"snippet(" + table + ", 6, '" + search.SearchMarkLeft + "', '" + search.SearchMarkRight + "', '...', 64) AS content"
	stmt := "SELECT " + projections + " FROM " + table + " WHERE `" + table + "` MATCH ?" + typeFilter
	stmt += " " + orderBy
	stmt += " LIMIT " + strconv.Itoa(pageSize) + " OFFSET " + strconv.Itoa((page-1)*pageSize)
	args := []any{buildAssetContentColumnFilter() + ":(" + query + ")"}
	args = append(args, typeArgs...)
	assetContents := sql.SelectAssetContentsRawStmtNoParseArgs(stmt, args, Conf.Search.Limit)
	ret = fromSQLAssetContents(&assetContents)
	if 1 > len(ret) {
		ret = []*AssetContent{}
	}

	matchedAssetCount = fullTextSearchAssetContentCount(query, typeFilter, typeArgs)
	return
}

func searchAssetContentBySQL(stmt string, page, pageSize int) (ret []*AssetContent, matchedAssetCount int, err error) {
	stmt = filterQueryInvisibleChars(stmt)
	stmt = strings.TrimSpace(stmt)
	if err = sql.CheckSingleStatement(stmt); err != nil {
		return
	}
	if err = sql.CheckAssetContentReadonlyStatement(stmt); err != nil {
		return
	}
	assetContents := sql.SelectAssetContentsRawStmt(stmt, page, pageSize)
	ret = fromSQLAssetContents(&assetContents)
	if 1 > len(ret) {
		ret = []*AssetContent{}
		return
	}

	stmt = strings.ToLower(stmt)
	stmt = strings.ReplaceAll(stmt, "select * ", "select COUNT(path) AS `assets` ")
	stmt = removeLimitClause(stmt)
	result, _ := sql.QueryAssetContentNoLimit(stmt)
	if 1 > len(result) {
		return
	}

	if assets, ok := result[0]["assets"].(int64); ok {
		matchedAssetCount = int(assets)
	}
	return
}

func fullTextSearchAssetContentCount(query, typeFilter string, typeArgs []any) (matchedAssetCount int) {
	query = filterQueryInvisibleChars(query)

	table := "asset_contents_fts_case_insensitive"
	stmt := "SELECT COUNT(path) AS `assets` FROM `" + table + "` WHERE `" + table + "` MATCH ?" + typeFilter
	args := []any{buildAssetContentColumnFilter() + ":(" + query + ")"}
	args = append(args, typeArgs...)
	result, _ := sql.QueryAssetContentNoLimitArgs(stmt, args...)
	if 1 > len(result) {
		return
	}
	matchedAssetCount = int(result[0]["assets"].(int64))
	return
}

func fromSQLAssetContents(assetContents *[]*sql.AssetContent) (ret []*AssetContent) {
	ret = []*AssetContent{}
	for _, assetContent := range *assetContents {
		ret = append(ret, fromSQLAssetContent(assetContent))
	}
	return
}

func fromSQLAssetContent(assetContent *sql.AssetContent) *AssetContent {
	content := util.EscapeHTML(assetContent.Content)
	if strings.Contains(content, search.SearchMarkLeft) {
		content = strings.ReplaceAll(content, search.SearchMarkLeft, "<mark>")
		content = strings.ReplaceAll(content, search.SearchMarkRight, "</mark>")
	}

	return &AssetContent{
		ID:      assetContent.ID,
		Name:    assetContent.Name,
		Ext:     assetContent.Ext,
		Path:    assetContent.Path,
		Size:    assetContent.Size,
		HSize:   humanize.BytesCustomCeil(uint64(assetContent.Size), 2),
		Updated: assetContent.Updated,
		Content: content,
	}
}

func buildAssetContentColumnFilter() string {
	return "{name content}"
}

func buildAssetContentTypeFilter(types map[string]bool) (clause string, args []any) {
	if 0 == len(types) {
		return
	}

	var enabledTypes []string
	for k, enabled := range types {
		if enabled {
			enabledTypes = append(enabledTypes, k)
		}
	}
	if 0 == len(enabledTypes) {
		clause = " AND 1 = 0"
		return
	}

	sort.Strings(enabledTypes)
	placeholders := strings.TrimSuffix(strings.Repeat("?, ", len(enabledTypes)), ", ")
	clause = " AND ext IN (" + placeholders + ")"
	for _, assetType := range enabledTypes {
		args = append(args, assetType)
	}
	return
}

func buildAssetContentOrderBy(orderBy int) string {
	switch orderBy {
	case 0:
		return "ORDER BY rank DESC"
	case 1:
		return "ORDER BY rank ASC"
	case 2:
		return "ORDER BY updated ASC"
	case 3:
		return "ORDER BY updated DESC"
	default:
		return "ORDER BY rank DESC"
	}
}

var assetContentSearcher = NewAssetsSearcher()

func removeIndexAssetContent(absPath string) {
	defer logging.Recover()

	assetsDir := util.GetDataAssetsAbsPath()
	p := "assets" + filepath.ToSlash(strings.TrimPrefix(absPath, assetsDir))
	sql.DeleteAssetContentsByPathQueue(p)
}

func indexAssetContent(absPath string) {
	defer logging.Recover()

	ext := filepath.Ext(absPath)
	parser := assetContentSearcher.GetParser(ext)
	if nil == parser {
		return
	}

	result := parser.Parse(absPath)
	if nil == result {
		return
	}

	info, err := os.Stat(absPath)
	if err != nil {
		logging.LogErrorf("stat [%s] failed: %s", absPath, err)
		return
	}

	assetsDir := util.GetDataAssetsAbsPath()
	p := "assets" + filepath.ToSlash(strings.TrimPrefix(absPath, assetsDir))

	assetContents := []*sql.AssetContent{
		{
			ID:      ast.NewNodeID(),
			Name:    util.RemoveID(filepath.Base(p)),
			Ext:     ext,
			Path:    p,
			Size:    info.Size(),
			Updated: info.ModTime().Unix(),
			Content: result.Content,
		},
	}

	sql.DeleteAssetContentsByPathQueue(p)
	sql.IndexAssetContentsQueue(assetContents)
}

func ReindexAssetContent() {
	task.AppendTask(task.AssetContentDatabaseIndexFull, fullReindexAssetContent)
	return
}

func fullReindexAssetContent() {
	util.PushMsg(Conf.Language(216), 7*1000)
	sql.InitAssetContentDatabase(true)

	assetContentSearcher.FullIndex()
	return
}

func init() {
	subscribeSQLAssetContentEvents()
}

func subscribeSQLAssetContentEvents() {
	eventbus.Subscribe(util.EvtSQLAssetContentRebuild, func() {
		ReindexAssetContent()
	})
}

type AssetsSearcher struct {
	parsers map[string]AssetParser
	lock    *sync.Mutex
}

func (searcher *AssetsSearcher) GetParser(ext string) AssetParser {
	searcher.lock.Lock()
	defer searcher.lock.Unlock()

	return searcher.parsers[strings.ToLower(ext)]
}

func (searcher *AssetsSearcher) FullIndex() {
	defer logging.Recover()

	assetsDir := util.GetDataAssetsAbsPath()
	if !gulu.File.IsDir(assetsDir) {
		return
	}

	var results []*AssetParseResult
	filelock.Walk(assetsDir, func(absPath string, d fs.DirEntry, err error) error {
		if err != nil {
			logging.LogErrorf("walk dir [%s] failed: %s", absPath, err)
			return err
		}

		if d.IsDir() {
			return nil
		}

		if IsEncryptedAssetPath(absPath) {
			return nil
		}

		ext := filepath.Ext(absPath)
		parser := searcher.GetParser(ext)
		if nil == parser {
			return nil
		}

		logging.LogInfof("parsing asset content [%s]", absPath)

		result := parser.Parse(absPath)
		if nil == result {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			logging.LogErrorf("stat file [%s] failed: %s", absPath, err)
			return nil
		}

		result.Path = "assets" + filepath.ToSlash(strings.TrimPrefix(absPath, assetsDir))
		result.Size = info.Size()
		result.Updated = info.ModTime().Unix()
		results = append(results, result)
		return nil
	})

	var assetContents []*sql.AssetContent
	for _, result := range results {
		assetContents = append(assetContents, &sql.AssetContent{
			ID:      ast.NewNodeID(),
			Name:    util.RemoveID(filepath.Base(result.Path)),
			Ext:     strings.ToLower(filepath.Ext(result.Path)),
			Path:    result.Path,
			Size:    result.Size,
			Updated: result.Updated,
			Content: result.Content,
		})
	}

	sql.IndexAssetContentsQueue(assetContents)
}

func NewAssetsSearcher() *AssetsSearcher {
	txtAssetParser := &TxtAssetParser{}
	return &AssetsSearcher{
		parsers: map[string]AssetParser{
			".txt":      txtAssetParser,
			".md":       txtAssetParser,
			".markdown": txtAssetParser,
			".json":     txtAssetParser,
			".log":      txtAssetParser,
			".sql":      txtAssetParser,
			".html":     txtAssetParser,
			".xml":      txtAssetParser,
			".java":     txtAssetParser,
			".h":        txtAssetParser,
			".c":        txtAssetParser,
			".cpp":      txtAssetParser,
			".go":       txtAssetParser,
			".rs":       txtAssetParser,
			".swift":    txtAssetParser,
			".kt":       txtAssetParser,
			".py":       txtAssetParser,
			".php":      txtAssetParser,
			".js":       txtAssetParser,
			".css":      txtAssetParser,
			".ts":       txtAssetParser,
			".sh":       txtAssetParser,
			".bat":      txtAssetParser,
			".cmd":      txtAssetParser,
			".ini":      txtAssetParser,
			".yaml":     txtAssetParser,
			".rst":      txtAssetParser,
			".adoc":     txtAssetParser,
			".textile":  txtAssetParser,
			".opml":     txtAssetParser,
			".org":      txtAssetParser,
			".wiki":     txtAssetParser,
			".cs":       txtAssetParser,
			".docx":     &DocxAssetParser{},
			".pptx":     &PptxAssetParser{},
			".xlsx":     &XlsxAssetParser{},
			".pdf":      &PdfAssetParser{},
			".epub":     &EpubAssetParser{},
		},

		lock: &sync.Mutex{},
	}
}

const (
	TxtAssetContentMaxSize = 1024 * 1024 * 4
	PDFAssetContentMaxPage = 1024
)

var (
	PDFAssetContentMaxSize uint64 = 1024 * 1024 * 128
)

type AssetParseResult struct {
	Path    string
	Size    int64
	Updated int64
	Content string
}

type AssetParser interface {
	Parse(absPath string) *AssetParseResult
}

type TxtAssetParser struct {
}

func (parser *TxtAssetParser) Parse(absPath string) (ret *AssetParseResult) {
	info, err := os.Stat(absPath)
	if err != nil {
		logging.LogErrorf("stat file [%s] failed: %s", absPath, err)
		return
	}

	if TxtAssetContentMaxSize < info.Size() {
		logging.LogWarnf("text asset [%s] is too large [%s]", absPath, humanize.BytesCustomCeil(uint64(info.Size()), 2))
		return
	}

	tmp := copyTempAsset(absPath)
	if "" == tmp {
		return
	}
	defer os.RemoveAll(tmp)

	data, err := os.ReadFile(tmp)
	if err != nil {
		logging.LogErrorf("read file [%s] failed: %s", absPath, err)
		return
	}

	if !utf8.Valid(data) {
		// Non-UTF-8 encoded text files are not included in asset file content searching
		logging.LogWarnf("text asset [%s] is not UTF-8 encoded", absPath)
		return
	}

	content := string(data)
	ret = &AssetParseResult{
		Content: content,
	}
	return
}

func normalizeNonTxtAssetContent(content string) (ret string) {
	ret = strings.Join(strings.Fields(content), " ")
	return
}

func copyTempAsset(absPath string) (ret string) {
	dir := filepath.Join(util.TempDir, "convert", "asset_content")
	if err := os.MkdirAll(dir, 0755); err != nil {
		logging.LogErrorf("mkdir [%s] failed: [%s]", dir, err)
		return
	}

	baseName := filepath.Base(absPath)
	if strings.HasPrefix(baseName, "~") {
		return
	}

	filelock.Lock(absPath)
	defer filelock.Unlock(absPath)

	ext := filepath.Ext(absPath)
	ret = filepath.Join(dir, gulu.Rand.String(7)+ext)
	if err := gulu.File.Copy(absPath, ret); err != nil {
		logging.LogErrorf("copy [src=%s, dest=%s] failed: %s", absPath, ret, err)
		return
	}
	return
}

type DocxAssetParser struct {
}

func (parser *DocxAssetParser) Parse(absPath string) (ret *AssetParseResult) {
	if !strings.HasSuffix(strings.ToLower(absPath), ".docx") {
		return
	}

	if !gulu.File.IsExist(absPath) {
		return
	}

	tmp := copyTempAsset(absPath)
	if "" == tmp {
		return
	}
	defer os.RemoveAll(tmp)

	data, err := extractDocxText(tmp)
	if err != nil {
		logging.LogErrorf("convert [%s] failed: [%s]", tmp, err)
		return
	}

	var content = normalizeNonTxtAssetContent(data)
	ret = &AssetParseResult{
		Content: content,
	}
	return
}

type PptxAssetParser struct {
}

func (parser *PptxAssetParser) Parse(absPath string) (ret *AssetParseResult) {
	if !strings.HasSuffix(strings.ToLower(absPath), ".pptx") {
		return
	}

	if !gulu.File.IsExist(absPath) {
		return
	}

	tmp := copyTempAsset(absPath)
	if "" == tmp {
		return
	}
	defer os.RemoveAll(tmp)

	data, err := extractPptxText(tmp)
	if err != nil {
		logging.LogErrorf("convert [%s] failed: [%s]", tmp, err)
		return
	}

	var content = normalizeNonTxtAssetContent(data)
	ret = &AssetParseResult{
		Content: content,
	}
	return
}

type XlsxAssetParser struct {
}

func (parser *XlsxAssetParser) Parse(absPath string) (ret *AssetParseResult) {
	if !strings.HasSuffix(strings.ToLower(absPath), ".xlsx") {
		return
	}

	if !gulu.File.IsExist(absPath) {
		return
	}

	tmp := copyTempAsset(absPath)
	if "" == tmp {
		return
	}
	defer os.RemoveAll(tmp)

	data, err := extractXlsxText(tmp)
	if err != nil {
		logging.LogErrorf("convert [%s] failed: [%s]", tmp, err)
		return
	}

	var content = normalizeNonTxtAssetContent(data)
	ret = &AssetParseResult{
		Content: content,
	}
	return
}

// PdfAssetParser parser factory product
type PdfAssetParser struct {
}

// Parse extracts searchable text from common text PDFs using Scribli's local parser.
func (parser *PdfAssetParser) Parse(absPath string) (ret *AssetParseResult) {
	if util.IsMobileContainer() {
		// PDF asset content searching is not supported on mobile platforms
		return
	}

	if !strings.HasSuffix(strings.ToLower(absPath), ".pdf") {
		return
	}

	if !gulu.File.IsExist(absPath) {
		return
	}

	tmp := copyTempAsset(absPath)
	if "" == tmp {
		return
	}
	defer os.RemoveAll(tmp)

	// PDF blob will be processed in-memory making sharing of PDF document data across worker goroutines possible
	pdfData, err := os.ReadFile(tmp)
	if err != nil {
		logging.LogErrorf("open [%s] failed: [%s]", tmp, err)
		return
	}

	if maxSizeVal := os.Getenv("SCRIBLI_PDF_ASSET_CONTENT_INDEX_MAX_SIZE"); "" != maxSizeVal {
		if maxSize, parseErr := strconv.ParseUint(maxSizeVal, 10, 64); nil == parseErr {
			if maxSize != PDFAssetContentMaxSize {
				PDFAssetContentMaxSize = maxSize
				logging.LogInfof("set PDF asset content index max size to [%s]", humanize.BytesCustomCeil(maxSize, 2))
			}
		} else {
			logging.LogWarnf("invalid env [SCRIBLI_PDF_ASSET_CONTENT_INDEX_MAX_SIZE]: [%s], parsing failed: %s", maxSizeVal, parseErr)
		}
	}

	if PDFAssetContentMaxSize < uint64(len(pdfData)) {
		// PDF files larger than 128MB are not included in asset file content searching
		logging.LogWarnf("ignore large PDF asset [%s] with [%s]", absPath, humanize.BytesCustomCeil(uint64(len(pdfData)), 2))
		return
	}

	pageCount := countPDFPages(pdfData)
	if PDFAssetContentMaxPage < pageCount {
		// PDF files longer than 1024 pages are not included in asset file content searching
		logging.LogWarnf("ignore large PDF asset [%s] with [%d] pages", absPath, pageCount)
		return
	}

	content, err := extractPDFText(pdfData)
	if err != nil {
		logging.LogWarnf("convert [%s] failed: [%s]", tmp, err)
		return
	}
	ret = &AssetParseResult{
		Content: normalizeNonTxtAssetContent(content),
	}
	return
}

type EpubAssetParser struct {
}

func (parser *EpubAssetParser) Parse(absPath string) (ret *AssetParseResult) {
	if !strings.HasSuffix(strings.ToLower(absPath), ".epub") {
		return
	}

	if !gulu.File.IsExist(absPath) {
		return
	}

	tmp := copyTempAsset(absPath)
	if "" == tmp {
		return
	}
	defer os.RemoveAll(tmp)

	f, err := os.Open(tmp)
	if err != nil {
		logging.LogErrorf("open [%s] failed: [%s]", tmp, err)
		return
	}
	defer f.Close()

	buf := bytes.Buffer{}
	if err = epub.ToTxt(tmp, &buf); err != nil {
		logging.LogErrorf("convert [%s] failed: [%s]", tmp, err)
		return
	}

	content := normalizeNonTxtAssetContent(buf.String())
	ret = &AssetParseResult{
		Content: content,
	}
	return
}
