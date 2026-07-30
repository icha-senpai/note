// DejaVu - Data snapshot and sync.
// Copyright (c) 2022-present, Scribli
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

package cloud

import (
	"errors"
	"strings"

	"github.com/icha-senpai/note/third_party/forks/github/dgraph-io/ristretto"
	"github.com/icha-senpai/note/third_party/forks/dejavu/entity"
	"github.com/icha-senpai/note/third_party/forks/github/klauspost/compress/zstd"
)

type Conf struct {
	Dir      string
	UserID   string
	RepoPath string
	Endpoint string
	Extras   map[string]interface{}

	S3 *ConfS3

	WebDAV *ConfWebDAV

	Local *ConfLocal

	AvailableSize int64 // Storage size budget in bytes for user-controlled backends that do not report free space.
}

type ConfS3 struct {
	Endpoint       string
	AccessKey      string // Access Key
	SecretKey      string // Secret Key
	Region         string
	Bucket         string
	PathStyle      bool
	SkipTlsVerify  bool
	Timeout        int
	ConcurrentReqs int
}

type ConfWebDAV struct {
	Endpoint       string
	Username       string
	Password       string
	SkipTlsVerify  bool
	Timeout        int
	ConcurrentReqs int
}

type ConfLocal struct {
	//
	//	"D:/path/to/repos/directory" // Windows
	//	"/path/to/repos/directory"   // Unix
	Endpoint       string
	Timeout        int
	ConcurrentReqs int
}

type Cloud interface {

	CreateRepo(name string) (err error)

	RemoveRepo(name string) (err error)

	GetRepos() (repos []*Repo, size int64, err error)

	UploadObject(filePath string, overwrite bool) (length int64, err error)

	UploadBytes(filePath string, data []byte, overwrite bool) (length int64, err error)

	DownloadObject(filePath string) (data []byte, err error)

	RemoveObject(filePath string) (err error)

	GetTags() (tags []*Ref, err error)

	GetIndexes(page int) (indexes []*entity.Index, pageCount, totalCount int, err error)

	GetRefsFiles() (fileIDs []string, refs []*Ref, err error)

	GetChunks(checkChunkIDs []string) (chunkIDs []string, err error)

	GetStat() (stat *Stat, err error)

	GetConf() *Conf

	GetAvailableSize() (size int64)

	AddTraffic(traffic *Traffic)

	ListObjects(pathPrefix string) (objInfos map[string]*entity.ObjectInfo, err error)

	GetIndex(id string) (index *entity.Index, err error)

	GetConcurrentReqs() int
}

type Traffic struct {
	UploadBytes   int64
	DownloadBytes int64
	APIGet        int
	APIPut        int
}

type Stat struct {
	Sync      *StatSync   `json:"sync"`
	Backup    *StatBackup `json:"backup"`
	AssetSize int64       `json:"assetSize"`
	RepoCount int         `json:"repoCount"`
}

type Repo struct {
	Name    string `json:"name"`
	Size    int64  `json:"size"`
	Updated string `json:"updated"`
}

type Ref struct {
	Name    string `json:"name"`
	ID      string `json:"id"`
	Updated string `json:"updated"`
}

type StatSync struct {
	Size      int64  `json:"size"`
	FileCount int    `json:"fileCount"`
	Updated   string `json:"updated"`
}

type StatBackup struct {
	Count     int    `json:"count"`
	Size      int64  `json:"size"`
	FileCount int    `json:"fileCount"`
	Updated   string `json:"updated"`
}

type Indexes struct {
	Indexes []*Index `json:"indexes"`
}

type Index struct {
	ID         string `json:"id"`
	SystemID   string `json:"systemID"`
	SystemName string `json:"systemName"`
	SystemOS   string `json:"systemOS"`
}

type BaseCloud struct {
	*Conf
	Cloud
}

func (baseCloud *BaseCloud) CreateRepo(name string) (err error) {
	err = ErrUnsupported
	return
}

func (baseCloud *BaseCloud) RemoveRepo(name string) (err error) {
	err = ErrUnsupported
	return
}

func (baseCloud *BaseCloud) GetRepos() (repos []*Repo, size int64, err error) {
	err = ErrUnsupported
	return
}

func (baseCloud *BaseCloud) UploadObject(filePath string, overwrite bool) (length int64, err error) {
	err = ErrUnsupported
	return
}

func (baseCloud *BaseCloud) UploadBytes(filePath string, data []byte, overwrite bool) (length int64, err error) {
	err = ErrUnsupported
	return
}

func (baseCloud *BaseCloud) DownloadObject(filePath string) (data []byte, err error) {
	err = ErrUnsupported
	return
}

func (baseCloud *BaseCloud) RemoveObject(filePath string) (err error) {
	err = ErrUnsupported
	return
}

func (baseCloud *BaseCloud) GetTags() (tags []*Ref, err error) {
	err = ErrUnsupported
	return
}

func (baseCloud *BaseCloud) GetIndexes(page int) (indexes []*entity.Index, pageCount, totalCount int, err error) {
	err = ErrUnsupported
	return
}

func (baseCloud *BaseCloud) GetRefsFiles() (fileIDs []string, refs []*Ref, err error) {
	err = ErrUnsupported
	return
}

func (baseCloud *BaseCloud) GetChunks(checkChunkIDs []string) (chunkIDs []string, err error) {
	err = ErrUnsupported
	return
}

func (baseCloud *BaseCloud) GetStat() (stat *Stat, err error) {
	stat = &Stat{
		Sync:   &StatSync{},
		Backup: &StatBackup{},
	}
	return
}

func (baseCloud *BaseCloud) ListObjects(pathPrefix string) (objInfos map[string]*entity.ObjectInfo, err error) {
	err = ErrUnsupported
	return
}

func (baseCloud *BaseCloud) GetIndex(id string) (index *entity.Index, err error) {
	err = ErrUnsupported
	return
}

func (baseCloud *BaseCloud) GetConcurrentReqs() int {
	return 8
}

func (baseCloud *BaseCloud) GetConf() *Conf {
	return baseCloud.Conf
}

func (baseCloud *BaseCloud) GetAvailableSize() int64 {
	return baseCloud.Conf.AvailableSize
}

func (baseCloud *BaseCloud) AddTraffic(*Traffic) {
	return
}

var (
	ErrUnsupported             = errors.New("not supported yet")
	ErrCloudObjectNotFound     = errors.New("cloud object not found")
	ErrCloudAuthFailed         = errors.New("cloud account auth failed")
	ErrCloudServiceUnavailable = errors.New("cloud service unavailable")
	ErrSystemTimeIncorrect     = errors.New("system time incorrect")
	ErrDeprecatedVersion       = errors.New("deprecated version")
	ErrCloudCheckFailed        = errors.New("cloud check failed")
	ErrCloudForbidden          = errors.New("cloud forbidden")
	ErrCloudTooManyRequests    = errors.New("cloud too many requests")
	ErrDecryptFailed           = errors.New("decrypt failed")
)

func IsValidCloudDirName(cloudDirName string) bool {
	if 63 < len(cloudDirName) || 1 > len(cloudDirName) {
		return false
	}

	chars := []byte{'~', '`', '!', '@', '#', '$', '%', '^', '&', '*', '(', ')', '+', '=',
		'[', ']', '{', '}', '\\', '|', ';', ':', '\'', '"', '<', ',', '>', '.', '?', '/', ' '}
	var charsStr string
	for _, char := range chars {
		charsStr += string(char)
	}

	if strings.ContainsAny(cloudDirName, charsStr) {
		return false
	}

	tmp := stripCtlFromUTF8(cloudDirName)
	return tmp == cloudDirName
}

func stripCtlFromUTF8(str string) string {
	return strings.Map(func(r rune) rune {
		if r >= 32 && r != 127 {
			return r
		}
		return -1
	}, str)
}

var (
	compressDecoder *zstd.Decoder
	cache           *ristretto.Cache
)

func init() {
	var err error
	compressDecoder, err = zstd.NewReader(nil, zstd.WithDecoderMaxMemory(16*1024*1024*1024))
	if nil != err {
		panic(err)
	}

	cache, err = ristretto.NewCache(&ristretto.Config{
		NumCounters: 200000,
		MaxCost:     1000 * 1000 * 32,
		BufferItems: 64,
	})
	if nil != err {
		panic(err)
	}
}

type objectInfo struct {
	Key     string `json:"key"`
	Size    int64  `json:"size"`
	Updated string `json:"updated"`
}
