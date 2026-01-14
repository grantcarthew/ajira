# Code Review Task

This document guides AI agents through comprehensive code review using Documentation Driven Development principles.

The primary goal is to ensure code is correct, maintainable, efficient, and simple.

---

## Workflow

1. Read project documentation:
   - README.md
   - AGENTS.md
   - Active project document (from AGENTS.md)
   - All relevant Design Records (.ai/design/design-records/)

2. Read the code being reviewed:
   - Start with entry points (main files, index files)
   - Follow code paths through the system
   - Understand the structure and organization

3. Analyze code using the review topics below

4. Document findings in a rectification project:
   - Create new project document: `p-NNN-code-review-rectification.md`
   - List all issues found with severity (critical, high, medium, low)
   - Reference specific files and line numbers
   - Include enough context for future implementation
   - This becomes a standalone project for addressing issues

---

## Review Principles

General approach:

- Use Design Records as source of truth for architectural decisions
- If code differs from DRs, document the discrepancy
- If DRs appear incorrect, note this for discussion
- Focus on meaningful issues, not cosmetic preferences
- Prioritize correctness over optimization

---

## 1. Correctness and Functionality

Does the code do what it's supposed to do?

Requirements:

- Does code implement the feature or fix described in requirements?
- Are success criteria from project documents met?

Logic:

- Are there flaws in algorithms or logic?
- Are calculations and transformations correct?
- Do conditionals cover all cases?

Edge Cases:

- Null/nil/undefined inputs
- Empty collections (arrays, lists, maps)
- Zero values and boundary conditions
- Off-by-one errors in loops and ranges
- Invalid or malformed input data

Data Validation:

- Are inputs validated before use?
- Are assumptions about data documented and checked?

---

## 2. Design and Architecture

Does the code fit well into the larger system?

Design Records Alignment:

- Does implementation match architectural decisions in DRs?
- Are DR-documented trade-offs respected?
- Are design patterns from DRs followed consistently?

Simplicity:

- Is code unnecessarily complex?
- Can logic be simplified without losing clarity?
- Are abstractions appropriate for the problem?

Separation of Concerns:

- Are modules/classes/functions focused on single responsibilities?
- Is business logic separated from infrastructure concerns?
- Is the code organized logically?

Coupling and Cohesion:

- Is code loosely coupled to other modules?
- Are dependencies clear and minimal?
- Are related functions grouped together?

Abstraction Levels:

- Do functions operate at consistent abstraction levels?
- Are low-level details hidden behind clear interfaces?
- Is there appropriate layering?

---

## 3. Code Quality and Idioms

Is the code written in the idiomatic style of its language?

Language Standards:

- Does code follow language conventions and best practices?
- Are language-specific linters/formatters applied?
- Are standard library features used appropriately?

Naming:

- Are names clear, descriptive, and consistent?
- Do names follow language conventions (camelCase, snake_case, etc.)?
- Are abbreviations avoided unless well-known?

Code Organization:

- Is file/module structure logical and navigable?
- Are imports/dependencies organized?
- Is code grouped by functionality?

---

## 4. Error Handling

How does the code handle failures and unexpected conditions?

Error Checking:

- Are all error conditions handled?
- Are errors never silently ignored?
- Is error handling consistent across the codebase?

Error Context:

- Do error messages provide enough context?
- Are errors wrapped/annotated with relevant information?
- Can errors be traced to their source?

Error Recovery:

- Are recoverable errors handled gracefully?
- Are resources cleaned up on error paths?
- Is retry logic appropriate (if present)?

User Experience:

- Are error messages user-friendly (for user-facing errors)?
- Is logging appropriate for debugging?

---

## 5. Concurrency and Parallelism

If code uses concurrent/parallel execution:

Thread Safety:

- Is shared state properly protected?
- Are race conditions prevented?
- Are atomic operations used correctly?

Resource Management:

- Are concurrent tasks properly synchronized?
- Can deadlocks occur?
- Are resources properly released?

Async Patterns:

- Are promises/futures/async-await used correctly?
- Are callbacks properly managed?
- Is cancellation handled appropriately?

---

## 6. Testing

Is the code adequately tested?

Test Coverage:

- Are critical paths tested?
- Are edge cases covered?
- Are error conditions tested?

Test Quality:

- Are tests clear and maintainable?
- Do tests verify behavior, not implementation?
- Are test names descriptive?

Test Organization:

- Are unit tests separated from integration tests?
- Are test fixtures and mocks appropriate?
- Can tests run independently?

---

## 7. Performance and Resource Management

Is the code efficient and well-behaved?

Resource Cleanup:

- Are resources (files, connections, memory) properly released?
- Are cleanup handlers (defer, finally, using) used correctly?
- Are there potential resource leaks?

Efficiency:

- Are algorithms appropriately efficient?
- Are unnecessary allocations avoided in critical paths?
- Is I/O batched where appropriate?

Scalability:

- Will the code handle expected load?
- Are there obvious bottlenecks?
- Is caching used appropriately?

---

## 8. Documentation and Readability

Is the code understandable?

Code Comments:

- Are complex sections explained?
- Do comments explain "why" not "what"?
- Are comments up-to-date?

API Documentation:

- Are public APIs documented?
- Are parameters and return values described?
- Are examples provided for complex APIs?

DR References:

- Are DR references included where design decisions are implemented?
- Format: `// See dr-042 for authentication strategy`
- Are references accurate and up-to-date?

Code Clarity:

- Is the code self-documenting where possible?
- Are magic numbers/strings avoided or explained?
- Is the intent clear from reading the code?

---

## 9. Security

Are security considerations addressed?

Input Validation:

- Is user input sanitized and validated?
- Are injection attacks (SQL, XSS, etc.) prevented?
- Are file paths validated to prevent traversal attacks?

Authentication and Authorization:

- Are auth checks present where needed?
- Are permissions verified before sensitive operations?
- Are security decisions documented in DRs?

Sensitive Data:

- Are passwords/secrets properly handled?
- Is sensitive data encrypted when appropriate?
- Are secrets in environment variables, not code?

Dependencies:

- Are dependencies up-to-date?
- Are known vulnerabilities addressed?

---

## Creating the Rectification Project

After completing the review, create a new project document:

File: `.ai/projects/p-NNN-code-review-rectification.md`

Structure:

```markdown
# P-NNN: Code Review Rectification

- Status: Proposed

## Overview

Address issues found during code review of [component/feature].

## Issues Found

### Critical Issues

1. [Issue description] - File: path/to/file.ext:123
   - Problem: [what's wrong]
   - Impact: [why it matters]
   - Solution: [how to fix]

### High Priority Issues

[same structure]

### Medium Priority Issues

[same structure]

### Low Priority Issues

[same structure]

## Success Criteria

- [ ] All critical issues resolved
- [ ] All high priority issues resolved
- [ ] Code passes tests
- [ ] DR references added to implementation

## Deliverables

- Updated code addressing all critical/high issues
- Tests for fixed issues
- Updated DRs if design changes needed
```

This creates a clear, actionable project for addressing review findings.
