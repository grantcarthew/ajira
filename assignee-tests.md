# Assignee Refactor Manual Tests

Manual test plan covering the staged changes: `-a/--assignee` flag on `issue create` and `epic create`, shared helper extraction (`resolveAssigneeInput`, `readText`, `width.PadRight`), and error-propagation cleanup.

## New `-a` assignee flag (issue create, epic create)

```bash
# Assign to yourself
./ajira issue create -s "Test assign me" -a me

# Assign by email
./ajira issue create -s "Test assign email" -a your.colleague@example.com

# Explicitly unassigned
./ajira issue create -s "Test unassigned" -a unassigned

# Assign by account ID (20+ char string skips search)
./ajira issue create -s "Test assign id" -a 5b10ac8d82e05b22cc7d4ef5

# Bad user should error cleanly
./ajira issue create -s "Test bad user" -a nobody@nowhere.invalid

# Same for epic
./ajira epic create -s "Test epic assign" -a me
```

## readText refactor (file/stdin/body)

```bash
# File path
echo "# From file" > /tmp/desc.md
./ajira issue create -s "File desc" -f /tmp/desc.md

# Stdin via "-"
echo "# From stdin" | ./ajira issue create -s "Stdin desc" -f -

# Body flag
./ajira issue create -s "Body desc" -d "Inline body"

# Comment add — file / stdin / body / positional
./ajira issue comment add PROJ-1 -f /tmp/desc.md
echo "stdin comment" | ./ajira issue comment add PROJ-1 -f -
./ajira issue comment add PROJ-1 -b "body comment"
./ajira issue comment add PROJ-1 "positional comment"

# Epic create from file
./ajira epic create -s "Epic from file" -f /tmp/desc.md
```

## width.PadRight (column rendering)

Run each list command and confirm columns align, including with non-ASCII issue titles or board names if available:

```bash
./ajira issue list -l 10
./ajira epic list -l 10
./ajira board list
./ajira sprint list
./ajira project list
./ajira issue link types
./ajira issue priority
./ajira issue status
./ajira issue type
```

## issue clone -a me empty-email guard

```bash
# Should error with "JIRA_EMAIL is required to resolve 'me'"
JIRA_EMAIL= ./ajira issue clone PROJ-1 -a me

# Normal path still works
./ajira issue clone PROJ-1 -a me
```

## User search Active field

```bash
# Inactive users should now show the correct active flag (previously always true)
./ajira user search "someone" --json | jq '.[] | {displayName, active}'
```

## Metadata parser (removed array-format fallback)

Any command that validates issue type exercises this path; confirm nothing regressed:

```bash
./ajira issue create -s "Type check" -t Task
./ajira epic create -s "Epic type check"
```

## Quick smoke regression

```bash
./ajira me
./ajira project list
./ajira issue list -l 3
go test ./...
```
