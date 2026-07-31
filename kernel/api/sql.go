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
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/icha-senpai/note/kernel/model"
	"github.com/icha-senpai/note/kernel/sql"
	"github.com/icha-senpai/note/kernel/util"
	"github.com/icha-senpai/note/third_party/forks/gulu"
)

func flushTransaction(c *gin.Context) {
	// Add internal kernel API `/api/sqlite/flushTransaction`
	ret := gulu.Ret.NewResult()
	defer c.JSON(http.StatusOK, ret)

	model.FlushTxQueue()
	sql.FlushQueue()
}

func SQL(c *gin.Context) {
	ret := gulu.Ret.NewResult()
	defer c.JSON(http.StatusOK, ret)

	arg, ok := util.JsonArg(c, ret)
	if !ok {
		return
	}

	var stmt, mode string
	if !util.ParseJsonArgs(arg, ret,
		util.BindJsonArg("stmt", &stmt, true, true),
		util.BindJsonArg("mode", &mode, false, false),
	) {
		return
	}

	switch mode {
	case "":

		if err := sql.CheckSingleStatement(stmt); err != nil {
			ret.Code = -1
			ret.Msg = err.Error()
			return
		}
	case "readonly":

		if err := sql.CheckSingleStatement(stmt); err != nil {
			ret.Code = -1
			ret.Msg = err.Error()
			return
		}
		if err := sql.CheckReadonlyStatement(stmt); err != nil {
			ret.Code = -1
			ret.Msg = err.Error()
			return
		}
	case "multiple":

	default:

		ret.Code = -1
		ret.Msg = "unknown [mode]"
		return
	}

	result, err := sql.Query(stmt, model.Conf.Search.Limit)
	if err != nil {
		ret.Code = 1
		ret.Msg = err.Error()
		return
	}

	ret.Data = result
}
