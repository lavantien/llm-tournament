package main

import (
	"database/sql"
	"flag"
	"llm-tournament/middleware"
	"net/http"
)

// Config holds application configuration
type Config struct {
	DBPath         string
	Port           string
	MigrateResults bool
}

// DefaultConfig returns default configuration values
func DefaultConfig() *Config {
	return &Config{
		DBPath: "data/tournament.db",
		Port:   ":8080",
	}
}

// ParseFlags parses command line flags into Config
// Note: This modifies flag.CommandLine state
func ParseFlags(args []string) (*Config, error) {
	cfg := DefaultConfig()

	fs := flag.NewFlagSet("llm-tournament", flag.ContinueOnError)
	fs.BoolVar(&cfg.MigrateResults, "migrate-results", false, "Migrate existing results to new scoring system")
	fs.StringVar(&cfg.DBPath, "db", cfg.DBPath, "SQLite database path")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	return cfg, nil
}

// InitDB initializes the database connection
func InitDB(dbPath string) error {
	return middleware.InitDB(dbPath)
}

// CloseDB closes the database connection
func CloseDB() {
	_ = middleware.CloseDB()
}

// GetDB returns the database instance
func GetDB() *sql.DB {
	return middleware.GetDB()
}

// RunMigration performs the results migration
func RunMigration() error {
	results := middleware.ReadResults()
	middleware.MigrateResults(results)
	suiteName := middleware.GetCurrentSuiteName()
	return middleware.WriteResults(suiteName, results)
}

// Routes returns the application routes map
func Routes() map[string]http.HandlerFunc {
	return routes
}

// SetupRoutes registers all routes on the given ServeMux
func SetupRoutes(mux *http.ServeMux) {
	for path, handler := range routes {
		mux.HandleFunc(path, handler)
	}
	mux.HandleFunc("/", router)
	mux.HandleFunc("/ws", middleware.HandleWebSocket)
	mux.Handle("/templates/", http.StripPrefix("/templates/", middleware.StaticFiles(http.Dir("templates"))))
	mux.Handle("/assets/", http.StripPrefix("/assets/", middleware.StaticFiles(http.Dir("assets"))))
}

// ServerHandler returns the fully wired HTTP handler: all routes behind the
// cross-origin guard that blocks browser-driven cross-site posts against the
// local server.
func ServerHandler() http.Handler {
	return middleware.CheckOrigin(NewServeMux())
}

// NewServeMux creates and configures a new ServeMux with all routes
func NewServeMux() *http.ServeMux {
	mux := http.NewServeMux()
	SetupRoutes(mux)
	return mux
}

// Router handles routing for unknown paths
func Router(w http.ResponseWriter, r *http.Request) {
	router(w, r)
}
