# AGENTS.md

## Non-Negotiables

- **Strict TDD is mandatory**: Write failing test first (test-as-documentation, one-at-a-time, regression-proof) -> minimal code to pass -> refactor -> using linters & formatters.
- **Adversarial Cooperation**: Rigorously check against linters and hostile unit tests or security exploits. If complexity requires, ultilize parallel Tasks, Concensus Voting, Synthetic and Fuzzy Test Case Generation with high-quality examples and high volume variations.
- **Common Pitfalls**:
  - **CGO_ENABLED=1**: Always prefix Go commands with this (SQLite requires CGO).
  - **Never edit `gen/` directories**: Run `go generate` to regenerate from OpenAPI specs.
  - **Commits**: No watermarks. No `Co-Authored-By` lines.
- **Only trust independent verification**: Never claim "done" without test output and command evidence.

## Core Workflow

### Requirements Contract (Non-Trivial Tasks)

Before coding, define: Goal, Acceptance Criteria (testable), Non-goals, Constraints, Verification Plan. **Rule**: If you cannot write acceptance criteria, pause and clarify.

Use Repomix MCP to explore code/structure and Context7 MCP to acquire up-to-date documentations and Playwright MCP for interactive browser-based testing. Use Web Search/Fetch if you see fit.
  -Context7: “Use context7 to acquire up-to-date, version-specific documentation for any library or API you touch.”
  -Repomix: “Use repomix to explore and pack the repository when you need a full-structure view or large-context reasoning.”
  -Playwright: “Use playwright MCP for interactive browser-based E2E tests and UI debugging instead of only static reasoning.”

Use GitHub CLI (`gh`) for GitHub related operations. And any doc command to stay correct, e.g. `go doc ...` if applicable.

### Verification Minimum

```bash
golangci-lint fmt && golangci-lint run && CGO_ENABLED=1 go test ./... -v -race -cover  # All formatters, linters, and tests with race detection. Or Powershell equivalent if running on Windows
```

### When Stuck (3 Failed Attempts)

1. Stop coding. Return to last green state.
2. Re-read requirements. Verify solving the RIGHT problem.
3. Decompose into atomic subtasks (<10 lines each).
4. Spawn 3 parallel diagnostic tasks via Task tool.
5. If still blocked → escalate to human with findings.

### Parallel Exploration (Task Tool)

Use for: uncertain decisions, codebase surveys, implementing and voting on approaches, subtasks (N=2-5).

- Use Git Worktree if necessary.
- Paraphrase prompts for each agent to ensure independence.
- Prefer simpler, more testable proposals when voting.
