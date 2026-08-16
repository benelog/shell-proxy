Developing shell-proxy
=========

Go 1.22 or later is required to build and test.

Dependencies differ by mode:

- **Stateless mode** uses only the Go standard library.
- **Interactive mode** pulls in [`github.com/creack/pty`](https://github.com/creack/pty) (PTY allocation) and [`github.com/gorilla/websocket`](https://github.com/gorilla/websocket) (WebSocket), plus the vendored [xterm.js](https://xtermjs.org/) front end under `internal/web/assets`.
  These are only reachable from the interactive endpoints; the stateless request path never touches them.

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

The enabled linters are declared in `.golangci.yml` (the `standard` set plus `misspell`, with `gofmt`/`goimports` as formatters).
The same checks run on every push and pull request through `.github/workflows/ci.yml`.

Layout
---------

    main.go                            entry point, flags, port parsing, startup banner
    internal/auth/auth.go              resolves the Basic auth credentials (OS user + random password)
    internal/executor/executor.go      runs a command via the system shell, captures output
    internal/server/server.go          HTTP server, Basic auth gate, /, /console and /exec handlers
    internal/server/interactive.go     PTY + WebSocket bridge for /term and /pty
    internal/web/home.html             mode chooser served at /
    internal/web/console.html          stateless CLI-style console served at /console
    internal/web/term.html             xterm.js interactive terminal UI
    internal/web/assets/               vendored xterm.js / CSS / fit-addon (go:embed)
    internal/web/web.go                embeds the HTML and assets into the binary

The interactive endpoints are always registered but gated at request time on the server's `interactive` flag, so `/term` and `/pty` return `404` until the server is started with `--interactive`.
Keep that gate in place when changing the PTY path, since it is what keeps the shell-spawning feature off by default.

`/` is a chooser, not a UI: it serves `home.html` when both modes are available and redirects to `ConsolePath` when only the stateless one is.
`/?command=...` is checked before either branch, because that contract predates the pages and scripts still rely on it.

Authentication wraps the whole mux in `server.New`, so a new route is protected the moment it is registered: there is no per-handler auth code to forget.
`New` takes the credentials as an argument rather than a setter, which makes an unauthenticated server impossible to construct.
The 401 response must keep its `WWW-Authenticate` header, otherwise browsers never show the login prompt and the web UI becomes unreachable.
Credential comparison goes through `crypto/subtle` on SHA-256 digests, so neither the password nor its length leaks through timing.

Basic auth is not enough on its own for `/pty`: WebSocket handshakes are exempt from CORS, and a browser may attach the credentials it cached for this origin to a handshake started by any page.
`sameOrigin` therefore requires the `Origin` header, when present, to match the request's `Host`, and a mismatch is logged before the 403 because the usual cause is a reverse proxy rewriting `Host` rather than an attack.
A missing `Origin` means a non-browser client and is allowed, since such a client can forge any value anyway and still has to authenticate.

Releasing
---------

The whole mechanical sequence lives in `.claude/skills/release/scripts/release.sh`:

    .claude/skills/release/scripts/release.sh vX.Y.Z /path/to/notes.md

It checks the preconditions (clean tree, tag not taken, `gh` authenticated, `README.md` pinned version bumped), then runs `make ci`, `make dist`, tags, pushes, and publishes the release with the notes file.
The equivalent manual steps, should the script be unavailable:

    make dist                                     # builds dist/shell-proxy-<os>-<arch>
    git tag -a vX.Y.Z -m "vX.Y.Z" && git push origin vX.Y.Z
    gh release create vX.Y.Z dist/* --title "vX.Y.Z" --notes-file notes.md

The download instructions in `README.md` point at `releases/latest`, so they need no update per release; only the pinned-version example does.
