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

package model

import (
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/88250/gulu"
	"github.com/88250/lute/parse"
	"github.com/siyuan-note/httpclient"
	"github.com/siyuan-note/logging"
	"github.com/siyuan-note/siyuan/kernel/conf"
	"github.com/siyuan-note/siyuan/kernel/util"
)

var ErrFailedToConnectCloudServer = errors.New("failed to connect cloud server")

func CloudChatGPT(msg string, contextMsgs []string) (ret string, stop bool, err error) {
	if util.OfficialServicesUnavailable() {
		err = ErrFailedToConnectCloudServer
		return
	}

	if nil == Conf.GetUser() {
		return
	}

	payload := map[string]any{}
	var messages []map[string]any
	for _, contextMsg := range contextMsgs {
		messages = append(messages, map[string]any{
			"role":    "user",
			"content": contextMsg,
		})
	}
	messages = append(messages, map[string]any{
		"role":    "user",
		"content": msg,
	})
	payload["messages"] = messages

	requestResult := gulu.Ret.NewResult()
	request := httpclient.NewCloudRequest30s()
	_, err = request.
		SetSuccessResult(requestResult).
		SetCookies(&http.Cookie{Name: "symphony", Value: Conf.GetUser().UserToken}).
		SetBody(payload).
		Post(util.GetCloudServer() + "/apis/siyuan/ai/chatGPT")
	if err != nil {
		logging.LogErrorf("chat gpt failed: %s", err)
		err = ErrFailedToConnectCloudServer
		return
	}
	if 0 != requestResult.Code {
		err = errors.New(requestResult.Msg)
		stop = true
		return
	}

	data := requestResult.Data.(map[string]any)
	choices := data["choices"].([]any)
	if 1 > len(choices) {
		stop = true
		return
	}
	choice := choices[0].(map[string]any)
	message := choice["message"].(map[string]any)
	ret = message["content"].(string)

	if nil != choice["finish_reason"] {
		finishReason := choice["finish_reason"].(string)
		if "length" == finishReason {
			stop = false
		} else {
			stop = true
		}
	} else {
		stop = true
	}
	return
}

func SetCloudBlockReminder(id, data string, timed int64) (err error) {
	if util.OfficialServicesUnavailable() {
		return ErrFailedToConnectCloudServer
	}

	requestResult := gulu.Ret.NewResult()
	payload := map[string]any{"dataId": id, "data": data, "timed": timed}
	request := httpclient.NewCloudRequest30s()
	resp, err := request.
		SetSuccessResult(requestResult).
		SetBody(payload).
		SetCookies(&http.Cookie{Name: "symphony", Value: Conf.GetUser().UserToken}).
		Post(util.GetCloudServer() + "/apis/siyuan/calendar/setBlockReminder")
	if err != nil {
		logging.LogErrorf("set block reminder failed: %s", err)
		return ErrFailedToConnectCloudServer
	}

	if 401 == resp.StatusCode {
		err = errors.New(Conf.Language(31))
		return
	}

	if 0 != requestResult.Code {
		logging.LogErrorf("set block reminder failed: %s", requestResult.Msg)
		return errors.New(requestResult.Msg)
	}
	return
}

var uploadToken = ""
var uploadTokenTime int64

func LoadUploadToken() (err error) {
	if util.OfficialServicesUnavailable() {
		return ErrFailedToConnectCloudServer
	}

	now := time.Now().Unix()
	if 3600 >= now-uploadTokenTime {
		return
	}

	requestResult := gulu.Ret.NewResult()
	request := httpclient.NewCloudRequest30s()
	resp, err := request.
		SetSuccessResult(requestResult).
		SetCookies(&http.Cookie{Name: "symphony", Value: Conf.GetUser().UserToken}).
		Post(util.GetCloudServer() + "/apis/siyuan/upload/token")
	if err != nil {
		logging.LogErrorf("get upload token failed: %s", err)
		return ErrFailedToConnectCloudServer
	}

	if 401 == resp.StatusCode {
		err = errors.New(Conf.Language(31))
		return
	}

	if 0 != requestResult.Code {
		logging.LogErrorf("get upload token failed: %s", requestResult.Msg)
		return
	}

	resultData := requestResult.Data.(map[string]any)
	uploadToken = resultData["uploadToken"].(string)
	uploadTokenTime = now
	return
}

func RefreshCheckJob2H() {
	if util.OfficialServicesUnavailable() {
		return
	}

	go refreshUser()
	go refreshAnnouncement()
}

func RefreshCheckJob6H() {
	if util.OfficialServicesUnavailable() {
		return
	}

	go refreshCheckDownloadInstallPkg()
}

func refreshUser() {
	defer logging.Recover()

	if nil != Conf.GetUser() {
		time.Sleep(2 * time.Minute)
		if nil != Conf.GetUser() {
			RefreshUser(Conf.GetUser().UserToken)
		}
	}
}

func refreshCheckDownloadInstallPkg() {
	defer logging.Recover()

	time.Sleep(3 * time.Minute)
	checkDownloadInstallPkg()
}

func refreshAnnouncement() {
	defer logging.Recover()

	time.Sleep(1 * time.Minute)
	announcementConf := filepath.Join(util.UserHomeConfDir(), "announcement.json")
	var existingAnnouncements, newAnnouncements []*Announcement
	if gulu.File.IsExist(announcementConf) {
		data, err := os.ReadFile(announcementConf)
		if err != nil {
			logging.LogErrorf("read announcement conf failed: %s", err)
			return
		}
		if err = gulu.JSON.UnmarshalJSON(data, &existingAnnouncements); err != nil {
			logging.LogErrorf("unmarshal announcement conf failed: %s", err)
			os.Remove(announcementConf)
			return
		}
	}

	for _, announcement := range getAnnouncements() {
		var exist bool
		for _, existingAnnouncement := range existingAnnouncements {
			if announcement.Id == existingAnnouncement.Id {
				exist = true
				break
			}
		}
		if !exist {
			existingAnnouncements = append(existingAnnouncements, announcement)
			if Conf.CloudRegion == announcement.Region {
				newAnnouncements = append(newAnnouncements, announcement)
			}
		}
	}

	data, err := gulu.JSON.MarshalJSON(existingAnnouncements)
	if err != nil {
		logging.LogErrorf("marshal announcement conf failed: %s", err)
		return
	}
	if err = os.WriteFile(announcementConf, data, 0644); err != nil {
		logging.LogErrorf("write announcement conf failed: %s", err)
		return
	}

	for _, newAnnouncement := range newAnnouncements {
		util.PushMsg(fmt.Sprintf(Conf.Language(11), newAnnouncement.URL, newAnnouncement.Title), 0)
	}
}

func RefreshUser(token string) {
	if util.OfficialServicesUnavailable() {
		return
	}

	threeDaysAfter := util.CurrentTimeMillis() + 1000*60*60*24*3
	if "" == token {
		user := Conf.GetUser()
		if nil == user && "" != Conf.UserData {
			user = loadUserFromConf()
			if nil != user {
				Conf.SetUser(user)
			}
		}
		if nil == user {
			return
		}

		var tokenExpireTime int64
		tokenExpireTime, err := strconv.ParseInt(user.UserTokenExpireTime+"000", 10, 64)
		if err != nil {
			logging.LogErrorf("convert token expire time [%s] failed: %s", user.UserTokenExpireTime, err)
			util.PushErrMsg(Conf.Language(19), 5000)
			return
		}

		if threeDaysAfter > tokenExpireTime {
			token = user.UserToken
			goto Net
		}
		return
	}

Net:
	start := time.Now()
	user, err := getUser(token)
	if err != nil {
		if nil == Conf.GetUser() || errors.Is(err, errInvalidUser) {
			util.PushErrMsg(Conf.Language(19), 5000)
			return
		}

		if errors.Is(err, errRequestUserFailed) {
			util.PushErrMsg(Conf.Language(18), 5000)
			return
		}

		var tokenExpireTime int64
		tokenExpireTime, err = strconv.ParseInt(Conf.GetUser().UserTokenExpireTime+"000", 10, 64)
		if err != nil {
			logging.LogErrorf("convert token expire time [%s] failed: %s", Conf.GetUser().UserTokenExpireTime, err)
			util.PushErrMsg(Conf.Language(19), 5000)
			return
		}

		if threeDaysAfter > tokenExpireTime {
			util.PushErrMsg(Conf.Language(19), 5000)
			return
		}
		return
	}

	Conf.SetUser(user)
	data, _ := gulu.JSON.MarshalJSON(user)
	Conf.UserData = util.AESEncrypt(string(data))
	Conf.Save()

	if elapsed := time.Since(start).Milliseconds(); 3000 < elapsed {
		logging.LogInfof("get cloud user elapsed [%dms]", elapsed)
	}
}

func loadUserFromConf() *conf.User {
	if "" == Conf.UserData {
		return nil
	}

	data := util.AESDecrypt(Conf.UserData)
	data, _ = hex.DecodeString(string(data))
	user := &conf.User{}
	if err := gulu.JSON.UnmarshalJSON(data, &user); err == nil {
		return user
	}
	return nil
}

func RemoveCloudShorthands(ids []string) (err error) {
	if util.OfficialServicesUnavailable() {
		return ErrFailedToConnectCloudServer
	}

	result := map[string]any{}
	request := httpclient.NewCloudRequest30s()
	body := map[string]any{
		"ids": ids,
	}
	resp, err := request.
		SetSuccessResult(&result).
		SetCookies(&http.Cookie{Name: "symphony", Value: Conf.GetUser().UserToken}).
		SetBody(body).
		Post(util.GetCloudServer() + "/apis/siyuan/inbox/removeCloudShorthands")
	if err != nil {
		logging.LogErrorf("remove cloud shorthands failed: %s", err)
		err = ErrFailedToConnectCloudServer
		return
	}

	if 401 == resp.StatusCode {
		err = errors.New(Conf.Language(31))
		return
	}

	code := result["code"].(float64)
	if 0 != code {
		logging.LogErrorf("remove cloud shorthands failed: %s", result["msg"])
		err = errors.New(result["msg"].(string))
		return
	}
	return
}

func GetCloudShorthand(id string) (ret map[string]any, err error) {
	if util.OfficialServicesUnavailable() {
		err = ErrFailedToConnectCloudServer
		return
	}

	result := map[string]any{}
	request := httpclient.NewCloudRequest30s()
	resp, err := request.
		SetSuccessResult(&result).
		SetCookies(&http.Cookie{Name: "symphony", Value: Conf.GetUser().UserToken}).
		Post(util.GetCloudServer() + "/apis/siyuan/inbox/getCloudShorthand?id=" + id)
	if err != nil {
		logging.LogErrorf("get cloud shorthand failed: %s", err)
		err = ErrFailedToConnectCloudServer
		return
	}

	if 401 == resp.StatusCode {
		err = errors.New(Conf.Language(31))
		return
	}

	code := result["code"].(float64)
	if 0 != code {
		logging.LogErrorf("get cloud shorthand failed: %s", result["msg"])
		err = errors.New(result["msg"].(string))
		return
	}
	ret = result["data"].(map[string]any)
	t, _ := strconv.ParseInt(id, 10, 64)
	hCreated := util.Millisecond2Time(t)
	ret["hCreated"] = hCreated.Format("2006-01-02 15:04")

	md := ret["shorthandContent"].(string)
	ret["shorthandMd"] = md

	luteEngine := NewLute()
	luteEngine.SetFootnotes(true)
	tree := parse.Parse("", []byte(md), luteEngine.ParseOptions)
	luteEngine.RenderOptions.ProtyleMarkNetImg = false
	content := luteEngine.ProtylePreview(tree, luteEngine.RenderOptions, luteEngine.ParseOptions)
	ret["shorthandContent"] = content
	return
}

func GetCloudShorthands(page int) (result map[string]any, err error) {
	if util.OfficialServicesUnavailable() {
		err = ErrFailedToConnectCloudServer
		return
	}

	result = map[string]any{}
	request := httpclient.NewCloudRequest30s()
	resp, err := request.
		SetSuccessResult(&result).
		SetCookies(&http.Cookie{Name: "symphony", Value: Conf.GetUser().UserToken}).
		Post(util.GetCloudServer() + "/apis/siyuan/inbox/getCloudShorthands?p=" + strconv.Itoa(page))
	if err != nil {
		logging.LogErrorf("get cloud shorthands failed: %s", err)
		err = ErrFailedToConnectCloudServer
		return
	}

	if 401 == resp.StatusCode {
		err = errors.New(Conf.Language(31))
		return
	}

	code := result["code"].(float64)
	if 0 != code {
		logging.LogErrorf("get cloud shorthands failed: %s", result["msg"])
		err = errors.New(result["msg"].(string))
		return
	}

	luteEngine := NewLute()
	audioRegexp := regexp.MustCompile("<audio.*>.*</audio>")
	videoRegexp := regexp.MustCompile("<video.*>.*</video>")
	fileRegexp := regexp.MustCompile("\\[文件]\\(.*\\)")
	shorthands := result["data"].(map[string]any)["shorthands"].([]any)
	for _, item := range shorthands {
		shorthand := item.(map[string]any)
		id := shorthand["oId"].(string)
		t, _ := strconv.ParseInt(id, 10, 64)
		hCreated := util.Millisecond2Time(t)
		shorthand["hCreated"] = hCreated.Format("2006-01-02 15:04")

		desc := shorthand["shorthandDesc"].(string)
		desc = audioRegexp.ReplaceAllString(desc, " 语音 ")
		desc = videoRegexp.ReplaceAllString(desc, " 视频 ")
		desc = fileRegexp.ReplaceAllString(desc, " 文件 ")
		desc = strings.ReplaceAll(desc, "\n\n", "")
		desc = strings.TrimSpace(desc)
		shorthand["shorthandDesc"] = desc

		md := shorthand["shorthandContent"].(string)
		shorthand["shorthandMd"] = md
		tree := parse.Parse("", []byte(md), luteEngine.ParseOptions)
		luteEngine.RenderOptions.ProtyleMarkNetImg = false
		content := luteEngine.ProtylePreview(tree, luteEngine.RenderOptions, luteEngine.ParseOptions)
		shorthand["shorthandContent"] = content
	}
	return
}

var (
	errInvalidUser       = errors.New("invalid user")
	errRequestUserFailed = errors.New("request user failed")
)

func getUser(token string) (*conf.User, error) {
	if util.OfficialServicesUnavailable() {
		return nil, errRequestUserFailed
	}

	result := map[string]any{}
	request := httpclient.NewCloudRequest30s()
	resp, err := request.
		SetSuccessResult(&result).
		SetBody(map[string]string{"token": token}).
		Post(util.GetCloudServer() + "/apis/siyuan/user")
	if err != nil {
		logging.LogErrorf("get community user failed: %s", err)
		return nil, errRequestUserFailed
	}
	if http.StatusOK != resp.StatusCode {
		logging.LogErrorf("get community user failed: %d", resp.StatusCode)
		return nil, errRequestUserFailed
	}

	code := result["code"].(float64)
	if 0 != code {
		if 255 == code {
			return nil, errInvalidUser
		}
		logging.LogErrorf("get community user failed: %s", result["msg"])
		return nil, errRequestUserFailed
	}

	dataStr := result["data"].(string)
	data := util.AESDecrypt(dataStr)
	user := &conf.User{}
	if err = gulu.JSON.UnmarshalJSON(data, &user); err != nil {
		logging.LogErrorf("get community user failed: %s", err)
		return nil, errRequestUserFailed
	}
	return user, nil
}

func LogoutUser() {
	Conf.UserData = ""
	Conf.SetUser(nil)
	Conf.Save()
}
