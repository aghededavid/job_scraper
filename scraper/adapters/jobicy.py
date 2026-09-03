"""
Jobicy Adapter
Source: https://jobicy.com/api/v2/remote-jobs (public JSON API)
Terms:  Free public API.
"""
from datetime import datetime

from scraper.adapters.base import BaseScraper
from scraper.models.job import StandardizedJob, FetchMethod, RemoteType


class JobicyScraper(BaseScraper):
    source_name = "Jobicy"
    source_slug = "jobicy"
    adapter_class = "JobicyScraper"
    base_url = "https://jobicy.com/api/v2/remote-jobs"
    fetch_method = FetchMethod.API
    schedule_cron = "*/10 * * * *"
    rate_limit_rpm = 30

    async def fetch_jobs(self) -> list[StandardizedJob]:
        client = await self.get_client()
        resp = await client.get(self.base_url, params={"count": 50, "geo": "worldwide"})
        resp.raise_for_status()
        data = resp.json()

        jobs = []
        for item in data.get("jobs", []):
            try:
                job = self._parse_job(item)
                if job:
                    jobs.append(job)
            except Exception as e:
                self.logger.warning(f"Failed to parse Jobicy job: {e}")
        return jobs

    def _parse_job(self, item: dict):
        tags = item.get("jobIndustry", [])
        if isinstance(tags, str):
            tags = [tags]

        salary_min, salary_max, currency = self.parse_salary(item.get("annualSalaryMin", "") or "")
        if not salary_min:
            salary_min, salary_max, currency = self.parse_salary(item.get("annualSalaryMax", "") or "")

        posted_at = None
        pub_date = item.get("pubDate")
        if pub_date:
            try:
                posted_at = datetime.fromisoformat(pub_date)
            except ValueError:
                pass

        job = StandardizedJob(
            source_slug=self.source_slug,
            source_url=item.get("url", ""),
            source_job_id=str(item.get("id", "")),
            title=item.get("jobTitle", ""),
            company_name=item.get("companyName", ""),
            company_logo_url=item.get("companyLogo"),
            description=self.clean_html(item.get("jobDescription", "")),
            apply_url=item.get("url", ""),
            is_remote=True,
            remote_type=RemoteType.WORLDWIDE,
            location=item.get("jobGeo", "Worldwide"),
            tags=tags,
            skills=item.get("jobSkills", []) if isinstance(item.get("jobSkills"), list) else [],
            salary_min=salary_min,
            salary_max=salary_max,
            salary_currency=currency,
            posted_at=posted_at,
        )
        job.compute_fingerprint()
        return job
