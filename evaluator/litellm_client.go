package evaluator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// LiteLLMClient handles communication with Python evaluation service
type LiteLLMClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewLiteLLMClient creates a new LiteLLM client
func NewLiteLLMClient(baseURL string) *LiteLLMClient {
	return &LiteLLMClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 180 * time.Second, // 3 minutes for thinking models
		},
	}
}

// Evaluate sends an evaluation request to the Python service
func (c *LiteLLMClient) Evaluate(req EvaluationRequest) (*EvaluationResponse, error) {
	jsonData, _ := json.Marshal(req)

	resp, err := c.httpClient.Post(
		c.baseURL+"/evaluate",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("evaluation failed with status %d: %s", resp.StatusCode, string(body))
	}

	var evalResp EvaluationResponse
	if err := json.NewDecoder(resp.Body).Decode(&evalResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &evalResp, nil
}

// EstimateCost estimates the cost of an evaluation
func (c *LiteLLMClient) EstimateCost(req CostEstimateRequest) (*CostEstimateResponse, error) {
	jsonData, _ := json.Marshal(req)

	resp, err := c.httpClient.Post(
		c.baseURL+"/estimate_cost",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("cost estimation failed with status %d: %s", resp.StatusCode, string(body))
	}

	var costResp CostEstimateResponse
	if err := json.NewDecoder(resp.Body).Decode(&costResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &costResp, nil
}

// HealthCheck checks if the Python service is healthy
func (c *LiteLLMClient) HealthCheck() error {
	resp, err := c.httpClient.Get(c.baseURL + "/health")
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("service unhealthy, status: %d", resp.StatusCode)
	}

	return nil
}
