package server

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

func TestTermIs404WhenInteractiveDisabled(t *testing.T) {
	s := newTestServer() // interactive defaults to off
	rec := httptest.NewRecorder()
	s.handleTerm(rec, httptest.NewRequest(http.MethodGet, "/term", nil))

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404 when interactive disabled", rec.Code)
	}
}

func TestPTYIs404WhenInteractiveDisabled(t *testing.T) {
	s := newTestServer()
	rec := httptest.NewRecorder()
	s.handlePTY(rec, httptest.NewRequest(http.MethodGet, "/pty", nil))

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404 when interactive disabled", rec.Code)
	}
}

func TestTermServesUIWhenInteractiveEnabled(t *testing.T) {
	s := newTestServer()
	s.SetInteractive(true)
	rec := httptest.NewRecorder()
	s.handleTerm(rec, httptest.NewRequest(http.MethodGet, "/term", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "xterm.js") {
		t.Errorf("term UI did not reference xterm.js")
	}
}

// The stateless console only advertises /term when that endpoint would
// actually answer, mirroring the 404 gate above.
func TestStatelessConsoleLinksToTermOnlyWhenInteractive(t *testing.T) {
	for _, tc := range []struct {
		interactive bool
		want        string
	}{
		// The <body> tag is what the CSS keys off; matching the bare
		// attribute would also hit the stylesheet's own selector.
		{false, `<body data-interactive="false">`},
		{true, `<body data-interactive="true">`},
	} {
		s := newTestServer()
		s.SetInteractive(tc.interactive)
		rec := httptest.NewRecorder()
		s.handleConsole(rec, httptest.NewRequest(http.MethodGet, ConsolePath, nil))

		if !strings.Contains(rec.Body.String(), tc.want) {
			t.Errorf("interactive=%t: page did not contain %s", tc.interactive, tc.want)
		}
		if !strings.Contains(rec.Body.String(), `href="/term"`) {
			t.Errorf("interactive=%t: page is missing the /term link markup", tc.interactive)
		}
	}
}

func TestSameOrigin(t *testing.T) {
	for _, tc := range []struct {
		name   string
		origin string
		host   string
		want   bool
	}{
		{"no origin means a non-browser client", "", "localhost:18080", true},
		{"page served by this server", "http://localhost:18080", "localhost:18080", true},
		{"https page on the same host", "https://box.local:8443", "box.local:8443", true},
		{"host comparison ignores case", "http://Box.Local:8443", "box.local:8443", true},
		{"another site entirely", "http://evil.example", "localhost:18080", false},
		{"same host, different port", "http://localhost:9999", "localhost:18080", false},
		{"unparseable origin", "://nonsense", "localhost:18080", false},
		{"null origin from a sandboxed frame", "null", "localhost:18080", false},
	} {
		r := httptest.NewRequest(http.MethodGet, "/pty", nil)
		r.Host = tc.host
		if tc.origin != "" {
			r.Header.Set("Origin", tc.origin)
		}
		if got := sameOrigin(r); got != tc.want {
			t.Errorf("%s: sameOrigin(origin=%q, host=%q) = %t, want %t",
				tc.name, tc.origin, tc.host, got, tc.want)
		}
	}
}

// A cross-origin handshake must be refused before the PTY is allocated, even
// when it carries valid credentials: browsers may attach cached credentials
// to a WebSocket handshake started by any page.
func TestPTYRejectsCrossOriginHandshake(t *testing.T) {
	s := newTestServer()
	s.SetInteractive(true)
	ts := httptest.NewServer(s.http.Handler)
	defer ts.Close()

	h := http.Header{}
	h.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(testUser+":"+testPassword)))
	h.Set("Origin", "http://evil.example")

	conn, resp, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(ts.URL, "http")+"/pty", h)
	if err == nil {
		_ = conn.Close()
		t.Fatal("cross-origin handshake succeeded, want it refused")
	}
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("status = %d, want 403", resp.StatusCode)
	}
}

func TestAssetsAreServed(t *testing.T) {
	s := newTestServer()
	ts := httptest.NewServer(s.http.Handler)
	defer ts.Close()

	req, err := http.NewRequest(http.MethodGet, ts.URL+"/assets/xterm.js", nil)
	if err != nil {
		t.Fatalf("bad request: %v", err)
	}
	req.SetBasicAuth(testUser, testPassword)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200 for /assets/xterm.js", resp.StatusCode)
	}
}
