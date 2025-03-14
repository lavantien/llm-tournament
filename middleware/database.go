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
		FOREIGN KEY (suite_id) REFERENCES suites(id) ON DELETE CASCADE
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

	// Process each suite
	for _, suiteName := range suites {
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

		// Migrate profiles
		profilesMap := make(map[string]int) // Maps profile name to ID
		profiles, err := ReadProfileSuiteFromJSON(suiteName)
		if err == nil && len(profiles) > 0 {
			stmt, err := tx.Prepare("INSERT INTO profiles (name, description, suite_id) VALUES (?, ?, ?)")
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
			}
			stmt.Close()
		}

		// Migrate prompts
		promptsMap := make(map[int]int) // Maps old index to new ID
		prompts, err := ReadPromptSuiteFromJSON(suiteName)
		if err == nil && len(prompts) > 0 {
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
			}
			stmt.Close()
		}

		// Migrate results
		results, err := ReadResultsFromJSON(suiteName)
		if err == nil && len(results) > 0 {
			// First, insert all models
			modelMap := make(map[string]int) // Maps model name to ID
			modelStmt, err := tx.Prepare("INSERT INTO models (name, suite_id) VALUES (?, ?)")
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
			}
			modelStmt.Close()

			// Then insert scores
			scoreStmt, err := tx.Prepare("INSERT INTO scores (model_id, prompt_id, score) VALUES (?, ?, ?)")
			if err != nil {
				return fmt.Errorf("failed to prepare score insert: %w", err)
			}
			
			for modelName, result := range results {
				modelID := modelMap[modelName]
				
				for i, score := range result.Scores {
					if promptID, exists := promptsMap[i]; exists {
						_, err := scoreStmt.Exec(modelID, promptID, score)
						if err != nil {
							scoreStmt.Close()
							return fmt.Errorf("failed to insert score: %w", err)
						}
					}
				}
			}
			scoreStmt.Close()
		}
	}

	// Commit the transaction
	return tx.Commit()
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
