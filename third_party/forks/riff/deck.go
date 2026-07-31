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
	"errors"
	"path/filepath"
	"sync"
	"time"

	"github.com/icha-senpai/note/third_party/forks/filelock"
	"github.com/icha-senpai/note/third_party/forks/logging"
	"github.com/vmihailenco/msgpack/v5"
)

type Deck struct {
	ID      string // ID
	Name    string
	Algo    Algo
	Desc    string
	Created int64
	Updated int64

	store Store
	lock  *sync.Mutex
}

func LoadDeck(saveDir, id string, requestRetention float64, maximumInterval int, weights string) (deck *Deck, err error) {
	created := time.Now().UnixMilli()
	deck = &Deck{
		ID:      id,
		Name:    id,
		Algo:    AlgoFSRS,
		Created: created,
		Updated: created,
		lock:    &sync.Mutex{},
	}

	dataPath := getDeckMsgpackPath(saveDir, id)
	if filelock.IsExist(dataPath) {
		var data []byte
		data, err = filelock.ReadFile(dataPath)
		if nil != err {
			logging.LogErrorf("load deck [%s] failed: %s", deck.Name, err)
			return
		}

		err = msgpack.Unmarshal(data, deck)
		if nil != err {
			logging.LogErrorf("load deck [%s] failed: %s", deck.Name, err)
			return
		}
	}

	var store Store
	switch deck.Algo {
	case AlgoFSRS:
		store = NewFSRSStore(deck.ID, saveDir, requestRetention, maximumInterval, weights)
		err = store.Load()
	default:
		err = errors.New("not supported yet")
		return
	}
	if nil != err {
		return
	}
	deck.store = store
	return
}

func (deck *Deck) AddCard(cardID, blockID string) {
	deck.lock.Lock()
	defer deck.lock.Unlock()

	card := deck.store.GetCard(cardID)
	if nil != card {
		return
	}

	deck.store.AddCard(cardID, blockID)
	deck.Updated = time.Now().UnixMilli()
}

func (deck *Deck) RemoveCard(cardID string) {
	deck.lock.Lock()
	defer deck.lock.Unlock()

	deck.store.RemoveCard(cardID)
	deck.Updated = time.Now().UnixMilli()
}

func (deck *Deck) SetCard(card Card) {
	deck.lock.Lock()
	defer deck.lock.Unlock()

	deck.store.SetCard(card)
}

func (deck *Deck) GetCard(cardID string) Card {
	deck.lock.Lock()
	defer deck.lock.Unlock()
	return deck.store.GetCard(cardID)
}

func (deck *Deck) GetCardsByBlockID(blockID string) (ret []Card) {
	deck.lock.Lock()
	defer deck.lock.Unlock()

	return deck.store.GetCardsByBlockID(blockID)
}

func (deck *Deck) GetCardsByBlockIDs(blockIDs []string) (ret []Card) {
	deck.lock.Lock()
	defer deck.lock.Unlock()

	return deck.store.GetCardsByBlockIDs(blockIDs)
}

func (deck *Deck) GetNewCardsByBlockIDs(blockIDs []string) (ret []Card) {
	deck.lock.Lock()
	defer deck.lock.Unlock()

	return deck.store.GetNewCardsByBlockIDs(blockIDs)
}

func (deck *Deck) GetDueCardsByBlockIDs(blockIDs []string) (ret []Card) {
	deck.lock.Lock()
	defer deck.lock.Unlock()

	return deck.store.GetDueCardsByBlockIDs(blockIDs)
}

func (deck *Deck) GetBlockIDs() (ret []string) {
	deck.lock.Lock()
	defer deck.lock.Unlock()

	return deck.store.GetBlockIDs()
}

func (deck *Deck) CountCards() int {
	deck.lock.Lock()
	defer deck.lock.Unlock()

	return deck.store.CountCards()
}

func (deck *Deck) Save() (err error) {
	deck.lock.Lock()
	defer deck.lock.Unlock()

	deck.Updated = time.Now().UnixMilli()
	err = deck.store.Save()
	if nil != err {
		logging.LogErrorf("save deck [%s] failed: %s", deck.Name, err)
		return
	}

	saveDir := deck.store.GetSaveDir()
	dataPath := getDeckMsgpackPath(saveDir, deck.ID)
	data, err := msgpack.Marshal(deck)
	if nil != err {
		logging.LogErrorf("save deck failed: %s", err)
		return
	}
	if err = filelock.WriteFile(dataPath, data); nil != err {
		logging.LogErrorf("save deck failed: %s", err)
		return
	}
	return
}

func (deck *Deck) SaveLog(log *Log) (err error) {
	deck.lock.Lock()
	defer deck.lock.Unlock()

	return deck.store.SaveLog(log)
}

func (deck *Deck) Review(cardID string, rating Rating) (ret *Log) {
	deck.lock.Lock()
	defer deck.lock.Unlock()

	ret = deck.store.Review(cardID, rating)
	deck.Updated = time.Now().UnixMilli()
	return
}

func (deck *Deck) Dues() (ret []Card) {
	deck.lock.Lock()
	defer deck.lock.Unlock()
	return deck.store.Dues()
}

func getDeckMsgpackPath(saveDir, id string) string {
	return filepath.Join(saveDir, id+".deck")
}
