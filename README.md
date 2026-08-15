shell-proxy
===========

A small HTTP server that executes shell commands and returns the result as JSON,
plus a terminal-style web UI to type commands from the browser. It has two modes:

- **Stateless mode** (default) — dependency-free, single static binary. One
  command in, one JSON result out. Great for scripting and automation.
- **Interactive mode** (opt-in via `--interactive`) — a real PTY streamed to an
  xterm.js browser terminal, so you can run full-screen programs like `vi`,
  `top`, or `htop`.

> ⚠️ **Security warning**: this server runs arbitrary shell commands with the
> privileges of the process that started it, and has **no authentication**.
> Interactive mode gives a full login shell. Run it only on trusted, isolated
> networks (localhost, a private lab, a container you control). Never expose it
> to the public internet.


## How to run

### Download a prebuilt binary

Download the binary for your platform from the
[latest release](https://github.com/benelog/shell-proxy/releases/latest):

	curl -LO https://github.com/benelog/shell-proxy/releases/latest/download/shell-proxy-linux-amd64
	chmod +x shell-proxy-linux-amd64
	./shell-proxy-linux-amd64

Available assets: `shell-proxy-linux-amd64`, `shell-proxy-linux-arm64`,
`shell-proxy-darwin-amd64`, `shell-proxy-darwin-arm64`,
`shell-proxy-windows-amd64.exe`. To pin a version, replace `latest/download`
with `download/v2.0.0`.

Prefer to build it yourself? See [DEVELOPMENT.md](DEVELOPMENT.md).

### Start the server

	./shell-proxy [port]                # stateless only, default port 18080
	./shell-proxy --interactive [port]  # also enable the /term PTY terminal

On start it prints the reachable web address, e.g. `http://192.168.0.10:18080`.


## Stateless mode (default)

Open the server address in a browser (`http://localhost:18080/`). You get a
CLI-style console: type a command, press Enter, and see stdout (white), stderr
(red), and the exit code inline. `↑`/`↓` walk through history; `clear` wipes the
screen.

This UI and its API have no external dependencies, ship one embedded HTML page,
and work offline.

### HTTP API

| Method   | Path                    | Description                                     |
| -------- | ----------------------- | ----------------------------------------------- |
| GET/POST | `/exec?command=<cmd>`   | Run `<cmd>`, return the result as JSON          |
| GET      | `/?command=<cmd>`       | Same as above                                   |
| GET      | `/`                     | Serve the web UI (when no `command` is present) |
| GET      | `/stop`                 | Shut the server down                            |

Example:

	curl "http://localhost:18080/exec?command=echo%20hello%20%7C%20wc%20-w"

Response:

```json
{
  "exitCode": 0,
  "standardOutput": "2\n",
  "standardError": "",
  "timedOut": false
}
```

Commands run through the system shell (`sh -c` on Unix, `cmd /C` on Windows),
so pipes, redirection and `&&` chaining work. Each command has a 61-second
timeout; a command killed by the timeout reports `"timedOut": true`.


## Interactive mode (`--interactive`)

Start with `--interactive` and open `http://localhost:18080/term`. This runs a
real login shell in a PTY and streams it to an xterm.js terminal over WebSocket,
so **interactive full-screen programs work**: `vi`, `top`, tab-completion,
colors, `Ctrl+C`, resizing, and so on.

| Path                    | Description                                                  |
| ----------------------- | ------------------------------------------------------------ |
| GET `/term`             | The xterm.js terminal UI                                     |
| WS  `/pty`              | WebSocket ↔ PTY bridge (binary = keystrokes, text = resize)  |
| GET `/pty?command=<c>`  | Run a specific program in the PTY instead of a login shell   |

When interactive mode is **not** enabled, `/term` and `/pty` return `404`, so the
feature is invisible unless you explicitly turn it on.

Notes:

- **Off by default** for safety — a PTY session is a full shell.
- **Unix only.** PTY allocation is not supported on Windows; `/pty` returns
  `501` there.
- Works offline: the xterm.js front end is bundled into the binary, no CDN.


## How this differs from similar open-source tools

There is a crowded field of "shell over the web" projects. They fall into two
camps; `shell-proxy` deliberately offers **both** in one binary, with the split
kept explicit.

**1. Interactive PTY streamers** — [gotty](https://github.com/yudai/gotty),
[ttyd](https://github.com/tsl0922/ttyd), [wetty](https://github.com/butlerx/wetty),
[webssh](https://github.com/huashengdun/webssh).
These allocate a real pseudo-terminal and stream it over WebSocket with xterm.js,
giving a fully interactive session. That is exactly what `shell-proxy`'s
`--interactive` mode does — but it is *one opt-in half* of the tool, off by
default, rather than the whole product.

**2. Command-to-HTTP mappers** —
[shell2http](https://github.com/msoap/shell2http),
[shellst](https://github.com/fdefelici/shellst),
[go-shell-run](https://github.com/harrisoncramer/go-shell-run).
These expose shell commands as HTTP endpoints, usually mapping *fixed* commands
to *fixed* routes (e.g. `/date` → `date`), aimed at webhooks and automation.

**Where `shell-proxy` differs:**

- **Two modes, one binary, clean separation.** The same tool gives you a
  scriptable stateless JSON API *and* an interactive PTY terminal — most projects
  pick one lane. The stateless path stays dependency-free; the PTY terminal loads
  only when you pass `--interactive`, and is 404 until then.
- **Stateless mode returns structured data, not a stream.** One command in, one
  JSON result out (`exitCode` + `stdout` + `stderr` + `timedOut`) — trivial to
  script against, and not limited to a pre-wired route table like the
  command-mappers.
- **Minimal by default.** With interactive mode off it is a small static binary
  with an embedded HTML console — no xterm.js bundle, no PTY layer in the request
  path.

Trade-off to be aware of: there is still **no authentication or TLS** in either
mode, so this is a trusted-network tool, not an internet-facing one — the
interactive streamers above generally ship basic-auth/TLS options that this
project intentionally leaves out.


## Development

Building from source, quality tooling, project layout, and the release process
live in [DEVELOPMENT.md](DEVELOPMENT.md).


## License

[MIT](LICENSE) © Sanghyuk Jung
