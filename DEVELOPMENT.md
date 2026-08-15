Developing shell-proxy
=========

Go 1.22 or later is required to build and test.

Dependencies differ by mode:

- **Stateless mode** uses only the Go standard library.
- **Interactive mode** pulls in [`github.com/creack/pty`](https://github.com/creack/pty)
  (PTY allocation) and [`github.com/gorilla/websocket`](https://github.com/gorilla/websocket)
  (WebSocket), plus the vendored [xterm.js](https://xtermjs.org/) front end under
  `internal/web/assets`. These are only reachable from the interactive endpoints;
  the stateless request path never touches them.

Build

    make build      # go build -o shell-proxy .
    make run        # build and start on the default port

Quality tools
---------

    make fmt        # goimports -w .
    make lint       # golangci-lint run ./...
    make test       # go test ./...
    make check      # fmt + lint + test, before committing
    make ci         # lint + test, what CI runs

`goimports` and `golangci-lint` (v2) are needed for `fmt` and `lint`:

    go install golang.org/x/tools/cmd/goimports@latest
    go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest

The enabled linters are declared in `.golangci.yml` (the `standard` set plus
`misspell`, with `gofmt`/`goimports` as formatters). The same checks run on every
push and pull request through `.github/workflows/ci.yml`.

Layout
---------

    main.go                            entry point, --interactive flag and port parsing
    internal/executor/executor.go      runs a command via the system shell, captures output
    internal/server/server.go          HTTP server, stateless /exec and / handlers
    internal/server/interactive.go     PTY + WebSocket bridge for /term and /pty
    internal/web/index.html            stateless CLI-style console
    internal/web/term.html             xterm.js interactive terminal UI
    internal/web/assets/               vendored xterm.js / CSS / fit-addon (go:embed)
    internal/web/web.go                embeds the HTML and assets into the binary

The interactive endpoints are always registered but gated at request time on the
server's `interactive` flag, so `/term` and `/pty` return `404` until the server
is started with `--interactive`. Keep that gate in place when changing the PTY
path — it is what keeps the shell-spawning feature off by default.

Releasing
---------

Cross-compile one static binary per platform, tag, then publish the assets:

    make dist                                     # builds dist/shell-proxy-<os>-<arch>
    git tag -a vX.Y.Z -m "vX.Y.Z" && git push origin vX.Y.Z
    gh release create vX.Y.Z dist/* --title "vX.Y.Z" --notes-file notes.md

The download instructions in `README.md` point at `releases/latest`, so they need
no update per release; only the pinned-version example does.
