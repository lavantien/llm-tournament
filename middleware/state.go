package middleware

import (
	"encoding/json"
	"log"
	"os"
	"strings"
)

type Prompt struct {
	Text     string `json:"text"`
	Solution string `json:"solution"`
}

type Result struct {
	Passes []bool `json:"passes"`
}

// Read prompts from prompts.json
func ReadPrompts() []Prompt {
	suiteName := GetCurrentSuiteName()
	prompts, _ := ReadPromptSuite(suiteName)
	return prompts
}

// Write prompts to prompts.json
func WritePrompts(prompts []Prompt) error {
	suiteName := GetCurrentSuiteName()
	return WritePromptSuite(suiteName, prompts)
}

// Read results from data/results-<suiteName>.json
func ReadResults() map[string]Result {
	suiteName := GetCurrentSuiteName()
	var filename string
	if suiteName == "default" {
		filename = "data/results-default.json"
	} else {
		filename = "data/results-" + suiteName + ".json"
	}
	data, _ := os.ReadFile(filename)
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

// Read prompt suite from data/prompts-<suiteName>.json
func ReadPromptSuite(suiteName string) ([]Prompt, error) {
	var filename string
	if suiteName == "default" {
		filename = "data/prompts-default.json"
	} else {
		filename = "data/prompts-" + suiteName + ".json"
	}
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var prompts []Prompt
	err = json.Unmarshal(data, &prompts)
	if err != nil {
		return nil, err
	}
	return prompts, nil
}

// Write prompt suite to data/prompts-<suiteName>.json
func WritePromptSuite(suiteName string, prompts []Prompt) error {
	var filename string
	if suiteName == "default" {
		filename = "data/prompts-default.json"
	} else {
		filename = "data/prompts-" + suiteName + ".json"
	}
	data, err := json.Marshal(prompts)
	if err != nil {
		return err
	}
	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return err
	}
	return nil
}

// List all prompt suites
func ListPromptSuites() ([]string, error) {
	var suites []string
	files, err := os.ReadDir("data")
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		if !file.IsDir() && strings.HasPrefix(file.Name(), "prompts-") && strings.HasSuffix(file.Name(), ".json") {
			suiteName := strings.TrimSuffix(strings.TrimPrefix(file.Name(), "prompts-"), ".json")
			suites = append(suites, suiteName)
		}
	}
	return suites, nil
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
