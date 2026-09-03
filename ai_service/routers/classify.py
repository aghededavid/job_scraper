"""
Classify router — POST /classify
Classifies a job into a category.
"""
from fastapi import APIRouter, HTTPException
from pydantic import BaseModel

router = APIRouter()


class ClassifyRequest(BaseModel):
    title: str
    description: str = ""
    tags: list[str] = []
    category: str = ""


class ClassifyResponse(BaseModel):
    category: str
    normalized_title: str


@router.post("", response_model=ClassifyResponse)
async def classify_job(req: ClassifyRequest):
    """Classify a job into a category and normalize its title."""
    from ai_service.main import pipeline

    if pipeline is None:
        raise HTTPException(status_code=503, detail="Pipeline not initialized")

    job_data = {
        "title": req.title,
        "description": req.description,
        "tags": req.tags,
        "category": req.category,
    }
    job_data = pipeline._classify_category(job_data)
    job_data = pipeline._normalize_title(job_data)

    return ClassifyResponse(
        category=job_data.get("category", "Software Engineering"),
        normalized_title=job_data.get("normalized_title", req.title),
    )
