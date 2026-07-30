// DejaVu - Data snapshot and sync.
// Copyright (c) 2022-present, b3log.org
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
	"errors"
	"path/filepath"
	"sync"
	"time"

	"github.com/icha-senpai/note/third_party/forks/gulu"
	"github.com/icha-senpai/note/third_party/forks/dejavu/cloud"
	"github.com/icha-senpai/note/third_party/forks/dejavu/entity"
	"github.com/icha-senpai/note/third_party/forks/logging"
)

func (repo *Repo) SyncDownload(context map[string]interface{}) (mergeResult *MergeResult, trafficStat *TrafficStat, err error) {
	lock.Lock()
	defer lock.Unlock()

	err = repo.tryLockCloud(repo.DeviceID, context)
	if nil != err {
		return
	}
	defer repo.unlockCloud(context)

	mergeResult = &MergeResult{Time: time.Now()}
	trafficStat = &TrafficStat{m: &sync.Mutex{}}

	latest, err := repo.Latest()
	if nil != err {
		logging.LogErrorf("get latest failed: %s", err)
		return
	}

	length, cloudLatest, err := repo.downloadCloudLatest(context)
	if nil != err {
		if !errors.Is(err, cloud.ErrCloudObjectNotFound) {
			logging.LogErrorf("download cloud latest failed: %s", err)
			return
		}
	}
	trafficStat.DownloadFileCount++
	trafficStat.DownloadBytes += length
	trafficStat.APIGet++

	if cloudLatest.ID == latest.ID || "" == cloudLatest.ID {
		return
	}

	fetchFileIDs, err := repo.localNotFoundFiles(cloudLatest.Files)
	if nil != err {
		logging.LogErrorf("get local not found files failed: %s", err)
		return
	}

	length, fetchedFiles, err := repo.downloadCloudFilesPut(fetchFileIDs, context)
	if nil != err {
		logging.LogErrorf("download cloud files put failed: %s", err)
		return
	}
	trafficStat.DownloadFileCount += len(fetchFileIDs)
	trafficStat.DownloadBytes += length
	trafficStat.APIGet += trafficStat.DownloadFileCount

	cloudLatestFiles, err := repo.getFiles(cloudLatest.Files)
	if nil != err {
		logging.LogErrorf("get cloud latest files failed: %s", err)
		return
	}

	cloudChunkIDs := repo.getChunks(cloudLatestFiles)

	fetchChunkIDs, err := repo.localNotFoundChunks(cloudChunkIDs)
	if nil != err {
		logging.LogErrorf("get local not found chunks failed: %s", err)
		return
	}

	length, err = repo.downloadCloudChunksPut(fetchChunkIDs, context)
	trafficStat.DownloadBytes += length
	trafficStat.DownloadChunkCount += len(fetchChunkIDs)
	trafficStat.APIGet += trafficStat.DownloadChunkCount

	latestFiles, err := repo.getFiles(latest.Files)
	if nil != err {
		logging.LogErrorf("get latest files failed: %s", err)
		return
	}
	latestSync := repo.latestSync()
	latestSyncFiles, err := repo.getFiles(latestSync.Files)
	if nil != err {
		logging.LogErrorf("get latest sync files failed: %s", err)
		return
	}
	localUpserts, localRemoves := repo.diffUpsertRemove(latestFiles, latestSyncFiles, false)
	localChanged := 0 < len(localUpserts) || 0 < len(localRemoves)

	mergeResult.Upserts, mergeResult.Removes = repo.diffUpsertRemove(cloudLatestFiles, latestFiles, false)

	var fetchedFileIDs []string
	for _, fetchedFile := range fetchedFiles {
		fetchedFileIDs = append(fetchedFileIDs, fetchedFile.ID)
	}

	for _, localUpsert := range localUpserts {
		if nil != repo.getFile(mergeResult.Upserts, localUpsert) || nil != repo.getFile(mergeResult.Removes, localUpsert) {
			mergeResult.Conflicts = append(mergeResult.Conflicts, localUpsert)
			logging.LogInfof("sync download conflict [%s, %s, %s]", localUpsert.ID, localUpsert.Path, time.UnixMilli(localUpsert.Updated).Format("2006-01-02 15:04:05"))
		}
	}

	if 0 < len(mergeResult.Conflicts) {
		now := mergeResult.Time.Format("2006-01-02-150405")
		temp := filepath.Join(repo.TempPath, "repo", "sync", "conflicts", now)
		for i, file := range mergeResult.Conflicts {
			var checkoutTmp *entity.File
			checkoutTmp, err = repo.store.GetFile(file.ID)
			if nil != err {
				logging.LogErrorf("get file failed: %s", err)
				return
			}

			err = repo.checkoutFile(checkoutTmp, temp, i+1, len(mergeResult.Conflicts), context)
			if nil != err {
				logging.LogErrorf("checkout file failed: %s", err)
				return
			}

			absPath := filepath.Join(temp, checkoutTmp.Path)
			err = repo.genSyncHistory(now, file.Path, absPath)
			if nil != err {
				logging.LogErrorf("generate sync history failed: %s", err)
				err = ErrCloudGenerateConflictHistory
				return
			}
		}
	}

	err = repo.restoreFiles(mergeResult, context)
	if nil != err {
		logging.LogErrorf("restore files failed: %s", err)
		return
	}

	err = repo.mergeSync(mergeResult, localChanged, false, latest, cloudLatest, cloudChunkIDs, trafficStat, context)
	if nil != err {
		logging.LogErrorf("merge sync failed: %s", err)
		return
	}

	go repo.cloud.AddTraffic(&cloud.Traffic{
		DownloadBytes: trafficStat.DownloadBytes,
		APIGet:        trafficStat.APIGet,
	})

	gulu.File.RemoveEmptyDirs(repo.DataPath, removeEmptyDirExcludes...)
	return
}

func (repo *Repo) SyncUpload(context map[string]interface{}) (trafficStat *TrafficStat, err error) {
	lock.Lock()
	defer lock.Unlock()

	err = repo.tryLockCloud(repo.DeviceID, context)
	if nil != err {
		return
	}
	defer repo.unlockCloud(context)

	trafficStat = &TrafficStat{m: &sync.Mutex{}}

	latest, err := repo.Latest()
	if nil != err {
		logging.LogErrorf("get latest failed: %s", err)
		return
	}

	length, cloudLatest, err := repo.downloadCloudLatest(context)
	if nil != err {
		if !errors.Is(err, cloud.ErrCloudObjectNotFound) {
			logging.LogErrorf("download cloud latest failed: %s", err)
			return
		}
	}
	trafficStat.DownloadFileCount++
	trafficStat.DownloadBytes += length
	trafficStat.APIPut++

	if cloudLatest.ID == latest.ID {
		return
	}

	availableSize := repo.cloud.GetAvailableSize()
	if availableSize <= cloudLatest.Size || availableSize <= latest.Size {
		err = ErrCloudStorageSizeExceeded
		return
	}

	var uploadFiles []*entity.File
	for _, localFileID := range latest.Files {
		if !gulu.Str.Contains(localFileID, cloudLatest.Files) {
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

	//uploadChunkIDs, err = repo.cloud.GetChunks(uploadChunkIDs)
	//if nil != err {
	//	logging.LogErrorf("get cloud repo upload chunks failed: %s", err)
	//	return
	//}

	length, err = repo.uploadChunks(uploadChunkIDs, context)
	if nil != err {
		logging.LogErrorf("upload chunks failed: %s", err)
		return
	}
	trafficStat.UploadChunkCount += len(uploadChunkIDs)
	trafficStat.UploadBytes += length
	trafficStat.APIPut += trafficStat.UploadChunkCount

	length, err = repo.uploadFiles(uploadFiles, context)
	if nil != err {
		logging.LogErrorf("upload files failed: %s", err)
		return
	}
	trafficStat.UploadChunkCount += len(uploadFiles)
	trafficStat.UploadBytes += length
	trafficStat.APIPut += trafficStat.UploadChunkCount

	err = repo.updateCloudIndexes(latest, trafficStat, context)
	if nil != err {
		logging.LogErrorf("update cloud indexes failed: %s", err)
		return
	}

	err = repo.UpdateLatestSync(latest)
	if nil != err {
		logging.LogErrorf("update latest sync failed: %s", err)
		return
	}

	go repo.cloud.AddTraffic(&cloud.Traffic{
		UploadBytes: trafficStat.UploadBytes,
		APIPut:      trafficStat.APIPut,
	})
	return
}
