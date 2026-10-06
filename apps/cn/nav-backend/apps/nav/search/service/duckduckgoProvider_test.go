package service

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gofurry/gofurry-nav-backend/common"
)

// Only the test transport dials a local TLS server; the production URL stays fixed.
func localProvider(t *testing.T, handler http.HandlerFunc) (*duckDuckGoProvider, *httptest.Server) {
	t.Helper()
	server := httptest.NewTLSServer(handler)
	t.Cleanup(server.Close)
	p := newDuckDuckGoProvider("")
	transport := p.client.Transport.(*http.Transport)
	transport.TLSClientConfig = testTLSConfig(server)
	transport.DialContext = func(ctx context.Context, network, _ string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, network, server.Listener.Addr().String())
	}
	t.Cleanup(transport.CloseIdleConnections)
	return p, server
}

func testTLSConfig(server *httptest.Server) *tls.Config {
	roots := x509.NewCertPool()
	roots.AddCert(server.Certificate())
	return &tls.Config{RootCAs: roots, ServerName: server.Certificate().DNSNames[0], MinVersion: tls.VersionTLS12}
}

func TestDuckDuckGoDecodeAndReuse(t *testing.T) {
	peers := map[string]bool{}
	var peersMu sync.Mutex
	p, _ := localProvider(t, func(w http.ResponseWriter, r *http.Request) {
		peersMu.Lock()
		peers[r.RemoteAddr] = true
		peersMu.Unlock()
		if r.Host != "duckduckgo.com" || r.URL.Path != "/ac/" || r.URL.Query().Get("q") != "兽 人&?" || len(r.URL.Query()) != 1 || r.Header.Get("User-Agent") != common.USER_AGENT {
			t.Errorf("unexpected request: %s %s", r.Host, r.URL)
		}
		_, _ = io.WriteString(w, `[{"phrase":" furry "},{"phrase":"furry"},{"phrase":" "},{"phrase":"兽人"}]`)
	})
	client, transport := p.client, p.client.Transport.(*http.Transport)
	if client.Timeout != 3*time.Second || transport.DisableKeepAlives || transport.TLSHandshakeTimeout == 0 {
		t.Fatal("lost standard transport defaults")
	}
	for i := 0; i < 2; i++ {
		items, err := p.Fetch(context.Background(), " 兽 人&? ")
		if err != nil || strings.Join(items, ",") != "furry,兽人" {
			t.Fatalf("items=%v error=%v", items, err)
		}
		if p.client != client || p.client.Transport != transport {
			t.Fatal("client recreated")
		}
	}
	peersMu.Lock()
	defer peersMu.Unlock()
	if len(peers) != 1 {
		t.Fatalf("expected connection reuse, peers=%d", len(peers))
	}
}

func TestDuckDuckGoFailureBoundaries(t *testing.T) {
	for _, tc := range []struct {
		name      string
		status    int
		body      string
		wantError bool
	}{
		{"empty", 200, `[]`, false},
		{"exact body bound", 200, `[]` + strings.Repeat(" ", searchSuggestMaxBodyBytes-2), false},
		{"non-success", 503, `[]`, true},
		{"redirect", 302, `[]`, true},
		{"malformed", 200, `<html>blocked</html>`, true},
		{"null", 200, `null`, true},
		{"shape drift", 200, `[{"unexpected":"value"}]`, true},
		{"trailing data", 200, `[] {}`, true},
		{"oversize even after valid JSON", 200, `[]` + strings.Repeat(" ", searchSuggestMaxBodyBytes), true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p, _ := localProvider(t, func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Location", "https://unrelated.invalid/")
				w.WriteHeader(tc.status)
				_, _ = io.WriteString(w, tc.body)
			})
			items, err := p.Fetch(context.Background(), "furry")
			if (err != nil) != tc.wantError {
				t.Fatalf("items=%v error=%v", items, err)
			}
			if !tc.wantError && (items == nil || len(items) != 0) {
				t.Fatal("empty must be a non-nil slice")
			}
		})
	}
}

func TestDuckDuckGoTimeoutAndCancellation(t *testing.T) {
	p, _ := localProvider(t, func(w http.ResponseWriter, r *http.Request) { <-r.Context().Done() })
	p.client.Timeout = 20 * time.Millisecond
	if _, err := p.Fetch(context.Background(), "furry"); err == nil {
		t.Fatal("timeout accepted")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := p.Fetch(ctx, "furry"); err != context.Canceled {
		t.Fatalf("cancel: %v", err)
	}
}

func TestDuckDuckGoConfiguredProxyAndReuse(t *testing.T) {
	var connects, requests atomic.Int64
	upstream := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		if r.URL.Query().Get("q") != "furry" {
			t.Error("proxy lost query")
		}
		_, _ = io.WriteString(w, `[{"phrase":"furry"}]`)
	}))
	defer upstream.Close()
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		connects.Add(1)
		if r.Method != http.MethodConnect || r.Host != "duckduckgo.com:443" {
			t.Errorf("unexpected CONNECT: %s %s", r.Method, r.Host)
			w.WriteHeader(400)
			return
		}
		remote, err := net.Dial("tcp", upstream.Listener.Addr().String())
		if err != nil {
			t.Error(err)
			w.WriteHeader(502)
			return
		}
		defer remote.Close()
		conn, _, err := w.(http.Hijacker).Hijack()
		if err != nil {
			t.Error(err)
			return
		}
		defer conn.Close()
		_, _ = io.WriteString(conn, "HTTP/1.1 200 Connection Established\r\n\r\n")
		go func() { _, _ = io.Copy(remote, conn); _ = remote.Close() }()
		_, _ = io.Copy(conn, remote)
	}))
	defer proxy.Close()
	p := newDuckDuckGoProvider(proxy.URL)
	transport := p.client.Transport.(*http.Transport)
	transport.TLSClientConfig = testTLSConfig(upstream)
	defer transport.CloseIdleConnections()
	// Any silent direct connection fails locally instead of accessing the internet.
	transport.DialContext = func(ctx context.Context, network, address string) (net.Conn, error) {
		if address != proxy.Listener.Addr().String() {
			t.Errorf("bypassed configured proxy: %s", address)
			return nil, net.ErrClosed
		}
		return (&net.Dialer{}).DialContext(ctx, network, address)
	}
	for i := 0; i < 2; i++ {
		if _, err := p.Fetch(context.Background(), "furry"); err != nil {
			t.Fatal(err)
		}
	}
	if connects.Load() != 1 || requests.Load() != 2 {
		t.Fatalf("CONNECT=%d requests=%d", connects.Load(), requests.Load())
	}
}

func TestDuckDuckGoInvalidProxyNeverCreatesDirectClient(t *testing.T) {
	for _, raw := range []string{"://bad", "proxy.example:1080", "http://", "ftp://proxy.example", "http://proxy.example:invalid", " "} {
		p := newDuckDuckGoProvider(raw)
		if p.client != nil {
			t.Fatalf("invalid proxy created a client: %q", raw)
		}
		if _, err := p.Fetch(context.Background(), "furry"); err == nil {
			t.Fatal("invalid proxy accepted")
		}
	}
}
