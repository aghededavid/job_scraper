"""
Remotive Adapter
Source: https://remotive.com/api/remote-jobs (public JSON API)
Terms:  Free public API. No auth required.
"""
from datetime import datetime

from scraper.adapters.base import BaseScraper
from scraper.models.job import StandardizedJob, FetchMethod, RemoteType


class RemotiveScraper(BaseScraper):
    source_name = "Remotive"
    source_slug = "remotive"
    adapter_class = "RemotiveScraper"
    base_url = "https://remotive.com/api/remote-jobs"
    fetch_method = FetchMethod.API
    schedule_cron = "*/10 * * * *"
    rate_limit_rpm = 30

    async def fetch_jobs(self) -> list[StandardizedJob]:
        client = await self.get_client()
        resp = await client.get(self.base_url, params={"limit": 200})
        resp.raise_for_status()
        data = resp.json()

        jobs = []
        for item in data.get("jobs", []):
            try:
                job = self._parse_job(item)
                if job:
                    jobs.append(job)
            except Exception as e:
                self.logger.warning(f"Failed to parse Remotive job {item.get('id')}: {e}")
        return jobs

    def _parse_job(self, item: dict):
        tags = item.get("tags", [])
        salary_min, salary_max, currency = self.parse_salary(item.get("salary", ""))

        posted_at = None
        pub_date = item.get("publication_date")
        if pub_date:
            try:
                posted_at = datetime.fromisoformat(pub_date.replace("Z", "+00:00"))
            except ValueError:
                pass

        job = StandardizedJob(
            source_slug=self.source_slug,
            source_url=item.get("url", ""),
            source_job_id=str(item.get("id", "")),
            title=item.get("title", ""),
            company_name=item.get("company_name", ""),
            company_logo_url=item.get("company_logo_url"),
            description=self.clean_html(item.get("description", "")),
            apply_url=item.get("url", ""),
            is_remote=True,
            remote_type=RemoteType.WORLDWIDE,
            location=item.get("candidate_required_location", "Worldwide"),
            category=item.get("job_type"),
            tags=tags,
            skills=tags,
            salary_min=salary_min,
            salary_max=salary_max,
            salary_currency=currency,
            salary_raw=item.get("salary"),
            posted_at=posted_at,
        )
        job.compute_fingerprint()
        return job
