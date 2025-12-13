"""GPT judge implementation using OpenAI API."""

import json
import time
from openai import OpenAI
from typing import Dict


class GPTJudge:
    """Judge using GPT-5.2 with extended thinking."""

    def __init__(self, api_key: str, config: Dict):
        """
        Initialize GPT judge.

        Args:
            api_key: OpenAI API key
            config: Model configuration from config.py
        """
        self.client = OpenAI(api_key=api_key)
        self.config = config
        self.name = "gpt_5.2"

    async def evaluate(self, judge_prompt: str, max_retries: int = 3) -> Dict:
        """
        Evaluate using GPT with extended thinking.

        Args:
            judge_prompt: Formatted prompt for the judge
            max_retries: Maximum number of retry attempts

        Returns:
            Dict with score, confidence, reasoning, and cost
        """
        for attempt in range(max_retries):
            try:
                # Call GPT (Note: Adjust parameters based on actual API)
                response = self.client.chat.completions.create(
                    model=self.config["model"],
                    max_tokens=self.config["max_tokens"],
                    messages=[
                        {
                            "role": "system",
                            "content": "You are an expert evaluator. Respond only with JSON."
                        },
                        {
                            "role": "user",
                            "content": judge_prompt
                        }
                    ],
                    response_format={"type": "json_object"}  # Force JSON response
                )

                # Parse response
                response_text = response.choices[0].message.content.strip()
                result = json.loads(response_text)

                # Calculate cost
                input_tokens = response.usage.prompt_tokens
                output_tokens = response.usage.completion_tokens
                cost = (input_tokens / 1000 * 0.010) + (output_tokens / 1000 * 0.030)

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
                    time.sleep(2 ** attempt)
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
