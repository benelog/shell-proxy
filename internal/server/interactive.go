package server

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/benelog/shell-proxy/internal/web"
	"github.com/creack/pty"
	"github.com/gorilla/websocket"
)

// resizeMessage is the JSON control frame the browser sends when the terminal
// is resized. Key input is sent as binary frames instead.
type resizeMessage struct {
	Cols uint16 `json:"cols"`
	Rows uint16 `json:"rows"`
}

// upgrader turns the /pty request into a WebSocket. WebSocket handshakes are
// not subject to CORS, so the Origin check below is the only thing stopping a
// page the user happens to visit from opening a shell here: browsers may
// attach the credentials they cached for this origin to such a handshake, and
// Basic auth alone would then wave it through.
var upgrader = websocket.Upgrader{
	CheckOrigin: sameOrigin,
}

// sameOrigin reports whether the handshake came from a page this server
// served. A request with no Origin header is not a browser (curl, wscat, a Go
// dialer), and those are allowed: Origin only means something when a browser
// is the one setting it, and a non-browser client can forge any value anyway.
// Credentials remain required either way. This is the same check gorilla
// applies when CheckOrigin is nil; it exists as our own function only so a
// rejection can be logged.
func sameOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}
	// r.Host carries the host:port the client addressed, which is what the
	// browser also puts in Origin, so this keeps working behind a proxy that
	// forwards the original Host.
	u, err := url.Parse(origin)
	if err == nil && strings.EqualFold(u.Host, r.Host) {
		return true
	}
	// Logged because a rejected handshake is otherwise a bare 403, and the
	// usual cause is a proxy rewriting Host rather than an actual attack.
	log.Printf("pty: rejected handshake with origin %q for host %q", origin, r.Host)
	return false
}

// handleTerm serves the interactive terminal UI. It 404s when interactive mode
// is disabled so the endpoint is invisible unless explicitly enabled.
func (s *ShellProxyServer) handleTerm(w http.ResponseWriter, r *http.Request) {
	if !s.interactive {
		interactiveDisabled(w)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(web.TermHTML)
}

// handlePTY bridges a WebSocket to a PTY running a login shell (or the command
// given via ?command=). Binary frames are keystrokes; text frames are resize
// control messages.
func (s *ShellProxyServer) handlePTY(w http.ResponseWriter, r *http.Request) {
	if !s.interactive {
		interactiveDisabled(w)
		return
	}
	if runtime.GOOS == "windows" {
		http.Error(w, "interactive PTY mode is not supported on Windows", http.StatusNotImplemented)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("pty: websocket upgrade failed: %v", err)
		return
	}
	defer func() { _ = conn.Close() }()

	cmd := ptyCommand(r.URL.Query().Get("command"))
	tty, err := pty.Start(cmd)
	if err != nil {
		log.Printf("pty: failed to start %v: %v", cmd.Args, err)
		_ = conn.WriteMessage(websocket.TextMessage, []byte("failed to start session: "+err.Error()))
		return
	}
	log.Printf("pty: session started: %v", cmd.Args)
	defer func() {
		_ = tty.Close()
		_ = cmd.Process.Kill()
		_, _ = cmd.Process.Wait()
		log.Printf("pty: session ended: %v", cmd.Args)
	}()

	// PTY output -> browser. When the shell exits, the read fails and this
	// goroutine closes the connection, unblocking the input loop below.
	go func() {
		buf := make([]byte, 32*1024)
		for {
			n, err := tty.Read(buf)
			if n > 0 {
				if werr := conn.WriteMessage(websocket.BinaryMessage, buf[:n]); werr != nil {
					return
				}
			}
			if err != nil {
				if err != io.EOF {
					log.Printf("pty: read error: %v", err)
				}
				_ = conn.Close()
				return
			}
		}
	}()

	// Browser -> PTY: binary = keystrokes, text = resize control frame.
	for {
		msgType, data, err := conn.ReadMessage()
		if err != nil {
			return
		}
		switch msgType {
		case websocket.TextMessage:
			var m resizeMessage
			if json.Unmarshal(data, &m) == nil && m.Cols > 0 && m.Rows > 0 {
				_ = pty.Setsize(tty, &pty.Winsize{Cols: m.Cols, Rows: m.Rows})
			}
		case websocket.BinaryMessage:
			if _, err := tty.Write(data); err != nil {
				return
			}
		}
	}
}

// ptyCommand builds the process to run inside the PTY. With no command it
// launches the user's login shell; otherwise it runs the command via that shell.
func ptyCommand(command string) *exec.Cmd {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}
	var cmd *exec.Cmd
	if command == "" {
		cmd = exec.Command(shell)
	} else {
		cmd = exec.Command(shell, "-c", command)
	}
	cmd.Env = append(os.Environ(), "TERM=xterm-256color")
	return cmd
}

func interactiveDisabled(w http.ResponseWriter) {
	http.Error(w, "interactive mode is disabled; start the server with --interactive to enable it", http.StatusNotFound)
}
