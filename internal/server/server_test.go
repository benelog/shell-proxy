package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/benelog/shell-proxy/internal/auth"
	"github.com/benelog/shell-proxy/internal/executor"
)

const (
	testUser     = "tester"
	testPassword = "s3cr3t"
)

// newTestServer builds a server with known credentials. Tests that call
// handlers directly bypass the auth wrapper; tests that go through
// s.http.Handler must send these credentials.
func newTestServer() *ShellProxyServer {
	return New(0, auth.Credentials{Username: testUser, Password: testPassword})
}

func TestExecReturnsCommandOutput(t *testing.T) {
	s := newTestServer()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/exec?command=echo+hello", nil)

	s.handleExec(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var got executor.ProcessExecutionResult
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("invalid JSON: %v (%s)", err, rec.Body.String())
	}
	if !strings.Contains(got.StandardOutput, "hello") {
		t.Errorf("stdout = %q, want it to contain %q", got.StandardOutput, "hello")
	}
}

func TestExecMissingCommandIsBadRequest(t *testing.T) {
	s := newTestServer()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/exec", nil)

	s.handleExec(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestConsoleServesUI(t *testing.T) {
	s := newTestServer()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, ConsolePath, nil)

	s.handleConsole(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Errorf("content-type = %q, want text/html", ct)
	}
	if !strings.Contains(rec.Body.String(), "type a command and press Enter") {
		t.Errorf("body did not contain the console UI markup")
	}
}

// With no interactive mode there is only one UI, so "/" is a detour.
func TestRootRedirectsToConsoleWhenNotInteractive(t *testing.T) {
	s := newTestServer()
	rec := httptest.NewRecorder()

	s.handleRoot(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusFound {
		t.Fatalf("status = %d, want 302", rec.Code)
	}
	if loc := rec.Header().Get("Location"); loc != ConsolePath {
		t.Errorf("Location = %q, want %q", loc, ConsolePath)
	}
}

func TestRootServesModeChooserWhenInteractive(t *testing.T) {
	s := newTestServer()
	s.SetInteractive(true)
	rec := httptest.NewRecorder()

	s.handleRoot(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	for _, want := range []string{`href="` + ConsolePath + `"`, `href="/term"`} {
		if !strings.Contains(rec.Body.String(), want) {
			t.Errorf("chooser is missing %s", want)
		}
	}
}

// "/?command=..." predates both pages and must keep executing in either mode,
// rather than being swallowed by the chooser or the redirect.
func TestRootExecutesWhenCommandGiven(t *testing.T) {
	for _, interactive := range []bool{false, true} {
		s := newTestServer()
		s.SetInteractive(interactive)
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/?command=echo+hi", nil)

		s.handleRoot(rec, req)

		if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
			t.Errorf("interactive=%t: content-type = %q, want application/json", interactive, ct)
		}
		if !strings.Contains(rec.Body.String(), "hi") {
			t.Errorf("interactive=%t: body = %q, want it to contain %q", interactive, rec.Body.String(), "hi")
		}
	}
}

func TestStopInvokesCallback(t *testing.T) {
	s := newTestServer()
	done := make(chan struct{})
	s.OnStop(func() { close(done) })

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/stop", nil)
	s.handleStop(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	<-done // handleStop runs the callback in a goroutine; blocks until called.
}

func TestStartAndStop(t *testing.T) {
	s := newTestServer() // port 0 => OS picks a free port
	errCh := make(chan error, 1)
	go func() { errCh <- s.Start() }()

	if err := s.Stop(); err != nil {
		t.Fatalf("stop failed: %v", err)
	}
	if err := <-errCh; err != nil {
		t.Fatalf("start returned error: %v", err)
	}
}
