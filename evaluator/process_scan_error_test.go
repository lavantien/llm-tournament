package evaluator

import "testing"

func TestProcessModelJob_PromptScanError_ReturnsError(t *testing.T) {
	db := setupEvaluatorTestDB(t)
	defer db.Close()

	if _, err := db.Exec("PRAGMA foreign_keys = OFF"); err != nil {
		t.Fatalf("disable foreign keys: %v", err)
	}
	if _, err := db.Exec("DROP TABLE prompts"); err != nil {
		t.Fatalf("drop prompts: %v", err)
	}
	if _, err := db.Exec("CREATE TABLE prompts (id TEXT PRIMARY KEY, suite_id INTEGER NOT NULL)"); err != nil {
		t.Fatalf("create prompts: %v", err)
	}
	if _, err := db.Exec("INSERT INTO prompts (id, suite_id) VALUES ('bad', 1)"); err != nil {
		t.Fatalf("insert prompt: %v", err)
	}

	e := &Evaluator{db: db}
	job := &EvaluationJob{ID: 1, SuiteID: 1, JobType: "model", TargetID: 1}

	if err := e.processModelJob(job, make(chan bool)); err == nil {
		t.Fatalf("expected scan error")
	}
}

func TestProcessPromptJob_ModelScanError_ReturnsError(t *testing.T) {
	db := setupEvaluatorTestDB(t)
	defer db.Close()

	if _, err := db.Exec("PRAGMA foreign_keys = OFF"); err != nil {
		t.Fatalf("disable foreign keys: %v", err)
	}
	if _, err := db.Exec("DROP TABLE models"); err != nil {
		t.Fatalf("drop models: %v", err)
	}
	if _, err := db.Exec("CREATE TABLE models (id TEXT PRIMARY KEY, suite_id INTEGER NOT NULL)"); err != nil {
		t.Fatalf("create models: %v", err)
	}
	if _, err := db.Exec("INSERT INTO models (id, suite_id) VALUES ('bad', 1)"); err != nil {
		t.Fatalf("insert model: %v", err)
	}

	e := &Evaluator{db: db}
	job := &EvaluationJob{ID: 1, SuiteID: 1, JobType: "prompt", TargetID: 1}

	if err := e.processPromptJob(job, make(chan bool)); err == nil {
		t.Fatalf("expected scan error")
	}
}

func TestProcessAllJob_ModelScanError_ReturnsError(t *testing.T) {
	db := setupEvaluatorTestDB(t)
	defer db.Close()

	if _, err := db.Exec("PRAGMA foreign_keys = OFF"); err != nil {
		t.Fatalf("disable foreign keys: %v", err)
	}
	if _, err := db.Exec("DROP TABLE models"); err != nil {
		t.Fatalf("drop models: %v", err)
	}
	if _, err := db.Exec("CREATE TABLE models (id TEXT PRIMARY KEY, suite_id INTEGER NOT NULL)"); err != nil {
		t.Fatalf("create models: %v", err)
	}
	if _, err := db.Exec("INSERT INTO models (id, suite_id) VALUES ('bad', 1)"); err != nil {
		t.Fatalf("insert model: %v", err)
	}

	e := &Evaluator{db: db}
	job := &EvaluationJob{ID: 1, SuiteID: 1, JobType: "all", ProgressTotal: 0}

	if err := e.processAllJob(job, make(chan bool)); err == nil {
		t.Fatalf("expected scan error")
	}
}

func TestProcessAllJob_PromptScanError_ReturnsError(t *testing.T) {
	db := setupEvaluatorTestDB(t)
	defer db.Close()

	if _, err := db.Exec("INSERT INTO models (name, suite_id) VALUES ('m1', 1)"); err != nil {
		t.Fatalf("insert model: %v", err)
	}

	if _, err := db.Exec("PRAGMA foreign_keys = OFF"); err != nil {
		t.Fatalf("disable foreign keys: %v", err)
	}
	if _, err := db.Exec("DROP TABLE prompts"); err != nil {
		t.Fatalf("drop prompts: %v", err)
	}
	if _, err := db.Exec("CREATE TABLE prompts (id TEXT PRIMARY KEY, suite_id INTEGER NOT NULL)"); err != nil {
		t.Fatalf("create prompts: %v", err)
	}
	if _, err := db.Exec("INSERT INTO prompts (id, suite_id) VALUES ('bad', 1)"); err != nil {
		t.Fatalf("insert prompt: %v", err)
	}

	e := &Evaluator{db: db}
	job := &EvaluationJob{ID: 1, SuiteID: 1, JobType: "all", ProgressTotal: 0}

	if err := e.processAllJob(job, make(chan bool)); err == nil {
		t.Fatalf("expected scan error")
	}
}
