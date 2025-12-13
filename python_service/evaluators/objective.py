"""Objective evaluator for prompts with clear expected solutions."""

import asyncio
import os
from typing import Dict, List
from .base import BaseEvaluator
from judges import ClaudeJudge, GPTJudge, GeminiJudge
from config import config


class ObjectiveEvaluator(BaseEvaluator):
    """Evaluator for objective prompts with semantic matching."""

    def _get_prompt_template_path(self) -> str:
        """Return path to objective judge prompt template."""
        return os.path.join(
            os.path.dirname(__file__),
            "..",
            "prompts",
            "objective_judge.txt"
        )

    async def evaluate(
        self,
        prompt: str,
        response: str,
        solution: str,
        judges: List[str],
        api_keys: Dict[str, str]
    ) -> List[Dict]:
        """
        Evaluate response using multiple judges in parallel.

        Args:
            prompt: The original prompt
            response: Model's response to evaluate
            solution: Expected solution
            judges: List of judge names to use
            api_keys: Dictionary of API keys

        Returns:
            List of judge results
        """
        # Format the judge prompt
        judge_prompt = self.format_judge_prompt(prompt, response, solution)

        # Initialize judges
        judge_instances = []
        for judge_name in judges:
            if judge_name not in config.JUDGE_MODELS:
                continue

            judge_config = config.JUDGE_MODELS[judge_name]
            provider = judge_config["provider"]
            api_key = api_keys.get(f"api_key_{provider}", "")

            if not api_key:
                continue

            # Create judge instance
            if provider == "anthropic":
                judge_instances.append(ClaudeJudge(api_key, judge_config))
            elif provider == "openai":
                judge_instances.append(GPTJudge(api_key, judge_config))
            elif provider == "google":
                judge_instances.append(GeminiJudge(api_key, judge_config))

        # Evaluate in parallel
        tasks = [judge.evaluate(judge_prompt) for judge in judge_instances]
        results = await asyncio.gather(*tasks)

        return results
