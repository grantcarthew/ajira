# DR-013: Issue Field Update Patterns

- Date: 2026-01-13
- Status: Accepted
- Category: CLI

## Problem

The issue commands (create, edit, move, delete) need enhancements to support additional Jira fields and operations:

- Parent/epic assignment for create and edit
- Components and fix versions for create and edit
- Label add/remove without replacing all labels
- Transition comments, resolution, and assignee changes during move
- Cascade delete for issues with subtasks

The Jira API uses two distinct patterns for field updates: the `fields` object for setting values and the `update` object for add/remove operations. The CLI needs a consistent approach for exposing these capabilities.

## Decision

### API Update Patterns

Use two request structures based on operation type:

Fields object (replace/set operations):

```json
{
  "fields": {
    "parent": {"key": "GCP-50"},
    "components": [{"name": "Backend"}],
    "fixVersions": [{"name": "1.0.0"}]
  }
}
```

Update object (add/remove operations):

```json
{
  "update": {
    "labels": [{"add": "new-label"}, {"remove": "old-label"}],
    "components": [{"add": {"name": "Frontend"}}]
  }
}
```

### Flag Design

New flags for issue create:

| Flag | Short | Type | Description |
|------|-------|------|-------------|
| --parent | | string | Parent issue or epic key |
| --component | -C | []string | Set component(s) |
| --fix-version | | []string | Set fix version(s) |

New flags for issue edit:

| Flag | Short | Type | Description |
|------|-------|------|-------------|
| --parent | | string | Parent issue key (empty/keyword removes) |
| --component | -C | []string | Replace all components |
| --add-component | | []string | Add component(s) |
| --remove-component | | []string | Remove component(s) |
| --fix-version | | []string | Replace all fix versions |
| --add-fix-version | | []string | Add fix version(s) |
| --remove-fix-version | | []string | Remove fix version(s) |
| --add-labels | | []string | Add label(s) |
| --remove-labels | | []string | Remove label(s) |

New flags for issue move:

| Flag | Short | Type | Description |
|------|-------|------|-------------|
| --comment | -m | string | Add comment during transition |
| --resolution | -R | string | Set resolution (e.g., Done, Won't Do) |
| --assignee | -a | string | Set assignee (email, accountId, me) |

New flags for issue delete:

| Flag | Short | Type | Description |
|------|-------|------|-------------|
| --cascade | | bool | Delete issue with all subtasks |

### Parent Removal

For edit, support multiple ways to remove a parent:

- `--parent` (no value, using NoOptDefVal)
- `--parent ""`
- `--parent none`
- `--parent remove`
- `--parent clear`
- `--parent unset`

Use `cmd.Flags().Changed("parent")` to detect explicit flag usage, then check for empty string or removal keywords.

### Conflicting Flags

Return an error when replace and add/remove flags are used together:

- Error if `--labels` used with `--add-labels` or `--remove-labels`
- Error if `--component` used with `--add-component` or `--remove-component`
- Error if `--fix-version` used with `--add-fix-version` or `--remove-fix-version`

### Validation

Let the Jira API validate field values. No upfront validation for:

- Component names
- Fix version names
- Resolution values

The API returns clear error messages when values are invalid.

## Why

API patterns:

- The `fields` vs `update` distinction matches how Jira's API works
- Replace operations use `fields` for simplicity
- Add/remove operations require `update` for atomic modifications

Flag design:

- `--parent` has no short flag to avoid conflict with `-P` (priority)
- `--fix-version` has no short flag for the same reason
- `-C` for component is memorable and available
- `-m` for comment matches git commit convention
- `-R` for resolution avoids conflict with `-r` (reporter in list)
- `-a` for assignee matches existing assign command pattern

Parent removal flexibility:

- Users have different mental models (none, remove, clear, unset)
- Empty value is intuitive but needs explicit flag detection
- Supporting multiple keywords costs nothing and helps usability

Error on conflicts:

- Combining replace with add/remove has ambiguous intent
- Explicit error is better than silent precedence rules
- Users can easily fix by removing one flag

API validation:

- Reduces code complexity
- Fewer API round trips
- Jira's error messages are clear enough
- Consistent with existing label handling

## Trade-offs

Accept:

- More flags to learn (9 new flags for edit alone)
- Users must understand replace vs add/remove distinction
- API error messages for invalid values (not custom messages)
- Parent removal keywords are magic strings

Gain:

- Complete control over Jira fields from CLI
- Atomic add/remove without fetching current values
- Flexible parent removal syntax
- Clear error on conflicting intent
- Simpler implementation without validation API calls

## Alternatives

Minus notation for removal:

- Syntax: `--labels foo,-bar` to add foo and remove bar
- Pro: Single flag for all operations
- Con: Parsing ambiguity with values starting with hyphen
- Con: Cobra flag parsing complications
- Con: Less discoverable
- Rejected: Explicit flags are clearer

Upfront validation:

- Make API calls to validate components/versions exist before create/edit
- Pro: Better error messages
- Con: Extra API calls on every operation
- Con: More code to maintain
- Rejected: API errors are clear enough, not worth the overhead

Silent precedence for conflicting flags:

- If both `--labels` and `--add-labels` provided, one wins
- Pro: No error, always works
- Con: User intent is unclear
- Con: Surprising behavior
- Rejected: Explicit error is better UX

## Usage Examples

Create with parent and components:

```bash
ajira issue create -s "Subtask" --parent GCP-50 --component Backend,API
```

Edit to add labels without removing existing:

```bash
ajira issue edit GCP-123 --add-labels urgent,reviewed
```

Remove parent from issue:

```bash
ajira issue edit GCP-123 --parent none
```

Transition with comment and resolution:

```bash
ajira issue move GCP-123 Done --comment "Completed implementation" --resolution Done
```

Delete with subtasks:

```bash
ajira issue delete GCP-50 --cascade
```
