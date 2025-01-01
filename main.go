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
	}
}

// Read prompts from prompts.txt
func readPrompts() []string {
	data, _ := os.ReadFile("data/prompts.txt")
	prompts := strings.Split(string(data), "\n")
	return prompts
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
	// Update the result in results.csv
	// Recalculate total scores and sort
	// Send back the updated scores
	w.Write([]byte("OK"))
}
