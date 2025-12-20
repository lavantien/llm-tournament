package testutil

import (
	"database/sql"
	"errors"
	"fmt"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

type fatalRecorder struct {
	called bool
	msg    string
}

func withFatalRecorder(t *testing.T) *fatalRecorder {
	t.Helper()
	rec := &fatalRecorder{}
	orig := fatalf
	fatalf = func(_ *testing.T, format string, args ...any) {
		rec.called = true
		rec.msg = fmt.Sprintf(format, args...)
	}
	t.Cleanup(func() { fatalf = orig })
	return rec
}

func TestSetupTestDB_ErrorPaths(t *testing.T) {
	t.Run("open error", func(t *testing.T) {
		rec := withFatalRecorder(t)

		origOpen := sqlOpen
		sqlOpen = func(string, string) (*sql.DB, error) { return nil, errors.New("open failed") }
		t.Cleanup(func() { sqlOpen = origOpen })

		db := SetupTestDB(t)
		if db != nil {
			_ = db.Close()
			t.Fatalf("expected nil db")
		}
		if !rec.called {
			t.Fatalf("expected fatalf to be called")
		}
	})

	t.Run("enable foreign keys error", func(t *testing.T) {
		rec := withFatalRecorder(t)

		realDB, err := sql.Open("sqlite3", ":memory:")
		if err != nil {
			t.Fatalf("open db: %v", err)
		}
		t.Cleanup(func() { _ = realDB.Close() })

		origOpen := sqlOpen
		sqlOpen = func(string, string) (*sql.DB, error) { return realDB, nil }
		t.Cleanup(func() { sqlOpen = origOpen })

		origEnable := enableForeignKeys
		enableForeignKeys = func(*sql.DB) error { return errors.New("pragma failed") }
		t.Cleanup(func() { enableForeignKeys = origEnable })

		db := SetupTestDB(t)
		if db != nil {
			_ = db.Close()
			t.Fatalf("expected nil db")
		}
		if !rec.called {
			t.Fatalf("expected fatalf to be called")
		}
	})

	t.Run("create schema error", func(t *testing.T) {
		rec := withFatalRecorder(t)

		realDB, err := sql.Open("sqlite3", ":memory:")
		if err != nil {
			t.Fatalf("open db: %v", err)
		}
		t.Cleanup(func() { _ = realDB.Close() })

		origOpen := sqlOpen
		sqlOpen = func(string, string) (*sql.DB, error) { return realDB, nil }
		t.Cleanup(func() { sqlOpen = origOpen })

		origCreate := createTestSchemaFunc
		createTestSchemaFunc = func(*sql.DB) error { return errors.New("schema failed") }
		t.Cleanup(func() { createTestSchemaFunc = origCreate })

		db := SetupTestDB(t)
		if db != nil {
			_ = db.Close()
			t.Fatalf("expected nil db")
		}
		if !rec.called {
			t.Fatalf("expected fatalf to be called")
		}
	})
}

func TestDBHelper_ErrorPaths(t *testing.T) {
	t.Run("CreateTestSuite exec error", func(t *testing.T) {
		rec := withFatalRecorder(t)
		db, err := sql.Open("sqlite3", ":memory:")
		if err != nil {
			t.Fatalf("open db: %v", err)
		}
		t.Cleanup(func() { _ = db.Close() })

		if got := CreateTestSuite(t, db, "suite"); got != 0 {
			t.Fatalf("expected 0, got %d", got)
		}
		if !rec.called {
			t.Fatalf("expected fatalf to be called")
		}
	})

	t.Run("CreateTestSuite lastInsertID error", func(t *testing.T) {
		rec := withFatalRecorder(t)
		db := SetupTestDB(t)
		t.Cleanup(func() { _ = db.Close() })

		origLast := lastInsertID
		lastInsertID = func(sql.Result) (int64, error) { return 0, errors.New("last insert failed") }
		t.Cleanup(func() { lastInsertID = origLast })

		if got := CreateTestSuite(t, db, "suite"); got != 0 {
			t.Fatalf("expected 0, got %d", got)
		}
		if !rec.called {
			t.Fatalf("expected fatalf to be called")
		}
	})

	t.Run("CreateTestProfile exec error", func(t *testing.T) {
		rec := withFatalRecorder(t)
		db, err := sql.Open("sqlite3", ":memory:")
		if err != nil {
			t.Fatalf("open db: %v", err)
		}
		t.Cleanup(func() { _ = db.Close() })

		if got := CreateTestProfile(t, db, 1, "p", "d"); got != 0 {
			t.Fatalf("expected 0, got %d", got)
		}
		if !rec.called {
			t.Fatalf("expected fatalf to be called")
		}
	})

	t.Run("CreateTestProfile lastInsertID error", func(t *testing.T) {
		rec := withFatalRecorder(t)
		db := SetupTestDB(t)
		t.Cleanup(func() { _ = db.Close() })

		suiteID := CreateTestSuite(t, db, "suite")

		origLast := lastInsertID
		lastInsertID = func(sql.Result) (int64, error) { return 0, errors.New("last insert failed") }
		t.Cleanup(func() { lastInsertID = origLast })

		if got := CreateTestProfile(t, db, suiteID, "p", "d"); got != 0 {
			t.Fatalf("expected 0, got %d", got)
		}
		if !rec.called {
			t.Fatalf("expected fatalf to be called")
		}
	})

	t.Run("CreateTestPrompt exec error", func(t *testing.T) {
		rec := withFatalRecorder(t)
		db, err := sql.Open("sqlite3", ":memory:")
		if err != nil {
			t.Fatalf("open db: %v", err)
		}
		t.Cleanup(func() { _ = db.Close() })

		if got := CreateTestPrompt(t, db, 1, "p", "s", nil, 0, "objective"); got != 0 {
			t.Fatalf("expected 0, got %d", got)
		}
		if !rec.called {
			t.Fatalf("expected fatalf to be called")
		}
	})

	t.Run("CreateTestPrompt lastInsertID error", func(t *testing.T) {
		rec := withFatalRecorder(t)
		db := SetupTestDB(t)
		t.Cleanup(func() { _ = db.Close() })

		suiteID := CreateTestSuite(t, db, "suite")

		origLast := lastInsertID
		lastInsertID = func(sql.Result) (int64, error) { return 0, errors.New("last insert failed") }
		t.Cleanup(func() { lastInsertID = origLast })

		if got := CreateTestPrompt(t, db, suiteID, "p", "s", nil, 0, "objective"); got != 0 {
			t.Fatalf("expected 0, got %d", got)
		}
		if !rec.called {
			t.Fatalf("expected fatalf to be called")
		}
	})

	t.Run("CreateTestModel exec error", func(t *testing.T) {
		rec := withFatalRecorder(t)
		db, err := sql.Open("sqlite3", ":memory:")
		if err != nil {
			t.Fatalf("open db: %v", err)
		}
		t.Cleanup(func() { _ = db.Close() })

		if got := CreateTestModel(t, db, 1, "m"); got != 0 {
			t.Fatalf("expected 0, got %d", got)
		}
		if !rec.called {
			t.Fatalf("expected fatalf to be called")
		}
	})

	t.Run("CreateTestModel lastInsertID error", func(t *testing.T) {
		rec := withFatalRecorder(t)
		db := SetupTestDB(t)
		t.Cleanup(func() { _ = db.Close() })

		suiteID := CreateTestSuite(t, db, "suite")

		origLast := lastInsertID
		lastInsertID = func(sql.Result) (int64, error) { return 0, errors.New("last insert failed") }
		t.Cleanup(func() { lastInsertID = origLast })

		if got := CreateTestModel(t, db, suiteID, "m"); got != 0 {
			t.Fatalf("expected 0, got %d", got)
		}
		if !rec.called {
			t.Fatalf("expected fatalf to be called")
		}
	})

	t.Run("CreateTestScore exec error", func(t *testing.T) {
		rec := withFatalRecorder(t)
		db, err := sql.Open("sqlite3", ":memory:")
		if err != nil {
			t.Fatalf("open db: %v", err)
		}
		t.Cleanup(func() { _ = db.Close() })

		CreateTestScore(t, db, 1, 1, 5)
		if !rec.called {
			t.Fatalf("expected fatalf to be called")
		}
	})

	t.Run("GetDefaultSuiteID scan error", func(t *testing.T) {
		rec := withFatalRecorder(t)
		db, err := sql.Open("sqlite3", ":memory:")
		if err != nil {
			t.Fatalf("open db: %v", err)
		}
		t.Cleanup(func() { _ = db.Close() })

		if got := GetDefaultSuiteID(t, db); got != 0 {
			t.Fatalf("expected 0, got %d", got)
		}
		if !rec.called {
			t.Fatalf("expected fatalf to be called")
		}
	})

	t.Run("SetCurrentSuite clear error", func(t *testing.T) {
		rec := withFatalRecorder(t)
		db, err := sql.Open("sqlite3", ":memory:")
		if err != nil {
			t.Fatalf("open db: %v", err)
		}
		t.Cleanup(func() { _ = db.Close() })

		SetCurrentSuite(t, db, 1)
		if !rec.called {
			t.Fatalf("expected fatalf to be called")
		}
	})

	t.Run("SetCurrentSuite set error", func(t *testing.T) {
		rec := withFatalRecorder(t)
		db, err := sql.Open("sqlite3", ":memory:")
		if err != nil {
			t.Fatalf("open db: %v", err)
		}
		t.Cleanup(func() { _ = db.Close() })

		if _, err := db.Exec(`CREATE TABLE suites (id INTEGER PRIMARY KEY, name TEXT, is_current INTEGER);`); err != nil {
			t.Fatalf("create suites: %v", err)
		}
		if _, err := db.Exec(`INSERT INTO suites (id, name, is_current) VALUES (1, 'default', 0);`); err != nil {
			t.Fatalf("insert suite: %v", err)
		}
		if _, err := db.Exec(`CREATE TRIGGER abort_set_current_suite
			BEFORE UPDATE ON suites
			WHEN NEW.is_current = 1
			BEGIN
				SELECT RAISE(FAIL, 'nope');
			END;`); err != nil {
			t.Fatalf("create trigger: %v", err)
		}

		SetCurrentSuite(t, db, 1)
		if !rec.called {
			t.Fatalf("expected fatalf to be called")
		}
	})

	t.Run("CreateTestSetting exec error", func(t *testing.T) {
		rec := withFatalRecorder(t)
		db, err := sql.Open("sqlite3", ":memory:")
		if err != nil {
			t.Fatalf("open db: %v", err)
		}
		t.Cleanup(func() { _ = db.Close() })

		CreateTestSetting(t, db, "k", "v")
		if !rec.called {
			t.Fatalf("expected fatalf to be called")
		}
	})

	t.Run("CreateTestEvaluationJob exec error", func(t *testing.T) {
		rec := withFatalRecorder(t)
		db, err := sql.Open("sqlite3", ":memory:")
		if err != nil {
			t.Fatalf("open db: %v", err)
		}
		t.Cleanup(func() { _ = db.Close() })

		if got := CreateTestEvaluationJob(t, db, 1, "job", "pending"); got != 0 {
			t.Fatalf("expected 0, got %d", got)
		}
		if !rec.called {
			t.Fatalf("expected fatalf to be called")
		}
	})

	t.Run("CreateTestEvaluationJob lastInsertID error", func(t *testing.T) {
		rec := withFatalRecorder(t)
		db := SetupTestDB(t)
		t.Cleanup(func() { _ = db.Close() })

		suiteID := CreateTestSuite(t, db, "suite")

		origLast := lastInsertID
		lastInsertID = func(sql.Result) (int64, error) { return 0, errors.New("last insert failed") }
		t.Cleanup(func() { lastInsertID = origLast })

		if got := CreateTestEvaluationJob(t, db, suiteID, "job", "pending"); got != 0 {
			t.Fatalf("expected 0, got %d", got)
		}
		if !rec.called {
			t.Fatalf("expected fatalf to be called")
		}
	})

	t.Run("CreateTestModelResponse exec error", func(t *testing.T) {
		rec := withFatalRecorder(t)
		db, err := sql.Open("sqlite3", ":memory:")
		if err != nil {
			t.Fatalf("open db: %v", err)
		}
		t.Cleanup(func() { _ = db.Close() })

		CreateTestModelResponse(t, db, 1, 1, "hello")
		if !rec.called {
			t.Fatalf("expected fatalf to be called")
		}
	})
}

