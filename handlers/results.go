package handlers

import (
	"encoding/json"
	"html/template"
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

// Handle results page
func ResultsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling results page")
	prompts := middleware.ReadPrompts()
	results := middleware.ReadResults()

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
	if len(models) == 0 {
		// If no results exist, get models from somewhere else if needed
		// This is just a fallback - you may need to adjust based on your data source
		models = []string{"Model1", "Model2"} // Example fallback
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
	t, err := template.New("results.html").Funcs(templates.FuncMap).ParseFiles("templates/results.html", "templates/nav.html")
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		http.Error(w, "Error parsing template: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if t == nil {
		log.Println("Error parsing template")
		http.Error(w, "Error parsing template", http.StatusInternalServerError)
		return
	}
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
		modelPassPercentages[model] = float64(totalScore) / float64(len(prompts)*100) * 100
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
	}

	err = t.Execute(w, templateData)
	if err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Error executing template", http.StatusInternalServerError)
		return
	}
	log.Println("Results page rendered successfully")
}

// Handle AJAX requests to update results
func UpdateResultHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling update result")
	r.ParseForm()
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

	suiteName := middleware.GetCurrentSuiteName()
	results := middleware.ReadResults()
	if results == nil {
		results = make(map[string]middleware.Result)
	}
	if _, ok := results[model]; !ok {
		results[model] = middleware.Result{
			Scores: make([]int, len(middleware.ReadPrompts())),
		}
	}

	prompts := middleware.ReadPrompts()
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
	err = middleware.WriteResults(suiteName, results)
	if err != nil {
		log.Printf("Error writing results: %v", err)
		http.Error(w, "Error writing results", http.StatusInternalServerError)
		return
	}

	middleware.BroadcastResults()

	_, err = w.Write([]byte("OK"))
	if err != nil {
		log.Printf("Error writing response: %v", err)
		http.Error(w, "Error writing response", http.StatusInternalServerError)
		return
	}
	log.Println("protocols.Result updated successfully")
}

// Handle reset results
func ResetResultsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling reset results")
	if r.Method == "GET" {
		t, err := template.ParseFiles("templates/reset_results.html")
		if err != nil {
			http.Error(w, "Error parsing template: "+err.Error(), http.StatusInternalServerError)
			return
		}
		err = t.Execute(w, nil)
		if err != nil {
			http.Error(w, "Error executing template: "+err.Error(), http.StatusInternalServerError)
			return
		}
	} else if r.Method == "POST" {
		emptyResults := make(map[string]middleware.Result)
		suiteName := middleware.GetCurrentSuiteName()
		err := middleware.WriteResults(suiteName, emptyResults)
		if err != nil {
			log.Printf("Error writing results: %v", err)
			http.Error(w, "Error writing results", http.StatusInternalServerError)
			return
		}
		log.Println("Results reset successfully")
		middleware.BroadcastResults()
		http.Redirect(w, r, "/results", http.StatusSeeOther)
	}
}

// Handle confirm refresh results
func ConfirmRefreshResultsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling confirm refresh results")
	if r.Method == "GET" {
		t, err := template.ParseFiles("templates/confirm_refresh_results.html")
		if err != nil {
			http.Error(w, "Error parsing template: "+err.Error(), http.StatusInternalServerError)
			return
		}
		err = t.Execute(w, nil)
		if err != nil {
			http.Error(w, "Error executing template: "+err.Error(), http.StatusInternalServerError)
			return
		}
	} else if r.Method == "POST" {
		results := middleware.ReadResults()
		for model := range results {
			results[model] = middleware.Result{
				Scores: make([]int, len(middleware.ReadPrompts())),
			}
		}
		suiteName := middleware.GetCurrentSuiteName()
		err := middleware.WriteResults(suiteName, results)
		if err != nil {
			log.Printf("Error writing results: %v", err)
			http.Error(w, "Error writing results", http.StatusInternalServerError)
			return
		}
		log.Println("Results refreshed successfully")
		middleware.BroadcastResults()
		http.Redirect(w, r, "/results", http.StatusSeeOther)
	}
}

// Handle refresh results
func RefreshResultsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling refresh results")
	if r.Method == "GET" {
		t, err := template.ParseFiles("templates/confirm_refresh_results.html")
		if err != nil {
			http.Error(w, "Error parsing template: "+err.Error(), http.StatusInternalServerError)
			return
		}
		err = t.Execute(w, nil)
		if err != nil {
			http.Error(w, "Error executing template: "+err.Error(), http.StatusInternalServerError)
			return
		}
	} else if r.Method == "POST" {
		results := middleware.ReadResults()
		for model := range results {
			results[model] = middleware.Result{Scores: make([]int, len(middleware.ReadPrompts()))}
		}
		suiteName := middleware.GetCurrentSuiteName()
		err := middleware.WriteResults(suiteName, results)
		if err != nil {
			log.Printf("Error writing results: %v", err)
			http.Error(w, "Error writing results", http.StatusInternalServerError)
			return
		}
		log.Println("Results refreshed successfully")
		middleware.BroadcastResults()
		http.Redirect(w, r, "/results", http.StatusSeeOther)
	}
}

// Handle evaluation of individual results
func EvaluateResult(w http.ResponseWriter, r *http.Request) {
	model := r.URL.Query().Get("model")
	promptIndexStr := r.URL.Query().Get("prompt")

	if r.Method == "POST" {
		scoreStr := r.FormValue("score")
		score, err := strconv.Atoi(scoreStr)
		if err != nil {
			http.Error(w, "Invalid score value", http.StatusBadRequest)
			return
		}

		results := middleware.ReadResults()
		if results == nil {
			results = make(map[string]middleware.Result)
		}

		result, exists := results[model]
		if !exists {
			// Initialize new result with scores array matching prompts length
			prompts := middleware.ReadPrompts()
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
		err = middleware.WriteResults(middleware.GetCurrentSuiteName(), results)
		if err != nil {
			http.Error(w, "Failed to save results", http.StatusInternalServerError)
			return
		}

		// Broadcast updated results to all clients
		middleware.BroadcastResults()

		// Add debug logging
		log.Printf("Updated score for model %s, prompt %d: %d", model, index, score)
		log.Printf("Current results for model %s: %v", model, result.Scores)

		// Redirect back to results page
		http.Redirect(w, r, "/results", http.StatusSeeOther)
		return
	}

	// Get current score for this model/prompt
	results := middleware.ReadResults()
	currentScore := 0
	if result, exists := results[model]; exists {
		if index, err := strconv.Atoi(promptIndexStr); err == nil && index < len(result.Scores) {
			currentScore = result.Scores[index]
		}
	}

	// Get the prompt text and solution for display
	prompts := middleware.ReadPrompts()
	var promptText, solution string
	promptIndex, err := strconv.Atoi(promptIndexStr)
	if err == nil && promptIndex >= 0 && promptIndex < len(prompts) {
		promptText = prompts[promptIndex].Text
		solution = prompts[promptIndex].Solution
	}

	data := struct {
		PageName     string
		Model        string
		PromptIndex  string
		ScoreOptions map[string]int
		CurrentScore int
		PromptText   string
		Solution     string
		TotalPrompts int
	}{
		PageName:     templates.PageNameEvaluate,
		Model:        model,
		PromptIndex:  promptIndexStr,
		ScoreOptions: templates.ScoreOptions,
		CurrentScore: currentScore,
		PromptText:   promptText,
		Solution:     solution,
		TotalPrompts: len(prompts),
	}

	t, err := template.New("evaluate.html").Funcs(templates.FuncMap).ParseFiles("templates/evaluate.html", "templates/nav.html")
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		http.Error(w, "Error parsing template: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = t.Execute(w, data)
	if err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Error executing template", http.StatusInternalServerError)
		return
	}
}

// Handle export results
func ExportResultsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling export results")
	results := middleware.ReadResults()

	// Convert results to JSON
	jsonData, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		log.Printf("Error marshaling results to JSON: %v", err)
		http.Error(w, "Error creating JSON export", http.StatusInternalServerError)
		return
	}

	// Set headers for JSON download
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment;filename=results.json")

	// Write JSON to response
	_, err = w.Write(jsonData)
	if err != nil {
		log.Printf("Error writing response: %v", err)
		http.Error(w, "Error writing response", http.StatusInternalServerError)
		return
	}
	log.Println("Results exported successfully as JSON")
}

// UpdateMockResultsHandler handles updating results with randomly generated mock data
// that ensures even distribution across all tier levels
func UpdateMockResultsHandler(w http.ResponseWriter, r *http.Request) {
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
	
	// We'll ignore the random scores from the client and generate evenly distributed ones
	prompts := middleware.ReadPrompts()
	
	// Get all model names
	models := mockData.Models
	if len(models) == 0 {
		// If no models passed, use models from existing results
		for model := range mockData.Results {
			models = append(models, model)
		}
	}
	
	// Define tier ranges to ensure even distribution
	tierRanges := []struct {
		min int
		max int
		name string
	}{
		{3000, 3300, "cosmic"},     // cosmic
		{2700, 2999, "divine"},     // divine
		{2400, 2699, "celestial"},  // celestial
		{2100, 2399, "ascendant"},  // ascendant
		{1800, 2099, "ethereal"},   // ethereal
		{1500, 1799, "mystic"},     // mystic
		{1200, 1499, "astral"},     // astral
		{900, 1199, "spiritual"},   // spiritual
		{600, 899, "primal"},       // primal
		{300, 599, "mortal"},       // mortal
		{0, 299, "primordial"},     // primordial
	}
	
	// Calculate how many models per tier
	modelsPerTier := len(models) / len(tierRanges)
	if modelsPerTier < 1 {
		modelsPerTier = 1
	}
	
	results := make(map[string]middleware.Result)
	
	// Distribute models evenly across tiers
	modelIndex := 0
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	
	for tierIndex, tierRange := range tierRanges {
		// Calculate how many models to put in this tier
		tierModelCount := modelsPerTier
		// Add extra models to the first few tiers if there are remainders
		if tierIndex < len(models)%len(tierRanges) {
			tierModelCount++
		}
		
		// Stop if we've used all models
		if modelIndex >= len(models) {
			break
		}
		
		// Assign models to this tier
		for i := 0; i < tierModelCount && modelIndex < len(models); i++ {
			model := models[modelIndex]
			
			// Calculate desired total score directly from tier range
			// This ensures we can reach all tiers including 3000+ and below 300
			totalPointsAvailable := len(prompts) * 100
			
			// Determine target score within the tier range
			var targetScore int
			if tierRange.min >= 3000 {
				// For cosmic tier, make some models go higher than 3000
				targetScore = tierRange.min + rng.Intn(tierRange.max-tierRange.min+200)
			} else if tierRange.max <= 299 {
				// For primordial tier, make some models go really low
				targetScore = rng.Intn(tierRange.max+1)
			} else {
				// For other tiers, distribute evenly within the range
				targetScore = tierRange.min + 
					(tierRange.max-tierRange.min)*i/max(1, tierModelCount-1)
			}
			
			// Calculate as percentage for the remaining algorithm
			targetPercentage := float64(targetScore) / float64(totalPointsAvailable) * 100
			
			// Calculate scores for individual prompts to achieve the target percentage
			scores := make([]int, len(prompts))
			remainingPoints := int(float64(totalPointsAvailable) * targetPercentage / 100)
			
			// Distribute points across prompts
			for j := 0; j < len(prompts); j++ {
				if j == len(prompts)-1 {
					// Last prompt gets all remaining points (if any) within 0-100 range
					scores[j] = remainingPoints
					if scores[j] > 100 {
						scores[j] = 100
					} else if scores[j] < 0 {
						scores[j] = 0
					}
				} else {
					// Other prompts get a distributed portion of points
					pointsForThisPrompt := 0
					if remainingPoints > 0 {
						// Determine a score that helps reach our target
						// but with some variability
						maxForThis := min(remainingPoints, 100)
						
						// More extreme distribution strategy based on tier
						// Create a distribution that fits the spiritual tier theme
						randomVariance := 0
						if tierRange.name == "cosmic" {
							// Cosmic tier: higher chance of perfect scores (enlightened performance)
							if rng.Float64() < 0.8 {  // 80% chance for max scores
								pointsForThisPrompt = 100
							} else if maxForThis > 10 {
								randomVariance = rng.Intn(maxForThis - 10)
								pointsForThisPrompt = min(maxForThis, 60+randomVariance)
							}
						} else if tierRange.name == "divine" || tierRange.name == "celestial" {
							// Divine/Celestial tiers: consistently high scores
							if rng.Float64() < 0.6 {  // 60% chance for max or near-max
								pointsForThisPrompt = 80 + rng.Intn(21) // 80-100
							} else if maxForThis > 10 {
								randomVariance = rng.Intn(maxForThis - 10)
								pointsForThisPrompt = min(maxForThis, 50+randomVariance)
							}
						} else if tierRange.name == "primordial" {
							// Primordial tier: higher chance of very low scores
							if rng.Float64() < 0.7 {  // 70% chance for low scores
								pointsForThisPrompt = rng.Intn(30)  // 0-29
							} else if maxForThis > 10 {
								randomVariance = rng.Intn(maxForThis - 10)
								pointsForThisPrompt = min(maxForThis, 20+randomVariance)
							}
						} else if tierRange.name == "mortal" || tierRange.name == "primal" {
							// Lower tiers: inconsistent performance
							if rng.Float64() < 0.5 {  // 50% chance for low-mid scores
								pointsForThisPrompt = 20 + rng.Intn(40)  // 20-59
							} else if maxForThis > 10 {
								randomVariance = rng.Intn(maxForThis - 10)
								pointsForThisPrompt = min(maxForThis, 30+randomVariance)
							}
						} else {
							// Middle tiers: balanced distribution with more variance
							if maxForThis > 10 {
								randomVariance = rng.Intn(maxForThis - 10)
								pointsForThisPrompt = min(maxForThis, 40+randomVariance)
							}
						}
					}
					scores[j] = pointsForThisPrompt
					remainingPoints -= pointsForThisPrompt
				}
			}
			
			// Add results for this model
			results[model] = middleware.Result{
				Scores: scores,
			}
			
			modelIndex++
		}
	}
	
	// Save the evenly distributed mock results
	suiteName := middleware.GetCurrentSuiteName()
	err = middleware.WriteResults(suiteName, results)
	if err != nil {
		log.Printf("Error writing mock results: %v", err)
		http.Error(w, "Error saving mock results", http.StatusInternalServerError)
		return
	}
	
	// Broadcast the updated results to all connected clients
	middleware.BroadcastResults()
	
	// Calculate totalScores and passPercentages for the response
	totalScores := make(map[string]int)
	passPercentages := make(map[string]float64)
	
	for model, result := range results {
		totalScore := 0
		for _, score := range result.Scores {
			totalScore += score
		}
		totalScores[model] = totalScore
		passPercentages[model] = float64(totalScore) / float64(len(prompts)*100) * 100
	}
	
	// Return success response with the generated data
	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"status": "success",
		"results": results,
		"models": models,
		"totalScores": totalScores,
		"passPercentages": passPercentages,
	}
	
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Printf("Error encoding response: %v", err)
	}
	
	log.Println("Mock results with even tier distribution updated successfully")
}
