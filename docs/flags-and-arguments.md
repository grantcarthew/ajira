# Flags and Arguments Reference

Quick reference for all ajira command flags and arguments.

## Global Flags

These persistent flags are inherited by all commands.

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--json` | `-j` | bool | false | Output in JSON format |
| `--project` | `-p` | string | $JIRA_PROJECT | Default project key |

## Commands

### me

```
ajira me
```

No arguments or local flags.

### project list

```
ajira project list
```

No arguments or local flags.

### issue list

```
ajira issue list [flags]
```

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--query` | `-q` | string | | JQL query (overrides other filters) |
| `--status` | `-s` | string | | Filter by status |
| `--type` | `-t` | string | | Filter by issue type |
| `--assignee` | `-a` | string | | Filter by assignee (email, accountId, 'me', 'unassigned') |
| `--reporter` | `-r` | string | | Filter by reporter (email, accountId, or 'me') |
| `--priority` | `-P` | string | | Filter by priority |
| `--labels` | `-L` | []string | | Filter by labels (comma-separated) |
| `--watching` | `-w` | bool | false | Filter to issues you are watching |
| `--order-by` | | string | updated | Sort field (created, updated, priority, key, rank) |
| `--reverse` | | bool | false | Reverse sort order (ASC instead of DESC) |
| `--limit` | `-l` | int | 50 | Maximum issues to return |

### issue view

```
ajira issue view <issue-key> [flags]
```

| Argument | Required | Description |
|----------|----------|-------------|
| `issue-key` | Yes | Issue key (e.g., PROJ-123) |

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--comments` | `-c` | int | 0 | Number of recent comments to show |

### issue create

```
ajira issue create [flags]
```

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--summary` | `-s` | string | | Issue summary (required) |
| `--description` | `-d` | string | | Issue description in Markdown |
| `--file` | `-f` | string | | Read description from file (- for stdin) |
| `--type` | `-t` | string | Task | Issue type |
| `--priority` | | string | | Issue priority |
| `--labels` | | []string | | Issue labels (comma-separated) |

### issue edit

```
ajira issue edit <issue-key> [flags]
```

| Argument | Required | Description |
|----------|----------|-------------|
| `issue-key` | Yes | Issue key to edit |

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--summary` | `-s` | string | | New issue summary |
| `--description` | `-d` | string | | New description in Markdown |
| `--file` | `-f` | string | | Read description from file (- for stdin) |
| `--type` | `-t` | string | | New issue type |
| `--priority` | | string | | New priority |
| `--labels` | | []string | | New labels (replaces existing) |

### issue assign

```
ajira issue assign <issue-key> <user>
```

| Argument | Required | Description |
|----------|----------|-------------|
| `issue-key` | Yes | Issue key to assign |
| `user` | Yes | User email, accountId, 'me', or 'unassigned' |

No local flags.

### issue move

```
ajira issue move <issue-key> [status] [flags]
```

| Argument | Required | Description |
|----------|----------|-------------|
| `issue-key` | Yes | Issue key to transition |
| `status` | No | Target status (omit to list available) |

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--list` | `-l` | bool | false | List available transitions |

### issue delete

```
ajira issue delete <issue-key>
```

| Argument | Required | Description |
|----------|----------|-------------|
| `issue-key` | Yes | Issue key to delete |

No local flags.

### issue comment add

```
ajira issue comment add <issue-key> [text] [flags]
```

| Argument | Required | Description |
|----------|----------|-------------|
| `issue-key` | Yes | Issue key |
| `text` | No | Comment text (alternative to flags) |

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--body` | `-b` | string | | Comment text in Markdown |
| `--file` | `-f` | string | | Read comment from file (- for stdin) |

### issue link add

```
ajira issue link add <outward-key> <link-type> <inward-key>
```

| Argument | Required | Description |
|----------|----------|-------------|
| `outward-key` | Yes | Issue that performs the action |
| `link-type` | Yes | Link type name (e.g., Blocks, Duplicate) |
| `inward-key` | Yes | Issue that receives the action |

No local flags.

### issue link remove

```
ajira issue link remove <issue-key> <link-type> <linked-key>
```

| Argument | Required | Description |
|----------|----------|-------------|
| `issue-key` | Yes | Issue containing the link |
| `link-type` | Yes | Link type name |
| `linked-key` | Yes | Linked issue key |

No local flags.

### issue link url

```
ajira issue link url <issue-key> <url> [title]
```

| Argument | Required | Description |
|----------|----------|-------------|
| `issue-key` | Yes | Issue key |
| `url` | Yes | URL to link |
| `title` | No | Link title (defaults to URL) |

No local flags.

### issue link types

```
ajira issue link types
```

No arguments or local flags.

### issue type

```
ajira issue type
```

No arguments or local flags.

### issue status

```
ajira issue status
```

No arguments or local flags.

### issue priority

```
ajira issue priority
```

No arguments or local flags.

## Short Flag Availability

Reserved globally: `-j`, `-p`

Used on `issue list`: `-q`, `-s`, `-t`, `-a`, `-r`, `-P`, `-L`, `-w`, `-l`

Available for new flags on `issue list`: `-b`, `-c`, `-d`, `-e`, `-f`, `-g`, `-h`, `-i`, `-k`, `-m`, `-n`, `-o`, `-u`, `-v`, `-x`, `-y`, `-z` and remaining uppercase variants.
