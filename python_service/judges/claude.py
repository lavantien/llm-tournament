"""Claude judge implementation using Anthropic API."""

import json
import time
from anthropic import Anthropic
from typing import Dict


class ClaudeJudge:
    """Judge using Claude Opus 4.5 with extended thinking."""

    def __init__(self, api_key: str, config: Dict):
        """
        Initialize Claude judge.

        Args:
            api_key: Anthropic API key
            config: Model configuration from config.py
        """
        self.client = Anthropic(api_key=api_key)
        self.config = config
        self.name = "claude_opus_4.5"

    async def evaluate(self, judge_prompt: str, max_retries: int = 3) -> Dict:
        """
        Evaluate using Claude with extended thinking.

        Args:
            judge_prompt: Formatted prompt for the judge
            max_retries: Maximum number of retry attempts

        Returns:
            Dict with score, confidence, reasoning, and cost
        """
        for attempt in range(max_retries):
            try:
                # Call Claude with extended thinking mode
                response = self.client.messages.create(
                    model=self.config["model"],
                    max_tokens=self.config["max_tokens"],
                    thinking={
                        "type": "enabled",
                        "budget_tokens": 10000  # High thinking budget
                    },
                    messages=[
                        {
                            "role": "user",
                            "content": judge_prompt
                        }
                    ]
                )

                # Extract the response text (skip thinking blocks)
                response_text = ""
                for block in response.content:
                    if block.type == "text":
                        response_text += block.text

                # Parse JSON response
                result = json.loads(response_text.strip())

                # Calculate cost
                input_tokens = response.usage.input_tokens
                output_tokens = response.usage.output_tokens
                cost = (input_tokens / 1000 * 0.015) + (output_tokens / 1000 * 0.075)

                return {
                    "judge": self.name,
                    "score": int(result["score"]),
                    "confidence": float(result["confidence"]),
                    "reasoning": str(result["reasoning"]),
                    "cost_usd": round(cost, 4),
                    "tokens": {
                        "input": input_tokens,
                        "output": output_tokens
                    }
                }

            except json.JSONDecodeError as e:
                if attempt < max_retries - 1:
                    time.sleep(2 ** attempt)  # Exponential backoff
                    continue
                return {
                    "judge": self.name,
                    "score": 0,
                    "confidence": 0.0,
                    "reasoning": f"Failed to parse JSON response: {str(e)}",
                    "cost_usd": 0.0,
                    "error": str(e)
                }

            except Exception as e:
                if attempt < max_retries - 1:
                    time.sleep(2 ** attempt)
                    continue
                return {
                    "judge": self.name,
                    "score": 0,
                    "confidence": 0.0,
                    "reasoning": f"Evaluation failed: {str(e)}",
                    "cost_usd": 0.0,
                    "error": str(e)
                }

        return {
            "judge": self.name,
            "score": 0,
            "confidence": 0.0,
            "reasoning": "Max retries exceeded",
            "cost_usd": 0.0,
            "error": "Max retries exceeded"
        }
