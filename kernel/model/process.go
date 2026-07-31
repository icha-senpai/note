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

package model

import (
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/icha-senpai/note/kernel/util"
	"github.com/icha-senpai/note/third_party/forks/logging"
	goPS "github.com/mitchellh/go-ps"
)

func HandleSignal() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	s := <-c
	logging.LogInfof("received os signal [%s], exit kernel process now", s)
	Close(false, true, 1)
}

var (
	firstRunHookDesktopUIProcJob = true
)

func HookDesktopUIProcJob() {
	if !util.AttachUI || util.ContainerStd != util.Container || "dev" == util.Mode {
		return
	}

	if firstRunHookDesktopUIProcJob {

		time.Sleep(30 * time.Second)
		firstRunHookDesktopUIProcJob = false
		return
	}

	if 0 < util.CountSessions() {

		return
	}

	uiProcCount := getAttachedUIProcCount()
	if 0 < uiProcCount {
		return
	}

	logging.LogWarnf("no active UI proc, continue to check from attached ui processes after 15s")
	time.Sleep(15 * time.Second)
	uiProcCount = getAttachedUIProcCount()
	if 0 < uiProcCount {
		return
	}
	logging.LogWarnf("no active UI proc, continue to check from all processes after 15s")
	time.Sleep(15 * time.Second)
	uiProcCount = getUIProcCount()
	if 0 < uiProcCount {
		logging.LogInfof("active UI proc count [%d]", uiProcCount)
		return
	}

	logging.LogWarnf("confirmed no active UI proc, exit kernel process now")
	Close(false, true, 1)
}

var uiProcNames = []string{"scribli", "electron"}

func getAttachedUIProcCount() (ret int) {
	util.UIProcessIDs.Range(func(uiProcIDArg, _ any) bool {
		uiProcID, err := strconv.Atoi(uiProcIDArg.(string))
		if err != nil {
			logging.LogErrorf("invalid UI proc ID [%s]: %s", uiProcIDArg, err)
			return true
		}

		proc, err := goPS.FindProcess(uiProcID)
		if err != nil {
			logging.LogErrorf("find UI proc [%d] failed: %s", uiProcID, err)
			return true
		}

		if nil == proc {
			return true
		}

		procName := strings.ToLower(proc.Executable())
		uiProcOk := false
		for _, name := range uiProcNames {
			if uiProcOk = strings.Contains(procName, name); uiProcOk {
				break
			}
		}
		if uiProcOk {
			ret++
		}
		return true
	})
	return
}

func getUIProcCount() (ret int) {
	pid := os.Getpid()
	procs, _ := goPS.Processes()
	for _, proc := range procs {
		if proc.Pid() == pid {
			continue
		}

		procName := strings.ToLower(proc.Executable())
		uiProcOk := false
		for _, name := range uiProcNames {
			if uiProcOk = strings.Contains(procName, name); uiProcOk {
				break
			}
		}
		if uiProcOk {
			ret++
		}
	}
	return
}
