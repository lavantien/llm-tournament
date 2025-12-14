package middleware

import (
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestCalculatePassPercentages(t *testing.T) {
	tests := []struct {
		name        string
		results     map[string]Result
		promptCount int
		want        map[string]float64
	}{
		{
			name:        "empty results",
			results:     map[string]Result{},
			promptCount: 0,
			want:        map[string]float64{},
		},
		{
			name: "all zeros",
			results: map[string]Result{
				"Model A": {Scores: []int{0, 0, 0}},
			},
			promptCount: 3,
			want: map[string]float64{
				"Model A": 0,
			},
		},
		{
			name: "all 100s",
			results: map[string]Result{
				"Model A": {Scores: []int{100, 100, 100}},
			},
			promptCount: 3,
			want: map[string]float64{
				"Model A": 100,
			},
		},
		{
			name: "mixed scores",
			results: map[string]Result{
				"Model A": {Scores: []int{100, 50, 0}},
			},
			promptCount: 3,
			want: map[string]float64{
				"Model A": 50, // 150/300 * 100 = 50
			},
		},
		{
			name: "multiple models",
			results: map[string]Result{
				"Model A": {Scores: []int{100, 100}},
				"Model B": {Scores: []int{50, 50}},
			},
			promptCount: 2,
			want: map[string]float64{
				"Model A": 100,
				"Model B": 50,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculatePassPercentages(tt.results, tt.promptCount)

			if len(got) != len(tt.want) {
				t.Errorf("expected %d results, got %d", len(tt.want), len(got))
			}

			for model, expected := range tt.want {
				if got[model] != expected {
					t.Errorf("model %q: expected %.2f, got %.2f", model, expected, got[model])
				}
			}
		})
	}
}

func TestCalculatePassPercentages_ZeroPrompts(t *testing.T) {
	results := map[string]Result{
		"Model A": {Scores: []int{}},
	}

	got := calculatePassPercentages(results, 0)

	// With zero prompts, percentage is NaN due to division by zero
	if !math.IsNaN(got["Model A"]) {
		t.Errorf("expected NaN for zero prompts, got %.2f", got["Model A"])
	}
}

func TestPromptsToStringArray(t *testing.T) {
	tests := []struct {
		name    string
		prompts []Prompt
		want    []string
	}{
		{
			name:    "empty prompts",
			prompts: []Prompt{},
			want:    []string{},
		},
		{
			name: "single prompt",
			prompts: []Prompt{
				{Text: "Prompt 1"},
			},
			want: []string{"Prompt 1"},
		},
		{
			name: "multiple prompts",
			prompts: []Prompt{
				{Text: "Prompt 1"},
				{Text: "Prompt 2"},
				{Text: "Prompt 3"},
			},
			want: []string{"Prompt 1", "Prompt 2", "Prompt 3"},
		},
		{
			name: "prompts with solutions and profiles",
			prompts: []Prompt{
				{Text: "Q1", Solution: "A1", Profile: "P1"},
				{Text: "Q2", Solution: "A2", Profile: "P2"},
			},
			want: []string{"Q1", "Q2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := promptsToStringArray(tt.prompts)

			if len(got) != len(tt.want) {
				t.Errorf("expected %d strings, got %d", len(tt.want), len(got))
			}

			for i, expected := range tt.want {
				if got[i] != expected {
					t.Errorf("index %d: expected %q, got %q", i, expected, got[i])
				}
			}
		})
	}
}

func TestProfileGroup_Struct(t *testing.T) {
	pg := ProfileGroup{
		ID:       "1",
		Name:     "Test Profile",
		StartCol: 0,
		EndCol:   5,
		Color:    "hsl(137, 70%, 50%)",
	}

	if pg.ID != "1" {
		t.Errorf("expected ID '1', got %q", pg.ID)
	}
	if pg.Name != "Test Profile" {
		t.Errorf("expected Name 'Test Profile', got %q", pg.Name)
	}
	if pg.StartCol != 0 {
		t.Errorf("expected StartCol 0, got %d", pg.StartCol)
	}
	if pg.EndCol != 5 {
		t.Errorf("expected EndCol 5, got %d", pg.EndCol)
	}
	if pg.Color != "hsl(137, 70%, 50%)" {
		t.Errorf("expected Color 'hsl(137, 70%%, 50%%)', got %q", pg.Color)
	}
}

// Helper function to create WebSocket test server
func createWebSocketTestServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, string) {
	t.Helper()
	server := httptest.NewServer(handler)
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	return server, wsURL
}

func TestHandleWebSocket_Connection(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	server, wsURL := createWebSocketTestServer(t, HandleWebSocket)
	defer server.Close()

	// Connect to WebSocket
	conn, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to connect to WebSocket: %v", err)
	}
	defer conn.Close()

	if resp.StatusCode != http.StatusSwitchingProtocols {
		t.Errorf("expected status %d, got %d", http.StatusSwitchingProtocols, resp.StatusCode)
	}

	// Verify client was registered
	clientsMutex.Lock()
	clientCount := len(clients)
	clientsMutex.Unlock()

	if clientCount != 1 {
		t.Errorf("expected 1 client, got %d", clientCount)
	}
}

func TestHandleWebSocket_CloseConnection(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	server, wsURL := createWebSocketTestServer(t, HandleWebSocket)
	defer server.Close()

	// Connect
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}

	// Close connection
	conn.Close()

	// Wait for cleanup
	time.Sleep(100 * time.Millisecond)

	// Verify client was removed
	clientsMutex.Lock()
	clientCount := len(clients)
	clientsMutex.Unlock()

	if clientCount != 0 {
		t.Errorf("expected 0 clients after close, got %d", clientCount)
	}
}

func TestHandleWebSocket_UpdatePromptsOrder(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// Create prompts
	prompts := []Prompt{
		{Text: "Prompt 1"},
		{Text: "Prompt 2"},
	}
	err = WritePromptSuite("default", prompts)
	if err != nil {
		t.Fatalf("WritePromptSuite failed: %v", err)
	}

	server, wsURL := createWebSocketTestServer(t, HandleWebSocket)
	defer server.Close()

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	// Send update_prompts_order message
	msg := map[string]interface{}{
		"type":  "update_prompts_order",
		"order": []int{2, 1},
	}
	err = conn.WriteJSON(msg)
	if err != nil {
		t.Fatalf("failed to send message: %v", err)
	}

	// Wait for processing
	time.Sleep(100 * time.Millisecond)
}

func TestHandleWebSocket_UnknownMessageType(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	server, wsURL := createWebSocketTestServer(t, HandleWebSocket)
	defer server.Close()

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	// Send unknown message type
	msg := map[string]interface{}{
		"type": "unknown_type",
	}
	err = conn.WriteJSON(msg)
	if err != nil {
		t.Fatalf("failed to send message: %v", err)
	}

	// Should not crash - wait a bit
	time.Sleep(50 * time.Millisecond)
}

func TestBroadcastResults_NoClients(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// Clear any existing clients
	clientsMutex.Lock()
	clients = make(map[*websocket.Conn]bool)
	clientsMutex.Unlock()

	// Should not panic with no clients
	BroadcastResults()
}

func TestBroadcastResults_WithClient(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// Create some test data
	prompts := []Prompt{
		{Text: "Test Prompt"},
	}
	err = WritePromptSuite("default", prompts)
	if err != nil {
		t.Fatalf("WritePromptSuite failed: %v", err)
	}

	server, wsURL := createWebSocketTestServer(t, HandleWebSocket)
	defer server.Close()

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	// Set read deadline
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))

	// Trigger broadcast
	go BroadcastResults()

	// Read the broadcasted message
	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read message: %v", err)
	}

	// Verify message structure
	var payload struct {
		Type string `json:"type"`
		Data struct {
			Results map[string]Result `json:"results"`
			Models  []string          `json:"models"`
		} `json:"data"`
	}
	err = json.Unmarshal(msg, &payload)
	if err != nil {
		t.Fatalf("failed to unmarshal message: %v", err)
	}

	if payload.Type != "results" {
		t.Errorf("expected type 'results', got %q", payload.Type)
	}
}

func TestBroadcastEvaluationProgress(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	server, wsURL := createWebSocketTestServer(t, HandleWebSocket)
	defer server.Close()

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(2 * time.Second))

	// Trigger broadcast
	go BroadcastEvaluationProgress(1, 5, 10, 0.50)

	// Read message
	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read message: %v", err)
	}

	var payload struct {
		Type string `json:"type"`
		Data struct {
			JobID   int     `json:"job_id"`
			Current int     `json:"current"`
			Total   int     `json:"total"`
			Cost    float64 `json:"cost"`
		} `json:"data"`
	}
	err = json.Unmarshal(msg, &payload)
	if err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if payload.Type != "evaluation_progress" {
		t.Errorf("expected type 'evaluation_progress', got %q", payload.Type)
	}
	if payload.Data.JobID != 1 {
		t.Errorf("expected job_id 1, got %d", payload.Data.JobID)
	}
	if payload.Data.Current != 5 {
		t.Errorf("expected current 5, got %d", payload.Data.Current)
	}
	if payload.Data.Total != 10 {
		t.Errorf("expected total 10, got %d", payload.Data.Total)
	}
}

func TestBroadcastEvaluationCompleted(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	server, wsURL := createWebSocketTestServer(t, HandleWebSocket)
	defer server.Close()

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(2 * time.Second))

	go BroadcastEvaluationCompleted(1, 1.50)

	// Read the completion message
	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read message: %v", err)
	}

	var payload struct {
		Type string `json:"type"`
		Data struct {
			JobID     int     `json:"job_id"`
			FinalCost float64 `json:"final_cost"`
		} `json:"data"`
	}
	err = json.Unmarshal(msg, &payload)
	if err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if payload.Type != "evaluation_completed" {
		t.Errorf("expected type 'evaluation_completed', got %q", payload.Type)
	}
	if payload.Data.JobID != 1 {
		t.Errorf("expected job_id 1, got %d", payload.Data.JobID)
	}
}

func TestBroadcastEvaluationFailed(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	server, wsURL := createWebSocketTestServer(t, HandleWebSocket)
	defer server.Close()

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(2 * time.Second))

	go BroadcastEvaluationFailed(1, "test error")

	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read message: %v", err)
	}

	var payload struct {
		Type string `json:"type"`
		Data struct {
			JobID int    `json:"job_id"`
			Error string `json:"error"`
		} `json:"data"`
	}
	err = json.Unmarshal(msg, &payload)
	if err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if payload.Type != "evaluation_failed" {
		t.Errorf("expected type 'evaluation_failed', got %q", payload.Type)
	}
	if payload.Data.Error != "test error" {
		t.Errorf("expected error 'test error', got %q", payload.Data.Error)
	}
}

func TestBroadcastCostAlert(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	server, wsURL := createWebSocketTestServer(t, HandleWebSocket)
	defer server.Close()

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(2 * time.Second))

	go BroadcastCostAlert(1, 95.0, 100.0)

	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read message: %v", err)
	}

	var payload struct {
		Type string `json:"type"`
		Data struct {
			SuiteID     int     `json:"suite_id"`
			CurrentCost float64 `json:"current_cost"`
			Threshold   float64 `json:"threshold"`
		} `json:"data"`
	}
	err = json.Unmarshal(msg, &payload)
	if err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if payload.Type != "cost_alert" {
		t.Errorf("expected type 'cost_alert', got %q", payload.Type)
	}
	if payload.Data.CurrentCost != 95.0 {
		t.Errorf("expected current_cost 95.0, got %f", payload.Data.CurrentCost)
	}
}
