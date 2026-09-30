package handlers

import (
	"llm-tournament/middleware"
	"llm-tournament/templates"
	"log"
	"net/http"
	"strconv"
)

func EvaluateResult(w http.ResponseWriter, r *http.Request) {
	DefaultHandler.EvaluateResultHandler(w, r)
}

// ExportResultsHandler handles exporting results (backward compatible wrapper)

func (h *Handler) EvaluateResultHandler(w http.ResponseWriter, r *http.Request) {
	model := r.URL.Query().Get("model")
	promptIndexStr := r.URL.Query().Get("prompt")

	// Validate required query parameters
	if model == "" || promptIndexStr == "" {
		http.Redirect(w, r, "/results", http.StatusSeeOther)
		return
	}
	if r.Method == http.MethodPost {
		scoreStr := r.FormValue("score")
		score, err := strconv.Atoi(scoreStr)
		if err != nil {
			http.Error(w, "Invalid score value", http.StatusBadRequest)
			return
		}

		results := h.DataStore.ReadResults()
		if results == nil {
			results = make(map[string]middleware.Result)
		}

		result, exists := results[model]
		if !exists {
			// Initialize new result with scores array matching prompts length
			prompts := h.DataStore.ReadPrompts()
			result = middleware.Result{
				Scores: make([]int, len(prompts)),
			}
		}

		index, err := strconv.Atoi(promptIndexStr)
		if err != nil || index < 0 || index >= len(result.Scores) {
			http.Error(w, "Invalid prompt index", http.StatusBadRequest)
			return
		}

		// Update the score (ensure it's within 0-100 range)
		if score < 0 {
			score = 0
		} else if score > 100 {
			score = 100
		}
		result.Scores[index] = score
		results[model] = result

		// Write updated results
		err = h.DataStore.WriteResults(h.DataStore.GetCurrentSuiteName(), results)
		if err != nil {
			http.Error(w, "Failed to save results", http.StatusInternalServerError)
			return
		}

		// Broadcast updated results to all clients
		h.DataStore.BroadcastResults()

		// Add debug logging
		log.Printf("Updated score for model %s, prompt %d: %d", model, index, score)
		log.Printf("Current results for model %s: %v", model, result.Scores)

		// Redirect back to results page
		http.Redirect(w, r, "/results", http.StatusSeeOther)
		return
	}

	// Get current score for this model/prompt
	results := h.DataStore.ReadResults()
	currentScore := 0
	if result, exists := results[model]; exists {
		if index, err := strconv.Atoi(promptIndexStr); err == nil && index >= 0 && index < len(result.Scores) {
			currentScore = result.Scores[index]
		}
	}

	// Get the prompt text and solution for display
	prompts := h.DataStore.ReadPrompts()
	var promptText, solution string
	promptIndex, err := strconv.Atoi(promptIndexStr)
	if err == nil && promptIndex >= 0 && promptIndex < len(prompts) {
		promptText = prompts[promptIndex].Text
		solution = prompts[promptIndex].Solution
	}

	// Get model response if available
	var modelResponse string
	db := middleware.GetDB()
	var modelID int
	var promptID int

	var suiteID int
	if suiteErr := db.QueryRow("SELECT id FROM suites WHERE is_current = 1").Scan(&suiteID); suiteErr == nil {
		// Get model_id from model name, scoped to the current suite
		err = db.QueryRow("SELECT id FROM models WHERE name = ? AND suite_id = ?", model, suiteID).Scan(&modelID)
		if err == nil {
			// Get prompt_id from the current suite using the 0-based prompt index
			err = db.QueryRow("SELECT id FROM prompts WHERE suite_id = ? ORDER BY display_order LIMIT 1 OFFSET ?", suiteID, promptIndex).Scan(&promptID)
			if err == nil {
				// Get the response for this model/prompt pair
				err = db.QueryRow("SELECT response_text FROM model_responses WHERE model_id = ? AND prompt_id = ?", modelID, promptID).Scan(&modelResponse)
				if err != nil {
					// No response found, leave empty
					modelResponse = ""
				}
			}
		}
	}

	data := struct {
		PageName      string
		Model         string
		PromptIndex   string
		ScoreOptions  map[string]int
		CurrentScore  int
		PromptText    string
		Solution      string
		TotalPrompts  int
		ModelResponse string
		ModelID       int
		PromptID      int
		CurrentPath   string
	}{
		PageName:      templates.PageNameEvaluate,
		Model:         model,
		PromptIndex:   promptIndexStr,
		ScoreOptions:  templates.ScoreOptions,
		CurrentScore:  currentScore,
		PromptText:    promptText,
		Solution:      solution,
		TotalPrompts:  len(prompts),
		ModelResponse: modelResponse,
		ModelID:       modelID,
		PromptID:      promptID,
		CurrentPath:   "/evaluate",
	}

	err = h.Renderer.Render(w, "evaluate.html", templates.FuncMap, data, "templates/evaluate.html", "templates/nav.html")
	if err != nil {
		log.Printf("Error rendering template: %v", err)
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		return
	}
}
