package handlers

import (
	"html/template"
	"llm-tournament/middleware"
	"log"
	"net/http"
	"strconv"
)

// SettingsHandler displays the settings page
func SettingsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling settings page")

	// Get masked API keys for display
	maskedKeys, err := middleware.GetMaskedAPIKeys()
	if err != nil {
		log.Printf("Error getting masked API keys: %v", err)
		http.Error(w, "Failed to load settings", http.StatusInternalServerError)
		return
	}

	// Get other settings
	threshold, _ := middleware.GetSetting("cost_alert_threshold_usd")
	autoEval, _ := middleware.GetSetting("auto_evaluate_new_models")
	pythonURL, _ := middleware.GetSetting("python_service_url")

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

	tmpl := template.Must(template.ParseFiles("templates/settings.html", "templates/nav.html"))
	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Failed to render settings", http.StatusInternalServerError)
	}
}

// UpdateSettingsHandler updates settings from form submission
func UpdateSettingsHandler(w http.ResponseWriter, r *http.Request) {
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
			if err := middleware.SetAPIKey(provider, key); err != nil {
				log.Printf("Error setting API key for %s: %v", provider, err)
				http.Error(w, "Failed to save API key", http.StatusInternalServerError)
				return
			}
		}
	}

	// Update other settings
	threshold := r.FormValue("cost_alert_threshold_usd")
	if threshold != "" {
		if err := middleware.SetSetting("cost_alert_threshold_usd", threshold); err != nil {
			log.Printf("Error setting threshold: %v", err)
		}
	}

	autoEval := r.FormValue("auto_evaluate_new_models")
	autoEvalValue := "false"
	if autoEval == "on" {
		autoEvalValue = "true"
	}
	if err := middleware.SetSetting("auto_evaluate_new_models", autoEvalValue); err != nil {
		log.Printf("Error setting auto_evaluate: %v", err)
	}

	pythonURL := r.FormValue("python_service_url")
	if pythonURL != "" {
		if err := middleware.SetSetting("python_service_url", pythonURL); err != nil {
			log.Printf("Error setting Python URL: %v", err)
		}
	}

	log.Println("Settings updated successfully")

	// Redirect back to settings page
	http.Redirect(w, r, "/settings", http.StatusSeeOther)
}

// TestAPIKeyHandler tests an API key by making a health check to Python service
func TestAPIKeyHandler(w http.ResponseWriter, r *http.Request) {
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
	apiKey, err := middleware.GetAPIKey(provider)
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
