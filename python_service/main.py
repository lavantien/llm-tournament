"""FastAPI server for LLM evaluation service."""

from fastapi import FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel, Field
from typing import List, Dict, Optional
import logging

from config import config
from evaluators import ObjectiveEvaluator, CreativeEvaluator

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s"
)
logger = logging.getLogger(__name__)

# Initialize FastAPI app
app = FastAPI(
    title="LLM Evaluation Service",
    description="Multi-judge LLM evaluation using LiteLLM",
    version="1.0.0"
)

# Add CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],  # Adjust for production
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Initialize evaluators
objective_evaluator = ObjectiveEvaluator()
creative_evaluator = CreativeEvaluator()


# Request/Response models
class EvaluationRequest(BaseModel):
    """Request model for evaluation endpoint."""
    prompt: str = Field(..., description="The original prompt")
    response: str = Field(..., description="Model's response to evaluate")
    solution: Optional[str] = Field(None, description="Expected solution (for objective)")
    type: str = Field("objective", description="Evaluation type: 'objective' or 'creative'")
    judges: List[str] = Field(
        default=["claude_opus_4.5", "gpt_5.2", "gemini_3_pro"],
        description="List of judges to use"
    )
    api_keys: Dict[str, str] = Field(..., description="API keys by provider")


class JudgeResult(BaseModel):
    """Result from a single judge."""
    judge: str
    score: int
    confidence: float
    reasoning: str
    cost_usd: float
    error: Optional[str] = None


class EvaluationResponse(BaseModel):
    """Response model for evaluation endpoint."""
    results: List[JudgeResult]
    total_cost_usd: float
    consensus_score: int
    avg_confidence: float


class CostEstimateRequest(BaseModel):
    """Request model for cost estimation."""
    prompt: str
    response: str
    solution: Optional[str] = None
    type: str = "objective"
    judges: List[str] = ["claude_opus_4.5", "gpt_5.2", "gemini_3_pro"]


class CostEstimateResponse(BaseModel):
    """Response model for cost estimation."""
    estimated_cost_usd: float
    breakdown: Dict[str, float]


@app.get("/health")
async def health_check():
    """Health check endpoint."""
    return {"status": "healthy", "service": "llm-evaluation"}


@app.post("/evaluate", response_model=EvaluationResponse)
async def evaluate(request: EvaluationRequest):
    """
    Evaluate a response using multiple judges.

    Args:
        request: Evaluation request with prompt, response, solution, etc.

    Returns:
        Evaluation response with judge results and consensus
    """
    try:
        logger.info(f"Evaluating {request.type} prompt with {len(request.judges)} judges")

        # Select evaluator based on type
        evaluator = objective_evaluator if request.type == "objective" else creative_evaluator

        # Run evaluation
        results = await evaluator.evaluate(
            prompt=request.prompt,
            response=request.response,
            solution=request.solution or "",
            judges=request.judges,
            api_keys=request.api_keys
        )

        # Calculate metrics
        total_cost = sum(r.get("cost_usd", 0.0) for r in results)

        # Calculate weighted consensus score
        valid_results = [r for r in results if r.get("confidence", 0) > 0]
        if valid_results:
            weighted_sum = sum(r["score"] * r["confidence"] for r in valid_results)
            total_weight = sum(r["confidence"] for r in valid_results)
            consensus_score = int(round(weighted_sum / total_weight)) if total_weight > 0 else 0
            avg_confidence = total_weight / len(valid_results)
        else:
            consensus_score = 0
            avg_confidence = 0.0

        logger.info(f"Evaluation complete. Consensus: {consensus_score}, Cost: ${total_cost:.4f}")

        return EvaluationResponse(
            results=[JudgeResult(**r) for r in results],
            total_cost_usd=round(total_cost, 4),
            consensus_score=consensus_score,
            avg_confidence=round(avg_confidence, 3)
        )

    except Exception as e:
        logger.error(f"Evaluation failed: {str(e)}")
        raise HTTPException(status_code=500, detail=str(e))


@app.post("/estimate_cost", response_model=CostEstimateResponse)
async def estimate_cost(request: CostEstimateRequest):
    """
    Estimate the cost of an evaluation without running it.

    Args:
        request: Cost estimation request

    Returns:
        Estimated cost breakdown
    """
    try:
        # Rough token estimation (4 characters ≈ 1 token)
        prompt_tokens = len(request.prompt) // 4
        response_tokens = len(request.response) // 4
        solution_tokens = len(request.solution or "") // 4
        template_tokens = 200  # Approximate template overhead

        input_tokens = prompt_tokens + response_tokens + solution_tokens + template_tokens
        output_tokens = 200  # Approximate judge response length

        breakdown = {}
        total_cost = 0.0

        for judge_name in request.judges:
            if judge_name in config.COSTS:
                costs = config.COSTS[judge_name]
                judge_cost = (input_tokens / 1000 * costs["input"]) + \
                            (output_tokens / 1000 * costs["output"])
                breakdown[judge_name] = round(judge_cost, 4)
                total_cost += judge_cost

        return CostEstimateResponse(
            estimated_cost_usd=round(total_cost, 4),
            breakdown=breakdown
        )

    except Exception as e:
        logger.error(f"Cost estimation failed: {str(e)}")
        raise HTTPException(status_code=500, detail=str(e))


@app.post("/evaluate_batch")
async def evaluate_batch(requests: List[EvaluationRequest]):
    """
    Batch evaluation endpoint.

    Args:
        requests: List of evaluation requests

    Returns:
        List of evaluation responses
    """
    results = []
    for req in requests:
        try:
            result = await evaluate(req)
            results.append(result)
        except Exception as e:
            logger.error(f"Batch evaluation item failed: {str(e)}")
            results.append({
                "error": str(e),
                "results": [],
                "total_cost_usd": 0.0,
                "consensus_score": 0,
                "avg_confidence": 0.0
            })

    return results


if __name__ == "__main__":
    import uvicorn
    logger.info(f"Starting LLM Evaluation Service on {config.HOST}:{config.PORT}")
    uvicorn.run(
        "main:app",
        host=config.HOST,
        port=config.PORT,
        reload=True,
        log_level="info"
    )
