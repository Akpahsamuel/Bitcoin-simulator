#!/usr/bin/env bash
# scripts/test-examples.sh
# Auto-discovers and executes language implementations under examples/.
# Contract: every example exits 0 on success, non-zero on failure.
#
# Usage:
#   scripts/test-examples.sh [options] [LANG] [EXAMPLE]
#
# Options (all optional; positional LANG/EXAMPLE still work for back-compat):
#   -l, --lang   LANGS     comma-separated: python,typescript,rust,go  (default: all)
#   -e, --example NAME     substring of a lab dir, e.g. 02  or  keys   (default: all)
#   -L, --list             list discovered <lab>/<lang> targets and exit
#   -h, --help             show this help and exit
#
# Examples:
#   scripts/test-examples.sh                 # every language, every lab
#   scripts/test-examples.sh rust            # just Rust, every lab
#   scripts/test-examples.sh --lang python,go --example 03
#   scripts/test-examples.sh -e 01           # every language, lab 01 only
set -eo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
EXAMPLES_DIR="$REPO_ROOT/examples"

ALL_LANGS="python typescript rust go"

# Print the contiguous comment header (everything from line 2 to the first
# non-comment line) as help text.
usage() { awk 'NR>1 && /^#/ {sub(/^# ?/, ""); print; next} NR>1 {exit}' "${BASH_SOURCE[0]}"; }

FILTER_LANGS=""
FILTER_EX=""
LIST_ONLY=0
POSITIONAL=()

while [ $# -gt 0 ]; do
    case "$1" in
        -h|--help)    usage; exit 0 ;;
        -L|--list)    LIST_ONLY=1; shift ;;
        -l|--lang)    FILTER_LANGS="$2"; shift 2 ;;
        --lang=*)     FILTER_LANGS="${1#*=}"; shift ;;
        -e|--example) FILTER_EX="$2"; shift 2 ;;
        --example=*)  FILTER_EX="${1#*=}"; shift ;;
        --)           shift; while [ $# -gt 0 ]; do POSITIONAL+=("$1"); shift; done ;;
        -*)           echo "Unknown option: $1" >&2; usage >&2; exit 2 ;;
        *)            POSITIONAL+=("$1"); shift ;;
    esac
done

# Back-compat: positional  [LANG] [EXAMPLE]  (only if the flags weren't used)
[ -z "$FILTER_LANGS" ] && [ "${#POSITIONAL[@]}" -ge 1 ] && FILTER_LANGS="${POSITIONAL[0]}"
[ -z "$FILTER_EX" ]    && [ "${#POSITIONAL[@]}" -ge 2 ] && FILTER_EX="${POSITIONAL[1]}"

# Normalise + validate the language filter.
FILTER_LANGS="${FILTER_LANGS//,/ }"
if [ -n "$FILTER_LANGS" ]; then
    for l in $FILTER_LANGS; do
        case " $ALL_LANGS " in
            *" $l "*) : ;;
            *) echo "Unknown language: '$l' (valid: ${ALL_LANGS// /, })" >&2; exit 2 ;;
        esac
    done
else
    FILTER_LANGS="$ALL_LANGS"
fi

lang_wanted() { case " $FILTER_LANGS " in *" $1 "*) return 0 ;; *) return 1 ;; esac; }

# Load .env if present
if [ -f "$REPO_ROOT/.env" ]; then
    set -a; # shellcheck disable=SC1091
    source "$REPO_ROOT/.env"; set +a
elif [ -f "$REPO_ROOT/.env.example" ]; then
    set -a; # shellcheck disable=SC1091
    source "$REPO_ROOT/.env.example"; set +a
fi

PASSED=0
FAILED=0
SKIPPED=0
MATCHED=0
FAILURES=()

GREEN='\033[0;32m'; RED='\033[0;31m'; YELLOW='\033[0;33m'; BLUE='\033[0;34m'; NC='\033[0m'

if [ "$LIST_ONLY" -eq 1 ]; then
    for ex_dir in "$EXAMPLES_DIR"/[0-9][0-9]-*; do
        [ -d "$ex_dir" ] || continue
        ex_name="$(basename "$ex_dir")"
        [ "$ex_name" = "00-cli-workshop" ] && continue
        [ -n "$FILTER_EX" ] && [[ "$ex_name" != *"$FILTER_EX"* ]] && continue
        for lang in $ALL_LANGS; do
            lang_wanted "$lang" || continue
            [ -d "$ex_dir/$lang" ] && echo "${ex_name}/${lang}"
        done
    done
    exit 0
fi

echo -e "${BLUE}======================================================${NC}"
echo -e "${BLUE}   Bitcoin Developer Sandbox — Example Test Runner   ${NC}"
echo -e "${BLUE}======================================================${NC}"
echo -e "  languages: ${FILTER_LANGS}"
echo -e "  labs:      ${FILTER_EX:-all}"
echo ""

for ex_dir in "$EXAMPLES_DIR"/[0-9][0-9]-*; do
    [ -d "$ex_dir" ] || continue
    ex_name="$(basename "$ex_dir")"

    # 00-cli-workshop is prose-only
    [ "$ex_name" = "00-cli-workshop" ] && continue

    if [ -n "$FILTER_EX" ] && [[ "$ex_name" != *"$FILTER_EX"* ]]; then
        continue
    fi

    echo -e "${BLUE}▶ Testing Lab: ${ex_name}${NC}"

    for lang in $ALL_LANGS; do
        lang_wanted "$lang" || continue

        lang_dir="$ex_dir/$lang"
        [ -d "$lang_dir" ] || continue

        MATCHED=$((MATCHED + 1))
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
                # Deps are an npm workspace rooted at examples/ — install hoists tsx and
                # the shared libs to examples/node_modules, which is an ancestor of every
                # examples/<NN>/typescript/main.ts so bare imports resolve.
                if [ -f "$EXAMPLES_DIR/node_modules/.bin/tsx" ]; then
                    (cd "$EXAMPLES_DIR" && node_modules/.bin/tsx "$lang_dir/main.ts") > /tmp/test_output.log 2>&1 || run_status=$?
                elif command -v npx >/dev/null 2>&1; then
                    (cd "$EXAMPLES_DIR" && npx --yes tsx "$lang_dir/main.ts") > /tmp/test_output.log 2>&1 || run_status=$?
                else
                    echo -e "${YELLOW}SKIP (node/npx not found — run: cd examples && npm install)${NC}"
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
            sed 's/^/    /' /tmp/test_output.log || true
            echo -e "${RED}-----------------------------------${NC}"
        fi
    done
    echo ""
done

if [ "$MATCHED" -eq 0 ]; then
    echo -e "${RED}No examples matched (languages: ${FILTER_LANGS}; labs: ${FILTER_EX:-all}).${NC}"
    echo "Try: scripts/test-examples.sh --list"
    exit 2
fi

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
