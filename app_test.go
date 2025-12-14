package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"llm-tournament/middleware"
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

func TestInitDB(t *testing.T) {
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
}

func TestInitDB_InvalidPath(t *testing.T) {
	// Try to create DB in a path that doesn't exist and can't be created
	// On most systems, this would fail
	err := InitDB("/nonexistent/deeply/nested/path/that/should/fail/test.db")
	// This might succeed on some systems if they auto-create dirs
	// so we just check it doesn't panic
	if err == nil {
		CloseDB()
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

func TestInitEvaluator(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer CloseDB()

	// Should not panic
	InitEvaluator(GetDB())
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
	middleware.WritePrompts([]middleware.Prompt{{Text: "Test prompt"}})
	middleware.WriteResults("default", map[string]middleware.Result{
		"Model1": {Scores: []int{50}},
	})

	// Run migration
	err = RunMigration()
	if err != nil {
		t.Errorf("RunMigration failed: %v", err)
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
		"/settings",
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

	// Should not panic
	SetupRoutes(mux)
}

func TestNewServeMux(t *testing.T) {
	mux := NewServeMux()

	if mux == nil {
		t.Fatal("expected non-nil ServeMux")
	}
}

func TestApp_Router_KnownRoute(t *testing.T) {
	// Save current directory and change to project root for template access
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	// Find project root
	projectRoot := findProjectRoot()
	if projectRoot != "" {
		os.Chdir(projectRoot)
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
	defer os.Chdir(originalDir)

	projectRoot := findProjectRoot()
	if projectRoot != "" {
		os.Chdir(projectRoot)
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
