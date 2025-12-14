package evaluator

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// setupEvaluatorTestDB creates a complete test database
func setupEvaluatorTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	schema := `
		CREATE TABLE IF NOT EXISTS suites (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT UNIQUE NOT NULL,
			is_current BOOLEAN DEFAULT FALSE
		);

		CREATE TABLE IF NOT EXISTS profiles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			description TEXT DEFAULT '',
			suite_id INTEGER NOT NULL,
			FOREIGN KEY (suite_id) REFERENCES suites(id) ON DELETE CASCADE,
			UNIQUE(name, suite_id)
		);

		CREATE TABLE IF NOT EXISTS prompts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			text TEXT NOT NULL,
			solution TEXT DEFAULT '',
			profile_id INTEGER,
			suite_id INTEGER NOT NULL,
			display_order INTEGER DEFAULT 0,
			type TEXT DEFAULT 'objective',
			FOREIGN KEY (suite_id) REFERENCES suites(id) ON DELETE CASCADE,
			FOREIGN KEY (profile_id) REFERENCES profiles(id) ON DELETE SET NULL
		);

		CREATE TABLE IF NOT EXISTS models (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			suite_id INTEGER NOT NULL,
			FOREIGN KEY (suite_id) REFERENCES suites(id) ON DELETE CASCADE,
			UNIQUE(name, suite_id)
		);

		CREATE TABLE IF NOT EXISTS scores (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			model_id INTEGER NOT NULL,
			prompt_id INTEGER NOT NULL,
			score INTEGER DEFAULT 0,
			FOREIGN KEY (model_id) REFERENCES models(id) ON DELETE CASCADE,
			FOREIGN KEY (prompt_id) REFERENCES prompts(id) ON DELETE CASCADE,
			UNIQUE(model_id, prompt_id)
		);

		CREATE TABLE IF NOT EXISTS settings (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			key TEXT UNIQUE NOT NULL,
			value TEXT DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS evaluation_jobs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			suite_id INTEGER NOT NULL,
			job_type TEXT NOT NULL,
			target_id INTEGER DEFAULT 0,
			status TEXT DEFAULT 'pending',
			progress_current INTEGER DEFAULT 0,
			progress_total INTEGER DEFAULT 0,
			estimated_cost_usd REAL DEFAULT 0,
			actual_cost_usd REAL DEFAULT 0,
			error_message TEXT DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			started_at DATETIME,
			completed_at DATETIME,
			FOREIGN KEY (suite_id) REFERENCES suites(id) ON DELETE CASCADE
		);

		CREATE TABLE IF NOT EXISTS model_responses (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			model_id INTEGER NOT NULL,
			prompt_id INTEGER NOT NULL,
			response_text TEXT DEFAULT '',
			response_source TEXT DEFAULT '',
			api_config TEXT DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (model_id) REFERENCES models(id) ON DELETE CASCADE,
			FOREIGN KEY (prompt_id) REFERENCES prompts(id) ON DELETE CASCADE
		);

		CREATE TABLE IF NOT EXISTS evaluation_history (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			job_id INTEGER NOT NULL,
			model_id INTEGER NOT NULL,
			prompt_id INTEGER NOT NULL,
			judge_name TEXT NOT NULL,
			judge_score INTEGER DEFAULT 0,
			judge_confidence REAL DEFAULT 0,
			judge_reasoning TEXT DEFAULT '',
			cost_usd REAL DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (job_id) REFERENCES evaluation_jobs(id) ON DELETE CASCADE,
			FOREIGN KEY (model_id) REFERENCES models(id) ON DELETE CASCADE,
			FOREIGN KEY (prompt_id) REFERENCES prompts(id) ON DELETE CASCADE
		);

		INSERT INTO suites (name, is_current) VALUES ('default', 1);
		INSERT INTO settings (key, value) VALUES ('api_key_anthropic', '');
		INSERT INTO settings (key, value) VALUES ('api_key_openai', '');
		INSERT INTO settings (key, value) VALUES ('api_key_google', '');
	`

	_, err = db.Exec(schema)
	if err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	return db
}

func TestNewEvaluator(t *testing.T) {
	db := setupEvaluatorTestDB(t)
	defer db.Close()

	evaluator := NewEvaluator(db, "http://localhost:8001")

	if evaluator == nil {
		t.Fatal("NewEvaluator returned nil")
	}
	if evaluator.db != db {
		t.Error("db not set correctly")
	}
	if evaluator.litellmClient == nil {
		t.Error("litellmClient should not be nil")
	}
	if evaluator.jobQueue == nil {
		t.Error("jobQueue should not be nil")
	}
	if len(evaluator.judges) != 3 {
		t.Errorf("expected 3 judges, got %d", len(evaluator.judges))
	}
}

func TestEvaluator_Judges(t *testing.T) {
	db := setupEvaluatorTestDB(t)
	defer db.Close()

	evaluator := NewEvaluator(db, "http://localhost:8001")

	expectedJudges := []string{"claude_opus_4.5", "gpt_5.2", "gemini_3_pro"}
	for i, judge := range expectedJudges {
		if evaluator.judges[i] != judge {
			t.Errorf("expected judge %q, got %q", judge, evaluator.judges[i])
		}
	}
}

func TestEvaluateAll_NoData(t *testing.T) {
	db := setupEvaluatorTestDB(t)
	defer db.Close()

	// Create evaluator without starting workers
	evaluator := &Evaluator{
		db:            db,
		litellmClient: NewLiteLLMClient("http://localhost:8001"),
		judges:        []string{"claude"},
		jobQueue: &JobQueue{
			db:      db,
			jobs:    make(chan *EvaluationJob, 100),
			workers: 0,
			running: make(map[int]bool),
			cancel:  make(map[int]chan bool),
		},
	}
	evaluator.jobQueue.evaluator = evaluator

	// With no prompts/models, total should be 0
	jobID, err := evaluator.EvaluateAll(1)
	if err != nil {
		t.Fatalf("EvaluateAll failed: %v", err)
	}
	if jobID == 0 {
		t.Error("expected non-zero job ID")
	}

	// Verify job was created
	var status string
	err = db.QueryRow("SELECT status FROM evaluation_jobs WHERE id = ?", jobID).Scan(&status)
	if err != nil {
		t.Fatalf("failed to query job: %v", err)
	}
	if status != "pending" {
		t.Errorf("expected status 'pending', got %q", status)
	}
}

func TestEvaluateAll_WithData(t *testing.T) {
	db := setupEvaluatorTestDB(t)
	defer db.Close()

	// Add test data
	_, err := db.Exec("INSERT INTO prompts (text, suite_id, display_order) VALUES ('prompt1', 1, 0)")
	if err != nil {
		t.Fatalf("failed to insert prompt: %v", err)
	}
	_, err = db.Exec("INSERT INTO prompts (text, suite_id, display_order) VALUES ('prompt2', 1, 1)")
	if err != nil {
		t.Fatalf("failed to insert prompt: %v", err)
	}
	_, err = db.Exec("INSERT INTO models (name, suite_id) VALUES ('model1', 1)")
	if err != nil {
		t.Fatalf("failed to insert model: %v", err)
	}

	evaluator := &Evaluator{
		db:            db,
		litellmClient: NewLiteLLMClient("http://localhost:8001"),
		judges:        []string{"claude"},
		jobQueue: &JobQueue{
			db:      db,
			jobs:    make(chan *EvaluationJob, 100),
			workers: 0,
			running: make(map[int]bool),
			cancel:  make(map[int]chan bool),
		},
	}
	evaluator.jobQueue.evaluator = evaluator

	jobID, err := evaluator.EvaluateAll(1)
	if err != nil {
		t.Fatalf("EvaluateAll failed: %v", err)
	}

	// Verify job has correct totals (2 prompts * 1 model = 2)
	var progressTotal int
	var estimatedCost float64
	err = db.QueryRow("SELECT progress_total, estimated_cost_usd FROM evaluation_jobs WHERE id = ?", jobID).Scan(&progressTotal, &estimatedCost)
	if err != nil {
		t.Fatalf("failed to query job: %v", err)
	}
	if progressTotal != 2 {
		t.Errorf("expected progress_total 2, got %d", progressTotal)
	}
	if estimatedCost != 0.1 { // 2 * 0.05
		t.Errorf("expected estimated_cost 0.1, got %f", estimatedCost)
	}
}

func TestEvaluateModel(t *testing.T) {
	db := setupEvaluatorTestDB(t)
	defer db.Close()

	// Add test data
	_, err := db.Exec("INSERT INTO prompts (text, suite_id, display_order) VALUES ('prompt1', 1, 0)")
	if err != nil {
		t.Fatalf("failed to insert prompt: %v", err)
	}
	_, err = db.Exec("INSERT INTO prompts (text, suite_id, display_order) VALUES ('prompt2', 1, 1)")
	if err != nil {
		t.Fatalf("failed to insert prompt: %v", err)
	}
	result, err := db.Exec("INSERT INTO models (name, suite_id) VALUES ('model1', 1)")
	if err != nil {
		t.Fatalf("failed to insert model: %v", err)
	}
	modelID, _ := result.LastInsertId()

	evaluator := &Evaluator{
		db:            db,
		litellmClient: NewLiteLLMClient("http://localhost:8001"),
		judges:        []string{"claude"},
		jobQueue: &JobQueue{
			db:      db,
			jobs:    make(chan *EvaluationJob, 100),
			workers: 0,
			running: make(map[int]bool),
			cancel:  make(map[int]chan bool),
		},
	}
	evaluator.jobQueue.evaluator = evaluator

	jobID, err := evaluator.EvaluateModel(int(modelID))
	if err != nil {
		t.Fatalf("EvaluateModel failed: %v", err)
	}

	// Verify job has correct type and target
	var jobType string
	var targetID int
	err = db.QueryRow("SELECT job_type, target_id FROM evaluation_jobs WHERE id = ?", jobID).Scan(&jobType, &targetID)
	if err != nil {
		t.Fatalf("failed to query job: %v", err)
	}
	if jobType != "model" {
		t.Errorf("expected job_type 'model', got %q", jobType)
	}
	if targetID != int(modelID) {
		t.Errorf("expected target_id %d, got %d", modelID, targetID)
	}
}

func TestEvaluateModel_NotFound(t *testing.T) {
	db := setupEvaluatorTestDB(t)
	defer db.Close()

	evaluator := &Evaluator{
		db:            db,
		litellmClient: NewLiteLLMClient("http://localhost:8001"),
		judges:        []string{"claude"},
		jobQueue: &JobQueue{
			db:      db,
			jobs:    make(chan *EvaluationJob, 100),
			workers: 0,
			running: make(map[int]bool),
			cancel:  make(map[int]chan bool),
		},
	}
	evaluator.jobQueue.evaluator = evaluator

	_, err := evaluator.EvaluateModel(999)
	if err == nil {
		t.Error("expected error for non-existent model")
	}
}

func TestEvaluatePrompt(t *testing.T) {
	db := setupEvaluatorTestDB(t)
	defer db.Close()

	// Add test data
	result, err := db.Exec("INSERT INTO prompts (text, suite_id, display_order) VALUES ('prompt1', 1, 0)")
	if err != nil {
		t.Fatalf("failed to insert prompt: %v", err)
	}
	promptID, _ := result.LastInsertId()

	_, err = db.Exec("INSERT INTO models (name, suite_id) VALUES ('model1', 1)")
	if err != nil {
		t.Fatalf("failed to insert model: %v", err)
	}
	_, err = db.Exec("INSERT INTO models (name, suite_id) VALUES ('model2', 1)")
	if err != nil {
		t.Fatalf("failed to insert model: %v", err)
	}

	evaluator := &Evaluator{
		db:            db,
		litellmClient: NewLiteLLMClient("http://localhost:8001"),
		judges:        []string{"claude"},
		jobQueue: &JobQueue{
			db:      db,
			jobs:    make(chan *EvaluationJob, 100),
			workers: 0,
			running: make(map[int]bool),
			cancel:  make(map[int]chan bool),
		},
	}
	evaluator.jobQueue.evaluator = evaluator

	jobID, err := evaluator.EvaluatePrompt(int(promptID))
	if err != nil {
		t.Fatalf("EvaluatePrompt failed: %v", err)
	}

	// Verify job has correct type and target
	var jobType string
	var targetID int
	err = db.QueryRow("SELECT job_type, target_id FROM evaluation_jobs WHERE id = ?", jobID).Scan(&jobType, &targetID)
	if err != nil {
		t.Fatalf("failed to query job: %v", err)
	}
	if jobType != "prompt" {
		t.Errorf("expected job_type 'prompt', got %q", jobType)
	}
	if targetID != int(promptID) {
		t.Errorf("expected target_id %d, got %d", promptID, targetID)
	}
}

func TestEvaluatePrompt_NotFound(t *testing.T) {
	db := setupEvaluatorTestDB(t)
	defer db.Close()

	evaluator := &Evaluator{
		db:            db,
		litellmClient: NewLiteLLMClient("http://localhost:8001"),
		judges:        []string{"claude"},
		jobQueue: &JobQueue{
			db:      db,
			jobs:    make(chan *EvaluationJob, 100),
			workers: 0,
			running: make(map[int]bool),
			cancel:  make(map[int]chan bool),
		},
	}
	evaluator.jobQueue.evaluator = evaluator

	_, err := evaluator.EvaluatePrompt(999)
	if err == nil {
		t.Error("expected error for non-existent prompt")
	}
}

func TestGetJobStatus(t *testing.T) {
	db := setupEvaluatorTestDB(t)
	defer db.Close()

	evaluator := &Evaluator{
		db:            db,
		litellmClient: NewLiteLLMClient("http://localhost:8001"),
		judges:        []string{"claude"},
		jobQueue: &JobQueue{
			db:      db,
			jobs:    make(chan *EvaluationJob, 100),
			workers: 0,
			running: make(map[int]bool),
			cancel:  make(map[int]chan bool),
		},
	}
	evaluator.jobQueue.evaluator = evaluator

	// Create a job
	jobID, err := evaluator.EvaluateAll(1)
	if err != nil {
		t.Fatalf("EvaluateAll failed: %v", err)
	}

	// Get status
	job, err := evaluator.GetJobStatus(jobID)
	if err != nil {
		t.Fatalf("GetJobStatus failed: %v", err)
	}

	if job.ID != jobID {
		t.Errorf("expected job ID %d, got %d", jobID, job.ID)
	}
	if job.Status != "pending" {
		t.Errorf("expected status 'pending', got %q", job.Status)
	}
}

func TestCancelJob(t *testing.T) {
	db := setupEvaluatorTestDB(t)
	defer db.Close()

	evaluator := &Evaluator{
		db:            db,
		litellmClient: NewLiteLLMClient("http://localhost:8001"),
		judges:        []string{"claude"},
		jobQueue: &JobQueue{
			db:      db,
			jobs:    make(chan *EvaluationJob, 100),
			workers: 0,
			running: make(map[int]bool),
			cancel:  make(map[int]chan bool),
		},
	}
	evaluator.jobQueue.evaluator = evaluator

	// Create a job
	jobID, err := evaluator.EvaluateAll(1)
	if err != nil {
		t.Fatalf("EvaluateAll failed: %v", err)
	}

	// Try to cancel (will fail because job isn't "running")
	err = evaluator.CancelJob(jobID)
	if err == nil {
		t.Error("expected error when cancelling non-running job")
	}
}
