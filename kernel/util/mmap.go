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

package util

import (
	"errors"
	"fmt"
	"os"

	mmap "github.com/icha-senpai/note/third_party/forks/github/edsrzf/mmap-go"
	"github.com/icha-senpai/note/third_party/forks/filelock"
	"github.com/icha-senpai/note/third_party/forks/logging"
)

//

//

func WriteFileByMmap(filePath string, data []byte) (err error) {
	f, err := filelock.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return
	}
	defer filelock.CloseFile(f)

	if err = f.Truncate(int64(len(data))); err != nil {
		msg := fmt.Sprintf("truncate file [%s] failed: %s", filePath, err)
		logging.LogError(msg)
		err = errors.New(msg)
		return
	}

	m, err := mmap.Map(f, mmap.RDWR, 0)
	if err != nil {
		msg := fmt.Sprintf("map file [%s] failed: %s", filePath, err)
		logging.LogError(msg)
		err = errors.New(msg)
		return
	}
	defer m.Unmap()

	copy(m, data)
	if err = m.Flush(); err != nil {
		msg := fmt.Sprintf("flush data [%s] failed: %s", filePath, err)
		logging.LogError(msg)
		err = errors.New(msg)
		return
	}
	return
}
