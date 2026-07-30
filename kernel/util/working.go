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
	"flag"
	"fmt"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/icha-senpai/note/third_party/forks/go-humanize"
	"github.com/icha-senpai/note/third_party/forks/gulu"
	figure "github.com/icha-senpai/note/third_party/forks/github/common-nighthawk/go-figure"
	"github.com/icha-senpai/note/third_party/forks/github/gofrs/flock"
	"github.com/icha-senpai/note/third_party/forks/filelock"
	"github.com/icha-senpai/note/third_party/forks/httpclient"
	"github.com/icha-senpai/note/third_party/forks/logging"
	"github.com/icha-senpai/note/third_party/forks/external/golang.org/x/mod/semver"
)

// var Mode = "dev"
var Mode = "prod"

const (
	Ver       = "3.7.3"
	IsInsider = false
)

func IsReleaseVer(ver string) bool {
	v := "v" + strings.TrimPrefix(ver, "v")
	return semver.IsValid(v) && semver.Prerelease(v) == ""
}

var (
	RunInContainer              = false
	ScribliAccessAuthCodeBypass = false
	AttachUI                    = false
	OfflineMode                 = false
	OfficialServicesDisabled    = true
)

func initEnvVars() {
	RunInContainer = isRunningInDockerContainer()
	var err error
	if ScribliAccessAuthCodeBypass, err = strconv.ParseBool(os.Getenv("SCRIBLI_ACCESS_AUTH_CODE_BYPASS")); err != nil {
		ScribliAccessAuthCodeBypass = false
	}
	if OfflineMode, err = strconv.ParseBool(os.Getenv("SCRIBLI_OFFLINE_MODE")); err != nil {
		OfflineMode = false
	}
}

var (
	bootProgress = atomic.Int32{}
	bootDetails  string
	HttpServer   *http.Server
	HttpServing  = false

	SafeMode = false
)

// If a commandline parameter is empty, fallback to the env var.
//
// "empty" means the parameter is not set or set to an empty string.
// It returns a pointer to string, to be a drop-in replacement for
// the commandline parameter itself.
func coalesceToEnvVar(fromCLI *string, envVarName string) *string {
	if fromCLI == nil || "" == *fromCLI {
		ret := os.Getenv(envVarName)
		return &ret
	}
	return fromCLI
}

func InitWorkspace(workspacePath, wdPath string) {
	initEnvVars()
	initMime()
	initHttpClient()

	if "" != wdPath {
		WorkingDir = wdPath
	}

	Container = ContainerStd
	if RunInContainer {
		Container = ContainerDocker
	}

	initWorkspaceDir(workspacePath)
	initPathDir()

	AppearancePath = filepath.Join(ConfDir, "appearance")
	if "dev" == Mode {
		ThemesPath = filepath.Join(WorkingDir, "appearance", "themes")
		IconsPath = filepath.Join(WorkingDir, "appearance", "icons")
	} else {
		ThemesPath = filepath.Join(AppearancePath, "themes")
		IconsPath = filepath.Join(AppearancePath, "icons")
	}

	LogPath = filepath.Join(TempDir, "scribli.log")
}

func Boot() {
	IncBootProgress(3, BootL10n(299, "Booting kernel..."))

	workspacePath := flag.String("workspace", "", "dir path of the workspace, default to ~/Scribli/")
	wdPath := flag.String("wd", WorkingDir, "working directory of Scribli")
	port := flag.String("port", "0", "port of the HTTP server")
	readOnly := flag.String("readonly", "false", "read-only mode")
	accessAuthCode := flag.String("accessAuthCode", "", "access auth code")
	ssl := flag.Bool("ssl", false, "for https and wss")
	attachUI := flag.Bool("attach-ui", false, "attach kernel lifecycle to desktop UI process (used by Electron)")
	lang := flag.String("lang", "", "ar/de/en/es/fr/he/hi/id/it/ja/ko/nl/pl/pt-BR/ru/sk/th/tr/uk/zh-CN/zh-TW")
	mode := flag.String("mode", "prod", "dev/prod")
	safeMode := flag.Bool("safe-mode", false, "boot in safe mode")
	flag.Parse()

	BootWithFlags(*workspacePath, *wdPath, *port, *readOnly, *accessAuthCode, *lang, *mode, *ssl, *attachUI, *safeMode)
}

func BootWithFlags(workspacePath, wdPath, port, readOnly, accessAuthCode, lang, mode string, ssl, attachUI, safeMode bool) {
	SafeMode = safeMode
	// Fallback to env vars if commandline args are not set
	// valid only for CLI args that default to "", as the
	// others have explicit (sane) defaults
	workspacePath = *coalesceToEnvVar(&workspacePath, "SCRIBLI_WORKSPACE_PATH")
	accessAuthCode = *coalesceToEnvVar(&accessAuthCode, "SCRIBLI_ACCESS_AUTH_CODE")
	lang = *coalesceToEnvVar(&lang, "SCRIBLI_LANG")

	if "" != lang {
		Lang = LangToBCP47(lang)
	}
	Mode = mode
	ServerPort = port
	ReadOnly, _ = strconv.ParseBool(readOnly)
	AttachUI = attachUI
	AccessAuthCode = accessAuthCode
	AccessAuthCode = RemoveInvalid(AccessAuthCode)
	AccessAuthCode = strings.TrimSpace(AccessAuthCode)
	Container = ContainerStd
	if RunInContainer {
		Container = ContainerDocker
		if "" == AccessAuthCode { // Still empty?
			interruptBoot := true

			if ScribliAccessAuthCodeBypass {
				interruptBoot = false
				fmt.Println("bypass access auth code check since the env [SCRIBLI_ACCESS_AUTH_CODE_BYPASS] is set to [true]")
			}

			if interruptBoot {
				// The access authorization code command line parameter must be set when deploying via Docker
				fmt.Printf("the access authorization code command line parameter (--accessAuthCode) must be set when deploying via Docker\n")
				fmt.Printf("or you can set the SCRIBLI_ACCESS_AUTH_CODE env var")
				os.Exit(logging.ExitCodeSecurityRisk)
			}
		}
	}
	if ContainerStd != Container {
		ServerPort = FixedPort
	}

	UserAgent = UserAgent + " " + Container + "/" + runtime.GOOS
	httpclient.SetUserAgent(UserAgent)

	InitWorkspace(workspacePath, wdPath)

	msStoreFilePath := filepath.Join(WorkingDir, "ms-store")
	ISMicrosoftStore = gulu.File.IsExist(msStoreFilePath)

	SSL = ssl
	logging.SetLogPath(LogPath)

	tryLockWorkspace()

	bootBanner := figure.NewColorFigure("Scribli", "isometric3", "green", true)
	logging.LogInfo("\n" + bootBanner.String())
	logBootInfo()
}

var bootDetailsLock = sync.Mutex{}

func setBootDetails(details string) {
	bootDetailsLock.Lock()
	bootDetails = "v" + Ver + " " + details
	bootDetailsLock.Unlock()
}

func SetBootDetails(details string) {
	if 100 <= bootProgress.Load() {
		return
	}
	setBootDetails(details)
}

//

func BootL10n(num int, fallback string) string {
	if s := Langs[Lang][num]; "" != s {
		return s
	}
	if s := Langs["en"][num]; "" != s {
		return s
	}
	return fallback
}

func IncBootProgress(progress int32, details string) {
	if 100 <= bootProgress.Load() {
		return
	}
	bootProgress.Add(progress)
	setBootDetails(details)
}

func IsBooted() bool {
	return 100 <= bootProgress.Load()
}

func GetBootProgressDetails() (progress int32, details string) {
	progress = bootProgress.Load()
	bootDetailsLock.Lock()
	details = bootDetails
	bootDetailsLock.Unlock()
	return
}

func GetBootProgress() int32 {
	return bootProgress.Load()
}

func SetBooted() {

	bootProgress.Store(100)
	setBootDetails(BootL10n(300, "Finishing boot..."))
	logging.LogInfof("kernel booted")
}

var (
	HomeDir, _    = gulu.OS.Home()
	WorkingDir, _ = os.Getwd()

	WorkspaceDir       string
	WorkspaceName      string
	WorkspaceLock      *flock.Flock
	ConfDir            string
	DataDir            string
	RepoDir            string
	HistoryDir         string
	TempDir            string
	QueueDir           string
	LogPath            string
	DBName             = "scribli.db"
	DBPath             string
	HistoryDBPath      string
	AssetContentDBPath string
	BlockTreeDBPath    string
	AppearancePath     string
	ThemesPath         string
	IconsPath          string
	SnippetsPath       string
	ShortcutsPath      string

	UIProcessIDs = sync.Map{}
)

func initWorkspaceDir(workspaceArg string) {
	userHomeConfDir := UserHomeConfDir()
	workspaceConf := filepath.Join(userHomeConfDir, "workspace.json")
	logging.SetLogPath(filepath.Join(userHomeConfDir, "kernel.log"))

	if !gulu.File.IsExist(workspaceConf) {
		if err := os.MkdirAll(userHomeConfDir, 0755); err != nil && !os.IsExist(err) {
			logging.LogErrorf("create user home conf folder [%s] failed: %s", userHomeConfDir, err)
			os.Exit(logging.ExitCodeInitWorkspaceErr)
		}
	}

	defaultWorkspaceDir := filepath.Join(HomeDir, "Scribli")
	if gulu.OS.IsWindows() {

		if userProfile := os.Getenv("USERPROFILE"); "" != userProfile {
			defaultWorkspaceDir = filepath.Join(userProfile, "Scribli")
		}
	} else if gulu.OS.IsDarwin() {
		// Change the initial workspace path to ~/Library/Application Support/Scribli on macOS
		defaultWorkspaceDir = filepath.Join(HomeDir, "Library", "Application Support", "Scribli")
	}

	var workspacePaths []string
	if !gulu.File.IsExist(workspaceConf) {
		WorkspaceDir = defaultWorkspaceDir
	} else {
		workspacePaths, _ = ReadWorkspacePaths()
		if 0 < len(workspacePaths) {
			WorkspaceDir = workspacePaths[len(workspacePaths)-1]
		} else {
			WorkspaceDir = defaultWorkspaceDir
		}
	}

	if "" != workspaceArg {
		WorkspaceDir = workspaceArg
	}

	WorkspaceDir = filepath.Clean(WorkspaceDir)

	if !gulu.File.IsDir(WorkspaceDir) {
		logging.LogWarnf("use the default workspace [%s] since the specified workspace [%s] is not a dir", defaultWorkspaceDir, WorkspaceDir)
		if err := os.MkdirAll(defaultWorkspaceDir, 0755); err != nil && !os.IsExist(err) {
			logging.LogErrorf("create default workspace folder [%s] failed: %s", defaultWorkspaceDir, err)
			os.Exit(logging.ExitCodeInitWorkspaceErr)
		}
		WorkspaceDir = defaultWorkspaceDir
	}
	workspacePaths = append(workspacePaths, WorkspaceDir)

	if err := WriteWorkspacePaths(workspacePaths); err != nil {
		logging.LogErrorf("write workspace conf [%s] failed: %s", workspaceConf, err)
		os.Exit(logging.ExitCodeInitWorkspaceErr)
	}

	WorkspaceName = filepath.Base(WorkspaceDir)
	ConfDir = filepath.Join(WorkspaceDir, "conf")
	DataDir = filepath.Join(WorkspaceDir, "data")
	RepoDir = filepath.Join(WorkspaceDir, "repo")
	HistoryDir = filepath.Join(WorkspaceDir, "history")
	TempDir = filepath.Join(WorkspaceDir, "temp")
	QueueDir = filepath.Join(TempDir, "queue")
	osTmpDir := filepath.Join(TempDir, "os")
	os.RemoveAll(osTmpDir)
	if err := os.MkdirAll(osTmpDir, 0755); err != nil {
		logging.LogErrorf("create os tmp dir [%s] failed: %s", osTmpDir, err)
		os.Exit(logging.ExitCodeInitWorkspaceErr)
	}
	os.RemoveAll(filepath.Join(TempDir, "repo"))

	os.RemoveAll(filepath.Join(TempDir, "export"))
	os.Setenv("TMPDIR", osTmpDir)
	os.Setenv("TEMP", osTmpDir)
	os.Setenv("TMP", osTmpDir)
	DBPath = filepath.Join(TempDir, DBName)
	HistoryDBPath = filepath.Join(TempDir, "history.db")
	AssetContentDBPath = filepath.Join(TempDir, "asset_content.db")
	BlockTreeDBPath = filepath.Join(TempDir, "blocktree.db")
	SnippetsPath = filepath.Join(DataDir, "snippets")
	ShortcutsPath = filepath.Join(userHomeConfDir, "shortcuts")
}

func DeduplicateWorkspacePaths(paths []string) []string {
	if !gulu.OS.IsWindows() {
		return gulu.Str.RemoveDuplicatedElem(paths)
	}
	seen := map[string]bool{}
	var result []string
	for _, p := range paths {
		key := strings.ToLower(filepath.Clean(p))
		if seen[key] {
			continue
		}
		seen[key] = true
		result = append(result, p)
	}
	return result
}

func RemoveWorkspacePath(paths []string, target string) []string {
	if !gulu.OS.IsWindows() {
		return gulu.Str.RemoveElem(paths, target)
	}
	targetLower := strings.ToLower(target)
	var result []string
	for _, p := range paths {
		if strings.ToLower(p) == targetLower {
			continue
		}
		result = append(result, p)
	}
	return result
}

func ReadWorkspacePaths() (ret []string, err error) {
	ret = []string{}
	workspaceConf := filepath.Join(UserHomeConfDir(), "workspace.json")
	data, err := os.ReadFile(workspaceConf)
	if err != nil {
		msg := fmt.Sprintf("read workspace conf [%s] failed: %s", workspaceConf, err)
		logging.LogError(msg)
		err = errors.New(msg)
		return
	}

	if err = gulu.JSON.UnmarshalJSON(data, &ret); err != nil {
		msg := fmt.Sprintf("unmarshal workspace conf [%s] failed: %s", workspaceConf, err)
		logging.LogError(msg)
		err = errors.New(msg)
		return
	}

	var tmp []string
	workspaceBaseDir := filepath.Dir(HomeDir)
	for _, d := range ret {
		if ContainerIOS == Container && strings.Contains(d, "/Documents/") {

			d = d[strings.Index(d, "/Documents/")+len("/Documents/"):]
			d = filepath.Join(workspaceBaseDir, d)
		}

		d = strings.TrimRight(d, " \t\n")
		d = filepath.Clean(d)
		if gulu.File.IsDir(d) {
			tmp = append(tmp, d)
		} else {
			logging.LogWarnf("workspace path [%s] is not a dir", d)
		}
	}
	ret = tmp
	ret = DeduplicateWorkspacePaths(ret)
	return
}

func WriteWorkspacePaths(workspacePaths []string) (err error) {
	workspacePaths = DeduplicateWorkspacePaths(workspacePaths)
	workspaceConf := filepath.Join(UserHomeConfDir(), "workspace.json")
	data, err := gulu.JSON.MarshalJSON(workspacePaths)
	if err != nil {
		msg := fmt.Sprintf("marshal workspace conf [%s] failed: %s", workspaceConf, err)
		logging.LogError(msg)
		err = errors.New(msg)
		return
	}

	if err = filelock.WriteFile(workspaceConf, data); err != nil {
		msg := fmt.Sprintf("write workspace conf [%s] failed: %s", workspaceConf, err)
		logging.LogError(msg)
		err = errors.New(msg)
		return
	}
	return
}

func UserHomeConfDir() string {
	return filepath.Join(HomeDir, ".config", "scribli")
}

var (
	ServerURL  *url.URL
	ServerPort = "0"

	ReadOnly       bool
	AccessAuthCode string
	Lang           = ""

	Container        string // docker, android, ios, harmony, std
	ISMicrosoftStore bool
)

const (
	ContainerStd     = "std"
	ContainerDocker  = "docker"
	ContainerAndroid = "android"
	ContainerIOS     = "ios"
	ContainerHarmony = "harmony"

	LocalHost = "127.0.0.1"
	FixedPort = "6806"
)

func IsMobileContainer() bool {
	return ContainerAndroid == Container || ContainerIOS == Container || ContainerHarmony == Container
}

func initPathDir() {
	if err := os.MkdirAll(ConfDir, 0755); err != nil && !os.IsExist(err) {
		logging.LogFatalf(logging.ExitCodeInitWorkspaceErr, "create conf folder [%s] failed: %s", ConfDir, err)
	}
	if err := os.MkdirAll(DataDir, 0755); err != nil && !os.IsExist(err) {
		logging.LogFatalf(logging.ExitCodeInitWorkspaceErr, "create data folder [%s] failed: %s", DataDir, err)
	}
	if err := os.MkdirAll(TempDir, 0755); err != nil && !os.IsExist(err) {
		logging.LogFatalf(logging.ExitCodeInitWorkspaceErr, "create temp folder [%s] failed: %s", TempDir, err)
	}

	assets := filepath.Join(DataDir, "assets")
	if err := os.MkdirAll(assets, 0755); err != nil && !os.IsExist(err) {
		logging.LogFatalf(logging.ExitCodeInitWorkspaceErr, "create data assets folder [%s] failed: %s", assets, err)
	}

	templates := filepath.Join(DataDir, "templates")
	if err := os.MkdirAll(templates, 0755); err != nil && !os.IsExist(err) {
		logging.LogFatalf(logging.ExitCodeInitWorkspaceErr, "create data templates folder [%s] failed: %s", templates, err)
	}

	widgets := filepath.Join(DataDir, "widgets")
	if err := os.MkdirAll(widgets, 0755); err != nil && !os.IsExist(err) {
		logging.LogFatalf(logging.ExitCodeInitWorkspaceErr, "create data widgets folder [%s] failed: %s", widgets, err)
	}

	plugins := filepath.Join(DataDir, "plugins")
	if err := os.MkdirAll(plugins, 0755); err != nil && !os.IsExist(err) {
		logging.LogFatalf(logging.ExitCodeInitWorkspaceErr, "create data plugins folder [%s] failed: %s", plugins, err)
	}

	emojis := filepath.Join(DataDir, "emojis")
	if err := os.MkdirAll(emojis, 0755); err != nil && !os.IsExist(err) {
		logging.LogFatalf(logging.ExitCodeInitWorkspaceErr, "create data emojis folder [%s] failed: %s", emojis, err)
	}

	queueDir := filepath.Join(TempDir, "queue")
	if err := os.MkdirAll(queueDir, 0755); err != nil && !os.IsExist(err) {
		logging.LogFatalf(logging.ExitCodeInitWorkspaceErr, "create queue folder [%s] failed: %s", queueDir, err)
	}

	// Support directly access `data/public/*` contents via URL link
	public := filepath.Join(DataDir, "public")
	if err := os.MkdirAll(public, 0755); err != nil && !os.IsExist(err) {
		logging.LogFatalf(logging.ExitCodeInitWorkspaceErr, "create data public folder [%s] failed: %s", public, err)
	}
}

func initMime() {

	//
	//
	mime.AddExtensionType(".css", "text/css")
	mime.AddExtensionType(".js", "text/javascript")
	mime.AddExtensionType(".mjs", "text/javascript")
	mime.AddExtensionType(".html", "text/html")
	mime.AddExtensionType(".json", "application/json")
	mime.AddExtensionType(".woff2", "font/woff2")

	mime.AddExtensionType(".doc", "application/msword")
	mime.AddExtensionType(".docx", "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
	mime.AddExtensionType(".xls", "application/vnd.ms-excel")
	mime.AddExtensionType(".xlsx", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	mime.AddExtensionType(".dwg", "image/x-dwg")
	mime.AddExtensionType(".dxf", "image/x-dxf")
	mime.AddExtensionType(".dwf", "drawing/x-dwf")
	mime.AddExtensionType(".pdf", "application/pdf")

	mime.AddExtensionType(".svg", "image/svg+xml")

	mime.AddExtensionType(".sy", "application/json")

	mime.AddExtensionType(".md", "text/markdown")
	mime.AddExtensionType(".markdown", "text/markdown")

	mime.AddExtensionType(".png", "image/png")
	mime.AddExtensionType(".jpg", "image/jpeg")
	mime.AddExtensionType(".jpeg", "image/jpeg")
	mime.AddExtensionType(".gif", "image/gif")
	mime.AddExtensionType(".bmp", "image/bmp")
	mime.AddExtensionType(".tiff", "image/tiff")
	mime.AddExtensionType(".tif", "image/tiff")
	mime.AddExtensionType(".webp", "image/webp")
	mime.AddExtensionType(".ico", "image/x-icon")
}

func GetDataAssetsAbsPath() (ret string) {
	ret = filepath.Join(DataDir, "assets")
	if IsSymlinkPath(ret) {

		var err error
		ret, err = filepath.EvalSymlinks(ret)
		if err != nil {
			logging.LogErrorf("read assets link failed: %s", err)
		}
	}
	return
}

func EncryptedDBPath(boxID string) string {
	return filepath.Join(TempDir, "scribli-encrypted-"+boxID+".db")
}

func EncryptedBlockTreeDBPath(boxID string) string {
	return filepath.Join(TempDir, "scribli-encrypted-"+boxID+"-blocktree.db")
}

func tryLockWorkspace() {
	WorkspaceLock = flock.New(filepath.Join(WorkspaceDir, ".lock"))
	ok, err := WorkspaceLock.TryLock()
	if ok {
		return
	}
	if err != nil {
		logging.LogErrorf("lock workspace [%s] failed: %s", WorkspaceDir, err)
	} else {
		logging.LogErrorf("lock workspace [%s] failed", WorkspaceDir)
	}
	os.Exit(logging.ExitCodeWorkspaceLocked)
}

func IsWorkspaceLocked(workspacePath string) bool {
	if !gulu.File.IsDir(workspacePath) {
		return false
	}

	lockFilePath := filepath.Join(workspacePath, ".lock")
	if !gulu.File.IsExist(lockFilePath) {
		return false
	}

	f := flock.New(lockFilePath)
	defer f.Unlock()
	ok, _ := f.TryLock()
	if ok {
		return false
	}
	return true
}

func UnlockWorkspace() {
	if nil == WorkspaceLock {
		return
	}

	if err := WorkspaceLock.Unlock(); err != nil {
		logging.LogErrorf("unlock workspace [%s] failed: %s", WorkspaceDir, err)
		return
	}

	if err := os.Remove(filepath.Join(WorkspaceDir, ".lock")); err != nil {
		logging.LogErrorf("remove workspace lock failed: %s", err)
		return
	}
}

func LogDatabaseSize(dbPath string) {
	dbFile, err := os.Stat(dbPath)
	if nil != err {
		return
	}

	dbSize := humanize.BytesCustomCeil(uint64(dbFile.Size()), 2)
	logging.LogInfof("database [%s] size [%s]", dbPath, dbSize)
}

func RemoveDatabaseFile(dbPath string) {
	if gulu.File.IsExist(dbPath) {
		err := os.RemoveAll(dbPath)
		if err != nil {
			logging.LogErrorf("remove database file [%s] failed: %s", dbPath, err)
			return
		}
	}

	if gulu.File.IsExist(dbPath + "-shm") {
		err := os.RemoveAll(dbPath + "-shm")
		if err != nil {
			logging.LogErrorf("remove database file [%s] failed: %s", dbPath+"-shm", err)
			return
		}
	}

	if gulu.File.IsExist(dbPath + "-wal") {
		err := os.RemoveAll(dbPath + "-wal")
		if err != nil {
			logging.LogErrorf("remove database file [%s] failed: %s", dbPath+"-wal", err)
			return
		}
	}
}
