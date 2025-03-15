package middleware

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
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

// Read profiles from database for current suite
func ReadProfiles() []Profile {
	suiteName := GetCurrentSuiteName()
	profiles, _ := ReadProfileSuite(suiteName)
	return profiles
}

// Write profiles to database
func WriteProfiles(profiles []Profile) error {
	suiteName := GetCurrentSuiteName()
	return WriteProfileSuite(suiteName, profiles)
}

// Read profile suite from database
func ReadProfileSuite(suiteName string) ([]Profile, error) {
	suiteID, err := GetSuiteID(suiteName)
	if err != nil {
		return nil, fmt.Errorf("failed to get suite ID: %w", err)
	}

	rows, err := db.Query("SELECT name, description FROM profiles WHERE suite_id = ? ORDER BY id", suiteID)
	if err != nil {
		return nil, fmt.Errorf("failed to query profiles: %w", err)
	}
	defer rows.Close()

	var profiles []Profile
	for rows.Next() {
		var p Profile
		if err := rows.Scan(&p.Name, &p.Description); err != nil {
			return nil, fmt.Errorf("failed to scan profile: %w", err)
		}
		profiles = append(profiles, p)
	}

	return profiles, nil
}

// Write profile suite to database
func WriteProfileSuite(suiteName string, profiles []Profile) error {
	suiteID, err := GetSuiteID(suiteName)
	if err != nil {
		return fmt.Errorf("failed to get suite ID: %w", err)
	}

	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Delete existing profiles for this suite
	_, err = tx.Exec("DELETE FROM profiles WHERE suite_id = ?", suiteID)
	if err != nil {
		return fmt.Errorf("failed to delete profiles: %w", err)
	}

	// Insert new profiles
	if len(profiles) > 0 {
		stmt, err := tx.Prepare("INSERT INTO profiles (name, description, suite_id) VALUES (?, ?, ?)")
		if err != nil {
			return fmt.Errorf("failed to prepare profile insert: %w", err)
		}
		defer stmt.Close()

		for _, profile := range profiles {
			_, err = stmt.Exec(profile.Name, profile.Description, suiteID)
			if err != nil {
				return fmt.Errorf("failed to insert profile: %w", err)
			}
		}
	}

	return tx.Commit()
}

// Read profile suite from JSON file (for migration)
func ReadProfileSuiteFromJSON(suiteName string) ([]Profile, error) {
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

// List all profile suites
func ListProfileSuites() ([]string, error) {
	return ListSuites()
}

func DeleteProfileSuite(suiteName string) error {
	return DeleteSuite(suiteName)
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

// Read results from database
func ReadResults() map[string]Result {
	suiteName := GetCurrentSuiteName()
	var suiteID int
	err := db.QueryRow("SELECT id FROM suites WHERE name = ?", suiteName).Scan(&suiteID)
	if err != nil {
		log.Printf("Error getting suite ID: %v", err)
		return make(map[string]Result)
	}

	// Get all prompts for this suite
	promptCountQuery := "SELECT COUNT(*) FROM prompts WHERE suite_id = ?"
	var promptCount int
	err = db.QueryRow(promptCountQuery, suiteID).Scan(&promptCount)
	if err != nil {
		log.Printf("Error counting prompts: %v", err)
		return make(map[string]Result)
	}

	// Get all models for this suite
	modelQuery := "SELECT id, name FROM models WHERE suite_id = ?"
	modelRows, err := db.Query(modelQuery, suiteID)
	if err != nil {
		log.Printf("Error querying models: %v", err)
		return make(map[string]Result)
	}
	defer modelRows.Close()

	results := make(map[string]Result)
	for modelRows.Next() {
		var modelID int
		var modelName string
		if err := modelRows.Scan(&modelID, &modelName); err != nil {
			log.Printf("Error scanning model: %v", err)
			continue
		}

		// Initialize scores array
		scores := make([]int, promptCount)

		// Get scores for this model
		scoreQuery := `
		SELECT p.display_order, s.score
		FROM scores s
		JOIN prompts p ON s.prompt_id = p.id
		WHERE s.model_id = ? AND p.suite_id = ?
		ORDER BY p.display_order
		`
		scoreRows, err := db.Query(scoreQuery, modelID, suiteID)
		if err != nil {
			log.Printf("Error querying scores: %v", err)
			continue
		}

		for scoreRows.Next() {
			var promptOrder, score int
			if err := scoreRows.Scan(&promptOrder, &score); err != nil {
				log.Printf("Error scanning score: %v", err)
				continue
			}
			
			if promptOrder >= 0 && promptOrder < promptCount {
				scores[promptOrder] = score
			}
		}
		scoreRows.Close()

		results[modelName] = Result{Scores: scores}
	}

	return results
}

// Read results from JSON file (for migration)
func ReadResultsFromJSON(suiteName string) (map[string]Result, error) {
	var filename string
	if suiteName == "default" {
		filename = "data/results-default.json"
	} else {
		filename = "data/results-" + suiteName + ".json"
	}
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var results map[string]Result
	err = json.Unmarshal(data, &results)
	if err != nil {
		return nil, err
	}
	return results, nil
}

// Read prompt suite from database
func ReadPromptSuite(suiteName string) ([]Prompt, error) {
	suiteID, err := GetSuiteID(suiteName)
	if err != nil {
		return nil, fmt.Errorf("failed to get suite ID: %w", err)
	}

	// Query to get prompts with profile names - ensure distinct results
	query := `
	SELECT p.text, p.solution, COALESCE(pr.name, '') as profile_name, p.display_order
	FROM prompts p
	LEFT JOIN profiles pr ON p.profile_id = pr.id
	WHERE p.suite_id = ?
	ORDER BY p.display_order
	`

	rows, err := db.Query(query, suiteID)
	if err != nil {
		return nil, fmt.Errorf("failed to query prompts: %w", err)
	}
	defer rows.Close()

	var prompts []Prompt
	seenTexts := make(map[string]bool) // Track unique prompts by text content
	
	for rows.Next() {
		var p Prompt
		var displayOrder int
		if err := rows.Scan(&p.Text, &p.Solution, &p.Profile, &displayOrder); err != nil {
			return nil, fmt.Errorf("failed to scan prompt: %w", err)
		}
		
		// Ensure we don't add duplicates
		if !seenTexts[p.Text] {
			prompts = append(prompts, p)
			seenTexts[p.Text] = true
		} else {
			log.Printf("Warning: Skipped duplicate prompt with text: %s", p.Text[:min(20, len(p.Text))])
		}
	}

	// Check for any errors during iteration
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating prompt rows: %w", err)
	}

	return prompts, nil
}

// min returns the smaller of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Write prompt suite to database
func WritePromptSuite(suiteName string, prompts []Prompt) error {
	suiteID, err := GetSuiteID(suiteName)
	if err != nil {
		return fmt.Errorf("failed to get suite ID: %w", err)
	}

	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Delete existing prompts for this suite
	_, err = tx.Exec("DELETE FROM prompts WHERE suite_id = ?", suiteID)
	if err != nil {
		return fmt.Errorf("failed to delete prompts: %w", err)
	}

	// Insert new prompts
	if len(prompts) > 0 {
		stmt, err := tx.Prepare(`
		INSERT INTO prompts (text, solution, profile_id, suite_id, display_order) 
		VALUES (?, ?, ?, ?, ?)
		`)
		if err != nil {
			return fmt.Errorf("failed to prepare prompt insert: %w", err)
		}
		defer stmt.Close()

		for i, prompt := range prompts {
			// Get profile ID if a profile is specified
			var profileID sql.NullInt64
			if prompt.Profile != "" {
				id, exists, err := GetProfileID(prompt.Profile, suiteID)
				if err != nil {
					return fmt.Errorf("failed to get profile ID: %w", err)
				}
				if exists {
					profileID.Int64 = int64(id)
					profileID.Valid = true
				}
			}

			_, err = stmt.Exec(prompt.Text, prompt.Solution, profileID, suiteID, i)
			if err != nil {
				return fmt.Errorf("failed to insert prompt: %w", err)
			}
		}
	}

	return tx.Commit()
}

// Read prompt suite from JSON file (for migration)
func ReadPromptSuiteFromJSON(suiteName string) ([]Prompt, error) {
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

// List all prompt suites
func ListPromptSuites() ([]string, error) {
	return ListSuites()
}

func DeletePromptSuite(suiteName string) error {
	return DeleteSuite(suiteName)
}

// RenameSuiteFiles renames all files associated with a suite
func RenameSuiteFiles(oldName, newName string) error {
	return RenameSuite(oldName, newName)
}

// SuiteExists checks if a suite with the given name exists
func SuiteExists(name string) bool {
	var exists int
	err := db.QueryRow("SELECT 1 FROM suites WHERE name = ?", name).Scan(&exists)
	return err == nil
}

// Get current suite name
func GetCurrentSuiteName() string {
	var name string
	err := db.QueryRow("SELECT name FROM suites WHERE is_current = 1").Scan(&name)
	if err != nil {
		if err == sql.ErrNoRows {
			// Set default suite as current if none is set
			_, err = db.Exec("UPDATE suites SET is_current = 1 WHERE name = 'default'")
			if err != nil {
				log.Printf("Error setting default suite as current: %v", err)
				return ""
			}
			return "default"
		}
		log.Printf("Error getting current suite name: %v", err)
		return ""
	}
	return name
}

// Write results to database
func WriteResults(suiteName string, results map[string]Result) error {
	suiteID, err := GetSuiteID(suiteName)
	if err != nil {
		return fmt.Errorf("failed to get suite ID: %w", err)
	}

	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Get all prompt IDs for this suite
	promptRows, err := tx.Query("SELECT id FROM prompts WHERE suite_id = ? ORDER BY display_order", suiteID)
	if err != nil {
		return fmt.Errorf("failed to query prompts: %w", err)
	}
	
	var promptIDs []int
	for promptRows.Next() {
		var id int
		if err := promptRows.Scan(&id); err != nil {
			promptRows.Close()
			return fmt.Errorf("failed to scan prompt ID: %w", err)
		}
		promptIDs = append(promptIDs, id)
	}
	promptRows.Close()
	
	if err := promptRows.Err(); err != nil {
		return fmt.Errorf("error iterating prompt rows: %w", err)
	}

	// Get current model names in the database
	modelNamesRows, err := tx.Query("SELECT name FROM models WHERE suite_id = ?", suiteID)
	if err != nil {
		return fmt.Errorf("failed to query model names: %w", err)
	}
	
	var dbModelNames []string
	for modelNamesRows.Next() {
		var name string
		if err := modelNamesRows.Scan(&name); err != nil {
			modelNamesRows.Close()
			return fmt.Errorf("failed to scan model name: %w", err)
		}
		dbModelNames = append(dbModelNames, name)
	}
	modelNamesRows.Close()
	
	// Delete models that are in the database but not in the results map
	for _, dbModelName := range dbModelNames {
		if _, exists := results[dbModelName]; !exists {
			_, err = tx.Exec("DELETE FROM models WHERE name = ? AND suite_id = ?", dbModelName, suiteID)
			if err != nil {
				return fmt.Errorf("failed to delete model: %w", err)
			}
		}
	}

	// Clear existing scores for this suite
	_, err = tx.Exec(`
		DELETE FROM scores 
		WHERE model_id IN (SELECT id FROM models WHERE suite_id = ?)
	`, suiteID)
	if err != nil {
		return fmt.Errorf("failed to delete scores: %w", err)
	}

	// Process each model
	for modelName, result := range results {
		// Get or create model
		var modelID int
		err := tx.QueryRow("SELECT id FROM models WHERE name = ? AND suite_id = ?", modelName, suiteID).Scan(&modelID)
		if err == sql.ErrNoRows {
			// Create new model
			modelResult, err := tx.Exec("INSERT INTO models (name, suite_id) VALUES (?, ?)", modelName, suiteID)
			if err != nil {
				return fmt.Errorf("failed to insert model: %w", err)
			}
			modelIDInt64, err := modelResult.LastInsertId()
			if err != nil {
				return fmt.Errorf("failed to get model ID: %w", err)
			}
			modelID = int(modelIDInt64)
		} else if err != nil {
			return fmt.Errorf("failed to query model: %w", err)
		}

		// Insert scores
		if len(result.Scores) > 0 {
			scoreStmt, err := tx.Prepare("INSERT INTO scores (model_id, prompt_id, score) VALUES (?, ?, ?)")
			if err != nil {
				return fmt.Errorf("failed to prepare score insert: %w", err)
			}
			
			for i, score := range result.Scores {
				if i < len(promptIDs) {
					_, err = scoreStmt.Exec(modelID, promptIDs[i], score)
					if err != nil {
						scoreStmt.Close()
						return fmt.Errorf("failed to insert score: %w", err)
					}
				}
			}
			scoreStmt.Close()
		}
	}

	return tx.Commit()
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
	
	suiteID, err := GetCurrentSuiteID()
	if err != nil {
		log.Printf("Error getting current suite ID: %v", err)
		return
	}
	
	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		log.Printf("Error beginning transaction: %v", err)
		return
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Get all prompt IDs for this suite
	promptRows, err := tx.Query("SELECT id FROM prompts WHERE suite_id = ? ORDER BY display_order", suiteID)
	if err != nil {
		log.Printf("Error querying prompts: %v", err)
		return
	}
	
	var promptIDs []int
	for promptRows.Next() {
		var id int
		if err := promptRows.Scan(&id); err != nil {
			promptRows.Close()
			log.Printf("Error scanning prompt ID: %v", err)
			return
		}
		promptIDs = append(promptIDs, id)
	}
	promptRows.Close()
	
	// Update each prompt's display_order
	for newOrder, oldIndex := range order {
		if oldIndex < 0 || oldIndex >= len(promptIDs) {
			log.Println("Invalid index in order")
			return
		}
		
		_, err = tx.Exec("UPDATE prompts SET display_order = ? WHERE id = ?", newOrder, promptIDs[oldIndex])
		if err != nil {
			log.Printf("Error updating prompt order: %v", err)
			return
		}
	}
	
	if err = tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		return
	}
	
	log.Println("Prompts order updated successfully")
	BroadcastResults()
}
