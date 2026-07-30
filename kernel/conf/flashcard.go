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

package conf

import (
	"bytes"
	"fmt"

	"github.com/icha-senpai/note/third_party/forks/github/open-spaced-repetition/go-fsrs/v3"
)

type Flashcard struct {
	NewCardLimit    int  `json:"newCardLimit"`
	ReviewCardLimit int  `json:"reviewCardLimit"`
	Mark            bool `json:"mark"`
	List            bool `json:"list"`
	SuperBlock      bool `json:"superBlock"`
	Heading         bool `json:"heading"`
	Deck            bool `json:"deck"`
	ReviewMode      int  `json:"reviewMode"`

	// Apply result optimized by FSRS optimizer
	RequestRetention float64 `json:"requestRetention"`
	MaximumInterval  int     `json:"maximumInterval"`
	Weights          string  `json:"weights"`
}

func NewFlashcard() *Flashcard {
	param := fsrs.DefaultParam()
	return &Flashcard{
		NewCardLimit:     20,
		ReviewCardLimit:  200,
		Mark:             true,
		List:             true,
		SuperBlock:       true,
		Heading:          true,
		Deck:             false,
		ReviewMode:       0,
		RequestRetention: param.RequestRetention,
		MaximumInterval:  int(param.MaximumInterval),
		Weights:          DefaultFSRSWeights(),
	}
}

func DefaultFSRSWeights() string {
	buf := bytes.Buffer{}
	defaultWs := fsrs.DefaultWeights()
	for i, w := range defaultWs {
		buf.WriteString(fmt.Sprintf("%v", w))
		if i < len(defaultWs)-1 {
			buf.WriteString(", ")
		}
	}
	return buf.String()
}
