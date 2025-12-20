package evaluator

import (
	"database/sql"
	"errors"
	"io"
	"log"
	"strings"
	"testing"
)

func TestJobQueue_Enqueue_InsertError_ReturnsError(t *testing.T) {
	db := setupTestJobQueueDB(t)
	defer db.Close()

	if _, err := db.Exec("DROP TABLE evaluation_jobs"); err != nil {
		t.Fatalf("drop evaluation_jobs: %v", err)
	}

	jq := &JobQueue{
		db:      db,
		jobs:    make(chan *EvaluationJob, 1),
		workers: 0,
		running: make(map[int]bool),
		cancel:  make(map[int]chan bool),
	}

	job := &EvaluationJob{
		SuiteID:       1,
		JobType:       "all",
		ProgressTotal: 1,
		EstimatedCost: 0.05,
	}

	err := jq.Enqueue(job)
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "failed to insert job") {
		t.Fatalf("expected insert error, got %v", err)
	}
}

func TestJobQueue_Enqueue_LastInsertIDError_ReturnsError(t *testing.T) {
	db := setupTestJobQueueDB(t)
	defer db.Close()

	jq := &JobQueue{
		db:      db,
		jobs:    make(chan *EvaluationJob, 1),
		workers: 0,
		running: make(map[int]bool),
		cancel:  make(map[int]chan bool),
	}

	original := lastInsertID
	lastInsertID = func(sql.Result) (int64, error) { return 0, errors.New("boom") }
	t.Cleanup(func() { lastInsertID = original })

	job := &EvaluationJob{
		SuiteID:       1,
		JobType:       "all",
		ProgressTotal: 1,
		EstimatedCost: 0.05,
	}

	err := jq.Enqueue(job)
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "failed to get job ID") {
		t.Fatalf("expected last-insert-id error, got %v", err)
	}
}

func TestWorker_ProcessJobError_SetsFailedAndCleansUp(t *testing.T) {
	db := setupTestJobQueueDB(t)
	defer db.Close()

	if _, err := db.Exec("DROP TABLE evaluation_jobs"); err != nil {
		t.Fatalf("drop evaluation_jobs: %v", err)
	}

	origLogOutput := log.Writer()
	log.SetOutput(io.Discard)
	t.Cleanup(func() { log.SetOutput(origLogOutput) })

	e := &Evaluator{}
	jq := &JobQueue{
		db:        db,
		jobs:      make(chan *EvaluationJob, 1),
		workers:   0,
		running:   make(map[int]bool),
		cancel:    make(map[int]chan bool),
		evaluator: e,
	}

	done := make(chan struct{})
	go func() {
		jq.worker(0)
		close(done)
	}()

	job := &EvaluationJob{ID: 123, SuiteID: 1, JobType: "nope"}
	jq.jobs <- job
	close(jq.jobs)
	<-done

	if job.Status != "failed" {
		t.Fatalf("expected failed status, got %q", job.Status)
	}
	if !strings.Contains(job.ErrorMessage, "unknown job type") {
		t.Fatalf("expected unknown job type error message, got %q", job.ErrorMessage)
	}
	if job.StartedAt == nil {
		t.Fatalf("expected StartedAt to be set")
	}
	if job.CompletedAt == nil {
		t.Fatalf("expected CompletedAt to be set")
	}

	if jq.running[job.ID] {
		t.Fatalf("expected running to be cleared for job %d", job.ID)
	}
	if _, ok := jq.cancel[job.ID]; ok {
		t.Fatalf("expected cancel channel to be cleared for job %d", job.ID)
	}
}

