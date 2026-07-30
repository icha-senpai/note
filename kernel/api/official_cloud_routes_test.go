package api

import (
	"strings"
	"testing"

	"github.com/icha-senpai/note/third_party/forks/github/gin-gonic/gin"
)

func TestOfficialCloudRoutesAreNotRegistered(t *testing.T) {
	gin.SetMode(gin.TestMode)
	server := gin.New()
	ServeAPI(server)

	removedRoutes := []string{
		"/api/cloud/",
		"/api/repo/purgeCloudRepo",
		"/api/repo/getCloudRepoTagSnapshots",
		"/api/repo/getCloudRepoSnapshots",
		"/api/repo/removeCloudRepoTagSnapshot",
		"/api/repo/uploadCloudSnapshot",
		"/api/repo/downloadCloudSnapshot",
		"/api/asset/uploadCloud",
		"/api/asset/uploadCloudByAssetsPaths",
		"/api/inbox/",
		"/api/setting/setAccount",
	}

	for _, route := range server.Routes() {
		for _, removed := range removedRoutes {
			if strings.HasPrefix(route.Path, removed) {
				t.Fatalf("official cloud route is still registered: %s", route.Path)
			}
		}
	}
}
