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

//go:build windows

//

//     https://learn.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-setclipboarddata

//     https://learn.microsoft.com/en-us/windows/win32/dataxchg/using-the-clipboard
//

package util

import (
	"encoding/binary"
	"errors"
	"runtime"
	"syscall"
	"time"
	"unsafe"

	"github.com/gonutz/w32/v2"
)

const (
	cfHDROP       = 15
	dropfilesSize = 20
)

//

func WriteFilePaths(paths []string) error {
	if len(paths) == 0 {
		return nil
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	data, err := buildDropfilesData(paths)
	if err != nil {
		return err
	}
	if len(data) == 0 {
		return nil
	}

	// https://learn.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-setclipboarddata
	size := uint32(len(data))
	hMem := w32.GlobalAlloc(w32.GMEM_MOVEABLE, size)
	if hMem == 0 {
		return syscall.Errno(w32.GetLastError())
	}

	ptr := w32.GlobalLock(hMem)
	if ptr == nil {
		w32.GlobalFree(hMem)
		return syscall.Errno(w32.GetLastError())
	}

	w32.MoveMemory(ptr, unsafe.Pointer(&data[0]), size)

	// https://learn.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-setclipboarddata
	w32.GlobalUnlock(hMem)

	if err := waitOpenClipboard(); err != nil {
		w32.GlobalFree(hMem)
		return err
	}
	defer w32.CloseClipboard()

	if !w32.EmptyClipboard() {
		w32.GlobalFree(hMem)
		return syscall.Errno(w32.GetLastError())
	}
	if w32.SetClipboardData(cfHDROP, w32.HANDLE(hMem)) == 0 {
		w32.GlobalFree(hMem)
		return syscall.Errno(w32.GetLastError())
	}

	// https://learn.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-setclipboarddata
	return nil
}

//

// https://learn.microsoft.com/en-us/windows/win32/api/shlobj_core/ns-shlobj_core-dropfiles

func buildDropfilesData(paths []string) ([]byte, error) {
	var totalLen = dropfilesSize
	for _, p := range paths {
		u16, err := syscall.UTF16FromString(p)
		if err != nil {
			return nil, err
		}
		totalLen += len(u16) * 2
	}
	totalLen += 2

	buf := make([]byte, totalLen)

	binary.LittleEndian.PutUint32(buf[0:4], 20)
	// pt.x, pt.y, fNC, fWide
	binary.LittleEndian.PutUint32(buf[16:20], 1)

	offset := dropfilesSize
	for _, p := range paths {
		u16, err := syscall.UTF16FromString(p)
		if err != nil {
			return nil, err
		}
		for _, c := range u16 {
			binary.LittleEndian.PutUint16(buf[offset:offset+2], c)
			offset += 2
		}
	}
	return buf, nil
}

// https://learn.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-openclipboard
func waitOpenClipboard() error {
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if w32.OpenClipboard(0) {
			return nil
		}
		time.Sleep(time.Millisecond)
	}
	return errors.New("open clipboard timeout")
}
