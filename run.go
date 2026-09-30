package main

import (
	"database/sql"
	"flag"
	"fmt"
	"io"
	"llm-tournament/middleware"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type runDeps struct {
	initDB              func(string) error
	closeDB             func() error
	readResults         func() map[string]middleware.Result
	migrateResults      func(map[string]middleware.Result) map[string]middleware.Result
	getCurrentSuiteName func() string
	writeResults        func(string, map[string]middleware.Result) error
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
		getDB:               middleware.GetDB,
		listenAndServe:      http.ListenAndServe,
	}
}

func run(args []string, deps runDeps) int {
	fs := flag.NewFlagSet("llm-tournament", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	migrateResults := fs.Bool("migrate-results", false, "Migrate existing results to new scoring system")
	dbPath := fs.String("db", "data/tournament.db", "SQLite database path")
	port := fs.String("port", DefaultConfig().Port, "HTTP listen port")

	if err := fs.Parse(args); err != nil {
		return 2
	}

	addr, err := normalizePort(*port)
	if err != nil {
		log.Printf("Invalid port: %v", err)
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

	log.Printf("Server is listening on %s", addr)
	if err := deps.listenAndServe(addr, ServerHandler()); err != nil {
		log.Printf("Error starting server: %v", err)
		return 1
	}

	return 0
}

// normalizePort turns a port flag value ("8080" or ":8080") into a listen
// address, rejecting anything that is not a valid TCP port number.
func normalizePort(port string) (string, error) {
	trimmed := strings.TrimPrefix(port, ":")
	n, err := strconv.Atoi(trimmed)
	if err != nil || n < 0 || n > 65535 {
		return "", fmt.Errorf("invalid port %q", port)
	}
	return fmt.Sprintf(":%d", n), nil
}

func main() {
	osExit(run(os.Args[1:], defaultRunDeps()))
}
