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

package model

import (
	"testing"

	"github.com/siyuan-note/siyuan/kernel/conf"
	"github.com/siyuan-note/siyuan/kernel/util"
)

func TestIsPetalsEnabled(t *testing.T) {
	originalConf := Conf
	originalContainer := util.Container
	t.Cleanup(func() {
		Conf = originalConf
		util.Container = originalContainer
	})

	Conf = NewAppConf()
	Conf.Extensions = conf.NewExtensions()
	Conf.Extensions.PetalDisabled = false
	Conf.Extensions.Trust = false

	for _, container := range []string{util.ContainerAndroid, util.ContainerIOS, util.ContainerHarmony} {
		util.Container = container
		if !IsPetalsEnabled() {
			t.Fatalf("petals should be enabled on mobile container [%s] without extension trust", container)
		}
	}

	for _, container := range []string{util.ContainerStd, util.ContainerDocker} {
		util.Container = container
		if IsPetalsEnabled() {
			t.Fatalf("petals should be disabled on container [%s] without extension trust", container)
		}
	}

	Conf.Extensions.Trust = true
	if !IsPetalsEnabled() {
		t.Fatal("petals should be enabled after extension trust")
	}

	Conf.Extensions.PetalDisabled = true
	if IsPetalsEnabled() {
		t.Fatal("petals should be disabled explicitly")
	}
}
