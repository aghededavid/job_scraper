"""
RemoteScope AI Service — FastAPI app
Processes jobs through the full AI pipeline:
1. Skill extraction
2. Title normalization
3. Category classification
4. Embedding generation
5. Deduplication
6. Summarization
7. Fake job detection
"""
import asyncio
import logging
import os
import sys

from contextlib import asynccontextmanager
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from ai_service.pipeline import AIPipeline
from ai_service.worker import PipelineWorker
from ai_service.routers import classify, embed, deduplicate, extract, summarize

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(name)s] %(levelname)s: %(message)s",
    handlers=[logging.StreamHandler(sys.stdout)],
)
logger = logging.getLogger("ai_service")

pipeline: AIPipeline | None = None
worker: PipelineWorker | None = None


@asynccontextmanager
async def lifespan(app: FastAPI):
    global pipeline, worker

    logger.info("Starting AI Service...")

    pipeline = AIPipeline()
    await pipeline.initialize()

    redis_url = os.getenv("REDIS_URL", "redis://localhost:6379/0")
    worker = PipelineWorker(pipeline=pipeline, redis_url=redis_url)

    # Start background worker
    worker_task = asyncio.create_task(worker.run())

    logger.info("✅ AI Service ready")
    yield

    # Shutdown
    worker_task.cancel()
    await pipeline.close()
    logger.info("AI Service shut down")


app = FastAPI(
    title="RemoteScope AI Service",
    description="AI processing pipeline for job classification, embeddings, and deduplication",
    version="1.0.0",
    lifespan=lifespan,
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_methods=["*"],
    allow_headers=["*"],
)

# ── Routers ──────────────────────────────────────────────────
app.include_router(embed.router, prefix="/embed", tags=["Embeddings"])
app.include_router(classify.router, prefix="/classify", tags=["Classification"])
app.include_router(deduplicate.router, prefix="/deduplicate", tags=["Deduplication"])
app.include_router(extract.router, prefix="/extract", tags=["Extraction"])
app.include_router(summarize.router, prefix="/summarize", tags=["Summarization"])


@app.get("/health")
async def health():
    return {"status": "ok", "service": "remotescope-ai"}


@app.post("/process")
async def process_job(job_data: dict):
    """Process a single job through the full AI pipeline."""
    if pipeline is None:
        return {"error": "pipeline not initialized"}
    result = await pipeline.process(job_data)
    return result
