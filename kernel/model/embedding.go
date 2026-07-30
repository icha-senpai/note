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
	"container/heap"
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/icha-senpai/note/third_party/forks/gulu"
	"github.com/icha-senpai/note/kernel/sql"
	"github.com/icha-senpai/note/kernel/task"
	"github.com/icha-senpai/note/kernel/util"
	"github.com/icha-senpai/note/third_party/forks/eventbus"
	"github.com/icha-senpai/note/third_party/forks/logging"
	ignore "github.com/icha-senpai/note/third_party/forks/github/sabhiram/go-gitignore"
)

const (
	embeddingBatchSize      = 10
	embeddingMaxConcurrency = 8
	embeddingMinTextLen     = 7
	embeddingMaxContentLen  = 12000
	embeddingVectorDim      = 4 // float32 = 4 bytes

	// delay = min(embeddingBackoffBase << (failCount-1), embeddingBackoffMax)
	embeddingBackoffBase  = 30
	embeddingBackoffMax   = 30 * 60
	embeddingMaxFailCount = 8

	embeddingIgnoredNone   = 0
	embeddingIgnoredByLen  = 1
	embeddingIgnoredByConf = 2
)

var (
	embeddingDirtyCh = make(chan string, 1024)
	embeddingTableOk bool

	embeddingIgnoreLoaded  bool
	embeddingIgnoreMatcher *ignore.GitIgnore
	embeddingIgnoreLock    sync.Mutex

	embeddingStop atomic.Bool

	embeddingErrNotified atomic.Bool

	embeddingIndexerRunning atomic.Bool
)

func checkEmbeddingTable() bool {
	_, err := sql.QueryNoLimit("SELECT COUNT(*) FROM block_embeddings")
	if err != nil {
		logging.LogWarnf("block_embeddings table not available, embedding indexer disabled: %s", err)
		return false
	}
	return true
}

func StartEmbeddingIndexer() {
	if !checkEmbeddingTable() || !isEmbeddingEnabled() {
		return
	}

	if !embeddingIndexerRunning.CompareAndSwap(false, true) {
		return
	}

	eventbus.Subscribe(eventbus.EvtEmbeddingDirty, func(id string) {
		select {
		case embeddingDirtyCh <- id:
		default:
		}
	})

	embeddingTableOk = true

	processPendingEmbeddings()

	for {
		select {
		case <-embeddingDirtyCh:
			processPendingEmbeddings()
		case <-time.After(30 * time.Second):
			processPendingEmbeddings()
		}
	}
}

func PrepareEmbeddingSearch() {
	if checkEmbeddingTable() && isEmbeddingEnabled() {
		embeddingTableOk = true
	}
}

type embeddingJob struct {
	texts  []string
	blocks []map[string]any
}

func processPendingEmbeddings() {
	if !isEmbeddingEnabled() {
		return
	}

	embeddingStop.Store(false)
	embeddingErrNotified.Store(false)

	workCh := make(chan embeddingJob, embeddingMaxConcurrency*2)

	var workersWg sync.WaitGroup
	for range embeddingMaxConcurrency {
		workersWg.Go(func() {
			for job := range workCh {
				if embeddingStop.Load() {

					recordFailedEmbedding(job.blocks, "round stopped due to earlier failure in this round")
					continue
				}
				doEmbedAndStore(job.texts, job.blocks)
			}
		})
	}

	go func() {
		defer close(workCh)
		for {
			if embeddingStop.Load() {
				return
			}

			now := time.Now().Unix()
			cutoff := now - int64(embeddingBackoffBase)
			results, err := sql.QueryNoLimitArgs(stmtPendingBlocks, embeddingMaxFailCount, cutoff)
			if err != nil {
				logging.LogErrorf("query pending embedding blocks failed: %s", err)
				return
			}

			if 1 > len(results) {
				return
			}

			var texts []string
			var blocks []map[string]any
			anySubmitted := false
			backoffSkipped := 0
			minRemaining := int64(embeddingBackoffMax)
			for _, row := range results {
				id, _ := row["id"].(string)
				rootID, _ := row["root_id"].(string)
				box, _ := row["box"].(string)
				path, _ := row["path"].(string)
				updated, _ := row["updated"].(string)
				content, _ := row["content"].(string)

				failCount, _ := row["fail_count"].(int64)
				lastTried, _ := row["last_tried"].(int64)
				if failCount > 0 {
					if failCount >= embeddingMaxFailCount {
						continue
					}
					required := int64(embeddingBackoffFor(int(failCount)) / time.Second)
					if elapsed := now - lastTried; elapsed < required {
						backoffSkipped++
						if remaining := required - elapsed; remaining < minRemaining {
							minRemaining = remaining
						}
						continue
					}
				}

				matcher := getEmbeddingIgnoreMatcher()
				if nil != matcher && matcher.MatchesPath("/"+box+path) {

					sql.Exec("INSERT OR IGNORE INTO block_embeddings (id, root_id, box, path, embedding, model, content_len, updated, fail_count, last_tried, ignored_type) VALUES (?, ?, ?, ?, ?, ?, ?, ?, 0, 0, ?)",
						id, rootID, box, path, []byte{}, embeddingModel(), 0, updated, embeddingIgnoredByConf)
					continue
				}
				if len(content) < embeddingMinTextLen || len(content) > embeddingMaxContentLen {

					sql.Exec("INSERT OR IGNORE INTO block_embeddings (id, root_id, box, path, embedding, model, content_len, updated, fail_count, last_tried, ignored_type) VALUES (?, ?, ?, ?, ?, ?, ?, ?, 0, 0, ?)",
						id, rootID, box, path, []byte{}, embeddingModel(), 0, updated, embeddingIgnoredByLen)
					continue
				}
				row["plain_text"] = content
				texts = append(texts, content)
				blocks = append(blocks, row)

				if len(texts) >= embeddingBatchSize {
					workCh <- embeddingJob{texts: texts, blocks: blocks}
					anySubmitted = true
					texts = nil
					blocks = nil
				}
			}
			if len(texts) > 0 {
				workCh <- embeddingJob{texts: texts, blocks: blocks}
				anySubmitted = true
			}

			if !anySubmitted && backoffSkipped > 0 {
				wait := max(time.Duration(minRemaining)*time.Second, time.Second)

				for wait > 0 && !embeddingStop.Load() {
					step := min(wait, time.Second)
					time.Sleep(step)
					wait -= step
				}
			}
		}
	}()

	workersWg.Wait()
}

//

const stmtPendingBlocks = "SELECT b.id, b.root_id, b.box, b.path, b.content, b.updated, " +
	"COALESCE(e.fail_count, 0) AS fail_count, COALESCE(e.last_tried, 0) AS last_tried " +
	"FROM blocks b " +
	"LEFT JOIN block_embeddings e ON b.id = e.id " +
	"WHERE e.id IS NULL " +
	"OR (e.fail_count > 0 AND e.fail_count < ? AND e.last_tried < ?) " +
	"ORDER BY fail_count ASC, b.updated DESC LIMIT 100"

func embeddingBackoffFor(failCount int) time.Duration {
	if failCount < 1 {
		return time.Duration(embeddingBackoffBase) * time.Second
	}
	shift := min(failCount-1,

		20)
	d := embeddingBackoffBase << uint(shift)
	if d > embeddingBackoffMax || d < 0 {
		return time.Duration(embeddingBackoffMax) * time.Second
	}
	return time.Duration(d) * time.Second
}

func encodeVector(vec []float32) []byte {
	buf := make([]byte, len(vec)*embeddingVectorDim)
	for i, v := range vec {
		binary.LittleEndian.PutUint32(buf[i*embeddingVectorDim:], math.Float32bits(v))
	}
	return buf
}

func decodeVector(b []byte) []float32 {
	if len(b) == 0 {
		return nil
	}
	return unsafe.Slice((*float32)(unsafe.Pointer(&b[0])), len(b)/embeddingVectorDim)
}

func recordFailedEmbedding(blocks []map[string]any, reason string) {
	embeddingStop.Store(true)
	logging.LogErrorf("create embeddings failed (%s), stop this round", reason)

	if embeddingErrNotified.CompareAndSwap(false, true) {
		util.PushErrMsg("Embedding request failed, indexing paused. Please check AI embedding config.", 5000)
	}

	now := time.Now().Unix()
	for _, row := range blocks {
		id, _ := row["id"].(string)
		rootID, _ := row["root_id"].(string)
		box, _ := row["box"].(string)
		path, _ := row["path"].(string)
		updated, _ := row["updated"].(string)

		sql.Exec("INSERT OR IGNORE INTO block_embeddings (id, root_id, box, path, embedding, model, content_len, updated, fail_count, last_tried, ignored_type) VALUES (?, ?, ?, ?, ?, ?, ?, ?, 0, 0, 0)",
			id, rootID, box, path, []byte{}, embeddingModel(), 0, updated)
		sql.Exec("UPDATE block_embeddings SET fail_count = fail_count + 1, last_tried = ?, embedding = ?, model = ?, content_len = 0, ignored_type = 0 WHERE id = ?",
			now, []byte{}, embeddingModel(), id)
	}
}

func doEmbedAndStore(texts []string, blocks []map[string]any) {
	vectors, err := util.BatchGetEmbeddings(texts, embeddingKey(), embeddingBaseURL(), embeddingModel(), embeddingDimensions(), embeddingTimeout())
	if err != nil {

		recordFailedEmbedding(blocks, err.Error())
		return
	}

	if len(vectors) != len(blocks) {
		recordFailedEmbedding(blocks, fmt.Sprintf("count mismatch: requested %d but got %d", len(blocks), len(vectors)))
		return
	}

	for i, row := range blocks {
		id, _ := row["id"].(string)
		rootID, _ := row["root_id"].(string)
		box, _ := row["box"].(string)
		path, _ := row["path"].(string)
		updated, _ := row["updated"].(string)
		plainText, _ := row["plain_text"].(string)

		buf := encodeVector(vectors[i])

		err = sql.Exec("INSERT OR REPLACE INTO block_embeddings (id, root_id, box, path, embedding, model, content_len, updated, fail_count, last_tried, ignored_type) VALUES (?, ?, ?, ?, ?, ?, ?, ?, 0, 0, 0)",
			id, rootID, box, path, buf, embeddingModel(), len(plainText), updated)
		if err != nil {
			logging.LogErrorf("store embedding failed for block [%s]: %s", id, err)
		}
	}
}

func getEmbeddingIgnoreMatcher() *ignore.GitIgnore {
	if embeddingIgnoreLoaded {
		return embeddingIgnoreMatcher
	}

	embeddingIgnoreLock.Lock()
	defer embeddingIgnoreLock.Unlock()

	if embeddingIgnoreLoaded {
		return embeddingIgnoreMatcher
	}

	embeddingIgnorePath := filepath.Join(util.DataDir, ".scribli", "embeddingignore")
	if !gulu.File.IsExist(embeddingIgnorePath) {
		return nil
	}

	data, err := os.ReadFile(embeddingIgnorePath)
	if err != nil {
		logging.LogErrorf("read embeddingignore [%s] failed: %s", embeddingIgnorePath, err)
		return nil
	}

	dataStr := string(data)
	dataStr = strings.ReplaceAll(dataStr, "\r\n", "\n")
	lines := strings.Split(dataStr, "\n")

	embeddingIgnoreMatcher = ignore.CompileIgnoreLines(lines...)
	embeddingIgnoreLoaded = true
	return embeddingIgnoreMatcher
}

func cosineSimilarity(a, b []float32) float32 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}

	var dotProduct, normA, normB float64
	for i := range a {
		dotProduct += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return float32(dotProduct / (math.Sqrt(normA) * math.Sqrt(normB)))
}

type scoredBlock struct {
	id    string
	score float32
}

type scoredHeap []scoredBlock

func (h scoredHeap) Len() int           { return len(h) }
func (h scoredHeap) Less(i, j int) bool { return h[i].score < h[j].score } // min-heap
func (h scoredHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *scoredHeap) Push(x any) {
	*h = append(*h, x.(scoredBlock))
}
func (h *scoredHeap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[:n-1]
	return x
}

func SemanticSearchBlock(query string, boxes, paths []string, types, subTypes map[string]bool, page, pageSize int) (blocks []*Block, matchedBlockCount, matchedRootCount, pageCount int) {
	blocks = []*Block{}

	if !embeddingTableOk || !isEmbeddingEnabled() || "" == query {
		return
	}

	vectors, err := util.BatchGetEmbeddings([]string{query}, embeddingKey(), embeddingBaseURL(), embeddingModel(), embeddingDimensions(), embeddingTimeout())
	if err != nil || 1 > len(vectors) {
		logging.LogErrorf("get query embedding failed")
		return
	}
	queryVec := vectors[0]

	boxFilter, boxArgs := buildBoxesFilter(boxes, "be.")
	pathFilter, pathArgs := buildPathsFilter(paths, "be.")
	boxDocFilter, boxDocArgs := buildRootIDExclusionFilter(hiddenBoxDocRootIDs(), "b.")
	typeFilter := buildTypeFilter(types, subTypes, "b.")
	hasFilter := 0 < len(boxes) || 0 < len(paths) || 0 < len(types) || "" != boxDocFilter
	hasTypeFilter := 0 < len(types)

	numWorkers := max(runtime.GOMAXPROCS(0), 1)

	topK := page * pageSize
	if isRerankEnabled() {
		topK = rerankCandidateCount()
	}
	h := &scoredHeap{}
	heap.Init(h)

	scanSize := 4096
	cursor := int64(0)

	for {
		var q string
		var args []any
		if hasFilter {
			q = fmt.Sprintf("SELECT be.rowid, be.id, be.embedding FROM block_embeddings be JOIN blocks b ON be.id = b.id WHERE be.embedding IS NOT NULL AND length(be.embedding) > 0 AND be.rowid > %d", cursor)
			if hasTypeFilter {
				q += " AND " + typeFilter
			}
			q += boxFilter + pathFilter + boxDocFilter

			args = append(append(append([]any{}, boxArgs...), pathArgs...), boxDocArgs...)
			q += fmt.Sprintf(" ORDER BY be.rowid LIMIT %d", scanSize)
		} else {
			q = fmt.Sprintf("SELECT rowid, id, embedding FROM block_embeddings WHERE embedding IS NOT NULL AND length(embedding) > 0 AND rowid > %d ORDER BY rowid LIMIT %d", cursor, scanSize)
		}
		rows, qErr := sql.QueryNoLimitArgs(q, args...)
		if qErr != nil {
			logging.LogErrorf("query embeddings for search failed: %s", qErr)
			break
		}
		if 1 > len(rows) {
			break
		}

		rawCursor, _ := rows[len(rows)-1]["rowid"].(int64)
		if rawCursor > cursor {
			cursor = rawCursor
		}

		chunkSize := (len(rows) + numWorkers - 1) / numWorkers
		scoredCh := make(chan []scoredBlock, numWorkers)
		var wg sync.WaitGroup

		for w := 0; w < numWorkers; w++ {
			start := w * chunkSize
			end := min(start+chunkSize, len(rows))
			if start >= end {
				continue
			}

			wg.Add(1)
			go func(chunk []map[string]any) {
				defer wg.Done()
				local := make([]scoredBlock, 0, len(chunk))
				for _, row := range chunk {
					embRaw := row["embedding"].([]byte)
					if len(embRaw) == 0 {
						continue
					}
					buf := make([]byte, len(embRaw))
					copy(buf, embRaw)
					vec := decodeVector(buf)
					score := cosineSimilarity(queryVec, vec)
					id, _ := row["id"].(string)
					local = append(local, scoredBlock{id: id, score: score})
				}
				scoredCh <- local
			}(rows[start:end])
		}

		wg.Wait()
		close(scoredCh)

		for ch := range scoredCh {
			for _, s := range ch {
				if h.Len() < topK {
					heap.Push(h, s)
				} else if s.score > (*h)[0].score {
					heap.Pop(h)
					heap.Push(h, s)
				}
			}
		}
	}

	matchedBlockCount = h.Len()
	if 1 > matchedBlockCount {
		pageCount = 0
		return
	}

	result := make([]scoredBlock, h.Len())
	for i := len(result) - 1; i >= 0; i-- {
		result[i] = heap.Pop(h).(scoredBlock)
	}

	var candidateIDs []string
	for _, s := range result {
		candidateIDs = append(candidateIDs, s.id)
	}

	sqlBlocks := sql.GetBlocks(candidateIDs)

	sqlBlocks = rerankSqlBlocks(query, sqlBlocks)

	offset := (page - 1) * pageSize
	if offset >= len(sqlBlocks) {
		pageCount = (matchedBlockCount + pageSize - 1) / pageSize
		return
	}

	end := min(offset+pageSize, len(sqlBlocks))

	rootIDSet := map[string]bool{}
	for i := offset; i < end; i++ {
		b := sqlBlocks[i]
		rootIDSet[b.RootID] = true
		blocks = append(blocks, fromSQLBlock(b, "", 36))
	}
	matchedRootCount = len(rootIDSet)
	pageCount = (matchedBlockCount + pageSize - 1) / pageSize

	return
}

func isEmbeddingEnabled() bool {
	return nil != Conf.AI.Embedding && Conf.AI.Embedding.Enabled && len(Conf.AI.Embedding.APIKey) > 0
}

func rerankSqlBlocks(query string, sqlBlocks []*sql.Block) []*sql.Block {
	if !isRerankEnabled() || len(sqlBlocks) < 2 {
		return sqlBlocks
	}

	documents := make([]string, len(sqlBlocks))
	for i, b := range sqlBlocks {
		documents[i] = b.Content
	}

	indices, _, err := util.Rerank(query, documents, rerankKey(), rerankEndpoint(), rerankModel(), 0, rerankTimeout())
	if nil != err {
		logging.LogErrorf("rerank failed, fallback to vector similarity order: %s", err)
		return sqlBlocks
	}
	if len(indices) != len(sqlBlocks) {

		logging.LogErrorf("rerank returned %d indices for %d documents, fallback", len(indices), len(sqlBlocks))
		return sqlBlocks
	}

	seen := make(map[int]bool, len(indices))
	for _, idx := range indices {
		if seen[idx] {
			logging.LogErrorf("rerank returned duplicate index %d, fallback", idx)
			return sqlBlocks
		}
		seen[idx] = true
	}

	reranked := make([]*sql.Block, len(indices))
	for i, idx := range indices {
		reranked[i] = sqlBlocks[idx]
	}
	return reranked
}

func ReindexEmbedding() {
	task.AppendTask(task.DatabaseIndexEmbeddingFull, fullReindexEmbedding)
}

func fullReindexEmbedding() {
	if !isEmbeddingEnabled() {
		logging.LogWarnf("embedding not enabled, skip reindex")
		return
	}
	if !checkEmbeddingTable() {
		logging.LogWarnf("block_embeddings table not available, skip reindex")
		return
	}
	if err := sql.Exec("DELETE FROM block_embeddings"); err != nil {
		logging.LogErrorf("clear block_embeddings failed: %s", err)
		return
	}
	logging.LogInfof("embedding vectors cleared, indexer will re-embed all blocks")

	if !embeddingIndexerRunning.Load() {
		go StartEmbeddingIndexer()
	} else {
		eventbus.Publish(eventbus.EvtEmbeddingDirty, "")
	}
}

func RetryFailedEmbedding() {
	task.AppendTask(task.DatabaseIndexEmbeddingRetryFailed, retryFailedEmbedding)
}

func retryFailedEmbedding() {
	if !isEmbeddingEnabled() {
		logging.LogWarnf("embedding not enabled, skip retry failed")
		return
	}
	if !checkEmbeddingTable() {
		logging.LogWarnf("block_embeddings table not available, skip retry failed")
		return
	}
	if err := sql.Exec("DELETE FROM block_embeddings WHERE fail_count > 0"); err != nil {
		logging.LogErrorf("delete failed embedding rows failed: %s", err)
		return
	}
	logging.LogInfof("failed embedding rows cleared, indexer will retry these blocks")

	if embeddingIndexerRunning.Load() {
		eventbus.Publish(eventbus.EvtEmbeddingDirty, "")
	} else {
		go StartEmbeddingIndexer()
	}
}

type EmbeddingStat struct {
	Total           int  `json:"total"`
	Indexed         int  `json:"indexed"`
	Pending         int  `json:"pending"`
	Failed          int  `json:"failed"`
	IgnoredByLen    int  `json:"ignoredByLen"`
	IgnoredByConfig int  `json:"ignoredByConfig"`
	Enabled         bool `json:"enabled"`
}

func GetEmbeddingStat() (ret *EmbeddingStat) {
	ret = &EmbeddingStat{Enabled: isEmbeddingEnabled()}
	if !checkEmbeddingTable() {
		return
	}

	rows, err := sql.QueryNoLimit("SELECT COUNT(*) AS total, SUM(CASE WHEN e.id IS NULL THEN 1 ELSE 0 END) AS pending FROM blocks b LEFT JOIN block_embeddings e ON b.id = e.id")
	if err != nil || 1 > len(rows) {
		logging.LogErrorf("query embedding total/pending stat failed: %s", err)
		return
	}
	if total, ok := rows[0]["total"].(int64); ok {
		ret.Total = int(total)
	}
	if pending, ok := rows[0]["pending"].(int64); ok {
		ret.Pending = int(pending)
	}

	rows, err = sql.QueryNoLimit("SELECT COUNT(*) AS c FROM block_embeddings WHERE length(embedding) > 0")
	if err == nil && 0 < len(rows) {
		if c, ok := rows[0]["c"].(int64); ok {
			ret.Indexed = int(c)
		}
	}

	rows, err = sql.QueryNoLimit("SELECT COUNT(*) AS c FROM block_embeddings WHERE fail_count > 0")
	if err == nil && 0 < len(rows) {
		if c, ok := rows[0]["c"].(int64); ok {
			ret.Failed = int(c)
		}
	}

	rows, err = sql.QueryNoLimit("SELECT SUM(CASE WHEN ignored_type = 1 THEN 1 ELSE 0 END) AS by_len, SUM(CASE WHEN ignored_type = 2 THEN 1 ELSE 0 END) AS by_conf FROM block_embeddings WHERE ignored_type > 0")
	if err == nil && 0 < len(rows) {
		if byLen, ok := rows[0]["by_len"].(int64); ok {
			ret.IgnoredByLen = int(byLen)
		}
		if byConf, ok := rows[0]["by_conf"].(int64); ok {
			ret.IgnoredByConfig = int(byConf)
		}
	}
	return
}

func embeddingKey() string {
	if nil != Conf.AI.Embedding && Conf.AI.Embedding.Enabled && "" != Conf.AI.Embedding.APIKey {
		return Conf.AI.Embedding.APIKey
	}
	if v := os.Getenv("SCRIBLI_OPENAI_EMBEDDING_API_KEY"); "" != v {
		return v
	}
	return ""
}

func embeddingBaseURL() string {
	if nil != Conf.AI.Embedding && Conf.AI.Embedding.Enabled && "" != Conf.AI.Embedding.BaseURL {
		return Conf.AI.Embedding.BaseURL
	}
	if v := os.Getenv("SCRIBLI_OPENAI_EMBEDDING_BASE_URL"); "" != v {
		return v
	}
	return ""
}

func embeddingTimeout() int {
	if nil != Conf.AI.Embedding && Conf.AI.Embedding.Enabled && 0 < Conf.AI.Embedding.Timeout {
		return Conf.AI.Embedding.Timeout
	}
	return 30
}

func embeddingDimensions() int {
	if nil != Conf.AI.Embedding && Conf.AI.Embedding.Enabled && 0 < Conf.AI.Embedding.Dimensions {
		return Conf.AI.Embedding.Dimensions
	}
	return 0
}

func embeddingModel() string {
	if nil != Conf.AI.Embedding && Conf.AI.Embedding.Enabled && "" != Conf.AI.Embedding.Name {
		return Conf.AI.Embedding.Name
	}
	if v := os.Getenv("SCRIBLI_OPENAI_EMBEDDING_MODEL"); "" != v {
		return v
	}
	return ""
}
