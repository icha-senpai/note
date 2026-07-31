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

package util

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/icha-senpai/note/third_party/forks/gulu"
	"github.com/icha-senpai/note/third_party/forks/logging"
)

var ErrEbookConvertNotFound = errors.New("not found executable ebook-convert")

func EbookConvert(binPath, inputPath, outputPath string, extraArgs ...string) error {
	if !IsValidEbookConvertBin(binPath) {
		return ErrEbookConvertNotFound
	}

	args := []string{inputPath, outputPath}
	args = append(args, extraArgs...)
	cmd := exec.Command(binPath, args...)
	gulu.CmdAttr(cmd)
	output, err := cmd.CombinedOutput()
	if err != nil {
		msg := gulu.DecodeCmdOutput(output)
		if msg == "" {
			msg = err.Error()
		}
		logging.LogErrorf("ebook-convert failed: %s", msg)
		return errors.New(msg)
	}
	return nil
}

func IsValidEbookConvertBin(binPath string) bool {
	if "" == binPath {
		return false
	}

	if real, err := filepath.EvalSymlinks(binPath); err == nil {
		binPath = real
	}

	fi, err := os.Stat(binPath)
	if err != nil || fi.IsDir() || !fi.Mode().IsRegular() {
		return false
	}

	f, err := os.Open(binPath)
	if err != nil {
		return false
	}
	defer f.Close()

	header := make([]byte, 16)
	n, _ := f.Read(header)
	header = header[:n]

	if bytes.HasPrefix(header, []byte("#!")) {
		return false
	}

	isBin := false
	if len(header) >= 4 {
		switch {
		case bytes.Equal(header[:4], []byte{0x7f, 'E', 'L', 'F'}):
			isBin = true
		case bytes.Equal(header[:4], []byte{0xfe, 0xed, 0xfa, 0xce}), bytes.Equal(header[:4], []byte{0xce, 0xfa, 0xed, 0xfe}):
			isBin = true
		case bytes.Equal(header[:4], []byte{0xfe, 0xed, 0xfa, 0xcf}), bytes.Equal(header[:4], []byte{0xcf, 0xfa, 0xed, 0xfe}):
			isBin = true
		case bytes.Equal(header[:4], []byte{0xca, 0xfe, 0xba, 0xbe}), bytes.Equal(header[:4], []byte{0xbe, 0xba, 0xfe, 0xca}):
			isBin = true
		}
	}
	if !isBin && len(header) >= 2 && bytes.Equal(header[:2], []byte{'M', 'Z'}) {
		isBin = true
	}
	if !isBin && gulu.OS.IsWindows() && strings.EqualFold(filepath.Ext(binPath), ".exe") {
		isBin = true
	}
	if !isBin {
		logging.LogWarnf("file [%s] is not a valid binary executable", binPath)
		return false
	}

	cmd := exec.Command(binPath, "--version")
	gulu.CmdAttr(cmd)
	data, err := cmd.CombinedOutput()
	return err == nil && strings.Contains(strings.ToLower(string(data)), "ebook-convert")
}
