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

package conf

type Publish struct {
	Enable bool       `json:"enable"`
	Port   uint16     `json:"port"`
	Auth   *BasicAuth `json:"auth"`
}

type BasicAuth struct {
	Enable   bool                `json:"enable"`
	Accounts []*BasicAuthAccount `json:"accounts"`
}

type BasicAuthAccount struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Memo     string `json:"memo"`
}

func NewPublish() *Publish {
	return &Publish{
		Enable: false,
		Port:   6808,
		Auth: &BasicAuth{
			Enable:   true,
			Accounts: []*BasicAuthAccount{},
		},
	}
}
