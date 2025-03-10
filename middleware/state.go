package middleware

import (
	"encoding/json"
	"fmt"
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
	Scores []int `json:"scores"`
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
	
	// Migrate any old format results
	results = MigrateResults(results)
	for model, result := range results {
		// Ensure Scores array is correct length
		if result.Scores == nil {
			result.Scores = make([]int, len(prompts))
		} else if len(result.Scores) < len(prompts) {
			result.Scores = append(result.Scores, make([]int, len(prompts)-len(result.Scores))...)
		} else if len(result.Scores) > len(prompts) {
			result.Scores = result.Scores[:len(prompts)]
		}
		
		results[model] = result
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

// RenameSuiteFiles renames all files associated with a suite
func RenameSuiteFiles(oldName, newName string) error {
    // Add validation
    if oldName == "" {
        return fmt.Errorf("original suite name cannot be empty")
    }
    if newName == "" {
        return fmt.Errorf("new suite name cannot be empty")
    }
    if strings.ContainsAny(newName, "/\\") {
        return fmt.Errorf("suite name contains invalid characters")
    }
    if SuiteExists(newName) {
        return fmt.Errorf("suite with name '%s' already exists", newName)
    }

    // Rename prompts file
    oldPrompts := "data/prompts-" + oldName + ".json"
    newPrompts := "data/prompts-" + newName + ".json"
    if err := os.Rename(oldPrompts, newPrompts); err != nil && !os.IsNotExist(err) {
        return fmt.Errorf("error renaming prompts: %w", err)
    }

    // Rename profiles file
    oldProfiles := "data/profiles-" + oldName + ".json"
    newProfiles := "data/profiles-" + newName + ".json"
    if err := os.Rename(oldProfiles, newProfiles); err != nil && !os.IsNotExist(err) {
        return fmt.Errorf("error renaming profiles: %w", err)
    }

    // Rename results file
    oldResults := "data/results-" + oldName + ".json"
    newResults := "data/results-" + newName + ".json"
    if err := os.Rename(oldResults, newResults); err != nil && !os.IsNotExist(err) {
        return fmt.Errorf("error renaming results: %w", err)
    }

    // Update current suite if it was the renamed one
    if GetCurrentSuiteName() == oldName {
        if err := os.WriteFile("data/current_suite.txt", []byte(newName), 0644); err != nil {
            return fmt.Errorf("error updating current suite: %w", err)
        }
    }
    
    return nil
}

// SuiteExists checks if a suite with the given name exists
func SuiteExists(name string) bool {
    suites, _ := ListPromptSuites()
    for _, s := range suites {
        if s == name {
            return true
        }
    }
    return false
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


	return nil
}


// MigrateResults converts old result formats to the current format
func MigrateResults(results map[string]Result) map[string]Result {
	migrated := make(map[string]Result)
	prompts := ReadPrompts()
	
	for model, result := range results {
		// If we have no Scores, initialize empty array
		if result.Scores == nil {
			result.Scores = make([]int, len(prompts))
		} else if len(result.Scores) < len(prompts) {
			// Ensure scores array has correct length
			newScores := make([]int, len(prompts))
			copy(newScores, result.Scores)
			result.Scores = newScores
		}
		
		// Ensure all scores are within valid range
		for i, score := range result.Scores {
			if score < 0 || score > 100 {
				result.Scores[i] = 0
			}
		}
		
		migrated[model] = result
	}
	return migrated
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
