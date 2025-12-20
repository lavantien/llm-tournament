package evaluator

import (
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"
)

// JobQueue manages evaluation jobs with concurrent workers
type JobQueue struct {
	db          *sql.DB
	jobs        chan *EvaluationJob
	workers     int
	mu          sync.Mutex
	running     map[int]bool // Track running jobs
	cancel      map[int]chan bool
	evaluator   *Evaluator
	resumeDelay time.Duration // Delay before resuming pending jobs (configurable for testing)
}

var lastInsertID = func(result sql.Result) (int64, error) { return result.LastInsertId() }

// NewJobQueue creates a new job queue with the specified number of workers
func NewJobQueue(db *sql.DB, workers int, evaluator *Evaluator) *JobQueue {
	return NewJobQueueWithDelay(db, workers, evaluator, 5*time.Second)
}

// NewJobQueueWithDelay creates a new job queue with a configurable resume delay (for testing)
func NewJobQueueWithDelay(db *sql.DB, workers int, evaluator *Evaluator, resumeDelay time.Duration) *JobQueue {
	jq := &JobQueue{
		db:          db,
		jobs:        make(chan *EvaluationJob, 100),
		workers:     workers,
		running:     make(map[int]bool),
		cancel:      make(map[int]chan bool),
		evaluator:   evaluator,
		resumeDelay: resumeDelay,
	}

	// Start worker goroutines
	for i := 0; i < workers; i++ {
		go jq.worker(i)
	}

	// Resume pending jobs on startup
	go jq.resumePendingJobs()

	return jq
}

// worker processes jobs from the queue
func (jq *JobQueue) worker(id int) {
	log.Printf("Worker %d started", id)
	for job := range jq.jobs {
		log.Printf("Worker %d processing job %d", id, job.ID)

		// Mark job as running
		jq.mu.Lock()
		jq.running[job.ID] = true
		cancelChan := make(chan bool, 1)
		jq.cancel[job.ID] = cancelChan
		jq.mu.Unlock()

		// Update job status to running
		now := time.Now()
		job.StartedAt = &now
		job.Status = "running"
		if err := jq.updateJob(job); err != nil {
			log.Printf("Failed to update job status: %v", err)
		}

		// Process the job
		err := jq.evaluator.processJob(job, cancelChan)

		// Update job status
		completedAt := time.Now()
		job.CompletedAt = &completedAt

		if err != nil {
			job.Status = "failed"
			job.ErrorMessage = err.Error()
			log.Printf("Job %d failed: %v", job.ID, err)
		} else {
			job.Status = "completed"
			log.Printf("Job %d completed successfully", job.ID)
		}

		if err := jq.updateJob(job); err != nil {
			log.Printf("Failed to update job completion: %v", err)
		}

		// Clean up
		jq.mu.Lock()
		delete(jq.running, job.ID)
		delete(jq.cancel, job.ID)
		jq.mu.Unlock()
	}
}

// Enqueue adds a job to the queue
func (jq *JobQueue) Enqueue(job *EvaluationJob) error {
	// Insert job into database
	result, err := jq.db.Exec(`
		INSERT INTO evaluation_jobs (suite_id, job_type, target_id, status, progress_total, estimated_cost_usd)
		VALUES (?, ?, ?, 'pending', ?, ?)
	`, job.SuiteID, job.JobType, job.TargetID, job.ProgressTotal, job.EstimatedCost)

	if err != nil {
		return fmt.Errorf("failed to insert job: %w", err)
	}

	jobID, err := lastInsertID(result)
	if err != nil {
		return fmt.Errorf("failed to get job ID: %w", err)
	}

	job.ID = int(jobID)
	job.Status = "pending"
	job.CreatedAt = time.Now()

	// Add to queue
	jq.jobs <- job

	return nil
}

// CancelJob cancels a running job
func (jq *JobQueue) CancelJob(jobID int) error {
	jq.mu.Lock()
	defer jq.mu.Unlock()

	if cancelChan, ok := jq.cancel[jobID]; ok {
		cancelChan <- true
		_, err := jq.db.Exec(`UPDATE evaluation_jobs SET status = 'cancelled' WHERE id = ?`, jobID)
		return err
	}

	return fmt.Errorf("job %d not running", jobID)
}

// GetJob retrieves a job by ID
func (jq *JobQueue) GetJob(jobID int) (*EvaluationJob, error) {
	job := &EvaluationJob{}
	var startedAt, completedAt sql.NullTime

	err := jq.db.QueryRow(`
		SELECT id, suite_id, job_type, target_id, status, progress_current, progress_total,
		       estimated_cost_usd, actual_cost_usd, error_message, created_at, started_at, completed_at
		FROM evaluation_jobs
		WHERE id = ?
	`, jobID).Scan(
		&job.ID, &job.SuiteID, &job.JobType, &job.TargetID, &job.Status,
		&job.ProgressCurrent, &job.ProgressTotal, &job.EstimatedCost, &job.ActualCost,
		&job.ErrorMessage, &job.CreatedAt, &startedAt, &completedAt,
	)

	if err != nil {
		return nil, err
	}

	if startedAt.Valid {
		job.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		job.CompletedAt = &completedAt.Time
	}

	return job, nil
}

// UpdateJobProgress updates job progress
func (jq *JobQueue) UpdateJobProgress(jobID, current, total int, cost float64) error {
	_, err := jq.db.Exec(`
		UPDATE evaluation_jobs
		SET progress_current = ?, progress_total = ?, actual_cost_usd = ?
		WHERE id = ?
	`, current, total, cost, jobID)

	return err
}

// updateJob updates a job in the database
func (jq *JobQueue) updateJob(job *EvaluationJob) error {
	_, err := jq.db.Exec(`
		UPDATE evaluation_jobs
		SET status = ?, progress_current = ?, progress_total = ?,
		    estimated_cost_usd = ?, actual_cost_usd = ?, error_message = ?,
		    started_at = ?, completed_at = ?
		WHERE id = ?
	`, job.Status, job.ProgressCurrent, job.ProgressTotal,
		job.EstimatedCost, job.ActualCost, job.ErrorMessage,
		job.StartedAt, job.CompletedAt, job.ID)

	return err
}

// resumePendingJobs resumes jobs that were interrupted
func (jq *JobQueue) resumePendingJobs() {
	if jq.resumeDelay > 0 {
		time.Sleep(jq.resumeDelay) // Wait for initialization
	}

	rows, err := jq.db.Query(`
		SELECT id, suite_id, job_type, target_id, progress_total, estimated_cost_usd
		FROM evaluation_jobs
		WHERE status = 'pending' OR status = 'running'
		ORDER BY created_at
	`)

	if err != nil {
		log.Printf("Failed to query pending jobs: %v", err)
		return
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		job := &EvaluationJob{}
		if err := rows.Scan(&job.ID, &job.SuiteID, &job.JobType, &job.TargetID,
			&job.ProgressTotal, &job.EstimatedCost); err != nil {
			log.Printf("Failed to scan job: %v", err)
			continue
		}

		job.Status = "pending"
		jq.jobs <- job
		count++
	}

	if count > 0 {
		log.Printf("Resumed %d pending jobs", count)
	}
}
