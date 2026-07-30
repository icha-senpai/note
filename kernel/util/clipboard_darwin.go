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

//go:build darwin && !ios

// This file writes file-path lists to macOS NSPasteboard by passing an NSURL array to writeObjects:
// (NSPasteboardTypeFileURL / public.file-url), allowing Finder and other apps to paste them as files.
//
// The logic follows Apple's official three-step pasteboard copy flow:
// 1) get the general pasteboard; 2) clearContents; 3) writeObjects: with NSPasteboardWriting objects.
// NSURL is a system-supported type. After writing file URLs, macOS automatically provides public.file-url,
// NSFilenamesPboardType, public.utf8-plain-text, and similar representations for Finder and older APIs.
//
// Official documentation and references:
//   - Pasteboard Programming Guide (macOS)
//     https://developer.apple.com/library/archive/documentation/Cocoa/Conceptual/PasteboardGuide106/Introduction/Introduction.html
//   - Copying to a Pasteboard (three-step flow and writeObjects:)
//     https://developer.apple.com/library/archive/documentation/Cocoa/Conceptual/PasteboardGuide106/Articles/pbCopying.html
//   - NSPasteboard
//     https://developer.apple.com/documentation/appkit/nspasteboard
//   - NSPasteboardWriting (implemented by NSURL, NSString, and others)
//     https://developer.apple.com/documentation/appkit/nspasteboardwriting
//
// The /* ... */ block below is inline Objective-C extracted and compiled by cgo; it is not commented-out code.

package util

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework AppKit -framework Foundation
#import <AppKit/AppKit.h>
#import <Foundation/Foundation.h>

// writeFilePathsToPasteboard writes paths to the general pasteboard using the Copying to a Pasteboard flow:
// 1) generalPasteboard; 2) clearContents; 3) writeObjects: with an NSURL array.
// NSURL conforms to NSPasteboardWriting, so macOS automatically provides public.file-url,
// NSFilenamesPboardType, and similar representations after writing.
// paths is an array of UTF-8 path strings, and count is its length.
static int writeFilePathsToPasteboard(const char** paths, int count) {
	if (count <= 0) return 0;
	NSMutableArray *arr = [NSMutableArray arrayWithCapacity:(NSUInteger)count];
	for (int i = 0; i < count; i++) {
		NSString *path = [NSString stringWithUTF8String:paths[i]];
		if (!path) continue;
		NSURL *url = [NSURL fileURLWithPath:path];
		if (url) [arr addObject:url];
	}
	// Return -2 when no valid path remains, such as invalid UTF-8 or values that cannot become NSURLs.
	if ([arr count] == 0) return -2;
	// Step 1: get the general pasteboard used by cut/copy/paste.
	NSPasteboard *pb = [NSPasteboard generalPasteboard];
	// Step 2: clear existing contents, then write only the current file paths.
	[pb clearContents];
	// Step 3: writeObjects: requires NSPasteboardWriting objects; NSURL already supports it.
	BOOL ok = [pb writeObjects:arr];
	return ok ? 0 : -1;
}
*/
import "C"

import (
	"errors"
	"unsafe"
)

// WriteFilePaths writes file paths to the system clipboard (general pasteboard), allowing Finder and similar
// apps to paste them as files. See Pasteboard Guide: Copying to a Pasteboard.
func WriteFilePaths(paths []string) error {
	if len(paths) == 0 {
		return nil
	}
	// Allocate a C char* array for passing into Objective-C.
	cPaths := make([]*C.char, len(paths))
	for i, p := range paths {
		cPaths[i] = C.CString(p)
	}
	defer func() {
		for _, c := range cPaths {
			C.free(unsafe.Pointer(c))
		}
	}()
	// Pass the first element address as const char**.
	ret := C.writeFilePathsToPasteboard((**C.char)(unsafe.Pointer(&cPaths[0])), C.int(len(paths)))
	switch ret {
	case 0:
		return nil
	case -2:
		return errors.New("no valid file paths to write (invalid UTF-8 or path)")
	default:
		return errors.New("failed to write file paths to pasteboard")
	}
}
