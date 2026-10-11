// Scribli contributors. SPDX-License-Identifier: AGPL-3.0-or-later

package model

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Cleanup must not mistake an unreadable template for one without references.
func scanTemplateDatabaseReferences(root string, ids []string) (ret []string, err error) {
	info, err := os.Stat(root)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("template root is not a directory")
	}
	found := map[string]bool{}
	err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		if !entry.Type().IsRegular() {
			return fmt.Errorf("cannot scan template references [%s]", path)
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		buf := make([]byte, 64*1024)
		tail := ""
		for {
			n, readErr := file.Read(buf)
			text := tail + string(buf[:n])
			for _, id := range ids {
				if strings.Contains(text, id) {
					found[id] = true
				}
			}
			// Database IDs have a fixed 22-byte width; preserve boundary matches.
			tail = text[max(0, len(text)-21):]
			if readErr == io.EOF {
				return nil
			}
			if readErr != nil {
				return readErr
			}
		}
	})
	if err != nil {
		return nil, err
	}
	for _, id := range ids {
		if found[id] {
			ret = append(ret, id)
		}
	}
	return
}
