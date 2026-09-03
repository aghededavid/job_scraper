"""
Summarize router — POST /summarize
Generates a 2-sentence AI summary of a job description.
"""
from fastapi import APIRouter, HTTPException
from pydantic import BaseModel

router = APIRouter()


class SummarizeRequest(BaseModel):
    title: str = ""
    company_name: str = ""
    description: str = ""


class SummarizeResponse(BaseModel):
    summary: str | None
    available: bool


@router.post("", response_model=SummarizeResponse)
async def summarize_job(req: SummarizeRequest):
    """Generate a 2-sentence summary of a job description using the AI model."""
    from ai_service.main import pipeline

    if pipeline is None:
        raise HTTPException(status_code=503, detail="Pipeline not initialized")

    if not pipeline.openai_key:
        return SummarizeResponse(summary=None, available=False)

    job_data = {
        "title": req.title,
        "company_name": req.company_name,
        "description": req.description,
    }
    job_data = await pipeline._summarize(job_data)
    summary = job_data.get("description_summary")

    return SummarizeResponse(summary=summary, available=True)
