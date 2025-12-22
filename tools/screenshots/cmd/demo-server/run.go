package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"io"
	"llm-tournament/handlers"
	"llm-tournament/middleware"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type runDeps struct {
	stdout io.Writer

	initDB         func(string) error
	closeDB        func() error
	seedDemoData   func() error
	initEvaluator  func(*sql.DB)
	getDB          func() *sql.DB
	listen         func(network, address string) (net.Listener, error)
	serve          func(*http.Server, net.Listener) error
	signalCh       <-chan os.Signal
	setLogOutput   func(io.Writer)
	ensureDemoKey  func()
	registerRoutes func(*http.ServeMux)
}

func defaultRunDeps() runDeps {
	return runDeps{
		stdout:         os.Stdout,
		initDB:         middleware.InitDB,
		closeDB:        middleware.CloseDB,
		seedDemoData:   seedDemoData,
		initEvaluator:  handlers.InitEvaluator,
		getDB:          middleware.GetDB,
		listen:         net.Listen,
		serve:          func(s *http.Server, ln net.Listener) error { return s.Serve(ln) },
		setLogOutput:   log.SetOutput,
		ensureDemoKey:  ensureDemoEncryptionKey,
		registerRoutes: registerRoutes,
	}
}

func run(args []string, deps runDeps) int {
	if deps.stdout == nil {
		deps.stdout = io.Discard
	}
	if deps.setLogOutput == nil {
		deps.setLogOutput = log.SetOutput
	}
	if deps.ensureDemoKey == nil {
		deps.ensureDemoKey = ensureDemoEncryptionKey
	}
	if deps.registerRoutes == nil {
		deps.registerRoutes = registerRoutes
	}

	fs := flag.NewFlagSet("demo-server", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	addr := fs.String("addr", "127.0.0.1:0", "listen address (use :0 for random port)")
	db := fs.String("db", "", "path to sqlite db file (required)")
	seed := fs.Bool("seed", true, "seed demo data")

	if err := fs.Parse(args); err != nil {
		return 2
	}

	if *db == "" {
		log.Print("missing required -db")
		return 1
	}

	deps.ensureDemoKey()

	if err := deps.initDB(*db); err != nil {
		log.Printf("init db: %v", err)
		return 1
	}
	defer func() { _ = deps.closeDB() }()

	if *seed {
		if err := deps.seedDemoData(); err != nil {
			log.Printf("seed demo data: %v", err)
			return 1
		}
	}

	deps.initEvaluator(deps.getDB())

	mux := http.NewServeMux()
	deps.registerRoutes(mux)

	server := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	ln, err := deps.listen("tcp", *addr)
	if err != nil {
		log.Printf("listen: %v", err)
		return 1
	}

	// stdout is reserved for machine-readable output (used by Playwright harness).
	_, _ = fmt.Fprintf(deps.stdout, "URL=http://%s\n", ln.Addr().String())

	shutdownCh := make(chan struct{}, 1)
	mux.HandleFunc("/__shutdown", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusOK)
		select {
		case shutdownCh <- struct{}{}:
		default:
		}
	})

	errCh := make(chan error, 1)
	go func() {
		errCh <- deps.serve(server, ln)
	}()

	sigCh := deps.signalCh
	if sigCh == nil {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		defer signal.Stop(c)
		sigCh = c
	}

	select {
	case <-sigCh:
	case <-shutdownCh:
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			log.Printf("server error: %v", err)
			return 1
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = server.Shutdown(ctx)

	return 0
}
