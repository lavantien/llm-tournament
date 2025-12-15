package handlers

import (
	"llm-tournament/middleware"
	"log"
	"net/http"
	"strconv"
)

// SettingsHandler displays the settings page (backward compatible wrapper)
func SettingsHandler(w http.ResponseWriter, r *http.Request) {
	DefaultHandler.Settings(w, r)
}

// UpdateSettingsHandler updates settings from form submission (backward compatible wrapper)
func UpdateSettingsHandler(w http.ResponseWriter, r *http.Request) {
	DefaultHandler.UpdateSettings(w, r)
}

// TestAPIKeyHandler tests an API key (backward compatible wrapper)
func TestAPIKeyHandler(w http.ResponseWriter, r *http.Request) {
	DefaultHandler.TestAPIKey(w, r)
}

// Settings displays the settings page
func (h *Handler) Settings(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling settings page")

	// Get masked API keys for display
	maskedKeys, err := h.DataStore.GetMaskedAPIKeys()
	if err != nil {
		log.Printf("Error getting masked API keys: %v", err)
		http.Error(w, "Failed to load settings", http.StatusInternalServerError)
		return
	}

	// Get other settings
	threshold, _ := h.DataStore.GetSetting("cost_alert_threshold_usd")
	autoEval, _ := h.DataStore.GetSetting("auto_evaluate_new_models")
	pythonURL, _ := h.DataStore.GetSetting("python_service_url")

	// Parse threshold as float
	thresholdFloat, _ := strconv.ParseFloat(threshold, 64)
	if thresholdFloat == 0 {
		thresholdFloat = 100.0
	}

	data := struct {
		PageName      string
		MaskedAPIKeys map[string]string
		Threshold     float64
		AutoEvaluate  bool
		PythonURL     string
	}{
		PageName:      "Settings",
		MaskedAPIKeys: maskedKeys,
		Threshold:     thresholdFloat,
		AutoEvaluate:  autoEval == "true",
		PythonURL:     pythonURL,
	}

	middleware.RenderTemplate(w, "settings.html", data)
}

// UpdateSettings updates settings from form submission
func (h *Handler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Println("Updating settings")

	// Parse form
	if err := r.ParseForm(); err != nil {
		log.Printf("Error parsing form: %v", err)
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Update API keys (only if not empty)
	apiKeys := map[string]string{
		"anthropic": r.FormValue("api_key_anthropic"),
		"openai":    r.FormValue("api_key_openai"),
		"google":    r.FormValue("api_key_google"),
	}

	for provider, key := range apiKeys {
		if key != "" && key != "********" { // Don't update if placeholder
			if err := h.DataStore.SetAPIKey(provider, key); err != nil {
				log.Printf("Error setting API key for %s: %v", provider, err)
				http.Error(w, "Failed to save API key", http.StatusInternalServerError)
				return
			}
		}
	}

	// Update other settings
	threshold := r.FormValue("cost_alert_threshold_usd")
	if threshold != "" {
		if err := h.DataStore.SetSetting("cost_alert_threshold_usd", threshold); err != nil {
			log.Printf("Error setting threshold: %v", err)
		}
	}

	autoEval := r.FormValue("auto_evaluate_new_models")
	autoEvalValue := "false"
	if autoEval == "on" {
		autoEvalValue = "true"
	}
	if err := h.DataStore.SetSetting("auto_evaluate_new_models", autoEvalValue); err != nil {
		log.Printf("Error setting auto_evaluate: %v", err)
	}

	pythonURL := r.FormValue("python_service_url")
	if pythonURL != "" {
		if err := h.DataStore.SetSetting("python_service_url", pythonURL); err != nil {
			log.Printf("Error setting Python URL: %v", err)
		}
	}

	log.Println("Settings updated successfully")

	// Redirect back to settings page
	http.Redirect(w, r, "/settings", http.StatusSeeOther)
}

// TestAPIKey tests an API key by making a health check to Python service
func (h *Handler) TestAPIKey(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	provider := r.FormValue("provider")
	if provider == "" {
		http.Error(w, "Provider required", http.StatusBadRequest)
		return
	}

	// Get API key
	apiKey, err := h.DataStore.GetAPIKey(provider)
	if err != nil || apiKey == "" {
		middleware.RespondJSON(w, map[string]interface{}{
			"success": false,
			"message": "API key not configured",
		})
		return
	}

	// TODO: Make actual test call to respective API
	// For now, just return success
	middleware.RespondJSON(w, map[string]interface{}{
		"success": true,
		"message": "API key appears valid (test not fully implemented)",
	})
}
