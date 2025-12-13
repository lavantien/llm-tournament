package evaluator

import (
	"database/sql"
	"fmt"
	"log"
)

// Evaluator orchestrates LLM evaluations
type Evaluator struct {
	db            *sql.DB
	litellmClient *LiteLLMClient
	jobQueue      *JobQueue
	judges        []string
}

// NewEvaluator creates a new evaluator instance
func NewEvaluator(db *sql.DB, pythonServiceURL string) *Evaluator {
	evaluator := &Evaluator{
		db:            db,
		litellmClient: NewLiteLLMClient(pythonServiceURL),
		judges:        []string{"claude_opus_4.5", "gpt_5.2", "gemini_3_pro"},
	}

	// Initialize job queue with 3 concurrent workers
	evaluator.jobQueue = NewJobQueue(db, 3, evaluator)

	return evaluator
}

// EvaluateAll evaluates all models against all prompts in a suite
func (e *Evaluator) EvaluateAll(suiteID int) (int, error) {
	// Get prompt and model counts
	var promptCount, modelCount int
	err := e.db.QueryRow("SELECT COUNT(*) FROM prompts WHERE suite_id = ?", suiteID).Scan(&promptCount)
	if err != nil {
		return 0, fmt.Errorf("failed to count prompts: %w", err)
	}

	err = e.db.QueryRow("SELECT COUNT(*) FROM models WHERE suite_id = ?", suiteID).Scan(&modelCount)
	if err != nil {
		return 0, fmt.Errorf("failed to count models: %w", err)
	}

	total := promptCount * modelCount

	// Estimate cost
	estimatedCost := float64(total) * 0.05 // ~$0.05 per evaluation

	// Create job
	job := &EvaluationJob{
		SuiteID:       suiteID,
		JobType:       "all",
		ProgressTotal: total,
		EstimatedCost: estimatedCost,
	}

	if err := e.jobQueue.Enqueue(job); err != nil {
		return 0, fmt.Errorf("failed to enqueue job: %w", err)
	}

	return job.ID, nil
}

// EvaluateModel evaluates one model against all prompts
func (e *Evaluator) EvaluateModel(modelID int) (int, error) {
	// Get suite ID and prompt count
	var suiteID, promptCount int
	err := e.db.QueryRow("SELECT suite_id FROM models WHERE id = ?", modelID).Scan(&suiteID)
	if err != nil {
		return 0, fmt.Errorf("failed to get model suite: %w", err)
	}

	err = e.db.QueryRow("SELECT COUNT(*) FROM prompts WHERE suite_id = ?", suiteID).Scan(&promptCount)
	if err != nil {
		return 0, fmt.Errorf("failed to count prompts: %w", err)
	}

	estimatedCost := float64(promptCount) * 0.05

	job := &EvaluationJob{
		SuiteID:       suiteID,
		JobType:       "model",
		TargetID:      modelID,
		ProgressTotal: promptCount,
		EstimatedCost: estimatedCost,
	}

	if err := e.jobQueue.Enqueue(job); err != nil {
		return 0, fmt.Errorf("failed to enqueue job: %w", err)
	}

	return job.ID, nil
}

// EvaluatePrompt evaluates all models for one prompt
func (e *Evaluator) EvaluatePrompt(promptID int) (int, error) {
	// Get suite ID and model count
	var suiteID, modelCount int
	err := e.db.QueryRow("SELECT suite_id FROM prompts WHERE id = ?", promptID).Scan(&suiteID)
	if err != nil {
		return 0, fmt.Errorf("failed to get prompt suite: %w", err)
	}

	err = e.db.QueryRow("SELECT COUNT(*) FROM models WHERE suite_id = ?", suiteID).Scan(&modelCount)
	if err != nil {
		return 0, fmt.Errorf("failed to count models: %w", err)
	}

	estimatedCost := float64(modelCount) * 0.05

	job := &EvaluationJob{
		SuiteID:       suiteID,
		JobType:       "prompt",
		TargetID:      promptID,
		ProgressTotal: modelCount,
		EstimatedCost: estimatedCost,
	}

	if err := e.jobQueue.Enqueue(job); err != nil {
		return 0, fmt.Errorf("failed to enqueue job: %w", err)
	}

	return job.ID, nil
}

// processJob processes an evaluation job (called by worker)
func (e *Evaluator) processJob(job *EvaluationJob, cancelChan chan bool) error {
	log.Printf("Processing job %d (type: %s)", job.ID, job.JobType)

	switch job.JobType {
	case "all":
		return e.processAllJob(job, cancelChan)
	case "model":
		return e.processModelJob(job, cancelChan)
	case "prompt":
		return e.processPromptJob(job, cancelChan)
	default:
		return fmt.Errorf("unknown job type: %s", job.JobType)
	}
}

// processAllJob evaluates all models × all prompts
func (e *Evaluator) processAllJob(job *EvaluationJob, cancelChan chan bool) error {
	// Get all models
	modelRows, err := e.db.Query("SELECT id FROM models WHERE suite_id = ?", job.SuiteID)
	if err != nil {
		return fmt.Errorf("failed to query models: %w", err)
	}
	defer modelRows.Close()

	var modelIDs []int
	for modelRows.Next() {
		var modelID int
		if err := modelRows.Scan(&modelID); err != nil {
			return err
		}
		modelIDs = append(modelIDs, modelID)
	}

	// Get all prompts
	promptRows, err := e.db.Query("SELECT id FROM prompts WHERE suite_id = ?", job.SuiteID)
	if err != nil {
		return fmt.Errorf("failed to query prompts: %w", err)
	}
	defer promptRows.Close()

	var promptIDs []int
	for promptRows.Next() {
		var promptID int
		if err := promptRows.Scan(&promptID); err != nil {
			return err
		}
		promptIDs = append(promptIDs, promptID)
	}

	// Process each combination
	current := 0
	totalCost := 0.0

	for _, modelID := range modelIDs {
		for _, promptID := range promptIDs {
			// Check for cancellation
			select {
			case <-cancelChan:
				return fmt.Errorf("job cancelled")
			default:
			}

			cost, err := e.evaluateModelPromptPair(job.ID, modelID, promptID)
			if err != nil {
				log.Printf("Failed to evaluate model %d, prompt %d: %v", modelID, promptID, err)
				// Continue with next evaluation
			}

			totalCost += cost
			current++

			// Update progress
			if err := e.jobQueue.UpdateJobProgress(job.ID, current, job.ProgressTotal, totalCost); err != nil {
				log.Printf("Failed to update progress: %v", err)
			}
		}
	}

	return nil
}

// processModelJob evaluates one model × all prompts
func (e *Evaluator) processModelJob(job *EvaluationJob, cancelChan chan bool) error {
	// Get all prompts
	promptRows, err := e.db.Query("SELECT id FROM prompts WHERE suite_id = ?", job.SuiteID)
	if err != nil {
		return fmt.Errorf("failed to query prompts: %w", err)
	}
	defer promptRows.Close()

	var promptIDs []int
	for promptRows.Next() {
		var promptID int
		if err := promptRows.Scan(&promptID); err != nil {
			return err
		}
		promptIDs = append(promptIDs, promptID)
	}

	current := 0
	totalCost := 0.0

	for _, promptID := range promptIDs {
		select {
		case <-cancelChan:
			return fmt.Errorf("job cancelled")
		default:
		}

		cost, err := e.evaluateModelPromptPair(job.ID, job.TargetID, promptID)
		if err != nil {
			log.Printf("Failed to evaluate prompt %d: %v", promptID, err)
		}

		totalCost += cost
		current++

		if err := e.jobQueue.UpdateJobProgress(job.ID, current, job.ProgressTotal, totalCost); err != nil {
			log.Printf("Failed to update progress: %v", err)
		}
	}

	return nil
}

// processPromptJob evaluates all models × one prompt
func (e *Evaluator) processPromptJob(job *EvaluationJob, cancelChan chan bool) error {
	// Get all models
	modelRows, err := e.db.Query("SELECT id FROM models WHERE suite_id = ?", job.SuiteID)
	if err != nil {
		return fmt.Errorf("failed to query models: %w", err)
	}
	defer modelRows.Close()

	var modelIDs []int
	for modelRows.Next() {
		var modelID int
		if err := modelRows.Scan(&modelID); err != nil {
			return err
		}
		modelIDs = append(modelIDs, modelID)
	}

	current := 0
	totalCost := 0.0

	for _, modelID := range modelIDs {
		select {
		case <-cancelChan:
			return fmt.Errorf("job cancelled")
		default:
		}

		cost, err := e.evaluateModelPromptPair(job.ID, modelID, job.TargetID)
		if err != nil {
			log.Printf("Failed to evaluate model %d: %v", modelID, err)
		}

		totalCost += cost
		current++

		if err := e.jobQueue.UpdateJobProgress(job.ID, current, job.ProgressTotal, totalCost); err != nil {
			log.Printf("Failed to update progress: %v", err)
		}
	}

	return nil
}

// evaluateModelPromptPair evaluates a single model-prompt pair
func (e *Evaluator) evaluateModelPromptPair(jobID, modelID, promptID int) (float64, error) {
	// Get prompt data
	var promptText, solution, promptType string
	var solutionNull sql.NullString
	err := e.db.QueryRow(`
		SELECT text, solution, type
		FROM prompts
		WHERE id = ?
	`, promptID).Scan(&promptText, &solutionNull, &promptType)

	if err != nil {
		return 0, fmt.Errorf("failed to get prompt: %w", err)
	}

	if solutionNull.Valid {
		solution = solutionNull.String
	}

	// Get model response (from model_responses table or generate placeholder)
	var response string
	err = e.db.QueryRow(`
		SELECT response_text
		FROM model_responses
		WHERE model_id = ? AND prompt_id = ?
	`, modelID, promptID).Scan(&response)

	if err == sql.ErrNoRows {
		// No response stored - skip evaluation
		log.Printf("No response for model %d, prompt %d - skipping", modelID, promptID)
		return 0, nil
	} else if err != nil {
		return 0, fmt.Errorf("failed to get model response: %w", err)
	}

	// Get API keys from settings
	apiKeys, err := e.getAPIKeys()
	if err != nil {
		return 0, fmt.Errorf("failed to get API keys: %w", err)
	}

	// Call Python evaluation service
	evalReq := EvaluationRequest{
		Prompt:   promptText,
		Response: response,
		Solution: solution,
		Type:     promptType,
		Judges:   e.judges,
		APIKeys:  apiKeys,
	}

	evalResp, err := e.litellmClient.Evaluate(evalReq)
	if err != nil {
		return 0, fmt.Errorf("evaluation failed: %w", err)
	}

	// Calculate consensus score
	consensusScore := RoundToValidScore(evalResp.ConsensusScore)

	// Update score in database
	_, err = e.db.Exec(`
		INSERT OR REPLACE INTO scores (model_id, prompt_id, score)
		VALUES (?, ?, ?)
	`, modelID, promptID, consensusScore)

	if err != nil {
		return 0, fmt.Errorf("failed to update score: %w", err)
	}

	// Save evaluation history
	for _, result := range evalResp.Results {
		_, err = e.db.Exec(`
			INSERT INTO evaluation_history (job_id, model_id, prompt_id, judge_name, judge_score, judge_confidence, judge_reasoning, cost_usd)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		`, jobID, modelID, promptID, result.Judge, result.Score, result.Confidence, result.Reasoning, result.CostUSD)

		if err != nil {
			log.Printf("Failed to save evaluation history: %v", err)
		}
	}

	return evalResp.TotalCostUSD, nil
}

// getAPIKeys retrieves API keys from settings
func (e *Evaluator) getAPIKeys() (map[string]string, error) {
	rows, err := e.db.Query(`
		SELECT key, value
		FROM settings
		WHERE key LIKE 'api_key_%'
	`)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	apiKeys := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}
		// TODO: Decrypt value using encryption module (Phase 4)
		apiKeys[key] = value
	}

	return apiKeys, nil
}

// GetJobStatus retrieves job status
func (e *Evaluator) GetJobStatus(jobID int) (*EvaluationJob, error) {
	return e.jobQueue.GetJob(jobID)
}

// CancelJob cancels a running job
func (e *Evaluator) CancelJob(jobID int) error {
	return e.jobQueue.CancelJob(jobID)
}
