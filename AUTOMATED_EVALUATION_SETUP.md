# Automated LLM Evaluation - Setup Guide

## Overview

The LLM Tournament Arena includes automated evaluation powered by AI judges (Claude Opus 4.5, GPT-5.2, Gemini 3 Pro). This guide covers installation, configuration, and usage.

## Prerequisites

- Python 3.8+ (for AI judge service)
- Go 1.24+ with CGO_ENABLED=1 (for main server)
- API keys for at least one AI provider:
  - Anthropic (Claude)
  - OpenAI (GPT)
  - Google (Gemini)

## 🚀 Installation

### 1. Install Python Dependencies

```bash
cd python_service
pip install -r requirements.txt
```

### 2. Generate Encryption Key

```bash
# Windows (PowerShell)
$key = -join ((0..31) | ForEach-Object { '{0:x2}' -f (Get-Random -Maximum 256) })
echo $key

# Linux/Mac
openssl rand -hex 32
```

### 3. Set Environment Variables

```bash
# Windows
set ENCRYPTION_KEY=<your-32-byte-hex-key>

# Linux/Mac
export ENCRYPTION_KEY=<your-32-byte-hex-key>
```

### 4. Start Python Service

```bash
cd python_service
python main.py
```

Expected output:
```
INFO:     Started server process
INFO:     Waiting for application startup.
INFO:     Application startup complete.
INFO:     Uvicorn running on http://0.0.0.0:8001
```

### 5. Start Go Server (in new terminal)

```bash
CGO_ENABLED=1 go run main.go
```

Expected output:
```
Initializing database...
Initializing evaluator...
Evaluator initialized with Python service URL: http://localhost:8001
Starting the server...
Server is listening on :8080
```

### 6. Configure API Keys

1. Open browser: `http://localhost:8080/settings`
2. Enter your API keys:
   - **Anthropic API Key** (for Claude Opus 4.5)
   - **OpenAI API Key** (for GPT-5.2)
   - **Google API Key** (for Gemini 3 Pro)
3. Set cost alert threshold (default: $100)
4. Save settings

---

## 📋 Usage Guide

### Triggering Evaluations

The evaluation system uses AI judges to score model responses. Before triggering an evaluation, ensure you have:
1. Added models to your suite
2. Created prompts with the appropriate type (`objective` or `creative`)
3. Configured API keys in settings

#### 1. Evaluate All
- **Button**: "Evaluate All" (top-right of Results page)
- **Action**: Evaluates all models × all prompts in current suite
- **Endpoint**: `POST /evaluate/all`
- **Cost Estimate**: Shown before execution

#### 2. Evaluate Per-Model
- **Button**: "Evaluate" (in model row header)
- **Action**: Evaluates one model × all prompts
- **Endpoint**: `POST /evaluate/model?id={model_id}`

#### 3. Evaluate Per-Prompt
- **Button**: "Evaluate" (in prompt column header)
- **Action**: Evaluates all models × one prompt
- **Endpoint**: `POST /evaluate/prompt?id={prompt_id}`

#### 4. Auto-Evaluate New Models
- **Setting**: Enable in `/settings`
- **Action**: Automatically evaluates when new models are added

### Monitoring Progress

Progress updates are broadcast via WebSocket:

```javascript
// Client-side (already implemented in templates)
ws.onmessage = (event) => {
  const msg = JSON.parse(event.data);

  switch (msg.type) {
    case 'evaluation_progress':
      // Update progress bar: msg.data.current / msg.data.total
      // Show cost: msg.data.cost
      break;

    case 'evaluation_completed':
      // Show success message
      // Final cost: msg.data.final_cost
      break;

    case 'evaluation_failed':
      // Show error: msg.data.error
      break;

    case 'cost_alert':
      // Warn user: threshold exceeded
      break;
  }
};
```

---

## 🧪 Testing the System

### 1. Health Check

```bash
# Test Python service
curl http://localhost:8001/health

# Expected: {"status":"healthy","service":"llm-evaluation"}
```

### 2. Cost Estimation

```bash
curl -X POST http://localhost:8001/estimate_cost \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "What is 2+2?",
    "response": "4",
    "solution": "4",
    "type": "objective",
    "judges": ["claude_opus_4.5", "gpt_5.2", "gemini_3_pro"]
  }'

# Expected: {"estimated_cost_usd":0.05,"breakdown":{...}}
```

### 3. Single Evaluation (with fake keys for testing)

```bash
curl -X POST http://localhost:8001/evaluate \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "What is the capital of France?",
    "response": "Paris is the capital of France.",
    "solution": "Paris",
    "type": "objective",
    "judges": ["claude_opus_4.5"],
    "api_keys": {
      "api_key_anthropic": "your-api-key-here"
    }
  }'
```

### 4. Test Encryption

```bash
# In Go code or via API:
# middleware.EncryptAPIKey("sk-test-key-12345")
# Expected: Base64-encoded encrypted string

# middleware.DecryptAPIKey(<encrypted-string>)
# Expected: "sk-test-key-12345"
```

---

## 🔒 Security Notes

### API Key Encryption
- **Algorithm**: AES-256-GCM
- **Key Source**: `ENCRYPTION_KEY` environment variable (32-byte hex = 64 characters)
- **Storage**: Encrypted values in `settings` table
- **Display**: Masked in UI (shows only last 4 characters)

### Best Practices
1. **Never commit** `ENCRYPTION_KEY` to version control
2. **Rotate keys** periodically (requires re-encrypting all stored keys)
3. **Use HTTPS** in production
4. **Set restrictive CORS** in production (currently allows all origins for dev)
5. **Monitor costs** via the cost tracking dashboard

**Cost Note**: Each evaluation costs approximately $0.05 (using 3 AI judges). Cost estimates are shown before execution, and you can set alert thresholds in settings.

---

## 🐛 Troubleshooting

### Python Service Won't Start

```bash
# Check port availability
netstat -an | grep 8001

# Check Python version (requires 3.8+)
python --version

# Reinstall dependencies
pip install --force-reinstall -r requirements.txt
```

### Encryption Key Error

```
Error: ENCRYPTION_KEY environment variable not set
```

**Solution**: Set the 64-character hex key:
```bash
export ENCRYPTION_KEY=$(openssl rand -hex 32)
```

### Database Migration Issues

```bash
# Backup first
cp data/tournament.db data/tournament.db.backup

# Check schema
sqlite3 data/tournament.db ".schema"

# Manually add missing columns if needed
sqlite3 data/tournament.db "ALTER TABLE prompts ADD COLUMN type TEXT NOT NULL DEFAULT 'objective';"
```

### WebSocket Connection Failed

```
Error: WebSocket connection closed unexpectedly
```

**Check**:
1. Go server is running
2. No firewall blocking :8080
3. Browser console for errors
4. Server logs for connection issues

### Evaluation Fails with "API key invalid"

**Steps**:
1. Go to `/settings`
2. Re-enter API keys
3. Click "Test API Key" buttons
4. Check Python service logs for detailed error

---

## 📚 Additional Resources

- **Main Documentation**: [README.md](README.md) - Complete feature list and architecture
- **Developer Reference**: [CLAUDE.md](CLAUDE.md) - Development patterns, database schema, testing
- **API Endpoints**: See README.md for complete endpoint reference

For issues or questions, please refer to the troubleshooting section above or check the main documentation.
