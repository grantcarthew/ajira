#!/bin/bash
#
# Round-trip test for ajira Markdown to ADF conversion
#
# Tests the full cycle: Markdown -> Jira ADF -> Markdown
# Compares key features to verify conversion fidelity
#
# Usage: ./testdata/roundtrip-test.sh [-y|--yes]
#
# Options:
#   -y, --yes    Auto-delete test issue without prompting
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
            echo "  -y, --yes    Auto-delete test issue without prompting"
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
TEST_FILE="$SCRIPT_DIR/comprehensive-markdown.md"
AJIRA_BIN="$PROJECT_DIR/ajira"
TIMESTAMP=$(date +%Y%m%d-%H%M%S)
ISSUE_SUMMARY="Round-Trip Test $TIMESTAMP"

log_step "Starting round-trip test"
log_info "Test file: $TEST_FILE"
log_info "Issue summary: $ISSUE_SUMMARY"

# Check test file exists
if [[ ! -f "$TEST_FILE" ]]; then
    log_error "Test file not found: $TEST_FILE"
    exit 1
fi
log_success "Test file exists"

# Check environment variables
log_step "Checking environment configuration"
if [[ -z "${JIRA_BASE_URL:-}" ]]; then
    log_error "JIRA_BASE_URL is not set"
    exit 1
fi
log_success "JIRA_BASE_URL is set"

if [[ -z "${JIRA_EMAIL:-}" ]]; then
    log_error "JIRA_EMAIL is not set"
    exit 1
fi
log_success "JIRA_EMAIL is set"

if [[ -z "${JIRA_API_TOKEN:-}${ATLASSIAN_API_TOKEN:-}" ]]; then
    log_error "No API token set (JIRA_API_TOKEN or ATLASSIAN_API_TOKEN)"
    exit 1
fi
log_success "API token is set"

if [[ -z "${JIRA_PROJECT:-}" ]]; then
    log_error "JIRA_PROJECT is not set"
    exit 1
fi
log_success "JIRA_PROJECT is set: $JIRA_PROJECT"

# Build ajira if needed
log_step "Building ajira"
cd "$PROJECT_DIR"
if [[ ! -f "$AJIRA_BIN" ]] || [[ -n "$(find cmd internal -name '*.go' -newer "$AJIRA_BIN" 2>/dev/null)" ]]; then
    log_info "Source files changed, rebuilding..."
    go build -o ajira ./cmd/ajira
    log_success "Build completed"
else
    log_info "Binary is up to date, skipping build"
fi

# Verify ajira works
AJIRA_VERSION=$("$AJIRA_BIN" --version 2>&1 || echo "unknown")
log_info "ajira version: $AJIRA_VERSION"

# Create the test issue
log_step "Creating test issue in Jira"
log_info "Uploading comprehensive-markdown.md as description..."

CREATE_OUTPUT=$(cat "$TEST_FILE" | "$AJIRA_BIN" issue create -s "$ISSUE_SUMMARY" -f - 2>&1)

# Extract issue key from output (URL format: https://xxx.atlassian.net/browse/PROJ-123)
ISSUE_URL=$(echo "$CREATE_OUTPUT" | grep -E "^https://" | head -1)
ISSUE_KEY=$(echo "$ISSUE_URL" | grep -oE '[A-Z]+-[0-9]+' | tail -1)

if [[ -z "$ISSUE_KEY" ]]; then
    log_error "Failed to create issue"
    echo "$CREATE_OUTPUT"
    exit 1
fi

log_success "Issue created successfully"
log_info "Issue key: $ISSUE_KEY"
log_info "Issue URL: $ISSUE_URL"

# View the issue back
log_step "Retrieving issue from Jira"
log_info "Fetching issue content as Markdown..."

TEMP_DIR=$(mktemp -d)
RETRIEVED_FILE="$TEMP_DIR/retrieved.md"

# Get the JSON output to extract description
"$AJIRA_BIN" issue view "$ISSUE_KEY" --json | jq -r '.description // ""' > "$RETRIEVED_FILE"

ORIGINAL_LINES=$(wc -l < "$TEST_FILE" | tr -d ' ')
RETRIEVED_LINES=$(wc -l < "$RETRIEVED_FILE" | tr -d ' ')

log_success "Issue retrieved successfully"
log_info "Original file: $ORIGINAL_LINES lines"
log_info "Retrieved description: $RETRIEVED_LINES lines"

# Run feature checks
log_step "Verifying round-trip conversion"

PASS_COUNT=0
FAIL_COUNT=0
WARN_COUNT=0

check_feature() {
    local name="$1"
    local pattern="$2"
    local file="$RETRIEVED_FILE"

    if grep -qE "$pattern" "$file"; then
        log_success "$name"
        PASS_COUNT=$((PASS_COUNT + 1))
    else
        log_error "$name - pattern not found: $pattern"
        FAIL_COUNT=$((FAIL_COUNT + 1))
    fi
}

check_feature_warn() {
    local name="$1"
    local pattern="$2"
    local file="$RETRIEVED_FILE"

    if grep -qE "$pattern" "$file"; then
        log_success "$name"
        PASS_COUNT=$((PASS_COUNT + 1))
    else
        log_warn "$name - pattern not found (known limitation)"
        WARN_COUNT=$((WARN_COUNT + 1))
    fi
}

# Text formatting
log_info "Checking text formatting..."
check_feature "Bold text" "\*\*bold text\*\*"
check_feature "Italic text" "\*italic text\*|_italic text_"
check_feature "Bold+italic text" "\*\*\*bold italic text\*\*\*"
check_feature "Strikethrough" "~~strikethrough~~"
check_feature "Inline code" "\`inline code\`"

# Headings
log_info "Checking headings..."
check_feature "H2 heading" "^## Text Formatting"
check_feature "H3 heading" "^### Basic Formatting"
check_feature "H4 heading" "^#### This is H4"
check_feature "H5 heading" "^##### This is H5"
check_feature "H6 heading" "^###### This is H6"

# Code blocks
log_info "Checking code blocks..."
check_feature "Go code block" '```go'
check_feature "Python code block" '```python'
check_feature "JavaScript code block" '```javascript'
check_feature "Bash code block" '```bash'
check_feature "HTML code block" '```html'
check_feature "Code block without language" '```$'
check_feature "Special chars in code (<, >, &)" "a < b && c > d"
check_feature "Backslashes in code" 'C:\\Users\\test'
check_feature "Regex in code" 'MustCompile'
check_feature "Quotes in code" 'She said \\"Hello\\"'
check_feature_warn "Indented code block" "4 spaces instead of fences"
check_feature_warn "Empty code block" '^```$'

# Lists
log_info "Checking lists..."
check_feature "Unordered list" "^- First item|^\* First item"
check_feature "Ordered list" "^1\. Step one"
check_feature "Nested list (child)" "Child item"
check_feature "Nested list (grandchild)" "Grandchild"
check_feature "Mixed nested list" "Mixed deep nesting"
check_feature "Deeply nested (Level 5)" "Level 5"

# Task lists
log_info "Checking task lists..."
check_feature_warn "Unchecked task" "\- \[ \]|\* \[ \]"
check_feature_warn "Checked task" "\- \[x\]|\* \[x\]"

# Tables
log_info "Checking tables..."
check_feature "Table header" "Feature.*Status.*Notes"
check_feature "Table row" "Headings.*Working"
check_feature "Table with code" "fmt\.Println"
check_feature_warn "Table empty cells" "Empty middle"
check_feature_warn "Table escaped pipes" "OR operation"
check_feature_warn "Table formatted headers" "Bold Header"
check_feature_warn "Table alignment" "Left Aligned.*Center Aligned.*Right Aligned"

# Links
log_info "Checking links..."
check_feature "External link" "\[Atlassian Documentation\]"
check_feature "Link URL" "https://developer.atlassian.com"
check_feature "Multiple links" "\[GitHub\].*\[Stack Overflow\]"
check_feature_warn "Link with title" "\[Go Documentation\]"
check_feature "Link with special chars" "search\?q=foo"
check_feature_warn "AutoLinks" "https://example.com"
check_feature "Reference-style link" "\[reference link\]"

# Blockquotes
log_info "Checking blockquotes..."
check_feature "Simple blockquote" "^> This is a simple blockquote"
check_feature "Blockquote with formatting" "^>.*\*\*Note:\*\*"
check_feature_warn "Nested blockquotes" "Outer blockquote level one"
check_feature "Blockquote with list" "First item in quote"
check_feature_warn "Blockquote with code" "Hello from blockquote"

# Horizontal rules
log_info "Checking horizontal rules..."
check_feature "Horizontal rule" "^---$|^-{3,}$"

# Unicode
log_info "Checking unicode..."
check_feature "Japanese text" "こんにちは世界"
check_feature "Chinese text" "你好世界"
check_feature "Korean text" "안녕하세요"
check_feature "Russian text" "Привет мир"
check_feature "Arabic text" "مرحبا بالعالم"
check_feature "Unicode emoji" "🚀"

# Special characters
log_info "Checking special characters..."
check_feature "Ampersand" "Ampersands &"
check_feature "Angle brackets" "< >"
check_feature "Quotes" '"double"'
check_feature "Mathematical symbols" "×|π|∞"
check_feature "Arrows" "→|←|↑|↓"

# Edge cases
log_info "Checking edge cases..."
check_feature_warn "Escaped asterisk" '\\\*not italic\\\*'
check_feature_warn "Escaped backtick" '\\\`not code\\\`'
check_feature_warn "Hard line break" "forces a line break"
check_feature_warn "Double-backtick code" "code with.*backtick"
check_feature "Long line preserved" "extremely long line"
check_feature "Consecutive code blocks" "First code block"
check_feature_warn "Paragraph breaks" "This is paragraph two"

# Cleanup temp files
rm -rf "$TEMP_DIR"

# Summary
log_step "Test Summary"
echo ""
TOTAL=$((PASS_COUNT + FAIL_COUNT + WARN_COUNT))
log_info "Total checks: $TOTAL"
log_info "Passed: $PASS_COUNT"
if [[ $WARN_COUNT -gt 0 ]]; then
    log_info "Warnings: $WARN_COUNT (known ADF limitations)"
fi
if [[ $FAIL_COUNT -gt 0 ]]; then
    log_info "Failed: $FAIL_COUNT"
fi
echo ""
log_info "Issue URL: $ISSUE_URL"
log_info "Issue key: $ISSUE_KEY"
echo ""

# Open in browser
log_step "Opening issue in browser"
if [[ "$OSTYPE" == "darwin"* ]]; then
    open "$ISSUE_URL"
    log_success "Opened in default browser (macOS)"
elif command -v xdg-open &> /dev/null; then
    xdg-open "$ISSUE_URL"
    log_success "Opened in default browser (Linux)"
elif command -v wslview &> /dev/null; then
    wslview "$ISSUE_URL"
    log_success "Opened in default browser (WSL)"
else
    log_warn "Could not detect browser opener - please open URL manually"
fi

# Cleanup
echo ""
log_step "Cleanup"

# Safety check: verify issue summary matches before deleting
delete_issue() {
    local key="$1"
    local expected_summary="$ISSUE_SUMMARY"

    # Fetch the issue and verify summary matches
    local actual_summary
    actual_summary=$("$AJIRA_BIN" issue view "$key" --json 2>/dev/null | jq -r '.summary // ""')

    if [[ -z "$actual_summary" ]]; then
        log_warn "Could not fetch issue $key for verification - skipping delete"
        return 1
    fi

    if [[ "$actual_summary" != "$expected_summary" ]]; then
        log_error "Safety check failed: issue summary does not match"
        log_info "Expected: $expected_summary"
        log_info "Actual:   $actual_summary"
        log_warn "Refusing to delete - please verify and delete manually"
        return 1
    fi

    log_info "Verified: issue summary matches test issue"
    if "$AJIRA_BIN" issue delete "$key" 2>/dev/null; then
        log_success "Test issue deleted"
        return 0
    else
        log_warn "Failed to delete issue - please delete manually"
        return 1
    fi
}

if [[ "$AUTO_DELETE" == "true" ]]; then
    delete_issue "$ISSUE_KEY"
else
    read -p "Delete test issue $ISSUE_KEY? [y/N] " -n 1 -r
    echo ""
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        delete_issue "$ISSUE_KEY"
    else
        log_info "Issue preserved: $ISSUE_KEY"
        log_info "Delete manually with: ajira issue delete $ISSUE_KEY"
    fi
fi

# Exit with appropriate code
if [[ $FAIL_COUNT -gt 0 ]]; then
    exit 1
else
    exit 0
fi
