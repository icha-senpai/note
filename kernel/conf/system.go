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

package conf

import (
	"github.com/siyuan-note/siyuan/kernel/util"
)

type System struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	KernelVersion    string `json:"kernelVersion"`
	OS               string `json:"os"`
	OSPlatform       string `json:"osPlatform"`
	Container        string `json:"container"` // docker, android, ios, harmony, std
	IsMicrosoftStore bool   `json:"isMicrosoftStore"`
	IsInsider        bool   `json:"isInsider"`

	HomeDir      string `json:"homeDir"`
	WorkspaceDir string `json:"workspaceDir"`
	AppDir       string `json:"appDir"`
	ConfDir      string `json:"confDir"`
	DataDir      string `json:"dataDir"`

	NetworkServe    bool          `json:"networkServe"`
	NetworkServeTLS bool          `json:"networkServeTLS"`
	NetworkProxy    *NetworkProxy `json:"networkProxy"`

	DownloadInstallPkg bool `json:"downloadInstallPkg"`
	AutoLaunch2        int  `json:"autoLaunch2"`
	LockScreenMode     int  `json:"lockScreenMode"`

	DisabledFeatures []string `json:"disabledFeatures"`

	MicrosoftDefenderExcluded bool `json:"microsoftDefenderExcluded"`

	SafeMode bool `json:"safeMode"`
}

const (
	OnboardingPending         = "pending"
	OnboardingNotebookCreated = "notebook-created"
	OnboardingCompleted       = "completed"
)

type Onboarding struct {
	State      string `json:"state"`
	NewUser    bool   `json:"newUser"`
	Dismissed  bool   `json:"dismissed"`
	NotebookID string `json:"notebookID"`
	DocumentID string `json:"documentID"`
}

func NewSystem() *System {
	return &System{
		ID:                 util.GetDeviceID(),
		Name:               util.GetDeviceName(),
		KernelVersion:      util.Ver,
		NetworkProxy:       &NetworkProxy{},
		DownloadInstallPkg: false,
	}
}

type NetworkProxy struct {
	Scheme string `json:"scheme"`
	Host   string `json:"host"`
	Port   string `json:"port"`
}

func (np *NetworkProxy) String() string {
	if "" == np.Scheme {
		return ""
	}
	return np.Scheme + "://" + np.Host + ":" + np.Port
}
