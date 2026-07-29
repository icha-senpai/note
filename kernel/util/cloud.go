// SiYuan - Refactor your thinking
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

package util

var CurrentCloudRegion = 1

func OfficialServicesUnavailable() bool {
	return OfficialServicesDisabled || OfflineMode
}

func OfficialServicesError() error {
	if OfflineMode {
		return ErrOfflineMode
	}
	return ErrOfficialServicesDisabled
}

func IsChinaCloud() bool {
	return 0 == CurrentCloudRegion
}

func GetCloudServer() string {
	if 0 == CurrentCloudRegion {
		return chinaServer
	}
	return northAmericaServer
}

func GetCloudWebSocketServer() string {
	if 0 == CurrentCloudRegion {
		return chinaWebSocketServer
	}
	return northAmericaWebSocketServer
}

func GetCloudSyncServer() string {
	if 0 == CurrentCloudRegion {
		return chinaSyncServer
	}
	return northAmericaSyncServer
}

func GetCloudAssetsServer() string {
	if 0 == CurrentCloudRegion {
		return chinaCloudAssetsServer
	}
	return northAmericaCloudAssetsServer
}

func GetCloudAccountServer() string {
	if 0 == CurrentCloudRegion {
		return chinaAccountServer
	}
	return northAmericaAccountServer
}

func GetCloudForumAssetsServer() string {
	if 0 == CurrentCloudRegion {
		return chinaForumAssetsServer
	}
	return northAmericaForumAssetsServer
}

const (
	chinaServer            = ""
	chinaWebSocketServer   = ""
	chinaSyncServer        = ""
	chinaCloudAssetsServer = ""
	chinaAccountServer     = ""
	chinaForumAssetsServer = ""

	northAmericaServer            = ""
	northAmericaWebSocketServer   = ""
	northAmericaSyncServer        = ""
	northAmericaCloudAssetsServer = ""
	northAmericaAccountServer     = ""
	northAmericaForumAssetsServer = ""
)
