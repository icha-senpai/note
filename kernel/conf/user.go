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

type User struct {
	UserId              string       `json:"userId"`
	UserName            string       `json:"userName"`
	UserAvatarURL       string       `json:"userAvatarURL"`
	UserHomeBImgURL     string       `json:"userHomeBImgURL"`
	UserTitles          []*UserTitle `json:"userTitles"`
	UserIntro           string       `json:"userIntro"`
	UserNickname        string       `json:"userNickname"`
	UserCreateTime      string       `json:"userCreateTime"`
	UserToken           string       `json:"userToken"`
	UserTokenExpireTime string       `json:"userTokenExpireTime"`
}

type UserTitle struct {
	Name string `json:"name"`
	Desc string `json:"desc"`
	Icon string `json:"icon"`
}
