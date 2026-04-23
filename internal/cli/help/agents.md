# ajira Jira CLI

- Defaults: `JIRA_PROJECT` `JIRA_BOARD`
- `-f -` reads description/comment body from stdin
- See `ajira help {schemas,markdown,agile}`
- Use `--json` only when parsing

## Markdown (not Jira wiki)

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
ajira issue comment list PROJ-123
ajira issue comment list PROJ-123 -l 20
ajira issue link list PROJ-123
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

- `issue view` shows 5 comments by default
- `issue move` without target lists transitions
- Comment IDs shown as `[date] [id] Author:`
- Attachment IDs shown as `[id]`

## Other Commands

Run `<cmd> --help` for usage.

- Auth/lookup: `me`, `project list`, `release list`, `user search`, `field list`
- Issue meta: `issue clone`, `issue delete`, `issue watch`, `issue type`, `issue status`, `issue priority`
- Browser: `open`

