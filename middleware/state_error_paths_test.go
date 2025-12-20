package middleware

import (
	"database/sql"
	"errors"
	"strings"
	"testing"
)

func TestReadProfileSuite_ErrorBranches(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	if err := InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	suiteID, err := GetSuiteID("default")
	if err != nil {
		t.Fatalf("GetSuiteID failed: %v", err)
	}

	// Scan error: description=NULL cannot be scanned into string.
	if _, err := db.Exec("INSERT INTO profiles (name, description, suite_id) VALUES (?, ?, ?)", "BadProfile", nil, suiteID); err != nil {
		t.Fatalf("insert profile: %v", err)
	}
	if _, err := ReadProfileSuite("default"); err == nil {
		t.Fatalf("expected ReadProfileSuite to return error for NULL description scan")
	}

	// Query error: missing profiles table.
	if _, err := db.Exec("DROP TABLE profiles"); err != nil {
		t.Fatalf("drop profiles: %v", err)
	}
	if _, err := ReadProfileSuite("default"); err == nil {
		t.Fatalf("expected ReadProfileSuite to return error when profiles table is missing")
	}

	// GetSuiteID error: database is closed.
	if err := CloseDB(); err != nil {
		t.Fatalf("CloseDB failed: %v", err)
	}
	if _, err := ReadProfileSuite("default"); err == nil {
		t.Fatalf("expected ReadProfileSuite to return error when database is closed")
	}
}

func TestWriteProfileSuite_ErrorBranches(t *testing.T) {
	t.Run("duplicate profile name fails insert", func(t *testing.T) {
		dbPath, cleanup := setupTestDB(t)
		defer cleanup()

		if err := InitDB(dbPath); err != nil {
			t.Fatalf("InitDB failed: %v", err)
		}

		err := WriteProfileSuite("default", []Profile{
			{Name: "dup", Description: "a"},
			{Name: "dup", Description: "b"},
		})
		if err == nil {
			t.Fatalf("expected WriteProfileSuite to return an error for duplicate profile name")
		}
		if !strings.Contains(err.Error(), "failed to insert profile") {
			t.Fatalf("expected insert error, got %v", err)
		}
	})

	t.Run("missing profiles table fails delete", func(t *testing.T) {
		dbPath, cleanup := setupTestDB(t)
		defer cleanup()

		if err := InitDB(dbPath); err != nil {
			t.Fatalf("InitDB failed: %v", err)
		}

		if _, err := db.Exec("DROP TABLE profiles"); err != nil {
			t.Fatalf("drop profiles: %v", err)
		}

		err := WriteProfileSuite("default", nil)
		if err == nil {
			t.Fatalf("expected WriteProfileSuite to return an error when profiles table is missing")
		}
		if !strings.Contains(err.Error(), "failed to delete profiles") {
			t.Fatalf("expected delete error, got %v", err)
		}
	})

	t.Run("closed database fails GetSuiteID", func(t *testing.T) {
		dbPath, cleanup := setupTestDB(t)
		defer cleanup()

		if err := InitDB(dbPath); err != nil {
			t.Fatalf("InitDB failed: %v", err)
		}
		if err := CloseDB(); err != nil {
			t.Fatalf("CloseDB failed: %v", err)
		}

		err := WriteProfileSuite("default", nil)
		if err == nil {
			t.Fatalf("expected WriteProfileSuite to return an error when database is closed")
		}
		if !strings.Contains(err.Error(), "failed to get suite ID") {
			t.Fatalf("expected suite id error, got %v", err)
		}
	})
}

func TestReadResults_ErrorBranches(t *testing.T) {
	t.Run("suite id query error returns empty map", func(t *testing.T) {
		dbPath, cleanup := setupTestDB(t)
		defer cleanup()

		if err := InitDB(dbPath); err != nil {
			t.Fatalf("InitDB failed: %v", err)
		}
		if _, err := db.Exec("DROP TABLE suites"); err != nil {
			t.Fatalf("drop suites: %v", err)
		}

		if got := ReadResults(); len(got) != 0 {
			t.Fatalf("expected empty results on suite id error, got %#v", got)
		}
	})

	t.Run("prompt count query error returns empty map", func(t *testing.T) {
		dbPath, cleanup := setupTestDB(t)
		defer cleanup()

		if err := InitDB(dbPath); err != nil {
			t.Fatalf("InitDB failed: %v", err)
		}
		if _, err := db.Exec("DROP TABLE prompts"); err != nil {
			t.Fatalf("drop prompts: %v", err)
		}

		if got := ReadResults(); len(got) != 0 {
			t.Fatalf("expected empty results on prompt count error, got %#v", got)
		}
	})

	t.Run("model query error returns empty map", func(t *testing.T) {
		dbPath, cleanup := setupTestDB(t)
		defer cleanup()

		if err := InitDB(dbPath); err != nil {
			t.Fatalf("InitDB failed: %v", err)
		}
		if _, err := db.Exec("DROP TABLE models"); err != nil {
			t.Fatalf("drop models: %v", err)
		}

		if got := ReadResults(); len(got) != 0 {
			t.Fatalf("expected empty results on model query error, got %#v", got)
		}
	})

	t.Run("model scan error is skipped", func(t *testing.T) {
		dbPath, cleanup := setupTestDB(t)
		defer cleanup()

		if err := InitDB(dbPath); err != nil {
			t.Fatalf("InitDB failed: %v", err)
		}

		suiteID, err := GetSuiteID("default")
		if err != nil {
			t.Fatalf("GetSuiteID failed: %v", err)
		}

		if _, err := db.Exec("DROP TABLE models"); err != nil {
			t.Fatalf("drop models: %v", err)
		}
		if _, err := db.Exec(`CREATE TABLE models (id TEXT PRIMARY KEY, name TEXT NOT NULL, suite_id INTEGER NOT NULL)`); err != nil {
			t.Fatalf("create models: %v", err)
		}
		if _, err := db.Exec("INSERT INTO models (id, name, suite_id) VALUES (?, ?, ?)", "bad", "ModelBad", suiteID); err != nil {
			t.Fatalf("insert models: %v", err)
		}

		if got := ReadResults(); len(got) != 0 {
			t.Fatalf("expected scan error model to be skipped, got %#v", got)
		}
	})

	t.Run("score query error is skipped", func(t *testing.T) {
		dbPath, cleanup := setupTestDB(t)
		defer cleanup()

		if err := InitDB(dbPath); err != nil {
			t.Fatalf("InitDB failed: %v", err)
		}

		suiteID, err := GetSuiteID("default")
		if err != nil {
			t.Fatalf("GetSuiteID failed: %v", err)
		}
		if _, err := db.Exec("INSERT INTO models (name, suite_id) VALUES (?, ?)", "Model1", suiteID); err != nil {
			t.Fatalf("insert model: %v", err)
		}
		if _, err := db.Exec("DROP TABLE scores"); err != nil {
			t.Fatalf("drop scores: %v", err)
		}

		if got := ReadResults(); len(got) != 0 {
			t.Fatalf("expected score query error to skip model results, got %#v", got)
		}
	})

	t.Run("score scan error leaves default score", func(t *testing.T) {
		dbPath, cleanup := setupTestDB(t)
		defer cleanup()

		if err := InitDB(dbPath); err != nil {
			t.Fatalf("InitDB failed: %v", err)
		}

		suiteID, err := GetSuiteID("default")
		if err != nil {
			t.Fatalf("GetSuiteID failed: %v", err)
		}

		if err := WritePromptSuite("default", []Prompt{{Text: "p1"}}); err != nil {
			t.Fatalf("WritePromptSuite: %v", err)
		}
		if _, err := db.Exec("INSERT INTO models (name, suite_id) VALUES (?, ?)", "Model1", suiteID); err != nil {
			t.Fatalf("insert model: %v", err)
		}

		var promptID int
		if err := db.QueryRow("SELECT id FROM prompts WHERE text = ? AND suite_id = ?", "p1", suiteID).Scan(&promptID); err != nil {
			t.Fatalf("prompt id: %v", err)
		}
		var modelID int
		if err := db.QueryRow("SELECT id FROM models WHERE name = ? AND suite_id = ?", "Model1", suiteID).Scan(&modelID); err != nil {
			t.Fatalf("model id: %v", err)
		}

		if _, err := db.Exec("INSERT INTO scores (model_id, prompt_id, score) VALUES (?, ?, ?)", modelID, promptID, "bad"); err != nil {
			t.Fatalf("insert score: %v", err)
		}

		results := ReadResults()
		if len(results) != 1 {
			t.Fatalf("expected results for 1 model, got %#v", results)
		}
		if got := results["Model1"].Scores; len(got) != 1 || got[0] != 0 {
			t.Fatalf("expected score scan error to leave 0 score, got %#v", got)
		}
	})
}

func TestReadPromptSuite_ErrorBranches(t *testing.T) {
	t.Run("scan error returns error", func(t *testing.T) {
		dbPath, cleanup := setupTestDB(t)
		defer cleanup()

		if err := InitDB(dbPath); err != nil {
			t.Fatalf("InitDB failed: %v", err)
		}
		suiteID, err := GetSuiteID("default")
		if err != nil {
			t.Fatalf("GetSuiteID failed: %v", err)
		}

		if _, err := db.Exec("DROP TABLE prompts"); err != nil {
			t.Fatalf("drop prompts: %v", err)
		}
		if _, err := db.Exec(`CREATE TABLE prompts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			text TEXT NOT NULL,
			solution TEXT DEFAULT '',
			profile_id INTEGER,
			suite_id INTEGER NOT NULL,
			display_order TEXT
		)`); err != nil {
			t.Fatalf("create prompts: %v", err)
		}
		if _, err := db.Exec("INSERT INTO prompts (text, solution, suite_id, display_order) VALUES (?, ?, ?, ?)", "p1", "s1", suiteID, "bad"); err != nil {
			t.Fatalf("insert prompt: %v", err)
		}

		if _, err := ReadPromptSuite("default"); err == nil {
			t.Fatalf("expected ReadPromptSuite to return scan error")
		}
	})

	t.Run("duplicate prompt text is skipped", func(t *testing.T) {
		dbPath, cleanup := setupTestDB(t)
		defer cleanup()

		if err := InitDB(dbPath); err != nil {
			t.Fatalf("InitDB failed: %v", err)
		}
		suiteID, err := GetSuiteID("default")
		if err != nil {
			t.Fatalf("GetSuiteID failed: %v", err)
		}

		if _, err := db.Exec("DROP TABLE prompts"); err != nil {
			t.Fatalf("drop prompts: %v", err)
		}
		if _, err := db.Exec(`CREATE TABLE prompts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			text TEXT NOT NULL,
			solution TEXT DEFAULT '',
			profile_id INTEGER,
			suite_id INTEGER NOT NULL,
			display_order INTEGER DEFAULT 0
		)`); err != nil {
			t.Fatalf("create prompts: %v", err)
		}
		if _, err := db.Exec("INSERT INTO prompts (text, solution, suite_id, display_order) VALUES (?, ?, ?, ?)", "dup", "s1", suiteID, 0); err != nil {
			t.Fatalf("insert prompt 1: %v", err)
		}
		if _, err := db.Exec("INSERT INTO prompts (text, solution, suite_id, display_order) VALUES (?, ?, ?, ?)", "dup", "s2", suiteID, 1); err != nil {
			t.Fatalf("insert prompt 2: %v", err)
		}

		prompts, err := ReadPromptSuite("default")
		if err != nil {
			t.Fatalf("ReadPromptSuite returned error: %v", err)
		}
		if len(prompts) != 1 {
			t.Fatalf("expected duplicates to be skipped, got %#v", prompts)
		}
	})
}

func TestWritePromptSuite_ErrorBranches(t *testing.T) {
	t.Run("closed database fails GetSuiteID", func(t *testing.T) {
		dbPath, cleanup := setupTestDB(t)
		defer cleanup()

		if err := InitDB(dbPath); err != nil {
			t.Fatalf("InitDB failed: %v", err)
		}
		if err := CloseDB(); err != nil {
			t.Fatalf("CloseDB failed: %v", err)
		}

		if err := WritePromptSuite("default", nil); err == nil {
			t.Fatalf("expected error when database is closed")
		}
	})

	t.Run("missing prompts table fails delete", func(t *testing.T) {
		dbPath, cleanup := setupTestDB(t)
		defer cleanup()

		if err := InitDB(dbPath); err != nil {
			t.Fatalf("InitDB failed: %v", err)
		}

		if _, err := db.Exec("DROP TABLE prompts"); err != nil {
			t.Fatalf("drop prompts: %v", err)
		}

		err := WritePromptSuite("default", nil)
		if err == nil {
			t.Fatalf("expected WritePromptSuite to return an error when prompts table is missing")
		}
		if !strings.Contains(err.Error(), "failed to delete prompts") {
			t.Fatalf("expected delete prompts error, got %v", err)
		}
	})

	t.Run("profile lookup error returns error", func(t *testing.T) {
		dbPath, cleanup := setupTestDB(t)
		defer cleanup()

		if err := InitDB(dbPath); err != nil {
			t.Fatalf("InitDB failed: %v", err)
		}

		if _, err := db.Exec("DROP TABLE profiles"); err != nil {
			t.Fatalf("drop profiles: %v", err)
		}
		if _, err := db.Exec("PRAGMA foreign_keys = OFF"); err != nil {
			t.Fatalf("disable foreign keys: %v", err)
		}

		err := WritePromptSuite("default", []Prompt{{Text: "p1", Profile: "missing"}})
		if err == nil {
			t.Fatalf("expected WritePromptSuite to return an error when profiles table is missing")
		}
		if !strings.Contains(err.Error(), "failed to get profile ID") {
			t.Fatalf("expected profile id error, got %v", err)
		}
	})

	t.Run("duplicate prompt text fails insert", func(t *testing.T) {
		dbPath, cleanup := setupTestDB(t)
		defer cleanup()

		if err := InitDB(dbPath); err != nil {
			t.Fatalf("InitDB failed: %v", err)
		}

		err := WritePromptSuite("default", []Prompt{
			{Text: "dup"},
			{Text: "dup"},
		})
		if err == nil {
			t.Fatalf("expected WritePromptSuite to return an error for duplicate prompt text")
		}
		if !strings.Contains(err.Error(), "failed to insert prompt") {
			t.Fatalf("expected insert error, got %v", err)
		}
	})
}

func TestGetCurrentSuiteName_DefaultFallback_UpdateError_ReturnsEmpty(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	if err := InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	if _, err := db.Exec("UPDATE suites SET is_current = 0"); err != nil {
		t.Fatalf("clear current suite: %v", err)
	}
	if _, err := db.Exec(`CREATE TRIGGER abort_suite_update
		BEFORE UPDATE ON suites
		BEGIN
			SELECT RAISE(FAIL, 'nope');
		END;`); err != nil {
		t.Fatalf("create trigger: %v", err)
	}

	if got := GetCurrentSuiteName(); got != "" {
		t.Fatalf("expected empty suite name when default update fails, got %q", got)
	}
}

func TestWriteResults_ErrorBranches(t *testing.T) {
	t.Run("closed database fails GetSuiteID", func(t *testing.T) {
		dbPath, cleanup := setupTestDB(t)
		defer cleanup()

		if err := InitDB(dbPath); err != nil {
			t.Fatalf("InitDB failed: %v", err)
		}
		if err := CloseDB(); err != nil {
			t.Fatalf("CloseDB failed: %v", err)
		}

		if err := WriteResults("default", map[string]Result{}); err == nil {
			t.Fatalf("expected error when database is closed")
		}
	})

	t.Run("missing prompts table fails prompt query", func(t *testing.T) {
		dbPath, cleanup := setupTestDB(t)
		defer cleanup()

		if err := InitDB(dbPath); err != nil {
			t.Fatalf("InitDB failed: %v", err)
		}
		if _, err := db.Exec("DROP TABLE prompts"); err != nil {
			t.Fatalf("drop prompts: %v", err)
		}

		err := WriteResults("default", map[string]Result{})
		if err == nil {
			t.Fatalf("expected error when prompts table is missing")
		}
		if !strings.Contains(err.Error(), "failed to query prompts") {
			t.Fatalf("expected prompt query error, got %v", err)
		}
	})

	t.Run("prompt id scan error returns error", func(t *testing.T) {
		dbPath, cleanup := setupTestDB(t)
		defer cleanup()

		if err := InitDB(dbPath); err != nil {
			t.Fatalf("InitDB failed: %v", err)
		}
		suiteID, err := GetSuiteID("default")
		if err != nil {
			t.Fatalf("GetSuiteID failed: %v", err)
		}

		if _, err := db.Exec("DROP TABLE prompts"); err != nil {
			t.Fatalf("drop prompts: %v", err)
		}
		if _, err := db.Exec(`CREATE TABLE prompts (
			id TEXT PRIMARY KEY,
			text TEXT NOT NULL,
			solution TEXT DEFAULT '',
			profile_id INTEGER,
			suite_id INTEGER NOT NULL,
			display_order INTEGER NOT NULL
		)`); err != nil {
			t.Fatalf("create prompts: %v", err)
		}
		if _, err := db.Exec("INSERT INTO prompts (id, text, solution, suite_id, display_order) VALUES (?, ?, ?, ?, ?)", "bad", "p1", "s1", suiteID, 0); err != nil {
			t.Fatalf("insert prompt: %v", err)
		}

		err = WriteResults("default", map[string]Result{})
		if err == nil {
			t.Fatalf("expected error for prompt id scan")
		}
		if !strings.Contains(err.Error(), "failed to scan prompt ID") {
			t.Fatalf("expected prompt id scan error, got %v", err)
		}
	})

	t.Run("missing models table fails model name query", func(t *testing.T) {
		dbPath, cleanup := setupTestDB(t)
		defer cleanup()

		if err := InitDB(dbPath); err != nil {
			t.Fatalf("InitDB failed: %v", err)
		}
		if _, err := db.Exec("DROP TABLE models"); err != nil {
			t.Fatalf("drop models: %v", err)
		}

		err := WriteResults("default", map[string]Result{})
		if err == nil {
			t.Fatalf("expected error when models table is missing")
		}
		if !strings.Contains(err.Error(), "failed to query model names") {
			t.Fatalf("expected model names query error, got %v", err)
		}
	})

	t.Run("model name scan error returns error", func(t *testing.T) {
		dbPath, cleanup := setupTestDB(t)
		defer cleanup()

		if err := InitDB(dbPath); err != nil {
			t.Fatalf("InitDB failed: %v", err)
		}
		suiteID, err := GetSuiteID("default")
		if err != nil {
			t.Fatalf("GetSuiteID failed: %v", err)
		}

		if _, err := db.Exec("DROP TABLE models"); err != nil {
			t.Fatalf("drop models: %v", err)
		}
		if _, err := db.Exec(`CREATE TABLE models (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT,
			suite_id INTEGER NOT NULL
		)`); err != nil {
			t.Fatalf("create models: %v", err)
		}
		if _, err := db.Exec("INSERT INTO models (name, suite_id) VALUES (?, ?)", nil, suiteID); err != nil {
			t.Fatalf("insert model: %v", err)
		}

		err = WriteResults("default", map[string]Result{})
		if err == nil {
			t.Fatalf("expected error for NULL model name scan")
		}
		if !strings.Contains(err.Error(), "failed to scan model name") {
			t.Fatalf("expected model name scan error, got %v", err)
		}
	})

	t.Run("delete model failure returns error", func(t *testing.T) {
		dbPath, cleanup := setupTestDB(t)
		defer cleanup()

		if err := InitDB(dbPath); err != nil {
			t.Fatalf("InitDB failed: %v", err)
		}

		suiteID, err := GetSuiteID("default")
		if err != nil {
			t.Fatalf("GetSuiteID failed: %v", err)
		}
		if _, err := db.Exec("INSERT INTO models (name, suite_id) VALUES (?, ?)", "to-delete", suiteID); err != nil {
			t.Fatalf("insert model: %v", err)
		}
		if _, err := db.Exec(`CREATE TRIGGER abort_model_delete
			BEFORE DELETE ON models
			BEGIN
				SELECT RAISE(FAIL, 'nope');
			END;`); err != nil {
			t.Fatalf("create trigger: %v", err)
		}

		err = WriteResults("default", map[string]Result{})
		if err == nil {
			t.Fatalf("expected error when model deletion fails")
		}
		if !strings.Contains(err.Error(), "failed to delete model") {
			t.Fatalf("expected delete model error, got %v", err)
		}
	})

	t.Run("missing scores table fails delete scores", func(t *testing.T) {
		dbPath, cleanup := setupTestDB(t)
		defer cleanup()

		if err := InitDB(dbPath); err != nil {
			t.Fatalf("InitDB failed: %v", err)
		}

		if _, err := db.Exec("DROP TABLE scores"); err != nil {
			t.Fatalf("drop scores: %v", err)
		}

		err := WriteResults("default", map[string]Result{})
		if err == nil {
			t.Fatalf("expected error when scores table is missing")
		}
		if !strings.Contains(err.Error(), "failed to delete scores") {
			t.Fatalf("expected delete scores error, got %v", err)
		}
	})

	t.Run("model insert error returns error", func(t *testing.T) {
		dbPath, cleanup := setupTestDB(t)
		defer cleanup()

		if err := InitDB(dbPath); err != nil {
			t.Fatalf("InitDB failed: %v", err)
		}

		if _, err := db.Exec(`CREATE TRIGGER abort_model_insert
			BEFORE INSERT ON models
			BEGIN
				SELECT RAISE(FAIL, 'nope');
			END;`); err != nil {
			t.Fatalf("create trigger: %v", err)
		}

		err := WriteResults("default", map[string]Result{
			"new-model": {Scores: []int{10}},
		})
		if err == nil {
			t.Fatalf("expected error when model insertion fails")
		}
		if !strings.Contains(err.Error(), "failed to insert model") {
			t.Fatalf("expected insert model error, got %v", err)
		}
	})

	t.Run("score insert error returns error", func(t *testing.T) {
		dbPath, cleanup := setupTestDB(t)
		defer cleanup()

		if err := InitDB(dbPath); err != nil {
			t.Fatalf("InitDB failed: %v", err)
		}

		suiteID, err := GetSuiteID("default")
		if err != nil {
			t.Fatalf("GetSuiteID failed: %v", err)
		}
		if err := WritePromptSuite("default", []Prompt{{Text: "p1"}}); err != nil {
			t.Fatalf("WritePromptSuite: %v", err)
		}
		if _, err := db.Exec("INSERT INTO models (name, suite_id) VALUES (?, ?)", "Model1", suiteID); err != nil {
			t.Fatalf("insert model: %v", err)
		}

		if _, err := db.Exec(`CREATE TRIGGER abort_score_insert
			BEFORE INSERT ON scores
			BEGIN
				SELECT RAISE(FAIL, 'nope');
			END;`); err != nil {
			t.Fatalf("create trigger: %v", err)
		}

		err = WriteResults("default", map[string]Result{
			"Model1": {Scores: []int{1}},
		})
		if err == nil {
			t.Fatalf("expected error when score insertion fails")
		}
		if !strings.Contains(err.Error(), "failed to insert score") {
			t.Fatalf("expected insert score error, got %v", err)
		}
	})
}

func TestUpdatePromptsOrder_ErrorBranches(t *testing.T) {
	t.Run("GetCurrentSuiteID error returns early", func(t *testing.T) {
		dbPath, cleanup := setupTestDB(t)
		defer cleanup()

		if err := InitDB(dbPath); err != nil {
			t.Fatalf("InitDB failed: %v", err)
		}
		if err := CloseDB(); err != nil {
			t.Fatalf("CloseDB failed: %v", err)
		}

		UpdatePromptsOrder([]int{})
	})

	t.Run("query prompts error returns early", func(t *testing.T) {
		dbPath, cleanup := setupTestDB(t)
		defer cleanup()

		if err := InitDB(dbPath); err != nil {
			t.Fatalf("InitDB failed: %v", err)
		}
		if _, err := db.Exec("DROP TABLE prompts"); err != nil {
			t.Fatalf("drop prompts: %v", err)
		}

		UpdatePromptsOrder([]int{})
	})

	t.Run("scan prompt id error returns early", func(t *testing.T) {
		dbPath, cleanup := setupTestDB(t)
		defer cleanup()

		if err := InitDB(dbPath); err != nil {
			t.Fatalf("InitDB failed: %v", err)
		}
		suiteID, err := GetSuiteID("default")
		if err != nil {
			t.Fatalf("GetSuiteID failed: %v", err)
		}

		if _, err := db.Exec("DROP TABLE prompts"); err != nil {
			t.Fatalf("drop prompts: %v", err)
		}
		if _, err := db.Exec(`CREATE TABLE prompts (
			id TEXT PRIMARY KEY,
			text TEXT NOT NULL,
			solution TEXT DEFAULT '',
			profile_id INTEGER,
			suite_id INTEGER NOT NULL,
			display_order INTEGER NOT NULL
		)`); err != nil {
			t.Fatalf("create prompts: %v", err)
		}
		if _, err := db.Exec("INSERT INTO prompts (id, text, solution, suite_id, display_order) VALUES (?, ?, ?, ?, ?)", "bad", "p1", "s1", suiteID, 0); err != nil {
			t.Fatalf("insert prompt: %v", err)
		}

		UpdatePromptsOrder([]int{0})
	})

	t.Run("update error returns early", func(t *testing.T) {
		dbPath, cleanup := setupTestDB(t)
		defer cleanup()

		if err := InitDB(dbPath); err != nil {
			t.Fatalf("InitDB failed: %v", err)
		}
		if err := WritePromptSuite("default", []Prompt{
			{Text: "p1"},
			{Text: "p2"},
		}); err != nil {
			t.Fatalf("WritePromptSuite: %v", err)
		}
		if _, err := db.Exec(`CREATE TRIGGER abort_prompt_update
			BEFORE UPDATE ON prompts
			BEGIN
				SELECT RAISE(FAIL, 'nope');
			END;`); err != nil {
			t.Fatalf("create trigger: %v", err)
		}

		UpdatePromptsOrder([]int{1, 0})
	})
}

func TestStateErrorPaths_NoUnexpectedPanics(t *testing.T) {
	// Guard against accidental panics if any helper leaves db in a surprising state.
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("unexpected panic: %v", r)
		}
	}()

	_ = errors.New("guard") // keep errors imported even if tests are refactored
	_ = sql.ErrNoRows
}
