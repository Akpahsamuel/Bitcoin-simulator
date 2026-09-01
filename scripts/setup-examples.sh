#!/usr/bin/env bash
# scripts/setup-examples.sh
# Installs the per-language dependencies the coded examples need. Idempotent —
# safe to re-run. Each language is independent: a missing toolchain is skipped,
# not an error, so you can set up only the language you care about.
#
# Run automatically as the devcontainer `postCreateCommand`, by CI, and usable
# by hand:  bash scripts/setup-examples.sh
set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$REPO_ROOT"

rc=0

# --- Python: virtualenv + pinned deps ---------------------------------------
if command -v python3 >/dev/null 2>&1; then
    # --copies (not symlinks) so the venv also works on bind-mounted workspaces
    # (OrbStack / Docker Desktop virtiofs).
    [ -x .venv/bin/python ] || python3 -m venv --copies .venv
    if ./.venv/bin/pip install --quiet --upgrade pip \
        && ./.venv/bin/pip install --quiet -r examples/python/requirements.txt; then
        echo "✓ python  — .venv ready"
    else
        echo "✗ python  — pip install failed"; rc=1
    fi
else
    echo "· python  — skipped (python3 not found)"
fi

# --- TypeScript: npm workspace rooted at examples/ -------------------------—-
if command -v npm >/dev/null 2>&1; then
    if ( cd examples && npm install --no-audit --no-fund --silent ); then
        echo "✓ typescript — examples/node_modules ready"
    else
        echo "✗ typescript — npm install failed"; rc=1
    fi
else
    echo "· typescript — skipped (npm not found)"
fi

# --- Rust: pre-fetch each lab crate so the first run needs no network -------—
if command -v cargo >/dev/null 2>&1; then
    ok=1
    for d in examples/[0-9][0-9]-*/rust; do
        [ -d "$d" ] || continue
        ( cd "$d" && cargo fetch --quiet ) || ok=0
    done
    [ "$ok" = 1 ] && echo "✓ rust    — cargo deps fetched" || { echo "✗ rust    — cargo fetch failed"; rc=1; }
else
    echo "· rust    — skipped (cargo not found)"
fi

# --- Go: download modules for each lab -------------------------------------—-
if command -v go >/dev/null 2>&1; then
    ok=1
    for d in examples/[0-9][0-9]-*/go; do
        [ -d "$d" ] || continue
        ( cd "$d" && go mod download ) || ok=0
    done
    [ "$ok" = 1 ] && echo "✓ go      — modules downloaded" || { echo "✗ go      — go mod download failed"; rc=1; }
else
    echo "· go      — skipped (go not found)"
fi

echo
if [ "$rc" -eq 0 ]; then
    echo "Setup complete. Run the suite:  bash scripts/test-examples.sh"
else
    echo "Setup finished with errors (exit $rc)." >&2
fi
exit "$rc"
