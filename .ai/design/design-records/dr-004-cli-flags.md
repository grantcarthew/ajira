# dr-004: CLI Global Flags

- Date: 2025-12-22
- Status: Accepted
- Category: CLI

## Problem

ajira needs consistent global flags across all commands for:

- Output format selection (text vs JSON)
- Default project specification
- Version and help information

Flag design affects usability and must be consistent with CLI conventions.

## Decision

Global persistent flags on root command:

| Flag | Short | Type | Default | Description |
| ---- | ----- | ---- | ------- | ----------- |
| --json | -j | bool | false | Output in JSON format |
| --project | -p | string | $JIRA_PROJECT | Default project key |
| --version | -v | bool | - | Display version information |
| --help | -h | bool | - | Display help (provided by Cobra) |

Flag behaviour:

- `--json` affects all commands that produce output
- `--project` can be overridden per-command or set via JIRA_PROJECT env var
- `--version` prints version and exits (standard Cobra behaviour)
- `--help` is automatic from Cobra

## Why

- Short flags (`-j`, `-p`) reduce typing for frequent operations
- `--json` is standard for machine-readable output in CLIs
- `--project` allows defaulting without repeated specification
- Persistent flags apply to all subcommands automatically

## Output Formats

Plain text (default):

- Human-readable, formatted for terminal
- One logical item per line where sensible
- No trailing punctuation

JSON (`--json`):

- Machine-readable, parseable
- Single JSON object or array per command
- No additional text or formatting

## Usage Examples

```bash
ajira me                    # Plain text output
ajira me -j                 # JSON output
ajira -p GCP issue list     # Use project GCP
ajira --version             # Show version
```

## Trade-offs

Accept:

- Global flags must be placed before subcommands (Cobra convention)
- Short flags are single letters only

Gain:

- Consistent interface across all commands
- Predictable behaviour for scripting
- Environment variable fallback reduces repetition

## Alternatives

Per-command output flags:

- Pro: More explicit
- Con: Repetitive, inconsistent
- Rejected: Global flag is simpler and standard

YAML output format:

- Pro: Human-readable structured format
- Con: Additional complexity
- Con: JSON is more universal for automation
- Rejected: JSON is sufficient, can add later if needed
