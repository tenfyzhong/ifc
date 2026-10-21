package main

import (
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestFormatStr(t *testing.T) {
	for _, tc := range []struct {
		input string
		max   int
		want  string
	}{
		{"", 4, ""},
		{"abc", 4, "abc"},
		{"abcd", 4, "abcd"},
		{"abcde", 4, "abcd"},
		{"abc", 0, ""},
	} {
		if got := formatStr(tc.input, tc.max); got != tc.want {
			t.Errorf("formatStr(%q, %d) = %q, want %q", tc.input, tc.max, got, tc.want)
		}
	}
}

// This checks the application's TLS path on the patched toolchain. It is not
// a reproduction of GO-2026-6809's compromised-intermediate denylist case.
func TestNewClientTLS(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/_ping" {
			t.Errorf("unexpected Docker API path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte("OK"))
	}))
	defer server.Close()

	dir, err := ioutil.TempDir("", "ifc-tls-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	caFile := filepath.Join(dir, "ca.pem")
	cert := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: server.Certificate().Raw})
	if err := ioutil.WriteFile(caFile, cert, 0600); err != nil {
		t.Fatal(err)
	}

	oldHost, oldTLS, oldCA, oldCert, oldKey := host, tls, tlscacert, tlscert, tlskey
	defer func() {
		host, tls, tlscacert, tlscert, tlskey = oldHost, oldTLS, oldCA, oldCert, oldKey
	}()
	host, tls, tlscacert, tlscert, tlskey = server.URL, true, caFile, "", ""

	for _, tc := range []struct {
		name       string
		untrusted  bool
		serverName string
	}{
		{name: "trusted certificate"},
		{name: "untrusted certificate", untrusted: true},
		{name: "wrong hostname", serverName: "not-the-docker-host.invalid"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			client, err := newClient()
			if err != nil {
				t.Fatal(err)
			}
			defer client.HTTPClient.CloseIdleConnections()
			if client.TLSConfig.InsecureSkipVerify {
				t.Fatal("certificate verification must be enabled when a CA is configured")
			}
			if tc.untrusted {
				client.TLSConfig.RootCAs = x509.NewCertPool()
			}
			client.TLSConfig.ServerName = tc.serverName
			err = client.Ping()
			wantError := tc.untrusted || tc.serverName != ""
			if (err != nil) != wantError {
				t.Fatalf("Ping() error = %v, want error %v", err, wantError)
			}
		})
	}
}
