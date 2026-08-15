package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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
