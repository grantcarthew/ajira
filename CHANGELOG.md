# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.0.0] - 2026-04-23

First stable release.

### Added

- Authentication and user verification (`me`, `user search`)
- Project discovery (`project list`, `field list`)
- Issue management: `list`, `view`, `create`, `edit`, `clone`, `delete`, `move`, `assign`, `watch`, `unwatch`, `open`
- Issue discovery metadata: `issue type`, `issue status`, `issue priority`
- Comments: `issue comment add`, `edit`, `list`
- Attachments: `issue attachment add`, `list`, `download`, `remove` with streaming upload and download
- Issue links: `issue link add`, `remove`, `list`, `types`, `url` (remote web links)
- Agile commands: `board list`, `sprint list`, `sprint add`, `epic list`, `epic create`, `epic add`, `epic remove`, `release list`
- Markdown to ADF conversion and ADF to Markdown rendering
- JSON output (`--json`) across all listing and detail commands
- Automation flags: `--dry-run`, `--quiet`, `--no-color`, `--verbose`
- Environment-variable configuration with shared `ATLASSIAN_*` and tool-specific `JIRA_*` overrides
- Token-efficient help topics for AI agents: `ajira help agents|agile|markdown|schemas`
- Shell completion generation for bash, zsh, and fish
- Homebrew installation via `grantcarthew/tap`

[Unreleased]: https://github.com/grantcarthew/ajira/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/grantcarthew/ajira/releases/tag/v1.0.0
