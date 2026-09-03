"""
Scraper Scheduler & Orchestrator
Runs all adapters on their configured schedules and publishes results to Redis queue.
"""
import asyncio
import logging
import os
import sys
from typing import Type

from apscheduler.schedulers.asyncio import AsyncIOScheduler
from apscheduler.triggers.cron import CronTrigger

from scraper.adapters.base import BaseScraper
from scraper.adapters.remoteok import RemoteOKScraper
from scraper.adapters.remotive import RemotiveScraper
from scraper.adapters.jobicy import JobicyScraper
from scraper.adapters.weworkremotely import WWRScraper
from scraper.adapters.greenhouse import GreenhouseScraper
from scraper.queue import QueuePublisher

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(name)s] %(levelname)s: %(message)s",
    handlers=[logging.StreamHandler(sys.stdout)],
)
logger = logging.getLogger("scheduler")

# ── Registry of all adapters ─────────────────────────────────
ADAPTERS: list[Type[BaseScraper]] = [
    RemoteOKScraper,
    RemotiveScraper,
    JobicyScraper,
    WWRScraper,
    GreenhouseScraper,
]


class ScraperScheduler:
    def __init__(self, publisher: QueuePublisher):
        self.publisher = publisher
        self.scheduler = AsyncIOScheduler(timezone="UTC")

    def register_adapters(self):
        """Register each adapter on its configured cron schedule."""
        for adapter_cls in ADAPTERS:
            adapter = adapter_cls()
            job_id = f"scrape_{adapter.source_slug}"

            self.scheduler.add_job(
                func=self._run_adapter,
                args=[adapter_cls],
                trigger=CronTrigger.from_crontab(adapter.schedule_cron),
                id=job_id,
                name=f"Scrape {adapter.source_name}",
                max_instances=1,
                coalesce=True,
                misfire_grace_time=60,
                replace_existing=True,
            )
            logger.info(f"Registered: {adapter.source_name} [{adapter.schedule_cron}]")

    async def _run_adapter(self, adapter_cls: Type[BaseScraper]):
        """Instantiate, run, and publish results for one adapter."""
        adapter = adapter_cls()
        logger.info(f"[{adapter.source_name}] Starting scrape run...")

        try:
            jobs = await adapter.run()
            if jobs:
                await self.publisher.publish_batch(jobs)
                logger.info(f"[{adapter.source_name}] Published {len(jobs)} jobs to queue")
            else:
                logger.info(f"[{adapter.source_name}] No jobs found")
        except Exception as e:
            logger.exception(f"[{adapter.source_name}] Scrape failed: {e}")

    async def run_all_now(self):
        """Trigger all adapters immediately (for testing/seeding)."""
        tasks = [self._run_adapter(cls) for cls in ADAPTERS]
        await asyncio.gather(*tasks, return_exceptions=True)

    def start(self):
        self.scheduler.start()
        logger.info(f"Scheduler started with {len(ADAPTERS)} adapters")

    def stop(self):
        self.scheduler.shutdown()


async def main():
    import argparse
    parser = argparse.ArgumentParser(description="RemoteScope Scraper")
    parser.add_argument("--run-now", action="store_true", help="Run all scrapers immediately and exit")
    parser.add_argument("--source", type=str, help="Run specific source by slug")
    args = parser.parse_args()

    publisher = QueuePublisher(redis_url=os.getenv("REDIS_URL", "redis://localhost:6379/0"))
    await publisher.connect()

    scheduler = ScraperScheduler(publisher)

    if args.run_now:
        if args.source:
            # Find adapter by slug
            for cls in ADAPTERS:
                if cls.source_slug == args.source:
                    await scheduler._run_adapter(cls)
                    break
            else:
                logger.error(f"Unknown source: {args.source}")
        else:
            await scheduler.run_all_now()
        await publisher.close()
        return

    # Normal scheduled mode
    scheduler.register_adapters()
    scheduler.start()

    # Keep running
    try:
        while True:
            await asyncio.sleep(3600)
    except (KeyboardInterrupt, SystemExit):
        scheduler.stop()
        await publisher.close()
        logger.info("Scheduler stopped")


if __name__ == "__main__":
    asyncio.run(main())
