"""Base evaluator class."""

from abc import ABC, abstractmethod
from typing import Dict, List
import os


class BaseEvaluator(ABC):
    """Base class for all evaluators."""

    def __init__(self):
        """Initialize evaluator."""
        self.prompt_template = self._load_prompt_template()

    @abstractmethod
    def _get_prompt_template_path(self) -> str:
        """Return path to the prompt template file."""
        pass

    def _load_prompt_template(self) -> str:
        """Load the prompt template from file."""
        template_path = self._get_prompt_template_path()
        with open(template_path, 'r', encoding='utf-8') as f:
            return f.read()

    def format_judge_prompt(self, prompt: str, response: str, solution: str = None) -> str:
        """
        Format the judge prompt with the given parameters.

        Args:
            prompt: The original prompt given to the model
            response: The model's response to evaluate
            solution: Expected solution (optional, used for objective evaluation)

        Returns:
            Formatted prompt string for the judge
        """
        return self.prompt_template.format(
            prompt=prompt,
            response=response,
            solution=solution or "N/A"
        )

    @abstractmethod
    async def evaluate(
        self,
        prompt: str,
        response: str,
        solution: str,
        judges: List[Dict],
        api_keys: Dict[str, str]
    ) -> List[Dict]:
        """
        Evaluate a response using multiple judges.

        Args:
            prompt: The original prompt
            response: Model's response to evaluate
            solution: Expected solution (may be None for creative tasks)
            judges: List of judge configurations
            api_keys: Dictionary of API keys by provider

        Returns:
            List of judge results with score, confidence, reasoning
        """
        pass
