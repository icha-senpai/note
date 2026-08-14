// Scribli - Refactor your thinking
// Copyright (c) 2020-present, Scribli
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package tools

import (
	"regexp"

	"github.com/icha-senpai/note/kernel/util"
)

var kramdownAttrRE = regexp.MustCompile(`\s([A-Za-z0-9_-]+)="([^"]*)"`)

func createdFromID(id string) string {
	if len(id) < 14 {
		return ""
	}
	return util.TimeFromID(id)
}

func ialUpdated(ial map[string]string, kramdown string) string {
	if ial != nil && ial["updated"] != "" {
		return ial["updated"]
	}
	return kramdownAttr(kramdown, "updated")
}

func kramdownAttr(kramdown, key string) string {
	matches := kramdownAttrRE.FindAllStringSubmatch(kramdown, -1)
	for _, match := range matches {
		if len(match) == 3 && match[1] == key {
			return match[2]
		}
	}
	return ""
}
