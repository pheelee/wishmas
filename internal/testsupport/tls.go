// Package testsupport provides helpers for isolated test processes only.
package testsupport

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"net/http"
	"net/http/httptest"
)

// TrustLocalTLSServers adds httptest's public fixture certificate to the process
// trust pool. Call only from TestMain before tests/goroutines start. Production
// code never calls this helper; TLS verification remains enabled in tests.
func TrustLocalTLSServers() error {
	fixture := httptest.NewTLSServer(http.NotFoundHandler())
	certificate := fixture.Certificate()
	fixture.Close()
	roots, err := x509.SystemCertPool()
	if err != nil {
		return err
	}
	base, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		return errors.New("test TLS trust requires the standard HTTP transport")
	}
	transport := base.Clone()
	roots.AddCert(certificate)
	transport.TLSClientConfig = &tls.Config{RootCAs: roots, MinVersion: tls.VersionTLS12}
	http.DefaultTransport = transport
	return nil
}
