# Design Phase Task

This document guides AI agents through design work using Documentation Driven Development (DDD).

IMPORTANT: Read `docs/design/dr-writing-guide.md` before continuing. That guide defines how to write Design Records (DRs).

---

## Design Phase Objectives

The design phase focuses on making and documenting key decisions. Through interactive creation of Design Records, you will:

1. Think deeply about the project by answering design questions
2. Make explicit decisions about architecture, technology, and approach
3. Document the "why" behind each decision
4. Identify dependencies and decision order
5. Build design clarity before (or alongside) implementation

## How to Work in Design Phase

### Your Role as Design Agent

You are here to:

- Ask probing questions that uncover design decisions
- Challenge assumptions to ensure decisions are well-thought-out
- Suggest alternatives the user may not have considered
- Create DRs for each significant decision
- Maintain the DR index as decisions are made
- Identify gaps in the design that need decisions

### Design Session Workflow

1. Identify next design decision - What needs to be decided?
2. Gather context - What information is needed to make this decision?
3. Explore alternatives - What are the options?
4. Discuss trade-offs - What are the pros/cons of each option?
5. Make decision - User chooses an approach
6. Document as DR - Create a Design Record following dr-writing-guide.md
7. Update DR index - Add to docs/design/design-records/README.md
8. Identify follow-up decisions - What new questions does this decision raise?
9. Repeat - Continue until design is complete (or ready to proceed)

When a decision supersedes an earlier one:

- Create the new DR documenting the new decision
- Update the old DR's status to "Superseded" and add link to new DR in header
- Move the old DR file to docs/design/design-records/superseded/
- Update the DR index (docs/design/design-records/README.md)

### When to Create a Design Record

Create a DR when the decision:

- Affects architecture - System structure, component relationships
- Has alternatives - Multiple valid approaches were considered
- Involves trade-offs - Choosing one thing means giving up another
- Has long-term impact - Will be hard to change later
- Needs explanation - Future developers will ask "why did we do it this way?"

Don't create DRs for:

- Trivial implementation details (variable names, file organization)
- Decisions with only one reasonable option
- Temporary/experimental code
- Standard practices (following language conventions)

### Question Types to Ask

Architecture Questions:

- How should the system be structured?
- What are the major components?
- How do components communicate?
- What are the boundaries?

Technology Questions:

- Which framework/library should we use?
- What database/storage approach?
- What build/deployment tools?
- What testing strategy?

Design Questions:

- What are the main entities/models?
- What are the key workflows?
- How should errors be handled?
- What are the extension points?

Process Questions:

- What's the development workflow?
- How do we ensure quality?
- What's the release process?
- How do we handle versioning?

Security Questions:

- How will authentication work?
- What authorization model should we use?
- How do we protect sensitive data?
- How are secrets/credentials managed?
- What are the security requirements?

Performance/Scalability Questions:

- What are the performance requirements?
- How will this scale?
- What are the known/potential bottlenecks?
- What caching strategy should we use?
- How do we handle high load?

Data Questions:

- What's the data lifecycle?
- How do we handle data privacy?
- What's the backup/recovery strategy?
- How do we handle data migrations?
- What's the data retention policy?

Integration Questions:

- What external services do we integrate with?
- How do we handle third-party APIs?
- What happens when dependencies are unavailable?
- How do we manage API versioning?
- What are the integration failure modes?

---

## Notes for AI Agents

### Reading Order

1. Active project document in docs/projects/ - Current context
2. docs/design/dr-writing-guide.md - How to write DRs
3. docs/design/design-records/README.md - DR index
4. All DRs in docs/design/design-records/ - Design decisions made so far
5. AGENTS.md - General agent guidance and project info

### Agent Instructions

When starting a design session:

- Read the active project document (referenced in AGENTS.md)
- Read all existing DRs to understand current state
- Ask user which area they want to focus on
- If no direction given, suggest tackling high-priority design questions

During design discussion:

- Ask clarifying questions
- Suggest alternatives user may not have considered
- Play devil's advocate to ensure decisions are robust
- Point out implications and consequences
- Help user think through edge cases

After each decision:

- Create DR immediately while discussion is fresh
- Update DR index (docs/design/design-records/README.md)
- Identify new questions that emerged from the decision

### DR Numbering

DRs are numbered sequentially starting from 001:

- `dr-001-first-decision.md`
- `dr-002-second-decision.md`
- etc.

Always check `docs/design/design-records/README.md` to find the next number.

### Quality Checks

Before creating a DR, ensure:

- [ ] Decision is significant (not trivial)
- [ ] Alternatives were considered
- [ ] Trade-offs are understood
- [ ] Context explains the "why"
- [ ] Consequences (both positive and negative) are documented

---

## Resources

- DR Writing Guide: `docs/design/dr-writing-guide.md`
- DR Index: `docs/design/design-records/README.md`
- Project Writing Guide: `docs/projects/p-writing-guide.md`
- Agent Guide: `AGENTS.md`
