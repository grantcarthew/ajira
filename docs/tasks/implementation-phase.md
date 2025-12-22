# Implementation Phase Task

This document guides AI agents through implementation work using Documentation Driven Development (DDD).

IMPORTANT: All design decisions have been documented as Design Records (DRs) in `docs/design/design-records/`. Implementation must follow these decisions.

---

## Implementation Phase Objectives

The implementation phase translates design decisions into working code. You will:

1. Read all DRs to understand design decisions and rationale
2. Implement according to DRs - Follow the decisions that were made
3. Reference DRs in code - Link code to the decisions that drove it
4. Flag design gaps - Identify when implementation needs a decision that wasn't made
5. Deliver working software - Build the system as designed

## How to Work in Implementation Phase

### Your Role as Implementation Agent

You are here to:

- Implement the design - Write code that realizes the DRs
- Follow decisions - Don't second-guess design choices (they're in DRs)
- Ask when unclear - If a DR is ambiguous, ask for clarification
- Reference DRs - Add comments linking code to relevant DRs
- Identify gaps - If you need a decision that wasn't made, flag it
- Write tests - Follow the testing strategy from DRs
- Maintain quality - Follow code style and best practices from DRs

### Implementation Session Workflow

1. Read design context
   - Review all DRs before starting
   - Understand the complete system design
   - Identify key decisions that affect your current task

2. Plan implementation
   - Break down into implementable units
   - Identify dependencies
   - Determine implementation order

3. Implement
   - Write code following DR decisions
   - Add comments referencing relevant DRs (e.g., `// See DR-042 for rationale`)
   - Follow established patterns from DRs

4. Test
   - Write tests according to testing strategy DRs
   - Ensure tests cover decision points

5. Review against DRs
   - Verify implementation matches design decisions
   - Check that all DR requirements are met

6. Update project status
   - Update active project document with progress
   - Document any new DRs if implementation reveals needed decisions

### When Design Doesn't Cover Something

If you encounter a situation where:

- No DR exists for a needed decision
- Existing DR is ambiguous or incomplete
- Implementation reveals a design gap

Do this:

1. Stop implementation of that specific part
2. Document the question - What decision is needed?
3. Ask the user - Present the question and options
4. Create a DR - Document the decision (even in implementation phase)
5. Resume implementation - Continue with the new decision

Note: Small implementation details don't need DRs. Only create new DRs for decisions with alternatives, trade-offs, or lasting impact.

---

## Notes for AI Agents

### Reading Order

1. Active project document in docs/projects/ - Current implementation context
2. docs/design/design-records/README.md - Index of all design decisions
3. All DRs in docs/design/design-records/ - Complete design (critical!)
4. AGENTS.md - General agent guidance and project info
5. Existing codebase - What's already implemented

### Agent Instructions

When starting an implementation session:

- Read all DRs (or at minimum, the critical ones)
- Read the active project document for current context
- Ask user what to work on
- Check what's already implemented

During implementation:

- Keep DRs in mind - implement according to decisions
- Add comments linking code to DRs when implementing decision points

  ```python
  # Use PostgreSQL for persistence (see DR-023)
  # Chose connection pooling to handle concurrent requests (see DR-031)
  ```

- If you discover a missing decision, document it and ask
- Write clean, maintainable code following project guidelines
- Write tests as you go (don't leave for later)

After implementing:

- Update the active project document with progress
- If you created new DRs (for gaps), update DR index
- Ensure continuity for the next session

### Referencing DRs in Code

When implementing a decision point, add a comment:

Good examples:

```python
# Singleton pattern for database connection (DR-015)
class DatabaseConnection:
    _instance = None
```

```javascript
// Retry with exponential backoff for API calls (DR-028)
async function fetchWithRetry(url, maxRetries = 3) {
```

```go
// Use channels for worker pool communication (DR-041)
jobs := make(chan Job, 100)
```

When to add DR references:

- Architectural patterns
- Algorithm choices
- Error handling approaches
- Performance optimizations
- Security measures
- Any non-obvious choice that has a DR

### Quality Checks

Before marking a component complete:

- [ ] Implements all requirements from relevant DRs
- [ ] Follows code style guidelines from DRs
- [ ] Has tests (unit, integration as appropriate)
- [ ] Has documentation (code comments, docstrings)
- [ ] References relevant DRs in code comments
- [ ] Passes all tests
- [ ] Follows security guidelines from DRs (if applicable)

### If You Need to Deviate from a DR

Sometimes implementation reveals that a design decision needs revision. If this happens:

1. Don't just change it - The DR exists for a reason
2. Discuss with user - Explain why the decision doesn't work
3. Create new DR - Document the new decision
4. Update old DR - Set status to "Superseded" and link to new DR in header
5. Move old DR - Move the old DR file to docs/design/design-records/superseded/
6. Update DR index - Reflect the superseding relationship
7. Update code - Implement the new decision

---

## Resources

- DR Index: `docs/design/design-records/README.md`
- DR Writing Guide: `docs/design/dr-writing-guide.md`
- Project Writing Guide: `docs/projects/p-writing-guide.md`
- Agent Guide: `AGENTS.md`
