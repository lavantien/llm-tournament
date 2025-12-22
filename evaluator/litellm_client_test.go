package evaluator

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewLiteLLMClient(t *testing.T) {
	client := NewLiteLLMClient("http://localhost:8001")

	if client == nil {
		t.Fatal("NewLiteLLMClient returned nil")
	}
	if client.baseURL != "http://localhost:8001" {
		t.Errorf("expected baseURL 'http://localhost:8001', got %q", client.baseURL)
	}
	if client.httpClient == nil {
		t.Error("httpClient should not be nil")
	}
	if client.httpClient.Timeout != 180*1e9 { // 180 seconds in nanoseconds
		t.Errorf("expected 180s timeout, got %v", client.httpClient.Timeout)
	}
}

func TestEvaluate_Success(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/evaluate" {
			t.Errorf("expected /evaluate path, got %s", r.URL.Path)
		}

		// Verify request body
		var req EvaluationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		// Send response
		resp := EvaluationResponse{
			Results: []JudgeResult{
				{Judge: "claude", Score: 80, Confidence: 0.9, Reasoning: "Good response"},
			},
			TotalCostUSD:   0.05,
			ConsensusScore: 80,
			AvgConfidence:  0.9,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewLiteLLMClient(server.URL)
	req := EvaluationRequest{
		Prompt:   "Test prompt",
		Response: "Test response",
		Solution: "Expected solution",
		Type:     "objective",
		Judges:   []string{"claude"},
	}

	resp, err := client.Evaluate(req)
	if err != nil {
		t.Fatalf("Evaluate failed: %v", err)
	}

	if len(resp.Results) != 1 {
		t.Errorf("expected 1 result, got %d", len(resp.Results))
	}
	if resp.Results[0].Score != 80 {
		t.Errorf("expected score 80, got %d", resp.Results[0].Score)
	}
	if resp.ConsensusScore != 80 {
		t.Errorf("expected consensus score 80, got %d", resp.ConsensusScore)
	}
}

func TestEvaluate_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("Internal server error"))
	}))
	defer server.Close()

	client := NewLiteLLMClient(server.URL)
	req := EvaluationRequest{
		Prompt: "Test prompt",
	}

	_, err := client.Evaluate(req)
	if err == nil {
		t.Error("expected error for server error response")
	}
}

func TestEvaluate_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("not valid json"))
	}))
	defer server.Close()

	client := NewLiteLLMClient(server.URL)
	req := EvaluationRequest{
		Prompt: "Test prompt",
	}

	_, err := client.Evaluate(req)
	if err == nil {
		t.Error("expected error for invalid JSON response")
	}
}

func TestEvaluate_ConnectionError(t *testing.T) {
	client := NewLiteLLMClient("http://localhost:99999")
	req := EvaluationRequest{
		Prompt: "Test prompt",
	}

	_, err := client.Evaluate(req)
	if err == nil {
		t.Error("expected error for connection failure")
	}
}

func TestEstimateCost_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/estimate_cost" {
			t.Errorf("expected /estimate_cost path, got %s", r.URL.Path)
		}

		resp := CostEstimateResponse{
			EstimatedCostUSD: 0.15,
			Breakdown: map[string]float64{
				"claude": 0.05,
				"gpt":    0.05,
				"gemini": 0.05,
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewLiteLLMClient(server.URL)
	req := CostEstimateRequest{
		Prompt:   "Test prompt",
		Response: "Test response",
		Judges:   []string{"claude", "gpt", "gemini"},
	}

	resp, err := client.EstimateCost(req)
	if err != nil {
		t.Fatalf("EstimateCost failed: %v", err)
	}

	if resp.EstimatedCostUSD != 0.15 {
		t.Errorf("expected estimated cost 0.15, got %f", resp.EstimatedCostUSD)
	}
	if len(resp.Breakdown) != 3 {
		t.Errorf("expected 3 breakdown entries, got %d", len(resp.Breakdown))
	}
}

func TestEstimateCost_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Bad request"))
	}))
	defer server.Close()

	client := NewLiteLLMClient(server.URL)
	req := CostEstimateRequest{}

	_, err := client.EstimateCost(req)
	if err == nil {
		t.Error("expected error for server error response")
	}
}

func TestEstimateCost_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("not valid json"))
	}))
	defer server.Close()

	client := NewLiteLLMClient(server.URL)
	req := CostEstimateRequest{
		Prompt: "Test prompt",
	}

	_, err := client.EstimateCost(req)
	if err == nil {
		t.Error("expected error for invalid JSON response")
	}
}

func TestEstimateCost_ConnectionError(t *testing.T) {
	client := NewLiteLLMClient("http://localhost:99999")
	req := CostEstimateRequest{
		Prompt: "Test prompt",
	}

	_, err := client.EstimateCost(req)
	if err == nil {
		t.Error("expected error for connection failure")
	}
}

func TestHealthCheck_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/health" {
			t.Errorf("expected /health path, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status": "healthy"}`))
	}))
	defer server.Close()

	client := NewLiteLLMClient(server.URL)
	err := client.HealthCheck()
	if err != nil {
		t.Fatalf("HealthCheck failed: %v", err)
	}
}

func TestHealthCheck_Unhealthy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	client := NewLiteLLMClient(server.URL)
	err := client.HealthCheck()
	if err == nil {
		t.Error("expected error for unhealthy service")
	}
}

func TestHealthCheck_ConnectionError(t *testing.T) {
	client := NewLiteLLMClient("http://localhost:99999")
	err := client.HealthCheck()
	if err == nil {
		t.Error("expected error for connection failure")
	}
}

func TestEvaluate_MultipleJudges(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := EvaluationResponse{
			Results: []JudgeResult{
				{Judge: "claude", Score: 80, Confidence: 0.9},
				{Judge: "gpt", Score: 70, Confidence: 0.8},
				{Judge: "gemini", Score: 75, Confidence: 0.85},
			},
			TotalCostUSD:   0.15,
			ConsensusScore: 75,
			AvgConfidence:  0.85,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewLiteLLMClient(server.URL)
	req := EvaluationRequest{
		Prompt:   "Test prompt",
		Response: "Test response",
		Judges:   []string{"claude", "gpt", "gemini"},
		APIKeys:  map[string]string{"anthropic": "sk-ant", "openai": "sk-oai"},
	}

	resp, err := client.Evaluate(req)
	if err != nil {
		t.Fatalf("Evaluate failed: %v", err)
	}

	if len(resp.Results) != 3 {
		t.Errorf("expected 3 results, got %d", len(resp.Results))
	}
}
