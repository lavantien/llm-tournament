package main

import (
	"database/sql"
	"flag"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"llm-tournament/handlers"
	"llm-tournament/middleware"
)

type runDeps struct {
	initDB              func(string) error
	closeDB             func() error
	readResults         func() map[string]middleware.Result
	migrateResults      func(map[string]middleware.Result) map[string]middleware.Result
	getCurrentSuiteName func() string
	writeResults        func(string, map[string]middleware.Result) error
	initEvaluator       func(*sql.DB)
	getDB               func() *sql.DB
	listenAndServe      func(string, http.Handler) error
}

var osExit = os.Exit

func defaultRunDeps() runDeps {
	return runDeps{
		initDB:              middleware.InitDB,
		closeDB:             middleware.CloseDB,
		readResults:         middleware.ReadResults,
		migrateResults:      middleware.MigrateResults,
		getCurrentSuiteName: middleware.GetCurrentSuiteName,
		writeResults:        middleware.WriteResults,
		initEvaluator:       handlers.InitEvaluator,
		getDB:               middleware.GetDB,
		listenAndServe:      http.ListenAndServe,
	}
}

func run(args []string, deps runDeps) int {
	rand.Seed(time.Now().UnixNano())

	fs := flag.NewFlagSet("llm-tournament", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	migrateResults := fs.Bool("migrate-results", false, "Migrate existing results to new scoring system")
	dbPath := fs.String("db", "data/tournament.db", "SQLite database path")

	if err := fs.Parse(args); err != nil {
		return 2
	}

	log.Println("Initializing database...")
	if err := deps.initDB(*dbPath); err != nil {
		log.Printf("Failed to initialize database: %v", err)
		return 1
	}
	defer func() { _ = deps.closeDB() }()

	if *migrateResults {
		log.Println("Migrating results to new scoring system...")
		results := deps.readResults()
		results = deps.migrateResults(results)

		suiteName := deps.getCurrentSuiteName()
		if err := deps.writeResults(suiteName, results); err != nil {
			log.Printf("Error migrating results: %v", err)
			return 1
		}
		log.Println("Migration completed successfully")
		return 0
	}

	log.Println("Initializing evaluator...")
	deps.initEvaluator(deps.getDB())

	mux := http.NewServeMux()
	mux.HandleFunc("/", router)
	mux.HandleFunc("/ws", middleware.HandleWebSocket)
	mux.Handle("/templates/", http.StripPrefix("/templates/", http.FileServer(http.Dir("templates"))))
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))

	log.Println("Server is listening on :8080")
	if err := deps.listenAndServe(":8080", mux); err != nil {
		log.Printf("Error starting server: %v", err)
		return 1
	}

	return 0
}

func main() {
	osExit(run(os.Args[1:], defaultRunDeps()))
}
