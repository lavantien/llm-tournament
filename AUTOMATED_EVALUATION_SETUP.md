# Automated Evaluation Setup

The LLM Tournament Arena supports **optional automated evaluation** using a Python FastAPI “judge service” (`python_service/`). The Go server stays the source of truth (SQLite + SSR templates + WebSockets); the Python service is only used for scoring.

## What You Get

- Multi-judge consensus scoring (default judges: Claude Opus 4.5, GPT-5.2, Gemini 3 Pro)
- Async evaluation jobs with progress + cost updates in the UI
- Encrypted API key storage (AES-256-GCM) via `ENCRYPTION_KEY`

## Prerequisites

- Go 1.24+ with a working CGO toolchain (SQLite requires CGO)
- Python 3.8+
- An API key for at least one provider (Anthropic / OpenAI / Google)

## Quick Start (Local)

### 1) Install Python Dependencies

```bash
cd python_service
pip install -r requirements.txt
```

### 2) Set `ENCRYPTION_KEY` (64 hex chars / 32 bytes)

macOS/Linux:

```bash
export ENCRYPTION_KEY=$(openssl rand -hex 32)
```

PowerShell:

```powershell
$env:ENCRYPTION_KEY = (python -c "import secrets; print(secrets.token_hex(32))")
```

### 3) Start the Python Judge Service (Terminal 1)

```bash
cd python_service
python main.py
```

Health check:

```bash
curl http://localhost:8001/health
```

### 4) Start the Go Server (Terminal 2)

```bash
CGO_ENABLED=1 go run .
```

Open the app:

- http://localhost:8080/settings (configure and test API keys)
- http://localhost:8080/results (run evaluations from the UI)

## How Evaluation Works

- **Evaluate All:** `POST /evaluate/all`
- **Evaluate a model:** `POST /evaluate/model?id={model_id}`
- **Evaluate a prompt:** `POST /evaluate/prompt?id={prompt_id}`
- **Progress:** `GET /evaluation/progress?id={job_id}` (also pushed via WebSocket)
- **Cancel:** `POST /evaluation/cancel?id={job_id}`

Prompts can be `objective` (semantic matching) or `creative` (quality assessment). The UI shows a cost estimate before running jobs.

## Environment Variables

Go server:

- `CGO_ENABLED=1` (required for SQLite)
- `ENCRYPTION_KEY` (64 hex chars / 32 bytes; required for encrypted API keys / automated evaluation)

Python judge service:

- `HOST` (default `0.0.0.0`)
- `PORT` (default `8001`)

## Security Notes

- Never commit `ENCRYPTION_KEY` or provider API keys.
- Use HTTPS in production and restrict CORS (the judge service is permissive by default for local dev).
- Cost varies by provider/model; the UI tracks spend and supports alert thresholds.

## Troubleshooting

### `ENCRYPTION_KEY` errors

If you see an error like “`ENCRYPTION_KEY environment variable not set`” or “must be 64 hex characters”, re-generate it and restart the Go server.

### Python service unavailable / evaluations fail immediately

- Confirm the service is running and healthy: `curl http://localhost:8001/health`
- Check Python logs for provider errors (invalid keys, rate limits, etc.)
- Confirm your keys in http://localhost:8080/settings and use the “Test API Key” buttons

### CGO / SQLite build failures

Install a working C compiler toolchain (CGO requires it to build SQLite support).

## Additional Resources

- [README.md](README.md) (UI tour, architecture, endpoints)
- [CLAUDE.md](CLAUDE.md) (developer notes)
