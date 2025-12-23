package evaluator

import (
	"io"
	"log"
	"strings"
	"testing"
)

func TestProcessModelJob_QueryPromptsError_ReturnsError(t *testing.T) {
	db := setupEvaluatorTestDB(t)
	defer func() { _ = db.Close() }()

	if _, err := db.Exec("DROP TABLE prompts"); err != nil {
		t.Fatalf("drop prompts: %v", err)
	}

	e := &Evaluator{db: db}
	job := &EvaluationJob{ID: 1, SuiteID: 1, JobType: "model", TargetID: 1, ProgressTotal: 0}

	err := e.processModelJob(job, make(chan bool))
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "failed to query prompts") {
		t.Fatalf("expected query prompts error, got %v", err)
	}
}

func TestProcessPromptJob_QueryModelsError_ReturnsError(t *testing.T) {
	db := setupEvaluatorTestDB(t)
	defer func() { _ = db.Close() }()

	if _, err := db.Exec("DROP TABLE models"); err != nil {
		t.Fatalf("drop models: %v", err)
	}

	e := &Evaluator{db: db}
	job := &EvaluationJob{ID: 1, SuiteID: 1, JobType: "prompt", TargetID: 1, ProgressTotal: 0}

	err := e.processPromptJob(job, make(chan bool))
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "failed to query models") {
		t.Fatalf("expected query models error, got %v", err)
	}
}

func TestProcessModelJob_EvaluateAndProgressErrors_Continue(t *testing.T) {
	db := setupEvaluatorTestDB(t)
	defer func() { _ = db.Close() }()

	origLogOutput := log.Writer()
	log.SetOutput(io.Discard)
	t.Cleanup(func() { log.SetOutput(origLogOutput) })

	if _, err := db.Exec("INSERT INTO prompts (text, suite_id, display_order) VALUES ('p1', 1, 0)"); err != nil {
		t.Fatalf("insert prompt: %v", err)
	}

	if _, err := db.Exec("DROP TABLE model_responses"); err != nil {
		t.Fatalf("drop model_responses: %v", err)
	}
	if _, err := db.Exec("DROP TABLE evaluation_jobs"); err != nil {
		t.Fatalf("drop evaluation_jobs: %v", err)
	}

	e := &Evaluator{
		db: db,
		jobQueue: &JobQueue{
			db: db,
		},
	}

	job := &EvaluationJob{ID: 1, SuiteID: 1, JobType: "model", TargetID: 1, ProgressTotal: 1}

	if err := e.processModelJob(job, make(chan bool)); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestProcessPromptJob_EvaluateAndProgressErrors_Continue(t *testing.T) {
	db := setupEvaluatorTestDB(t)
	defer func() { _ = db.Close() }()

	origLogOutput := log.Writer()
	log.SetOutput(io.Discard)
	t.Cleanup(func() { log.SetOutput(origLogOutput) })

	if _, err := db.Exec("INSERT INTO prompts (text, suite_id, display_order) VALUES ('p1', 1, 0)"); err != nil {
		t.Fatalf("insert prompt: %v", err)
	}
	if _, err := db.Exec("INSERT INTO models (name, suite_id) VALUES ('m1', 1)"); err != nil {
		t.Fatalf("insert model: %v", err)
	}

	if _, err := db.Exec("DROP TABLE model_responses"); err != nil {
		t.Fatalf("drop model_responses: %v", err)
	}
	if _, err := db.Exec("DROP TABLE evaluation_jobs"); err != nil {
		t.Fatalf("drop evaluation_jobs: %v", err)
	}

	e := &Evaluator{
		db: db,
		jobQueue: &JobQueue{
			db: db,
		},
	}

	job := &EvaluationJob{ID: 1, SuiteID: 1, JobType: "prompt", TargetID: 1, ProgressTotal: 1}

	if err := e.processPromptJob(job, make(chan bool)); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
