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
	Profile  string `json:"profile"`
}

type Result struct {
	Passes []bool `json:"passes"`
	Scores []int  `json:"scores"`
}

type Profile struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Read profiles from data/profiles-default.json
func ReadProfiles() []Profile {
	suiteName := GetCurrentSuiteName()
	profiles, _ := ReadProfileSuite(suiteName)
	return profiles
}

// Write profiles to data/profiles-default.json
func WriteProfiles(profiles []Profile) error {
	suiteName := GetCurrentSuiteName()
	return WriteProfileSuite(suiteName, profiles)
}

// Read profile suite from data/profiles-<suiteName>.json
func ReadProfileSuite(suiteName string) ([]Profile, error) {
	var filename string
	if suiteName == "default" {
		filename = "data/profiles-default.json"
	} else {
		filename = "data/profiles-" + suiteName + ".json"
	}
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var profiles []Profile
	err = json.Unmarshal(data, &profiles)
	if err != nil {
		return nil, err
	}
	return profiles, nil
}

// Write profile suite to data/profiles-<suiteName>.json
func WriteProfileSuite(suiteName string, profiles []Profile) error {
	var filename string
	if suiteName == "default" {
		filename = "data/profiles-default.json"
	} else {
		filename = "data/profiles-" + suiteName + ".json"
	}
	data, err := json.Marshal(profiles)
	if err != nil {
		return err
	}
	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return err
	}
	return nil
}

// List all profile suites
func ListProfileSuites() ([]string, error) {
	var suites []string
	files, err := os.ReadDir("data")
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		if !file.IsDir() && strings.HasPrefix(file.Name(), "profiles-") && strings.HasSuffix(file.Name(), ".json") {
			suiteName := strings.TrimSuffix(strings.TrimPrefix(file.Name(), "profiles-"), ".json")
			suites = append(suites, suiteName)
		}
	}
	return suites, nil
}

func DeleteProfileSuite(suiteName string) error {
	filename := "data/profiles-" + suiteName + ".json"
	err := os.Remove(filename)
	if err != nil {
		return err
	}
	return nil
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

func DeletePromptSuite(suiteName string) error {
	filename := "data/prompts-" + suiteName + ".json"
	err := os.Remove(filename)
	if err != nil {
		return err
	}
	return nil
}

// Get current suite name
func GetCurrentSuiteName() string {
	data, err := os.ReadFile("data/current_suite.txt")
	if err != nil {
		return ""
	}
	suiteName := string(data)
	return strings.TrimSpace(suiteName)
}

// Write results to data/results-<suiteName>.json
func WriteResults(suiteName string, results map[string]Result) error {
	var filename string
	if suiteName == "default" {
		filename = "data/results-default.json"
	} else {
		filename = "data/results-" + suiteName + ".json"
	}
	data, err := json.Marshal(results)
	if err != nil {
		return err
	}
	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return err
	}

	prompts := ReadPrompts()
	for model, result := range results {
		if len(result.Passes) < len(prompts) {
			result.Passes = append(result.Passes, make([]bool, len(prompts)-len(result.Passes))...)
			results[model] = result
		} else if len(result.Passes) > len(prompts) {
			results[model] = Result{Passes: result.Passes[:len(prompts)]}
		}
	}

	return nil
}

func migrateResults(results map[string]Result) {
	for model, result := range results {
		// If we have Passes but no Scores, migrate from Passes to Scores
		if len(result.Passes) > 0 && len(result.Scores) == 0 {
			result.Scores = boolToScore(result.Passes)
		}
		// If we have Scores but no Passes, migrate from Scores to Passes
		if len(result.Scores) > 0 && len(result.Passes) == 0 {
			result.Passes = scoreToBool(result.Scores)
		}
		results[model] = result
	}
}

func scoreToBool(scores []int) []bool {
	passes := make([]bool, len(scores))
	for i, score := range scores {
		passes[i] = score >= 50 // Consider 50+ as pass
	}
	return passes
}

func boolToScore(passes []bool) []int {
	scores := make([]int, len(passes))
	for i, pass := range passes {
		if pass {
			scores[i] = 100
		} else {
			scores[i] = 0
		}
	}
	return scores
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
