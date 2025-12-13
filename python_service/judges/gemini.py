"""Gemini judge implementation using Google Generative AI."""

import json
import time
import google.generativeai as genai
from typing import Dict


class GeminiJudge:
    """Judge using Gemini 3 Pro with extended thinking."""

    def __init__(self, api_key: str, config: Dict):
        """
        Initialize Gemini judge.

        Args:
            api_key: Google API key
            config: Model configuration from config.py
        """
        genai.configure(api_key=api_key)
        self.model = genai.GenerativeModel(config["model"])
        self.config = config
        self.name = "gemini_3_pro"

    async def evaluate(self, judge_prompt: str, max_retries: int = 3) -> Dict:
        """
        Evaluate using Gemini with extended thinking.

        Args:
            judge_prompt: Formatted prompt for the judge
            max_retries: Maximum number of retry attempts

        Returns:
            Dict with score, confidence, reasoning, and cost
        """
        for attempt in range(max_retries):
            try:
                # Call Gemini
                response = self.model.generate_content(
                    judge_prompt,
                    generation_config=genai.types.GenerationConfig(
                        max_output_tokens=self.config["max_tokens"],
                        temperature=0.1  # Low temperature for consistent evaluation
                    )
                )

                # Parse response
                response_text = response.text.strip()
                result = json.loads(response_text)

                # Estimate cost (Google's API doesn't always return token counts)
                # Using rough estimates based on character count
                input_tokens = len(judge_prompt) // 4
                output_tokens = len(response_text) // 4
                cost = (input_tokens / 1000 * 0.0125) + (output_tokens / 1000 * 0.050)

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
