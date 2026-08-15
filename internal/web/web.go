// Package web holds the static assets for the UIs: the dependency-free
// stateless console and the xterm.js-based interactive terminal.
package web

import (
	"embed"
	"io/fs"
)

//go:embed index.html term.html assets
var files embed.FS

// mustRead returns the contents of an embedded file, panicking on error.
// The files are compiled into the binary, so a failure is a build-time bug.
func mustRead(name string) []byte {
	b, err := files.ReadFile(name)
	if err != nil {
		panic("web: missing embedded asset " + name + ": " + err.Error())
	}
	return b
}

// IndexHTML is the stateless CLI-style console served at "/".
var IndexHTML = mustRead("index.html")

// TermHTML is the interactive xterm.js terminal served at "/term".
var TermHTML = mustRead("term.html")

// Assets is the embedded /assets tree (xterm.js, xterm.css, addon-fit.js),
// ready to hand to http.FileServer.
func Assets() fs.FS {
	sub, err := fs.Sub(files, "assets")
	if err != nil {
		panic("web: cannot open assets sub-tree: " + err.Error())
	}
	return sub
}
