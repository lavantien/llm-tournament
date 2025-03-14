package main

import (
	"flag"
	"log"
	"math/rand"
	"net/http"
	"time"

	"llm-tournament/handlers"
	"llm-tournament/middleware"
)

func main() {
	// Seed the random number generator for mock data generation
	rand.Seed(time.Now().UnixNano())
	
	// Parse command line flags
	migrateResults := flag.Bool("migrate-results", false, "Migrate existing results to new scoring system")
	migrateToSQLite := flag.Bool("migrate-to-sqlite", false, "Migrate data from JSON files to SQLite database")
	remigrateScores := flag.Bool("remigrate-scores", false, "Remigrate just the scores from JSON to SQLite")
	dbPath := flag.String("db", "data/tournament.db", "SQLite database path")
	flag.Parse()

	// Initialize the database
	log.Println("Initializing database...")
	if err := middleware.InitDB(*dbPath); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer middleware.CloseDB()

	// Handle migration from JSON to SQLite
	if *migrateToSQLite {
		log.Println("Migrating data from JSON files to SQLite database...")
		if err := middleware.MigrateFromJSON(); err != nil {
			log.Fatalf("Error migrating data to SQLite: %v", err)
		}
		log.Println("Migration to SQLite completed successfully")
		return
	}
	
	// Handle remigration of scores
	if *remigrateScores {
		log.Println("Remigrating scores from JSON files to SQLite database...")
		if err := middleware.RemigrateScores(); err != nil {
			log.Fatalf("Error remigrating scores to SQLite: %v", err)
		}
		log.Println("Score remigration completed successfully")
		return
	}

	// Handle old result format migration
	if *migrateResults {
		log.Println("Migrating results to new scoring system...")
		results := middleware.ReadResults()
		middleware.MigrateResults(results)
		suiteName := middleware.GetCurrentSuiteName()
		err := middleware.WriteResults(suiteName, results)
		if err != nil {
			log.Fatalf("Error migrating results: %v", err)
		}
		log.Println("Migration completed successfully")
		return
	}

	log.Println("Starting the server...")
	http.HandleFunc("/", router)
	http.HandleFunc("/ws", middleware.HandleWebSocket)
	http.Handle("/templates/", http.StripPrefix("/templates/", http.FileServer(http.Dir("templates"))))
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))
	log.Println("Server is listening on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}

var routes = map[string]http.HandlerFunc{
	"/import_error":            middleware.ImportErrorHandler,
	"/prompts":                 handlers.PromptListHandler,
	"/add_model":               handlers.AddModelHandler,
	"/edit_model":              handlers.EditModelHandler,
	"/delete_model":            handlers.DeleteModelHandler,
	"/add_prompt":              handlers.AddPromptHandler,
	"/edit_prompt":             handlers.EditPromptHandler,
	"/delete_prompt":           handlers.DeletePromptHandler,
	"/move_prompt":             handlers.MovePromptHandler,
	"/import_results":          handlers.ImportResultsHandler,
	"/export_prompts":          handlers.ExportPromptsHandler,
	"/import_prompts":          handlers.ImportPromptsHandler,
	"/update_prompts_order":    handlers.UpdatePromptsOrderHandler,
	"/reset_prompts":           handlers.ResetPromptsHandler,
	"/bulk_delete_prompts":     handlers.BulkDeletePromptsHandler,
	"/prompts/suites/new":      handlers.NewPromptSuiteHandler,
	"/prompts/suites/edit":     handlers.EditPromptSuiteHandler,
	"/prompts/suites/delete":   handlers.DeletePromptSuiteHandler,
	"/prompts/suites/select":   handlers.SelectPromptSuiteHandler,
	"/results":                 handlers.ResultsHandler,
	"/update_result":           handlers.UpdateResultHandler,
	"/reset_results":           handlers.ResetResultsHandler,
	"/confirm_refresh_results": handlers.ConfirmRefreshResultsHandler,
	"/refresh_results":         handlers.RefreshResultsHandler,
	"/export_results":          handlers.ExportResultsHandler,
	"/update_mock_results":     handlers.UpdateMockResultsHandler,
	"/evaluate":                handlers.EvaluateResult,
	"/profiles":                handlers.ProfilesHandler,
	"/add_profile":             handlers.AddProfileHandler,
	"/edit_profile":            handlers.EditProfileHandler,
	"/delete_profile":          handlers.DeleteProfileHandler,
	"/reset_profiles":          handlers.ResetProfilesHandler,
	"/stats":                   handlers.StatsHandler,
}

func router(w http.ResponseWriter, r *http.Request) {
	log.Printf("Request received: %s %s", r.Method, r.URL.Path)

	if handler, exists := routes[r.URL.Path]; exists {
		handler(w, r)
	} else {
		log.Printf("Redirecting to /prompts from %s", r.URL.Path)
		http.Redirect(w, r, "/prompts", http.StatusSeeOther)
	}
}
