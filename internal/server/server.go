// Package server exposes the shell executor over HTTP.
package server

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/benelog/shell-proxy/internal/executor"
	"github.com/benelog/shell-proxy/internal/web"
)

// defaultTimeout matches the original Java servlet's 61s command timeout.
const defaultTimeout = 61 * time.Second

// ShellProxyServer wraps an http.Server that executes shell commands.
type ShellProxyServer struct {
	http *http.Server
	exec *executor.ShellExecutor
	// interactive enables the PTY/WebSocket terminal endpoints. Off by default.
	interactive bool
	// stopFn is invoked by the /stop endpoint. Defaults to os.Exit(0)-like
	// behavior via the Starter; tests can override it.
	stopFn func()
}

// New builds a ShellProxyServer listening on the given port.
func New(port int) *ShellProxyServer {
	s := &ShellProxyServer{
		exec: executor.New(defaultTimeout),
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleRoot)
	mux.HandleFunc("/exec", s.handleExec)
	mux.HandleFunc("/stop", s.handleStop)

	// Interactive endpoints are always registered but gated at request time on
	// s.interactive, so they stay invisible (404) until explicitly enabled.
	mux.HandleFunc("/term", s.handleTerm)
	mux.HandleFunc("/pty", s.handlePTY)
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.FS(web.Assets()))))

	s.http = &http.Server{
		Addr:    ":" + strconv.Itoa(port),
		Handler: mux,
	}
	return s
}

// SetInteractive enables or disables the interactive PTY terminal mode.
func (s *ShellProxyServer) SetInteractive(enabled bool) {
	s.interactive = enabled
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

// handleRoot serves the UI when no command is given, and — for backward
// compatibility with the original "/?command=..." contract — executes the
// command when one is present.
func (s *ShellProxyServer) handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if r.URL.Query().Get("command") == "" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(web.IndexHTML)
		return
	}
	s.handleExec(w, r)
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
