package household

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"io"
	"log"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

// @test CON-009-AC19: direct construction cannot bypass startup validation.
func TestS3RejectsHTTPBeforeNetwork(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { requests.Add(1) }))
	defer server.Close()
	_, err := NewS3Backend(t.TempDir(), StorageConfig{Endpoint: server.URL, Bucket: "bucket", AccessKey: "fixture", SecretKey: "fixture-secret"})
	if err == nil {
		t.Fatal("HTTP backend accepted")
	}
	if requests.Load() != 0 {
		t.Fatal("HTTP request sent before rejecting configuration")
	}
}

// @test CON-009-AC19: an HTTPS response cannot downgrade subsequent requests.
func TestS3RejectsHTTPRedirect(t *testing.T) {
	var requests atomic.Int32
	plaintext := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { requests.Add(1) }))
	defer plaintext.Close()
	secure := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, plaintext.URL, http.StatusFound) }))
	defer secure.Close()
	b, err := NewS3Backend(t.TempDir(), StorageConfig{Endpoint: secure.URL, Bucket: "bucket", AccessKey: "fixture", SecretKey: "fixture-secret"})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, _, err := b.Load(); err == nil {
		t.Fatal("HTTP redirect accepted")
	}
	if requests.Load() != 0 {
		t.Fatal("redirect sent a plaintext request")
	}
	request, err := http.NewRequest(http.MethodGet, secure.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := s3Redirect(request, make([]*http.Request, 10)); err == nil {
		t.Fatal("redirect limit lost")
	}
	if err := s3Redirect(request, nil); err != nil {
		t.Fatal("HTTPS redirect rejected", err)
	}
}

// @test CON-009-AC19: trusting fixtures must not disable certificate validation.
func TestS3RejectsUntrustedTLS(t *testing.T) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	certificate := &x509.Certificate{SerialNumber: big.NewInt(1), NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour), IPAddresses: []net.IP{net.ParseIP("127.0.0.1")}, KeyUsage: x509.KeyUsageDigitalSignature, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}}
	der, err := x509.CreateCertificate(rand.Reader, certificate, certificate, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	var requests atomic.Int32
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { requests.Add(1) }))
	server.TLS = &tls.Config{MinVersion: tls.VersionTLS12, Certificates: []tls.Certificate{{Certificate: [][]byte{der}, PrivateKey: key}}}
	server.Config.ErrorLog = log.New(io.Discard, "", 0)
	server.StartTLS()
	defer server.Close()
	b, err := NewS3Backend(t.TempDir(), StorageConfig{Endpoint: server.URL, Bucket: "bucket", AccessKey: "fixture", SecretKey: "fixture-secret"})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, _, err := b.Load(); err == nil {
		t.Fatal("untrusted certificate accepted")
	}
	if requests.Load() != 0 {
		t.Fatal("request reached server with untrusted certificate")
	}
}
