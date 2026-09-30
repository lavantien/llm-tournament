package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"llm-tournament/middleware"
)

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
