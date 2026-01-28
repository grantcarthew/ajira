# ajira Agent Reference

Non-interactive Jira CLI. Text output is token-efficient; use --json only when parsing.

- JIRA_PROJECT env sets default project
- Text output returns issue URLs on success
- Use -f - to read description/comment content from stdin
- See `ajira help schemas` for JSON fields

## Markdown (not Jira wiki)

Wiki markup won't render. Use Markdown:

- `*bold*` → `**bold**`
- `h2. Title` → `## Title`
- `[text|url]` → `[text](url)`
- `{{code}}` → `` `code` ``
- `||H1||H2||` → `| H1 | H2 |`

## Core Commands

```
ajira issue list
ajira issue list -l 10 --status "In Progress" -t Bug -a me
ajira issue list -r me -P High -L bug,urgent -w
ajira issue list --order-by created --reverse
ajira issue list -q "status = Done AND updated >= -7d"
ajira issue view PROJ-123
ajira issue view PROJ-123 -c 10
ajira issue create -s "Fix login bug"
ajira issue create -s "Add feature" -t Story -d "Description"
ajira issue create -s "From file" -f description.md
echo "Description" | ajira issue create -s "From stdin" -f -
ajira issue create -s "Subtask" -t Sub-task --parent PROJ-50
ajira issue create -s "Full" --labels bug --priority High -C Backend --fix-version 1.0
ajira issue edit PROJ-123 -s "New summary" -d "New description"
ajira issue edit PROJ-123 --parent PROJ-50
ajira issue edit PROJ-123 --add-labels urgent --remove-labels stale
ajira issue edit PROJ-123 --add-component Frontend --add-fix-version 1.1.0
ajira issue assign PROJ-123 me
ajira issue assign PROJ-123 user@example.com
ajira issue assign PROJ-123 unassigned
ajira issue move PROJ-123 "In Progress"
ajira issue move PROJ-123 Done -m "Completed"
ajira issue move PROJ-123
ajira issue comment add PROJ-123 "Comment text"
ajira issue comment add PROJ-123 -f comment.md
ajira issue comment edit PROJ-123 12345 "Updated text"
ajira issue link types
ajira issue link add PROJ-123 Blocks PROJ-456
ajira issue link remove PROJ-123 PROJ-456
ajira issue link url PROJ-123 https://example.com "Docs"
ajira issue attachment list PROJ-123
ajira issue attachment add PROJ-123 file.pdf
ajira issue attachment add PROJ-123 *.log
ajira issue attachment download PROJ-123 10001
ajira issue attachment download PROJ-123 10001 -o custom.pdf
ajira issue attachment remove PROJ-123 10001
```

Note: `issue view` shows 5 comments by default. Comment IDs shown as `[date] [id] Author:`. Attachment IDs shown as `[id]`.

## Other Commands

See --help: `me`, `open`, `project list`, `board list`, `release list`, `user search`, `field list`, `issue clone`, `issue delete`, `issue watch`, `issue type`, `issue status`, `issue priority`, `epic create`, `epic list`, `epic add`, `epic remove`, `sprint list`, `sprint add`

## Chaining (JSON)

```
KEY=$(ajira issue create -s "New task" --json | jq -r .key)
ajira issue assign $KEY me
```

```
ajira issue list --status "To Do" --json | jq -r '.[].key' | while read key; do
  ajira issue move "$key" "In Progress"
done
```
