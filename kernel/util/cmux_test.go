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
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"io"
	"math/big"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/soheilhy/cmux"
)

func writeSelfSignedCert(t *testing.T) (certPath, keyPath string) {
	t.Helper()

	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key failed: %s", err)
	}

	template := x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "scribli-test"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IsCA:                  true,
		BasicConstraintsValid: true,
		IPAddresses:           []net.IP{net.IPv4(127, 0, 0, 1)},
	}

	der, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		t.Fatalf("create certificate failed: %s", err)
	}

	keyDER, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		t.Fatalf("marshal key failed: %s", err)
	}

	dir := t.TempDir()
	certPath = filepath.Join(dir, "cert.pem")
	keyPath = filepath.Join(dir, "key.pem")
	if err = os.WriteFile(certPath, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), 0o600); err != nil {
		t.Fatalf("write cert failed: %s", err)
	}
	if err = os.WriteFile(keyPath, pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER}), 0o600); err != nil {
		t.Fatalf("write key failed: %s", err)
	}
	return certPath, keyPath
}

func newTestHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "ok")
	})
	return mux
}

func awaitReady(t *testing.T, addr string) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if c, err := net.Dial("tcp", addr); err == nil {
			c.Close()
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("server at %s did not become ready", addr)
}

func TestCmuxDerivedListenerCloseClosesRoot(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := ln.Addr().String()

	m := cmux.New(ln)
	derived := m.Match(cmux.Any())

	serveErrCh := make(chan error, 1)
	go func() { serveErrCh <- m.Serve() }()

	if err := derived.Close(); err != nil {
		t.Logf("derived listener close: %s", err)
	}

	select {
	case err := <-serveErrCh:
		if errors.Is(err, cmux.ErrListenerClosed) {
			t.Fatalf("m.Serve should NOT return cmux.ErrListenerClosed, got %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("m.Serve did not return after closing derived listener")
	}

	if c, err := net.Dial("tcp", addr); err == nil {
		c.Close()
		t.Fatal("expected root listener to be closed, but connection succeeded")
	}
}

func TestServeMultiplexed_HTTPAndHTTPSMustUseSeparateServers(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := ln.Addr().String()
	handler := newTestHandler()
	certPath, keyPath := writeSelfSignedCert(t)

	pubHTTP := &http.Server{Handler: handler}
	pubHTTPS := &http.Server{Handler: handler}
	if pubHTTP == pubHTTPS {
		t.Fatal("HTTP and HTTPS servers must be independent instances")
	}

	serveErrCh := make(chan error, 1)
	go func() {
		_, _, e := ServeMultiplexed(ln, handler, certPath, keyPath, pubHTTP, pubHTTPS)
		serveErrCh <- e
	}()

	awaitReady(t, addr)

	pubHTTP.Close()

	select {
	case <-serveErrCh:

	case <-time.After(3 * time.Second):
		t.Fatal("ServeMultiplexed did not return after closing HTTP server (timeout)")
	}

	pubHTTPS.Close()
}

//

//

func TestServeMultiplexed_HTTPSMustNotReuseExternalServer(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := ln.Addr().String()
	handler := newTestHandler()
	certPath, keyPath := writeSelfSignedCert(t)

	externalServer := &http.Server{Handler: handler}

	type result struct {
		httpSrv  *http.Server
		httpsSrv *http.Server
		err      error
	}
	resultCh := make(chan result, 1)
	go func() {
		h, hs, e := ServeMultiplexed(ln, handler, certPath, keyPath, externalServer, nil)
		resultCh <- result{h, hs, e}
	}()

	awaitReady(t, addr)

	externalServer.Close()

	var res result
	select {
	case res = <-resultCh:
	case <-time.After(5 * time.Second):
		t.Fatal("ServeMultiplexed did not return after externalServer.Close() (timeout)")
	}

	if res.httpSrv != externalServer {
		t.Fatal("returned http server should be the external one")
	}

	if res.httpsSrv == nil {
		t.Fatal("returned https server should be non-nil")
	}
	if res.httpsSrv == externalServer {
		t.Fatal("returned https server must NOT reuse the external httpServer (would close cmux root on Close)")
	}

	if res.err == nil {
		t.Fatal("ServeMultiplexed should return non-nil error after external server close")
	}
	if !errors.Is(res.err, net.ErrClosed) {
		t.Fatalf("returned error should match net.ErrClosed (got %v), otherwise serve.go would os.Exit(21)", res.err)
	}
}

func TestServeMultiplexed_CloseDropsActiveConnections(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := ln.Addr().String()
	handler := newTestHandler()
	certPath, keyPath := writeSelfSignedCert(t)

	pubHTTP := &http.Server{Handler: handler}
	pubHTTPS := &http.Server{Handler: handler}

	serveErrCh := make(chan error, 1)
	go func() {
		_, _, e := ServeMultiplexed(ln, handler, certPath, keyPath, pubHTTP, pubHTTPS)
		serveErrCh <- e
	}()

	awaitReady(t, addr)

	resp, err := http.Get("http://" + addr + "/")
	if err != nil {
		t.Fatalf("request before shutdown failed: %s", err)
	}
	resp.Body.Close()

	if err := ln.Close(); err != nil {
		t.Logf("listener close: %s", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	_ = pubHTTP.Shutdown(ctx)
	_ = pubHTTPS.Shutdown(ctx)
	cancel()
	pubHTTP.Close()
	pubHTTPS.Close()

	select {
	case <-serveErrCh:

	case <-time.After(5 * time.Second):
		t.Fatal("ServeMultiplexed did not return after shutdown (timeout)")
	}

	if c, err := net.DialTimeout("tcp", addr, 2*time.Second); err == nil {
		c.Close()
		t.Fatal("expected connection to be refused after publish service shutdown, but dial succeeded")
	}
}
