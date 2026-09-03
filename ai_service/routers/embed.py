"""
Embed router — POST /embed
Generates a vector embedding for arbitrary text.
"""
from typing import Optional
from fastapi import APIRouter, HTTPException
from pydantic import BaseModel

router = APIRouter()


class EmbedRequest(BaseModel):
    text: str
    model: Optional[str] = None


class EmbedResponse(BaseModel):
    embedding: list[float]
    model: str
    dimensions: int


@router.post("", response_model=EmbedResponse)
async def embed_text(req: EmbedRequest):
    """Generate an embedding vector for the provided text."""
    from ai_service.main import pipeline

    if pipeline is None:
        raise HTTPException(status_code=503, detail="Pipeline not initialized")

    embedding = await pipeline.embed_text(req.text)
    if embedding is None:
        raise HTTPException(
            status_code=503,
            detail="Embedding unavailable — check OPENAI_API_KEY",
        )

    model = pipeline.embedding_model
    return EmbedResponse(embedding=embedding, model=model, dimensions=len(embedding))
