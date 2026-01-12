#!/bin/bash
#
# Integration test for ajira issue clone command
#
# Tests cloning issues with various options:
# - Basic clone (copy all fields)
# - Clone with field overrides
# - Clone with link to original
#
# Usage: ./testdata/clone-test.sh [-y|--yes]
#
# Options:
#   -y, --yes    Auto-delete test issues without prompting
#

set -euo pipefail

# Parse arguments
AUTO_DELETE=false
while [[ $# -gt 0 ]]; do
    case $1 in
        -y|--yes)
            AUTO_DELETE=true
            shift
            ;;
        -h|--help)
            echo "Usage: $0 [-y|--yes]"
            echo ""
            echo "Options:"
            echo "  -y, --yes    Auto-delete test issues without prompting"
            echo "  -h, --help   Show this help message"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            echo "Usage: $0 [-y|--yes]"
            exit 1
            ;;
    esac
done

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[OK]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_step() {
    echo -e "\n${GREEN}==>${NC} $1"
}

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
AJIRA_BIN="$PROJECT_DIR/ajira"
TIMESTAMP=$(date +%Y%m%d-%H%M%S)

# Track created issues for cleanup
CREATED_ISSUES=()

cleanup() {
    if [[ ${#CREATED_ISSUES[@]} -gt 0 ]]; then
        log_step "Cleanup"
        if [[ "$AUTO_DELETE" == "true" ]]; then
            for issue in "${CREATED_ISSUES[@]}"; do
                log_info "Deleting $issue..."
                "$AJIRA_BIN" issue delete "$issue" 2>/dev/null || log_warn "Failed to delete $issue"
            done
            log_success "Cleanup complete"
        else
            echo ""
            log_warn "Created issues: ${CREATED_ISSUES[*]}"
            read -p "Delete test issues? (y/N) " -n 1 -r
            echo
            if [[ $REPLY =~ ^[Yy]$ ]]; then
                for issue in "${CREATED_ISSUES[@]}"; do
                    log_info "Deleting $issue..."
                    "$AJIRA_BIN" issue delete "$issue" 2>/dev/null || log_warn "Failed to delete $issue"
                done
                log_success "Cleanup complete"
            else
                log_info "Issues kept for inspection"
            fi
        fi
    fi
}

trap cleanup EXIT

# Check binary exists
if [[ ! -x "$AJIRA_BIN" ]]; then
    log_error "ajira binary not found at $AJIRA_BIN"
    log_info "Run: go build -o ajira ./cmd/ajira"
    exit 1
fi

# Check JIRA_PROJECT is set
if [[ -z "${JIRA_PROJECT:-}" ]]; then
    log_error "JIRA_PROJECT environment variable not set"
    exit 1
fi

log_step "Starting clone integration test"
log_info "Project: $JIRA_PROJECT"
log_info "Timestamp: $TIMESTAMP"

# Test 1: Create source issue
log_step "Test 1: Create source issue"
SOURCE_SUMMARY="Clone Test Source $TIMESTAMP"
SOURCE_DESC="This is the source issue for clone testing.

It has a description with **markdown** formatting."

SOURCE_KEY=$("$AJIRA_BIN" issue create \
    -s "$SOURCE_SUMMARY" \
    -d "$SOURCE_DESC" \
    -t Task \
    --priority Major \
    --labels "test,clone-source" \
    --json | jq -r '.key')

if [[ -z "$SOURCE_KEY" || "$SOURCE_KEY" == "null" ]]; then
    log_error "Failed to create source issue"
    exit 1
fi
CREATED_ISSUES+=("$SOURCE_KEY")
log_success "Created source issue: $SOURCE_KEY"

# Test 2: Basic clone
log_step "Test 2: Basic clone (copy all fields)"
CLONE1_KEY=$("$AJIRA_BIN" issue clone "$SOURCE_KEY" --json | jq -r '.clonedKey')

if [[ -z "$CLONE1_KEY" || "$CLONE1_KEY" == "null" ]]; then
    log_error "Failed to clone issue"
    exit 1
fi
CREATED_ISSUES+=("$CLONE1_KEY")
log_success "Cloned to: $CLONE1_KEY"

# Verify cloned fields
CLONE1_DATA=$("$AJIRA_BIN" issue view "$CLONE1_KEY" --json)
CLONE1_SUMMARY=$(echo "$CLONE1_DATA" | jq -r '.summary')
CLONE1_TYPE=$(echo "$CLONE1_DATA" | jq -r '.type')
CLONE1_PRIORITY=$(echo "$CLONE1_DATA" | jq -r '.priority')

if [[ "$CLONE1_SUMMARY" != "$SOURCE_SUMMARY" ]]; then
    log_error "Summary mismatch: expected '$SOURCE_SUMMARY', got '$CLONE1_SUMMARY'"
    exit 1
fi
log_success "Summary matches"

if [[ "$CLONE1_TYPE" != "Task" ]]; then
    log_error "Type mismatch: expected 'Task', got '$CLONE1_TYPE'"
    exit 1
fi
log_success "Type matches"

if [[ "$CLONE1_PRIORITY" != "Major" ]]; then
    log_error "Priority mismatch: expected 'Major', got '$CLONE1_PRIORITY'"
    exit 1
fi
log_success "Priority matches"

# Test 3: Clone with overrides
log_step "Test 3: Clone with field overrides"
OVERRIDE_SUMMARY="Overridden Clone $TIMESTAMP"
CLONE2_KEY=$("$AJIRA_BIN" issue clone "$SOURCE_KEY" \
    -s "$OVERRIDE_SUMMARY" \
    --priority Minor \
    -L "override-label" \
    --json | jq -r '.clonedKey')

if [[ -z "$CLONE2_KEY" || "$CLONE2_KEY" == "null" ]]; then
    log_error "Failed to clone with overrides"
    exit 1
fi
CREATED_ISSUES+=("$CLONE2_KEY")
log_success "Cloned with overrides to: $CLONE2_KEY"

# Verify overridden fields
CLONE2_DATA=$("$AJIRA_BIN" issue view "$CLONE2_KEY" --json)
CLONE2_SUMMARY=$(echo "$CLONE2_DATA" | jq -r '.summary')
CLONE2_PRIORITY=$(echo "$CLONE2_DATA" | jq -r '.priority')

if [[ "$CLONE2_SUMMARY" != "$OVERRIDE_SUMMARY" ]]; then
    log_error "Override summary mismatch: expected '$OVERRIDE_SUMMARY', got '$CLONE2_SUMMARY'"
    exit 1
fi
log_success "Summary override works"

if [[ "$CLONE2_PRIORITY" != "Minor" ]]; then
    log_error "Override priority mismatch: expected 'Minor', got '$CLONE2_PRIORITY'"
    exit 1
fi
log_success "Priority override works"

# Test 4: Clone with link
log_step "Test 4: Clone with link to original"
CLONE3_RESULT=$("$AJIRA_BIN" issue clone "$SOURCE_KEY" --link --json)
CLONE3_KEY=$(echo "$CLONE3_RESULT" | jq -r '.clonedKey')
CLONE3_LINKED=$(echo "$CLONE3_RESULT" | jq -r '.linked')
CLONE3_LINKTYPE=$(echo "$CLONE3_RESULT" | jq -r '.linkType')

if [[ -z "$CLONE3_KEY" || "$CLONE3_KEY" == "null" ]]; then
    log_error "Failed to clone with link"
    exit 1
fi
CREATED_ISSUES+=("$CLONE3_KEY")
log_success "Cloned with link to: $CLONE3_KEY"

if [[ "$CLONE3_LINKED" != "true" ]]; then
    log_error "Link not created"
    exit 1
fi
log_success "Link created"

if [[ "$CLONE3_LINKTYPE" != "Clones" ]]; then
    log_warn "Link type: $CLONE3_LINKTYPE (expected 'Clones' - may vary by instance)"
fi

# Verify link exists on cloned issue
CLONE3_DATA=$("$AJIRA_BIN" issue view "$CLONE3_KEY" --json)
CLONE3_LINKS=$(echo "$CLONE3_DATA" | jq -r '.links | length')
if [[ "$CLONE3_LINKS" -lt 1 ]]; then
    log_error "No links found on cloned issue"
    exit 1
fi
log_success "Link visible on cloned issue"

# Test 5: Text output (URL)
log_step "Test 5: Text output returns URL"
CLONE4_OUTPUT=$("$AJIRA_BIN" issue clone "$SOURCE_KEY")
if [[ ! "$CLONE4_OUTPUT" =~ atlassian.net ]]; then
    log_error "Expected URL output, got: $CLONE4_OUTPUT"
    exit 1
fi
# Extract key from URL for cleanup
CLONE4_KEY=$(echo "$CLONE4_OUTPUT" | grep -oE '[A-Z]+-[0-9]+' | tail -1)
if [[ -n "$CLONE4_KEY" ]]; then
    CREATED_ISSUES+=("$CLONE4_KEY")
fi
log_success "Text output is URL: $CLONE4_OUTPUT"

# Summary
log_step "All tests passed!"
echo ""
log_info "Created issues:"
for issue in "${CREATED_ISSUES[@]}"; do
    echo "  - $issue"
done
