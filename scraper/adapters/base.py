from abc import ABC, abstractmethod
from typing import Optional
import hashlib
import logging
import httpx
import asyncio

from scraper.models.job import StandardizedJob, FetchMethod

logger = logging.getLogger(__name__)


class BaseScraper(ABC):
    """
    Abstract base class for all job source adapters.

    Each adapter must implement `fetch_jobs()` and set the class-level attributes.
    The base class handles rate limiting, error handling, retry logic, and normalization.
    """

    # Required class-level attributes
    source_name: str = ""
    source_slug: str = ""
    adapter_class: str = ""
    base_url: str = ""
    fetch_method: FetchMethod = FetchMethod.API
    schedule_cron: str = "*/15 * * * *"
    rate_limit_rpm: int = 30
    respects_robots: bool = True

    def __init__(self):
        self.logger = logging.getLogger(self.__class__.__name__)
        self._client: Optional[httpx.AsyncClient] = None

    async def get_client(self) -> httpx.AsyncClient:
        if self._client is None or self._client.is_closed:
            self._client = httpx.AsyncClient(
                timeout=httpx.Timeout(30.0),
                headers={
                    "User-Agent": "RemoteScope/1.0 (job aggregator; contact@remotescope.app)",
                    "Accept": "application/json, text/html, application/xml",
                },
                follow_redirects=True,
            )
        return self._client

    @abstractmethod
    async def fetch_jobs(self) -> list[StandardizedJob]:
        """
        Fetch and return a list of StandardizedJob objects from the source.
        Must be implemented by each adapter.
        """
        ...

    async def run(self) -> list[StandardizedJob]:
        """
        Execute the scraper with error handling and rate limiting.
        """
        try:
            self.logger.info(f"[{self.source_name}] Starting scrape...")
            # Rate limit: sleep to respect RPM
            delay = 60.0 / self.rate_limit_rpm
            await asyncio.sleep(delay)

            jobs = await self.fetch_jobs()
            self.logger.info(f"[{self.source_name}] Fetched {len(jobs)} jobs")
            return jobs
        except httpx.HTTPStatusError as e:
            self.logger.error(f"[{self.source_name}] HTTP error: {e.response.status_code} - {e}")
            return []
        except httpx.TimeoutException:
            self.logger.error(f"[{self.source_name}] Request timed out")
            return []
        except Exception as e:
            self.logger.exception(f"[{self.source_name}] Unexpected error: {e}")
            return []
        finally:
            if self._client:
                await self._client.aclose()
                self._client = None

    @staticmethod
    def make_fingerprint(title: str, company: str, description: str = "") -> str:
        """Generate a deduplication fingerprint for a job."""
        normalized = f"{title.lower().strip()}|{company.lower().strip()}"
        if description:
            # Include first 500 chars of description for better dedup
            normalized += f"|{description[:500].lower().strip()}"
        return hashlib.sha256(normalized.encode()).hexdigest()[:32]

    @staticmethod
    def clean_html(text: str) -> str:
        """Remove basic HTML tags from text."""
        import re
        text = re.sub(r'<[^>]+>', ' ', text)
        text = re.sub(r'\s+', ' ', text)
        return text.strip()

    @staticmethod
    def parse_salary(salary_str: str) -> tuple[Optional[int], Optional[int], str]:
        """
        Parse salary string to (min, max, currency).
        Examples: "$120k-$160k", "€80,000 - €100,000", "$200/hr"
        """
        import re
        if not salary_str:
            return None, None, "USD"

        salary_str = salary_str.strip()
        currency = "USD"

        if "€" in salary_str or "EUR" in salary_str:
            currency = "EUR"
        elif "£" in salary_str or "GBP" in salary_str:
            currency = "GBP"
        elif "CAD" in salary_str:
            currency = "CAD"
        elif "AUD" in salary_str:
            currency = "AUD"

        # Extract numbers
        numbers = re.findall(r'[\d,]+(?:\.\d+)?', salary_str.replace(',', ''))
        nums = [int(float(n)) for n in numbers if n]

        if not nums:
            return None, None, currency

        # Handle "k" suffix
        if 'k' in salary_str.lower():
            nums = [n * 1000 if n < 1000 else n for n in nums]

        if len(nums) == 1:
            return nums[0], nums[0], currency
        if len(nums) >= 2:
            return min(nums[:2]), max(nums[:2]), currency

        return None, None, currency
