package main

import (
	"llm-tournament/middleware"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.DBPath != "data/tournament.db" {
		t.Errorf("expected DBPath 'data/tournament.db', got %q", cfg.DBPath)
	}
	if cfg.Port != ":8080" {
		t.Errorf("expected Port ':8080', got %q", cfg.Port)
	}
	if cfg.MigrateResults {
		t.Error("expected MigrateResults to be false by default")
	}
}

func TestParseFlags_Defaults(t *testing.T) {
	cfg, err := ParseFlags([]string{})
	if err != nil {
		t.Fatalf("ParseFlags failed: %v", err)
	}

	if cfg.DBPath != "data/tournament.db" {
		t.Errorf("expected default DBPath, got %q", cfg.DBPath)
	}
	if cfg.MigrateResults {
		t.Error("expected MigrateResults to be false by default")
	}
}

func TestParseFlags_CustomDBPath(t *testing.T) {
	cfg, err := ParseFlags([]string{"-db", "/custom/path.db"})
	if err != nil {
		t.Fatalf("ParseFlags failed: %v", err)
	}

	if cfg.DBPath != "/custom/path.db" {
		t.Errorf("expected DBPath '/custom/path.db', got %q", cfg.DBPath)
	}
}

func TestParseFlags_MigrateResults(t *testing.T) {
	cfg, err := ParseFlags([]string{"-migrate-results"})
	if err != nil {
		t.Fatalf("ParseFlags failed: %v", err)
	}

	if !cfg.MigrateResults {
		t.Error("expected MigrateResults to be true")
	}
}

func TestParseFlags_AllOptions(t *testing.T) {
	cfg, err := ParseFlags([]string{"-db", "/my/db.sqlite", "-migrate-results"})
	if err != nil {
		t.Fatalf("ParseFlags failed: %v", err)
	}

	if cfg.DBPath != "/my/db.sqlite" {
		t.Errorf("expected DBPath '/my/db.sqlite', got %q", cfg.DBPath)
	}
	if !cfg.MigrateResults {
		t.Error("expected MigrateResults to be true")
	}
}

func TestParseFlags_InvalidFlag(t *testing.T) {
	_, err := ParseFlags([]string{"-invalid-flag"})
	if err == nil {
		t.Error("expected error for invalid flag")
	}
}

func TestInitDB_InvalidPath(t *testing.T) {
	// InitDB creates the directory tree, so a merely nonexistent path
	// succeeds; a directory component occupied by a regular file must fail.
	blocker := filepath.Join(t.TempDir(), "blocker")
	if err := os.WriteFile(blocker, []byte("not a directory"), 0o644); err != nil {
		t.Fatalf("failed to create blocker file: %v", err)
	}

	err := InitDB(filepath.Join(blocker, "nested", "test.db"))
	if err == nil {
		CloseDB()
		t.Fatal("expected InitDB to fail when a directory component is a regular file")
	}
}

func TestCloseDB(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// Should not panic
	CloseDB()
}

func TestGetDB(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer CloseDB()

	db := GetDB()
	if db == nil {
		t.Error("expected non-nil database")
	}

	// Verify DB is functional
	err = db.Ping()
	if err != nil {
		t.Errorf("database ping failed: %v", err)
	}
}

func TestRunMigration(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer CloseDB()

	// Run migration on empty database - should succeed
	err = RunMigration()
	if err != nil {
		t.Errorf("RunMigration failed: %v", err)
	}
}

func TestRunMigration_WithData(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer CloseDB()

	// Add some test data
	if err := middleware.WritePrompts([]middleware.Prompt{{Text: "Test prompt"}}); err != nil {
		t.Fatalf("failed to write prompts: %v", err)
	}
	if err := middleware.WriteResults("default", map[string]middleware.Result{
		"Model1": {Scores: []int{50}},
	}); err != nil {
		t.Fatalf("failed to write results: %v", err)
	}

	// Run migration
	err = RunMigration()
	if err != nil {
		t.Errorf("RunMigration failed: %v", err)
	}

	// Verify the data survived: prompts unchanged, results readable and
	// still holding the written score (1 prompt, in range, so unchanged)
	gotPrompts := middleware.ReadPrompts()
	wantPrompts := []middleware.Prompt{{ID: 1, Text: "Test prompt"}}
	if !reflect.DeepEqual(gotPrompts, wantPrompts) {
		t.Errorf("prompts after migration: got %#v, want %#v", gotPrompts, wantPrompts)
	}

	gotResults := middleware.ReadResults()
	wantResults := map[string]middleware.Result{"Model1": {Scores: []int{50}}}
	if !reflect.DeepEqual(gotResults, wantResults) {
		t.Errorf("results after migration: got %#v, want %#v", gotResults, wantResults)
	}
}

func TestRoutes(t *testing.T) {
	routes := Routes()

	if routes == nil {
		t.Fatal("expected non-nil routes")
	}

	// Check some expected routes exist
	expectedRoutes := []string{
		"/prompts",
		"/results",
		"/profiles",
		"/stats",
		"/evaluate",
		"/save_model_response",
		"/add_model",
		"/delete_model",
	}

	for _, route := range expectedRoutes {
		if _, ok := routes[route]; !ok {
			t.Errorf("expected route %q not found", route)
		}
	}
}

func TestSetupRoutes(t *testing.T) {
	mux := http.NewServeMux()

	SetupRoutes(mux)

	// Every route in the routes map must be registered on the mux: an exact
	// request for the pattern resolves back to that same pattern (duplicate
	// registrations panic inside SetupRoutes, so a full census is impossible
	// without an enumeration API; exact resolution is the observable contract).
	for pattern := range routes {
		req := httptest.NewRequest(http.MethodGet, pattern, nil)
		handler, matched := mux.Handler(req)
		if handler == nil {
			t.Errorf("route pattern %q did not resolve to a handler", pattern)
		}
		if matched != pattern {
			t.Errorf("path %q resolved to pattern %q, want %q", pattern, matched, pattern)
		}
	}

	// Plus the four mux-level patterns registered by SetupRoutes itself
	for path, wantPattern := range map[string]string{
		"/":                     "/",
		"/ws":                   "/ws",
		"/templates/output.css": "/templates/",
		"/assets/app.js":        "/assets/",
	} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		handler, matched := mux.Handler(req)
		if handler == nil {
			t.Errorf("path %q did not resolve to a handler", path)
		}
		if matched != wantPattern {
			t.Errorf("path %q resolved to pattern %q, want %q", path, matched, wantPattern)
		}
	}

	// An unregistered path must fall through to the catch-all "/" pattern
	req := httptest.NewRequest(http.MethodGet, "/definitely/not/registered", nil)
	_, matched := mux.Handler(req)
	if matched != "/" {
		t.Errorf("unregistered path resolved to pattern %q, want %q", matched, "/")
	}
}

func TestNewServeMux(t *testing.T) {
	mux := NewServeMux()

	if mux == nil {
		t.Fatal("expected non-nil ServeMux")
	}
}

func TestServerHandler_RejectsCrossOriginPost(t *testing.T) {
	originalDir, _ := os.Getwd()
	defer func() {
		_ = os.Chdir(originalDir)
	}()
	if projectRoot := findProjectRoot(); projectRoot != "" {
		if err := os.Chdir(projectRoot); err != nil {
			t.Fatalf("failed to change dir: %v", err)
		}
	}

	tmpDir := t.TempDir()
	if err := InitDB(filepath.Join(tmpDir, "test.db")); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer CloseDB()

	handler := ServerHandler()

	req := httptest.NewRequest("POST", "/reset_results", nil)
	req.Host = "localhost:8080"
	req.Header.Set("Origin", "http://evil.example")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected cross-origin POST to be rejected with %d, got %d", http.StatusForbidden, rr.Code)
	}

	req = httptest.NewRequest("POST", "/reset_results", nil)
	req.Host = "localhost:8080"
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code == http.StatusForbidden {
		t.Errorf("expected same-origin POST without Origin header to pass, got %d", rr.Code)
	}
}

func TestServerHandler_TemplatesStaticServing(t *testing.T) {
	originalDir, _ := os.Getwd()
	defer func() {
		_ = os.Chdir(originalDir)
	}()
	if projectRoot := findProjectRoot(); projectRoot != "" {
		if err := os.Chdir(projectRoot); err != nil {
			t.Fatalf("failed to change dir: %v", err)
		}
	}

	handler := ServerHandler()

	for path, wantStatus := range map[string]int{
		"/templates/output.css":          http.StatusOK,
		"/templates/utils.js":            http.StatusOK,
		"/templates/shared.go":           http.StatusNotFound,
		"/templates/prompt_list.html":    http.StatusNotFound,
		"/templates/design_preview.html": http.StatusNotFound,
	} {
		req := httptest.NewRequest("GET", path, nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != wantStatus {
			t.Errorf("GET %s: expected status %d, got %d", path, wantStatus, rr.Code)
		}
	}
}

func TestApp_Router_KnownRoute(t *testing.T) {
	// Save current directory and change to project root for template access
	originalDir, _ := os.Getwd()
	defer func() {
		_ = os.Chdir(originalDir)
	}()

	// Find project root
	projectRoot := findProjectRoot()
	if projectRoot != "" {
		if err := os.Chdir(projectRoot); err != nil {
			t.Fatalf("failed to change dir: %v", err)
		}
	}

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer CloseDB()

	req := httptest.NewRequest("GET", "/prompts", nil)
	rr := httptest.NewRecorder()

	Router(rr, req)

	// Should not be a redirect (303) since /prompts is a known route
	// It should render the prompts page (200) or handle it
	if rr.Code == http.StatusSeeOther {
		t.Errorf("expected non-redirect for known route /prompts, got %d", rr.Code)
	}
}

func TestApp_Router_UnknownRoute(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer CloseDB()

	req := httptest.NewRequest("GET", "/unknown/route/that/does/not/exist", nil)
	rr := httptest.NewRecorder()

	Router(rr, req)

	// Unknown routes should redirect to /prompts
	if rr.Code != http.StatusSeeOther {
		t.Errorf("expected redirect (303) for unknown route, got %d", rr.Code)
	}

	location := rr.Header().Get("Location")
	if location != "/prompts" {
		t.Errorf("expected redirect to /prompts, got %q", location)
	}
}

func TestNewServeMux_Integration(t *testing.T) {
	// Save current directory and change to project root for template access
	originalDir, _ := os.Getwd()
	defer func() {
		_ = os.Chdir(originalDir)
	}()

	projectRoot := findProjectRoot()
	if projectRoot != "" {
		if err := os.Chdir(projectRoot); err != nil {
			t.Fatalf("failed to change dir: %v", err)
		}
	}

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer CloseDB()

	mux := NewServeMux()

	// Test a route through the mux
	req := httptest.NewRequest("GET", "/results", nil)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	// Results page should render (200) or redirect
	// Not be a 404
	if rr.Code == http.StatusNotFound {
		t.Error("expected route to be found")
	}
}

// findProjectRoot looks for the go.mod file to find project root
func findProjectRoot() string {
	dir, _ := os.Getwd()
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

func TestSetupRoutes_WithMux(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer CloseDB()

	mux := http.NewServeMux()
	SetupRoutes(mux)

	// End-to-end dispatch behavior through the mux (pattern registration is
	// censused by TestSetupRoutes)
	testCases := []struct {
		path       string
		wantStatus int
		location   string
	}{
		{path: "/", wantStatus: http.StatusSeeOther, location: "/prompts"},
		{path: "/unknown-route", wantStatus: http.StatusSeeOther, location: "/prompts"},
		{path: "/save_model_response", wantStatus: http.StatusMethodNotAllowed},
	}

	for _, tc := range testCases {
		req := httptest.NewRequest(http.MethodGet, tc.path, nil)
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)
		if rr.Code != tc.wantStatus {
			t.Errorf("%s: expected status %d, got %d", tc.path, tc.wantStatus, rr.Code)
		}
		if tc.location != "" {
			if location := rr.Header().Get("Location"); location != tc.location {
				t.Errorf("%s: expected redirect to %q, got %q", tc.path, tc.location, location)
			}
		}
	}
}

func TestNewServeMux_AllRoutesRegistered(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer CloseDB()

	mux := NewServeMux()

	// Test that known routes are registered
	knownRoutes := []string{
		"/prompts",
		"/results",
		"/stats",
	}

	for _, route := range knownRoutes {
		req := httptest.NewRequest("GET", route, nil)
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		// Route should be found (not 404 for pattern mismatch)
		// Note: 500 errors are acceptable in test environment without templates
		if rr.Code == http.StatusNotFound {
			t.Errorf("route %s returned 404 - not registered", route)
		}
	}
}

func TestParseFlags_CombinedFlags(t *testing.T) {
	cfg, err := ParseFlags([]string{"-db=/tmp/test.db", "-migrate-results"})
	if err != nil {
		t.Fatalf("ParseFlags failed: %v", err)
	}

	if cfg.DBPath != "/tmp/test.db" {
		t.Errorf("expected DBPath '/tmp/test.db', got %q", cfg.DBPath)
	}

	if !cfg.MigrateResults {
		t.Error("expected MigrateResults to be true")
	}
}

func TestInitDB_AndGetDB(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	db := GetDB()
	if db == nil {
		t.Error("GetDB returned nil")
	}

	// Test DB is functional
	err = db.Ping()
	if err != nil {
		t.Errorf("DB ping failed: %v", err)
	}

	CloseDB()
}
