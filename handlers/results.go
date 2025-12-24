package handlers

import (
	"encoding/json"
	"io"
	"log"
	"math/rand"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"llm-tournament/middleware"
	"llm-tournament/templates"
)

// min returns the smaller of x or y
func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

// max returns the larger of x or y
func max(x, y int) int {
	if x > y {
		return x
	}
	return y
}

// initRand returns a new random number generator seeded with the current time
func initRand() *rand.Rand {
	source := rand.NewSource(time.Now().UnixNano())
	return rand.New(source)
}

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

// RefreshResultsHandler handles refresh results (backward compatible wrapper)
func RefreshResultsHandler(w http.ResponseWriter, r *http.Request) {
	DefaultHandler.RefreshResults(w, r)
}

// EvaluateResult handles evaluating individual results (backward compatible wrapper)
func EvaluateResult(w http.ResponseWriter, r *http.Request) {
	DefaultHandler.EvaluateResultHandler(w, r)
}

// ExportResultsHandler handles exporting results (backward compatible wrapper)
func ExportResultsHandler(w http.ResponseWriter, r *http.Request) {
	DefaultHandler.ExportResults(w, r)
}

// UpdateMockResultsHandler handles updating mock results (backward compatible wrapper)
func UpdateMockResultsHandler(w http.ResponseWriter, r *http.Request) {
	DefaultHandler.UpdateMockResults(w, r)
}

// Results handles the results page
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
	case "GET":
		if err := h.Renderer.RenderTemplateSimple(w, "reset_results.html", nil); err != nil {
			log.Printf("Error rendering template: %v", err)
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
		}
	case "POST":
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
	case "GET":
		if err := h.Renderer.RenderTemplateSimple(w, "confirm_refresh_results.html", nil); err != nil {
			log.Printf("Error rendering template: %v", err)
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
		}
	case "POST":
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

// RefreshResults handles refresh results
func (h *Handler) RefreshResults(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling refresh results")
	switch r.Method {
	case "GET":
		if err := h.Renderer.RenderTemplateSimple(w, "confirm_refresh_results.html", nil); err != nil {
			log.Printf("Error rendering template: %v", err)
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
		}
	case "POST":
		results := h.DataStore.ReadResults()
		for model := range results {
			results[model] = middleware.Result{Scores: make([]int, len(h.DataStore.ReadPrompts()))}
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
func (h *Handler) EvaluateResultHandler(w http.ResponseWriter, r *http.Request) {
	model := r.URL.Query().Get("model")
	promptIndexStr := r.URL.Query().Get("prompt")

	if r.Method == "POST" {
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
		if index, err := strconv.Atoi(promptIndexStr); err == nil && index < len(result.Scores) {
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

	// Get model_id from model name
	err = db.QueryRow("SELECT id FROM models WHERE name = ?", model).Scan(&modelID)
	if err == nil {
		// Get prompt_id from database using prompt index (1-indexed)
		err = db.QueryRow("SELECT id FROM prompts WHERE suite_id = 1 ORDER BY display_order LIMIT 1 OFFSET ?", promptIndex).Scan(&promptID)
		if err == nil {
			// Get the response for this model/prompt pair
			err = db.QueryRow("SELECT response_text FROM model_responses WHERE model_id = ? AND prompt_id = ?", modelID, promptID).Scan(&modelResponse)
			if err != nil {
				// No response found, leave empty
				modelResponse = ""
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
	}

	err = h.Renderer.Render(w, "evaluate.html", templates.FuncMap, data, "templates/evaluate.html", "templates/nav.html")
	if err != nil {
		log.Printf("Error rendering template: %v", err)
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		return
	}
}

// ExportResults handles export results
func (h *Handler) ExportResults(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling export results")
	results := h.DataStore.ReadResults()

	// Convert results to JSON
	jsonData, _ := json.MarshalIndent(results, "", "  ")

	// Set headers for JSON download
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment;filename=results.json")

	// Write JSON to response
	_, err := w.Write(jsonData)
	if err != nil {
		log.Printf("Error writing response: %v", err)
		http.Error(w, "Error writing response", http.StatusInternalServerError)
		return
	}
	log.Println("Results exported successfully as JSON")
}

// UpdateMockResults handles updating results with randomly generated mock data
// that ensures even distribution across all tier levels
func (h *Handler) UpdateMockResults(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling update mock results")

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse the JSON request body
	var mockData struct {
		Results         map[string]middleware.Result `json:"results"`
		Models          []string                     `json:"models"`
		PassPercentages map[string]float64           `json:"passPercentages"`
		TotalScores     map[string]int               `json:"totalScores"`
	}

	log.Println("Received mock data request")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(body, &mockData)
	if err != nil {
		log.Printf("Error decoding mock data: %v", err)
		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}

	// Use client-provided scores instead of generating new ones
	log.Println("Using client-provided scores for mock data")

	prompts := h.DataStore.ReadPrompts()

	// If no prompts exist, create mock prompts
	if len(prompts) == 0 {
		db := middleware.GetDB()
		suiteID, err := middleware.GetCurrentSuiteID()
		if err != nil || suiteID == 0 {
			suiteID = 1
		}
		mockPrompts := []string{
			"What is 2 + 2?",
			"Explain recursion in programming",
			"What is the capital of France?",
			"Write a function to reverse a string",
			"What is photosynthesis?",
		}
		for i, text := range mockPrompts {
			_, err = db.Exec("INSERT INTO prompts (text, suite_id, display_order, type) VALUES (?, ?, ?, 'objective')", text, suiteID, i)
			if err != nil {
				log.Printf("Error inserting mock prompt: %v", err)
			}
		}
		prompts = h.DataStore.ReadPrompts()
		log.Printf("Created %d mock prompts", len(prompts))

		// Create mock profiles
		profiles := []string{"Math", "Philosophy", "Programming", "Science", "Writing"}
		for _, name := range profiles {
			_, err = db.Exec("INSERT INTO profiles (name, suite_id) VALUES (?, ?)", name, suiteID)
			if err != nil {
				log.Printf("Error inserting mock profile: %v", err)
			}
		}
		log.Printf("Created %d mock profiles", len(profiles))
	}

	// Get all model names
	models := mockData.Models
	if len(models) == 0 {
		// If no models passed, use models from existing results
		for model := range mockData.Results {
			models = append(models, model)
		}
	}

	// Use the client's results directly
	results := mockData.Results

	// Generate mock models if both models and results are empty
	if len(models) == 0 && len(results) == 0 {
		results = make(map[string]middleware.Result)
		tiers := []string{"Cosmic", "Transcendent", "Ethereal", "Celestial", "Infinite",
			"Quantum", "Nebular", "Stellar", "Galactic", "Universal", "Dimensional"}
		for i := 0; i < 15; i++ {
			tier := tiers[i%len(tiers)]
			num := i/len(tiers) + 1
			modelName := tier + "-" + strconv.Itoa(num)
			models = append(models, modelName)
			results[modelName] = middleware.Result{Scores: make([]int, len(prompts))}
		}
	}

	// Validate that all scores are legitimate values: 0, 20, 40, 60, 80, 100
	for model, result := range results {
		for i, score := range result.Scores {
			// Only allow valid score values
			switch score {
			case 0, 20, 40, 60, 80, 100:
				// Valid score, keep it
			default:
				// Invalid score, set to 0
				log.Printf("Correcting invalid score %d for model %s prompt %d", score, model, i)
				result.Scores[i] = 0
			}
		}
		results[model] = result
	}

	// Skip the evenly distributed tier generation since we're using client scores

	// Save the evenly distributed mock results
	suiteName := h.DataStore.GetCurrentSuiteName()
	err = h.DataStore.WriteResults(suiteName, results)
	if err != nil {
		log.Printf("Error writing mock results: %v", err)
		http.Error(w, "Error saving mock results", http.StatusInternalServerError)
		return
	}

	// Generate mock responses for each model and prompt combination
	// Get database for inserting mock responses
	db := middleware.GetDB()
	suiteID, err := middleware.GetCurrentSuiteID()
	if err != nil {
		suiteID = 1 // fallback to default suite
	}

	// Get all prompts for response generation
	for _, modelName := range models {
		// Get model ID
		var modelID int
		err = db.QueryRow("SELECT id FROM models WHERE name = ? AND suite_id = ?", modelName, suiteID).Scan(&modelID)
		if err != nil {
			continue // model might not exist yet
		}

		// Generate mock response for each prompt
		for promptIdx := range prompts {
			var promptID int
			err = db.QueryRow("SELECT id FROM prompts WHERE suite_id = ? ORDER BY display_order LIMIT 1 OFFSET ?",
				suiteID, promptIdx).Scan(&promptID)
			if err != nil {
				continue
			}

			// Generate Lorem ipsum mock response
			loremPhrases := []string{
				"Lorem ipsum dolor sit amet, consectetur adipiscing elit.",
				"Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.",
				"Ut enim ad minim veniam, quis nostrud exercitation ullamco.",
				"Duis aute irure dolor in reprehenderit in voluptate velit esse.",
				"Excepteur sint occaecat cupidatat non proident sunt in culpa.",
			}

			// Build 3-5 random sentences
			numSentences := 3 + rand.Intn(3)
			var responseParts []string
			for i := 0; i < numSentences; i++ {
				responseParts = append(responseParts, loremPhrases[rand.Intn(len(loremPhrases))])
			}
			mockResponse := strings.Join(responseParts, " ")

			// Insert or update mock response
			_, err = db.Exec(
				"INSERT INTO model_responses (model_id, prompt_id, response_text, response_source) "+
					"VALUES (?, ?, ?, 'mock') "+
					"ON CONFLICT(model_id, prompt_id) DO UPDATE SET "+
					"response_text = excluded.response_text, response_source = 'mock', updated_at = CURRENT_TIMESTAMP",
				modelID, promptID, mockResponse)
			if err != nil {
				log.Printf("Error inserting mock response for model %s prompt %d: %v", modelName, promptIdx, err)
			}
		}
	}

	// Broadcast the updated results to all connected clients
	h.DataStore.BroadcastResults()

	// Calculate totalScores and passPercentages for the response
	totalScores := make(map[string]int)
	passPercentages := make(map[string]float64)

	log.Println("Calculating total scores for each model:")
	for model, result := range results {
		totalScore := 0
		for _, score := range result.Scores {
			totalScore += score
		}
		totalScores[model] = totalScore
		// Avoid division by zero when there are no prompts
		if len(prompts) > 0 {
			passPercentages[model] = float64(totalScore) / float64(len(prompts)*100) * 100
		} else {
			passPercentages[model] = 0
		}

		log.Printf("Model %s: total score = %d, pass percentage = %.2f%%",
			model, totalScore, passPercentages[model])
	}

	// Sort models by total score in descending order
	sort.Slice(models, func(i, j int) bool {
		return totalScores[models[i]] > totalScores[models[j]]
	})

	log.Printf("Sorted models after mock generation: %v", models[:min(5, len(models))])

	// Return success response with the generated data
	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"status":          "success",
		"results":         results,
		"models":          models, // Now sorted by score
		"totalScores":     totalScores,
		"passPercentages": passPercentages,
	}

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Printf("Error encoding response: %v", err)
	}

	log.Println("Mock results with even tier distribution updated successfully")
}

// RandomizeScoresHandler handles randomizing scores (backward compatible wrapper)
func RandomizeScoresHandler(w http.ResponseWriter, r *http.Request) {
	DefaultHandler.RandomizeScores(w, r)
}

// RandomizeScores randomizes existing scores in the database without creating new models or prompts
func (h *Handler) RandomizeScores(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling randomize scores")

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	db := middleware.GetDB()
	suiteID, err := middleware.GetCurrentSuiteID()
	if err != nil {
		log.Printf("Error getting suite ID: %v", err)
		http.Error(w, "Error getting suite ID", http.StatusInternalServerError)
		return
	}

	modelRows, err := db.Query("SELECT id, name FROM models WHERE suite_id = ?", suiteID)
	if err != nil {
		log.Printf("Error querying models: %v", err)
		http.Error(w, "Error querying models", http.StatusInternalServerError)
		return
	}
	defer modelRows.Close()

	var models []struct {
		ID   int
		Name string
	}

	for modelRows.Next() {
		var m struct {
			ID   int
			Name string
		}
		if err := modelRows.Scan(&m.ID, &m.Name); err != nil {
			continue
		}
		models = append(models, m)
	}

	promptRows, err := db.Query("SELECT id FROM prompts WHERE suite_id = ? ORDER BY display_order", suiteID)
	if err != nil {
		http.Error(w, "Error querying prompts", http.StatusInternalServerError)
		return
	}
	defer promptRows.Close()

	var promptIDs []int
	for promptRows.Next() {
		var id int
		if err := promptRows.Scan(&id); err != nil {
			continue
		}
		promptIDs = append(promptIDs, id)
	}

	validScores := []int{0, 20, 40, 60, 80, 100}
	rng := initRand()

	for _, model := range models {
		for _, promptID := range promptIDs {
			randomScore := validScores[rng.Intn(len(validScores))]
			db.Exec(
				"INSERT INTO scores (model_id, prompt_id, score) VALUES (?, ?, ?) "+
					"ON CONFLICT(model_id, prompt_id) DO UPDATE SET score = excluded.score",
				model.ID, promptID, randomScore)
		}
	}

	h.DataStore.BroadcastResults()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}
