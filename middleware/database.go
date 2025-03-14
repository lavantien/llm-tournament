package middleware

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

// InitDB initializes the SQLite database connection
func InitDB(dbPath string) error {
	// Ensure data directory exists
	dataDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	var err error
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Set pragmas for better performance
	_, err = db.Exec(`PRAGMA journal_mode = WAL;
                     PRAGMA synchronous = NORMAL;
                     PRAGMA foreign_keys = ON;`)
	if err != nil {
		return fmt.Errorf("failed to set database pragmas: %w", err)
	}

	// Create schema if it doesn't exist
	if err = createTables(); err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	return nil
}

// CloseDB closes the database connection
func CloseDB() error {
	if db != nil {
		return db.Close()
	}
	return nil
}

// createTables creates all the necessary tables if they don't exist
func createTables() error {
	schema := `
	CREATE TABLE IF NOT EXISTS suites (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT UNIQUE NOT NULL,
		is_current BOOLEAN DEFAULT FALSE
	);

	CREATE TABLE IF NOT EXISTS profiles (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		description TEXT,
		suite_id INTEGER NOT NULL,
		FOREIGN KEY (suite_id) REFERENCES suites(id) ON DELETE CASCADE,
		UNIQUE(name, suite_id)
	);

	CREATE TABLE IF NOT EXISTS prompts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		text TEXT NOT NULL,
		solution TEXT,
		profile_id INTEGER,
		suite_id INTEGER NOT NULL,
		display_order INTEGER NOT NULL,
		FOREIGN KEY (profile_id) REFERENCES profiles(id) ON DELETE SET NULL,
		FOREIGN KEY (suite_id) REFERENCES suites(id) ON DELETE CASCADE,
		UNIQUE(text, suite_id)
	);

	CREATE TABLE IF NOT EXISTS models (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		suite_id INTEGER NOT NULL,
		FOREIGN KEY (suite_id) REFERENCES suites(id) ON DELETE CASCADE,
		UNIQUE(name, suite_id)
	);

	CREATE TABLE IF NOT EXISTS scores (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		model_id INTEGER NOT NULL,
		prompt_id INTEGER NOT NULL,
		score INTEGER NOT NULL DEFAULT 0,
		FOREIGN KEY (model_id) REFERENCES models(id) ON DELETE CASCADE,
		FOREIGN KEY (prompt_id) REFERENCES prompts(id) ON DELETE CASCADE,
		UNIQUE(model_id, prompt_id)
	);

	-- Add the default suite if it doesn't exist
	INSERT OR IGNORE INTO suites (name, is_current) VALUES ('default', 1);
	`

	_, err := db.Exec(schema)
	return err
}

// GetSuiteID returns the ID of the specified suite or creates it if it doesn't exist
func GetSuiteID(suiteName string) (int, error) {
	if suiteName == "" {
		suiteName = "default"
	}

	// Try to get the existing suite
	var suiteID int
	err := db.QueryRow("SELECT id FROM suites WHERE name = ?", suiteName).Scan(&suiteID)
	if err == nil {
		return suiteID, nil
	}

	// If suite doesn't exist, create it
	if err == sql.ErrNoRows {
		result, err := db.Exec("INSERT INTO suites (name) VALUES (?)", suiteName)
		if err != nil {
			return 0, fmt.Errorf("failed to create suite: %w", err)
		}
		id, err := result.LastInsertId()
		if err != nil {
			return 0, fmt.Errorf("failed to get suite ID: %w", err)
		}
		return int(id), nil
	}

	return 0, err
}

// GetCurrentSuiteID returns the ID of the current suite
func GetCurrentSuiteID() (int, error) {
	var suiteID int
	err := db.QueryRow("SELECT id FROM suites WHERE is_current = 1").Scan(&suiteID)
	if err == sql.ErrNoRows {
		// If no current suite, set default as current
		_, err = db.Exec("UPDATE suites SET is_current = 1 WHERE name = 'default'")
		if err != nil {
			return 0, fmt.Errorf("failed to set default suite as current: %w", err)
		}
		return GetCurrentSuiteID()
	}
	return suiteID, err
}

// SetCurrentSuite sets the specified suite as the current one
func SetCurrentSuite(suiteName string) error {
	if suiteName == "" {
		suiteName = "default"
	}

	// Get or create the suite
	suiteID, err := GetSuiteID(suiteName)
	if err != nil {
		return fmt.Errorf("failed to get suite: %w", err)
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

	// Clear current suite flag
	_, err = tx.Exec("UPDATE suites SET is_current = 0")
	if err != nil {
		return fmt.Errorf("failed to clear current suite flag: %w", err)
	}

	// Set new current suite
	_, err = tx.Exec("UPDATE suites SET is_current = 1 WHERE id = ?", suiteID)
	if err != nil {
		return fmt.Errorf("failed to set current suite: %w", err)
	}

	return tx.Commit()
}

// ListSuites returns a list of all suite names
func ListSuites() ([]string, error) {
	rows, err := db.Query("SELECT name FROM suites ORDER BY name")
	if err != nil {
		return nil, fmt.Errorf("failed to query suites: %w", err)
	}
	defer rows.Close()

	var suites []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("failed to scan suite name: %w", err)
		}
		suites = append(suites, name)
	}

	return suites, nil
}

// DeleteSuite deletes a suite and all associated data
func DeleteSuite(suiteName string) error {
	if suiteName == "default" {
		return fmt.Errorf("cannot delete the default suite")
	}

	suiteID, err := GetSuiteID(suiteName)
	if err != nil {
		return fmt.Errorf("failed to get suite ID: %w", err)
	}

	// Check if it's the current suite
	var isCurrent bool
	err = db.QueryRow("SELECT is_current FROM suites WHERE id = ?", suiteID).Scan(&isCurrent)
	if err != nil {
		return fmt.Errorf("failed to check if suite is current: %w", err)
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

	// Delete the suite (cascade will delete related data)
	_, err = tx.Exec("DELETE FROM suites WHERE id = ?", suiteID)
	if err != nil {
		return fmt.Errorf("failed to delete suite: %w", err)
	}

	// If it was the current suite, set default as current
	if isCurrent {
		_, err = tx.Exec("UPDATE suites SET is_current = 1 WHERE name = 'default'")
		if err != nil {
			return fmt.Errorf("failed to set default suite as current: %w", err)
		}
	}

	return tx.Commit()
}

// RenameSuite renames a suite
func RenameSuite(oldName, newName string) error {
	if oldName == "default" {
		return fmt.Errorf("cannot rename the default suite")
	}
	if newName == "" {
		return fmt.Errorf("new suite name cannot be empty")
	}
	if strings.ContainsAny(newName, "/\\") {
		return fmt.Errorf("suite name contains invalid characters")
	}

	// Check if new name already exists
	var exists int
	err := db.QueryRow("SELECT 1 FROM suites WHERE name = ?", newName).Scan(&exists)
	if err == nil {
		return fmt.Errorf("suite with name '%s' already exists", newName)
	} else if err != sql.ErrNoRows {
		return fmt.Errorf("failed to check if suite exists: %w", err)
	}

	// Rename the suite
	_, err = db.Exec("UPDATE suites SET name = ? WHERE name = ?", newName, oldName)
	if err != nil {
		return fmt.Errorf("failed to rename suite: %w", err)
	}

	return nil
}

// MigrateFromJSON migrates data from JSON files to SQLite database
func MigrateFromJSON() error {
	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Get list of all suite names from files
	suites, err := ListPromptSuites()
	if err != nil {
		log.Printf("Warning: could not list prompt suites from files: %v", err)
		// Continue with default suite
		suites = []string{"default"}
	}

	log.Printf("Migrating %d suites: %v", len(suites), suites)

	// Process each suite
	for _, suiteName := range suites {
		log.Printf("Migrating suite: %s", suiteName)
		
		// Get or create suite in DB
		stmt, err := tx.Prepare("INSERT OR IGNORE INTO suites (name, is_current) VALUES (?, ?)")
		if err != nil {
			return fmt.Errorf("failed to prepare suite insert: %w", err)
		}
		
		isCurrent := 0
		currentSuiteName := GetCurrentSuiteName()
		if suiteName == currentSuiteName {
			isCurrent = 1
		}
		
		_, err = stmt.Exec(suiteName, isCurrent)
		stmt.Close()
		if err != nil {
			return fmt.Errorf("failed to insert suite: %w", err)
		}

		var suiteID int
		err = tx.QueryRow("SELECT id FROM suites WHERE name = ?", suiteName).Scan(&suiteID)
		if err != nil {
			return fmt.Errorf("failed to get suite ID: %w", err)
		}
		log.Printf("Suite ID for %s: %d", suiteName, suiteID)

		// Migrate profiles
		profilesMap := make(map[string]int) // Maps profile name to ID
		profiles, err := ReadProfileSuiteFromJSON(suiteName)
		if err == nil && len(profiles) > 0 {
			log.Printf("Migrating %d profiles", len(profiles))
			stmt, err := tx.Prepare("INSERT OR REPLACE INTO profiles (name, description, suite_id) VALUES (?, ?, ?)")
			if err != nil {
				return fmt.Errorf("failed to prepare profile insert: %w", err)
			}
			
			for _, profile := range profiles {
				result, err := stmt.Exec(profile.Name, profile.Description, suiteID)
				if err != nil {
					stmt.Close()
					return fmt.Errorf("failed to insert profile: %w", err)
				}
				
				profileID, err := result.LastInsertId()
				if err != nil {
					stmt.Close()
					return fmt.Errorf("failed to get profile ID: %w", err)
				}
				
				profilesMap[profile.Name] = int(profileID)
				log.Printf("Migrated profile %s with ID %d", profile.Name, profileID)
			}
			stmt.Close()
		} else if err != nil {
			log.Printf("No profiles found or error: %v", err)
		}

		// Migrate prompts
		promptsMap := make(map[int]int) // Maps old index to new ID
		prompts, err := ReadPromptSuiteFromJSON(suiteName)
		if err == nil && len(prompts) > 0 {
			log.Printf("Migrating %d prompts", len(prompts))
			
			// First, delete any existing prompts for this suite to avoid duplicates
			_, err := tx.Exec("DELETE FROM prompts WHERE suite_id = ?", suiteID)
			if err != nil {
				return fmt.Errorf("failed to clear existing prompts: %w", err)
			}
			log.Printf("Cleared existing prompts for suite %s", suiteName)
			
			stmt, err := tx.Prepare("INSERT INTO prompts (text, solution, profile_id, suite_id, display_order) VALUES (?, ?, ?, ?, ?)")
			if err != nil {
				return fmt.Errorf("failed to prepare prompt insert: %w", err)
			}
			
			for i, prompt := range prompts {
				var profileID sql.NullInt64
				if prompt.Profile != "" {
					if id, exists := profilesMap[prompt.Profile]; exists {
						profileID.Int64 = int64(id)
						profileID.Valid = true
					}
				}
				
				result, err := stmt.Exec(prompt.Text, prompt.Solution, profileID, suiteID, i)
				if err != nil {
					stmt.Close()
					return fmt.Errorf("failed to insert prompt: %w", err)
				}
				
				promptID, err := result.LastInsertId()
				if err != nil {
					stmt.Close()
					return fmt.Errorf("failed to get prompt ID: %w", err)
				}
				
				promptsMap[i] = int(promptID)
				log.Printf("Migrated prompt %d with ID %d", i, promptID)
			}
			stmt.Close()
		} else if err != nil {
			log.Printf("No prompts found or error: %v", err)
		}

		// Migrate results
		results, err := ReadResultsFromJSON(suiteName)
		if err == nil && len(results) > 0 {
			log.Printf("Migrating results for %d models", len(results))
			
			// First, insert all models
			modelMap := make(map[string]int) // Maps model name to ID
			modelStmt, err := tx.Prepare("INSERT OR REPLACE INTO models (name, suite_id) VALUES (?, ?)")
			if err != nil {
				return fmt.Errorf("failed to prepare model insert: %w", err)
			}
			
			for modelName := range results {
				result, err := modelStmt.Exec(modelName, suiteID)
				if err != nil {
					modelStmt.Close()
					return fmt.Errorf("failed to insert model: %w", err)
				}
				
				modelID, err := result.LastInsertId()
				if err != nil {
					modelStmt.Close()
					return fmt.Errorf("failed to get model ID: %w", err)
				}
				
				modelMap[modelName] = int(modelID)
				log.Printf("Migrated model %s with ID %d", modelName, modelID)
			}
			modelStmt.Close()

			// Debug the result of the models insert
			var count int
			err = tx.QueryRow("SELECT COUNT(*) FROM models WHERE suite_id = ?", suiteID).Scan(&count)
			if err != nil {
				log.Printf("Error checking models count: %v", err)
			} else {
				log.Printf("Verified %d models were inserted for suite %s", count, suiteName)
			}

			// Then insert scores
			scoreStmt, err := tx.Prepare("INSERT INTO scores (model_id, prompt_id, score) VALUES (?, ?, ?)")
			if err != nil {
				return fmt.Errorf("failed to prepare score insert: %w", err)
			}
			
			scoreCount := 0
			for modelName, result := range results {
				modelID, exists := modelMap[modelName]
				if !exists {
					log.Printf("Warning: Model ID not found for %s", modelName)
					continue
				}
				
				log.Printf("Migrating %d scores for model %s (ID: %d)", len(result.Scores), modelName, modelID)
				
				for i, score := range result.Scores {
					promptID, exists := promptsMap[i]
					if !exists {
						log.Printf("Warning: Prompt ID not found for index %d", i)
						continue
					}
					
					_, err := scoreStmt.Exec(modelID, promptID, score)
					if err != nil {
						scoreStmt.Close()
						return fmt.Errorf("failed to insert score for model %s, prompt %d: %w", modelName, i, err)
					}
					scoreCount++
				}
			}
			scoreStmt.Close()
			log.Printf("Total of %d scores migrated for suite %s", scoreCount, suiteName)
			
			// Verify scores were inserted
			err = tx.QueryRow("SELECT COUNT(*) FROM scores s JOIN models m ON s.model_id = m.id WHERE m.suite_id = ?", suiteID).Scan(&count)
			if err != nil {
				log.Printf("Error checking scores count: %v", err)
			} else {
				log.Printf("Verified %d scores were inserted for suite %s", count, suiteName)
			}
		} else if err != nil {
			log.Printf("No results found or error: %v", err)
		} else {
			log.Printf("No results to migrate for suite %s", suiteName)
		}
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	
	log.Println("Migration completed successfully!")
	return nil
}

// CleanupDuplicatePrompts removes duplicate prompts from the database
func CleanupDuplicatePrompts() error {
	log.Println("Starting cleanup of duplicate prompts...")
	
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
	
	// Get all suites
	rows, err := tx.Query("SELECT id, name FROM suites")
	if err != nil {
		return fmt.Errorf("failed to query suites: %w", err)
	}
	
	var suites []struct {
		ID   int
		Name string
	}
	
	for rows.Next() {
		var suite struct {
			ID   int
			Name string
		}
		if err := rows.Scan(&suite.ID, &suite.Name); err != nil {
			rows.Close()
			return fmt.Errorf("failed to scan suite: %w", err)
		}
		suites = append(suites, suite)
	}
	rows.Close()
	
	for _, suite := range suites {
		log.Printf("Cleaning up duplicates for suite: %s (ID: %d)", suite.Name, suite.ID)
		
		// First find duplicate prompts by text within the same suite
		duplicateRows, err := tx.Query(`
		SELECT MIN(id) as keep_id, text, COUNT(*) as count
		FROM prompts 
		WHERE suite_id = ?
		GROUP BY text
		HAVING COUNT(*) > 1
		`, suite.ID)
		if err != nil {
			return fmt.Errorf("failed to find duplicates: %w", err)
		}
		
		type DuplicateGroup struct {
			KeepID int
			Text   string
			Count  int
		}
		
		var duplicates []DuplicateGroup
		for duplicateRows.Next() {
			var dg DuplicateGroup
			if err := duplicateRows.Scan(&dg.KeepID, &dg.Text, &dg.Count); err != nil {
				duplicateRows.Close()
				return fmt.Errorf("failed to scan duplicate row: %w", err)
			}
			duplicates = append(duplicates, dg)
		}
		duplicateRows.Close()
		
		if len(duplicates) == 0 {
			log.Printf("No duplicates found for suite %s", suite.Name)
			continue
		}
		
		log.Printf("Found %d duplicate prompt groups in suite %s", len(duplicates), suite.Name)
		
		// For each duplicate group, keep the one with the lowest ID and delete the rest
		for _, dg := range duplicates {
			// Get the IDs of all duplicates for this text
			dupIDs, err := tx.Query(`
			SELECT id FROM prompts 
			WHERE text = ? AND suite_id = ? AND id != ?
			`, dg.Text, suite.ID, dg.KeepID)
			if err != nil {
				return fmt.Errorf("failed to query duplicate IDs: %w", err)
			}
			
			var idsToDelete []int
			for dupIDs.Next() {
				var id int
				if err := dupIDs.Scan(&id); err != nil {
					dupIDs.Close()
					return fmt.Errorf("failed to scan duplicate ID: %w", err)
				}
				idsToDelete = append(idsToDelete, id)
			}
			dupIDs.Close()
			
			if len(idsToDelete) > 0 {
				log.Printf("Deleting %d duplicates for prompt text: %s...", len(idsToDelete), dg.Text[:min(20, len(dg.Text))])
				
				// Update any scores to point to the kept prompt
				for _, idToDelete := range idsToDelete {
					_, err = tx.Exec(`
					UPDATE scores SET prompt_id = ? 
					WHERE prompt_id = ?
					`, dg.KeepID, idToDelete)
					if err != nil {
						return fmt.Errorf("failed to update scores for duplicate prompt: %w", err)
					}
				}
				
				// Now delete the duplicate prompts
				for _, idToDelete := range idsToDelete {
					_, err = tx.Exec("DELETE FROM prompts WHERE id = ?", idToDelete)
					if err != nil {
						return fmt.Errorf("failed to delete duplicate prompt: %w", err)
					}
				}
			}
		}
		
		// Finally, reorder the display_order values to be consecutive
		log.Printf("Reordering prompts for suite %s", suite.Name)
		promptRows, err := tx.Query(`
		SELECT id FROM prompts 
		WHERE suite_id = ? 
		ORDER BY display_order
		`, suite.ID)
		if err != nil {
			return fmt.Errorf("failed to query prompts for reordering: %w", err)
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
		
		for i, id := range promptIDs {
			_, err = tx.Exec("UPDATE prompts SET display_order = ? WHERE id = ?", i, id)
			if err != nil {
				return fmt.Errorf("failed to update prompt order: %w", err)
			}
		}
	}
	
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	
	log.Println("Duplicate prompt cleanup completed successfully!")
	return nil
}

// RemigrateScores specifically remigrates just the scores
func RemigrateScores() error {
	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Get list of all suites from database
	rows, err := tx.Query("SELECT id, name FROM suites")
	if err != nil {
		return fmt.Errorf("failed to query suites: %w", err)
	}
	
	var suites []struct {
		ID   int
		Name string
	}
	
	for rows.Next() {
		var suite struct {
			ID   int
			Name string
		}
		if err := rows.Scan(&suite.ID, &suite.Name); err != nil {
			rows.Close()
			return fmt.Errorf("failed to scan suite: %w", err)
		}
		suites = append(suites, suite)
	}
	rows.Close()
	
	log.Printf("Re-migrating scores for %d suites", len(suites))

	// Clear existing scores
	_, err = tx.Exec("DELETE FROM scores")
	if err != nil {
		return fmt.Errorf("failed to clear scores: %w", err)
	}
	log.Println("Cleared existing scores")

	for _, suite := range suites {
		log.Printf("Processing suite: %s (ID: %d)", suite.Name, suite.ID)
		
		// Get all prompts for this suite with their index
		promptRows, err := tx.Query("SELECT id, display_order FROM prompts WHERE suite_id = ? ORDER BY display_order", suite.ID)
		if err != nil {
			return fmt.Errorf("failed to query prompts: %w", err)
		}
		
		promptMap := make(map[int]int) // Maps display_order to ID
		for promptRows.Next() {
			var id, displayOrder int
			if err := promptRows.Scan(&id, &displayOrder); err != nil {
				promptRows.Close()
				return fmt.Errorf("failed to scan prompt: %w", err)
			}
			promptMap[displayOrder] = id
		}
		promptRows.Close()
		
		log.Printf("Found %d prompts for suite %s", len(promptMap), suite.Name)
		
		// Get all models for this suite
		modelRows, err := tx.Query("SELECT id, name FROM models WHERE suite_id = ?", suite.ID)
		if err != nil {
			return fmt.Errorf("failed to query models: %w", err)
		}
		
		modelMap := make(map[string]int) // Maps name to ID
		for modelRows.Next() {
			var id int
			var name string
			if err := modelRows.Scan(&id, &name); err != nil {
				modelRows.Close()
				return fmt.Errorf("failed to scan model: %w", err)
			}
			modelMap[name] = id
		}
		modelRows.Close()
		
		log.Printf("Found %d models for suite %s", len(modelMap), suite.Name)
		
		// Get results from JSON
		results, err := ReadResultsFromJSON(suite.Name)
		if err != nil {
			log.Printf("Warning: could not read results for suite %s: %v", suite.Name, err)
			continue
		}
		
		log.Printf("Found results for %d models in JSON for suite %s", len(results), suite.Name)
		
		// Insert scores
		scoreStmt, err := tx.Prepare("INSERT INTO scores (model_id, prompt_id, score) VALUES (?, ?, ?)")
		if err != nil {
			return fmt.Errorf("failed to prepare score insert: %w", err)
		}
		
		scoreCount := 0
		for modelName, result := range results {
			modelID, modelExists := modelMap[modelName]
			if !modelExists {
				log.Printf("Warning: Model %s not found in database, creating it", modelName)
				// Create the model if it doesn't exist
				res, err := tx.Exec("INSERT OR REPLACE INTO models (name, suite_id) VALUES (?, ?)", modelName, suite.ID)
				if err != nil {
					scoreStmt.Close()
					return fmt.Errorf("failed to insert model %s: %w", modelName, err)
				}
				
				id, err := res.LastInsertId()
				if err != nil {
					scoreStmt.Close()
					return fmt.Errorf("failed to get model ID: %w", err)
				}
				
				modelID = int(id)
				log.Printf("Created model %s with ID %d", modelName, modelID)
			}
			
			for i, score := range result.Scores {
				promptID, promptExists := promptMap[i]
				if !promptExists {
					log.Printf("Warning: Prompt with index %d not found for suite %s", i, suite.Name)
					continue
				}
				
				_, err := scoreStmt.Exec(modelID, promptID, score)
				if err != nil {
					scoreStmt.Close()
					return fmt.Errorf("failed to insert score for model %s, prompt %d: %w", modelName, i, err)
				}
				scoreCount++
			}
		}
		scoreStmt.Close()
		
		log.Printf("Inserted %d scores for suite %s", scoreCount, suite.Name)
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	
	log.Println("Score remigration completed successfully!")
	return nil
}

// GetProfileID returns the profile ID for a given name and suite
func GetProfileID(profileName string, suiteID int) (int, bool, error) {
	if profileName == "" {
		return 0, false, nil
	}

	var profileID int
	err := db.QueryRow("SELECT id FROM profiles WHERE name = ? AND suite_id = ?", profileName, suiteID).Scan(&profileID)
	if err == sql.ErrNoRows {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	return profileID, true, nil
}
