"""
Extract router — POST /extract
Extracts skills and technologies from job text.
"""
from fastapi import APIRouter, HTTPException
from pydantic import BaseModel

router = APIRouter()


class ExtractRequest(BaseModel):
    title: str = ""
    description: str = ""
    tags: list[str] = []
    skills: list[str] = []
    technologies: list[str] = []


class ExtractResponse(BaseModel):
    skills: list[str]
    technologies: list[str]


@router.post("", response_model=ExtractResponse)
async def extract_skills(req: ExtractRequest):
    """Extract skills and technologies from job title, description, and tags."""
    from ai_service.main import pipeline

    if pipeline is None:
        raise HTTPException(status_code=503, detail="Pipeline not initialized")

    job_data = {
        "title": req.title,
        "description": req.description,
        "tags": req.tags,
        "skills": req.skills,
        "technologies": req.technologies,
    }
    job_data = pipeline._extract_skills(job_data)

    return ExtractResponse(
        skills=job_data.get("skills", []),
        technologies=job_data.get("technologies", []),
    )
