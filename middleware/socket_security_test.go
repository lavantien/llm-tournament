package middleware

import (
	"net/http"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

// serverHost extracts the host:port portion of a ws:// test URL.
func serverHost(wsURL string) string {
	return strings.TrimPrefix(wsURL, "ws://")
}

func TestHandleWebSocket_MismatchedOriginRejected(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	if err := InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	clientsMutex.Lock()
	clients = make(map[*websocket.Conn]bool)
	clientsMutex.Unlock()

	server, wsURL := createWebSocketTestServer(t, HandleWebSocket)
	defer server.Close()

	header := http.Header{}
	header.Set("Origin", "http://evil.example")
	conn, resp, err := websocket.DefaultDialer.Dial(wsURL, header)
	if err == nil {
		_ = conn.Close()
		t.Fatal("expected dial with cross-origin Origin header to be rejected")
	}
	if resp == nil {
		t.Fatal("expected an HTTP response alongside the rejection")
	}
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("expected status %d for rejected origin, got %d", http.StatusForbidden, resp.StatusCode)
	}

	clientsMutex.Lock()
	got := len(clients)
	clientsMutex.Unlock()
	if got != 0 {
		t.Fatalf("expected 0 clients after rejected origin, got %d", got)
	}
}

func TestHandleWebSocket_EmptyOriginAllowed(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	if err := InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	server, wsURL := createWebSocketTestServer(t, HandleWebSocket)
	defer server.Close()

	// Non-browser clients (CLI tools, tests) send no Origin header.
	conn, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("expected dial without Origin header to succeed: %v", err)
	}
	defer func() { _ = conn.Close() }()

	if resp.StatusCode != http.StatusSwitchingProtocols {
		t.Errorf("expected status %d, got %d", http.StatusSwitchingProtocols, resp.StatusCode)
	}

	waitForWebSocketClientRegistration(t, 1)
}

func TestHandleWebSocket_SameOriginAllowed(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	if err := InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	server, wsURL := createWebSocketTestServer(t, HandleWebSocket)
	defer server.Close()

	// Origin host matches the request host (the browser's own page).
	header := http.Header{}
	header.Set("Origin", "http://"+serverHost(wsURL))
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, header)
	if err != nil {
		t.Fatalf("expected dial with matching Origin header to succeed: %v", err)
	}
	defer func() { _ = conn.Close() }()

	waitForWebSocketClientRegistration(t, 1)
}
