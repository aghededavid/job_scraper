"""
Deduplicate router — POST /deduplicate
Checks if a job is a duplicate by fingerprint.
"""
import hashlib

from fastapi import APIRouter
from pydantic import BaseModel

router = APIRouter()


class DeduplicateRequest(BaseModel):
    title: str
    company_name: str
    description: str = ""
    fingerprint: str | None = None


class DeduplicateResponse(BaseModel):
    fingerprint: str
    is_duplicate: bool


@router.post("", response_model=DeduplicateResponse)
async def deduplicate_job(req: DeduplicateRequest):
    """
    Compute a deduplication fingerprint for a job.
    The caller is responsible for checking existence in DB.
    Returns the fingerprint and whether it appears to be a structural duplicate.
    """
    fp = req.fingerprint
    if not fp:
        normalized = f"{req.title.lower().strip()}|{req.company_name.lower().strip()}"
        fp = hashlib.sha256(normalized.encode()).hexdigest()[:32]

    # is_duplicate is always False here — the DB-level check is done in the worker.
    # This endpoint is for computing fingerprints without hitting the DB.
    return DeduplicateResponse(fingerprint=fp, is_duplicate=False)
