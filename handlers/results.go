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
		score := 0
		for _, pass := range result.Passes {
			if pass {
				score++
			}
		}
		modelScores[model] = score
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
		resultsForTemplate[model] = result
	}
	modelPassPercentages := make(map[string]float64)
	modelTotalScores := make(map[string]int)
	promptIndices := make([]int, len(prompts))
	for i := range prompts {
		promptIndices[i] = i + 1
	}
	for model, result := range filteredResults {
		score := 0
		for _, pass := range result.Passes {
			if pass {
				score++
			}
		}
		modelPassPercentages[model] = float64(score) / float64(len(prompts)) * 100
		modelTotalScores[model] = score * 100
	}

	err = t.Execute(w, struct {
		PageName        string
		Prompts         []string
		Results         map[string][]bool
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
	})
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
		results[model] = middleware.Result{Passes: make([]bool, len(middleware.ReadPrompts()))}
	}

	prompts := middleware.ReadPrompts()
	result := results[model]
	if len(result.Passes) < len(prompts) {
		result.Passes = append(result.Passes, make([]bool, len(prompts)-len(result.Passes))...)
	}

	if promptIndex >= 0 && promptIndex < len(result.Passes) {
		result.Passes[promptIndex] = pass
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
			results[model] = middleware.Result{Passes: make([]bool, len(middleware.ReadPrompts()))}
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
			results[model] = middleware.Result{Passes: make([]bool, len(middleware.ReadPrompts()))}
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
		score, _ := strconv.Atoi(scoreStr)

		results := middleware.ReadResults()
		result := results[model]

		index, _ := strconv.Atoi(promptIndexStr)
		if index >= 0 && index < len(result.Scores) {
			result.Scores[index] = score
			results[model] = result
			middleware.WriteResults(middleware.GetCurrentSuiteName(), results)
			middleware.BroadcastResults()
		}

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
		for _, pass := range result.Passes {
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
