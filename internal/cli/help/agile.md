# ajira Agile

Epic, sprint, and board commands. See `ajira help {agents,schemas}` for general rules and JSON fields.

- Sprint ops need `JIRA_BOARD` or `--board`
- `board list` to discover the id for `JIRA_BOARD`
- `sprint add` requires a future or active sprint (not closed)
- `epic remove` takes only issue keys (removes from current epic)

## Commands

```
ajira board list
ajira board list -l 10
ajira sprint list
ajira sprint list --current
ajira sprint list --state closed -l 5
ajira sprint add 42 PROJ-123 PROJ-124
echo -e "PROJ-1\nPROJ-2" | ajira sprint add 42 --stdin
ajira epic list
ajira epic list --status "In Progress"
ajira epic create -s "Auth Epic"
ajira epic create -s "API" -d "Description" -P Major -a me
ajira epic create -s "API" -f description.md
ajira epic add EPIC-1 PROJ-123 PROJ-124
ajira epic remove PROJ-123 PROJ-124
echo -e "PROJ-1\nPROJ-2" | ajira epic remove --stdin
```
