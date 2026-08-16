shell-proxy
===========

A small HTTP server that executes shell commands and returns the result as JSON, plus a terminal-style web UI to type commands from the browser.
It has two modes:

- **Stateless mode** (default): dependency-free, single static binary.
  One command in, one JSON result out.
  Great for scripting and automation.
- **Interactive mode** (opt-in via `--interactive`): a real PTY streamed to an xterm.js browser terminal, so you can run full-screen programs like `vi`, `top`, or `htop`.

> ⚠️ **Security warning**: this server runs arbitrary shell commands with the privileges of the process that started it.
> Every endpoint is behind HTTP Basic auth (see [Authentication](#authentication)), but the traffic is plain **HTTP with no TLS**, so credentials travel unencrypted.
> Interactive mode gives a full login shell.
> Run it only on trusted, isolated networks (localhost, a private lab, a container you control).
> Never expose it to the public internet.


## How to run

### Download a prebuilt binary

Download the binary for your platform from the [latest release](https://github.com/benelog/shell-proxy/releases/latest):

	curl -LO https://github.com/benelog/shell-proxy/releases/latest/download/shell-proxy-linux-amd64
	chmod +x shell-proxy-linux-amd64
	./shell-proxy-linux-amd64

Available assets: `shell-proxy-linux-amd64`, `shell-proxy-linux-arm64`, `shell-proxy-darwin-amd64`, `shell-proxy-darwin-arm64`, `shell-proxy-windows-amd64.exe`.
To pin a version, replace `latest/download` with `download/v2.0.0`.

Prefer to build it yourself?
See [DEVELOPMENT.md](DEVELOPMENT.md).

### Start the server

	./shell-proxy [options] [port]      # stateless only, default port 18080
	./shell-proxy --interactive [port]  # also enable the /term PTY terminal

On start it prints the reachable web address, e.g. `http://192.168.0.10:18080`, followed by the credentials needed to reach it.

## Authentication

Every endpoint (`/`, `/exec`, `/stop`, `/term`, `/pty`, `/assets/`) requires **HTTP Basic auth**.
There is no way to turn it off.

By default:

- the **username** is the OS account the server runs as (`whoami`);
- the **password** is randomly generated (128 bits) each time the server starts, and printed to the console once, at startup.

	-----------------------------
	Authentication: HTTP Basic auth is required on every endpoint
	   (/, /exec, /stop, /term, /pty, /assets/).
	   Username : alice  (OS login account)
	   Password : 4jddSK1CdMHklMzAifdkiQ  (randomly generated for this run)
	   The password changes on every restart. Use --user / --password to fix it.

	   Browsers prompt for these credentials on the first request.
	   From the command line:
	      curl -u 'alice:4jddSK1CdMHklMzAifdkiQ' 'http://localhost:18080/exec?command=whoami'
	-----------------------------

Because a generated password changes on every restart, set both explicitly when you want stable credentials (for a script, a service unit, or a container):

	./shell-proxy --user ops --password 's3cr3t' 18080

A password given with `--password` is never echoed to the console.

| Option              | Default                       |
| ------------------- | ----------------------------- |
| `--user <name>`     | the OS login account          |
| `--password <pass>` | randomly generated at startup |

In a browser, the first request pops up the standard login prompt; the same credentials are then reused for `/term` and its `/pty` WebSocket connection.


## Choosing a mode

`/` explains the two modes side by side and links to each, because they are not interchangeable: one returns a JSON result a script can read, the other gives you a terminal.
With `--interactive` off there is nothing to choose, so `/` redirects to `/console`.


## Stateless mode (default)

Open `http://localhost:18080/console`.
You get a CLI-style console: type a command, press Enter, and see stdout (white), stderr (red), and the exit code inline.
`↑`/`↓` walk through history; `clear` wipes the screen.

This UI and its API have no external dependencies, ship one embedded HTML page, and work offline.

Each command runs without a terminal attached, so full-screen programs cannot work here: `vi` produces no screen and sits there until the 61-second timeout, and `top` exits with `failed tty get`.
Those need [interactive mode](#interactive-mode---interactive); when the server is started with `--interactive`, this console shows a link to `/term` in its header.

### HTTP API

| Method   | Path                    | Description                                                          |
| -------- | ----------------------- | -------------------------------------------------------------------- |
| GET/POST | `/exec?command=<cmd>`   | Run `<cmd>`, return the result as JSON                               |
| GET      | `/?command=<cmd>`       | Same as above                                                        |
| GET      | `/console`              | Serve the stateless web console                                      |
| GET      | `/`                     | Mode chooser, or a redirect to `/console` without `--interactive`    |
| GET      | `/stop`                 | Shut the server down                                                 |

Example (`-u` carries the Basic credentials printed at startup):

	curl -u 'alice:4jddSK1CdMHklMzAifdkiQ' \
	  "http://localhost:18080/exec?command=echo%20hello%20%7C%20wc%20-w"

Response:

```json
{
  "exitCode": 0,
  "standardOutput": "2\n",
  "standardError": "",
  "timedOut": false
}
```

Commands run through the system shell (`sh -c` on Unix, `cmd /C` on Windows), so pipes, redirection and `&&` chaining work.
Each command has a 61-second timeout; a command killed by the timeout reports `"timedOut": true`.


## Interactive mode (`--interactive`)

Start with `--interactive` and open `http://localhost:18080/term`.
This runs a real login shell in a PTY and streams it to an xterm.js terminal over WebSocket, so **interactive full-screen programs work**: `vi`, `top`, `claude`, tab-completion, colors, `Ctrl+C`, resizing, and so on.
The startup banner prints the `/term` address, and both `/` and the console header link to it, so you do not have to remember the path.

| Path                    | Description                                                  |
| ----------------------- | ------------------------------------------------------------ |
| GET `/term`             | The xterm.js terminal UI                                     |
| WS  `/pty`              | WebSocket ↔ PTY bridge (binary = keystrokes, text = resize)  |
| GET `/pty?command=<c>`  | Run a specific program in the PTY instead of a login shell   |

When interactive mode is **not** enabled, `/term` and `/pty` return `404`, so the feature is invisible unless you explicitly turn it on.

Notes:

- **Off by default** for safety, because a PTY session is a full shell.
- **Same-origin only.**
  The `/pty` handshake is refused with `403` unless its `Origin` matches the host you addressed, so a page you happen to visit cannot open a shell here with credentials your browser cached.
  Requests without an `Origin` header (curl, wscat, any non-browser client) are still accepted; they always need the credentials.
- Each browser tab is its own session: a separate login shell with its own working directory and history.
  Closing or reloading the tab ends that shell; there is no reattach.
- **Unix only.**
  PTY allocation is not supported on Windows; `/pty` returns `501` there.
- Works offline: the xterm.js front end is bundled into the binary, no CDN.


## How this differs from similar open-source tools

There is a crowded field of "shell over the web" projects.
They fall into two camps; `shell-proxy` deliberately offers **both** in one binary, with the split kept explicit.

**1. Interactive PTY streamers**: [gotty](https://github.com/yudai/gotty), [ttyd](https://github.com/tsl0922/ttyd), [wetty](https://github.com/butlerx/wetty), [webssh](https://github.com/huashengdun/webssh).
These allocate a real pseudo-terminal and stream it over WebSocket with xterm.js, giving a fully interactive session.
That is exactly what `shell-proxy`'s `--interactive` mode does, but it is *one opt-in half* of the tool, off by default, rather than the whole product.

**2. Command-to-HTTP mappers**: [shell2http](https://github.com/msoap/shell2http), [shellst](https://github.com/fdefelici/shellst), [go-shell-run](https://github.com/harrisoncramer/go-shell-run).
These expose shell commands as HTTP endpoints, usually mapping *fixed* commands to *fixed* routes (e.g. `/date` → `date`), aimed at webhooks and automation.

**Where `shell-proxy` differs:**

- **Two modes, one binary, clean separation.**
  The same tool gives you a scriptable stateless JSON API *and* an interactive PTY terminal, while most projects pick one lane.
  The stateless path stays dependency-free; the PTY terminal loads only when you pass `--interactive`, and is 404 until then.
- **Stateless mode returns structured data, not a stream.**
  One command in, one JSON result out (`exitCode` + `stdout` + `stderr` + `timedOut`), so it is trivial to script against and not limited to a pre-wired route table like the command-mappers.
- **Minimal by default.**
  With interactive mode off it is a small static binary with an embedded HTML console: no xterm.js bundle, no PTY layer in the request path.
- **Authenticated by default, with zero setup.**
  Both modes are behind HTTP Basic auth from the first run: no config file, no flag to remember, because the server generates a password at startup and prints it.
  The streamers above ship basic-auth options too, but they are opt-in flags.

Trade-off to be aware of: there is still **no TLS** in either mode, so credentials and command output travel in the clear.
This remains a trusted-network tool, not an internet-facing one; put it behind a reverse proxy if you need transport encryption.


## Development

Building from source, quality tooling, project layout, and the release process live in [DEVELOPMENT.md](DEVELOPMENT.md).


## License

[MIT](LICENSE) © Sanghyuk Jung
