#!/usr/bin/env bash
# Release shell-proxy: preflight checks, build, tag, push, publish.
#
# Usage: release.sh vX.Y.Z /path/to/notes.md
#
# Everything mechanical lives here. Judgment calls (choosing the version,
# writing the notes, bumping the pinned version in README.md) happen before
# this script runs; the preflight below verifies they actually happened.

set -euo pipefail

die() { echo "release.sh: $*" >&2; exit 1; }

VERSION="${1:-}"
NOTES="${2:-}"

[[ -n "$VERSION" && -n "$NOTES" ]] || die "usage: release.sh vX.Y.Z /path/to/notes.md"
[[ "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]] || die "version must look like vX.Y.Z, got: $VERSION"
[[ -s "$NOTES" ]] || die "notes file missing or empty: $NOTES"

cd "$(git rev-parse --show-toplevel)"

echo "==> Preflight"
[[ -z "$(git status --porcelain)" ]] || die "working tree is not clean; commit or stash first"
! git rev-parse -q --verify "refs/tags/$VERSION" >/dev/null || die "tag $VERSION already exists locally"
[[ -z "$(git ls-remote --tags origin "refs/tags/$VERSION")" ]] || die "tag $VERSION already exists on origin"
gh auth status >/dev/null || die "gh is not authenticated"
grep -q "download/$VERSION" README.md \
    || die "README.md pinned-version example does not mention download/$VERSION; update it before releasing"
grep -q "$VERSION" "$NOTES" || die "notes file never mentions $VERSION; wrong file?"

echo "==> Lint and test (make ci)"
make ci

echo "==> Cross-compile release binaries (make dist)"
make dist

echo "==> Tag and push"
git tag -a "$VERSION" -m "$VERSION"
# Push the current branch (this repo uses master), then the tag.
git push origin HEAD
git push origin "$VERSION"

echo "==> Publish GitHub release"
gh release create "$VERSION" dist/* --title "$VERSION" --notes-file "$NOTES"

echo "==> Verify"
gh release view "$VERSION"

echo "==> Done: $VERSION released"
