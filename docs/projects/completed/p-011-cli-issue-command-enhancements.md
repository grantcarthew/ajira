# P-011: Issue Command Enhancements

- Status: Complete
- Started: 2026-01-13
- Completed: 2026-01-13

## Overview

Enhance existing issue commands (create, edit, move, delete) with additional field support. This project adds parent/epic assignment, components, fix versions, transition enhancements, and cascade delete functionality.

Custom field support is explicitly excluded from this project due to complexity and moved to a future project.

## Goals

1. Add parent/epic field support to create and edit
2. Add component and fix version support
3. Enhance move command with comment, resolution, and assignee
4. Add cascade delete for subtasks
5. Support label add/remove syntax in edit

## Scope

### In Scope

Create enhancements (`issue create`):

- `--parent` - Attach to parent issue or epic (no short flag)
- `--component`, `-C` - Set component(s)
- `--fix-version` - Set fix version(s) (no short flag)

Edit enhancements (`issue edit`):

- `--parent` - Change parent/epic (no short flag)
- `--component`, `-C` - Set/add/remove components
- `--fix-version` - Set/add/remove fix versions (no short flag)
- `--add-labels` - Add labels without replacing existing
- `--remove-labels` - Remove specific labels

Move enhancements (`issue move`):

- `--comment`, `-m` - Add comment during transition
- `--resolution`, `-R` - Set resolution (for done transitions)
- `--assignee`, `-a` - Set assignee during transition

Delete enhancements (`issue delete`):

- `--cascade` - Delete issue with all subtasks

### Out of Scope

- Custom field support (complexity warrants separate project)
- Attachment management (separate project)
- Bulk operations (covered in P-013)
- Minus notation for labels (replaced with explicit `--add-labels` / `--remove-labels`)

## Implementation Reference

### Existing Code Patterns

Flag definitions use Cobra in `init()` functions:

```go
// Pattern from issue_create.go
func init() {
    issueCreateCmd.Flags().StringVarP(&createSummary, "summary", "s", "", "Issue summary (required)")
    issueCreateCmd.Flags().StringSliceVar(&createLabels, "labels", nil, "Issue labels (comma-separated)")
    issueCmd.AddCommand(issueCreateCmd)
}
```

Validation before API calls:

```go
// Pattern from issue_create.go
if err := jira.ValidateIssueType(ctx, client, projectKey, createType); err != nil {
    return fmt.Errorf("%v", err)
}
```

### Flag Conflict Resolution

The `-P` short flag is already assigned to `--priority` in create, edit, and clone commands:

```go
// issue_create.go:76
issueCreateCmd.Flags().StringVarP(&createPriority, "priority", "P", "", "Issue priority")
```

Resolution: `--parent` and `--fix-version` have no short flags to avoid conflicts.

### API Patterns

Issue create uses typed struct (`issue_create.go:25-36`):

```go
type issueCreateFields struct {
    Project     projectKey     `json:"project"`
    Summary     string         `json:"summary"`
    Description *converter.ADF `json:"description,omitempty"`
    IssueType   issueTypeName  `json:"issuetype"`
    Priority    *priorityName  `json:"priority,omitempty"`
    Labels      []string       `json:"labels,omitempty"`
}
```

Issue edit uses dynamic map (`issue_edit.go:18-20`):

```go
type issueEditRequest struct {
    Fields map[string]any `json:"fields"`
}
```

Transition uses simple struct (`issue_move.go:29-36`):

```go
type transitionRequest struct {
    Transition transitionRef `json:"transition"`
}
```

### API Client Methods

Located in `internal/api/client.go`:

- `client.Get(ctx, path)` - GET request
- `client.Post(ctx, path, body)` - POST request
- `client.Put(ctx, path, body)` - PUT request
- `client.Delete(ctx, path)` - DELETE request (needs extension for query params)

## Technical Specifications

### 1. Parent Field

The `parent` field sets an issue's parent (epic or parent issue for subtasks).

API payload for create:

```json
{
  "fields": {
    "project": {"key": "GCP"},
    "summary": "Child issue",
    "issuetype": {"name": "Task"},
    "parent": {"key": "GCP-50"}
  }
}
```

API payload for edit:

```json
{
  "fields": {
    "parent": {"key": "GCP-50"}
  }
}
```

To remove parent, set to null:

```json
{
  "fields": {
    "parent": null
  }
}
```

Implementation notes:

- Works for both epics and subtask parents
- Accepts issue key format (e.g., GCP-50)
- No validation needed; API returns clear error if parent invalid

New flag:

| Flag | Short | Type | Description |
|------|-------|------|-------------|
| --parent | | string | Parent issue or epic key |

### 2. Components

Components are project-specific categories for issues.

API payload for create:

```json
{
  "fields": {
    "components": [{"name": "Backend"}, {"name": "API"}]
  }
}
```

API payload for edit (replace all):

```json
{
  "fields": {
    "components": [{"name": "Backend"}, {"name": "API"}]
  }
}
```

API payload for edit (add/remove):

```json
{
  "update": {
    "components": [
      {"add": {"name": "NewComponent"}},
      {"remove": {"name": "OldComponent"}}
    ]
  }
}
```

Implementation notes:

- Create always sets components (replaces any default)
- Edit with `--component` replaces all components
- Edit with `--add-component` adds without replacing
- Edit with `--remove-component` removes specific components
- Multiple values comma-separated: `--component Backend,API`

New flags for create:

| Flag | Short | Type | Description |
|------|-------|------|-------------|
| --component | -C | []string | Set component(s) |

New flags for edit:

| Flag | Short | Type | Description |
|------|-------|------|-------------|
| --component | -C | []string | Replace all components |
| --add-component | | []string | Add component(s) |
| --remove-component | | []string | Remove component(s) |

### 3. Fix Versions

Fix versions track which release will include an issue.

API payload for create:

```json
{
  "fields": {
    "fixVersions": [{"name": "1.0.0"}, {"name": "1.1.0"}]
  }
}
```

API payload for edit (replace all):

```json
{
  "fields": {
    "fixVersions": [{"name": "1.0.0"}]
  }
}
```

API payload for edit (add/remove):

```json
{
  "update": {
    "fixVersions": [
      {"add": {"name": "1.1.0"}},
      {"remove": {"name": "1.0.0"}}
    ]
  }
}
```

Implementation notes:

- Same pattern as components
- Version names must exist in project
- Multiple values comma-separated: `--fix-version 1.0.0,1.1.0`

New flags for create:

| Flag | Short | Type | Description |
|------|-------|------|-------------|
| --fix-version | | []string | Set fix version(s) |

New flags for edit:

| Flag | Short | Type | Description |
|------|-------|------|-------------|
| --fix-version | | []string | Replace all fix versions |
| --add-fix-version | | []string | Add fix version(s) |
| --remove-fix-version | | []string | Remove fix version(s) |

### 4. Label Add/Remove for Edit

Currently `--labels` replaces all labels. Add explicit add/remove flags.

API payload for add/remove:

```json
{
  "update": {
    "labels": [
      {"add": "new-label"},
      {"remove": "old-label"}
    ]
  }
}
```

Note: Labels use string values directly, not objects like components.

New flags for edit:

| Flag | Short | Type | Description |
|------|-------|------|-------------|
| --add-labels | | []string | Add label(s) |
| --remove-labels | | []string | Remove label(s) |

Existing `--labels` flag behaviour unchanged (replaces all).

### 5. Transition Enhancements

Enhance the move command to support comment, resolution, and assignee during transition.

Current request structure:

```go
type transitionRequest struct {
    Transition transitionRef `json:"transition"`
}
```

Enhanced request structure:

```go
type transitionRequest struct {
    Transition transitionRef  `json:"transition"`
    Fields     map[string]any `json:"fields,omitempty"`
    Update     map[string]any `json:"update,omitempty"`
}
```

API payload with comment:

```json
{
  "transition": {"id": "31"},
  "update": {
    "comment": [
      {
        "add": {
          "body": {
            "version": 1,
            "type": "doc",
            "content": [{"type": "paragraph", "content": [{"type": "text", "text": "Comment text"}]}]
          }
        }
      }
    ]
  }
}
```

API payload with resolution:

```json
{
  "transition": {"id": "31"},
  "fields": {
    "resolution": {"name": "Done"}
  }
}
```

API payload with assignee:

```json
{
  "transition": {"id": "31"},
  "fields": {
    "assignee": {"accountId": "abc123"}
  }
}
```

Implementation notes:

- Comment uses ADF format (use existing `converter.MarkdownToADF`)
- Resolution only valid for transitions to "Done" category
- Assignee uses accountId (use existing `resolveUser` function)
- All three can be combined in single request

New flags:

| Flag | Short | Type | Description |
|------|-------|------|-------------|
| --comment | -m | string | Add comment during transition |
| --resolution | -R | string | Set resolution (e.g., Done, Won't Do) |
| --assignee | -a | string | Set assignee (email, accountId, 'me') |

### 6. Cascade Delete

Delete an issue and all its subtasks.

Current delete implementation (`issue_delete.go:57-60`):

```go
func deleteIssue(ctx context.Context, client *api.Client, key string) error {
    path := fmt.Sprintf("/issue/%s", key)
    _, err := client.Delete(ctx, path)
    return err
}
```

API endpoint with cascade:

```
DELETE /rest/api/3/issue/{issueIdOrKey}?deleteSubtasks=true
```

Implementation approach:

Option A: Add query parameter support to delete function:

```go
func deleteIssue(ctx context.Context, client *api.Client, key string, cascade bool) error {
    path := fmt.Sprintf("/issue/%s", key)
    if cascade {
        path += "?deleteSubtasks=true"
    }
    _, err := client.Delete(ctx, path)
    return err
}
```

Option B: Add `DeleteWithQuery` method to API client.

Recommendation: Option A is simpler and sufficient for this use case.

New flag:

| Flag | Short | Type | Description |
|------|-------|------|-------------|
| --cascade | | bool | Delete issue with all subtasks |

## Implementation Order

Implement in this order to build incrementally:

1. **Cascade delete** - Simplest change, modifies one function
2. **Transition enhancements** - Self-contained in issue_move.go
3. **Parent field** - Simple addition to create and edit
4. **Label add/remove** - Introduces `update` field pattern for edit
5. **Components** - Uses same `update` pattern
6. **Fix versions** - Uses same `update` pattern

## File Changes

### issue_create.go

Add fields to `issueCreateFields` struct:

```go
type issueCreateFields struct {
    Project     projectKey     `json:"project"`
    Summary     string         `json:"summary"`
    Description *converter.ADF `json:"description,omitempty"`
    IssueType   issueTypeName  `json:"issuetype"`
    Priority    *priorityName  `json:"priority,omitempty"`
    Labels      []string       `json:"labels,omitempty"`
    Parent      *parentKey     `json:"parent,omitempty"`      // NEW
    Components  []componentName `json:"components,omitempty"` // NEW
    FixVersions []versionName  `json:"fixVersions,omitempty"` // NEW
}

type parentKey struct {
    Key string `json:"key"`
}

type componentName struct {
    Name string `json:"name"`
}

type versionName struct {
    Name string `json:"name"`
}
```

Add new flag variables and flag definitions in `init()`.

Update `createIssue` function to populate new fields.

### issue_edit.go

Change request structure to support both `fields` and `update`:

```go
type issueEditRequest struct {
    Fields map[string]any `json:"fields,omitempty"`
    Update map[string]any `json:"update,omitempty"`
}
```

Add new flag variables and flag definitions in `init()`.

Update `runIssueEdit` to:

- Use `fields` for replace operations (parent, component, fix-version, labels)
- Use `update` for add/remove operations (add-component, remove-component, etc.)

### issue_move.go

Update request structure:

```go
type transitionRequest struct {
    Transition transitionRef  `json:"transition"`
    Fields     map[string]any `json:"fields,omitempty"`
    Update     map[string]any `json:"update,omitempty"`
}
```

Add new flag variables and flag definitions in `init()`.

Update `doTransition` to accept optional fields and update maps.

Import `converter` package for comment ADF conversion.

### issue_delete.go

Add `--cascade` flag variable.

Update `deleteIssue` function signature and implementation.

## Test Requirements

### Unit Tests

Add to `issue_test.go`:

- `TestCreateIssue_WithParent` - Verify parent field in request
- `TestCreateIssue_WithComponents` - Verify components array in request
- `TestCreateIssue_WithFixVersions` - Verify fixVersions array in request
- `TestUpdateIssue_WithParent` - Verify parent field update
- `TestUpdateIssue_AddComponents` - Verify update.components add operation
- `TestUpdateIssue_RemoveComponents` - Verify update.components remove operation
- `TestUpdateIssue_AddLabels` - Verify update.labels add operation
- `TestUpdateIssue_RemoveLabels` - Verify update.labels remove operation
- `TestDoTransition_WithComment` - Verify comment in update field
- `TestDoTransition_WithResolution` - Verify resolution in fields
- `TestDoTransition_WithAssignee` - Verify assignee in fields
- `TestDeleteIssue_Cascade` - Verify deleteSubtasks query parameter

### Integration Tests

Manual testing against Jira instance:

- Create issue with parent, verify parent link in Jira
- Edit issue to add/remove components, verify in Jira
- Move issue with comment, verify comment appears
- Delete parent issue with cascade, verify subtasks deleted

## Success Criteria

- [x] `issue create --parent GCP-50` attaches issue to epic
- [x] `issue create --component Backend,API` sets components
- [x] `issue create --fix-version 1.0.0` sets fix version
- [x] `issue edit --parent GCP-50` changes parent
- [x] `issue edit --add-component Frontend` adds component
- [x] `issue edit --remove-component Backend` removes component
- [x] `issue edit --add-fix-version 1.1.0` adds fix version
- [x] `issue edit --remove-fix-version 1.0.0` removes fix version
- [x] `issue edit --add-labels urgent` adds label
- [x] `issue edit --remove-labels stale` removes label
- [x] `issue move PROJ-123 Done --comment "Completed"` adds comment during transition
- [x] `issue move PROJ-123 Done --resolution "Done"` sets resolution
- [x] `issue move PROJ-123 "In Progress" --assignee me` sets assignee during transition
- [x] `issue delete PROJ-123 --cascade` deletes issue and subtasks
- [x] All new flags have help text
- [x] Unit tests pass for all new functionality

## Deliverables

- Updated `internal/cli/issue_create.go` - Parent, components, fix versions
- Updated `internal/cli/issue_edit.go` - All new fields with add/remove support
- Updated `internal/cli/issue_move.go` - Comment, resolution, assignee
- Updated `internal/cli/issue_delete.go` - Cascade option
- Updated `internal/cli/issue_test.go` - Unit tests for all changes
- DR-012: Issue Field Update Patterns (documents `fields` vs `update` usage)

## Dependencies

- P-010 (Agile Features) - Complete (provides epic context)

## Notes

### Why No Custom Field Support

Custom fields require:

1. Field discovery via `/rest/api/3/field` API
2. Mapping field names to IDs (e.g., "Story Points" -> "customfield_10016")
3. Different value formats per field type (string, number, option, etc.)
4. Project-specific field availability

This complexity warrants a dedicated project with proper design.

### Why Explicit Add/Remove Flags Instead of Minus Notation

The minus notation (`--labels -old-label`) has parsing issues:

- Ambiguous with flag values starting with hyphen
- Cobra flag parsing complications
- Less discoverable for users

Explicit flags (`--add-labels`, `--remove-labels`) are clearer and consistent with Jira's API model.
