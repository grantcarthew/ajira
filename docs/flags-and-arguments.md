# Flags and Arguments Reference

Quick reference for all ajira command flags and arguments.

## Global Flags

These persistent flags are inherited by all commands.

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--json` | `-j` | bool | false | Output in JSON format |
| `--project` | `-p` | string | $JIRA_PROJECT | Default project key |
| `--board` | | string | $JIRA_BOARD | Default board ID for agile commands |

## Commands

### me

```
ajira me
```

No arguments or local flags.

### project list

```
ajira project list [flags]
```

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--query` | `-q` | string | | Filter by project name/key |
| `--limit` | `-l` | int | 0 | Maximum projects to return (0 = all) |

### issue list

```
ajira issue list [flags]
```

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--query` | `-q` | string | | JQL query (overrides other filters) |
| `--status` | | string | | Filter by status |
| `--type` | `-t` | string | | Filter by issue type |
| `--assignee` | `-a` | string | | Filter by assignee (email, accountId, 'me', 'unassigned') |
| `--reporter` | `-r` | string | | Filter by reporter (email, accountId, or 'me') |
| `--priority` | `-P` | string | | Filter by priority |
| `--labels` | `-L` | []string | | Filter by labels (comma-separated) |
| `--watching` | `-w` | bool | false | Filter to issues you are watching |
| `--order-by` | | string | updated | Sort field (created, updated, priority, key, rank) |
| `--reverse` | | bool | false | Reverse sort order (ASC instead of DESC) |
| `--limit` | `-l` | int | 50 | Maximum issues to return |
| `--sprint` | | string | | Filter by sprint ID |
| `--epic` | | string | | Filter by epic key |

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
| `--priority` | `-P` | string | | Issue priority |
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
| `--priority` | `-P` | string | | New priority |
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
| `--list` | | bool | false | List available transitions |

### issue clone

```
ajira issue clone <issue-key> [flags]
```

| Argument | Required | Description |
|----------|----------|-------------|
| `issue-key` | Yes | Issue key to clone |

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--summary` | `-s` | string | | Override summary |
| `--type` | `-t` | string | | Override issue type |
| `--priority` | `-P` | string | | Override priority |
| `--assignee` | `-a` | string | | Override assignee (email, accountId, 'me', 'unassigned') |
| `--reporter` | `-r` | string | | Override reporter (email, accountId, or 'me') |
| `--labels` | `-L` | []string | | Override labels (comma-separated) |
| `--link` | | string | | Link to original issue (default: Clones, or specify type) |

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

### board list

```
ajira board list [flags]
```

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--limit` | `-l` | int | 0 | Maximum boards to return (0 = all) |

### sprint list

```
ajira sprint list [flags]
```

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--state` | | string | | Filter by state (active, future, closed) |
| `--current` | | bool | false | Show current active sprints (shorthand for --state active) |
| `--limit` | `-l` | int | 0 | Maximum sprints to return (0 = all) |

Note: Requires `--board` flag or `JIRA_BOARD` environment variable.

### sprint add

```
ajira sprint add <sprint-id> <issue-keys...>
```

| Argument | Required | Description |
|----------|----------|-------------|
| `sprint-id` | Yes | Target sprint ID |
| `issue-keys` | Yes | One or more issue keys to add |

No local flags.

### epic list

```
ajira epic list [flags]
```

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--status` | | string | | Filter by status |
| `--assignee` | `-a` | string | | Filter by assignee (email, accountId, 'me', 'unassigned') |
| `--priority` | `-P` | string | | Filter by priority |
| `--limit` | `-l` | int | 50 | Maximum epics to return |

### epic create

```
ajira epic create [flags]
```

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--summary` | `-s` | string | | Epic summary (required) |
| `--description` | `-d` | string | | Epic description in Markdown |
| `--file` | `-f` | string | | Read description from file (- for stdin) |
| `--priority` | `-P` | string | | Epic priority |
| `--labels` | | []string | | Epic labels (comma-separated) |

### epic add

```
ajira epic add <epic-key> <issue-keys...>
```

| Argument | Required | Description |
|----------|----------|-------------|
| `epic-key` | Yes | Target epic key |
| `issue-keys` | Yes | One or more issue keys to add |

No local flags.

### epic remove

```
ajira epic remove <issue-keys...>
```

| Argument | Required | Description |
|----------|----------|-------------|
| `issue-keys` | Yes | One or more issue keys to remove from their epic |

No local flags.

## Short Flag Availability

Reserved globally: `-j`, `-p`

Used on `issue list`: `-q`, `-t`, `-a`, `-r`, `-P`, `-L`, `-w`, `-l`

Available for new flags on `issue list`: `-b`, `-c`, `-d`, `-e`, `-f`, `-g`, `-h`, `-i`, `-k`, `-m`, `-n`, `-o`, `-s`, `-u`, `-v`, `-x`, `-y`, `-z` and remaining uppercase variants.
