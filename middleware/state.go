package middleware

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
)

type Prompt struct {
	Text string `json:"text"`
}

type Result struct {
	Passes []bool `json:"passes"`
}

// Read prompts from prompts.json
func ReadPrompts() []Prompt {
	data, _ := os.ReadFile("data/prompts.json")
	var prompts []Prompt
	json.Unmarshal(data, &prompts)
	return prompts
}

// Write prompts to prompts.json
func WritePrompts(prompts []Prompt) error {
	data, err := json.Marshal(prompts)
	if err != nil {
		return err
	}
	err = os.WriteFile("data/prompts.json", data, 0644)
	if err != nil {
		return err
	}
	return nil
}

// Read results from results.json
func ReadResults() map[string]Result {
	data, _ := os.ReadFile("data/results.json")
	var results map[string]Result
	json.Unmarshal(data, &results)

	prompts := ReadPrompts()
	if results == nil {
		return make(map[string]Result)
	}
	for model, result := range results {
		if len(result.Passes) < len(prompts) {
			result.Passes = append(result.Passes, make([]bool, len(prompts)-len(result.Passes))...)
			results[model] = result
		} else if len(result.Passes) > len(prompts) {
			results[model] = Result{Passes: result.Passes[:len(prompts)]}
		}
	}
	return results
}

// Write results to results.json
func WriteResults(results map[string]Result) error {
	data, err := json.Marshal(results)
	if err != nil {
		return err
	}
	err = os.WriteFile("data/results.json", data, 0644)
	if err != nil {
		return err
	}
	return nil
}

// Handle import error
func ImportErrorHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling import error")
	t, err := template.ParseFiles("templates/import_error.html")
	if err != nil {
		http.Error(w, "Error parsing template: "+err.Error(), http.StatusInternalServerError)
		return
	}
	err = t.Execute(w, nil)
	if err != nil {
		http.Error(w, "Error executing template: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func UpdatePromptsOrder(order []int) {
	prompts := ReadPrompts()
	if len(order) != len(prompts) {
		log.Println("Invalid order length")
		return
	}
	orderedPrompts := make([]Prompt, len(prompts))
	for i, index := range order {
		if index < 0 || index >= len(prompts) {
			log.Println("Invalid index in order")
			return
		}
		orderedPrompts[i] = prompts[index]
	}
	err := WritePrompts(orderedPrompts)
	if err != nil {
		log.Printf("Error writing prompts: %v", err)
		return
	}
	log.Println("Prompts order updated successfully")
	BroadcastResults()
}
