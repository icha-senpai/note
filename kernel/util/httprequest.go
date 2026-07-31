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

package util

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/icha-senpai/note/third_party/forks/httpclient"
	"github.com/imroc/req/v3"
)

const (
	maxHTTPRequestBytes     = 5 * 1024 * 1024
	maxHTTPRequestFileBytes = 10 * 1024 * 1024
	maxHTTPRequestChars     = 50000
)

func CheckHostSSRF(host string) error {
	ips, err := net.LookupIP(host)
	if err != nil {
		return errors.New("failed to resolve host: " + err.Error())
	}
	for _, ip := range ips {
		if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsPrivate() || ip.IsUnspecified() {
			return errors.New("access to private/internal IP is prohibited")
		}
	}
	return nil
}

func HTTPRequest(method, rawURL string, headers map[string]string, body string) (statusCode int, contentType string, text string, err error) {
	u, err := url.Parse(rawURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
		return 0, "", "", errors.New("URL must start with http:// or https://")
	}
	if u.Host == "" {
		return 0, "", "", errors.New("URL has no host")
	}

	if serr := CheckHostSSRF(u.Hostname()); serr != nil {
		return 0, "", "", serr
	}

	method = strings.ToUpper(strings.TrimSpace(method))
	if method == "" {
		method = "GET"
	}

	request := httpclient.NewBrowserRequest()
	for k, v := range headers {
		request.SetHeader(k, v)
	}
	if body != "" && method != "GET" && method != "HEAD" {
		request.SetBody(body)
	}

	resp, err := sendByMethod(request, method, rawURL)
	if err != nil {
		return 0, "", "", errors.New("request failed: " + err.Error())
	}
	if resp == nil {
		return 0, "", "", errors.New("nil response")
	}
	defer resp.Body.Close()

	statusCode = resp.StatusCode
	contentType = resp.Header.Get("Content-Type")

	maxReadBytes := int64(maxHTTPRequestBytes)
	if !isTextContentType(contentType) {
		maxReadBytes = maxHTTPRequestFileBytes
	}

	if resp.ContentLength > maxReadBytes {
		return statusCode, contentType, "", errors.New("response too large")
	}

	respBody, rerr := io.ReadAll(io.LimitReader(resp.Body, maxReadBytes))
	if rerr != nil {
		return statusCode, contentType, "", errors.New("read body failed: " + rerr.Error())
	}

	if !isTextContentType(contentType) {
		importDir := filepath.Join(TempDir, "import")
		if merr := os.MkdirAll(importDir, 0755); merr != nil {
			return statusCode, contentType, "", errors.New("create import dir failed: " + merr.Error())
		}
		filename := extractFilename(rawURL, contentType)
		filePath := filepath.Join(importDir, filename)
		if werr := os.WriteFile(filePath, respBody, 0644); werr != nil {
			return statusCode, contentType, "", errors.New("write file failed: " + werr.Error())
		}
		return statusCode, contentType, fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(respBody)), nil
	}

	return statusCode, contentType, truncateRunes(string(respBody), maxHTTPRequestChars), nil
}

func sendByMethod(request *req.Request, method, rawURL string) (*req.Response, error) {
	switch method {
	case "GET", "":
		return request.Get(rawURL)
	case "POST":
		return request.Post(rawURL)
	case "PUT":
		return request.Put(rawURL)
	case "DELETE":
		return request.Delete(rawURL)
	case "PATCH":
		return request.Patch(rawURL)
	default:
		return nil, fmt.Errorf("unsupported method: %s", method)
	}
}

func isTextContentType(contentType string) bool {
	ct := strings.ToLower(strings.TrimSpace(strings.SplitN(contentType, ";", 2)[0]))
	if ct == "" {
		return false
	}
	if strings.HasPrefix(ct, "text/") {
		return true
	}
	switch ct {
	case "application/json", "application/xml":
		return true
	}
	if strings.HasPrefix(ct, "application/") && (strings.HasSuffix(ct, "+json") || strings.HasSuffix(ct, "+xml")) {
		return true
	}
	return false
}
