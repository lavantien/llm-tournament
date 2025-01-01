package main

import (
	"html/template"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

func main() {
	http.HandleFunc("/", router)
	http.Handle("/templates/", http.StripPrefix("/templates/", http.FileServer(http.Dir("templates"))))
	http.ListenAndServe(":8080", nil)
}

func router(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/prompts" {
		promptListHandler(w, r)
	} else if r.URL.Path == "/results" {
		resultsHandler(w, r)
	} else if r.URL.Path == "/update_result" {
		updateResultHandler(w, r)
	} else if r.URL.Path == "/add_model" {
		addModelHandler(w, r)
	} else if r.URL.Path == "/add_prompt" {
		addPromptHandler(w, r)
	} else if r.URL.Path == "/edit_prompt" {
		editPromptHandler(w, r)
	} else if r.URL.Path == "/delete_prompt" {
		deletePromptHandler(w, r)
	} else if r.URL.Path == "/reset_results" {
		resetResultsHandler(w, r)
	} else if r.URL.Path == "/export_results" {
		exportResultsHandler(w, r)
	} else if r.URL.Path == "/import_results" {
		importResultsHandler(w, r)
	} else if r.URL.Path == "/export_prompts" {
		exportPromptsHandler(w, r)
	} else if r.URL.Path == "/import_prompts" {
		importPromptsHandler(w, r)
	} else {
		http.Redirect(w, r, "/prompts", http.StatusSeeOther)
	}
}

func add(a, b int) int {
	return a + b
}

// Handle add prompt
func addPromptHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}
	promptText := r.Form.Get("prompt")
	if promptText == "" {
		http.Error(w, "Prompt text cannot be empty", http.StatusBadRequest)
		return
	}
	prompts := readPrompts()
	prompts = append(prompts, Prompt{Text: promptText})
	err = writePrompts(prompts)
	if err != nil {
		http.Error(w, "Error writing prompts", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/prompts", http.StatusSeeOther)
}

// Handle add model
func addModelHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}
	modelName := r.Form.Get("model")
	if modelName == "" {
		http.Error(w, "Model name cannot be empty", http.StatusBadRequest)
		return
	}
	results := readResults()
	if results == nil {
		results = make(map[string]Result)
	}
	if _, ok := results[modelName]; !ok {
		results[modelName] = Result{Passes: make([]bool, len(readPrompts()))}
	}
	err = writeResults(results)
	if err != nil {
		http.Error(w, "Error writing results", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/results", http.StatusSeeOther)
}

// Handle export prompts
func exportPromptsHandler(w http.ResponseWriter, r *http.Request) {
	prompts := readPrompts()

	// Create CSV string
	csvString := "Prompt\n"
	for _, prompt := range prompts {
		csvString += prompt.Text + "\n"
	}

	// Set headers for CSV download
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment;filename=prompts.csv")

	// Write CSV to response
	_, err := w.Write([]byte(csvString))
	if err != nil {
		http.Error(w, "Error writing response", http.StatusInternalServerError)
		return
	}
}

// Handle import prompts
func importPromptsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		file, _, err := r.FormFile("prompts_file")
		if err != nil {
			http.Error(w, "Error uploading file", http.StatusBadRequest)
			return
		}
		defer file.Close()

		// Read the file content
		data := make([]byte, 0)
		buf := make([]byte, 1024)
		for {
			n, err := file.Read(buf)
			if err != nil && err.Error() != "EOF" {
				http.Error(w, "Error reading file", http.StatusInternalServerError)
				return
			}
			if n > 0 {
				data = append(data, buf[:n]...)
			}
			if err != nil {
				break
			}
		}

		// Parse CSV data
		lines := strings.Split(string(data), "\n")
		if len(lines) <= 1 {
			http.Error(w, "Invalid CSV format: No data found", http.StatusBadRequest)
			return
		}

		var prompts []Prompt
		for _, line := range lines {
			if line == "" || line == "Prompt" {
				continue
			}
			prompts = append(prompts, Prompt{Text: line})
		}
		writePrompts(prompts)
		http.Redirect(w, r, "/prompts", http.StatusSeeOther)
	} else {
		t, _ := template.ParseFiles("templates/import_prompts.html")
		t.Execute(w, nil)
	}
}

// Handle edit prompt
func editPromptHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}
		indexStr := r.Form.Get("index")
		index, err := strconv.Atoi(indexStr)
		if err != nil {
			http.Error(w, "Invalid index", http.StatusBadRequest)
			return
		}
		prompts := readPrompts()
		if index >= 0 && index < len(prompts) {
			t, err := template.ParseFiles("templates/edit_prompt.html")
			if err != nil {
				http.Error(w, "Error parsing template", http.StatusInternalServerError)
				return
			}
			err = t.Execute(w, struct {
				Index  int
				Prompt string
			}{
				Index:  index,
				Prompt: prompts[index].Text,
			})
			if err != nil {
				http.Error(w, "Error executing template", http.StatusInternalServerError)
				return
			}
		}
	} else if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}
		indexStr := r.Form.Get("index")
		index, err := strconv.Atoi(indexStr)
		if err != nil {
			http.Error(w, "Invalid index", http.StatusBadRequest)
			return
		}
		editedPrompt := r.Form.Get("prompt")
		if editedPrompt == "" {
			http.Error(w, "Prompt text cannot be empty", http.StatusBadRequest)
			return
		}
		prompts := readPrompts()
		if index >= 0 && index < len(prompts) {
			prompts[index].Text = editedPrompt
		}
		err = writePrompts(prompts)
		if err != nil {
			http.Error(w, "Error writing prompts", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/prompts", http.StatusSeeOther)
	}
}

// Handle delete prompt
func deletePromptHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}
		indexStr := r.Form.Get("index")
		index, err := strconv.Atoi(indexStr)
		if err != nil {
			http.Error(w, "Invalid index", http.StatusBadRequest)
			return
		}
		prompts := readPrompts()
		if index >= 0 && index < len(prompts) {
			t, err := template.ParseFiles("templates/delete_prompt.html")
			if err != nil {
				http.Error(w, "Error parsing template", http.StatusInternalServerError)
				return
			}
			err = t.Execute(w, struct {
				Index  int
				Prompt string
			}{
				Index:  index,
				Prompt: prompts[index].Text,
			})
			if err != nil {
				http.Error(w, "Error executing template", http.StatusInternalServerError)
				return
			}
		}
	} else if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}
		indexStr := r.Form.Get("index")
		index, err := strconv.Atoi(indexStr)
		if err != nil {
			http.Error(w, "Invalid index", http.StatusBadRequest)
			return
		}
		prompts := readPrompts()
		if index >= 0 && index < len(prompts) {
			prompts = append(prompts[:index], prompts[index+1:]...)
		}
		err = writePrompts(prompts)
		if err != nil {
			http.Error(w, "Error writing prompts", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/prompts", http.StatusSeeOther)
	}
}

// Handle prompt list page
func promptListHandler(w http.ResponseWriter, r *http.Request) {
	prompts := readPrompts()
	promptTexts := make([]string, len(prompts))
	promptIndices := make([]int, len(prompts))
	for i, prompt := range prompts {
		promptTexts[i] = prompt.Text
		promptIndices[i] = i + 1
	}
	funcMap := template.FuncMap{
		"inc": func(i int) int {
			return i + 1
		},
	}
	t, err := template.New("prompt_list.html").Funcs(funcMap).ParseFiles("templates/prompt_list.html", "templates/nav.html")
	if err != nil {
		http.Error(w, "Error parsing template: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if t == nil {
		http.Error(w, "Error parsing template", http.StatusInternalServerError)
		return
	}
	err = t.Execute(w, promptTexts)
	if err != nil {
		http.Error(w, "Error executing template", http.StatusInternalServerError)
		return
	}
}

// Handle results page
func resultsHandler(w http.ResponseWriter, r *http.Request) {
	prompts := readPrompts()
	results := readResults()

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

	// Sort models by score in descending order
	models := make([]string, 0, len(results))
	for model := range results {
		models = append(models, model)
	}
	sort.Slice(models, func(i, j int) bool {
		return modelScores[models[i]] > modelScores[models[j]]
	})

	funcMap := template.FuncMap{
		"inc": func(i int) int {
			// slog.Info(strconv.Itoa(i))
			return i + 1
		},
	}
	t, err := template.New("results.html").Funcs(funcMap).ParseFiles("templates/results.html", "templates/nav.html")
	if err != nil {
		http.Error(w, "Error parsing template: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if t == nil {
		http.Error(w, "Error parsing template", http.StatusInternalServerError)
		return
	}
	promptTexts := make([]string, len(prompts))
	for i, prompt := range prompts {
		promptTexts[i] = prompt.Text
	}
	resultsForTemplate := make(map[string][]bool)
	for model, result := range results {
		resultsForTemplate[model] = result.Passes
	}
	modelPassPercentages := make(map[string]float64)
	modelTotalScores := make(map[string]int)
	promptIndices := make([]int, len(prompts))
	for i := range prompts {
		promptIndices[i] = i + 1
	}
	for model, result := range results {
		score := 0
		for _, pass := range result.Passes {
			if pass {
				score++
			}
		}
		modelPassPercentages[model] = float64(score) / float64(len(prompts)) * 100
		modelTotalScores[model] = score
	}

	modelFilter := r.FormValue("model_filter")

	err = t.Execute(w, struct {
		Prompts         []string
		Results         map[string][]bool
		Models          []string
		PassPercentages map[string]float64
		ModelFilter     string
		TotalScores     map[string]int
		PromptIndices   []int
	}{
		Prompts:         promptTexts,
		Results:         resultsForTemplate,
		Models:          models,
		PassPercentages: modelPassPercentages,
		ModelFilter:     modelFilter,
		TotalScores:     modelTotalScores,
		PromptIndices:   promptIndices,
	})
	if err != nil {
		http.Error(w, "Error executing template", http.StatusInternalServerError)
		return
	}
}

// Handle AJAX requests to update results
func updateResultHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	model := r.Form.Get("model")
	promptIndexStr := r.Form.Get("promptIndex")
	passStr := r.Form.Get("pass")
	promptIndex, _ := strconv.Atoi(promptIndexStr)
	pass, err := strconv.ParseBool(passStr)
	if err != nil {
		http.Error(w, "Invalid pass value", http.StatusBadRequest)
		return
	}

	results := readResults()
	if results == nil {
		results = make(map[string]Result)
	}
	if _, ok := results[model]; !ok {
		results[model] = Result{Passes: make([]bool, len(readPrompts()))}
	}

	prompts := readPrompts()
	result := results[model]
	if len(result.Passes) < len(prompts) {
		result.Passes = append(result.Passes, make([]bool, len(prompts)-len(result.Passes))...)
	}

	if promptIndex >= 0 && promptIndex < len(result.Passes) {
		result.Passes[promptIndex] = pass
	}
	results[model] = result
	err = writeResults(results)
	if err != nil {
		http.Error(w, "Error writing results", http.StatusInternalServerError)
		return
	}

	_, err = w.Write([]byte("OK"))
	if err != nil {
		http.Error(w, "Error writing response", http.StatusInternalServerError)
		return
	}
}

// Handle reset results
func resetResultsHandler(w http.ResponseWriter, r *http.Request) {
	emptyResults := make(map[string]Result)
	err := writeResults(emptyResults)
	if err != nil {
		http.Error(w, "Error writing results", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/results", http.StatusSeeOther)
}

// Handle export results
func exportResultsHandler(w http.ResponseWriter, r *http.Request) {
	results := readResults()
	prompts := readPrompts()

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
	w.Write([]byte(csvString))
}

// Handle import results
func importResultsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		file, _, err := r.FormFile("results_file")
		if err != nil {
			http.Error(w, "Error uploading file", http.StatusBadRequest)
			return
		}
		defer file.Close()

		// Read the file content
		data := make([]byte, 0)
		buf := make([]byte, 1024)
		for {
			n, err := file.Read(buf)
			if n > 0 {
				data = append(data, buf[:n]...)
			}
			if err != nil {
				break
			}
		}

		// Parse CSV data
		lines := strings.Split(string(data), "\n")
		if len(lines) <= 1 {
			http.Error(w, "Invalid CSV format: No data found", http.StatusBadRequest)
			return
		}

		results := make(map[string]Result)
		prompts := readPrompts()
		for i, line := range lines {
			if i == 0 || line == "" {
				continue
			}
			parts := strings.Split(line, ",")
			if len(parts) < 2 {
				continue
			}
			model := parts[0]
			var passes []bool
			for _, passStr := range parts[1:] {
				pass, _ := strconv.ParseBool(passStr)
				passes = append(passes, pass)
			}
			if len(passes) < len(prompts) {
				passes = append(passes, make([]bool, len(prompts)-len(passes))...)
			}
			results[model] = Result{Passes: passes}
		}
		err = writeResults(results)
		if err != nil {
			http.Error(w, "Error writing results", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/results", http.StatusSeeOther)
	} else {
		t, err := template.ParseFiles("templates/import_results.html")
		if err != nil {
			http.Error(w, "Error parsing template", http.StatusInternalServerError)
			return
		}
		err = t.Execute(w, nil)
		if err != nil {
			http.Error(w, "Error executing template", http.StatusInternalServerError)
			return
		}
	}
}
