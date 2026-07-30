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

package cmd

import (
	"github.com/icha-senpai/note/kernel/cache"
	"github.com/icha-senpai/note/kernel/job"
	"github.com/icha-senpai/note/kernel/model"
	"github.com/icha-senpai/note/kernel/plugin"
	"github.com/icha-senpai/note/kernel/server"
	"github.com/icha-senpai/note/kernel/sql"
	"github.com/icha-senpai/note/kernel/util"
	"github.com/icha-senpai/note/third_party/forks/logging"

	"github.com/icha-senpai/note/third_party/forks/github/spf13/cobra"
)

var (
	serveWdPath         string
	servePort           string
	serveReadOnly       string
	serveAccessAuthCode string
	serveLang           string
	serveMode           string
	serveSSL            bool
	serveAttachUI       bool
	serveSafeMode       bool
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start kernel HTTP server",
	Long:  "Start kernel HTTP server. All serving-related options below are passed to the kernel boot.",

	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {

		if "" != logLevel {
			logging.SetLogLevel(logLevel)
			util.CLILogLevel = logLevel
		}
		return nil // bypass root's init — BootWithFlags() handles it
	},
	Run: func(cmd *cobra.Command, args []string) {

		ws := workspacePath

		util.BootWithFlags(ws, serveWdPath, servePort, serveReadOnly, serveAccessAuthCode, serveLang, serveMode, serveSSL, serveAttachUI, serveSafeMode)

		model.InitJwtKey()
		model.InitConf()
		go server.Serve(false, model.Conf.CookieKey)
		model.InitAppearance()
		sql.InitDatabase(false)
		sql.InitHistoryDatabase(false)
		sql.InitAssetContentDatabase(false)
		sql.SetCaseSensitive(model.Conf.Search.CaseSensitive)
		sql.SetIndexAssetPath(model.Conf.Search.IndexAssetPath)

		model.BootSyncData()
		model.InitBoxes()
		model.LoadFlashcards()
		util.LoadAssetsTexts()

		util.SetBooted()
		util.PushClearAllMsg()

		job.StartCron()

		go model.AutoGenerateFileHistory()
		go cache.LoadAssets()
		go util.CheckFileSysStatus()
		go plugin.InitManager()
		go model.StartEmbeddingIndexer()

		model.WatchAssets()
		model.WatchEmojis()
		model.WatchThemes()
		model.HandleSignal()
	},
}

func init() {

	serveCmd.Flags().StringVar(&serveWdPath, "wd", resolveWorkingDir(), "working directory of Scribli")
	serveCmd.Flags().StringVar(&servePort, "port", "0", "port of the HTTP server")
	serveCmd.Flags().StringVar(&serveReadOnly, "readonly", "false", "read-only mode")
	serveCmd.Flags().StringVar(&serveAccessAuthCode, "accessAuthCode", "", "access auth code")
	serveCmd.Flags().StringVar(&serveLang, "lang", "", "ar/de/en/es/fr/he/hi/id/it/ja/ko/nl/pl/pt-BR/ru/sk/th/tr/uk/zh-CN/zh-TW")
	serveCmd.Flags().StringVar(&serveMode, "mode", "prod", "dev/prod")
	serveCmd.Flags().BoolVar(&serveSSL, "ssl", false, "for https and wss")
	serveCmd.Flags().BoolVar(&serveAttachUI, "attach-ui", false, "attach kernel lifecycle to desktop UI process (used by Electron)")
	serveCmd.Flags().BoolVar(&serveSafeMode, "safe-mode", false, "boot in safe mode")

	rootCmd.AddCommand(serveCmd)
}
