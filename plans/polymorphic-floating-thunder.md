# Refactoring Plan: Achieve 90%+ Test Coverage via Dependency Injection

## Current State
- **Overall coverage**: 74.3%
- **Key gaps**: Template error paths, database transaction failures, WebSocket errors
- **Root cause**: Global state (`var db`, `var clients`, `globalEvaluator`) and direct `template.ParseFiles()` calls

## Target
- **90%+ overall coverage** through interface-based dependency injection

---

## Phase 1: TemplateRenderer Interface

**Goal**: Enable mocking template operations

### New Files
- `middleware/renderer.go` - Interface + default implementation
- `middleware/renderer_test.go` - Tests for renderer

### Interface Definition
```go
type TemplateRenderer interface {
    Render(w http.ResponseWriter, name string, funcMap template.FuncMap, data interface{}, files ...string) error
}

type FileRenderer struct{}  // Default implementation
var DefaultRenderer TemplateRenderer = &FileRenderer{}
```

### Files to Modify
- `middleware/handler_utils.go` - Refactor `RenderTemplate()` to use interface

---

## Phase 2: Refactor Handlers (By Priority)

Modify each handler to use `DefaultRenderer` (injectable for tests):

| Handler File | Template Calls | Coverage Impact |
|--------------|----------------|-----------------|
| `handlers/prompt.go` | 8 | High |
| `handlers/results.go` | 5 | High |
| `handlers/profiles.go` | 4 | Medium |
| `handlers/suites.go` | 3 | Medium |
| `handlers/models.go` | 2 | Low |
| `handlers/stats.go` | 1 | Low |
| `handlers/settings.go` | 1 (uses Must) | Low |

### Pattern Change
```go
// Before (direct call)
t, err := template.ParseFiles("templates/foo.html", "templates/nav.html")

// After (uses interface)
if err := renderer.Render(w, "foo.html", funcMap, data,
    "templates/foo.html", "templates/nav.html"); err != nil {
    http.Error(w, "Error rendering", 500)
    return
}
```

### Test Files to Update
- `handlers/prompt_test.go`
- `handlers/results_test.go`
- `handlers/profiles_test.go`
- `handlers/suites_test.go`
- `handlers/models_test.go`
- `handlers/stats_test.go`
- `handlers/settings_test.go`

---

## Phase 3: DataStore Interface

**Goal**: Enable mocking database operations

### New Files
- `middleware/datastore.go` - Interface definition
- `middleware/datastore_sqlite.go` - SQLite implementation (wraps existing code)

### Interface Definition (Core Methods)
```go
type DataStore interface {
    // Suites
    GetCurrentSuiteID() (int, error)
    ListSuites() ([]string, error)
    SetCurrentSuite(name string) error

    // Profiles
    ReadProfiles() []Profile
    WriteProfiles(profiles []Profile) error

    // Prompts
    ReadPrompts() []Prompt
    WritePrompts(prompts []Prompt) error

    // Results
    ReadResults() map[string]Result
    WriteResults(suite string, results map[string]Result) error

    // Settings
    GetSetting(key string) (string, error)
    SetSetting(key, value string) error
}

var DefaultDataStore DataStore = &SQLiteDataStore{}
```

### Files to Modify
- `middleware/database.go` - Wrap functions in SQLiteDataStore
- `middleware/state.go` - Delegate to DataStore interface

---

## Phase 4: Handler DataStore Integration

Update handlers to use `DefaultDataStore`:
- `handlers/prompt.go`
- `handlers/results.go`
- `handlers/profiles.go`
- `handlers/evaluation.go`
- `handlers/settings.go`
- `handlers/suites.go`

---

## Phase 5: Mock Infrastructure

### New Test Files
- `testutil/mock_renderer.go` - MockRenderer with error injection
- `testutil/mock_datastore.go` - MockDataStore with error injection

### Mock Capabilities
```go
type MockRenderer struct {
    RenderError error  // Inject error
    RenderCalls []RenderCall  // Track calls
}

type MockDataStore struct {
    ReadError  error
    WriteError error
    Data       map[string]interface{}  // In-memory storage
}
```

---

## Phase 6: Error Path Tests

Add tests for each handler covering:
1. Template parse failure -> 500
2. Template execute failure -> 500
3. Database read failure -> 500
4. Database write failure -> 500
5. Invalid input validation -> 400

---

## Execution Order

```
Phase 1 (Renderer Interface)
    |
    v
Phase 2a (prompt.go refactor) --> Phase 2b (results.go) --> Phase 2c (profiles.go)
    |
    v
Phase 3 (DataStore Interface)
    |
    v
Phase 4 (Handler DataStore Integration)
    |
    v
Phase 5 (Mock Infrastructure)
    |
    v
Phase 6 (Error Path Tests)
```

---

## Coverage Projections

| Phase | handlers | middleware | Overall |
|-------|----------|------------|---------|
| Current | 74.4% | 80.4% | 74.3% |
| After Phase 2 | 82% | 82% | 78% |
| After Phase 4 | 88% | 87% | 85% |
| After Phase 6 | 92%+ | 91%+ | 90%+ |

---

## Critical Files Summary

### Create New
- `middleware/renderer.go`
- `middleware/renderer_test.go`
- `middleware/datastore.go`
- `middleware/datastore_sqlite.go`
- `testutil/mock_renderer.go`
- `testutil/mock_datastore.go`

### Modify
- `middleware/handler_utils.go`
- `handlers/prompt.go` (highest impact)
- `handlers/results.go`
- `handlers/profiles.go`
- `handlers/suites.go`
- `handlers/models.go`
- `handlers/stats.go`
- `handlers/settings.go`
- All corresponding `*_test.go` files

---

## TDD Approach

For each phase:
1. Write failing test for error path
2. Implement interface/refactor
3. Verify test passes
4. Run `CGO_ENABLED=1 go test ./... -v -race -cover`
5. Commit changes
