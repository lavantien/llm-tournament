package handlers

import (
	"html/template"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"llm-tournament/middleware"
	"llm-tournament/templates"
)

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

	data := struct {
		PageName     string
		Model        string
		PromptIndex  string
		ScoreOptions map[string]int
		CurrentScore int
	}{
		PageName:     templates.PageNameEvaluate,
		Model:        model,
		PromptIndex:  promptIndexStr,
		ScoreOptions: templates.ScoreOptions,
		CurrentScore: currentScore,
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
	prompts := middleware.ReadPrompts()

	// Create CSV string
	csvString := "Model,"
	for i := range prompts {
		csvString += "Prompt " + strconv.Itoa(i+1) + ","
	}
	csvString += "\n"

	for model, result := range results {
		csvString += model + ","
		for _, score := range result.Scores {
			// Convert score to pass/fail (100 = pass, 0 = fail)
			pass := score == 100
			csvString += strconv.FormatBool(pass) + ","
		}
		csvString += "\n"
	}

	// Set headers for CSV download
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment;filename=results.csv")

	// Write CSV to response
	_, err := w.Write([]byte(csvString))
	if err != nil {
		log.Printf("Error writing response: %v", err)
		http.Error(w, "Error writing response", http.StatusInternalServerError)
		return
	}
	log.Println("Results exported successfully")
}
