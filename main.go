package main

import (
	"html/template"
	"net/http"
	"sort"
    "strings"
	"strconv"
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

// Handle add prompt
func addPromptHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	promptText := r.Form.Get("prompt")
	prompts := readPrompts()
	prompts = append(prompts, Prompt{Text: promptText})
	writePrompts(prompts)
	http.Redirect(w, r, "/prompts", http.StatusSeeOther)
}

// Handle add model
func addModelHandler(w http.ResponseWriter, r *http.Request) {
    r.ParseForm()
    modelName := r.Form.Get("model")
    results := readResults()
    if _, ok := results[modelName]; !ok {
        results[modelName] = Result{Passes: make([]bool, len(readPrompts()))}
    }
    writeResults(results)
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
    w.Write([]byte(csvString))
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
            http.Error(w, "Invalid CSV format", http.StatusBadRequest)
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
        r.ParseForm()
        indexStr := r.Form.Get("index")
        index, _ := strconv.Atoi(indexStr)
        prompts := readPrompts()
        if index >= 0 && index < len(prompts) {
            t, _ := template.ParseFiles("templates/edit_prompt.html")
            t.Execute(w, struct {
                Index int
                Prompt string
            }{
                Index: index,
                Prompt: prompts[index].Text,
            })
        }
    } else if r.Method == "POST" {
        r.ParseForm()
        indexStr := r.Form.Get("index")
        index, _ := strconv.Atoi(indexStr)
        editedPrompt := r.Form.Get("prompt")
        prompts := readPrompts()
        if index >= 0 && index < len(prompts) {
            prompts[index].Text = editedPrompt
        }
        writePrompts(prompts)
        http.Redirect(w, r, "/prompts", http.StatusSeeOther)
    }
}

// Handle delete prompt
func deletePromptHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		r.ParseForm()
		indexStr := r.Form.Get("index")
		index, _ := strconv.Atoi(indexStr)
		prompts := readPrompts()
		if index >= 0 && index < len(prompts) {
			t, _ := template.ParseFiles("templates/delete_prompt.html")
			t.Execute(w, struct {
				Index  int
				Prompt string
			}{
				Index:  index,
				Prompt: prompts[index].Text,
			})
		}
	} else if r.Method == "POST" {
		r.ParseForm()
		indexStr := r.Form.Get("index")
		index, _ := strconv.Atoi(indexStr)
		prompts := readPrompts()
		if index >= 0 && index < len(prompts) {
			prompts = append(prompts[:index], prompts[index+1:]...)
		}
		writePrompts(prompts)
		http.Redirect(w, r, "/prompts", http.StatusSeeOther)
    }
}


// Handle prompt list page
func promptListHandler(w http.ResponseWriter, r *http.Request) {
	prompts := readPrompts()
    promptTexts := make([]string, len(prompts))
    for i, prompt := range prompts {
        promptTexts[i] = prompt.Text
    }
	t, _ := template.ParseFiles("templates/prompt_list.html")
	t.Execute(w, promptTexts)
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

	t, _ := template.ParseFiles("templates/results.html")
    promptTexts := make([]string, len(prompts))
    for i, prompt := range prompts {
        promptTexts[i] = prompt.Text
    }
    resultsForTemplate := make(map[string][]bool)
    for model, result := range results {
        resultsForTemplate[model] = result.Passes
    }
    modelPassPercentages := make(map[string]float64)
    for model, result := range results {
        score := 0
        for _, pass := range result.Passes {
            if pass {
                score++
            }
        }
        modelPassPercentages[model] = float64(score) / float64(len(prompts)) * 100
    }

    modelFilter := r.FormValue("model_filter")

	t.Execute(w, struct {
		Prompts  []string
		Results  map[string][]bool
		Models   []string
        PassPercentages map[string]float64
        ModelFilter string
	}{
		Prompts:  promptTexts,
		Results:  resultsForTemplate,
		Models:   models,
        PassPercentages: modelPassPercentages,
        ModelFilter: modelFilter,
	})
}

// Handle AJAX requests to update results
func updateResultHandler(w http.ResponseWriter, r *http.Request) {
    r.ParseForm()
    model := r.Form.Get("model")
    promptIndexStr := r.Form.Get("promptIndex")
    passStr := r.Form.Get("pass")
    promptIndex, _ := strconv.Atoi(promptIndexStr)
    pass, _ := strconv.ParseBool(passStr)

    results := readResults()
    if _, ok := results[model]; !ok {
        results[model] = Result{Passes: make([]bool, len(readPrompts()))}
    }
    
    prompts := readPrompts()
    result := results[model]
    if len(result.Passes) < len(prompts) {
        result.Passes = append(result.Passes, make([]bool, len(prompts) - len(result.Passes))...)
    }
    
    if promptIndex >= 0 && promptIndex < len(result.Passes) {
        result.Passes[promptIndex] = pass
    }
    results[model] = result
    writeResults(results)

    w.Write([]byte("OK"))
}

// Handle reset results
func resetResultsHandler(w http.ResponseWriter, r *http.Request) {
    emptyResults := make(map[string]Result)
    writeResults(emptyResults)
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
			http.Error(w, "Invalid CSV format", http.StatusBadRequest)
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
				passes = append(passes, make([]bool, len(prompts) - len(passes))...)
			}
			results[model] = Result{Passes: passes}
		}
		writeResults(results)
		http.Redirect(w, r, "/results", http.StatusSeeOther)
	} else {
		t, _ := template.ParseFiles("templates/import_results.html")
		t.Execute(w, nil)
	}
}
