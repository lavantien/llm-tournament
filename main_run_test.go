package main

import (
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	"llm-tournament/middleware"
)

func TestRun_MigrateResults_UsesMigratedMap(t *testing.T) {
	original := map[string]middleware.Result{
		"model-a": {Scores: nil},
	}
	migrated := map[string]middleware.Result{
		"model-a": {Scores: []int{1, 2, 3}},
	}

	var (
		initDBCalled       bool
		closeDBCalled      bool
		migrateCalled      bool
		writeResultsCalled bool
	)

	deps := runDeps{
		initDB: func(path string) error {
			initDBCalled = true
			if path != "db.sqlite" {
				t.Fatalf("initDB called with %q", path)
			}
			return nil
		},
		closeDB: func() error {
			closeDBCalled = true
			return nil
		},
		readResults: func() map[string]middleware.Result {
			return original
		},
		migrateResults: func(in map[string]middleware.Result) map[string]middleware.Result {
			migrateCalled = true
			if !reflect.DeepEqual(in, original) {
				t.Fatalf("migrateResults called with unexpected input: %#v", in)
			}
			return migrated
		},
		getCurrentSuiteName: func() string {
			return "default"
		},
		writeResults: func(suite string, results map[string]middleware.Result) error {
			writeResultsCalled = true
			if suite != "default" {
				t.Fatalf("writeResults called with suite %q", suite)
			}
			if !reflect.DeepEqual(results, migrated) {
				t.Fatalf("writeResults called with non-migrated results: %#v", results)
			}
			return nil
		},
		initEvaluator: func(*sql.DB) {
			t.Fatalf("initEvaluator should not be called when migrating results")
		},
		getDB: func() *sql.DB {
			return nil
		},
		listenAndServe: func(string, http.Handler) error {
			t.Fatalf("listenAndServe should not be called when migrating results")
			return nil
		},
	}

	exitCode := run([]string{"-db", "db.sqlite", "-migrate-results"}, deps)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}
	if !initDBCalled {
		t.Fatalf("expected initDB to be called")
	}
	if !migrateCalled {
		t.Fatalf("expected migrateResults to be called")
	}
	if !writeResultsCalled {
		t.Fatalf("expected writeResults to be called")
	}
	if !closeDBCalled {
		t.Fatalf("expected closeDB to be called via defer")
	}
}

func TestRun_FlagParseError_Returns2(t *testing.T) {
	deps := runDeps{
		initDB: func(string) error {
			t.Fatalf("initDB should not be called on flag parse error")
			return nil
		},
	}

	if exitCode := run([]string{"-no-such-flag"}, deps); exitCode != 2 {
		t.Fatalf("expected exit code 2, got %d", exitCode)
	}
}

func TestRun_InitDBError_Returns1(t *testing.T) {
	var closeDBCalled bool

	deps := runDeps{
		initDB: func(string) error {
			return errors.New("init db failed")
		},
		closeDB: func() error {
			closeDBCalled = true
			return nil
		},
		listenAndServe: func(string, http.Handler) error {
			t.Fatalf("listenAndServe should not be called when initDB fails")
			return nil
		},
	}

	if exitCode := run([]string{"-db", "db.sqlite"}, deps); exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}
	if closeDBCalled {
		t.Fatalf("expected closeDB to not be called when initDB fails")
	}
}

func TestRun_MigrateResults_WriteResultsError_Returns1(t *testing.T) {
	var closeDBCalled bool

	deps := runDeps{
		initDB: func(string) error {
			return nil
		},
		closeDB: func() error {
			closeDBCalled = true
			return nil
		},
		readResults: func() map[string]middleware.Result {
			return map[string]middleware.Result{}
		},
		migrateResults: func(in map[string]middleware.Result) map[string]middleware.Result {
			return in
		},
		getCurrentSuiteName: func() string { return "default" },
		writeResults: func(string, map[string]middleware.Result) error {
			return errors.New("write failed")
		},
		initEvaluator: func(*sql.DB) {
			t.Fatalf("initEvaluator should not be called when migrating results")
		},
		getDB: func() *sql.DB { return nil },
		listenAndServe: func(string, http.Handler) error {
			t.Fatalf("listenAndServe should not be called when migrating results")
			return nil
		},
	}

	if exitCode := run([]string{"-migrate-results"}, deps); exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}
	if !closeDBCalled {
		t.Fatalf("expected closeDB to be called via defer")
	}
}

func TestRun_ServePath_CallsListenAndServeWithRouter(t *testing.T) {
	var (
		initEvaluatorCalled bool
		closeDBCalled       bool
	)

	deps := runDeps{
		initDB: func(string) error {
			return nil
		},
		closeDB: func() error {
			closeDBCalled = true
			return nil
		},
		initEvaluator: func(*sql.DB) {
			initEvaluatorCalled = true
		},
		getDB: func() *sql.DB { return nil },
		listenAndServe: func(addr string, handler http.Handler) error {
			if addr != ":8080" {
				t.Fatalf("listenAndServe called with addr %q", addr)
			}
			if handler == nil {
				t.Fatalf("listenAndServe called with nil handler")
			}

			req := httptest.NewRequest(http.MethodGet, "/unknown", nil)
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != http.StatusSeeOther {
				t.Fatalf("expected redirect status %d, got %d", http.StatusSeeOther, rr.Code)
			}
			if location := rr.Header().Get("Location"); location != "/prompts" {
				t.Fatalf("expected redirect location %q, got %q", "/prompts", location)
			}

			return nil
		},
	}

	if exitCode := run([]string{"-db", "db.sqlite"}, deps); exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}
	if !initEvaluatorCalled {
		t.Fatalf("expected initEvaluator to be called")
	}
	if !closeDBCalled {
		t.Fatalf("expected closeDB to be called via defer")
	}
}

func TestRun_ServePath_ListenAndServeError_Returns1(t *testing.T) {
	var closeDBCalled bool

	deps := runDeps{
		initDB: func(string) error {
			return nil
		},
		closeDB: func() error {
			closeDBCalled = true
			return nil
		},
		initEvaluator: func(*sql.DB) {},
		getDB:         func() *sql.DB { return nil },
		listenAndServe: func(string, http.Handler) error {
			return errors.New("listen failed")
		},
	}

	if exitCode := run([]string{"-db", "db.sqlite"}, deps); exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}
	if !closeDBCalled {
		t.Fatalf("expected closeDB to be called via defer")
	}
}

func TestDefaultRunDeps_HasRequiredDeps(t *testing.T) {
	deps := defaultRunDeps()

	if deps.initDB == nil ||
		deps.closeDB == nil ||
		deps.readResults == nil ||
		deps.migrateResults == nil ||
		deps.getCurrentSuiteName == nil ||
		deps.writeResults == nil ||
		deps.initEvaluator == nil ||
		deps.getDB == nil ||
		deps.listenAndServe == nil {
		t.Fatalf("expected all default run deps to be non-nil: %#v", deps)
	}
}

func TestMain_UsesOsExit(t *testing.T) {
	origArgs := os.Args
	origExit := osExit
	t.Cleanup(func() {
		os.Args = origArgs
		osExit = origExit
	})

	os.Args = []string{"llm-tournament", "-no-such-flag"}

	var gotCode int
	osExit = func(code int) {
		gotCode = code
	}

	main()

	if gotCode != 2 {
		t.Fatalf("expected exit code 2, got %d", gotCode)
	}
}
