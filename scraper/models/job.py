from __future__ import annotations

from datetime import datetime
from enum import Enum
from typing import Optional
from pydantic import BaseModel, Field, HttpUrl


class FetchMethod(str, Enum):
    API = "api"
    RSS = "rss"
    HTML = "html"
    WEBHOOK = "webhook"
    GRAPHQL = "graphql"


class ExperienceLevel(str, Enum):
    INTERNSHIP = "internship"
    ENTRY = "entry"
    JUNIOR = "junior"
    MID = "mid"
    SENIOR = "senior"
    STAFF = "staff"
    PRINCIPAL = "principal"
    ANY = "any"


class EmploymentType(str, Enum):
    FULL_TIME = "full_time"
    PART_TIME = "part_time"
    CONTRACT = "contract"
    FREELANCE = "freelance"
    TEMPORARY = "temporary"
    ANY = "any"


class RemoteType(str, Enum):
    WORLDWIDE = "worldwide"
    REGION_RESTRICTED = "region_restricted"
    ANY = "any"


class StandardizedJob(BaseModel):
    """
    The canonical job schema that every adapter must output.
    This is the contract between scrapers and the AI/ingestion pipeline.
    """

    # Source metadata
    source_slug: str = Field(..., description="Adapter source slug e.g. 'remoteok'")
    source_url: str = Field(..., description="Original URL on source site")
    source_job_id: Optional[str] = Field(None, description="ID used by source site")

    # Core job info
    title: str = Field(..., min_length=1)
    company_name: str = Field(..., min_length=1)
    company_website: Optional[str] = None
    company_logo_url: Optional[str] = None
    description: str = Field(default="")
    apply_url: str = Field(..., description="Direct application URL")

    # Location & remote
    is_remote: bool = True
    remote_type: RemoteType = RemoteType.WORLDWIDE
    location: Optional[str] = None
    timezone_restrictions: list[str] = Field(default_factory=list)

    # Classification (raw from source, AI will normalize)
    category: Optional[str] = None
    tags: list[str] = Field(default_factory=list)

    # Compensation
    salary_min: Optional[int] = None
    salary_max: Optional[int] = None
    salary_currency: str = "USD"
    salary_raw: Optional[str] = None  # Original salary string

    # Employment details
    experience_level: ExperienceLevel = ExperienceLevel.ANY
    employment_type: EmploymentType = EmploymentType.FULL_TIME
    skills: list[str] = Field(default_factory=list)
    technologies: list[str] = Field(default_factory=list)
    benefits: list[str] = Field(default_factory=list)
    visa_sponsorship: Optional[bool] = None

    # Timestamps
    posted_at: Optional[datetime] = None
    expires_at: Optional[datetime] = None

    # Deduplication
    fingerprint: Optional[str] = None

    class Config:
        use_enum_values = True

    def compute_fingerprint(self) -> str:
        """Compute and set the deduplication fingerprint."""
        import hashlib
        normalized = f"{self.title.lower().strip()}|{self.company_name.lower().strip()}"
        fp = hashlib.sha256(normalized.encode()).hexdigest()[:32]
        self.fingerprint = fp
        return fp

    def to_db_dict(self) -> dict:
        """Convert to a dict ready for database insertion."""
        return {
            "title": self.title,
            "company_name": self.company_name,
            "company_website": self.company_website,
            "company_logo_url": self.company_logo_url,
            "description": self.description,
            "apply_url": self.apply_url,
            "source_url": self.source_url,
            "source_slug": self.source_slug,
            "is_remote": self.is_remote,
            "remote_type": self.remote_type,
            "location": self.location,
            "timezone_restrictions": self.timezone_restrictions,
            "category": self.category,
            "tags": self.tags,
            "salary_min": self.salary_min,
            "salary_max": self.salary_max,
            "salary_currency": self.salary_currency,
            "experience_level": self.experience_level,
            "employment_type": self.employment_type,
            "skills": self.skills,
            "technologies": self.technologies,
            "benefits": self.benefits,
            "visa_sponsorship": self.visa_sponsorship,
            "posted_at": self.posted_at.isoformat() if self.posted_at else None,
            "expires_at": self.expires_at.isoformat() if self.expires_at else None,
            "fingerprint": self.fingerprint,
        }
