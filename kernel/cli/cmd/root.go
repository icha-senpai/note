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
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"

	"github.com/icha-senpai/note/kernel/model"
	"github.com/icha-senpai/note/kernel/sql"
	"github.com/icha-senpai/note/kernel/treenode"
	"github.com/icha-senpai/note/kernel/util"
	"github.com/icha-senpai/note/third_party/forks/logging"

	"github.com/icha-senpai/note/third_party/forks/github/spf13/cobra"
)

var (
	workspacePath string
	outputFormat  string
	dryRun        bool
	logLevel      string
)

var rootCmd = &cobra.Command{
	Use:     "Scribli-Kernel",
	Version: util.Ver,
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {

		name := cmd.Name()

		if name == "serve" || (cmd.Parent() != nil && cmd.Parent().Name() == "workspace") {
			return nil
		}
		model.FlushTxQueue()
		sql.FlushQueue()
		return nil
	},
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {

		if cmd.Parent() != nil && cmd.Parent().Name() == "workspace" {
			return nil
		}

		if workingDir := resolveWorkingDir(); workingDir != "" {
			util.WorkingDir = workingDir
		}

		langsDir := filepath.Join(util.WorkingDir, "appearance", "langs")
		if _, err := os.Stat(langsDir); os.IsNotExist(err) {
			return fmt.Errorf("appearance files not found at [%s]", langsDir)
		}

		if workspacePath == "" {
			workspacePath = os.Getenv("SCRIBLI_WORKSPACE_PATH")
		}
		if workspacePath == "" {
			workspacePath = filepath.Join(util.HomeDir, "Scribli")
		}

		if _, err := os.Stat(workspacePath); os.IsNotExist(err) {
			return fmt.Errorf("directory not found: %s", workspacePath)
		}
		if !util.IsWorkspaceDir(workspacePath) {
			return fmt.Errorf("not a valid workspace: %s", workspacePath)
		}

		util.Mode = "prod"
		util.InitWorkspace(workspacePath, util.WorkingDir)

		logging.SetLogPath(filepath.Join(util.TempDir, "siyuan-cli.log"))
		logging.SetLogToStdout(false)

		effectiveLevel := logLevel
		if "" == effectiveLevel {
			effectiveLevel = "warn"
		}
		logging.SetLogLevel(effectiveLevel)
		util.CLILogLevel = effectiveLevel

		model.InitConf()
		sql.InitDatabase(false)
		sql.InitHistoryDatabase(false)
		sql.InitAssetContentDatabase(false)
		sql.SetCaseSensitive(model.Conf.Search.CaseSensitive)
		sql.SetIndexAssetPath(model.Conf.Search.IndexAssetPath)

		model.PrepareEmbeddingSearch()
		if err := rejectEncryptedNotebookCLI(cmd, args); err != nil {
			return err
		}
		return nil
	},
}

func rejectEncryptedNotebookCLI(cmd *cobra.Command, args []string) error {
	if cmd == serveCmd {
		return nil
	}
	if (cmd == notebookRandomIconCmd && !cmd.Flags().Changed("id")) || cmd == exportDataCmd {
		boxID, err := firstEncryptedNotebookID()
		if err != nil {
			return err
		}
		if boxID != "" {
			return fmt.Errorf("CLI does not support encrypted notebook [%s]", boxID)
		}
	}

	var encryptedTarget string
	checkID := func(id string) bool {
		if id == "" {
			return false
		}
		if model.IsEncryptedBox(id) {
			encryptedTarget = id
			return true
		}
		if bt := treenode.GetBlockTree(id); bt != nil && model.IsEncryptedBox(bt.BoxID) {
			encryptedTarget = bt.BoxID
			return true
		}
		return false
	}

	for _, flagName := range []string{"notebook", "box", "id", "ids", "parent", "previous", "block"} {
		flag := cmd.Flags().Lookup(flagName)
		if flag == nil {
			continue
		}
		values := []string{flag.Value.String()}
		if flag.Value.Type() == "stringArray" {
			values, _ = cmd.Flags().GetStringArray(flagName)
		}
		for _, value := range values {
			for id := range strings.SplitSeq(value, ",") {
				if checkID(strings.TrimSpace(id)) {
					return fmt.Errorf("CLI does not support encrypted notebook [%s]", encryptedTarget)
				}
			}
		}
	}

	if cmd.Parent() == fileCmd {
		if slices.ContainsFunc(args, isEncryptedNotebookWorkspacePath) {
			return fmt.Errorf("CLI does not support files in encrypted notebooks")
		}
		if pathFlag := cmd.Flags().Lookup("path"); pathFlag != nil && pathFlag.Value.String() != "" && isEncryptedNotebookWorkspacePath(pathFlag.Value.String()) {
			return fmt.Errorf("CLI does not support files in encrypted notebooks")
		}
	}
	if cmd.Parent() == assetCmd {
		if pathFlag := cmd.Flags().Lookup("path"); pathFlag != nil && pathFlag.Value.String() != "" {
			assetPath := pathFlag.Value.String()
			if !filepath.IsAbs(assetPath) {
				assetPath = filepath.Join("data", assetPath)
			}
			if isEncryptedNotebookWorkspacePath(assetPath) {
				return fmt.Errorf("CLI does not support files in encrypted notebooks")
			}
		}
	}
	return nil
}

func firstEncryptedNotebookID() (string, error) {
	boxes, err := model.ListNotebooks()
	if err != nil {
		return "", err
	}
	for _, box := range boxes {
		if model.IsEncryptedBox(box.ID) {
			return box.ID, nil
		}
	}
	return "", nil
}

func isEncryptedNotebookWorkspacePath(p string) bool {
	return isEncryptedNotebookWorkspacePathWith(p, util.WorkspaceDir, util.DataDir, model.IsEncryptedBox)
}

func isEncryptedNotebookWorkspacePathWith(p, workspaceDir, dataDir string, isEncryptedBox func(string) bool) bool {
	abs := p
	if !filepath.IsAbs(abs) {
		abs = filepath.Join(workspaceDir, p)
	}
	rel, err := filepath.Rel(dataDir, filepath.Clean(abs))
	if err != nil || rel == "." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || rel == ".." {
		return false
	}
	boxID := strings.Split(rel, string(filepath.Separator))[0]
	return isEncryptedBox(boxID)
}

func resolveWorkingDir() string {
	if exePath, err := os.Executable(); err == nil {
		if resolved, err2 := filepath.EvalSymlinks(exePath); err2 == nil {
			exePath = resolved
		}
		exeDir := filepath.Dir(exePath)
		candidates := []string{
			filepath.Join(exeDir, ".."),              // resources/kernel/ → resources/ (production)
			filepath.Join(exeDir, "..", "app"),       // kernel/cli/ → kernel/ → app/
			filepath.Join(exeDir, "app"),             // kernel/ → app/
			filepath.Join(exeDir, "..", "..", "app"), // kernel/cli/cmd/... → .../app/
		}

		if runtime.GOOS == "darwin" {
			candidates = append(candidates,
				filepath.Join(exeDir, "..", "..", "..", "..", "Resources"),
			)
		}
		for _, d := range candidates {
			langsDir := filepath.Join(d, "appearance", "langs")
			if fi, err := os.Stat(langsDir); err == nil && fi.IsDir() {
				return d
			}
		}
	}
	return ""
}

func init() {
	rootCmd.Use = strings.TrimSuffix(filepath.Base(os.Args[0]), ".exe")
	rootCmd.Short = "Scribli Kernel v" + util.Ver
	rootCmd.Long = "Scribli Kernel v" + util.Ver + ". Manage workspace data directly or start the HTTP server."

	rootCmd.PersistentFlags().StringVarP(&workspacePath, "workspace", "w", "", "workspace path")
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "format", "f", "table", "output format: table | json")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "dry run mode: validate and print what would happen without making changes")
	rootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "v", "", "log level: off | trace | debug | info | warn | error | fatal (defaults to conf.json system.logLevel)")
}

func Execute() error {
	return rootCmd.Execute()
}

func HasSubCommand(name string) bool {
	for _, c := range rootCmd.Commands() {
		if c.Name() == name {
			return true
		}
	}
	return false
}
