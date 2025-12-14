package evaluator

import (
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// setupTestJobQueueDB creates an in-memory database for testing
func setupTestJobQueueDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	// Create schema
	schema := `
		CREATE TABLE IF NOT EXISTS suites (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT UNIQUE NOT NULL,
			is_current BOOLEAN DEFAULT FALSE
		);

		CREATE TABLE IF NOT EXISTS evaluation_jobs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			suite_id INTEGER NOT NULL,
			job_type TEXT NOT NULL,
			target_id INTEGER DEFAULT 0,
			status TEXT NOT NULL DEFAULT 'pending',
			progress_current INTEGER DEFAULT 0,
			progress_total INTEGER DEFAULT 0,
			estimated_cost_usd REAL DEFAULT 0.0,
			actual_cost_usd REAL DEFAULT 0.0,
			error_message TEXT DEFAULT '',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			started_at TIMESTAMP,
			completed_at TIMESTAMP,
			FOREIGN KEY (suite_id) REFERENCES suites(id) ON DELETE CASCADE
		);

		INSERT INTO suites (name, is_current) VALUES ('default', 1);
	`
	_, err = db.Exec(schema)
	if err != nil {
		t.Fatalf("failed to create test schema: %v", err)
	}

	return db
}

func TestNewJobQueue(t *testing.T) {
	db := setupTestJobQueueDB(t)
	defer db.Close()

	// Create a mock evaluator (nil is ok for this test)
	jq := NewJobQueue(db, 1, nil)

	if jq == nil {
		t.Fatal("NewJobQueue returned nil")
	}
	if jq.workers != 1 {
		t.Errorf("expected 1 worker, got %d", jq.workers)
	}
	if jq.db != db {
		t.Error("db not set correctly")
	}
	if jq.running == nil {
		t.Error("running map should be initialized")
	}
	if jq.cancel == nil {
		t.Error("cancel map should be initialized")
	}
}

func TestJobQueue_Enqueue(t *testing.T) {
	db := setupTestJobQueueDB(t)
	defer db.Close()

	// Create job queue without starting workers
	jq := &JobQueue{
		db:      db,
		jobs:    make(chan *EvaluationJob, 100),
		workers: 0,
		running: make(map[int]bool),
		cancel:  make(map[int]chan bool),
	}

	job := &EvaluationJob{
		SuiteID:       1,
		JobType:       "all",
		ProgressTotal: 10,
		EstimatedCost: 0.5,
	}

	err := jq.Enqueue(job)
	if err != nil {
		t.Fatalf("Enqueue failed: %v", err)
	}

	if job.ID == 0 {
		t.Error("job ID should be set after enqueue")
	}
	if job.Status != "pending" {
		t.Errorf("expected status 'pending', got %q", job.Status)
	}

	// Verify job is in database
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM evaluation_jobs WHERE id = ?", job.ID).Scan(&count)
	if err != nil {
		t.Fatalf("failed to query job: %v", err)
	}
	if count != 1 {
		t.Error("job should be in database")
	}
}

func TestJobQueue_GetJob(t *testing.T) {
	db := setupTestJobQueueDB(t)
	defer db.Close()

	jq := &JobQueue{
		db:      db,
		jobs:    make(chan *EvaluationJob, 100),
		workers: 0,
		running: make(map[int]bool),
		cancel:  make(map[int]chan bool),
	}

	// Enqueue a job
	job := &EvaluationJob{
		SuiteID:       1,
		JobType:       "model",
		TargetID:      5,
		ProgressTotal: 20,
		EstimatedCost: 1.0,
	}
	err := jq.Enqueue(job)
	if err != nil {
		t.Fatalf("Enqueue failed: %v", err)
	}

	// Get the job
	retrieved, err := jq.GetJob(job.ID)
	if err != nil {
		t.Fatalf("GetJob failed: %v", err)
	}

	if retrieved.ID != job.ID {
		t.Errorf("expected ID %d, got %d", job.ID, retrieved.ID)
	}
	if retrieved.JobType != "model" {
		t.Errorf("expected job type 'model', got %q", retrieved.JobType)
	}
	if retrieved.TargetID != 5 {
		t.Errorf("expected target ID 5, got %d", retrieved.TargetID)
	}
}

func TestJobQueue_GetJob_NotFound(t *testing.T) {
	db := setupTestJobQueueDB(t)
	defer db.Close()

	jq := &JobQueue{
		db:      db,
		jobs:    make(chan *EvaluationJob, 100),
		workers: 0,
		running: make(map[int]bool),
		cancel:  make(map[int]chan bool),
	}

	_, err := jq.GetJob(999)
	if err == nil {
		t.Error("expected error for non-existent job")
	}
}

func TestJobQueue_CancelJob_NotRunning(t *testing.T) {
	db := setupTestJobQueueDB(t)
	defer db.Close()

	jq := &JobQueue{
		db:      db,
		jobs:    make(chan *EvaluationJob, 100),
		workers: 0,
		running: make(map[int]bool),
		cancel:  make(map[int]chan bool),
	}

	err := jq.CancelJob(999)
	if err == nil {
		t.Error("expected error when cancelling non-running job")
	}
}

func TestJobQueue_CancelJob_Running(t *testing.T) {
	db := setupTestJobQueueDB(t)
	defer db.Close()

	jq := &JobQueue{
		db:      db,
		jobs:    make(chan *EvaluationJob, 100),
		workers: 0,
		running: make(map[int]bool),
		cancel:  make(map[int]chan bool),
	}

	// Enqueue a job first
	job := &EvaluationJob{
		SuiteID: 1,
		JobType: "all",
	}
	err := jq.Enqueue(job)
	if err != nil {
		t.Fatalf("Enqueue failed: %v", err)
	}

	// Simulate job running
	jq.mu.Lock()
	jq.running[job.ID] = true
	jq.cancel[job.ID] = make(chan bool, 1)
	jq.mu.Unlock()

	// Cancel the job
	err = jq.CancelJob(job.ID)
	if err != nil {
		t.Fatalf("CancelJob failed: %v", err)
	}

	// Verify job is cancelled in database
	var status string
	err = db.QueryRow("SELECT status FROM evaluation_jobs WHERE id = ?", job.ID).Scan(&status)
	if err != nil {
		t.Fatalf("failed to query job: %v", err)
	}
	if status != "cancelled" {
		t.Errorf("expected status 'cancelled', got %q", status)
	}
}

func TestJobQueue_UpdateJobProgress(t *testing.T) {
	db := setupTestJobQueueDB(t)
	defer db.Close()

	jq := &JobQueue{
		db:      db,
		jobs:    make(chan *EvaluationJob, 100),
		workers: 0,
		running: make(map[int]bool),
		cancel:  make(map[int]chan bool),
	}

	// Enqueue a job
	job := &EvaluationJob{
		SuiteID:       1,
		JobType:       "all",
		ProgressTotal: 100,
	}
	err := jq.Enqueue(job)
	if err != nil {
		t.Fatalf("Enqueue failed: %v", err)
	}

	// Update progress (jobID, current, total, cost)
	err = jq.UpdateJobProgress(job.ID, 50, 100, 0.25)
	if err != nil {
		t.Fatalf("UpdateJobProgress failed: %v", err)
	}

	// Verify progress was updated
	var progress, total int
	var cost float64
	err = db.QueryRow("SELECT progress_current, progress_total, actual_cost_usd FROM evaluation_jobs WHERE id = ?", job.ID).Scan(&progress, &total, &cost)
	if err != nil {
		t.Fatalf("failed to query job: %v", err)
	}
	if progress != 50 {
		t.Errorf("expected progress 50, got %d", progress)
	}
	if total != 100 {
		t.Errorf("expected total 100, got %d", total)
	}
	if cost != 0.25 {
		t.Errorf("expected cost 0.25, got %f", cost)
	}
}

func TestEvaluationJob_Fields(t *testing.T) {
	now := time.Now()
	job := &EvaluationJob{
		ID:              1,
		SuiteID:         2,
		JobType:         "prompt",
		TargetID:        3,
		Status:          "running",
		ProgressCurrent: 5,
		ProgressTotal:   10,
		EstimatedCost:   0.5,
		ActualCost:      0.3,
		ErrorMessage:    "",
		CreatedAt:       now,
		StartedAt:       &now,
		CompletedAt:     nil,
	}

	if job.ID != 1 {
		t.Errorf("expected ID 1, got %d", job.ID)
	}
	if job.JobType != "prompt" {
		t.Errorf("expected JobType 'prompt', got %q", job.JobType)
	}
	if job.StartedAt == nil {
		t.Error("StartedAt should not be nil")
	}
	if job.CompletedAt != nil {
		t.Error("CompletedAt should be nil")
	}
}

func TestJobQueue_MultipleJobs(t *testing.T) {
	db := setupTestJobQueueDB(t)
	defer db.Close()

	jq := &JobQueue{
		db:      db,
		jobs:    make(chan *EvaluationJob, 100),
		workers: 0,
		running: make(map[int]bool),
		cancel:  make(map[int]chan bool),
	}

	// Enqueue multiple jobs
	jobIDs := []int{}
	for i := 0; i < 5; i++ {
		job := &EvaluationJob{
			SuiteID:       1,
			JobType:       "all",
			ProgressTotal: 10,
		}
		err := jq.Enqueue(job)
		if err != nil {
			t.Fatalf("Enqueue failed: %v", err)
		}
		jobIDs = append(jobIDs, job.ID)
	}

	// Verify all jobs are in database
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM evaluation_jobs").Scan(&count)
	if err != nil {
		t.Fatalf("failed to query jobs: %v", err)
	}
	if count != 5 {
		t.Errorf("expected 5 jobs, got %d", count)
	}

	// Verify each job has unique ID
	uniqueIDs := make(map[int]bool)
	for _, id := range jobIDs {
		if uniqueIDs[id] {
			t.Errorf("duplicate job ID: %d", id)
		}
		uniqueIDs[id] = true
	}
}
