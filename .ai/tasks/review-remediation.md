# Review Remediation

Remediate a single open code review.

See `.ai/reviews/README.md` for review status and links.

## Workflow

1. Read the review document
2. Analyse referenced code/files
3. Triage findings with user: fix or dismiss
4. Implement fixes, update/add tests as needed
5. Run `go test ./...` and `gl`
6. If fix requires significant refactoring, consult user
7. If fix involves architectural decision, consult user about DR
8. Update `.ai/reviews/README.md` status to Actioned

## Completion

Task is complete when the review has been processed and marked Actioned.
