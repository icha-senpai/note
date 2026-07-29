// SiYuan - Refactor your thinking
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
	"fmt"
	"regexp"
	"sync"
	"sync/atomic"
	"time"

	"github.com/88250/gulu"
	"github.com/88250/lute/ast"
	"github.com/siyuan-note/siyuan/kernel/treenode"
)

type UndoEntry struct {
	id             string
	doOperations   []*Operation
	undoOperations []*Operation
	timestamp      int64
	mutatedRootIDs []string
}

func (e *UndoEntry) DoOperationsForReplay() []*Operation {
	return cloneOperations(e.doOperations)
}

func (e *UndoEntry) UndoOperationsForReplay() []*Operation {
	return cloneOperations(e.undoOperations)
}

func (e *UndoEntry) MutatedRootIDs() []string {
	if nil == e.mutatedRootIDs {
		return nil
	}
	ret := make([]string, len(e.mutatedRootIDs))
	copy(ret, e.mutatedRootIDs)
	return ret
}

type undoStack struct {
	undoStack []*UndoEntry
	redoStack []*UndoEntry
	hasUndo   bool
}

type UndoLog struct {
	mu     sync.Mutex
	stacks map[string]*undoStack
	max    int
}

var GlobalUndoLog = newUndoLog(64)

var undoEntrySeq uint64

func newUndoLog(max int) *UndoLog {
	return &UndoLog{
		stacks: map[string]*undoStack{},
		max:    max,
	}
}

func newUndoEntryID() string {
	seq := atomic.AddUint64(&undoEntrySeq, 1)
	return fmt.Sprintf("undo-%d-%d", time.Now().UnixNano(), seq)
}

func (l *UndoLog) stack(rootID string) *undoStack {
	return l.stacks[rootID]
}

func (l *UndoLog) stackOrCreate(rootID string) *undoStack {
	s := l.stacks[rootID]
	if nil == s {
		s = &undoStack{}
		l.stacks[rootID] = s
	}
	return s
}

func (l *UndoLog) Record(tx *Transaction) {
	if !tx.fromAPI || 0 == len(tx.UndoOperations) || tx.isReplay {
		return
	}

	rootIDs := tx.GetMutatedRootIDs()
	if 0 == len(rootIDs) {

		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	entry := &UndoEntry{
		id:             newUndoEntryID(),
		doOperations:   cloneOperations(tx.DoOperations),
		undoOperations: cloneOperations(tx.UndoOperations),
		timestamp:      time.Now().UnixMilli(),
		mutatedRootIDs: rootIDs,
	}

	for _, rootID := range rootIDs {
		s := l.stackOrCreate(rootID)
		s.undoStack = append(s.undoStack, entry)
		if s.hasUndo {
			s.redoStack = nil
			s.hasUndo = false
		}
		if l.max < len(s.undoStack) {
			s.undoStack = s.undoStack[len(s.undoStack)-l.max:]
		}
	}
}

func (l *UndoLog) Peek(rootID string) *UndoEntry {
	l.mu.Lock()
	defer l.mu.Unlock()

	s := l.stack(rootID)
	if nil == s || 0 == len(s.undoStack) {
		return nil
	}
	return s.undoStack[len(s.undoStack)-1]
}

func (l *UndoLog) Undo(rootID string) *UndoEntry {
	l.mu.Lock()
	defer l.mu.Unlock()

	s := l.stack(rootID)
	if nil == s || 0 == len(s.undoStack) {
		return nil
	}

	entry := s.undoStack[len(s.undoStack)-1]
	s.undoStack = s.undoStack[:len(s.undoStack)-1]

	s.redoStack = append(s.redoStack, entry)
	if l.max < len(s.redoStack) {
		s.redoStack = s.redoStack[len(s.redoStack)-l.max:]
	}
	s.hasUndo = true
	return entry
}

func (l *UndoLog) UndoCommit(entry *UndoEntry, rootID string) {
	if nil == entry {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	for _, r := range entry.mutatedRootIDs {
		if r == rootID {
			continue
		}
		l.removeEntry(r, entry.id)
	}
}

func (l *UndoLog) UndoRollback(entry *UndoEntry, rootID string) {
	if nil == entry {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	s := l.stack(rootID)
	if nil == s {
		return
	}

	if 0 < len(s.redoStack) && s.redoStack[len(s.redoStack)-1].id == entry.id {
		s.redoStack = s.redoStack[:len(s.redoStack)-1]
	}

	s.undoStack = append(s.undoStack, entry)
	s.hasUndo = false
}

func (l *UndoLog) Redo(rootID string) *UndoEntry {
	l.mu.Lock()
	defer l.mu.Unlock()

	s := l.stack(rootID)
	if nil == s || 0 == len(s.redoStack) {
		return nil
	}

	entry := s.redoStack[len(s.redoStack)-1]
	s.redoStack = s.redoStack[:len(s.redoStack)-1]
	s.undoStack = append(s.undoStack, entry)
	return entry
}

func (l *UndoLog) RedoCommit(entry *UndoEntry, rootID string) {
	if nil == entry {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	for _, r := range entry.mutatedRootIDs {
		if r == rootID {
			continue
		}
		rs := l.stackOrCreate(r)
		rs.undoStack = append(rs.undoStack, entry)
		if l.max < len(rs.undoStack) {
			rs.undoStack = rs.undoStack[len(rs.undoStack)-l.max:]
		}
	}
}

func (l *UndoLog) RedoRollback(entry *UndoEntry, rootID string) {
	if nil == entry {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	s := l.stack(rootID)
	if nil == s {
		return
	}

	if 0 < len(s.undoStack) && s.undoStack[len(s.undoStack)-1].id == entry.id {
		s.undoStack = s.undoStack[:len(s.undoStack)-1]
	}

	s.redoStack = append(s.redoStack, entry)
	if l.max < len(s.redoStack) {
		s.redoStack = s.redoStack[len(s.redoStack)-l.max:]
	}
}

func (l *UndoLog) State(rootID string) (canUndo, canRedo bool, peekMutatedRootIDs []string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	s := l.stack(rootID)
	if nil == s {
		return
	}
	canUndo = 0 < len(s.undoStack)
	canRedo = 0 < len(s.redoStack)
	if canUndo {
		top := s.undoStack[len(s.undoStack)-1]
		peekMutatedRootIDs = append(peekMutatedRootIDs, top.mutatedRootIDs...)
		peekMutatedRootIDs = gulu.Str.RemoveDuplicatedElem(peekMutatedRootIDs)
	}
	return
}

func (l *UndoLog) Clear(rootID string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if "" == rootID {
		l.stacks = map[string]*undoStack{}
		return
	}

	s := l.stacks[rootID]
	if nil == s {
		return
	}

	linkedIDs := map[string]bool{}
	for _, e := range s.undoStack {
		for _, r := range e.mutatedRootIDs {
			if r != rootID {
				linkedIDs[e.id] = true
			}
		}
	}
	for _, e := range s.redoStack {
		for _, r := range e.mutatedRootIDs {
			if r != rootID {
				linkedIDs[e.id] = true
			}
		}
	}
	delete(l.stacks, rootID)
	for otherID, other := range l.stacks {
		for id := range linkedIDs {
			other.undoStack = removeEntryByID(other.undoStack, id)
			other.redoStack = removeEntryByID(other.redoStack, id)
		}
		_ = otherID
	}
}

func (l *UndoLog) removeEntry(rootID, id string) {
	s := l.stacks[rootID]
	if nil == s {
		return
	}
	s.undoStack = removeEntryByID(s.undoStack, id)
}

func removeEntryByID(stack []*UndoEntry, id string) []*UndoEntry {
	for i, e := range stack {
		if e.id == id {
			return append(stack[:i], stack[i+1:]...)
		}
	}
	return stack
}

func cloneOperations(ops []*Operation) []*Operation {
	if nil == ops {
		return nil
	}
	ret := make([]*Operation, len(ops))
	for i, op := range ops {
		cloned := *op
		ret[i] = &cloned
	}
	return ret
}

var dataNodeIDPattern = regexp.MustCompile(`data-node-id="([^"]+)"`)
var refcountAttrPattern = regexp.MustCompile(`\s*refcount="[^"]*"`)
var refcountDivPattern = regexp.MustCompile(`<div class="protyle-attr--refcount[^"]*"[^>]*>.*?</div>`)

func ResolveReplayDuplicateIds(tx *Transaction) {
	if nil == tx || !tx.isReplay {
		return
	}

	ids := map[string]struct{}{}
	collect := func(ops []*Operation) {
		for _, op := range ops {
			if "insert" != op.Action {
				continue
			}
			if "" != op.ID && ast.IsNodeIDPattern(op.ID) {
				ids[op.ID] = struct{}{}
			}
			data, ok := op.Data.(string)
			if !ok {
				continue
			}
			for _, m := range dataNodeIDPattern.FindAllStringSubmatch(data, -1) {
				if ast.IsNodeIDPattern(m[1]) {
					ids[m[1]] = struct{}{}
				}
			}
		}
	}
	collect(tx.DoOperations)

	if 0 == len(ids) {
		return
	}

	idList := make([]string, 0, len(ids))
	for id := range ids {
		idList = append(idList, id)
	}
	exist := treenode.ExistBlockTrees(idList)

	replacements := map[string]string{}
	for _, id := range idList {
		if exist[id] {
			replacements[id] = ast.NewNodeID()
		}
	}
	if 0 == len(replacements) {
		return
	}

	apply := func(ops []*Operation) {
		for _, op := range ops {

			_, idReplaced := replacements[op.ID]

			// https://github.com/siyuan-note/siyuan/issues/18012
			if "insert" == op.Action {
				if newID, ok := replacements[op.ID]; ok {
					op.ID = newID
				}
			}
			if newID, ok := replacements[op.ParentID]; ok {
				op.ParentID = newID
			}
			if newID, ok := replacements[op.PreviousID]; ok {
				op.PreviousID = newID
			}
			if newID, ok := replacements[op.NextID]; ok {
				op.NextID = newID
			}
			data, ok := op.Data.(string)
			if !ok {
				continue
			}
			for oldID, newID := range replacements {
				data = dataNodeIDPattern.ReplaceAllStringFunc(data, func(match string) string {
					if sub := dataNodeIDPattern.FindStringSubmatch(match); len(sub) > 1 && sub[1] == oldID {
						return `data-node-id="` + newID + `"`
					}
					return match
				})
			}

			if idReplaced {
				data = refcountDivPattern.ReplaceAllString(data, "")
				data = refcountAttrPattern.ReplaceAllString(data, "")
			}
			op.Data = data
		}
	}
	apply(tx.DoOperations)
}
