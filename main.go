package main

import (
	"html/template"
	"net/http"
	"os"
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
	} else {
		http.Redirect(w, r, "/prompts", http.StatusSeeOther)
	} else if r.URL.Path == "/add_prompt" {
		addPromptHandler(w, r)
	} else if r.URL.Path == "/edit_prompt" {
		editPromptHandler(w, r)
	} else if r.URL.Path == "/delete_prompt" {
		deletePromptHandler(w, r)
	}
	else {
		http.Redirect(w, r, "/prompts", http.StatusSeeOther)
	}
}

// Read prompts from prompts.txt
func readPrompts() []string {
	data, _ := os.ReadFile("data/prompts.txt")
	prompts := strings.Split(string(data), "\n")
	return prompts
}

// Write prompts to prompts.txt
func writePrompts(prompts []string) {
	data := strings.Join(prompts, "\n")
	os.WriteFile("data/prompts.txt", []byte(data), 0644)
}

// Handle add prompt
func addPromptHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	prompt := r.Form.Get("prompt")
	prompts := readPrompts()
	prompts = append(prompts, prompt)
	writePrompts(prompts)
	http.Redirect(w, r, "/prompts", http.StatusSeeOther)
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
                Prompt: prompts[index],
            })
        } else {
            http.Redirect(w, r, "/prompts", http.StatusSeeOther)
        }
    } else if r.Method == "POST" {
        r.ParseForm()
        indexStr := r.Form.Get("index")
        index, _ := strconv.Atoi(indexStr)
        editedPrompt := r.Form.Get("prompt")
        prompts := readPrompts()
		if index >= 0 && index < len(prompts) {
			prompts[index] = editedPrompt
		}
        writePrompts(prompts)
        http.Redirect(w, r, "/prompts", http.StatusSeeOther)
    }
}

// Handle delete prompt
func deletePromptHandler(w http.ResponseWriter, r *http.Request) {
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

// Read results from results.csv
func readResults() map[string][]bool {
	data, _ := os.ReadFile("data/results.csv")
	lines := strings.Split(string(data), "\n")
	results := make(map[string][]bool)
	prompts := readPrompts()
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) != len(prompts)+1 {
			continue
		}
		model := parts[0]
		var passes []bool
		for _, passStr := range parts[1:] {
			passes = append(passes, passStr == "true")
		}
		results[model] = passes
	}
	return results
}

// Write results to results.csv
func writeResults(results map[string][]bool) {
	var lines []string
	for model, passes := range results {
		line := model
		for _, pass := range passes {
			line += "," + strconv.FormatBool(pass)
		}
		lines = append(lines, line)
	}
	data := strings.Join(lines, "\n")
	os.WriteFile("data/results.csv", []byte(data), 0644)
}

// Handle prompt list page
func promptListHandler(w http.ResponseWriter, r *http.Request) {
	prompts := readPrompts()
	t, _ := template.ParseFiles("templates/prompt_list.html")
	t.Execute(w, prompts)
}

// Handle results page
func resultsHandler(w http.ResponseWriter, r *http.Request) {
	prompts := readPrompts()
	results := readResults()
	// Sort models by total score
	models := make([]string, 0)
	for model := range results {
		models = append(models, model)
	}
	// Calculate total scores and sort
	t, _ := template.ParseFiles("templates/results.html")
	t.Execute(w, struct {
		Prompts  []string
		Results  map[string][]bool
		Models   []string
	}{
		Prompts:  prompts,
		Results:  results,
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
        results[model] = make([]bool, len(readPrompts()))
    }
    if promptIndex >= 0 && promptIndex < len(results[model]) {
        results[model][promptIndex] = pass
    }
    writeResults(results)

	w.Write([]byte("OK"))
}
