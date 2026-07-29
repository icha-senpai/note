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

package av

type ViewGroup struct {
	Field     string      `json:"field"`
	Method    GroupMethod `json:"method"`
	Range     *GroupRange `json:"range,omitempty"`
	Order     GroupOrder  `json:"order"`
	HideEmpty bool        `json:"hideEmpty"`
}

type GroupMethod int

const (
	GroupMethodValue GroupMethod = iota
	GroupMethodRangeNum
	GroupMethodDateRelative
	GroupMethodDateDay
	GroupMethodDateWeek
	GroupMethodDateMonth
	GroupMethodDateYear
)

type GroupRange struct {
	NumStart float64 `json:"numStart"`
	NumEnd   float64 `json:"numEnd"`
	NumStep  float64 `json:"numStep"`
}

type GroupOrder int

const (
	GroupOrderAsc = iota
	GroupOrderDesc
	GroupOrderMan
	GroupOrderSelectOption
)
