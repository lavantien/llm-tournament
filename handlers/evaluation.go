package handlers

import (
	"database/sql"
	"fmt"
	"llm-tournament/evaluator"
	"llm-tournament/middleware"
	"log"
	"net/http"
	"strconv"
)

var globalEvaluator *evaluator.Evaluator

// InitEvaluator initializes the global evaluator instance
func InitEvaluator(db *sql.DB) {
	pythonURL, _ := middleware.GetSetting("python_service_url")
	if pythonURL == "" {
		pythonURL = "http://localhost:8001"
	}
	globalEvaluator = evaluator.NewEvaluator(db, pythonURL)
	log.Printf("Evaluator initialized with Python service URL: %s", pythonURL)
}

// EvaluateAllHandler triggers evaluation of all models × all prompts
func EvaluateAllHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	suiteID, err := middleware.GetCurrentSuiteID()
	if err != nil {
		log.Printf("Error getting current suite: %v", err)
		http.Error(w, "Failed to get current suite", http.StatusInternalServerError)
		return
	}

	jobID, err := globalEvaluator.EvaluateAll(suiteID)
	if err != nil {
		log.Printf("Error starting evaluation: %v", err)
		http.Error(w, fmt.Sprintf("Failed to start evaluation: %v", err), http.StatusInternalServerError)
		return
	}

	middleware.RespondJSON(w, map[string]interface{}{
		"success": true,
		"job_id":  jobID,
		"message": "Evaluation started",
	})
}

// EvaluateModelHandler triggers evaluation of one model × all prompts
func EvaluateModelHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	modelIDStr := r.URL.Query().Get("id")
	if modelIDStr == "" {
		http.Error(w, "Model ID required", http.StatusBadRequest)
		return
	}

	modelID, err := strconv.Atoi(modelIDStr)
	if err != nil {
		http.Error(w, "Invalid model ID", http.StatusBadRequest)
		return
	}

	jobID, err := globalEvaluator.EvaluateModel(modelID)
	if err != nil {
		log.Printf("Error starting model evaluation: %v", err)
		http.Error(w, fmt.Sprintf("Failed to start evaluation: %v", err), http.StatusInternalServerError)
		return
	}

	middleware.RespondJSON(w, map[string]interface{}{
		"success": true,
		"job_id":  jobID,
		"message": "Model evaluation started",
	})
}

// EvaluatePromptHandler triggers evaluation of all models × one prompt
func EvaluatePromptHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	promptIDStr := r.URL.Query().Get("id")
	if promptIDStr == "" {
		http.Error(w, "Prompt ID required", http.StatusBadRequest)
		return
	}

	promptID, err := strconv.Atoi(promptIDStr)
	if err != nil {
		http.Error(w, "Invalid prompt ID", http.StatusBadRequest)
		return
	}

	jobID, err := globalEvaluator.EvaluatePrompt(promptID)
	if err != nil {
		log.Printf("Error starting prompt evaluation: %v", err)
		http.Error(w, fmt.Sprintf("Failed to start evaluation: %v", err), http.StatusInternalServerError)
		return
	}

	middleware.RespondJSON(w, map[string]interface{}{
		"success": true,
		"job_id":  jobID,
		"message": "Prompt evaluation started",
	})
}

// EvaluationProgressHandler returns the status of an evaluation job
func EvaluationProgressHandler(w http.ResponseWriter, r *http.Request) {
	jobIDStr := r.URL.Query().Get("id")
	if jobIDStr == "" {
		http.Error(w, "Job ID required", http.StatusBadRequest)
		return
	}

	jobID, err := strconv.Atoi(jobIDStr)
	if err != nil {
		http.Error(w, "Invalid job ID", http.StatusBadRequest)
		return
	}

	job, err := globalEvaluator.GetJobStatus(jobID)
	if err != nil {
		log.Printf("Error getting job status: %v", err)
		http.Error(w, "Failed to get job status", http.StatusInternalServerError)
		return
	}

	middleware.RespondJSON(w, map[string]interface{}{
		"job_id":            job.ID,
		"status":            job.Status,
		"progress_current":  job.ProgressCurrent,
		"progress_total":    job.ProgressTotal,
		"estimated_cost":    job.EstimatedCost,
		"actual_cost":       job.ActualCost,
		"error":             job.ErrorMessage,
	})
}

// CancelEvaluationHandler cancels a running evaluation job
func CancelEvaluationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	jobIDStr := r.URL.Query().Get("id")
	if jobIDStr == "" {
		http.Error(w, "Job ID required", http.StatusBadRequest)
		return
	}

	jobID, err := strconv.Atoi(jobIDStr)
	if err != nil {
		http.Error(w, "Invalid job ID", http.StatusBadRequest)
		return
	}

	if err := globalEvaluator.CancelJob(jobID); err != nil {
		log.Printf("Error cancelling job: %v", err)
		http.Error(w, fmt.Sprintf("Failed to cancel job: %v", err), http.StatusInternalServerError)
		return
	}

	middleware.RespondJSON(w, map[string]interface{}{
		"success": true,
		"message": "Job cancelled",
	})
}
