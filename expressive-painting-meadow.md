# Automated LLM Evaluation System - Implementation Plan

## Overview

Transform the LLM Tournament Arena from manual scoring to automated evaluation using:
- **3 AI judges**: Claude Opus 4.5 (high think), GPT-5.2 (high think), Gemini 3 Pro (high think)
- **LiteLLM**: Python service for unified API access
- **Weighted consensus**: Aggregate judge scores by confidence
- **Hybrid I/O**: Support both manual and API-based model responses
- **Async execution**: Background jobs with real-time WebSocket progress tracking
- **Cost control**: Estimate and track API costs with alerts

## User Requirements Summary

**Evaluation Methods:**
- Objective prompts: LLM semantic matching against solution field
- Creative prompts: Judge evaluation (no expected answer)

**Evaluation Triggers (all 4):**
- Manual "Evaluate All" button
- Per-model "Evaluate" button
- Per-prompt evaluation across all models
- Auto-evaluate when new models added

**Other:**
- Re-evaluation with confirmation dialog
- UI settings page for encrypted API keys
- Cost tracking with alerts ($100 default threshold)

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    Go Web Server                         │
│                                                           │
│  HTTP Handlers ↔ Job Queue ↔ WebSocket ↔ SQLite DB     │
│       ↓                                                   │
│  Evaluation Orchestrator                                 │
│       ↓ (HTTP)                                           │
└───────┼─────────────────────────────────────────────────┘
        │
        ↓
┌───────────────────────────────────────────────────────────┐
│              Python LiteLLM Service (FastAPI)             │
│                                                            │
│  Endpoints: /evaluate, /evaluate_batch, /estimate_cost   │
│       ↓                                                    │
│  Judge Evaluators (3 parallel):                          │
│    - Claude Opus 4.5 (high think)                        │
│    - GPT-5.2 (high think)                                │
│    - Gemini 3 Pro (high think)                           │
│  Each returns: {score: 0-100, confidence: 0-1, reasoning}│
└───────────────────────────────────────────────────────────┘
```

## Database Schema Changes

### 1. Add `type` field to prompts
```sql
ALTER TABLE prompts ADD COLUMN type TEXT NOT NULL DEFAULT 'objective';
-- Values: 'objective' (semantic matching) or 'creative' (judge evaluation)
```

### 2. New tables

**settings** - Encrypted API keys and config
```sql
CREATE TABLE settings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    key TEXT UNIQUE NOT NULL,           -- e.g., 'api_key_anthropic'
    value TEXT NOT NULL,                -- Encrypted with AES-256-GCM
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**evaluation_jobs** - Async job tracking
```sql
CREATE TABLE evaluation_jobs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    suite_id INTEGER NOT NULL,
    job_type TEXT NOT NULL,             -- 'all', 'model', 'prompt'
    target_id INTEGER,                  -- model_id or prompt_id
    status TEXT NOT NULL DEFAULT 'pending', -- 'pending', 'running', 'completed', 'failed', 'cancelled'
    progress_current INTEGER DEFAULT 0,
    progress_total INTEGER DEFAULT 0,
    estimated_cost_usd REAL DEFAULT 0.0,
    actual_cost_usd REAL DEFAULT 0.0,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    FOREIGN KEY (suite_id) REFERENCES suites(id) ON DELETE CASCADE
);
```

**model_responses** - Store model outputs (hybrid I/O)
```sql
CREATE TABLE model_responses (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    model_id INTEGER NOT NULL,
    prompt_id INTEGER NOT NULL,
    response_text TEXT,                 -- Model's actual response
    response_source TEXT NOT NULL DEFAULT 'manual', -- 'manual' or 'api'
    api_config TEXT,                    -- JSON config for API models
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (model_id) REFERENCES models(id) ON DELETE CASCADE,
    FOREIGN KEY (prompt_id) REFERENCES prompts(id) ON DELETE CASCADE,
    UNIQUE(model_id, prompt_id)
);
```

**evaluation_history** - Audit trail of judge evaluations
```sql
CREATE TABLE evaluation_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    job_id INTEGER NOT NULL,
    model_id INTEGER NOT NULL,
    prompt_id INTEGER NOT NULL,
    judge_name TEXT NOT NULL,           -- 'claude_opus_4.5', 'gpt_5.2', 'gemini_3_pro'
    judge_score INTEGER,                -- 0-100
    judge_confidence REAL,              -- 0.0-1.0
    judge_reasoning TEXT,               -- Judge's explanation
    cost_usd REAL DEFAULT 0.0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (job_id) REFERENCES evaluation_jobs(id) ON DELETE CASCADE,
    FOREIGN KEY (model_id) REFERENCES models(id) ON DELETE CASCADE,
    FOREIGN KEY (prompt_id) REFERENCES prompts(id) ON DELETE CASCADE
);
```

**cost_tracking** - Daily budget monitoring
```sql
CREATE TABLE cost_tracking (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    suite_id INTEGER NOT NULL,
    date DATE NOT NULL,
    total_cost_usd REAL DEFAULT 0.0,
    evaluation_count INTEGER DEFAULT 0,
    FOREIGN KEY (suite_id) REFERENCES suites(id) ON DELETE CASCADE,
    UNIQUE(suite_id, date)
);
```

## Implementation Phases

### Phase 1: Database Foundation ⭐ CRITICAL
**Files:** `middleware/database.go`

1. Add all new tables to `createTables()` function
2. Create migration function `MigrateEvaluationSchema()`
3. Add indexes for performance
4. Initialize default settings (empty API keys, $100 threshold)

**Testing:** Run migration on test database, verify schema

---

### Phase 2: Python LiteLLM Service ⭐ CRITICAL
**New directory:** `python_service/`

**Structure:**
```
python_service/
├── main.py                     # FastAPI server
├── evaluators/
│   ├── base.py                # Base evaluator class
│   ├── objective.py           # Objective evaluation logic
│   └── creative.py            # Creative evaluation logic
├── judges/
│   ├── claude.py              # Claude Opus 4.5 high think
│   ├── gpt.py                 # GPT-5.2 high think
│   └── gemini.py              # Gemini 3 Pro high think
├── prompts/
│   ├── objective_judge.txt    # Judge prompt template
│   └── creative_judge.txt     # Judge prompt template
├── requirements.txt
└── config.py
```

**Key endpoints:**
- `POST /evaluate`: Single prompt evaluation
  - Input: `{prompt, response, solution, type, judges: [...]}`
  - Output: `{results: [{judge, score, confidence, reasoning}], avg_cost}`

- `POST /evaluate_batch`: Batch evaluation

- `GET /estimate_cost`: Token counting

**Judge implementation:**
```python
class Judge:
    def evaluate(self, prompt, response, solution=None, type='objective'):
        # Call LiteLLM with high think mode
        # Parse JSON response: {score, confidence, reasoning}
        return {
            'judge': self.name,
            'score': 0-100,
            'confidence': 0.0-1.0,
            'reasoning': '...'
        }
```

**Testing:** Unit tests for each judge with mock API responses

---

### Phase 3: Go Evaluation Orchestrator ⭐ CRITICAL
**New directory:** `evaluator/`

**Files:**
- `evaluator.go` - Main orchestration logic
- `job_queue.go` - Job queue with workers
- `litellm_client.go` - HTTP client for Python service
- `consensus.go` - Weighted average calculation
- `cost_estimator.go` - Cost estimation

**Key functions:**

```go
// evaluator.go
func (e *Evaluator) EvaluateAll(suiteID int) (jobID int, error)
func (e *Evaluator) EvaluateModel(modelID int) (jobID int, error)
func (e *Evaluator) EvaluatePrompt(promptID int) (jobID int, error)
func (e *Evaluator) processJob(job *EvaluationJob) error

// consensus.go
func CalculateConsensusScore(results []JudgeResult) int {
    // Weighted average by confidence
    weightedSum := 0.0
    totalWeight := 0.0
    for _, r := range results {
        weightedSum += float64(r.Score) * r.Confidence
        totalWeight += r.Confidence
    }
    return int(round(weightedSum / totalWeight))
}
```

**Job queue:**
- In-memory channel with configurable workers (default: 3)
- Jobs persist to database for restart recovery
- Concurrent execution with rate limiting

**Testing:** Integration tests with mock Python service

---

### Phase 4: API Key Management & Security
**New files:**
- `middleware/encryption.go`
- `middleware/settings.go`
- `handlers/settings.go`
- `templates/settings.html`

**Encryption:**
```go
// AES-256-GCM encryption
func EncryptAPIKey(plaintext string) (string, error)
func DecryptAPIKey(ciphertext string) (string, error)
```

**Settings UI:**
- Form fields for each provider's API key (masked display)
- Test button to validate keys
- Cost threshold slider
- Auto-evaluate toggle

**Security:**
- Encryption key from `ENCRYPTION_KEY` env var (32-byte hex)
- Never log decrypted keys
- HTTPS for production

---

### Phase 5: Evaluation Handlers
**New file:** `handlers/evaluation.go`

**Endpoints:**
```go
POST /evaluate/all              # Evaluate all models × all prompts
POST /evaluate/model/:id        # Evaluate one model × all prompts
POST /evaluate/prompt/:id       # Evaluate all models × one prompt
GET  /evaluation/progress/:id   # Get job status
POST /evaluation/cancel/:id     # Cancel job
```

**Flow:**
1. Check for existing scores → show confirmation dialog if re-evaluation
2. Estimate cost → show confirmation dialog
3. Create job in database
4. Queue job for processing
5. Return job ID to client
6. Client polls /evaluation/progress or receives WebSocket updates

---

### Phase 6: WebSocket Progress Updates
**Modified file:** `middleware/socket.go`

**New message types:**
```json
{
  "type": "evaluation_started",
  "data": {"job_id": 123, "estimated_cost": 5.50}
}

{
  "type": "evaluation_progress",
  "data": {"job_id": 123, "current": 10, "total": 100, "cost": 0.50}
}

{
  "type": "evaluation_completed",
  "data": {"job_id": 123, "final_cost": 4.85}
}

{
  "type": "evaluation_failed",
  "data": {"job_id": 123, "error": "API key invalid"}
}

{
  "type": "cost_alert",
  "data": {"suite_id": 1, "current_cost": 105.00, "threshold": 100.00}
}
```

**Broadcast functions:**
```go
func BroadcastEvaluationProgress(jobID, current, total int, cost float64)
func BroadcastEvaluationCompleted(jobID int, finalCost float64)
func BroadcastCostAlert(suiteID int, currentCost, threshold float64)
```

---

### Phase 7: UI Implementation
**Modified files:**
- `templates/results.html` - Add evaluation buttons
- `templates/prompt_list.html` - Add type field dropdown
- `handlers/prompt.go` - Handle type field in CRUD

**New files:**
- `templates/settings.html` - Settings page
- `templates/evaluation_progress.html` - Progress modal

**UI elements:**
1. **Results page:**
   - "Evaluate All" button (top-right)
   - Per-model "Evaluate" button (in model header)
   - Per-prompt "Evaluate" button (in prompt column)
   - Status indicators (pending/running/completed)

2. **Progress modal:**
   - Progress bar
   - Current/total counts (e.g., "15 / 100")
   - Estimated vs actual cost
   - Cancel button

3. **Settings page:**
   - API key inputs (masked: "sk-...abc123")
   - Test button per key
   - Cost threshold slider ($10 - $1000)
   - Auto-evaluate toggle

---

### Phase 8: Cost Control
**File:** `evaluator/cost_estimator.go`

**Cost formulas:**
```go
// Per evaluation (~500 input + 200 output tokens per judge)
const (
    ClaudeOpus45Cost  = (500*0.015 + 200*0.075) / 1000  // $0.0225
    GPT52Cost         = (500*0.010 + 200*0.030) / 1000  // $0.0110
    Gemini3ProCost    = (500*0.0125 + 200*0.050) / 1000 // $0.0163
)

// Total per evaluation: ~$0.05 (3 judges)
```

**Example costs:**
- 10 models × 42 prompts = 420 evaluations × $0.05 = **$21**
- 50 models × 100 prompts = 5,000 evaluations × $0.05 = **$250**

**Controls:**
1. Estimate before job creation
2. Show confirmation dialog with cost
3. Track actual cost per API call
4. Alert via WebSocket if threshold exceeded
5. Allow job cancellation

---

## Critical Files to Create/Modify

### New Files (Priority Order):

1. **`middleware/database.go`** (MODIFY)
   - Add new tables to `createTables()`
   - Add migration function

2. **`python_service/main.py`** (CREATE)
   - FastAPI server with evaluation endpoints

3. **`evaluator/evaluator.go`** (CREATE)
   - Main orchestration logic

4. **`evaluator/litellm_client.go`** (CREATE)
   - HTTP client for Python service

5. **`evaluator/consensus.go`** (CREATE)
   - Weighted consensus calculation

6. **`handlers/evaluation.go`** (CREATE)
   - Evaluation trigger handlers

7. **`middleware/socket.go`** (MODIFY)
   - Add new WebSocket message types

8. **`handlers/settings.go`** (CREATE)
   - Settings CRUD handlers

9. **`middleware/encryption.go`** (CREATE)
   - API key encryption/decryption

10. **`templates/settings.html`** (CREATE)
    - Settings UI page

### Existing Files to Modify:

- `main.go` - Add routes for `/settings`, `/evaluate/*`
- `templates/results.html` - Add evaluation buttons
- `handlers/prompt.go` - Add type field to CRUD

---

## Key Design Decisions

### 1. Weighted Consensus by Confidence
Judges with higher confidence have more influence:
```
Example:
Claude: score=85, confidence=0.9
GPT:    score=90, confidence=0.7
Gemini: score=80, confidence=0.8

Weighted avg = (85×0.9 + 90×0.7 + 80×0.8) / (0.9+0.7+0.8) = 85
```

### 2. Objective vs Creative Evaluation
- **Objective:** Judge compares response to solution field (semantic match)
- **Creative:** Judge evaluates quality without expected answer
- Different judge prompt templates for each type

### 3. Hybrid Model I/O
- Manual mode: User provides responses manually (current workflow)
- API mode: System fetches responses via LiteLLM
- Tracked via `response_source` field in `model_responses` table

### 4. In-Memory Job Queue
- No external dependencies (Redis, RabbitMQ)
- Jobs persist to database (survive restarts)
- Configurable worker count (default: 3)
- Rate limiting per judge to avoid API limits

### 5. API Key Security
- AES-256-GCM encryption
- Encryption key from `ENCRYPTION_KEY` env var
- Never log decrypted keys
- Masked display in UI

---

## Risks & Mitigations

### Risk: API Rate Limits
**Mitigation:**
- Exponential backoff with jitter
- Limit concurrent calls (3 workers)
- Track rate limits per provider

### Risk: High Costs
**Mitigation:**
- Estimate before execution (show confirmation)
- Track actual costs in real-time
- Alert if threshold exceeded
- Allow cancellation

### Risk: Python Service Crashes
**Mitigation:**
- Health check endpoint (`/health`)
- Retry failed requests (3 attempts)
- Store job status in database (resume after restart)

### Risk: Long Evaluation Times
**Mitigation:**
- Real-time progress updates via WebSocket
- Show estimated time remaining
- Allow cancellation
- Non-blocking background execution

---

## Testing Strategy

### Unit Tests:
- Encryption/decryption
- Consensus algorithm
- Cost estimation

### Integration Tests:
- Go ↔ Python HTTP communication (mock service)
- Database CRUD for new tables
- WebSocket broadcasting
- End-to-end evaluation flow (mock judges)

### Manual Testing:
- UI: Settings → Add keys → Trigger evaluation → Monitor progress
- Error handling: Invalid keys, service down, rate limits
- Cost alerts: Exceed threshold
- Cancellation: Cancel mid-execution

---

## Deployment

### Development:
```bash
# 1. Install Python dependencies
pip install -r python_service/requirements.txt

# 2. Start Python service
python python_service/main.py  # Port 8001

# 3. Set encryption key
export ENCRYPTION_KEY=$(openssl rand -hex 32)

# 4. Run migration
go run main.go --migrate-evaluation-schema

# 5. Start Go server
go run main.go
```

### Production:
```bash
# 1. Run migration
./llm-tournament --migrate-evaluation-schema --db=data/tournament.db

# 2. Build Python Docker image
docker build -t llm-eval-service python_service/

# 3. Run Python container
docker run -d -p 8001:8001 llm-eval-service

# 4. Set environment
export ENCRYPTION_KEY=<32-byte-hex>
export PYTHON_SERVICE_URL=http://localhost:8001

# 5. Start Go server
./llm-tournament --db=data/tournament.db
```

---

## Migration Path

### Backward Compatibility:
- Existing prompts: Default `type='objective'`
- Existing scores: Unchanged, not affected by evaluation system
- Manual scoring: Still works (existing handlers unchanged)

### Rollback:
If issues arise:
1. Stop Go server
2. Restore database backup
3. Revert to previous binary
4. Investigate logs, fix, retry

---

## Estimated Costs

### Development Time:
- Phase 1 (Database): 1 day
- Phase 2 (Python service): 3 days
- Phase 3 (Orchestrator): 3 days
- Phase 4 (Security): 2 days
- Phase 5 (Handlers): 2 days
- Phase 6 (WebSocket): 1 day
- Phase 7 (UI): 3 days
- Phase 8 (Cost control): 1 day

**Total: ~16 days** (assumes TDD approach)

### API Costs (per evaluation):
- ~$0.05 per model-prompt pair (3 judges)
- Example: 10 models × 42 prompts = $21
- Recommendation: Set $100/day default threshold

---

## Success Criteria

✅ Users can configure API keys via UI settings page
✅ Users can trigger evaluations (all/model/prompt)
✅ System fetches responses via API (for configured models)
✅ 3 judges evaluate each response independently
✅ Consensus score calculated by weighted confidence
✅ Objective prompts match against solution field
✅ Creative prompts evaluated without expected answer
✅ Real-time progress updates via WebSocket
✅ Cost estimation before execution
✅ Cost tracking with alerts
✅ Job cancellation mid-execution
✅ Re-evaluation with confirmation
✅ Manual scoring still works (backward compatible)
✅ All data persists correctly in SQLite

---

## Next Steps

After plan approval:
1. Create feature branch: `git checkout -b feature/automated-evaluation`
2. Start with Phase 1 (database schema)
3. Build Python service (Phase 2)
4. Implement Go orchestrator (Phase 3)
5. Add security layer (Phase 4)
6. Wire up handlers (Phase 5)
7. Add WebSocket updates (Phase 6)
8. Build UI (Phase 7)
9. Add cost controls (Phase 8)
10. Test thoroughly at each phase
11. Create pull request with full test coverage
