// Command shell-proxy starts an HTTP server that executes shell commands,
// with a terminal-style web UI.
//
// It is a Go port of the original Java net.benelog.shellproxy project.
package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/benelog/shell-proxy/internal/auth"
	"github.com/benelog/shell-proxy/internal/server"
)

// DefaultPort matches the original Java Starter.DEFAULT_PORT.
const DefaultPort = 18080

func main() {
	os.Exit(run(os.Args[1:]))
}

// run wires up the server and blocks until it stops. It returns a process
// exit code so it stays testable.
func run(args []string) int {
	fs := flag.NewFlagSet("shell-proxy", flag.ContinueOnError)
	interactive := fs.Bool("interactive", false, "enable the interactive PTY terminal at /term (allows vi, top, etc.)")
	username := fs.String("user", "", "HTTP Basic auth username (default: the OS login account)")
	password := fs.String("password", "", "HTTP Basic auth password (default: randomly generated at startup)")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	creds, err := auth.Resolve(*username, *password)
	if err != nil {
		fmt.Fprintln(os.Stderr, "cannot generate a password:", err)
		return 1
	}

	printUsage()

	port := parsePort(fs.Args())
	ip := localIP()
	fmt.Printf("Web address: http://%s:%d\n", ip, port)

	srv := server.New(port, creds)
	srv.SetInteractive(*interactive)
	printAuthPolicy(creds, port)
	printModePolicy(*interactive, ip, port)
	srv.OnStop(func() {
		fmt.Println("Server stop")
		_ = srv.Stop()
	})

	// Stop cleanly on Ctrl+C / SIGTERM, mirroring the Java shutdown hook.
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		fmt.Println("Server stop")
		_ = srv.Stop()
	}()

	if err := srv.Start(); err != nil {
		fmt.Fprintln(os.Stderr, "server error:", err)
		return 1
	}
	return 0
}

// parsePort returns the port from args[0], falling back to DefaultPort when the
// argument is missing or not a number.
func parsePort(args []string) int {
	if len(args) == 0 {
		return DefaultPort
	}
	port, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Println("Cannot parse port : " + args[0])
		return DefaultPort
	}
	return port
}

// printModePolicy spells out which of the two UIs is reachable. The web
// address above points at the stateless console, where full-screen programs
// such as vi or top cannot work: they need the PTY terminal, so its URL is
// printed here rather than left for the reader to guess.
func printModePolicy(interactive bool, ip string, port int) {
	fmt.Println("-----------------------------")
	if interactive {
		fmt.Println("Interactive PTY mode ENABLED (a real login shell; trusted networks only).")
		fmt.Printf("   Pick a mode   : http://%s:%d/          the two modes, side by side\n", ip, port)
		fmt.Printf("   Terminal      : http://%s:%d/term      vi, top, and other full-screen programs\n", ip, port)
		fmt.Printf("   JSON console  : http://%s:%d/console   one command in, one JSON result out\n", ip, port)
	} else {
		fmt.Println("Interactive PTY mode disabled: /term and /pty return 404.")
		fmt.Printf("   JSON console  : http://%s:%d/console   one command in, one JSON result out\n", ip, port)
		fmt.Println("   \"/\" redirects there, since it is the only mode available.")
		fmt.Println("   Full-screen programs (vi, top, ...) cannot run here; they need a real")
		fmt.Println("   terminal, so restart with --interactive to get one at /term.")
	}
	fmt.Println("-----------------------------")
	fmt.Println()
}

// localIP returns the machine's primary outbound IP, or "localhost" if it
// cannot be determined.
func localIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "localhost"
	}
	defer func() { _ = conn.Close() }()
	return conn.LocalAddr().(*net.UDPAddr).IP.String()
}

// printAuthPolicy explains the credentials policy at boot. A generated
// password is printed because this is the only chance to see it; an
// operator-supplied one is never echoed.
func printAuthPolicy(creds auth.Credentials, port int) {
	fmt.Println("-----------------------------")
	fmt.Println("Authentication: HTTP Basic auth is required on every endpoint")
	fmt.Println("   (/, /exec, /stop, /term, /pty, /assets/).")
	fmt.Println("   Username : " + creds.Username + usernameOrigin(creds))
	if creds.GeneratedPassword {
		fmt.Println("   Password : " + creds.Password + "  (randomly generated for this run)")
		fmt.Println("   The password changes on every restart. Use --user / --password to fix it.")
	} else {
		fmt.Println("   Password : (hidden; set with --password)")
	}
	fmt.Println()
	fmt.Println("   Browsers prompt for these credentials on the first request.")
	fmt.Printf("   From the command line:\n")
	fmt.Printf("      curl -u '%s:%s' 'http://localhost:%d/exec?command=whoami'\n",
		creds.Username, passwordForExample(creds), port)
	fmt.Println("-----------------------------")
	fmt.Println()
}

func usernameOrigin(creds auth.Credentials) string {
	if creds.DefaultedUsername {
		return "  (OS login account)"
	}
	return "  (from --user)"
}

func passwordForExample(creds auth.Credentials) string {
	if creds.GeneratedPassword {
		return creds.Password
	}
	return "<password>"
}

func printUsage() {
	fmt.Println("-----------------------------")
	fmt.Println("Usage:")
	fmt.Println("   Prompt> shell-proxy [options] [port]")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("   --interactive        enable the PTY terminal at /term")
	fmt.Println("   --user <name>        Basic auth username (default: OS login account)")
	fmt.Println("   --password <secret>  Basic auth password (default: randomly generated)")
	fmt.Println("-----------------------------")
	fmt.Println()
}
