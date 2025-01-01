package main

import (
	"html/template"
	"net/http"
	"encoding/json"
	"os"
	"sort"
	"strconv"
)

type Prompt struct {
	Text string `json:"text"`
}

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
    } else {
		http.Redirect(w, r, "/prompts", http.StatusSeeOther)
	}
}

// Read prompts from prompts.json
func readPrompts() []Prompt {
	data, _ := os.ReadFile("data/prompts.json")
	var prompts []Prompt
	json.Unmarshal(data, &prompts)
	return prompts
}

// Write prompts to prompts.json
func writePrompts(prompts []Prompt) {
	data, _ := json.Marshal(prompts)
	os.WriteFile("data/prompts.json", data, 0644)
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

type Result struct {
	Passes []bool `json:"passes"`
}

// Read results from results.json
func readResults() map[string]Result {
    data, _ := os.ReadFile("data/results.json")
    var results map[string]Result
    json.Unmarshal(data, &results)
    return results
}

// Write results to results.json
func writeResults(results map[string]Result) {
    data, _ := json.Marshal(results)
    os.WriteFile("data/results.json", data, 0644)
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
	t.Execute(w, struct {
		Prompts  []string
		Results  map[string][]bool
		Models   []string
	}{
		Prompts:  promptTexts,
		Results:  resultsForTemplate,
		Models:   models,
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
