# Project Writing Guide

Creating and maintaining Project documents.

Location: `docs/projects/p-NNN-title.md`

Read when: Writing/updating project documents or planning work.

---

## Markdown Formatting Note

IMPORTANT: Avoid using bold Markdown (`**text**`) in project documents. Bold formatting adds no semantic value to large language models but significantly increases token count. Use section headers, lists, and clear structure instead.

Good: Use `## Section Name` and bullet points
Bad: Use `**Section Name:**` with bold

Note: CLI documentation (docs/cli/) is intended for human readers and should use bold for readability. This guidance applies only to DR and Project documents that are primarily consumed by AI agents.

---

## Project Document Schema

Project structure:

### Header

```markdown
# P-NNN: Title

- Status: Proposed | In Progress | Completed | Blocked
- Started: YYYY-MM-DD (when work begins)
- Completed: YYYY-MM-DD (when finished)
```

### Required Sections

**Overview:**

- Brief description of what this project accomplishes
- Why this project exists
- How it fits into the larger system

**Goals:**

- Specific, measurable outcomes
- What we're trying to achieve
- 3-7 concrete goals typically

**Scope:**

- What's included in this project
- What's explicitly excluded
- Boundaries and constraints

**Success Criteria:**

- How we know the project is complete
- Measurable, testable outcomes
- Must be unambiguous

**Deliverables:**

- Concrete outputs (files, documents, code, DRs)
- What gets created or updated
- What gets published or deployed

### Optional Sections

Add as needed:

- **Dependencies** - Other projects or resources needed before starting
- **Technical Approach** - High-level strategy for accomplishing goals
- **Questions & Uncertainties** - What we need to figure out during the project
- **Research Areas** - Topics to investigate
- **Design Decisions** - DRs that will likely emerge from this project
- **Testing Strategy** - How we'll validate the deliverables
- **Timeline Estimates** - Rough effort estimates (optional, not commitments)
- **Notes** - Additional context, learnings, observations
- **Updates** - Historical changes with dates

---

## Projects vs Design Records

### Projects Are

- Work packages with clear deliverables
- Time-bounded efforts
- Focus on **what to build** and **how to validate it**
- May generate multiple DRs
- Can span research, design, and implementation
- Answer: "What are we doing?"

### Design Records Are

- Architectural decisions
- Focus on **why we chose** a specific approach
- Document trade-offs and alternatives
- Created **during** project execution
- Permanent record of reasoning
- Answer: "Why did we do it this way?"

### Relationship

A single project may create multiple DRs. For example:

- **P-001: CUE Foundation & Architecture** might generate:
  - DR-001: CUE Configuration Format Decision
  - DR-002: Schema Design Patterns
  - DR-003: Module Organization Strategy

Projects describe the work to be done. DRs document the decisions made while doing that work.

---

## What Belongs in Project Documents

### ✅ Goals and Objectives

Clear statements of what we're trying to achieve:

**Goals:**

1. Understand CUE's type system and validation capabilities
2. Design CUE schemas for roles, tasks, contexts, and agents
3. Document how CUE modules and packages work
4. Create architecture foundation for CUE-based system

### ✅ Success Criteria

Measurable outcomes that define completion:

**Success Criteria:**

- [ ] Can explain CUE's type system and how it applies to our use case
- [ ] Have working CUE schemas that validate correctly
- [ ] Documented module hierarchy and organization
- [ ] Created DR-001 documenting architectural decisions

### ✅ Scope Boundaries

Clear in/out of scope definitions:

**Scope:**

In Scope:

- Research CUE language features relevant to configuration
- Design schemas for core concepts
- Test CUE validation and composition

Out of Scope:

- Go integration (covered in P-005)
- CLI implementation (covered in P-004)
- Registry publishing (covered in P-003)

### ✅ Deliverables

Concrete outputs:

**Deliverables:**

- `docs/cue/schema-design.md` - Schema design documentation
- `examples/role.cue` - Example role schema
- `examples/task.cue` - Example task schema
- DR-001: CUE Configuration Format Decision
- DR-002: Schema Design Patterns

### ✅ Questions & Uncertainties

What we need to figure out:

**Questions & Uncertainties:**

- How do we handle schema inheritance vs composition?
- Can CUE validate command template syntax?
- What's the best way to structure module hierarchy?
- How do defaults and constraints interact?

### ✅ Research Areas

Topics to investigate:

**Research Areas:**

- CUE type system and unification
- Module and package structure
- CUE Central Registry package format
- Go API for loading and validating CUE

---

## What Does NOT Belong in Projects

### ❌ Implementation Code

Do not include actual source code - that goes in the repository:

```go
// Bad: Implementation code in project doc
func LoadConfig(path string) (*Config, error) {
    // ... implementation ...
}
```

Exception: Pseudocode in Technical Approach for clarity.

### ❌ Detailed Design Decisions

Design decisions belong in DRs, not project docs:

Bad: Detailed trade-off analysis in project doc

Good: "This project will generate DR-001 to document the CUE format decision"

### ❌ Step-by-Step Instructions

Projects define **what** and **why**, not detailed **how**:

Bad: "First run this command, then edit this file, then..."

Good: "Implement CUE loading from Go using the official API"

---

## Project Statuses

### Proposed

- Project defined but not started
- Goals and scope documented
- Waiting to begin work

### In Progress

- Currently being worked on
- Update with notes and learnings as you go
- Track completed goals/criteria

### Completed

- All success criteria met
- All deliverables created
- Set completion date
- Move to `completed/`

### Blocked

- Cannot proceed due to external dependency
- Document what's blocking progress
- Link to dependency or issue

---

## Project Numbering and Naming

### File Naming Convention

Format: `p-<NNN>-<category>-<title>.md`

- `NNN` = Three-digit number with leading zeros (001, 002, 003...)
- `category` = Technology/component/area (cue, cli, assets, distribution, orchestration, etc.)
- `title` = KISS description of the project
- All lowercase kebab-case (words separated by hyphens, no underscores or spaces)

Examples:

- `p-001-cue-foundation-architecture.md`
- `p-002-assets-concrete-tasks.md`
- `p-003-distribution-strategy.md`
- `p-004-cli-minimal-viable.md`

### Numbering

- Sequential: P-001, P-002, P-003, etc.
- Get next number from `docs/projects/README.md` index
- Never reuse numbers
- Gaps are acceptable (deferred/cancelled projects)

---

## Writing a Good Project Document

### Keep Projects Focused

Good project scope:

- Clear boundaries
- Achievable in reasonable timeframe
- Single area of concern
- Generates clear deliverables

Too broad:

- "Build the entire CLI" → Split into multiple projects

Too narrow:

- "Write one function" → Probably just a task, not a project

### Make Success Criteria Testable

Good criteria:

- [ ] Can load and validate CUE schemas from Go
- [ ] Example configs pass validation
- [ ] Created DR documenting format decision

Bad criteria:

- "Understand CUE better" (not measurable)
- "Make good schemas" (subjective)

### Document Uncertainties

Projects often start with unknowns - that's okay!

**Questions & Uncertainties:**

- How does CUE handle circular dependencies?
- What's the performance impact of validation?
- Can we dynamically load schemas?

These get answered **during** the project.

### Link to Related Work

**Dependencies:**

- Requires: P-001 (architecture foundation)
- Blocks: P-005 (needs schemas from this project)

**Related DRs:**

- Will create: DR-003 (package structure)
- Builds on: DR-001 (CUE format decision)

---

## Project Lifecycle

1. **Create** - Define goals, scope, success criteria
2. **Review** - Ensure project is well-scoped and achievable
3. **Start** - Set status to "In Progress", add start date
4. **Execute** - Work on deliverables, create DRs, document learnings
5. **Update** - Add notes, learnings, questions as you go
6. **Complete** - Verify success criteria, set completion date
7. **Archive** - Move to completed status, update README index

---

## Examples

### Good Project Document

```markdown
# P-001: CUE Foundation & Architecture

- Status: In Progress
- Started: 2025-12-01

## Overview

Research CUE capabilities and design the foundational architecture for the
CUE-based start tool. This project establishes how we'll use CUE for
configuration, validation, and asset management.

## Goals

1. Understand CUE's type system and validation
2. Design schemas for core concepts
3. Document module organization strategy

## Success Criteria

- [ ] Can explain CUE type system and how it applies
- [ ] Have working schemas that validate
- [ ] Created DR-001 documenting decisions

## Deliverables

- `docs/cue/schema-design.md`
- `examples/role.cue`
- DR-001: CUE Configuration Format
```

### Bad Project Document

```markdown
# P-001: Learn CUE

Make CUE work.

## TODO

- Read docs
- Write code
- Test stuff
```

Problems:

- Vague goals
- No success criteria
- No clear deliverables
- Not measurable

---

## Contributing

When creating a new project:

1. Get next number from `docs/projects/README.md`
2. Use template structure above
3. Define clear, measurable success criteria
4. Update README index
5. Link dependencies and related DRs
