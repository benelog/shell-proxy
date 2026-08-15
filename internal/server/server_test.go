package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/benelog/shell-proxy/internal/executor"
)

func TestExecReturnsCommandOutput(t *testing.T) {
	s := New(0)
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
	s := New(0)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/exec", nil)

	s.handleExec(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestRootServesUIWhenNoCommand(t *testing.T) {
	s := New(0)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	s.handleRoot(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Errorf("content-type = %q, want text/html", ct)
	}
	if !strings.Contains(rec.Body.String(), "shell-proxy") {
		t.Errorf("body did not contain the UI markup")
	}
}

func TestRootExecutesWhenCommandGiven(t *testing.T) {
	s := New(0)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/?command=echo+hi", nil)

	s.handleRoot(rec, req)

	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Errorf("content-type = %q, want application/json", ct)
	}
	if !strings.Contains(rec.Body.String(), "hi") {
		t.Errorf("body = %q, want it to contain %q", rec.Body.String(), "hi")
	}
}

func TestStopInvokesCallback(t *testing.T) {
	s := New(0)
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
	s := New(0) // port 0 => OS picks a free port
	errCh := make(chan error, 1)
	go func() { errCh <- s.Start() }()

	if err := s.Stop(); err != nil {
		t.Fatalf("stop failed: %v", err)
	}
	if err := <-errCh; err != nil {
		t.Fatalf("start returned error: %v", err)
	}
}
