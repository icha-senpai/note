// Riff - Spaced repetition.
// Copyright (c) 2022-present, Scribli
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

package riff

import (
	"math/rand"
	"path/filepath"
	"sync"
	"time"
)

type Store interface {
	AddCard(id, blockID string) Card

	GetCard(id string) Card

	SetCard(card Card)

	RemoveCard(id string) Card

	GetCardsByBlockID(blockID string) []Card

	GetCardsByBlockIDs(blockIDs []string) []Card

	GetNewCardsByBlockIDs(blockIDs []string) []Card

	GetDueCardsByBlockIDs(blockIDs []string) []Card

	GetBlockIDs() []string

	CountCards() int

	Review(id string, rating Rating) (ret *Log)

	Dues() []Card

	ID() string

	Algo() Algo

	Load() (err error)

	Save() error

	SaveLog(log *Log) error

	GetSaveDir() string
}

type BaseStore struct {
	id      string
	algo    Algo
	saveDir string
	lock    *sync.Mutex
}

func NewBaseStore(id string, algo Algo, saveDir string) *BaseStore {
	return &BaseStore{
		id:      id,
		algo:    algo,
		saveDir: saveDir,
		lock:    &sync.Mutex{},
	}
}

func (store *BaseStore) ID() string {
	return store.id
}

func (store *BaseStore) Algo() Algo {
	return store.algo
}

func (store *BaseStore) GetSaveDir() string {
	return store.saveDir
}

func (store *BaseStore) getMsgPackPath() string {
	return filepath.Join(store.saveDir, store.id+".cards")
}

type Rating int8

const (
	Again Rating = iota + 1
	Hard
	Good
	Easy
)

type Algo string

const (
	AlgoFSRS Algo = "fsrs"
	AlgoSM2  Algo = "sm2"
)

type State int8

const (
	New State = iota
	Learning
	Review
	Relearning
)

func newID() string {
	now := time.Now()
	return now.Format("20060102150405") + "-" + randStr(7)
}

func randStr(length int) string {
	letter := []rune("abcdefghijklmnopqrstuvwxyz0123456789")
	b := make([]rune, length)
	for i := range b {
		b[i] = letter[rand.Intn(len(letter))]
	}
	return string(b)
}
