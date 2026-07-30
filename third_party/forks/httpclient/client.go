// HttpClient - HTTP client for Scribli.
// Copyright (c) 2022-present, Scribli
//
// HttpClient is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//         http://license.coscl.org.cn/MulanPSL2
//
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
//
// See the Mulan PSL v2 for more details.

package httpclient

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/icha-senpai/note/third_party/forks/github/imroc/req/v3"
	"github.com/icha-senpai/note/third_party/forks/external/golang.org/x/net/http/httpproxy"
)

var (
	browserClient, cloudClientTimeout30s, cloudFileClientTimeout2Min *req.Client

	browserUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/107.0.0.0 Safari/537.36"
	scribliUserAgent = "Scribli/0.0.0"
)

func CloseIdleConnections() {
	if nil != browserClient {
		browserClient.GetClient().CloseIdleConnections()
	}
	if nil != cloudClientTimeout30s {
		cloudClientTimeout30s.GetClient().CloseIdleConnections()
	}
	if nil != cloudFileClientTimeout2Min {
		cloudFileClientTimeout2Min.GetClient().CloseIdleConnections()
	}
}

func GetCloudFileClient2Min() *http.Client {
	if nil == cloudFileClientTimeout2Min {
		newCloudFileClient2m()
	}
	return cloudFileClientTimeout2Min.GetClient()
}

func SetUserAgent(scribliUA string) {
	scribliUserAgent = scribliUA
}

// UserAgentTransport injects the Scribli User-Agent on outbound requests.
// It replaces any existing UA so user-controlled sync providers see a stable app identity.
type UserAgentTransport struct {
	Base http.RoundTripper
}

func (t *UserAgentTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	base := t.Base
	if base == nil {
		base = http.DefaultTransport
	}
	clone := req.Clone(req.Context())
	clone.Header.Set("User-Agent", scribliUserAgent)
	return base.RoundTrip(clone)
}

// NewUserAgentRoundTripper returns a RoundTripper that injects the Scribli UA.
func NewUserAgentRoundTripper(base http.RoundTripper) http.RoundTripper {
	return &UserAgentTransport{Base: base}
}

// NewUserAgentClient returns an http.Client that injects the Scribli UA.
func NewUserAgentClient(base http.RoundTripper) *http.Client {
	if base == nil {
		base = NewTransport(false)
	}
	return &http.Client{Transport: &UserAgentTransport{Base: base}}
}

func NewBrowserRequest() (ret *req.Request) {
	if nil == browserClient {
		browserClient = req.C().
			SetUserAgent(browserUserAgent).
			SetTimeout(30 * time.Second).
			DisableInsecureSkipVerify().
			SetProxy(ProxyFromEnvironment)
	}
	ret = browserClient.R()
	ret.SetRetryCount(1).SetRetryFixedInterval(3 * time.Second)
	return
}

func NewCloudFileRequest2m() *req.Request {
	if nil == cloudFileClientTimeout2Min {
		newCloudFileClient2m()
	}
	return cloudFileClientTimeout2Min.R()
}

func newCloudFileClient2m() {
	cloudFileClientTimeout2Min = req.C().
		EnableForceHTTP1().
		SetCommonHeader("Cache-Control", "no-cache, no-store, must-revalidate").
		SetCommonHeader("Pragma", "no-cache").
		SetCommonHeader("Expires", "0").
		SetUserAgent(scribliUserAgent).
		SetTimeout(2 * time.Minute).
		SetCommonRetryCount(1).
		SetCommonRetryFixedInterval(3 * time.Second).
		SetCommonRetryCondition(retryCondition).
		DisableInsecureSkipVerify().
		SetProxy(ProxyFromEnvironment)
}

func NewCloudRequest30s() *req.Request {
	if nil == cloudClientTimeout30s {
		cloudClientTimeout30s = req.C().
			SetUserAgent(scribliUserAgent).
			SetTimeout(30 * time.Second).
			SetCommonRetryCount(1).
			SetCommonRetryFixedInterval(3 * time.Second).
			SetCommonRetryCondition(retryCondition).
			DisableInsecureSkipVerify().
			SetProxy(ProxyFromEnvironment)
	}
	return cloudClientTimeout30s.R()
}

func retryCondition(resp *req.Response, err error) bool {
	if nil != err {
		return true
	}
	if nil == resp || nil == resp.Response {
		return true
	}
	if 503 == resp.StatusCode {
		return true
	}
	return false
}

func NewTransport(skipTlsVerify bool) *http.Transport {
	return &http.Transport{
		Proxy: ProxyFromEnvironment,
		DialContext: defaultTransportDialContext(&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}),
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          10,
		MaxIdleConnsPerHost:   2,
		MaxConnsPerHost:       2,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   7 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: skipTlsVerify}}
}

func defaultTransportDialContext(dialer *net.Dialer) func(context.Context, string, string) (net.Conn, error) {
	return dialer.DialContext
}

func ProxyFromEnvironment(req *http.Request) (*url.URL, error) {
	return httpproxy.FromEnvironment().ProxyFunc()(req.URL)
}
