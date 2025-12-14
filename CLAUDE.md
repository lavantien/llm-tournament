# CLAUDE.md (Local)

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

---

## 0) Core Goal

Deliver **end-to-end working software** with **high confidence** via:
- **Dialectical autocoding** (Coach ↔ Player loop; adversarial cooperation).
- **MDAP/MAKER reliability principles** (maximal decomposition, step-level error correction, red-flagging).
- **Strict TDD** (red → green → refactor), enforced by repository hooks.

The default mindset is: **do not trust self-reported completion**. Only trust **independent verification**.

---

## 1) Non-Negotiables (Repository Constraints)

- **TDD is mandatory**
  - Always write a failing test first.
  - TDD-Guard hooks will block edits if tests fail.
  - Refactor only after tests are green and linters passed.
- **Run the linter** for correctness and style.
- **Go race checks**
  - Prefix commands with `CGO_ENABLED=1` when running race checks.
- **Generated code**
  - Never edit generated code in any `gen/` directory.
  - Run `go generate` to regenerate from OpenAPI specs.
- **Commits**
  - No Claude Code watermarks.
  - No `Co-Authored-By` lines in commits.

---

## 2) The Requirements Contract (Prevents “Vibe Coding Drift”)

For any non-trivial task, create a **requirements contract** before touching code. This is the anchor for the Coach.

**Contract must include:**
- **Goal** (what success looks like)
- **Acceptance criteria** (observable + testable)
- **Non-goals** (explicit out-of-scope)
- **Constraints** (language, APIs, performance, security, repo rules)
- **Verification plan** (exact commands/tests you will run)

**Rule:** If you cannot write acceptance criteria, you do not understand the task yet—pause and clarify.

---

## 3) Dialectical Autocoding Loop (Adversarial Cooperation)

Use a bounded Coach/Player loop to avoid premature “done” claims and to systematically close gaps.

### Roles

**PLAYER (Implementation)**
- Implements, writes tests/harnesses, runs commands.
- Focuses on minimal correct changes.
- Responds only to specific Coach feedback.

**COACH (Independent Review)**
- Validates against requirements contract (not Player claims).
- Runs/requests concrete checks (tests, lint, build, manual repro).
- Produces crisp, actionable deltas.

### Turn Structure (Bounded Loop)

Default bounds:
- **Max turns:** 6 (increase to 10 only for large tasks)
- **Each turn ends with a verifiable artifact:** tests run, diff, or reproducible command results.

#### Player Turn Output Format (mandatory)

1. **Plan for this turn** (≤ 8 bullet points)
2. **Changes made** (files + what/why)
3. **Commands run + results**
4. **What remains** (if anything)

#### Coach Turn Output Format (mandatory)

1. **Requirements compliance matrix**
   - ✅ satisfied / ❌ missing / ⚠️ partial / ? unknown
2. **Verification evidence**
   - What was run, what failed, what is unverified
3. **Defects / gaps**
   - Prioritized list, with concrete pointers
4. **Next actions (atomic)**
   - “Do X in file Y”, “Add test Z”, “Run command Q”, etc.
5. **Approval gate**
   - Only say **COACH APPROVED** if all acceptance criteria are met and verified.

**Rule:** The loop ends only with **COACH APPROVED** or with an explicit “Blocked / needs human input”.

---

## 4) MDAP / MAKER Protocol (Reliability at Scale)

When tasks become complex, long-horizon, or error-prone, switch from “smart monolith” to **massively decomposed agentic process**:

### 4.1 Maximal Decomposition (MAD)
Break work into the **smallest atomic steps** that cannot be misunderstood, e.g.:
- “Create file X”
- “Add struct Y with fields …”
- “Add failing test for scenario Z”
- “Implement minimal code to pass only that test”

**Rule of thumb:** If you are guessing or writing >10 lines of logic without a test, you broke MAD/TDD. Stop and decompose.

### 4.2 Step-Level Error Correction (Voting)
For uncertain design/implementation decisions, use **multi-agent voting** rather than single-agent confidence.

**Default voting rule (practical “first-to-ahead-by-k” analogue):**
- Get **N = 2k − 1** independent proposals (or runs), with **k = 2 or 3** depending on risk.
- Choose the option that is:
  1) most consistent with the requirements contract, and  
  2) easiest to verify with tests, and  
  3) minimal in surface area.

Examples of what to vote on:
- API shape, edge-case behavior, error semantics
- Parsing/validation rules
- Concurrency approach
- Test strategy for hard-to-reproduce bugs

**Important:** Voting is only useful if samples are independent:
- Use separate git worktrees + Haiku subagents when possible.
- Paraphrase the subtask prompt slightly per agent to reduce correlated mistakes.

### 4.3 Red-Flagging (Discard Risky Samples)
Do not “repair” questionable outputs. Discard and resample when you hit red flags.

**Red flags in coding workflows include:**
- Claims of success without commands/test output.
- Leaving `TODO`, commented-out tests, or “should work” statements.
- Suggestions that violate repo rules (editing `gen/`, skipping TDD, etc.).
- Massive diffs that exceed the declared scope.
- Unbounded refactors before a failing test exists.
- Adding dependencies without justification + lockfile updates + security review.
- Any ambiguity about requirements or acceptance tests.

When a red flag occurs:
1) Stop
2) Narrow the step
3) Resample (or spawn subagent)
4) Verify with a minimal test

---

## 5) Tool / Model Utilization (Mapped to Coach/Player)

Use tools/models intentionally (this mirrors the coach/player paradigm and reduces context pollution).

- **Opus (Coach / Overseer)**
  - Planning, requirements contract creation, risk review, acceptance criteria, test strategy.
  - Reviews Player output and produces delta-focused feedback.
- **Sonnet (Player)**
  - Code edits, running commands, iterative TDD cycles, mechanical refactors.
- **Haiku subagents (Microagents)**
  - Parallel exploration on separate worktrees.
  - Used for voting, alternative test approaches, or competing implementations.
- **Web Search + Context7 MCP (from MCP_DOCKER)**
  - Only when you need up-to-date docs, APIs, behaviors, or niche info.
- **Playwright MCP (from MCP_DOCKER)**
  - Browser-driven verification tasks.

---

## 6) Testing & Verification Discipline (No “Trust Me”)

### Always prefer executable verification:
- Unit tests for pure logic
- Integration tests for I/O boundaries
- End-to-end tests for user-visible behavior (when feasible)

### Minimum verification set for most changes:
- `go test ./...`
- Add `-race` when appropriate: `CGO_ENABLED=1 go test -race ./...`
- Lint command(s) used by the repo
- Any golden-file or snapshot tests updated intentionally (with explanation)

**Rule:** Never declare “done” unless you can list:
- tests run,
- commands run,
- and the evidence that acceptance criteria are met.

---

## 7) “Stuck” Protocol (After 3 Failed Attempts)

If you hit three failed attempts (same class of failure):
1. Use `/rewind` to return to a known-good state.
2. Re-read the requirements contract.
3. Switch to MAD: decompose further until steps are trivial.
4. Spawn a **council** (swarm voting):
   - At least 3 Haiku microagents on separate worktrees.
   - Vote on: root cause hypothesis + smallest fix + minimal test that proves it.

---

## 8) MAKER/DAP Protocol (Local Variant)

**Trigger:** complex task, repeated errors, uncertain correctness, or long horizon.

**Principle:** Replace “cleverness” with **extreme decomposition + verification + error correction**.

### MAKER Phase (Decompose)
- Stop coding.
- Break into atomic subtasks.
- Each subtask must have a crisp expected outcome.
- Sequence subtasks so each produces evidence for the next.

### DAP/TDD Phase (Execute)
For each subtask:
1. **Red:** write the smallest failing test.
2. **Green:** minimal code to pass that test.
3. **Verify:** run tests + lint.
4. **Vote (when uncertain):** gather independent solutions.
5. **Refactor:** only when green.
6. Mark done; move on.

---

## 9) Design Hygiene (Keep It Maintainable)

- Prefer small, composable functions.
- Avoid deep nesting; refactor into helpers.
- Don’t over-engineer for hypothetical futures.
- Keep dependencies minimal; justify additions.
- When refactoring, preserve behavior with tests first.

---

## 10) Definition of Done (Coach Approval Criteria)

A task is done only when:
- All requirements contract acceptance criteria are ✅
- Tests are green and reported with commands
- Lint is clean (or exceptions are documented)
- No repo rules were violated
- No critical TODOs remain
- The Coach explicitly outputs: **COACH APPROVED**

---

## Working with the Codebase

### Project Overview

LLM Tournament Arena is a Go-based web application for benchmarking and evaluating Large Language Models. It uses:
- Single-binary deployment with embedded templates and assets
- WebSocket-based real-time updates across all connected clients
- SQLite for persistent data storage (requires CGO_ENABLED=1)
- Server-side rendering with Go templates, Markdown support (Blackfriday), and XSS protection (Bluemonday)
- Client-side vanilla JavaScript with Chart.js for visualizations

### Common Commands

**Development:**
```bash
./dev.sh              # Quick start with auto-recompile (using aider)
make run              # Run the application
go run .              # Alternative run command
make build            # Build for Linux/Mac
make buildwindows     # Build for Windows
```

**Testing:**
```bash
make test             # Run all tests with race detection and coverage
CGO_ENABLED=1 go test ./... -v -race -cover
```

**Database Operations:**
```bash
make setenv           # Setup CGO (required for SQLite)
make migrate          # Migrate from JSON to SQLite
make dedup            # Remove duplicate prompts
go run main.go --remigrate-scores        # Remigrate only scores from JSON
go run main.go --migrate-results         # Migrate old result format to new scoring system
```

**CRITICAL:** Always prefix Go commands with `CGO_ENABLED=1` when race checks are needed, as this project uses SQLite (via go-sqlite3) which requires CGO.

### Architecture

**3-Tier Structure:**

1. **Handlers** (`handlers/`): HTTP request handling and template rendering
   - `models.go`: Model CRUD operations (add, edit, delete models)
   - `profiles.go`: Profile management (evaluation categories)
   - `prompt.go`: Prompt operations (add, edit, delete, move, bulk operations, import/export)
   - `results.go`: Results display, score updates, mock data generation
   - `stats.go`: Analytics, tier classification, score breakdowns
   - `suites.go`: Test suite management (create, rename, delete, switch)

2. **Middleware** (`middleware/`): Business logic and data operations
   - `database.go`: SQLite schema, CRUD operations, migration logic from JSON
   - `state.go`: Core data models (Prompt, Result, Profile), read/write functions
   - `socket.go`: WebSocket handling, real-time broadcast of updates
   - `utils.go`: Profile grouping utilities, color generation
   - `handler_utils.go`: HTTP response helpers
   - `import_error.go`: Import error page handler

3. **Templates** (`templates/`): HTML templates, CSS, and JavaScript
   - Server-side: Go templates with custom functions (markdown, inc, json, etc.)
   - Styling: `style.css` (centralized, no inline styles)
   - Client-side utilities:
     - `score-utils.js`: Centralized score color management
     - `utils.js`: Common JavaScript utilities
     - `constants.js`: Shared constants
   - Pages: Separate HTML files for each route (prompts, results, stats, profiles, etc.)

**Request Flow:**
```
HTTP Request → Router (main.go) → Handler → Middleware → SQLite Database → Response
WebSocket: Score Update → BroadcastResults() → All Connected Clients → UI Update
```

**Suite Management:**
- All data (prompts, profiles, models, scores) is scoped to test suites
- One suite is "current" at a time (tracked by `is_current` flag in database)
- Switching suites instantly updates the entire UI via WebSocket broadcast
- Default suite ("default") is always present and cannot be deleted

**Database Schema (SQLite):**
- `suites`: Test suite definitions (id, name, is_current)
- `profiles`: Evaluation categories (id, name, description, suite_id)
- `prompts`: Test prompts (id, text, solution, profile_id, suite_id, display_order)
- `models`: LLM models being evaluated (id, name, suite_id)
- `scores`: Individual prompt scores per model (id, model_id, prompt_id, score)

**Key Relationships:**
- Foreign keys with CASCADE DELETE maintain referential integrity
- `display_order` (not ID) controls prompt ordering within suites
- Profile assignment is optional for prompts (NULL profile_id = "Uncategorized")
- UNIQUE constraints prevent duplicates: (text, suite_id) for prompts, (name, suite_id) for profiles/models

### Key Implementation Details

**Scoring System:**
- 6-level scale: 0, 20, 40, 60, 80, 100 points per prompt
- 12 performance tiers based on total score:
  - Transcendental (≥3780), Cosmic (3360-3779), Divine (2700-3359), Celestial (2400-2699)
  - Ascendant (2100-2399), Ethereal (1800-2099), Mystic (1500-1799), Astral (1200-1499)
  - Spiritual (900-1199), Primal (600-899), Mortal (300-599), Primordial (<300)
- Score color mapping centralized in `templates/score-utils.js`
- Mock score generation uses weighted distributions reflecting performance tiers

**Real-Time Updates (WebSocket):**
- Connection endpoint: `/ws`
- Managed in `middleware/socket.go`
- `BroadcastResults()` function pushes updates to all connected clients
- Called after any data modification: prompts, models, profiles, scores, suite changes
- Auto-reconnection with connection status monitoring on client side
- Message types: `results` (data updates), `update_prompts_order` (drag-and-drop reordering)

**Profile Groups:**
- Dynamic grouping of prompts by assigned profile
- Color-coded visual borders for group separation in results table
- Utility function `GetProfileGroups()` in `middleware/utils.go` generates groups
- Profile group contiguity maintained (prompts cannot be moved out of their profile group)
- Empty profiles are displayed in the UI even if no prompts assigned

**Migration System:**
- JSON-to-SQLite migration (`--migrate-to-sqlite`) preserves all historical data
- Duplicate cleanup (`--cleanup-duplicates`) removes prompts with identical text within same suite
- Score remigration (`--remigrate-scores`) updates only scores without touching other data
- All migrations use transactions for atomic operations and rollback on error
- Legacy result format migration (`--migrate-results`) converts old scoring to 0-100 scale

**Template Rendering:**
- Markdown support via Blackfriday (GitHub-flavored Markdown)
- XSS protection via Bluemonday sanitization (UGCPolicy)
- Custom template functions registered in handlers:
  - `markdown`: Converts Markdown to sanitized HTML
  - `inc`: Increments integers (for 1-based indexing)
  - `json`: Marshals data to JSON strings
  - `string`: Converts integers to strings
  - `tolower`, `contains`: String utilities
- Shared navigation bar in `templates/nav.html`
- Responsive design with sticky headers, tooltips, and keyboard navigation

### Development Workflow

**Adding New Features:**
1. **Handler Layer:** Create/modify handler function in appropriate `handlers/*.go` file
   - Parse request parameters and form data
   - Call middleware functions for business logic
   - Render template with data or redirect
   - Handle errors with appropriate HTTP status codes

2. **Middleware Layer:** Add business logic and data operations in `middleware/*.go`
   - Implement CRUD operations using prepared statements
   - Use transactions for multi-step operations
   - Call `BroadcastResults()` after state changes to update all clients
   - Log operations for debugging

3. **Database Layer:** Update schema in `middleware/database.go` if needed
   - Add new tables or columns in `createTables()` function
   - Use foreign keys with appropriate CASCADE rules
   - Add migration logic if modifying existing schema
   - Test with both empty and populated databases

4. **Template Layer:** Create/update HTML templates in `templates/`
   - Use shared navigation (`templates/nav.html`)
   - Follow existing CSS classes and patterns in `style.css`
   - Centralize JavaScript utilities in separate `.js` files
   - Use template functions for Markdown rendering and sanitization

5. **Routing:** Register route in `main.go` routes map
   ```go
   "/your_endpoint": handlers.YourHandler,
   ```

6. **Real-Time Updates:** Call `middleware.BroadcastResults()` after state changes
   - Called automatically in most CRUD operations
   - Ensures all connected clients see updates immediately

**Working with Prompts:**
- Prompts have three fields: text (required), solution (Markdown), profile assignment (optional)
- `display_order` field controls ordering, not ID (allows reordering without changing IDs)
- Drag-and-drop reordering sends WebSocket message: `{"type": "update_prompts_order", "order": [...]}`
- Bulk operations supported: selection, deletion, export to JSON
- Import validates JSON structure and prevents duplicate prompts within suite
- Moving prompts restricted to maintain profile group contiguity

**Working with Models:**
- Models are simple name strings scoped to a suite (no complex metadata)
- Adding new model automatically initializes scores array (all zeros) for all prompts
- Deleting model cascades to remove all associated scores (foreign key constraint)
- Models sorted by total score in descending order on results page
- Mock score generation uses tiered weighted distributions for realistic prototyping

**Working with Profiles:**
- Profiles have name and description (Markdown) fields
- Used to categorize prompts (e.g., "Coding", "Math", "Writing")
- Renaming profile automatically updates all associated prompts (referential integrity)
- Deleting profile sets prompts' profile_id to NULL (SET NULL constraint)
- Profile-based filtering in prompt list view for focused evaluation

**Error Handling:**
- Import errors redirect to `/import_error` with descriptive message
- WebSocket errors logged, connections automatically cleaned from clients map
- Database errors return HTTP 500 with error message logged
- Validation errors return HTTP 400 with specific error description
- In case of strange errors or uncertainty, ask user to perform web search for latest information instead of guessing

**Security:**
- **XSS Protection:** Bluemonday UGCPolicy sanitizes all Markdown content before rendering
- **CORS:** WebSocket upgrader `CheckOrigin` returns true (configured for local dev, restrict in production)
- **SQL Injection:** Parameterized queries with prepared statements throughout
- **Input Validation:** Form parsing with error handling, type checking, and bounds validation
- **File Uploads:** JSON structure validation for imports, reject malformed data

### Code Quality Guidelines

**Fundamental Principles (from system prompts):**
- Write clean, simple, readable code (KISS - Keep It Simple, Stupid)
- Implement features in the simplest possible way
- Avoid adding functionality until necessary (YAGNI - You Aren't Gonna Need It)
- Don't repeat yourself; refactor to eliminate duplication (DRY - Don't Repeat Yourself)
- Keep files small and focused (ideally <200 lines, though some files in this project are larger)
- Test after every meaningful change (TDD approach)
- Focus on core functionality before optimization
- Use clear, consistent naming conventions
- Think thoroughly before coding; write 2-3 reasoning paragraphs for complex changes
- Use clear and easy-to-understand language in code comments

**Error Fixing Approach:**
- DO NOT JUMP TO CONCLUSIONS! Consider multiple possible causes before deciding
- Explain the problem in plain English before attempting fixes
- Make minimal necessary changes, changing as few lines of code as possible
- Break the problem down into smallest required steps
- If encountering strange errors or uncertainty, ask user to perform web search for latest information

**Project-Specific Patterns:**
- Handlers focus on HTTP concerns (parsing, validation, rendering)
- Middleware contains business logic and database operations
- Templates use custom functions for common transformations
- WebSocket broadcasts used for all state changes affecting UI
- Transactions wrap multi-step database operations
- Logging at INFO level for normal operations, ERROR for failures

### Testing Notes

- **Current State:** No test files exist (`*_test.go`)
- **When Adding Tests:** Follow TDD approach from local instructions
  - Write failing test first (Red)
  - Implement minimal code to pass (Green)
  - Refactor while keeping tests passing
- **Race Detection:** Enabled in `make test` target
- **CGO Requirement:** Must enable CGO for tests involving SQLite: `CGO_ENABLED=1 go test ./...`
- **Coverage:** Use `-cover` flag to track test coverage
- **Test Structure:** Place tests in same package as code being tested (e.g., `handlers/models_test.go`)

### Server Configuration

- **Port:** `:8080`
- **Access URL:** `http://localhost:8080`
- **WebSocket Endpoint:** `/ws`
- **Default Route:** All unmatched routes redirect to `/prompts`
- **Static Assets:** Served from `templates/` and `assets/` directories
- **Database Path:** `data/tournament.db` (default, configurable via `--db` flag)

### File Structure Summary

```
llm-tournament/
├── main.go                    # Entry point, routing, server setup
├── handlers/                  # HTTP request handlers
│   ├── models.go             # Model CRUD
│   ├── profiles.go           # Profile management
│   ├── prompt.go             # Prompt operations
│   ├── results.go            # Results display, scoring
│   ├── stats.go              # Analytics, tier classification
│   └── suites.go             # Suite management
├── middleware/                # Business logic, data layer
│   ├── database.go           # SQLite schema, migrations
│   ├── state.go              # Data models, CRUD functions
│   ├── socket.go             # WebSocket handling
│   ├── utils.go              # Profile grouping utilities
│   ├── handler_utils.go      # HTTP helpers
│   └── import_error.go       # Error handling
├── templates/                 # HTML, CSS, JavaScript
│   ├── *.html                # Page templates
│   ├── style.css             # Centralized styles
│   ├── score-utils.js        # Score color utilities
│   ├── utils.js              # JavaScript utilities
│   └── constants.js          # Shared constants
├── data/                      # SQLite database, legacy JSON
├── tools/                     # Complementary utilities (TTS, background removal, LLM integration)
├── makefile                   # Build commands
├── go.mod                     # Go dependencies
└── CLAUDE.md                  # This file
```

### Debugging Tips

- **WebSocket Issues:** Check browser console for connection errors, verify `/ws` endpoint accessibility
- **Database Errors:** Enable SQLite logging, check foreign key constraints, verify transactions committed
- **Template Rendering Errors:** Check for nil data, verify template function signatures match usage
- **Score Calculation Issues:** Verify `display_order` values are consecutive and start at 0
- **Profile Grouping Issues:** Check `GetProfileGroups()` logic, verify profile names match exactly (case-sensitive)
- **Migration Failures:** Check transaction rollback logs, verify source JSON format matches expected structure

### Common Gotchas

- **CGO Required:** SQLite driver requires CGO, always use `CGO_ENABLED=1` for builds/tests
- **display_order vs ID:** Prompt ordering uses `display_order` field, not primary key `id`
- **Suite Scoping:** All queries must filter by `suite_id` to avoid cross-suite data leakage
- **WebSocket State:** Clients must reconnect after server restart, no persistent connection recovery
- **Cascade Deletes:** Deleting suite/model cascades to related data, ensure user confirmation prompts
- **Markdown Rendering:** Always sanitize user input through Bluemonday before displaying as HTML
- **Transaction Rollback:** Always defer rollback in transaction error paths to avoid locks

### Recommended Development Sequence for New Features

1. **Design Phase:**
   - Think through the complete feature requirement (2-3 reasoning paragraphs)
   - Identify affected components (handler, middleware, database, templates)
   - Consider WebSocket broadcast needs and real-time update requirements
   - Plan database schema changes if needed

2. **Implementation Phase (MAKER/DAP Protocol for Complex Features):**
   - Break into smallest atomic subtasks
   - Write test for each subtask (if TDD approach)
   - Implement minimal code to pass test
   - Verify and refactor
   - Call `BroadcastResults()` after state changes
   - Test manually with browser

3. **Integration Phase:**
   - Verify WebSocket updates work across multiple browser tabs
   - Test suite switching preserves feature functionality
   - Check error handling paths
   - Verify transaction rollback on failures

4. **Documentation Phase:**
   - Add code comments explaining non-obvious logic
   - Update this CLAUDE.md if adding new patterns or conventions
   - Consider adding example usage if feature is complex
