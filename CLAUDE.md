# CLAUDE.md

## Non-Negotiables

- **TDD is mandatory**: Write failing test first → minimal code to pass → refactor. TDD-Guard hooks block edits if tests fail.
- **Adversarial Cooperation**: Rigorously check against linters and hostile unit tests or security exploits. If complexity requires, ultilize parallel Tasks, Concensus Voting, Synthetic and Fuzzy Test Case Generation with high-quality examples and high volume variations.
- **Common Pitfalls**:
    - **CGO_ENABLED=1**: Always prefix Go commands with this (SQLite requires CGO).
    - **Never edit `gen/` directories**: Run `go generate` to regenerate from OpenAPI specs.
    - **Commits**: No Claude Code watermarks. No `Co-Authored-By` lines.
- **Only trust independent verification**: Never claim "done" without test output and command evidence.

## Core Workflow

### Requirements Contract (Non-Trivial Tasks)
Before coding, define: Goal, Acceptance Criteria (testable), Non-goals, Constraints, Verification Plan.

**Rule**: If you cannot write acceptance criteria, pause and clarify.

### Verification Minimum
```bash
CGO_ENABLED=1 go test ./... -v -race -cover  # All tests with race detection
```

### When Stuck (3 Failed Attempts)
1. Stop coding. Return to last green state.
2. Re-read requirements. Verify solving the RIGHT problem.
3. Decompose into atomic subtasks (<10 lines each).
4. Spawn 3 parallel diagnostic tasks via Task tool.
5. If still blocked → escalate to human with findings.

### Parallel Exploration (Task Tool)
Use for: uncertain decisions, codebase surveys, voting on approaches (N=3).
- Paraphrase prompts for each agent to ensure independence.
- Prefer simpler, more testable proposals when voting.

## Commands

```bash
make run                  # Go server on :8080
make test                 # TDD-guard + race detection + coverage
make test-verbose         # Verbose output, bypasses TDD-guard

# Python evaluation service
cd python_service && python main.py  # Runs on :8001

# With evaluation (requires ENCRYPTION_KEY)
export ENCRYPTION_KEY=$(openssl rand -hex 32)
CGO_ENABLED=1 go run main.go
```

## Architecture

```
handlers/     → HTTP handlers (models, profiles, prompts, results, stats, suites, settings, evaluation)
middleware/   → Business logic (database.go, state.go, socket.go, encryption.go)
evaluator/    → Async LLM evaluation (job_queue.go, litellm_client.go, consensus.go)
templates/    → HTML, CSS, JS (style.css, score-utils.js)
python_service/ → AI Judge service (FastAPI on :8001)
```

**Request Flow**: HTTP → Handler → Middleware → SQLite → Response
**WebSocket**: State change → `BroadcastResults()` → All clients update

## Database Schema

**Core**: `suites`, `profiles`, `prompts` (text, solution, profile_id, display_order, type), `models`, `scores`
**Evaluation**: `settings` (encrypted API keys), `evaluation_jobs`, `model_responses`, `evaluation_history`, `cost_tracking`

**Key Constraints**:
- Foreign keys with CASCADE DELETE
- `display_order` controls prompt ordering (not `id`)
- All queries must filter by `suite_id`
- `prompts.type`: 'objective' | 'creative'

## Key Patterns

- **Real-time updates**: Call `middleware.BroadcastResults()` after any state change
- **Transactions**: Wrap multi-step DB operations; always defer rollback
- **XSS protection**: All Markdown sanitized via Bluemonday before rendering
- **Profile groups**: Prompts grouped by profile; maintain contiguity on reorder

## Scoring

- 6-level scale: 0, 20, 40, 60, 80, 100 per prompt
- 12 tiers: Transcendental (>=3780) → Primordial (<300)
- Colors centralized in `templates/score-utils.js`

## Common Gotchas

| Issue | Solution |
|-------|----------|
| SQLite errors | Use `CGO_ENABLED=1` |
| Wrong prompt order | Use `display_order`, not `id` |
| Cross-suite data leak | Filter all queries by `suite_id` |
| Evaluation fails | Check `ENCRYPTION_KEY` env var (64 hex chars) |
| Python service down | Start `python main.py` on :8001 first |
| WebSocket not updating | Ensure `BroadcastResults()` called after state change |
| Transaction locks | Always `defer tx.Rollback()` in error paths |

## Environment

- Go Server: `:8080`
- Python Service: `:8001`
- WebSocket: `/ws`
- Database: `data/tournament.db`
- Required: `ENCRYPTION_KEY` (for evaluation), `CGO_ENABLED=1`
