# Project Documents

This directory contains project documents. Each project represents a focused effort with clear goals, scope, and success criteria.

See [p-writing-guide.md](./p-writing-guide.md) for guidelines on creating and maintaining project documents.

---

## Quick Reference

| Project | Title | Status | Started | Completed |
|---------|-------|--------|---------|-----------|
| P-001 | Project Initialization | Completed | 2025-12-22 | 2025-12-22 |
| P-002 | CLI Core Infrastructure | Completed | 2025-12-22 | 2025-12-22 |
| P-003 | Markdown/ADF Conversion | Completed | 2025-12-24 | 2026-01-05 |
| P-004 | Issue Commands | Completed | 2026-01-05 | 2026-01-05 |
| P-005 | Comment Functionality | Completed | 2026-01-05 | 2026-01-05 |
| P-006 | Integration Testing | Completed | 2026-01-05 | 2026-01-07 |
| P-007 | Issue Linking | Completed | 2026-01-08 | 2026-01-08 |
| P-008 | Issue List Enhancements | Completed | 2026-01-09 | 2026-01-09 |
| P-009 | Issue Clone | Completed | 2026-01-12 | 2026-01-12 |
| P-010 | Agile Features | Completed | 2026-01-13 | 2026-01-13 |
| P-011 | Issue Command Enhancements | Completed | 2026-01-13 | 2026-01-13 |
| P-012 | Time Tracking | Proposed | | |
| P-013 | Automation Support | Proposed | | |
| P-014 | Auxiliary Commands | Proposed | | |
| P-015 | CLI Help System | Completed | 2026-01-07 | 2026-01-07 |

Note: Completed projects are in `completed/`

---

## Status Values

- **Proposed** - Project defined, not yet started
- **In Progress** - Currently being worked on
- **Completed** - All success criteria met, deliverables created (move to `completed/`)
- **Blocked** - Waiting on external dependency or decision

---

## Projects vs Design Records

**Projects** are work packages that define **what to build** and **how to validate it**.

**Design Records (DRs)** document **why we chose** a specific approach and the trade-offs.

A single project may generate multiple DRs. Projects describe the work; DRs document the decisions made during that work.

See [p-writing-guide.md](./p-writing-guide.md) for detailed guidance.

---

## Contributing

When creating a new project:

1. List directory to find next number: `ls docs/projects/p-*.md`
2. Use format: `p-<NNN>-<category>-<title>.md`
3. Follow the structure in [p-writing-guide.md](./p-writing-guide.md)
4. Define clear, measurable success criteria
5. Update this README with project entry
6. Link dependencies to other projects
