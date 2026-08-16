// Package server exposes the shell executor over HTTP.
package server

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/benelog/shell-proxy/internal/auth"
	"github.com/benelog/shell-proxy/internal/executor"
	"github.com/benelog/shell-proxy/internal/web"
)

// defaultTimeout matches the original Java servlet's 61s command timeout.
const defaultTimeout = 61 * time.Second

// authRealm is the HTTP Basic realm shown in the browser's login prompt.
const authRealm = "shell-proxy"

// ShellProxyServer wraps an http.Server that executes shell commands.
type ShellProxyServer struct {
	http *http.Server
	exec *executor.ShellExecutor
	// interactive enables the PTY/WebSocket terminal endpoints. Off by default.
	interactive bool
	// creds are the HTTP Basic credentials every request must present.
	creds auth.Credentials
	// stopFn is invoked by the /stop endpoint. Defaults to os.Exit(0)-like
	// behavior via the Starter; tests can override it.
	stopFn func()
}

// ConsolePath is where the stateless console lives. "/" is the mode chooser
// instead, so the two UIs are presented side by side rather than one of them
// being the thing you land on by accident.
const ConsolePath = "/console"

// New builds a ShellProxyServer listening on the given port. The credentials
// are required: every endpoint is behind HTTP Basic auth, and passing them at
// construction time makes it impossible to start an unprotected server.
func New(port int, creds auth.Credentials) *ShellProxyServer {
	s := &ShellProxyServer{
		exec:  executor.New(defaultTimeout),
		creds: creds,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleRoot)
	mux.HandleFunc(ConsolePath, s.handleConsole)
	mux.HandleFunc("/exec", s.handleExec)
	mux.HandleFunc("/stop", s.handleStop)

	// Interactive endpoints are always registered but gated at request time on
	// s.interactive, so they stay invisible (404) until explicitly enabled.
	mux.HandleFunc("/term", s.handleTerm)
	mux.HandleFunc("/pty", s.handlePTY)
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.FS(web.Assets()))))

	s.http = &http.Server{
		Addr: ":" + strconv.Itoa(port),
		// Every route, including /assets/ and the /pty WebSocket handshake,
		// goes through the Basic auth gate.
		Handler: s.withAuth(mux),
	}
	return s
}

// SetInteractive enables or disables the interactive PTY terminal mode.
func (s *ShellProxyServer) SetInteractive(enabled bool) {
	s.interactive = enabled
}

// withAuth rejects requests that do not carry the expected Basic credentials.
func (s *ShellProxyServer) withAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || !s.credentialsMatch(user, pass) {
			log.Printf("unauthorized request from %s: %s %s", r.RemoteAddr, r.Method, r.URL.Path)
			// The realm header is what makes the browser show a login prompt.
			w.Header().Set("WWW-Authenticate", `Basic realm="`+authRealm+`", charset="UTF-8"`)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// credentialsMatch compares in constant time. Hashing first keeps the
// comparison length-independent, so neither value's length leaks.
func (s *ShellProxyServer) credentialsMatch(user, pass string) bool {
	gotUser, wantUser := sha256.Sum256([]byte(user)), sha256.Sum256([]byte(s.creds.Username))
	gotPass, wantPass := sha256.Sum256([]byte(pass)), sha256.Sum256([]byte(s.creds.Password))
	userOK := subtle.ConstantTimeCompare(gotUser[:], wantUser[:]) == 1
	passOK := subtle.ConstantTimeCompare(gotPass[:], wantPass[:]) == 1
	return userOK && passOK
}

// OnStop registers the callback invoked when /stop is requested.
func (s *ShellProxyServer) OnStop(fn func()) {
	s.stopFn = fn
}

// Start begins serving and blocks until the server is closed.
func (s *ShellProxyServer) Start() error {
	err := s.http.ListenAndServe()
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}

// Stop gracefully shuts the server down.
func (s *ShellProxyServer) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.http.Shutdown(ctx)
}

// handleRoot picks between the two UIs. With only one mode available there is
// nothing to choose, so it redirects to the console. The "/?command=..."
// contract predates both pages and still executes, whatever the mode.
func (s *ShellProxyServer) handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if r.URL.Query().Get("command") != "" {
		s.handleExec(w, r)
		return
	}
	if !s.interactive {
		http.Redirect(w, r, ConsolePath, http.StatusFound)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(web.HomeHTML)
}

// handleConsole serves the stateless CLI-style console.
func (s *ShellProxyServer) handleConsole(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(s.consoleHTML())
}

// interactiveOff / interactiveOn are the <body> attribute the stateless page
// ships with and the value it gets when the PTY terminal is available. The
// page hides its /term link unless the attribute is flipped, so the stateless
// console never advertises an endpoint that would answer 404.
var (
	interactiveOff = []byte(`data-interactive="false"`)
	interactiveOn  = []byte(`data-interactive="true"`)
)

// consoleInteractiveHTML is the console page with that attribute flipped,
// built once since both variants are static.
var consoleInteractiveHTML = bytes.Replace(web.ConsoleHTML, interactiveOff, interactiveOn, 1)

// consoleHTML returns the stateless console, telling it whether the
// interactive terminal is reachable. Full-screen programs (vi, top, and the
// like) cannot work in stateless mode, so this link is how users get out.
func (s *ShellProxyServer) consoleHTML() []byte {
	if !s.interactive {
		return web.ConsoleHTML
	}
	return consoleInteractiveHTML
}

// handleExec runs the "command" parameter and returns the result as JSON.
func (s *ShellProxyServer) handleExec(w http.ResponseWriter, r *http.Request) {
	command := r.FormValue("command")
	if command == "" {
		http.Error(w, `{"error":"missing 'command' parameter"}`, http.StatusBadRequest)
		return
	}

	result, err := s.exec.Execute(command)
	if err != nil {
		log.Printf("failed to execute %q: %v", command, err)
		http.Error(w, `{"error":"execution failed"}`, http.StatusInternalServerError)
		return
	}
	log.Printf("result: exitCode=%d timedOut=%t", result.ExitCode, result.TimedOut)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		log.Printf("failed to encode response: %v", err)
	}
}

// handleStop terminates the server. Mirrors the original StopServlet.
func (s *ShellProxyServer) handleStop(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("Server stop"))
	if s.stopFn != nil {
		// Run after the response is flushed so the client sees the reply.
		go s.stopFn()
	}
}
