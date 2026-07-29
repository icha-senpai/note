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

type Sync struct {
	CloudName           string  `json:"cloudName"`
	Enabled             bool    `json:"enabled"`
	Perception          bool    `json:"perception"`
	Mode                int     `json:"mode"`
	Interval            int     `json:"interval"`
	Synced              int64   `json:"synced"`
	Stat                string  `json:"stat"`
	GenerateConflictDoc bool    `json:"generateConflictDoc"`
	Provider            int     `json:"provider"`
	S3                  *S3     `json:"s3"`
	WebDAV              *WebDAV `json:"webdav"`
	Local               *Local  `json:"local"`
}

func NewSync() *Sync {
	return &Sync{
		CloudName:           "main",
		Enabled:             false,
		Perception:          false,
		Mode:                1,
		GenerateConflictDoc: false,
		Provider:            ProviderLocal,
		Interval:            30,
	}
}

type S3 struct {
	Endpoint       string `json:"endpoint"`
	AccessKey      string `json:"accessKey"` // Access Key
	SecretKey      string `json:"secretKey"` // Secret Key
	Bucket         string `json:"bucket"`
	Region         string `json:"region"`
	PathStyle      bool   `json:"pathStyle"`
	SkipTlsVerify  bool   `json:"skipTlsVerify"`
	Timeout        int    `json:"timeout"`
	ConcurrentReqs int    `json:"concurrentReqs"`
}

type WebDAV struct {
	Endpoint       string `json:"endpoint"`
	Username       string `json:"username"`
	Password       string `json:"password"`
	SkipTlsVerify  bool   `json:"skipTlsVerify"`
	Timeout        int    `json:"timeout"`
	ConcurrentReqs int    `json:"concurrentReqs"`
}

type Local struct {
	Endpoint       string `json:"endpoint"`
	Timeout        int    `json:"timeout"`
	ConcurrentReqs int    `json:"concurrentReqs"`
}

const (
	ProviderSiYuan = 0
	ProviderS3     = 2
	ProviderWebDAV = 3
	ProviderLocal  = 4
)

func ProviderToStr(provider int) string {
	switch provider {
	case ProviderSiYuan:
		return "Scribli"
	case ProviderS3:
		return "S3"
	case ProviderWebDAV:
		return "WebDAV"
	case ProviderLocal:
		return "Local File System"
	}
	return "Unknown"
}
