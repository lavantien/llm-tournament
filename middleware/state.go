package middleware

import (
	"database/sql"
	"fmt"
	"log"
)

var (
	rowsErr  = func(rows *sql.Rows) error { return rows.Err() }
	txCommit = func(tx *sql.Tx) error { return tx.Commit() }
)

type Prompt struct {
	ID       int    `json:"-"`
	Text     string `json:"text"`
	Solution string `json:"solution"`
	Profile  string `json:"profile"`
}

type Result struct {
	Scores []int `json:"scores"`
}

type Profile struct {
	ID          int    `json:"-"`
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

	rows, err := db.Query("SELECT id, name, description FROM profiles WHERE suite_id = ? ORDER BY id", suiteID)
	if err != nil {
		return nil, fmt.Errorf("failed to query profiles: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var profiles []Profile
	for rows.Next() {
		var p Profile
		if err := rows.Scan(&p.ID, &p.Name, &p.Description); err != nil {
			return nil, fmt.Errorf("failed to scan profile: %w", err)
		}
		profiles = append(profiles, p)
	}

	return profiles, nil
}

// Write profile suite to database. Surviving rows (matched by Profile.ID)
// are updated in place so prompts keep pointing at them; rows absent from
// profiles are deleted, and profiles carrying no current ID (fresh or stale)
// are inserted.
func WriteProfileSuite(suiteName string, profiles []Profile) error {
	suiteID, err := GetSuiteID(suiteName)
	if err != nil {
		return fmt.Errorf("failed to get suite ID: %w", err)
	}

	// Begin transaction
	tx, err := dbBegin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	existing, err := suiteRowIDs(tx, "SELECT id FROM profiles WHERE suite_id = ?", suiteID)
	if err != nil {
		return fmt.Errorf("failed to load profile IDs: %w", err)
	}

	kept := make(map[int]bool)
	for _, profile := range profiles {
		if profile.ID != 0 && existing[profile.ID] {
			kept[profile.ID] = true
		}
	}

	// Delete removed rows before inserting, so a fresh profile can reuse the
	// name of a replaced one without tripping UNIQUE(name, suite_id).
	for id := range existing {
		if !kept[id] {
			if _, err = tx.Exec("DELETE FROM profiles WHERE id = ?", id); err != nil {
				return fmt.Errorf("failed to delete removed profile: %w", err)
			}
		}
	}

	// Update surviving profiles and insert new ones
	if len(profiles) > 0 {
		stmt, err := tx.Prepare("INSERT INTO profiles (name, description, suite_id) VALUES (?, ?, ?)")
		if err != nil {
			return fmt.Errorf("failed to prepare profile insert: %w", err)
		}
		defer func() { _ = stmt.Close() }()

		for _, profile := range profiles {
			if kept[profile.ID] {
				_, err = tx.Exec(`
				UPDATE profiles
				SET name = ?, description = ?
				WHERE id = ? AND suite_id = ?
				`, profile.Name, profile.Description, profile.ID, suiteID)
				if err != nil {
					return fmt.Errorf("failed to update profile: %w", err)
				}
				continue
			}

			_, err = stmt.Exec(profile.Name, profile.Description, suiteID)
			if err != nil {
				return fmt.Errorf("failed to insert profile: %w", err)
			}
		}
	}

	return tx.Commit()
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
	defer func() { _ = modelRows.Close() }()

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
		_ = scoreRows.Close()

		results[modelName] = Result{Scores: scores}
	}

	return results
}

// Read prompt suite from database
func ReadPromptSuite(suiteName string) ([]Prompt, error) {
	suiteID, err := GetSuiteID(suiteName)
	if err != nil {
		return nil, fmt.Errorf("failed to get suite ID: %w", err)
	}

	// Query to get prompts with profile names - ensure distinct results
	query := `
	SELECT p.id, p.text, p.solution, COALESCE(pr.name, '') as profile_name, p.display_order
	FROM prompts p
	LEFT JOIN profiles pr ON p.profile_id = pr.id
	WHERE p.suite_id = ?
	ORDER BY p.display_order
	`

	rows, err := db.Query(query, suiteID)
	if err != nil {
		return nil, fmt.Errorf("failed to query prompts: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var prompts []Prompt
	seenTexts := make(map[string]bool) // Track unique prompts by text content

	for rows.Next() {
		var p Prompt
		var displayOrder int
		if err := rows.Scan(&p.ID, &p.Text, &p.Solution, &p.Profile, &displayOrder); err != nil {
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
	if err = rowsErr(rows); err != nil {
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

// suiteRowIDs returns the set of row ids produced by a suite-scoped id query.
// The suite writers use it to tell surviving rows (updated in place) from
// removed rows (deleted so cascades clean up only their dependents).
func suiteRowIDs(tx *sql.Tx, query string, suiteID int) (map[int]bool, error) {
	rows, err := tx.Query(query, suiteID)
	if err != nil {
		return nil, fmt.Errorf("failed to query row IDs: %w", err)
	}
	defer func() { _ = rows.Close() }()

	ids := make(map[int]bool)
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan row ID: %w", err)
		}
		ids[id] = true
	}
	if err := rowsErr(rows); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}
	return ids, nil
}

// Write prompt suite to database. Surviving rows (matched by Prompt.ID) are
// updated in place so cascaded scores and model responses survive the write.
// Rows absent from prompts are deleted, and prompts carrying no current ID
// (fresh or stale) are inserted.
func WritePromptSuite(suiteName string, prompts []Prompt) error {
	suiteID, err := GetSuiteID(suiteName)
	if err != nil {
		return fmt.Errorf("failed to get suite ID: %w", err)
	}

	// Begin transaction
	tx, err := dbBegin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	existing, err := suiteRowIDs(tx, "SELECT id FROM prompts WHERE suite_id = ?", suiteID)
	if err != nil {
		return fmt.Errorf("failed to load prompt IDs: %w", err)
	}

	kept := make(map[int]bool)
	for _, prompt := range prompts {
		if prompt.ID != 0 && existing[prompt.ID] {
			kept[prompt.ID] = true
		}
	}

	// Delete removed rows before inserting, so a fresh prompt can reuse the
	// text of a replaced one without tripping UNIQUE(text, suite_id).
	for id := range existing {
		if !kept[id] {
			if _, err = tx.Exec("DELETE FROM prompts WHERE id = ?", id); err != nil {
				return fmt.Errorf("failed to delete removed prompt: %w", err)
			}
		}
	}

	// Update surviving prompts and insert new ones
	if len(prompts) > 0 {
		stmt, err := tx.Prepare(`
		INSERT INTO prompts (text, solution, profile_id, suite_id, display_order)
		VALUES (?, ?, ?, ?, ?)
		`)
		if err != nil {
			return fmt.Errorf("failed to prepare prompt insert: %w", err)
		}
		defer func() { _ = stmt.Close() }()

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

			if kept[prompt.ID] {
				_, err = tx.Exec(`
				UPDATE prompts
				SET text = ?, solution = ?, profile_id = ?, display_order = ?
				WHERE id = ? AND suite_id = ?
				`, prompt.Text, prompt.Solution, profileID, i, prompt.ID, suiteID)
				if err != nil {
					return fmt.Errorf("failed to update prompt: %w", err)
				}
				continue
			}

			_, err = stmt.Exec(prompt.Text, prompt.Solution, profileID, suiteID, i)
			if err != nil {
				return fmt.Errorf("failed to insert prompt: %w", err)
			}
		}
	}

	return tx.Commit()
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
			// Same recovery as GetCurrentSuiteID so the default suite is
			// actually created and both resolvers agree on one current row.
			_, name, err := recoverDefaultSuite()
			if err != nil {
				log.Printf("Error recovering default suite: %v", err)
				return ""
			}
			return name
		}
		log.Printf("Error getting current suite name: %v", err)
		return ""
	}
	return name
}

// Write results to database
func WriteResults(suiteName string, results map[string]Result) (err error) {
	suiteID, err := GetSuiteID(suiteName)
	if err != nil {
		return fmt.Errorf("failed to get suite ID: %w", err)
	}

	// Begin transaction
	tx, err := dbBegin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
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
			_ = promptRows.Close()
			return fmt.Errorf("failed to scan prompt ID: %w", err)
		}
		promptIDs = append(promptIDs, id)
	}
	_ = promptRows.Close()

	if err := rowsErr(promptRows); err != nil {
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
			_ = modelNamesRows.Close()
			return fmt.Errorf("failed to scan model name: %w", err)
		}
		dbModelNames = append(dbModelNames, name)
	}
	_ = modelNamesRows.Close()

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
			modelIDInt64, err := lastInsertID(modelResult)
			if err != nil {
				return fmt.Errorf("failed to get model ID: %w", err)
			}
			modelID = int(modelIDInt64)
		} else if err != nil {
			return fmt.Errorf("failed to query model: %w", err)
		}

		// Insert scores
		if len(result.Scores) > 0 {
			if dropped := len(result.Scores) - len(promptIDs); dropped > 0 {
				log.Printf("Warning: dropped %d score(s) for model %q: suite has only %d prompt(s)", dropped, modelName, len(promptIDs))
			}
			scoreStmt, err := tx.Prepare("INSERT INTO scores (model_id, prompt_id, score) VALUES (?, ?, ?)")
			if err != nil {
				return fmt.Errorf("failed to prepare score insert: %w", err)
			}

			for i, score := range result.Scores {
				if i < len(promptIDs) {
					_, err = scoreStmt.Exec(modelID, promptIDs[i], score)
					if err != nil {
						_ = scoreStmt.Close()
						return fmt.Errorf("failed to insert score: %w", err)
					}
				}
			}
			_ = scoreStmt.Close()
		}
	}

	return tx.Commit()
}

// RenameModel renames a model within a suite in place. Keeping the row id is
// the point: model_responses cascade on model_id, so a rename done by delete
// and reinsert would destroy every saved response.
func RenameModel(suiteName, oldName, newName string) error {
	suiteID, err := GetSuiteID(suiteName)
	if err != nil {
		return fmt.Errorf("failed to get suite ID: %w", err)
	}

	if _, err := db.Exec("UPDATE models SET name = ? WHERE name = ? AND suite_id = ?", newName, oldName, suiteID); err != nil {
		return fmt.Errorf("failed to rename model: %w", err)
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
	if !isPermutationOfIndices(order) {
		log.Println("Invalid order: values must be a permutation of 0..n-1 without duplicates")
		return
	}

	// Begin transaction
	tx, err := dbBegin()
	if err != nil {
		log.Printf("Error beginning transaction: %v", err)
		return
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// Update each prompt's display_order. The prompts slice is ordered by
	// display_order and the permutation check guarantees every index is in
	// range. Updating by the row ID read with the prompt keeps the row
	// identity, and with it the cascaded scores and responses.
	for newOrder, oldIndex := range order {
		_, err = tx.Exec("UPDATE prompts SET display_order = ? WHERE id = ?", newOrder, prompts[oldIndex].ID)
		if err != nil {
			log.Printf("Error updating prompt order: %v", err)
			return
		}
	}

	if err = txCommit(tx); err != nil {
		log.Printf("Error committing transaction: %v", err)
		return
	}

	log.Println("Prompts order updated successfully")
	BroadcastResults()
}

// isPermutationOfIndices reports whether order contains every index in
// 0..len(order)-1 exactly once. A duplicate or omitted index would leave two
// prompts sharing a display_order and corrupt ReadResults score indexing.
func isPermutationOfIndices(order []int) bool {
	seen := make([]bool, len(order))
	for _, index := range order {
		if index < 0 || index >= len(order) || seen[index] {
			return false
		}
		seen[index] = true
	}
	return true
}
