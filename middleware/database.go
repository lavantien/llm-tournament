package middleware

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

var lastInsertID = func(result sql.Result) (int64, error) { return result.LastInsertId() }
var sqlOpen = sql.Open
var execPragmas = func(conn *sql.DB) error {
	_, err := conn.Exec(`PRAGMA journal_mode = WAL;
                     PRAGMA synchronous = NORMAL;
                     PRAGMA foreign_keys = ON;`)
	return err
}
var createTablesFunc = createTables
var dbBegin = func() (*sql.Tx, error) { return db.Begin() }

// GetDB returns the database connection
func GetDB() *sql.DB {
	return db
}

// InitDB initializes the SQLite database connection
func InitDB(dbPath string) error {
	// Ensure data directory exists
	dataDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	var err error
	db, err = sqlOpen("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Set pragmas for better performance
	if err = execPragmas(db); err != nil {
		return fmt.Errorf("failed to set database pragmas: %w", err)
	}

	// Create schema if it doesn't exist
	if err = createTablesFunc(); err != nil {
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
		type TEXT NOT NULL DEFAULT 'objective',
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

	CREATE TABLE IF NOT EXISTS settings (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		key TEXT UNIQUE NOT NULL,
		value TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS evaluation_jobs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		suite_id INTEGER NOT NULL,
		job_type TEXT NOT NULL,
		target_id INTEGER,
		status TEXT NOT NULL DEFAULT 'pending',
		progress_current INTEGER DEFAULT 0,
		progress_total INTEGER DEFAULT 0,
		estimated_cost_usd REAL DEFAULT 0.0,
		actual_cost_usd REAL DEFAULT 0.0,
		error_message TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		started_at TIMESTAMP,
		completed_at TIMESTAMP,
		FOREIGN KEY (suite_id) REFERENCES suites(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS model_responses (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		model_id INTEGER NOT NULL,
		prompt_id INTEGER NOT NULL,
		response_text TEXT,
		response_source TEXT NOT NULL DEFAULT 'manual',
		api_config TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (model_id) REFERENCES models(id) ON DELETE CASCADE,
		FOREIGN KEY (prompt_id) REFERENCES prompts(id) ON DELETE CASCADE,
		UNIQUE(model_id, prompt_id)
	);

	CREATE TABLE IF NOT EXISTS evaluation_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		job_id INTEGER NOT NULL,
		model_id INTEGER NOT NULL,
		prompt_id INTEGER NOT NULL,
		judge_name TEXT NOT NULL,
		judge_score INTEGER,
		judge_confidence REAL,
		judge_reasoning TEXT,
		cost_usd REAL DEFAULT 0.0,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (job_id) REFERENCES evaluation_jobs(id) ON DELETE CASCADE,
		FOREIGN KEY (model_id) REFERENCES models(id) ON DELETE CASCADE,
		FOREIGN KEY (prompt_id) REFERENCES prompts(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS cost_tracking (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		suite_id INTEGER NOT NULL,
		date DATE NOT NULL,
		total_cost_usd REAL DEFAULT 0.0,
		evaluation_count INTEGER DEFAULT 0,
		FOREIGN KEY (suite_id) REFERENCES suites(id) ON DELETE CASCADE,
		UNIQUE(suite_id, date)
	);

	-- Create indexes
	CREATE INDEX IF NOT EXISTS idx_settings_key ON settings(key);
	CREATE INDEX IF NOT EXISTS idx_evaluation_jobs_status ON evaluation_jobs(status);
	CREATE INDEX IF NOT EXISTS idx_evaluation_jobs_suite ON evaluation_jobs(suite_id);
	CREATE INDEX IF NOT EXISTS idx_model_responses_lookup ON model_responses(model_id, prompt_id);
	CREATE INDEX IF NOT EXISTS idx_evaluation_history_job ON evaluation_history(job_id);
	CREATE INDEX IF NOT EXISTS idx_evaluation_history_lookup ON evaluation_history(model_id, prompt_id);
	CREATE INDEX IF NOT EXISTS idx_cost_tracking_suite_date ON cost_tracking(suite_id, date);

	-- Add the default suite if it doesn't exist
	INSERT OR IGNORE INTO suites (name, is_current) VALUES ('default', 1);

	-- Initialize default settings
	INSERT OR IGNORE INTO settings (key, value) VALUES
		('api_key_anthropic', ''),
		('api_key_openai', ''),
		('api_key_google', ''),
		('cost_alert_threshold_usd', '100.0'),
		('auto_evaluate_new_models', 'false'),
		('python_service_url', 'http://localhost:8001');
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
			id, err := lastInsertID(result)
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
	tx, err := dbBegin()
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
	tx, err := dbBegin()
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
