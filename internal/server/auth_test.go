package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// doRequest sends a request through the full handler chain, including the
// Basic auth wrapper. Empty user means "send no credentials at all".
func doRequest(t *testing.T, s *ShellProxyServer, path, user, pass string) *http.Response {
	t.Helper()
	ts := httptest.NewServer(s.http.Handler)
	t.Cleanup(ts.Close)

	req, err := http.NewRequest(http.MethodGet, ts.URL+path, nil)
	if err != nil {
		t.Fatalf("bad request: %v", err)
	}
	if user != "" {
		req.SetBasicAuth(user, pass)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	t.Cleanup(func() { _ = resp.Body.Close() })
	return resp
}

func TestRequestWithoutCredentialsIsUnauthorized(t *testing.T) {
	resp := doRequest(t, newTestServer(), "/exec?command=echo+hi", "", "")

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", resp.StatusCode)
	}
	// Without this header the browser never shows a login prompt.
	challenge := resp.Header.Get("WWW-Authenticate")
	if !strings.HasPrefix(challenge, "Basic ") || !strings.Contains(challenge, `realm="shell-proxy"`) {
		t.Errorf("WWW-Authenticate = %q, want a Basic challenge with the shell-proxy realm", challenge)
	}
}

func TestRequestWithWrongPasswordIsUnauthorized(t *testing.T) {
	resp := doRequest(t, newTestServer(), "/exec?command=echo+hi", testUser, "wrong")

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", resp.StatusCode)
	}
}

func TestRequestWithWrongUsernameIsUnauthorized(t *testing.T) {
	resp := doRequest(t, newTestServer(), "/exec?command=echo+hi", "someone-else", testPassword)

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", resp.StatusCode)
	}
}

func TestRequestWithCorrectCredentialsSucceeds(t *testing.T) {
	resp := doRequest(t, newTestServer(), "/exec?command=echo+hi", testUser, testPassword)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
}

// Every route sits behind the same gate, not just /exec.
func TestAllRoutesRequireCredentials(t *testing.T) {
	for _, path := range []string{"/", ConsolePath, "/exec", "/stop", "/term", "/pty", "/assets/xterm.js"} {
		s := newTestServer()
		s.SetInteractive(true)
		resp := doRequest(t, s, path, "", "")
		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("GET %s without credentials = %d, want 401", path, resp.StatusCode)
		}
	}
}
