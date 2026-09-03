"""
We Work Remotely Adapter
Source: https://weworkremotely.com (RSS feeds)
Terms:  RSS feeds are publicly available.
"""
import xml.etree.ElementTree as ET
from datetime import datetime
from email.utils import parsedate_to_datetime

from scraper.adapters.base import BaseScraper
from scraper.models.job import StandardizedJob, FetchMethod, RemoteType

RSS_FEEDS = [
    ("https://weworkremotely.com/categories/remote-programming-jobs.rss", "Engineering"),
    ("https://weworkremotely.com/categories/remote-devops-sysadmin-jobs.rss", "DevOps"),
    ("https://weworkremotely.com/categories/remote-design-jobs.rss", "Design"),
    ("https://weworkremotely.com/categories/remote-product-jobs.rss", "Product"),
    ("https://weworkremotely.com/categories/remote-marketing-jobs.rss", "Marketing"),
    ("https://weworkremotely.com/categories/remote-customer-support-jobs.rss", "Customer Support"),
    ("https://weworkremotely.com/categories/remote-sales-jobs.rss", "Sales"),
    ("https://weworkremotely.com/categories/remote-writing-jobs.rss", "Writing"),
    ("https://weworkremotely.com/remote-jobs.rss", "All"),
]


class WWRScraper(BaseScraper):
    source_name = "We Work Remotely"
    source_slug = "weworkremotely"
    adapter_class = "WWRScraper"
    base_url = "https://weworkremotely.com"
    fetch_method = FetchMethod.RSS
    schedule_cron = "*/15 * * * *"
    rate_limit_rpm = 10

    async def fetch_jobs(self) -> list[StandardizedJob]:
        client = await self.get_client()
        jobs = []
        seen_urls = set()

        for feed_url, category in RSS_FEEDS:
            try:
                resp = await client.get(feed_url)
                resp.raise_for_status()
                parsed = self._parse_rss(resp.text, category)
                for job in parsed:
                    if job.source_url not in seen_urls:
                        seen_urls.add(job.source_url)
                        jobs.append(job)
            except Exception as e:
                self.logger.warning(f"Failed to fetch WWR feed {feed_url}: {e}")

        return jobs

    def _parse_rss(self, xml_text: str, category: str) -> list[StandardizedJob]:
        jobs = []
        try:
            root = ET.fromstring(xml_text)
            channel = root.find("channel")
            if channel is None:
                return jobs

            for item in channel.findall("item"):
                try:
                    title_elem = item.find("title")
                    link_elem = item.find("link")
                    desc_elem = item.find("description")
                    pub_elem = item.find("pubDate")

                    raw_title = title_elem.text if title_elem is not None else ""
                    if not raw_title:
                        continue

                    # WWR title format: "Company: Job Title"
                    parts = raw_title.split(": ", 1)
                    company = parts[0].strip() if len(parts) > 1 else "Unknown"
                    title = parts[1].strip() if len(parts) > 1 else raw_title

                    url = link_elem.text if link_elem is not None else ""
                    description = self.clean_html(desc_elem.text if desc_elem is not None else "")

                    posted_at = None
                    if pub_elem is not None and pub_elem.text:
                        try:
                            posted_at = parsedate_to_datetime(pub_elem.text)
                        except Exception:
                            pass

                    if not title or not url:
                        continue

                    job = StandardizedJob(
                        source_slug=self.source_slug,
                        source_url=url,
                        title=title,
                        company_name=company,
                        description=description,
                        apply_url=url,
                        is_remote=True,
                        remote_type=RemoteType.WORLDWIDE,
                        category=category,
                        tags=[category],
                        posted_at=posted_at,
                    )
                    job.compute_fingerprint()
                    jobs.append(job)
                except Exception as e:
                    self.logger.warning(f"Failed to parse WWR item: {e}")
        except ET.ParseError as e:
            self.logger.error(f"XML parse error: {e}")
        return jobs
