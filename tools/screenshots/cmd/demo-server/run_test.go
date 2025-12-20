package main

import (
	"bytes"
	"database/sql"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

type stubListener struct {
	addr net.Addr
}

func (l stubListener) Accept() (net.Conn, error) { return nil, errors.New("accept not implemented") }
func (l stubListener) Close() error              { return nil }
func (l stubListener) Addr() net.Addr            { return l.addr }

type stubAddr struct {
	network string
	address string
}

func (a stubAddr) Network() string { return a.network }
func (a stubAddr) String() string  { return a.address }

func TestRun_MissingDBFlag_Returns1(t *testing.T) {
	deps := runDeps{
		stdout: io.Discard,
		initDB: func(string) error {
			t.Fatalf("initDB should not be called when -db is missing")
			return nil
		},
	}

	if exitCode := run([]string{"-seed=false"}, deps); exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}
}

func TestRun_FlagParseError_Returns2(t *testing.T) {
	deps := runDeps{
		stdout: io.Discard,
		initDB: func(string) error {
			t.Fatalf("initDB should not be called on flag parse error")
			return nil
		},
	}

	if exitCode := run([]string{"-no-such-flag"}, deps); exitCode != 2 {
		t.Fatalf("expected exit code 2, got %d", exitCode)
	}
}

func TestRun_Success_PrintsURL_RegistersShutdown_AndClosesDB(t *testing.T) {
	var (
		closeCalled         bool
		seedCalled          bool
		initEvaluatorCalled bool
	)

	out := &bytes.Buffer{}
	releaseServe := make(chan struct{})
	t.Cleanup(func() { close(releaseServe) })

	deps := runDeps{
		stdout: out,
		initDB: func(string) error { return nil },
		closeDB: func() error {
			closeCalled = true
			return nil
		},
		seedDemoData: func() error {
			seedCalled = true
			return nil
		},
		initEvaluator: func(*sql.DB) {
			initEvaluatorCalled = true
		},
		getDB: func() *sql.DB { return nil },
		listen: func(network, address string) (net.Listener, error) {
			return stubListener{addr: stubAddr{network: network, address: "127.0.0.1:12345"}}, nil
		},
		serve: func(server *http.Server, ln net.Listener) error {
			// Exercise the shutdown endpoint handler (GET -> 405, POST -> 200).
			req := httptest.NewRequest(http.MethodGet, "/__shutdown", nil)
			rr := httptest.NewRecorder()
			server.Handler.ServeHTTP(rr, req)
			if rr.Code != http.StatusMethodNotAllowed {
				t.Fatalf("GET /__shutdown expected %d, got %d", http.StatusMethodNotAllowed, rr.Code)
			}

			req = httptest.NewRequest(http.MethodPost, "/__shutdown", nil)
			rr = httptest.NewRecorder()
			server.Handler.ServeHTTP(rr, req)
			if rr.Code != http.StatusOK {
				t.Fatalf("POST /__shutdown expected %d, got %d", http.StatusOK, rr.Code)
			}

			// Second POST should be non-blocking (buffered channel); this hits the default path.
			req = httptest.NewRequest(http.MethodPost, "/__shutdown", nil)
			rr = httptest.NewRecorder()
			server.Handler.ServeHTTP(rr, req)

			<-releaseServe
			return http.ErrServerClosed
		},
		signalCh: make(chan os.Signal, 1),
	}

	exitCode := run([]string{"-db", "db.sqlite"}, deps)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}

	if !seedCalled {
		t.Fatalf("expected seedDemoData to be called when -seed is true by default")
	}
	if !initEvaluatorCalled {
		t.Fatalf("expected initEvaluator to be called")
	}
	if !closeCalled {
		t.Fatalf("expected closeDB to be called via defer")
	}
	if got := out.String(); !strings.Contains(got, "URL=http://127.0.0.1:12345") {
		t.Fatalf("expected URL output, got %q", got)
	}
}

func TestRun_ListenError_Returns1(t *testing.T) {
	var closeCalled bool
	deps := runDeps{
		stdout:       io.Discard,
		initDB:       func(string) error { return nil },
		closeDB:      func() error { closeCalled = true; return nil },
		seedDemoData: func() error { return nil },
		initEvaluator: func(*sql.DB) {
		},
		getDB: func() *sql.DB { return nil },
		listen: func(string, string) (net.Listener, error) {
			return nil, errors.New("listen failed")
		},
		serve:    func(*http.Server, net.Listener) error { return nil },
		signalCh: make(chan os.Signal, 1),
	}

	if exitCode := run([]string{"-db", "db.sqlite"}, deps); exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}
	if !closeCalled {
		t.Fatalf("expected closeDB to be called via defer")
	}
}

func TestRun_ServerError_Returns1(t *testing.T) {
	var closeCalled bool
	deps := runDeps{
		stdout:       io.Discard,
		initDB:       func(string) error { return nil },
		closeDB:      func() error { closeCalled = true; return nil },
		seedDemoData: func() error { return nil },
		initEvaluator: func(*sql.DB) {
		},
		getDB: func() *sql.DB { return nil },
		listen: func(string, string) (net.Listener, error) {
			return stubListener{addr: stubAddr{network: "tcp", address: "127.0.0.1:1"}}, nil
		},
		serve: func(*http.Server, net.Listener) error {
			return errors.New("boom")
		},
		signalCh: make(chan os.Signal, 1),
	}

	if exitCode := run([]string{"-db", "db.sqlite"}, deps); exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}
	if !closeCalled {
		t.Fatalf("expected closeDB to be called via defer")
	}
}

func TestDefaultRunDeps_PopulatesDependencies(t *testing.T) {
	deps := defaultRunDeps()

	if deps.stdout == nil {
		t.Fatalf("expected stdout to be set")
	}
	if deps.initDB == nil || deps.closeDB == nil {
		t.Fatalf("expected InitDB/CloseDB to be set")
	}
	if deps.seedDemoData == nil {
		t.Fatalf("expected seedDemoData to be set")
	}
	if deps.initEvaluator == nil || deps.getDB == nil {
		t.Fatalf("expected initEvaluator/getDB to be set")
	}
	if deps.listen == nil || deps.serve == nil {
		t.Fatalf("expected listen/serve to be set")
	}
	if deps.setLogOutput == nil {
		t.Fatalf("expected setLogOutput to be set")
	}
	if deps.ensureDemoKey == nil {
		t.Fatalf("expected ensureDemoKey to be set")
	}
	if deps.registerRoutes == nil {
		t.Fatalf("expected registerRoutes to be set")
	}
}

func TestRun_StdoutNil_UsesDiscardAndReturnsMissingDBError(t *testing.T) {
	if exitCode := run([]string{"-seed=false"}, runDeps{}); exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}
}

func TestRun_InitDBError_Returns1(t *testing.T) {
	deps := runDeps{
		stdout:        io.Discard,
		ensureDemoKey: func() {},
		initDB:        func(string) error { return errors.New("boom") },
	}

	if exitCode := run([]string{"-db", "db.sqlite", "-seed=false"}, deps); exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}
}

func TestRun_SeedDemoDataError_Returns1_AndClosesDB(t *testing.T) {
	var closeCalled bool
	deps := runDeps{
		stdout:        io.Discard,
		ensureDemoKey: func() {},
		initDB:        func(string) error { return nil },
		closeDB:       func() error { closeCalled = true; return nil },
		seedDemoData:  func() error { return errors.New("seed failed") },
	}

	if exitCode := run([]string{"-db", "db.sqlite"}, deps); exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}
	if !closeCalled {
		t.Fatalf("expected closeDB to be called via defer")
	}
}

func TestRun_NilSignalChannel_ExitsOnErrServerClosed(t *testing.T) {
	var closeCalled bool
	deps := runDeps{
		stdout:        io.Discard,
		ensureDemoKey: func() {},
		initDB:        func(string) error { return nil },
		closeDB:       func() error { closeCalled = true; return nil },
		initEvaluator: func(*sql.DB) {},
		getDB:         func() *sql.DB { return nil },
		listen: func(network, address string) (net.Listener, error) {
			return stubListener{addr: stubAddr{network: network, address: "127.0.0.1:0"}}, nil
		},
		serve:          func(*http.Server, net.Listener) error { return http.ErrServerClosed },
		registerRoutes: func(*http.ServeMux) {},
	}

	if exitCode := run([]string{"-db", "db.sqlite", "-seed=false"}, deps); exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}
	if !closeCalled {
		t.Fatalf("expected closeDB to be called via defer")
	}
}

func TestDefaultRunDeps_ServeDelegatesToHTTPServer(t *testing.T) {
	deps := defaultRunDeps()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	_ = ln.Close()

	server := &http.Server{Handler: http.NewServeMux()}
	if err := deps.serve(server, ln); err == nil {
		t.Fatalf("expected serve to return an error for closed listener")
	}
}

func TestMain_UsesOsExit(t *testing.T) {
	origArgs := os.Args
	origExit := osExit
	origLogOutput := log.Writer()
	t.Cleanup(func() {
		os.Args = origArgs
		osExit = origExit
		log.SetOutput(origLogOutput)
	})

	os.Args = []string{"demo-server", "-no-such-flag"}

	var gotCode int
	osExit = func(code int) {
		gotCode = code
	}

	main()

	if gotCode != 2 {
		t.Fatalf("expected exit code 2, got %d", gotCode)
	}
}
