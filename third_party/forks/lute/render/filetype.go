// Copyright (c) 2019-present, Scribli
//
// Lute is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//         http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package render

import "strings"

func isFileExt(pos, length int, runes *[]rune) bool {
	max := pos + maxCommonFileTypeLen
	if max > length {
		max = length
	}

	ext := string((*runes)[pos:max])
	for j := 0; j < commonFileTypesLen; j++ {
		if strings.HasPrefix(ext, commonFileTypes[j]) {
			return true
		}
	}
	return false
}

var commonFileTypesLen = len(commonFileTypes)
var maxCommonFileTypeLen = 10 // textbundle

var commonFileTypes = []string{

	"jpg",
	"png",
	"gif",
	"webp",
	"cr2",
	"tif",
	"bmp",
	"heif",
	"jxr",
	"psd",
	"ico",
	"dwg",


	"mp4",
	"m4v",
	"mkv",
	"webm",
	"mov",
	"avi",
	"wmv",
	"mpg",
	"flv",
	"3gp",


	"mid",
	"mp3",
	"m4a",
	"ogg",
	"flac",
	"wav",
	"amr",
	"aac",


	"epub",
	"zip",
	"tar",
	"rar",
	"gz",
	"bz2",
	"7z",
	"xz",
	"pdf",
	"exe",
	"swf",
	"rtf",
	"iso",
	"eot",
	"ps",
	"sqli",
	"nes",
	"crx",
	"cab",
	"deb",
	"ar",
	"Z",
	"lz",
	"rpm",
	"elf",
	"dcm",


	"doc",
	"docx",
	"xls",
	"xlsx",
	"ppt",
	"pptx",
	"md",
	"txt",


	"woff",
	"woff2",
	"ttf",
	"otf",


	"wasm",
	"exe",


	"html",
	"js",
	"css",
	"go",
	"java",


	"textbundle",
}
