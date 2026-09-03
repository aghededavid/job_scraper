"""
Pipeline Worker — consumes jobs from Redis queue and runs AI pipeline,
then stores results in PostgreSQL.
"""
import asyncio
import json
import logging
import os
from datetime import datetime
from typing import Optional

import asyncpg
import redis.asyncio as aioredis

from ai_service.pipeline import AIPipeline

logger = logging.getLogger(__name__)

QUEUE_PARSE = "q:parse"
QUEUE_NOTIFY = "q:notify"


class PipelineWorker:
    def __init__(self, pipeline: AIPipeline, redis_url: str):
        self.pipeline = pipeline
        self.redis_url = redis_url
        self._redis: Optional[aioredis.Redis] = None
        self._db: Optional[asyncpg.Pool] = None
        self.concurrency = int(os.getenv("WORKER_CONCURRENCY", "5"))

    async def _connect(self):
        self._redis = aioredis.from_url(self.redis_url, encoding="utf-8", decode_responses=True)
        db_url = os.getenv("DATABASE_URL", "").replace("postgresql://", "postgres://")
        self._db = await asyncpg.create_pool(db_url, min_size=2, max_size=10)
        logger.info("Worker connected to Redis and PostgreSQL")

    async def run(self):
        await self._connect()
        sem = asyncio.Semaphore(self.concurrency)

        async def process_one():
            async with sem:
                result = await self._redis.blpop(QUEUE_PARSE, timeout=5)
                if not result:
                    return
                _, data = result
                job_data = json.loads(data)
                await self._process_and_store(job_data)

        logger.info(f"Worker running with concurrency={self.concurrency}")
        while True:
            try:
                await process_one()
            except asyncio.CancelledError:
                break
            except Exception as e:
                logger.exception(f"Worker error: {e}")
                await asyncio.sleep(1)

    async def _process_and_store(self, job_data: dict):
        try:
            # Run AI pipeline
            processed = await self.pipeline.process(job_data)

            # Check deduplication
            fingerprint = processed.get("fingerprint")
            if fingerprint:
                existing = await self._db.fetchval(
                    "SELECT id FROM jobs WHERE fingerprint = $1", fingerprint
                )
                if existing:
                    logger.debug(f"Duplicate job skipped: {fingerprint}")
                    return

            # Upsert company
            company_id = await self._upsert_company(processed)

            # Insert job
            job_id = await self._insert_job(processed, company_id)

            # Store embedding
            if processed.get("embedding") and job_id:
                await self._store_embedding(job_id, processed["embedding"])

            # Enqueue notification for instant-freq users
            if job_id:
                await self._enqueue_notification(job_id, processed)

            logger.info(f"Stored job: {processed.get('title')} @ {processed.get('company_name')}")

        except Exception as e:
            logger.exception(f"Failed to process/store job: {e}")

    async def _upsert_company(self, job_data: dict) -> Optional[str]:
        company_name = job_data.get("company_name", "").strip()
        if not company_name:
            return None

        import re
        slug = re.sub(r"[^\w\s-]", "", company_name.lower()).strip()
        slug = re.sub(r"[\s_-]+", "-", slug)

        try:
            row = await self._db.fetchrow(
                """
                INSERT INTO companies (id, name, slug, website, logo_url)
                VALUES (uuid_generate_v4(), $1, $2, $3, $4)
                ON CONFLICT (slug) DO UPDATE SET
                    logo_url = COALESCE(EXCLUDED.logo_url, companies.logo_url),
                    updated_at = NOW()
                RETURNING id
                """,
                company_name,
                slug,
                job_data.get("company_website"),
                job_data.get("company_logo_url"),
            )
            return str(row["id"]) if row else None
        except Exception as e:
            logger.warning(f"Company upsert failed: {e}")
            return None

    async def _insert_job(self, job_data: dict, company_id: Optional[str]) -> Optional[str]:
        try:
            posted_at = None
            if job_data.get("posted_at"):
                try:
                    posted_at = datetime.fromisoformat(str(job_data["posted_at"]))
                except (ValueError, TypeError):
                    pass

            row = await self._db.fetchrow(
                """
                INSERT INTO jobs (
                    id, title, normalized_title, company_id, description, description_summary,
                    is_remote, remote_type, location, timezone_restrictions,
                    salary_min, salary_max, salary_currency, salary_estimated,
                    experience_level, employment_type,
                    skills, technologies, benefits, visa_sponsorship,
                    category, keywords, apply_url,
                    fingerprint, posted_at, freshness_score,
                    is_fake_suspected, fake_score, ai_processed
                ) VALUES (
                    uuid_generate_v4(), $1, $2, $3::uuid, $4, $5,
                    $6, $7, $8, $9,
                    $10, $11, $12, $13,
                    $14, $15,
                    $16, $17, $18, $19,
                    $20, $21, $22,
                    $23, $24, $25,
                    $26, $27, $28
                )
                ON CONFLICT (fingerprint) DO NOTHING
                RETURNING id
                """,
                job_data.get("title", ""),
                job_data.get("normalized_title"),
                company_id,
                job_data.get("description", ""),
                job_data.get("description_summary"),
                True,
                job_data.get("remote_type", "worldwide"),
                job_data.get("location"),
                job_data.get("timezone_restrictions", []),
                job_data.get("salary_min"),
                job_data.get("salary_max"),
                job_data.get("salary_currency", "USD"),
                False,
                job_data.get("experience_level", "any"),
                job_data.get("employment_type", "full_time"),
                job_data.get("skills", []),
                job_data.get("technologies", []),
                job_data.get("benefits", []),
                job_data.get("visa_sponsorship"),
                job_data.get("category"),
                job_data.get("tags", []),
                job_data.get("apply_url", ""),
                job_data.get("fingerprint"),
                posted_at,
                1.0,
                job_data.get("is_fake_suspected", False),
                job_data.get("fake_score"),
                job_data.get("ai_processed", False),
            )
            return str(row["id"]) if row else None
        except Exception as e:
            logger.exception(f"Job insert failed: {e}")
            return None

    async def _store_embedding(self, job_id: str, embedding: list[float]):
        try:
            vec_str = "[" + ",".join(str(x) for x in embedding) + "]"
            await self._db.execute(
                """
                INSERT INTO job_embeddings (id, job_id, embedding, model_version)
                VALUES (uuid_generate_v4(), $1::uuid, $2::vector, $3)
                ON CONFLICT (job_id) DO UPDATE SET embedding = EXCLUDED.embedding, updated_at = NOW()
                """,
                job_id,
                vec_str,
                os.getenv("OPENAI_EMBEDDING_MODEL", "text-embedding-3-small"),
            )
        except Exception as e:
            logger.warning(f"Embedding store failed: {e}")

    async def _enqueue_notification(self, job_id: str, job_data: dict):
        """Notify users with instant alerts if job matches their saved searches."""
        try:
            payload = json.dumps({
                "type": "new_job",
                "job_id": job_id,
                "title": job_data.get("title"),
                "company": job_data.get("company_name"),
                "category": job_data.get("category"),
                "skills": job_data.get("skills", []),
            })
            await self._redis.rpush(QUEUE_NOTIFY, payload)
        except Exception as e:
            logger.warning(f"Notification enqueue failed: {e}")
