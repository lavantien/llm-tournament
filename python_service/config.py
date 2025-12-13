"""Configuration management for LiteLLM evaluation service."""

import os
from dotenv import load_dotenv

load_dotenv()

class Config:
    """Service configuration."""

    # API Keys (loaded from environment or passed from Go service)
    ANTHROPIC_API_KEY = os.getenv("ANTHROPIC_API_KEY", "")
    OPENAI_API_KEY = os.getenv("OPENAI_API_KEY", "")
    GOOGLE_API_KEY = os.getenv("GOOGLE_API_KEY", "")

    # Server settings
    HOST = os.getenv("HOST", "0.0.0.0")
    PORT = int(os.getenv("PORT", "8001"))

    # Judge models with thinking modes
    JUDGE_MODELS = {
        "claude_opus_4.5": {
            "model": "claude-opus-4.5-20250929",
            "provider": "anthropic",
            "thinking": "high",
            "max_tokens": 4096,
        },
        "gpt_5.2": {
            "model": "gpt-5.2",
            "provider": "openai",
            "thinking": "high",
            "max_tokens": 4096,
        },
        "gemini_3_pro": {
            "model": "gemini-3-pro",
            "provider": "google",
            "thinking": "high",
            "max_tokens": 4096,
        },
    }

    # Cost estimation (per 1K tokens)
    COSTS = {
        "claude_opus_4.5": {
            "input": 0.015,
            "output": 0.075,
        },
        "gpt_5.2": {
            "input": 0.010,
            "output": 0.030,
        },
        "gemini_3_pro": {
            "input": 0.0125,
            "output": 0.050,
        },
    }

    # Timeout settings (seconds)
    JUDGE_TIMEOUT = 120  # Thinking models can be slow

    # Retry settings
    MAX_RETRIES = 3
    RETRY_DELAY = 2  # seconds

config = Config()
