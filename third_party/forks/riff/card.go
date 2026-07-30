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

import "time"

type Card interface {
	ID() string

	BlockID() string

	NextDues() map[Rating]time.Time

	SetNextDues(map[Rating]time.Time)

	SetDue(time.Time)

	GetLapses() int

	GetReps() int

	GetState() State

	GetLastReview() time.Time

	Clone() Card

	Impl() interface{}

	SetImpl(c interface{})
}

type BaseCard struct {
	CID   string
	BID   string
	NDues map[Rating]time.Time
}

func (card *BaseCard) NextDues() map[Rating]time.Time {
	return card.NDues
}

func (card *BaseCard) SetNextDues(dues map[Rating]time.Time) {
	card.NDues = dues
}

func (card *BaseCard) ID() string {
	return card.CID
}

func (card *BaseCard) BlockID() string {
	return card.BID
}
