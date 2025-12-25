# CLAUDE.md

## Non-Negotiables

* Strict TDD is mandatory: Write failing test first (test-as-documentation, one-at-a-time, regression-proof, table-driven, test-doubles) -> minimal code to pass -> refactor -> using linters & formatters.
* Adversarial Cooperation: Rigorously check against linters and hostile unit tests or security exploits. If complexity requires, utilize parallel Tasks, Consensus Voting, Synthetic and Fuzzy Test Case Generation with high-quality examples and high volume variations.
* Only trust independent verification: Never claim "done" without test output and command evidence.
* Commits & Comments: No watermarks. No `Co-Authored-By` lines. Only plain simple text, maybe with unordered dash list or numbered list, avoid em/en dashes or bolting or italicizing or emojis. For comments, always in my humble voice and stay as unconfrontational as possible and phrase most things as constructive questions.
  * Conventions: Use Conventional Commits (feat, fix, docs, refactor, test, chore).
  * Granularity: Atomic commits. If the logic changes, the test must be committed in the same SHA.
  * Security: Never commit secrets. If a test requires a secret, it must use environment variables or skipped if the variable is missing.

## Test Doubles Principle

Test Doubles for Dependencies: Use dependency injection via interfaces/traits: Define behaviors as contracts, inject implementations (real or mock) at runtime. Mock only external boundaries or owned components to avoid brittle tests; prefer stubs for data provision and mocks for interaction verification. In TDD, write tests assuming isolated units first, then mock as needed for failing specs.

Universal Rule: Define behavior contracts (interfaces/traits), inject dependencies via constructor, mock only at external boundaries.

1. Contract First: Every mockable dependency gets an interface/trait.
2. Constructor Injection: Pass dependencies explicitly—never rely on global state.
3. Mock Boundaries Only: Mock external I/O (network, filesystem, databases), not internal logic.
4. Prefer Stubs over Mocks: Use stubs for data provision; reserve mocks for interaction verification.

## Core Workflow

### Requirements Contract (Non-Trivial Tasks)

#### Before coding, define

1. Goal: What are we solving?
2. Acceptance Criteria: Testable conditions for success.
3. Definition of Done: Explicitly state what files will **NOT** be touched to prevent scope creep.
4. Non-goals & Constraints: What are we avoiding?
5. Verification Plan: How will we prove it works?
6. Security Review: Briefly scan input/output for injection risks or PII leaks.

If you cannot write acceptance criteria, pause and clarify.

#### Tool Usage

* Repomix: Use to explore and pack the repository for full-structure views.
* Context7: Use to acquire up-to-date, version-specific documentation for any library/API.
* Playwright: Use for interactive browser-based E2E tests and UI debugging.
* GitHub CLI: Use `gh` for PRs/Issues.
* Offline Docs: Use `go doc` or equivalences for accurate references.
* Web Search or Fetch: Use if internal docs are insufficient.

### Verification Minimum

Detect the environment and run the **strict** verification chain. If a `Makefile`, `Justfile`, or `Taskfile` exists, prioritize the below first and then apply standard targets after (e.g., `make check`, `just test`).

Prioritize infra if detected (Dockerfile, Chart.yaml, main.tf).

#### Go Verification

```bash
go mod tidy && golangci-lint fmt && golangci-lint run --no-config --timeout=5m && CGO_ENABLED=1 go test ./... -race -cover && CGO_ENABLED=1 go test -bench
```

#### Rust Verification

```bash
cargo fmt -- --check && cargo clippy -- -D warnings && cargo test --all-features && cargo nextest
```

#### TypeScript / Node Verification

```bash
# Detect manager (npm, pnpm, yarn, bun)
<manager> run lint --fix && <manager> run type-check && <manager> test
```

#### Python Verification

```bash
# Prefer Ruff if available, otherwise standard chain
ruff check --fix . && ruff format . && pytest
# OR
black . && isort . && flake8 && pytest
```

#### Java Verification

```bash
mvn clean verify
# OR
./gradlew check
```

#### C# Verification

```bash
dotnet restore && dotnet format && dotnet build && dotnet test
```

#### Scala Verification

```bash
sbt scalafmtCheck && sbt scalafmtSbtCheck && sbt compile && sbt test
```

#### PHP Verification

```bash
composer validate && composer lint --fix && ./vendor/bin/phpstan analyse && ./vendor/bin/phpunit
```

#### GodotScript Verification

```bash
godot --path . -s addons/gut/gut_cmdln.gd -gexit_on_success
```

#### Docker

* Lint with Hadolint, then build.
* Commands: `hadolint Dockerfile && docker build -t app .`

#### Helm

* Lint chart structure and templates.
* Commands: `helm lint chart/ && helm template chart/ --dry-run --debug --validate`

#### Terraform

* Format check, validate syntax, dry-run plan.
* Commands: `terraform fmt -check && terraform validate && terraform plan -out=tfplan`

#### Other Languages / Unknown Environment

1. Check for `Makefile`, `Justfile`, `package.json`, or the like.
2. Read `CONTRIBUTING.md` or `README.md` or the like.
3. Propose a verification command based on findings and ask for user confirmation before running.

### Context Hygiene

If a conversation exceeds 64 turns or context becomes stale:

1. Summarize: Create `checkpoint.md` capturing: Current Goal, Recent Changes, Next Immediate Step, List of Open Questions.
2. Verify: Ensure `checkpoint.md` is committed.
3. Reset: Instruct user to `/reset` (or clear context) and read `checkpoint.md`.

### When Stuck (3 Failed Attempts)

1. Stop coding. Return to last green state (git reset).
2. Re-read requirements. Verify you are solving the RIGHT problem.
3. Decompose into atomic TDD increments: Recursively break the feature into smallest testable units—one behavior or assertion per test. Each subtask targets a single red-green-refactor cycle (<10 lines of code), starting from the leaves (e.g., simplest function) and building up, to maintain steady progress and isolate failures.
4. Constraint: You are forbidden from modifying the test logic to force a pass unless the Requirements Contract has changed.
5. Spawn 2-5 parallel diagnostic tasks via Task tool.
6. If still blocked → escalate to human with findings.

### Parallel Exploration (Task Tool)

Use for: uncertain decisions, codebase surveys, implementing and voting on approaches.

* Cleanup: Use Git Worktree if necessary, but strictly ensure cleanup (`git worktree remove` and branch deletion) occurs regardless of success/failure via a `defer` or `trap` mechanism, or just standard branching if sufficient.
* Independence: Paraphrase prompts for each agent to ensure cognitive diversity.
* Voting: Prefer simpler, more testable proposals.
* Consensus Protocol: When agents disagree, prioritize the solution with the fewest dependencies and highest test coverage. Discard "clever" solutions in favor of "boring" standard library usage.

### Workflow Exception: Trivial Edits

For simple typo fixes, comment updates, or one-line non-logic changes:

1. Skip the "Requirements Contract."
2. Run the linter/formatter only.
3. Commit immediately.

## Language Specific Pitfalls

### Go

* CGO_ENABLED=1: Always prefix Go commands with this (SQLite and Race Detection require CGO).
* Gen Directories: Never edit `gen/`. Run `go generate`, `protoc`, or `sqlc` to regenerate.

### TypeScript/JS

* Type Safety: No `any`. Use `unknown` + narrowing if necessary.
* Lockfiles: Do not mix package managers (pnpm/npm/yarn/bun).

### Ecosystem Libraries

| Language | Preferred Library | Notes |
|----------|-------------------|-------|
| Go | `moq` | Or existing: `testify`, `gomock` |
| Rust | `mockall` | Use `#[automock]` on traits |
| TypeScript | `jest.mock` | Avoid mocking primitives |
| Python | `unittest.mock` | `@patch` decorator |
| Java | Mockito | `@Mock`, `@InjectMocks` |
| C# | Moq | `Mock<T>` pattern |
| Scala | `mockito-scala` | Or `ScalaMock` for strict typing |
| PHP | Mockery | Or PHPUnit built-in mocks |
| GDScript | GUT doubles | Via GUT addon |

> Detect existing patterns in codebase before introducing new libraries.

