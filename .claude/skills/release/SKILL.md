---
name: release
description: Release a new version of shell-proxy to GitHub (cross-compiled binaries, tag, release notes). Use this whenever the user asks to release, publish, tag, or ship a version (e.g. "release v2.2.0", "cut a release", "publish the current state"), even if they don't say "release" explicitly but clearly want a new tagged version on GitHub.
---

# Releasing shell-proxy

A release is: cross-compiled static binaries in `dist/`, an annotated git tag
`vX.Y.Z`, and a GitHub release carrying the binaries and hand-written notes.
The mechanical part is fully scripted in `scripts/release.sh`; your job is the
three judgment calls it cannot make, in this order.

## 1. Choose the version

Look at what changed since the last tag:

    git describe --tags --abbrev=0        # last version
    git log $(git describe --tags --abbrev=0)..HEAD --oneline

Semver: user-visible behavior added -> minor bump; docs/text only -> patch;
breaking change -> major. If the user already named a version, use it.

## 2. Update the pinned-version example in README.md

The download instructions point at `releases/latest` and need no change, but
the one line "To pin a version, replace `latest/download` with
`download/vX.Y.Z`" must be bumped to the new version. The script refuses to
run until this is done. Commit this change (together with anything else that
belongs in the release) so the working tree is clean.

## 3. Write the release notes

Write the notes to a file **outside the repo** (scratchpad or /tmp) so they
do not dirty the working tree. Match the established format of previous
releases (`gh release view <last-tag>` shows one):

    <one-paragraph prose summary: what kind of release, whether behavior changed>

    ## Changes

    - <user-facing change, one bullet per theme, not per commit>

    ## Download

    Pick the binary for your platform below, or:

        curl -LO https://github.com/benelog/shell-proxy/releases/download/vX.Y.Z/shell-proxy-linux-amd64
        chmod +x shell-proxy-linux-amd64
        ./shell-proxy-linux-amd64

While summarizing the commits, actively look for changes that alter behavior
for existing users (new required auth, changed defaults, removed endpoints,
renamed flags). Those get an explicit callout in the prose summary; the
commit log alone won't flag them for you.

## 4. Run the script

    .claude/skills/release/scripts/release.sh vX.Y.Z /path/to/notes.md

It preflights (clean tree, tag free locally and on origin, gh authenticated,
README pinned version bumped, notes mention the version), runs `make ci` and
`make dist`, tags, pushes the current branch and the tag, publishes the
release with `gh release create`, and prints the result. It is fail-fast:
if a preflight check trips, fix the cause and rerun; nothing irreversible
happens until the tag push.

Note: this repo's default branch is **master** and the script pushes
`origin HEAD` deliberately; do not "correct" it to main.
