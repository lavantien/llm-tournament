package handlers

import (
	"database/sql"
	"encoding/json"
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
		"job_id":           job.ID,
		"status":           job.Status,
		"progress_current": job.ProgressCurrent,
		"progress_total":   job.ProgressTotal,
		"estimated_cost":   job.EstimatedCost,
		"actual_cost":      job.ActualCost,
		"error":            job.ErrorMessage,
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

// SaveModelResponseHandler saves or updates a model's response for a prompt
func SaveModelResponseHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var reqBody struct {
		ModelID      int    `json:"model_id"`
		PromptID     int    `json:"prompt_id"`
		ResponseText string `json:"response_text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil || reqBody.ModelID == 0 {
		http.Error(w, "model_id is required", http.StatusBadRequest)
		return
	}
	if reqBody.PromptID == 0 {
		http.Error(w, "prompt_id is required", http.StatusBadRequest)
		return
	}
	if reqBody.ResponseText == "" {
		http.Error(w, "response_text is required", http.StatusBadRequest)
		return
	}

	// Get database connection
	db := middleware.GetDB()

	// Insert or update the model response
	query := `
		INSERT INTO model_responses (model_id, prompt_id, response_text, response_source)
		VALUES (?, ?, ?, 'manual')
		ON CONFLICT(model_id, prompt_id) DO UPDATE SET
			response_text = excluded.response_text,
			updated_at = CURRENT_TIMESTAMP
	`
	_, err := db.Exec(query, reqBody.ModelID, reqBody.PromptID, reqBody.ResponseText)
	if err != nil {
		log.Printf("Error saving model response: %v", err)
		http.Error(w, "Failed to save response", http.StatusInternalServerError)
		return
	}

	// Return success
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Response saved successfully",
	}); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}
