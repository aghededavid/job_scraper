"""
Greenhouse Adapter
Source: https://boards-api.greenhouse.io/v1/boards/{company_token}/jobs
Terms:  Public job board API. No auth required. Permissioned for public access.
        We only query companies that have public Greenhouse boards.
"""
from datetime import datetime

from scraper.adapters.base import BaseScraper
from scraper.models.job import StandardizedJob, FetchMethod, RemoteType

# Known remote-friendly companies using Greenhouse
# Add more as discovered
GREENHOUSE_COMPANIES = [
    "airbnb", "stripe", "cloudflare", "hashicorp", "datadog",
    "mongodb", "elastic", "netlify", "supabase", "vercel",
    "gitlab", "figma", "notion", "linear", "loom",
    "brex", "plaid", "razorpay", "mercury", "ramp",
    "hubspot", "zendesk", "intercom", "mixpanel", "segment",
    "snowflake", "databricks", "confluent", "dbt-labs",
    "scale-ai", "weights-biases", "hugging-face",
]


class GreenhouseScraper(BaseScraper):
    source_name = "Greenhouse"
    source_slug = "greenhouse"
    adapter_class = "GreenhouseScraper"
    base_url = "https://boards-api.greenhouse.io/v1/boards"
    fetch_method = FetchMethod.API
    schedule_cron = "*/30 * * * *"
    rate_limit_rpm = 60

    async def fetch_jobs(self) -> list[StandardizedJob]:
        client = await self.get_client()
        jobs = []

        for company_token in GREENHOUSE_COMPANIES:
            try:
                url = f"{self.base_url}/{company_token}/jobs"
                resp = await client.get(url, params={"content": "true"})
                if resp.status_code == 404:
                    continue
                resp.raise_for_status()
                data = resp.json()

                for item in data.get("jobs", []):
                    try:
                        job = self._parse_job(item, company_token)
                        if job and self._is_remote(item):
                            jobs.append(job)
                    except Exception as e:
                        self.logger.warning(f"Failed to parse Greenhouse job: {e}")
            except Exception as e:
                self.logger.warning(f"Failed to fetch Greenhouse {company_token}: {e}")

        return jobs

    def _is_remote(self, item: dict) -> bool:
        """Check if job is remote based on location or title keywords."""
        location = (item.get("location", {}) or {}).get("name", "").lower()
        title = item.get("title", "").lower()
        remote_keywords = ["remote", "anywhere", "worldwide", "distributed"]
        return any(kw in location or kw in title for kw in remote_keywords)

    def _parse_job(self, item: dict, company_token: str):
        departments = [d.get("name", "") for d in (item.get("departments") or [])]
        offices = [(o.get("name") or "") for o in (item.get("offices") or [])]

        posted_at = None
        updated_at = item.get("updated_at") or item.get("created_at")
        if updated_at:
            try:
                posted_at = datetime.fromisoformat(updated_at.replace("Z", "+00:00"))
            except ValueError:
                pass

        content = item.get("content", "")
        description = self.clean_html(content) if content else ""

        # Infer company name from token
        company_name = company_token.replace("-", " ").title()

        location_str = (item.get("location", {}) or {}).get("name", "Remote")

        job = StandardizedJob(
            source_slug=self.source_slug,
            source_url=item.get("absolute_url", ""),
            source_job_id=str(item.get("id", "")),
            title=item.get("title", ""),
            company_name=company_name,
            description=description,
            apply_url=item.get("absolute_url", ""),
            is_remote=True,
            remote_type=RemoteType.WORLDWIDE,
            location=location_str,
            category=departments[0] if departments else None,
            tags=departments,
            posted_at=posted_at,
        )
        job.compute_fingerprint()
        return job
