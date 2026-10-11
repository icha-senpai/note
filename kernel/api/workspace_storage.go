package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/icha-senpai/note/kernel/util"
	"github.com/icha-senpai/note/third_party/forks/gulu"
)

func getWorkspaceStorage(c *gin.Context) {
	ret := gulu.Ret.NewResult()
	defer c.JSON(http.StatusOK, ret)
	stats, err := util.GetWorkspaceStorage(c.Request.Context(), util.WorkspaceDir)
	if err != nil {
		ret.Code = -1
		ret.Msg = err.Error()
		return
	}
	ret.Data = map[string]any{"directories": stats.Directories, "assetsSize": stats.AssetsSize, "totalSize": stats.TotalSize, "calculatedAt": stats.CalculatedAt}
}
