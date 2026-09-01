#!/usr/bin/env bash
# scripts/test-examples.sh
# Auto-discovers and executes all language implementations under examples/
# Contract: Every example must exit 0 on success and non-zero on failure.
set -eo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
EXAMPLES_DIR="$REPO_ROOT/examples"

# Load .env if present
if [ -f "$REPO_ROOT/.env" ]; then
    set -a
    # shellcheck disable=SC1091
    source "$REPO_ROOT/.env"
    set +a
elif [ -f "$REPO_ROOT/.env.example" ]; then
    set -a
    # shellcheck disable=SC1091
    source "$REPO_ROOT/.env.example"
    set +a
fi

PASSED=0
FAILED=0
SKIPPED=0
FAILURES=()

# Colors for terminal output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}======================================================${NC}"
echo -e "${BLUE}   Bitcoin Developer Sandbox — Example Test Runner   ${NC}"
echo -e "${BLUE}======================================================${NC}"
echo ""

# Filter options
FILTER_LANG="${1:-}"
FILTER_EX="${2:-}"

# Find all example directories matching [0-9][0-9]-* except 00-cli-workshop
for ex_dir in "$EXAMPLES_DIR"/[0-9][0-9]-*; do
    [ -d "$ex_dir" ] || continue
    ex_name="$(basename "$ex_dir")"
    
    # 00-cli-workshop is prose-only
    if [ "$ex_name" = "00-cli-workshop" ]; then
        continue
    fi

    if [ -n "$FILTER_EX" ] && [[ "$ex_name" != *"$FILTER_EX"* ]]; then
        continue
    fi

    echo -e "${BLUE}▶ Testing Lab: ${ex_name}${NC}"

    for lang in python typescript rust go; do
        if [ -n "$FILTER_LANG" ] && [ "$FILTER_LANG" != "$lang" ]; then
            continue
        fi

        lang_dir="$ex_dir/$lang"
        if [ ! -d "$lang_dir" ]; then
            continue
        fi

        test_target="${ex_name}/${lang}"
        echo -n "  • [${lang}] Running ${test_target} ... "

        run_status=0
        case "$lang" in
            python)
                if [ -d "$REPO_ROOT/.venv" ] && [ -f "$REPO_ROOT/.venv/bin/python" ]; then
                    PY_BIN="$REPO_ROOT/.venv/bin/python"
                else
                    PY_BIN="$(command -v python3 || command -v python || true)"
                fi

                if [ -z "$PY_BIN" ]; then
                    echo -e "${YELLOW}SKIP (python not found)${NC}"
                    SKIPPED=$((SKIPPED + 1))
                    continue
                fi

                PYTHONPATH="$REPO_ROOT/examples/python" "$PY_BIN" "$lang_dir/main.py" > /tmp/test_output.log 2>&1 || run_status=$?
                ;;

            typescript)
                if command -v npx >/dev/null 2>&1 && [ -f "$REPO_ROOT/examples/typescript/node_modules/.bin/tsx" ]; then
                    (cd "$REPO_ROOT/examples/typescript" && npx tsx "$lang_dir/main.ts") > /tmp/test_output.log 2>&1 || run_status=$?
                elif command -v tsx >/dev/null 2>&1; then
                    (cd "$lang_dir" && tsx main.ts) > /tmp/test_output.log 2>&1 || run_status=$?
                elif command -v npx >/dev/null 2>&1; then
                    (cd "$lang_dir" && npx --yes tsx main.ts) > /tmp/test_output.log 2>&1 || run_status=$?
                else
                    echo -e "${YELLOW}SKIP (node/npx not found)${NC}"
                    SKIPPED=$((SKIPPED + 1))
                    continue
                fi
                ;;

            rust)
                if ! command -v cargo >/dev/null 2>&1; then
                    echo -e "${YELLOW}SKIP (cargo not found)${NC}"
                    SKIPPED=$((SKIPPED + 1))
                    continue
                fi

                (cd "$lang_dir" && cargo run --quiet) > /tmp/test_output.log 2>&1 || run_status=$?
                ;;

            go)
                if ! command -v go >/dev/null 2>&1; then
                    echo -e "${YELLOW}SKIP (go not found)${NC}"
                    SKIPPED=$((SKIPPED + 1))
                    continue
                fi

                (cd "$lang_dir" && go run .) > /tmp/test_output.log 2>&1 || run_status=$?
                ;;
        esac

        if [ "$run_status" -eq 0 ]; then
            echo -e "${GREEN}PASS ✓${NC}"
            PASSED=$((PASSED + 1))
        else
            echo -e "${RED}FAIL ✗ (exit code ${run_status})${NC}"
            FAILED=$((FAILED + 1))
            FAILURES+=("$test_target")
            echo -e "${RED}--- Output Log for ${test_target} ---${NC}"
            cat /tmp/test_output.log | sed 's/^/    /' || true
            echo -e "${RED}-----------------------------------${NC}"
        fi
    done
    echo ""
done

echo "======================================================"
echo -e "Summary: ${GREEN}${PASSED} passed${NC}, ${RED}${FAILED} failed${NC}, ${YELLOW}${SKIPPED} skipped${NC}"
echo "======================================================"

if [ "$FAILED" -gt 0 ]; then
    echo -e "${RED}Failed tests:${NC}"
    for f in "${FAILURES[@]}"; do
        echo -e "  ${RED}✗ $f${NC}"
    done
    exit 1
fi

exit 0
