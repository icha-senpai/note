// Scribli - Refactor your thinking
// Copyright (c) 2020-present, Scribli
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

package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/icha-senpai/note/kernel/model"
	"github.com/icha-senpai/note/kernel/sql"
	"github.com/icha-senpai/note/kernel/util"
	"github.com/icha-senpai/note/third_party/forks/gulu"
)

func getBookmarkLabels(c *gin.Context) {
	ret := gulu.Ret.NewResult()
	defer c.JSON(http.StatusOK, ret)

	ret.Data = model.BookmarkLabels()
}

func batchGetBlockAttrs(c *gin.Context) {
	ret := gulu.Ret.NewResult()
	defer c.JSON(http.StatusOK, ret)

	arg, ok := util.JsonArg(c, ret)
	if !ok {
		return
	}

	ids := arg["ids"].([]any)
	var idList []string
	for _, id := range ids {
		idList = append(idList, id.(string))
	}

	ret.Data = sql.BatchGetBlockAttrs(idList)
}

func getBlockAttrs(c *gin.Context) {
	ret := gulu.Ret.NewResult()
	defer c.JSON(http.StatusOK, ret)

	arg, ok := util.JsonArg(c, ret)
	if !ok {
		return
	}

	id := arg["id"].(string)
	if util.InvalidIDPattern(id, ret) {
		return
	}

	ret.Data = sql.GetBlockAttrs(id)
}

func setBlockAttrs(c *gin.Context) {
	ret := gulu.Ret.NewResult()
	defer c.JSON(http.StatusOK, ret)

	arg, ok := util.JsonArg(c, ret)
	if !ok {
		return
	}

	id := arg["id"].(string)
	if util.InvalidIDPattern(id, ret) {
		return
	}

	attrs := arg["attrs"].(map[string]any)
	nameValues := map[string]string{}
	for name, value := range attrs {
		if nil == value {
			nameValues[name] = ""
		} else {
			strValue, ok := value.(string)
			if !ok {
				ret.Code = -1
				ret.Msg = fmt.Sprintf("the value of attr [%s] must be a string", name)
				return
			}
			nameValues[name] = strValue
		}
	}
	err := model.SetBlockAttrs(id, nameValues)
	if err != nil {
		ret.Code = -1
		ret.Msg = err.Error()
		return
	}
}

func batchSetBlockAttrs(c *gin.Context) {
	ret := gulu.Ret.NewResult()
	defer c.JSON(http.StatusOK, ret)

	arg, ok := util.JsonArg(c, ret)
	if !ok {
		return
	}

	blockAttrsArg := arg["blockAttrs"].([]any)
	var blockAttrs []map[string]any
	for _, blockAttrArg := range blockAttrsArg {
		blockAttr := blockAttrArg.(map[string]any)
		id := blockAttr["id"].(string)
		if util.InvalidIDPattern(id, ret) {
			return
		}

		attrs := blockAttr["attrs"].(map[string]any)
		nameValues := map[string]string{}
		for name, value := range attrs {
			if nil == value {
				nameValues[name] = ""
			} else {
				strValue, ok := value.(string)
				if !ok {
					ret.Code = -1
					ret.Msg = fmt.Sprintf("the value of attr [%s] must be a string", name)
					return
				}
				nameValues[name] = strValue
			}
		}

		blockAttrs = append(blockAttrs, map[string]any{
			"id":    id,
			"attrs": nameValues,
		})
	}

	err := model.BatchSetBlockAttrs(blockAttrs)
	if err != nil {
		ret.Code = -1
		ret.Msg = err.Error()
		return
	}
}
