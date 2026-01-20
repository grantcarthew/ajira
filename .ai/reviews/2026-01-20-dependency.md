# Dependency Review

- Date: 2026-01-20
- Reviewer: Claude (Opus 4.5)
- Scope: Third-party package usage and dependency management

## Executive Summary

The ajira project has **6 direct dependencies** and **23 indirect dependencies** - a reasonable footprint for a CLI application. All dependencies are necessary, well-maintained, and use permissive licenses compatible with MPL-2.0.

**Dependency Hygiene: Excellent**

---

## 1. Hygiene Checks

| Check | Status |
|-------|--------|
| `go mod tidy` | Clean - no changes needed |
| `go mod verify` | All modules verified |
| `govulncheck` | No vulnerabilities found |
| go.sum in VCS | Yes |

**Status:** Good

---

## 2. License Analysis

All dependencies use permissive licenses compatible with MPL-2.0:

| License | Count | Packages |
|---------|-------|----------|
| MIT | 20 | charmbracelet/*, fatih/color, goldmark, mattn/*, muesli/*, etc. |
| BSD-3-Clause | 7 | google/uuid, gorilla/css, bluemonday, spf13/pflag, golang.org/x/* |
| Apache-2.0 | 1 | spf13/cobra |

**Status:** Good - no license concerns

---

## 3. Direct Dependencies

### charmbracelet/glamour v0.10.0

- **Purpose:** Terminal Markdown rendering
- **Used in:** `internal/cli/root.go`
- **Justification:** No stdlib equivalent for styled terminal Markdown output
- **Maintenance:** Actively maintained by Charm (organisation)
- **Version:** Latest stable
- **Assessment:** Appropriate

### fatih/color v1.18.0

- **Purpose:** Terminal colour output
- **Used in:** 8 CLI files (board, issue lists, types, statuses, priorities)
- **Justification:** Provides cross-platform colour support with NO_COLOR respect
- **Maintenance:** Actively maintained, widely used
- **Version:** Latest stable
- **Assessment:** Appropriate

### google/uuid v1.6.0

- **Purpose:** UUID generation
- **Used in:** `internal/converter/markdown_to_adf.go`
- **Justification:** Generates UUIDs for ADF task list localIds
- **Maintenance:** Maintained by Google
- **Version:** Latest stable
- **Assessment:** Appropriate

### spf13/cobra v1.10.2

- **Purpose:** CLI framework
- **Used in:** All command implementations
- **Justification:** Industry-standard Go CLI framework
- **Maintenance:** Actively maintained, widely adopted
- **Version:** Latest stable
- **Assessment:** Appropriate

### yuin/goldmark v1.7.14

- **Purpose:** Markdown parsing
- **Used in:** `internal/converter/markdown_to_adf.go`
- **Justification:** Required for Markdown-to-ADF conversion
- **Maintenance:** Actively maintained
- **Version:** **Update available: v1.7.16**
- **Assessment:** Appropriate

### golang.org/x/term v0.38.0

- **Purpose:** Terminal handling
- **Used in:** Terminal width detection
- **Justification:** Official Go sub-repository
- **Maintenance:** Maintained by Go team
- **Version:** **Update available: v0.39.0**
- **Assessment:** Appropriate

---

## 4. Transitive Dependencies

Total indirect dependencies: 23

**Dependency tree analysis:**

- `glamour` is the largest contributor (pulls in lipgloss, chroma, bluemonday, etc.)
- `cobra` pulls in pflag and mousetrap
- No circular dependencies
- No version conflicts

**Notable transitives:**

| Package | Source | Notes |
|---------|--------|-------|
| alecthomas/chroma | glamour | Syntax highlighting |
| microcosm-cc/bluemonday | glamour | HTML sanitisation |
| charmbracelet/lipgloss | glamour | Terminal styling |
| xo/terminfo | lipgloss chain | Stable terminfo reader (pseudo-version from 2022 is normal for mature package) |

---

## 5. Pre-release Versions

glamour v0.10.0 pulls in pre-release versions of Charmbracelet packages:

```
charmbracelet/colorprofile@v0.2.3-0.20250311203215-f60798e515dc
charmbracelet/lipgloss@v1.1.1-0.20250404203927-76690c660834
charmbracelet/x/exp/slice@v0.0.0-20250327172914-2fdc97757edf
```

**Assessment:** Normal for Charmbracelet ecosystem. These are controlled by glamour's go.mod, not directly by this project. Not a concern as glamour is a stable release.

---

## 6. Red Flags Check

| Criterion | Status |
|-----------|--------|
| Dependencies with no updates in 2+ years | None |
| Single-maintainer critical dependencies | None |
| Excessive transitive dependencies | No (23 is reasonable) |
| Known unfixed vulnerabilities | None |
| Problematic licenses | None |
| Duplicate stdlib functionality | None |

**Status:** No red flags

---

## 7. Recommendations

### Minor updates available

```bash
go get github.com/yuin/goldmark@v1.7.16
go get golang.org/x/term@v0.39.0
go mod tidy
```

### Ongoing maintenance

- Run `govulncheck ./...` periodically or in CI
- Consider dependabot or renovate for automated updates
- Review major version upgrades carefully before adopting

---

## Verdict

**Dependency hygiene is excellent.** The project uses well-chosen, minimal, well-maintained dependencies with no security or licensing concerns. The dependency footprint is appropriate for the functionality provided.
