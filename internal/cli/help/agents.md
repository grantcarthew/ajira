# ajira Agent Reference

Non-interactive Jira CLI. Text output is token-efficient; use --json only when parsing.

See `ajira help schemas` for JSON output field lists.

## Key Behaviours

- JIRA_PROJECT env sets default project (no -p needed)
- Text output returns issue URLs on success
- Markdown for descriptions and comments (NOT Jira wiki markup)
- Use -f - to read from stdin

IMPORTANT: Use Markdown syntax, not Jira wiki markup.
Jira wiki (||, h2., {{}}) will NOT render correctly.
Run `ajira help markdown` for syntax reference.

## Find Issues

```
ajira issue list
ajira issue list -l 10
ajira issue list --status "In Progress"
ajira issue list -t Bug -a me
ajira issue list -r me -P High
ajira issue list -L bug,urgent
ajira issue list -w
ajira issue list --order-by created --reverse
ajira issue list -q "status = Done AND updated >= -7d"
```

## View Issue

```
ajira issue view PROJ-123
ajira issue view PROJ-123 -c 5
```

## Create Issue

```
ajira issue create -s "Fix login bug"
ajira issue create -s "Add feature" -t Story -d "Description here"
ajira issue create -s "From file" -f description.md
echo "Stdin description" | ajira issue create -s "From stdin" -f -
ajira issue create -s "With labels" --labels bug,urgent --priority High
ajira issue create -s "Subtask" -t Sub-task --parent PROJ-50
ajira issue create -s "With components" -C Backend,API
ajira issue create -s "With version" --fix-version 1.0.0
```

## Modify Issue

```
ajira issue edit PROJ-123 -s "New summary"
ajira issue edit PROJ-123 -d "New description"
ajira issue edit PROJ-123 --parent PROJ-50
ajira issue edit PROJ-123 --parent none
ajira issue edit PROJ-123 --add-labels urgent,reviewed
ajira issue edit PROJ-123 --remove-labels stale
ajira issue edit PROJ-123 --add-component Frontend
ajira issue edit PROJ-123 --add-fix-version 1.1.0
ajira issue assign PROJ-123 user@example.com
ajira issue assign PROJ-123 me
ajira issue assign PROJ-123 unassigned
ajira issue move PROJ-123 "In Progress"
ajira issue move PROJ-123 Done -m "Completed"
ajira issue move PROJ-123
ajira issue delete PROJ-123
ajira issue delete PROJ-123 --cascade
```

## Comments

```
ajira issue comment add PROJ-123 "Comment text"
ajira issue comment add PROJ-123 -f comment.md
echo "Stdin comment" | ajira issue comment add PROJ-123 -f -
ajira issue comment edit PROJ-123 12345 "Updated text"
ajira issue comment edit PROJ-123 12345 -f updated.md
```

Note: Comment IDs are shown in `issue view -c N` output as `[date] [id] Author:`.

## Issue Links

```
ajira issue link types
ajira issue link add PROJ-123 Blocks PROJ-456
ajira issue link remove PROJ-123 PROJ-456
ajira issue link url PROJ-123 https://example.com "Documentation"
```

## Available Values

```
ajira issue type
ajira issue status
ajira issue priority
```

## Chaining (JSON required)

```
KEY=$(ajira issue create -s "New task" --json | jq -r .key)
ajira issue assign $KEY me
```

```
ajira issue list --status "To Do" --json | jq -r '.[].key' | while read key; do
  ajira issue move "$key" "In Progress"
done
```
