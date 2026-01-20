# ajira - Release Process

> **Purpose**: Repeatable process for releasing new versions of ajira
> **Audience**: AI agents and maintainers performing releases
> **Last Updated**: 2026-01-20

This document provides step-by-step instructions for releasing ajira. Execute each step in order.

---

## Prerequisites

Verify before starting:

- Write access to `gcarthew/ajira` repository
- Write access to `grantcarthew/homebrew-tap` repository
- Go 1.25.5+ installed
- Git configured with proper credentials
- GitHub CLI (`gh`) installed and authenticated
- All planned features/fixes merged to main branch

---

## Release Process

**Steps**:

1. Pre-release review
2. Run pre-release validation
3. Determine version number
4. Update CHANGELOG.md
5. Commit changes
6. Create and push git tag
7. Create GitHub Release
8. Update Homebrew tap
9. Verify installation
10. Clean up

**Estimated Time**: 25-35 minutes

---

## Step 1: Pre-Release Review

Perform a brief holistic review of the codebase before release. This is a quick glance to identify obvious issues, not a full code review.

**Review the following:**

1. **Active project status** - Check `AGENTS.md` for the active project. Verify it is complete and ready for release, or confirm no active project blocks the release.

2. **Recent changes** - Review commits since the last release tag. Look for:
   - Incomplete work (TODO, FIXME, XXX comments in changed files)
   - Obvious errors or missing error handling
   - Changes that lack corresponding tests

3. **Documentation currency** - Quick check that:
   - `README.md` reflects current functionality
   - Command help text matches implementation (`ajira help`)
   - `AGENTS.md` is accurate

4. **Code cleanliness** - Scan `internal/` for:
   - Dead code or commented-out blocks
   - Debug statements (fmt.Println, log.Println not part of normal output)
   - Hardcoded values that should be configurable

**Commands to assist review:**

```bash
# Find TODOs/FIXMEs in Go files
rg -i "TODO|FIXME|XXX" --type go

# Show commits since last release
PREV_VERSION=$(git tag -l | tail -1)
git log ${PREV_VERSION}..HEAD --oneline

# List recently modified Go files
git diff --name-only ${PREV_VERSION}..HEAD -- "*.go"
```

**Decision:** Report **GO** if no blocking issues found, or **NO-GO** with specific concerns that must be addressed before release.

---

## Step 2: Pre-Release Validation

Run validation checks:

```bash
# Ensure on main branch with latest changes
git checkout main
git pull origin main

# Check formatting (should produce no output)
gofmt -l .

# Run linters
go vet ./...
golangci-lint run
staticcheck ./...
ineffassign ./...
govulncheck ./...

# Check cyclomatic complexity (functions over 15)
gocyclo -over 15 .

# Verify all tests pass
go test -v ./...

# Verify build works
go build -o ajira ./cmd/ajira
./ajira --version
./ajira me  # Quick connectivity test
rm ajira

# Verify clean working directory
git status
```

**Expected results**:

- `gofmt -l .` produces no output (all files formatted)
- `go vet ./...` reports no issues
- `golangci-lint run` reports no errors (warnings acceptable)
- `staticcheck ./...` reports no issues
- `ineffassign ./...` reports no issues
- `govulncheck ./...` reports no vulnerabilities
- `gocyclo -over 15 .` reports no functions (or acceptable exceptions)
- All tests pass
- Build completes without errors
- `ajira --version` shows current version
- `ajira me` returns current user info
- `git status` shows clean working tree

**Linter installation** (if not already installed):

```bash
# golangci-lint (comprehensive linter)
brew install golangci-lint
# or: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Individual linters
go install honnef.co/go/tools/cmd/staticcheck@latest
go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
go install github.com/gordonklaus/ineffassign@latest
go install golang.org/x/vuln/cmd/govulncheck@latest
```

**If any validation fails, stop and fix issues before proceeding.**

---

## Step 3: Determine Version Number

Set the version number using [Semantic Versioning](https://semver.org/):

- **MAJOR**: Breaking API changes (1.0.0 → 2.0.0)
- **MINOR**: New features, backward compatible (1.0.0 → 1.1.0)
- **PATCH**: Bug fixes only (1.0.0 → 1.0.1)

```bash
# Check current version
git tag -l | tail -1

# Set new version (example: v0.0.1)
export VERSION="0.0.1"
echo "Releasing version: v${VERSION}"
```

---

## Step 4: Update CHANGELOG.md

> **Note:** Skip this step for v0.x.x releases. CHANGELOG.md is not maintained during initial development. Start maintaining the changelog from v1.0.0.

Review changes since last release and update CHANGELOG.md:

```bash
# Show changes since previous version (or all commits if first release)
PREV_VERSION=$(git tag -l | tail -1)
if [ -z "$PREV_VERSION" ]; then
  echo "First release - showing all commits:"
  git log --oneline
else
  echo "Changes since ${PREV_VERSION}:"
  git log ${PREV_VERSION}..HEAD --oneline
fi

# Review the changes and categorize them
# Then edit CHANGELOG.md manually
```

Update CHANGELOG.md by adding a new version section with:

- **Added**: New features
- **Changed**: Changes to existing functionality
- **Fixed**: Bug fixes
- **Deprecated**: Features marked for removal
- **Removed**: Removed features
- **Security**: Security fixes

Example format:

```markdown
## [0.0.1] - 2026-01-07

### Added

- Initial release with core Jira functionality
- `me` command for user info
- `project list` command
- `issue` commands: list, view, create, edit, delete, assign, move
- `issue comment add` command
- `issue status`, `issue type`, `issue priority` metadata commands
- Markdown input with ADF conversion
- JSON output support

[Unreleased]: https://github.com/gcarthew/ajira/compare/v0.0.1...HEAD
[0.0.1]: https://github.com/gcarthew/ajira/releases/tag/v0.0.1
```

---

## Step 5: Commit Changes

> **Note:** For v0.x.x releases, skip the CHANGELOG commit but still verify clean state.

Commit the CHANGELOG (if updated):

```bash
# Stage and commit changes
git add CHANGELOG.md
git commit -m "chore: prepare for v${VERSION} release"
git push origin main
```

Verify clean working directory before tagging:

```bash
git status  # Should show "nothing to commit, working tree clean"
```

---

## Step 6: Create and Push Git Tag

Create an annotated git tag:

```bash
# Get previous version and review changes
PREV_VERSION=$(git tag -l | tail -1)
if [ -z "$PREV_VERSION" ]; then
  git log --oneline | head -10
else
  git log ${PREV_VERSION}..HEAD --oneline
fi

# Create one-line summary from the changes above
# Examples: "Initial release", "Add issue linking"
SUMMARY="Your one-line summary here"

# Create and push annotated tag
git tag -a "v${VERSION}" -m "Release v${VERSION} - ${SUMMARY}"
git push origin "v${VERSION}"

# Verify tag exists
git tag -l -n9 "v${VERSION}"
```

---

## Step 7: Create GitHub Release

Create the GitHub Release with release notes:

```bash
# Wait for tarball to be generated (usually immediate)
sleep 5

# Get tarball SHA256 for Homebrew (will use in Step 7)
TARBALL_URL="https://github.com/gcarthew/ajira/archive/refs/tags/v${VERSION}.tar.gz"
# macOS:
TARBALL_SHA256=$(curl -sL "$TARBALL_URL" | shasum -a 256 | cut -d' ' -f1)
# Linux:
# TARBALL_SHA256=$(curl -sL "$TARBALL_URL" | sha256sum | cut -d' ' -f1)
echo "Tarball SHA256: $TARBALL_SHA256"

# Create GitHub Release using gh CLI
PREV_VERSION=$(git tag -l | grep -v "v${VERSION}" | tail -1)
if [ -z "$PREV_VERSION" ]; then
  # First release
  NOTES=$(git log --pretty=format:"- %s" --reverse | head -20)
else
  NOTES=$(git log ${PREV_VERSION}..v${VERSION} --pretty=format:"- %s" --reverse)
fi

gh release create "v${VERSION}" \
  --title "Release v${VERSION}" \
  --notes "$(cat <<EOF
## Changes

${NOTES}

See [CHANGELOG.md](https://github.com/gcarthew/ajira/blob/main/CHANGELOG.md) for details.
EOF
)"

# Verify release was created
gh release view "v${VERSION}"
```

**Note**: GitHub automatically attaches source archives (tar.gz, zip) to releases. Homebrew builds from the tar.gz archive.

---

## Step 8: Update Homebrew Tap

Update the Homebrew formula with the new version:

```bash
# Navigate to homebrew-tap directory
cd ~/Projects/homebrew-tap
git pull origin main

# Display tarball info from Step 6
echo "Tarball URL: $TARBALL_URL"
echo "Tarball SHA256: $TARBALL_SHA256"

# Edit Formula/ajira.rb and update:
# 1. url line: Update version in URL
# 2. sha256 line: Update with TARBALL_SHA256
# 3. ldflags: Update version in "-X github.com/gcarthew/ajira/internal/cli.Version=X.X.X"
# 4. test: Update expected version in assert_match

# After editing, commit and push
git add Formula/ajira.rb
git commit -m "ajira: update to ${VERSION}"
git push origin main

# Return to ajira directory
cd -
```

**Formula example** (Formula/ajira.rb):

```ruby
class Ajira < Formula
  desc "Atlassian Jira CLI for AI agents and automation"
  homepage "https://github.com/gcarthew/ajira"
  url "https://github.com/gcarthew/ajira/archive/refs/tags/v0.0.1.tar.gz"
  sha256 "abc123..."  # Use TARBALL_SHA256 value
  license "MIT"

  depends_on "go" => :build

  def install
    ldflags = "-s -w -X github.com/gcarthew/ajira/internal/cli.Version=#{version}"
    system "go", "build", *std_go_args(ldflags: ldflags), "./cmd/ajira"
  end

  test do
    assert_match "0.0.1", shell_output("#{bin}/ajira --version")
  end
end
```

---

## Step 9: Verify Installation

Test the Homebrew installation:

```bash
# Update and reinstall
brew update
brew reinstall grantcarthew/tap/ajira

# Verify version
ajira --version  # Should show new version

# Test basic functionality (requires JIRA_* env vars)
ajira me
```

**Expected results**:

- `ajira --version` displays new version
- `ajira me` returns user info (if configured)
- No errors during installation

**If installation fails**, debug with:

```bash
brew audit --strict grantcarthew/tap/ajira
brew install --verbose grantcarthew/tap/ajira
```

---

## Step 10: Clean Up

Complete the release:

```bash
# Verify release is live
gh release view "v${VERSION}"

# Check Homebrew tap was updated
cd ~/Projects/homebrew-tap
git log -1
cd -

# Verify clean state
git status
```

**Release is complete!**

Monitor for issues:

- Watch GitHub issues for bug reports
- Monitor Homebrew installation feedback
- Be ready to release a patch if critical issues arise

---

## Rollback Procedure

If critical issues are discovered after release:

**Option 1: Patch Release** (Recommended)

```bash
# Fix the issue, then release patch version (e.g., v0.0.2)
# Follow the standard release process
```

**Option 2: Delete Release** (Last resort - use only for critical security issues)

```bash
# Delete GitHub release
gh release delete "v${VERSION}" --yes

# Delete tags
git push origin --delete "v${VERSION}"
git tag -d "v${VERSION}"

# Revert Homebrew tap
cd ~/Projects/homebrew-tap
git revert HEAD
git push origin main
cd -
```

---

## Quick Reference

One-command release workflow:

```bash
# Set version
export VERSION="0.0.1"

# Get previous version for change summary
PREV_VERSION=$(git tag -l | tail -1)

# 1. Pre-release review (see Step 1 for details)
rg -i "TODO|FIXME|XXX" --type go  # Should be empty or acceptable

# 2. Validation
go test -v ./...
golangci-lint run
staticcheck ./...
govulncheck ./...
git status  # Should be clean

# 3. Update CHANGELOG.md (skip for v0.x.x releases)
# git add CHANGELOG.md
# git commit -m "chore: prepare for v${VERSION} release"
# git push origin main
git status  # Verify clean working directory

# 4. Create tag with summary
SUMMARY="Your summary here"
git tag -a "v${VERSION}" -m "Release v${VERSION} - ${SUMMARY}"
git push origin "v${VERSION}"

# 5. Create GitHub Release
gh release create "v${VERSION}" --title "Release v${VERSION}" \
  --notes "$(git log ${PREV_VERSION}..v${VERSION} --pretty=format:'- %s' 2>/dev/null || git log --pretty=format:'- %s' | head -20)"

# 6. Get tarball SHA256 (macOS)
TARBALL_SHA256=$(curl -sL "https://github.com/gcarthew/ajira/archive/refs/tags/v${VERSION}.tar.gz" | shasum -a 256 | cut -d' ' -f1)
echo "SHA256: $TARBALL_SHA256"

# 7. Update Homebrew (edit Formula/ajira.rb with VERSION and SHA256)
cd ~/Projects/homebrew-tap
# Edit Formula/ajira.rb
git add Formula/ajira.rb
git commit -m "ajira: update to ${VERSION}"
git push origin main
cd -

# 8. Test
brew update && brew reinstall grantcarthew/tap/ajira
ajira --version
```

---

## Troubleshooting

**Tests failing**

- Run: `go test -v ./...` to see detailed output
- Fix all failures before proceeding
- Never release with failing tests

**Tarball not available**

- Wait 1-2 minutes after pushing tag
- Verify tag exists: `git ls-remote --tags origin | grep v${VERSION}`
- Check: <https://github.com/gcarthew/ajira/tags>

**Homebrew formula issues**

- Audit: `brew audit --strict grantcarthew/tap/ajira`
- Common: Incorrect SHA256, wrong URL format, Ruby syntax
- Fix and push updated formula

**Installation fails**

- Verbose output: `brew install --verbose grantcarthew/tap/ajira`
- View formula: `brew cat grantcarthew/tap/ajira`
- Verify tarball: `curl -I https://github.com/gcarthew/ajira/archive/refs/tags/v${VERSION}.tar.gz`

---

## Related Documents

- `AGENTS.md` - Repository context for AI agents
- `CHANGELOG.md` - Version history
- `.ai/design/design-records/` - Design decisions and rationale
- `README.md` - User-facing documentation

---

**End of Release Process**
