package main

import (
	"encoding/json"
	"os"
)

type Prompt struct {
	Text string `json:"text"`
}

type Result struct {
	Passes []bool `json:"passes"`
}

// Read prompts from prompts.json
func readPrompts() []Prompt {
	data, _ := os.ReadFile("data/prompts.json")
	var prompts []Prompt
	json.Unmarshal(data, &prompts)
	return prompts
}

// Write prompts to prompts.json
func writePrompts(prompts []Prompt) error {
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
func readResults() map[string]Result {
    data, _ := os.ReadFile("data/results.json")
    var results map[string]Result
    json.Unmarshal(data, &results)
    return results
}

// Write results to results.json
func writeResults(results map[string]Result) error {
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
