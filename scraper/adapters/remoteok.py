"""
RemoteOK Adapter
Source: https://remoteok.com/remote-jobs.json (public JSON API)
Terms:  Public API, no auth required. Respect rate limits.
"""
from datetime import datetime
from typing import Optional

from scraper.adapters.base import BaseScraper
from scraper.models.job import StandardizedJob, FetchMethod, RemoteType, EmploymentType


class RemoteOKScraper(BaseScraper):
    source_name = "RemoteOK"
    source_slug = "remoteok"
    adapter_class = "RemoteOKScraper"
    base_url = "https://remoteok.com/remote-jobs.json"
    fetch_method = FetchMethod.API
    schedule_cron = "*/5 * * * *"
    rate_limit_rpm = 10  # Be conservative with RemoteOK

    async def fetch_jobs(self) -> list[StandardizedJob]:
        client = await self.get_client()
        resp = await client.get(self.base_url)
        resp.raise_for_status()
        data = resp.json()

        jobs = []
        # First item is metadata, skip it
        for item in data[1:]:
            try:
                job = self._parse_job(item)
                if job:
                    jobs.append(job)
            except Exception as e:
                self.logger.warning(f"Failed to parse job {item.get('id')}: {e}")
        return jobs

    def _parse_job(self, item: dict) -> Optional[StandardizedJob]:
        # RemoteOK provides tags as job category/skills
        tags = item.get("tags", [])

        # Extract salary if present
        salary_min = item.get("salary_min") or item.get("salary")
        salary_max = item.get("salary_max") or salary_min

        # Parse posted timestamp
        epoch = item.get("epoch") or item.get("date")
        posted_at = None
        if epoch:
            try:
                posted_at = datetime.utcfromtimestamp(int(epoch))
            except (ValueError, TypeError):
                posted_at = None

        company = item.get("company", "Unknown")
        title = item.get("position", "")
        if not title or not company:
            return None

        job = StandardizedJob(
            source_slug=self.source_slug,
            source_url=f"https://remoteok.com/remote-jobs/{item.get('id', '')}",
            source_job_id=str(item.get("id", "")),
            title=title,
            company_name=company,
            company_logo_url=item.get("company_logo"),
            description=self.clean_html(item.get("description", "")),
            apply_url=item.get("url") or f"https://remoteok.com/remote-jobs/{item.get('id','')}",
            is_remote=True,
            remote_type=RemoteType.WORLDWIDE,
            location=item.get("location", "Worldwide"),
            tags=tags,
            skills=tags,
            salary_min=int(salary_min) if salary_min else None,
            salary_max=int(salary_max) if salary_max else None,
            salary_currency="USD",
            employment_type=EmploymentType.FULL_TIME,
            posted_at=posted_at,
        )
        job.compute_fingerprint()
        return job
