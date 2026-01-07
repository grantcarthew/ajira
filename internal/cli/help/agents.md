# ajira Agent Reference

Non-interactive Jira CLI. Text output is token-efficient; use --json only when parsing.

See `ajira help schemas` for JSON output field lists.

## Key Behaviours

- JIRA_PROJECT env sets default project (no -p needed)
- Text output returns issue URLs on success
- Markdown for descriptions and comments
- Use -f - to read from stdin

## Find Issues

```
ajira issue list
ajira issue list -l 10
ajira issue list -s "In Progress"
ajira issue list -t Bug -a me
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
```

## Modify Issue

```
ajira issue edit PROJ-123 -s "New summary"
ajira issue edit PROJ-123 -d "New description"
ajira issue assign PROJ-123 user@example.com
ajira issue assign PROJ-123 me
ajira issue assign PROJ-123 unassigned
ajira issue move PROJ-123 "In Progress"
ajira issue move PROJ-123
ajira issue delete PROJ-123
```

## Comments

```
ajira issue comment add PROJ-123 "Comment text"
ajira issue comment add PROJ-123 -f comment.md
echo "Stdin comment" | ajira issue comment add PROJ-123 -f -
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
ajira issue list -s "To Do" --json | jq -r '.[].key' | while read key; do
  ajira issue move "$key" "In Progress"
done
```
