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

package dejavu

import (
	"github.com/icha-senpai/note/third_party/forks/dejavu/cloud"
	"os"
	"path"
	"strings"

	"github.com/icha-senpai/note/third_party/forks/gulu"
	"github.com/icha-senpai/note/third_party/forks/dejavu/entity"
	"github.com/icha-senpai/note/third_party/forks/logging"
)

func (repo *Repo) DownloadIndex(id string, context map[string]interface{}) (downloadFileCount, downloadChunkCount int, downloadBytes int64, err error) {
	lock.Lock()
	defer lock.Unlock()

	downloadFileCount, downloadChunkCount, downloadBytes, err = repo.downloadIndex(id, context)
	return
}

func (repo *Repo) DownloadTagIndex(tag, id string, context map[string]interface{}) (downloadFileCount, downloadChunkCount int, downloadBytes int64, err error) {
	lock.Lock()
	defer lock.Unlock()

	downloadFileCount, downloadChunkCount, downloadBytes, err = repo.downloadIndex(id, context)

	err = repo.AddTag(id, tag)
	if nil != err {
		logging.LogErrorf("add tag failed: %s", err)
		return
	}

	return
}

func (repo *Repo) downloadIndex(id string, context map[string]interface{}) (downloadFileCount, downloadChunkCount int, downloadBytes int64, err error) {
	length, index, err := repo.downloadCloudIndex(id, context)
	if nil != err {
		logging.LogErrorf("download cloud index failed: %s", err)
		return
	}
	downloadFileCount++
	downloadBytes += length
	apiGet := 1

	fetchFileIDs, err := repo.localNotFoundFiles(index.Files)
	if nil != err {
		logging.LogErrorf("get local not found files failed: %s", err)
		return
	}

	length, fetchedFiles, err := repo.downloadCloudFilesPut(fetchFileIDs, context)
	if nil != err {
		logging.LogErrorf("download cloud files put failed: %s", err)
		return
	}
	downloadBytes += length
	downloadFileCount = len(fetchFileIDs)
	apiGet += downloadFileCount

	cloudChunkIDs := repo.getChunks(fetchedFiles)

	fetchChunkIDs, err := repo.localNotFoundChunks(cloudChunkIDs)
	if nil != err {
		logging.LogErrorf("get local not found chunks failed: %s", err)
		return
	}

	length, err = repo.downloadCloudChunksPut(fetchChunkIDs, context)
	downloadBytes += length
	downloadChunkCount = len(fetchChunkIDs)
	apiGet += downloadChunkCount

	err = repo.store.PutIndex(index)
	if nil != err {
		logging.LogErrorf("put index failed: %s", err)
		return
	}

	go repo.cloud.AddTraffic(&cloud.Traffic{DownloadBytes: downloadBytes, APIGet: apiGet})
	return
}

func (repo *Repo) UploadTagIndex(tag, id string, context map[string]interface{}) (uploadFileCount, uploadChunkCount int, uploadBytes int64, err error) {
	lock.Lock()
	defer lock.Unlock()

	uploadFileCount, uploadChunkCount, uploadBytes, err = repo.uploadTagIndex(tag, id, context)
	if e, ok := err.(*os.PathError); ok && os.IsNotExist(err) {
		p := e.Path
		if !strings.Contains(p, "objects") {
			return
		}

		logging.LogErrorf("upload tag index failed: %s", err)
		err = ErrRepoFatal
	}
	return
}

func (repo *Repo) uploadTagIndex(tag, id string, context map[string]interface{}) (uploadFileCount, uploadChunkCount int, uploadBytes int64, err error) {
	index, err := repo.store.GetIndex(id)
	if nil != err {
		logging.LogErrorf("get index failed: %s", err)
		return
	}

	availableSize := repo.cloud.GetAvailableSize()
	if availableSize <= index.Size {
		err = ErrCloudStorageSizeExceeded
		return
	}

	cloudRepoSize, cloudBackupCount, err := repo.getCloudRepoStat()
	if nil != err {
		logging.LogErrorf("get cloud repo stat failed: %s", err)
		return
	}
	if 12 <= cloudBackupCount {
		err = ErrCloudBackupCountExceeded
		return
	}

	if availableSize <= cloudRepoSize+index.Size {
		err = ErrCloudStorageSizeExceeded
		return
	}

	cloudFileIDs, refs, err := repo.cloud.GetRefsFiles()
	if nil != err {
		logging.LogErrorf("get cloud repo refs files failed: %s", err)
		return
	}
	apiGet := len(refs) + 1

	var uploadFiles []*entity.File
	for _, localFileID := range index.Files {
		if !gulu.Str.Contains(localFileID, cloudFileIDs) {
			var uploadFile *entity.File
			uploadFile, err = repo.store.GetFile(localFileID)
			if nil != err {
				logging.LogErrorf("get file failed: %s", err)
				return
			}
			uploadFiles = append(uploadFiles, uploadFile)
		}
	}

	uploadChunkIDs := repo.getChunks(uploadFiles)

	uploadChunkIDs, err = repo.cloud.GetChunks(uploadChunkIDs)
	if nil != err {
		logging.LogErrorf("get cloud repo upload chunks failed: %s", err)
		return
	}
	apiGet += len(uploadChunkIDs)

	length, err := repo.uploadChunks(uploadChunkIDs, context)
	if nil != err {
		logging.LogErrorf("upload chunks failed: %s", err)
		return
	}
	uploadChunkCount = len(uploadChunkIDs)
	uploadBytes += length
	apiPut := uploadChunkCount

	length, err = repo.uploadFiles(uploadFiles, context)
	if nil != err {
		logging.LogErrorf("upload files failed: %s", err)
		return
	}
	uploadFileCount = len(uploadFiles)
	uploadBytes += length
	apiPut += uploadFileCount

	length, err = repo.uploadIndex(index, context)
	uploadFileCount++
	uploadBytes += length
	apiPut++

	length, err = repo.updateCloudRef("refs/tags/"+tag, context)
	uploadFileCount++
	uploadBytes += length
	apiPut++

	go repo.cloud.AddTraffic(&cloud.Traffic{UploadBytes: uploadBytes, APIGet: apiGet, APIPut: apiPut})
	return
}

func (repo *Repo) getCloudRepoStat() (repoSize int64, backupCount int, err error) {
	repoStat, err := repo.cloud.GetStat()
	if nil != err {
		return
	}

	repoSize = repoStat.Sync.Size + repoStat.Backup.Size
	backupCount = repoStat.Backup.Count
	return
}

func (repo *Repo) RemoveCloudRepoTag(tag string) (err error) {
	key := path.Join("refs", "tags", tag)
	return repo.cloud.RemoveObject(key)
}
