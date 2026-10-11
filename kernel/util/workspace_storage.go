// Based on SiYuan; Copyright (c) 2020-present, b3log.org.
// SPDX-License-Identifier: AGPL-3.0-or-later

package util

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/icha-senpai/note/third_party/forks/filelock"
	"github.com/icha-senpai/note/third_party/forks/lute/ast"
)

type WorkspaceStorage struct {
	Directories  map[string]int64
	AssetsSize   int64
	TotalSize    int64
	CalculatedAt int64
}

type workspaceStorageScan struct {
	done chan struct{}
	data WorkspaceStorage
	err  error
}

var workspaceStorageScans = struct {
	sync.Mutex
	pending map[string]*workspaceStorageScan
}{pending: map[string]*workspaceStorageScan{}}

// GetWorkspaceStorage coalesces concurrent scans without caching completed results.
func GetWorkspaceStorage(ctx context.Context, workspacePath string) (WorkspaceStorage, error) {
	if err := ctx.Err(); err != nil {
		return WorkspaceStorage{}, err
	}
	workspaceStorageScans.Lock()
	scan := workspaceStorageScans.pending[workspacePath]
	if scan == nil {
		scan = &workspaceStorageScan{done: make(chan struct{})}
		workspaceStorageScans.pending[workspacePath] = scan
		go func() {
			// A closed page does not cancel a shared scan; bound its background lifetime.
			scanCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
			defer cancel()
			scan.data, scan.err = measureWorkspaceStorage(scanCtx, workspacePath, filelock.Walk)
			workspaceStorageScans.Lock()
			delete(workspaceStorageScans.pending, workspacePath)
			close(scan.done)
			workspaceStorageScans.Unlock()
		}()
	}
	workspaceStorageScans.Unlock()
	select {
	case <-ctx.Done():
		return WorkspaceStorage{}, ctx.Err()
	case <-scan.done:
		return scan.data, scan.err
	}
}

// measureWorkspaceStorage reads metadata without decrypting or downloading files.
func measureWorkspaceStorage(ctx context.Context, workspacePath string, walk func(string, fs.WalkDirFunc) error) (ret WorkspaceStorage, err error) {
	// Convert walker panics to errors and release callers waiting for the scan.
	defer func() {
		if recovered := recover(); recovered != nil {
			ret = WorkspaceStorage{}
			err = fmt.Errorf("workspace storage scan failed: %v", recovered)
		}
	}()
	ret = WorkspaceStorage{Directories: map[string]int64{"data": 0, "repo": 0, "history": 0, "temp": 0, "conf": 0, "other": 0}}
	root, err := filepath.EvalSymlinks(workspacePath)
	if err != nil {
		return WorkspaceStorage{}, err
	}
	root, err = filepath.Abs(root)
	if err != nil {
		return WorkspaceStorage{}, err
	}
	info, err := os.Stat(root)
	if err != nil {
		return WorkspaceStorage{}, err
	}
	if !info.IsDir() {
		return WorkspaceStorage{}, fmt.Errorf("workspace is not a directory")
	}
	err = walk(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		if walkErr != nil {
			if path != root && errors.Is(walkErr, fs.ErrNotExist) {
				return nil
			}
			return walkErr
		}
		// Use Lstat to identify links consistently and avoid counting their targets.
		info, err := os.Lstat(path)
		if err != nil {
			// Sync or cache cleanup may remove files after directory enumeration.
			if path != root && errors.Is(err, fs.ErrNotExist) {
				return nil
			}
			return err
		}
		if info.Mode()&fs.ModeSymlink != 0 {
			if entry.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return fmt.Errorf("workspace entry is outside the workspace")
		}
		parts := strings.Split(filepath.ToSlash(rel), "/")
		group := "other"
		if len(parts) > 1 {
			if _, ok := ret.Directories[parts[0]]; ok {
				group = parts[0]
			}
		}
		ret.Directories[group] += info.Size()
		ret.TotalSize += info.Size()
		if group == "data" && workspaceStorageAsset(parts[1:]) {
			ret.AssetsSize += info.Size()
		}
		return nil
	})
	if err == nil {
		err = ctx.Err()
	}
	if err != nil {
		// Do not return partial statistics as a complete measurement.
		return WorkspaceStorage{}, err
	}
	ret.CalculatedAt = time.Now().UnixMilli()
	return ret, nil
}

// workspaceStorageAsset includes global, notebook and document assets, excluding package assets.
func workspaceStorageAsset(parts []string) bool {
	for _, part := range parts[:len(parts)-1] {
		if part == "assets" {
			return true
		}
		if !ast.IsNodeIDPattern(part) {
			return false
		}
	}
	return false
}
