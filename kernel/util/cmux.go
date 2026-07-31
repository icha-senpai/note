package util

import (
	"crypto/tls"
	"errors"
	"net"
	"net/http"

	"github.com/icha-senpai/note/third_party/forks/logging"
	"github.com/soheilhy/cmux"
)

//

//

func ServeMultiplexed(ln net.Listener, handler http.Handler, certPath, keyPath string, httpServer, httpsServer *http.Server) (*http.Server, *http.Server, error) {
	m := cmux.New(ln)

	tlsL := m.Match(cmux.TLS())
	httpL := m.Match(cmux.Any())

	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		logging.LogErrorf("failed to load TLS cert for multiplexing: %s", err)
		return nil, nil, err
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		NextProtos:   []string{"h2", "http/1.1"},
	}

	tlsListener := tls.NewListener(tlsL, tlsConfig)

	if httpServer == nil {
		httpServer = &http.Server{Handler: handler}
	} else {
		httpServer.Handler = handler
	}
	if httpsServer == nil {
		httpsServer = &http.Server{Handler: handler}
	} else {
		httpsServer.Handler = handler
	}

	go func() {
		if serveErr := httpServer.Serve(httpL); serveErr != nil && !errors.Is(serveErr, cmux.ErrListenerClosed) && !errors.Is(serveErr, http.ErrServerClosed) {
			logging.LogErrorf("multiplexed HTTP server error: %s", serveErr)
		}
	}()

	go func() {
		if serveErr := httpsServer.Serve(tlsListener); serveErr != nil && !errors.Is(serveErr, cmux.ErrListenerClosed) && !errors.Is(serveErr, http.ErrServerClosed) {
			logging.LogErrorf("multiplexed HTTPS server error: %s", serveErr)
		}
	}()

	return httpServer, httpsServer, m.Serve()
}
