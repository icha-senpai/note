// FileLock - Read and write files with lock.
// Copyright (c) 2022-present, Scribli


package filelock

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/icha-senpai/note/third_party/forks/httpclient"
	"github.com/icha-senpai/note/third_party/forks/logging"
)

var (
	Container         = ""
	AndroidServerPort = 6906
)

func Walk(root string, fn fs.WalkDirFunc) error {
	if "harmony" == Container {
		return filepath.WalkDir(root, fn)
	}

	if strings.Contains(runtime.GOOS, "android") {
		// Data sync may cause data loss on Android 14

		start := time.Now()
		req := httpclient.NewCloudFileRequest2m()
		req.SetBody(map[string]interface{}{"dir": root})
		req.SetContentType("application/json; charset=utf-8")
		resp, err := req.Post("http://[::1]:" + fmt.Sprintf("%d", AndroidServerPort) + "/api/walkDir")
		logging.LogInfof("walk dir [%s] cost [%s]", root, time.Since(start))
		if nil != err {
			logging.LogErrorf("walk dir [%s] failed: %s", root, err)
			return filepath.WalkDir(root, fn)
		}
		if 200 != resp.StatusCode {
			logging.LogErrorf("walk dir [%s] failed: %d", root, resp.StatusCode)
			return filepath.WalkDir(root, fn)
		}

		result := map[string]interface{}{}
		if err = resp.UnmarshalJson(&result); nil != err {
			logging.LogErrorf("walk dir [%s] failed: %s", root, err)
			return errors.New("walk dir failed")
		}

		code := result["code"].(float64)
		if 0 != code {
			msgResult := result["msg"]
			var msg string
			if nil != msgResult {
				msg = msgResult.(string)
			}
			logging.LogErrorf("walk dir [%s] failed: %f, %s", root, code, msg)
			return errors.New("walk dir failed")
		}

		data := result["data"].(map[string]interface{})
		filesData := data["files"].([]interface{})
		var infos []*RemoteFile
		for _, f := range filesData {
			info := &RemoteFile{info: f.(map[string]interface{})}
			infos = append(infos, info)
		}

		sort.Slice(infos, func(i, j int) bool {
			return infos[i].Path() < infos[j].Path()
		})

		var skipPaths []string
		for _, info := range infos {
			p := info.Path()

			isSkipped := false
			for _, sp := range skipPaths {
				if strings.HasPrefix(p, sp) {
					isSkipped = true
					break
				}
			}
			if isSkipped {
				continue
			}

			err = fn(p, fs.FileInfoToDirEntry(info), nil)
			if nil != err {
				if errors.Is(err, fs.SkipDir) {
					skipPaths = append(skipPaths, p+"/")
					continue
				}
				if errors.Is(err, fs.SkipAll) {
					return nil
				}
				return err
			}
		}
		return nil
	}
	return filepath.WalkDir(root, fn)
}

type RemoteFile struct {
	info map[string]interface{}
}

func (f *RemoteFile) Path() string {
	return f.info["path"].(string)
}

func (f *RemoteFile) Name() string {
	return f.info["name"].(string)
}

func (f *RemoteFile) Size() int64 {
	return int64(f.info["size"].(float64))
}

func (f *RemoteFile) Mode() fs.FileMode {
	if f.IsDir() {
		return 0755
	}
	return 0644
}

func (f *RemoteFile) ModTime() time.Time {
	ms := int64(f.info["updated"].(float64))
	return time.UnixMilli(ms)
}

func (f *RemoteFile) IsDir() bool {
	return f.info["isDir"].(bool)
}

func (f *RemoteFile) Sys() interface{} {
	return nil
}
