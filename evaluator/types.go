package evaluator

import "time"

// EvaluationJob represents an evaluation job
type EvaluationJob struct {
	ID              int
	SuiteID         int
	JobType         string // 'all', 'model', 'prompt'
	TargetID        int
	Status          string
	ProgressCurrent int
	ProgressTotal   int
	EstimatedCost   float64
	ActualCost      float64
	ErrorMessage    string
	CreatedAt       time.Time
	StartedAt       *time.Time
	CompletedAt     *time.Time
}

// JudgeResult represents the result from a single judge
type JudgeResult struct {
	Judge      string  `json:"judge"`
	Score      int     `json:"score"`
	Confidence float64 `json:"confidence"`
	Reasoning  string  `json:"reasoning"`
	CostUSD    float64 `json:"cost_usd"`
	Error      string  `json:"error,omitempty"`
}

// EvaluationRequest represents a request to the Python service
type EvaluationRequest struct {
	Prompt   string            `json:"prompt"`
	Response string            `json:"response"`
	Solution string            `json:"solution"`
	Type     string            `json:"type"`
	Judges   []string          `json:"judges"`
	APIKeys  map[string]string `json:"api_keys"`
}

// EvaluationResponse represents a response from the Python service
type EvaluationResponse struct {
	Results        []JudgeResult `json:"results"`
	TotalCostUSD   float64       `json:"total_cost_usd"`
	ConsensusScore int           `json:"consensus_score"`
	AvgConfidence  float64       `json:"avg_confidence"`
}

// CostEstimateRequest represents a cost estimation request
type CostEstimateRequest struct {
	Prompt   string   `json:"prompt"`
	Response string   `json:"response"`
	Solution string   `json:"solution"`
	Type     string   `json:"type"`
	Judges   []string `json:"judges"`
}

// CostEstimateResponse represents a cost estimation response
type CostEstimateResponse struct {
	EstimatedCostUSD float64            `json:"estimated_cost_usd"`
	Breakdown        map[string]float64 `json:"breakdown"`
}
