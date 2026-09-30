package handlers

import (
	"database/sql"
	"llm-tournament/middleware"
	"llm-tournament/templates"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

// lastInsertID is swappable for error-path tests, mirroring middleware
var lastInsertID = func(result sql.Result) (int64, error) { return result.LastInsertId() }

// GroupedPrompt represents a prompt with its profile information
type GroupedPrompt struct {
	Index       int
	Text        string
	ProfileID   string
	ProfileName string
}

// ResultsHandler handles the results page (backward compatible wrapper)
func ResultsHandler(w http.ResponseWriter, r *http.Request) {
	DefaultHandler.Results(w, r)
}

// UpdateResultHandler handles updating results (backward compatible wrapper)
func UpdateResultHandler(w http.ResponseWriter, r *http.Request) {
	DefaultHandler.UpdateResult(w, r)
}

// ResetResultsHandler handles resetting results (backward compatible wrapper)
func ResetResultsHandler(w http.ResponseWriter, r *http.Request) {
	DefaultHandler.ResetResults(w, r)
}

// ConfirmRefreshResultsHandler handles confirm refresh results (backward compatible wrapper)
func ConfirmRefreshResultsHandler(w http.ResponseWriter, r *http.Request) {
	DefaultHandler.ConfirmRefreshResults(w, r)
}

func (h *Handler) Results(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling results page")
	prompts := h.DataStore.ReadPrompts()
	results := h.DataStore.ReadResults()

	// Group prompts by profile
	var orderedPrompts []GroupedPrompt

	// Get all profiles first (to include empty ones)
	profiles := h.DataStore.ReadProfiles()

	// Get profile groups using the utility function
	profileGroups, profileMap := middleware.GetProfileGroups(prompts, profiles)

	// Check if we have any uncategorized prompts
	hasUncategorized := false
	for _, prompt := range prompts {
		if prompt.Profile == "" {
			hasUncategorized = true
			break
		}
	}

	// Add a group for prompts with no profile only if needed
	if hasUncategorized {
		noProfileGroup := &middleware.ProfileGroup{
			ID:       "none",
			Name:     "Uncategorized",
			Color:    "hsl(0, 0%, 50%)",
			StartCol: -1,
			EndCol:   -1,
		}
		profileGroups = append(profileGroups, noProfileGroup)
		profileMap[""] = noProfileGroup
	}

	// Process prompts and assign them to profile groups
	currentCol := 0
	for i, prompt := range prompts {
		profileName := prompt.Profile

		group := profileMap[profileName]

		if group.StartCol == -1 {
			group.StartCol = currentCol
		}
		group.EndCol = currentCol

		orderedPrompts = append(orderedPrompts, GroupedPrompt{
			Index:       i,
			Text:        prompt.Text,
			ProfileID:   group.ID,
			ProfileName: profileName,
		})

		currentCol++
	}

	log.Println("Calculating total scores for each model")
	// Calculate total scores for each model
	modelScores := make(map[string]int)
	for model, result := range results {
		totalScore := 0
		for _, score := range result.Scores {
			totalScore += score
		}
		modelScores[model] = totalScore
	}

	log.Println("Sorting models by score in descending order")
	// Sort models by score in descending order
	models := make([]string, 0, len(results))
	for model := range results {
		models = append(models, model)
	}
	sort.Slice(models, func(i, j int) bool {
		return modelScores[models[i]] > modelScores[models[j]]
	})
	log.Printf("Sorted models: %v", models)

	modelFilter := r.FormValue("model_filter")
	searchQuery := strings.ToLower(r.FormValue("search"))

	filteredResults := make(map[string]middleware.Result)
	for model, result := range results {
		// Apply model filter if specified
		if modelFilter != "" && model != modelFilter {
			continue
		}
		// Apply search filter if specified
		if searchQuery != "" && !strings.Contains(strings.ToLower(model), searchQuery) {
			continue
		}
		filteredResults[model] = result
	}

	pageName := templates.PageNameResults
	promptTexts := make([]string, len(prompts))
	for i, prompt := range prompts {
		promptTexts[i] = prompt.Text
	}
	resultsForTemplate := make(map[string]middleware.Result)
	for model, result := range filteredResults {
		// Initialize scores array if nil
		if result.Scores == nil {
			result.Scores = make([]int, len(prompts))
		}

		// Ensure scores array matches prompts length
		if len(result.Scores) != len(prompts) {
			newScores := make([]int, len(prompts))
			copy(newScores, result.Scores)
			result.Scores = newScores
		}

		// Ensure all scores are valid (0-100)
		for i, score := range result.Scores {
			if score < 0 || score > 100 {
				result.Scores[i] = 0
			}
		}

		// Create a new Result struct to ensure proper initialization
		resultsForTemplate[model] = middleware.Result{
			Scores: result.Scores,
		}
	}
	modelPassPercentages := make(map[string]float64)
	modelTotalScores := make(map[string]int)
	promptIndices := make([]int, len(prompts))
	for i := range prompts {
		promptIndices[i] = i + 1
	}
	for model, result := range filteredResults {
		totalScore := 0
		for _, score := range result.Scores {
			totalScore += score
		}
		// Avoid division by zero when there are no prompts
		if len(prompts) > 0 {
			modelPassPercentages[model] = float64(totalScore) / float64(len(prompts)*100) * 100
		} else {
			modelPassPercentages[model] = 0
		}
		modelTotalScores[model] = totalScore
	}

	// Log the data we're about to send to the template for debugging
	if len(models) > 0 && len(promptTexts) > 0 {
		log.Printf("First model: %s, scores: %v", models[0], resultsForTemplate[models[0]].Scores)
	}

	templateData := struct {
		PageName        string
		Prompts         []string
		Results         map[string]middleware.Result
		Models          []string
		PassPercentages map[string]float64
		ModelFilter     string
		TotalScores     map[string]int
		PromptIndices   []int
		SearchQuery     string
		ProfileGroups   []*middleware.ProfileGroup
		OrderedPrompts  []GroupedPrompt
		CurrentPath     string
	}{
		PageName:        pageName,
		Prompts:         promptTexts,
		Results:         resultsForTemplate,
		Models:          models,
		PassPercentages: modelPassPercentages,
		ModelFilter:     modelFilter,
		TotalScores:     modelTotalScores,
		PromptIndices:   promptIndices,
		SearchQuery:     searchQuery,
		ProfileGroups:   profileGroups,
		OrderedPrompts:  orderedPrompts,
		CurrentPath:     "/results",
	}

	err := h.Renderer.Render(w, "results.html", templates.FuncMap, templateData, "templates/results.html", "templates/nav.html")
	if err != nil {
		log.Printf("Error rendering template: %v", err)
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		return
	}
	log.Println("Results page rendered successfully")
}

// UpdateResult handles AJAX requests to update results
func (h *Handler) UpdateResult(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling update result")
	_ = r.ParseForm()
	model := r.Form.Get("model")
	promptIndexStr := r.Form.Get("promptIndex")
	passStr := r.Form.Get("pass")
	promptIndex, _ := strconv.Atoi(promptIndexStr)
	pass, err := strconv.ParseBool(passStr)
	if err != nil {
		log.Printf("Invalid pass value: %v", err)
		http.Error(w, "Invalid pass value", http.StatusBadRequest)
		return
	}

	suiteName := h.DataStore.GetCurrentSuiteName()
	results := h.DataStore.ReadResults()
	if results == nil {
		results = make(map[string]middleware.Result)
	}
	if _, ok := results[model]; !ok {
		results[model] = middleware.Result{
			Scores: make([]int, len(h.DataStore.ReadPrompts())),
		}
	}

	prompts := h.DataStore.ReadPrompts()
	result := results[model]
	if len(result.Scores) < len(prompts) {
		result.Scores = append(result.Scores, make([]int, len(prompts)-len(result.Scores))...)
	}
	if promptIndex >= 0 && promptIndex < len(result.Scores) {
		if pass {
			result.Scores[promptIndex] = 100
		} else {
			result.Scores[promptIndex] = 0
		}
	}
	results[model] = result
	err = h.DataStore.WriteResults(suiteName, results)
	if err != nil {
		log.Printf("Error writing results: %v", err)
		http.Error(w, "Error writing results", http.StatusInternalServerError)
		return
	}

	h.DataStore.BroadcastResults()

	_, err = w.Write([]byte("OK"))
	if err != nil {
		log.Printf("Error writing response: %v", err)
		http.Error(w, "Error writing response", http.StatusInternalServerError)
		return
	}
	log.Println("protocols.Result updated successfully")
}

// ResetResults handles resetting results
func (h *Handler) ResetResults(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling reset results")
	switch r.Method {
	case http.MethodGet:
		if err := h.Renderer.RenderTemplateSimple(w, "reset_results.html", nil); err != nil {
			log.Printf("Error rendering template: %v", err)
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
		}
	case http.MethodPost:
		emptyResults := make(map[string]middleware.Result)
		suiteName := h.DataStore.GetCurrentSuiteName()
		err := h.DataStore.WriteResults(suiteName, emptyResults)
		if err != nil {
			log.Printf("Error writing results: %v", err)
			http.Error(w, "Error writing results", http.StatusInternalServerError)
			return
		}
		log.Println("Results reset successfully")
		h.DataStore.BroadcastResults()
		http.Redirect(w, r, "/results", http.StatusSeeOther)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// ConfirmRefreshResults handles confirm refresh results
func (h *Handler) ConfirmRefreshResults(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling confirm refresh results")
	switch r.Method {
	case http.MethodGet:
		if err := h.Renderer.RenderTemplateSimple(w, "confirm_refresh_results.html", nil); err != nil {
			log.Printf("Error rendering template: %v", err)
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
		}
	case http.MethodPost:
		results := h.DataStore.ReadResults()
		for model := range results {
			results[model] = middleware.Result{
				Scores: make([]int, len(h.DataStore.ReadPrompts())),
			}
		}
		suiteName := h.DataStore.GetCurrentSuiteName()
		err := h.DataStore.WriteResults(suiteName, results)
		if err != nil {
			log.Printf("Error writing results: %v", err)
			http.Error(w, "Error writing results", http.StatusInternalServerError)
			return
		}
		log.Println("Results refreshed successfully")
		h.DataStore.BroadcastResults()
		http.Redirect(w, r, "/results", http.StatusSeeOther)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// EvaluateResultHandler handles evaluation of individual results
