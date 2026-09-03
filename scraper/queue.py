"""
Queue Publisher — publishes scraped jobs to Redis for processing.
"""
import asyncio
import json
import logging
from typing import Optional

import redis.asyncio as aioredis

from scraper.models.job import StandardizedJob

logger = logging.getLogger(__name__)

QUEUE_PARSE = "q:parse"
QUEUE_DEAD_LETTER = "q:dead-letter"


class QueuePublisher:
    def __init__(self, redis_url: str):
        self.redis_url = redis_url
        self._redis: Optional[aioredis.Redis] = None

    async def connect(self):
        self._redis = aioredis.from_url(
            self.redis_url,
            encoding="utf-8",
            decode_responses=True,
        )
        await self._redis.ping()
        logger.info("Connected to Redis queue")

    async def close(self):
        if self._redis:
            await self._redis.aclose()

    async def publish(self, job: StandardizedJob) -> bool:
        """Publish a single job to the parse queue."""
        if not self._redis:
            raise RuntimeError("Redis not connected")
        try:
            payload = json.dumps(job.to_db_dict(), default=str)
            await self._redis.rpush(QUEUE_PARSE, payload)
            return True
        except Exception as e:
            logger.error(f"Failed to publish job: {e}")
            await self._dead_letter(job, str(e))
            return False

    async def publish_batch(self, jobs: list[StandardizedJob]) -> int:
        """Publish multiple jobs. Returns number successfully published."""
        if not self._redis:
            raise RuntimeError("Redis not connected")

        pipeline = self._redis.pipeline()
        for job in jobs:
            try:
                payload = json.dumps(job.to_db_dict(), default=str)
                pipeline.rpush(QUEUE_PARSE, payload)
            except Exception as e:
                logger.warning(f"Failed to serialize job: {e}")

        try:
            results = await pipeline.execute()
            return sum(1 for r in results if r)
        except Exception as e:
            logger.error(f"Pipeline error: {e}")
            return 0

    async def _dead_letter(self, job: StandardizedJob, error: str):
        """Send failed job to dead letter queue."""
        if self._redis:
            payload = json.dumps({
                "job": job.to_db_dict(),
                "error": error,
            }, default=str)
            await self._redis.rpush(QUEUE_DEAD_LETTER, payload)


class QueueConsumer:
    """For the AI service to consume from the parse queue."""

    def __init__(self, redis_url: str, queue: str = QUEUE_PARSE):
        self.redis_url = redis_url
        self.queue = queue
        self._redis: Optional[aioredis.Redis] = None

    async def connect(self):
        self._redis = aioredis.from_url(
            self.redis_url,
            encoding="utf-8",
            decode_responses=True,
        )
        logger.info(f"Consumer connected to Redis queue: {self.queue}")

    async def consume(self, timeout: int = 5) -> Optional[dict]:
        """Blocking pop from queue. Returns parsed job dict or None."""
        if not self._redis:
            raise RuntimeError("Redis not connected")
        result = await self._redis.blpop(self.queue, timeout=timeout)
        if result:
            _, data = result
            return json.loads(data)
        return None

    async def consume_loop(self, handler, concurrency: int = 5):
        """Continuously consume and process with given handler."""
        sem = asyncio.Semaphore(concurrency)

        async def process_one():
            async with sem:
                job_data = await self.consume()
                if job_data:
                    await handler(job_data)

        while True:
            try:
                await process_one()
            except Exception as e:
                logger.exception(f"Consumer error: {e}")
                await asyncio.sleep(1)
