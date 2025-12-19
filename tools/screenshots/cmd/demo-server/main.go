package main

import (
	"context"
	"flag"
	"fmt"
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

func main() {
	log.SetOutput(os.Stderr)

	var (
		addr = flag.String("addr", "127.0.0.1:0", "listen address (use :0 for random port)")
		db   = flag.String("db", "", "path to sqlite db file (required)")
		seed = flag.Bool("seed", true, "seed demo data")
	)
	flag.Parse()

	if *db == "" {
		log.Fatal("missing required -db")
	}

	ensureDemoEncryptionKey()

	if err := middleware.InitDB(*db); err != nil {
		log.Fatalf("init db: %v", err)
	}
	defer func() { _ = middleware.CloseDB() }()

	if *seed {
		if err := seedDemoData(); err != nil {
			log.Fatalf("seed demo data: %v", err)
		}
	}

	handlers.InitEvaluator(middleware.GetDB())

	mux := http.NewServeMux()
	registerRoutes(mux)

	server := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	ln, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	// stdout is reserved for machine-readable output (used by Playwright harness).
	fmt.Printf("URL=http://%s\n", ln.Addr().String())

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
		errCh <- server.Serve(ln)
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	select {
	case <-sigCh:
	case <-shutdownCh:
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = server.Shutdown(ctx)
}
