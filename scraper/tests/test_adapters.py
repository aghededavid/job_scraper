"""
Tests for scraper adapters.
Run with: cd scraper && python -m pytest tests/ -v
"""
import pytest
from scraper.adapters.base import BaseScraper
from scraper.models.job import StandardizedJob, FetchMethod, RemoteType


def test_make_fingerprint_deterministic():
    fp1 = BaseScraper.make_fingerprint("Senior Python Dev", "Acme Corp")
    fp2 = BaseScraper.make_fingerprint("Senior Python Dev", "Acme Corp")
    assert fp1 == fp2
    assert len(fp1) == 32


def test_make_fingerprint_different_titles():
    fp1 = BaseScraper.make_fingerprint("Backend Engineer", "Acme Corp")
    fp2 = BaseScraper.make_fingerprint("Frontend Engineer", "Acme Corp")
    assert fp1 != fp2


def test_make_fingerprint_case_insensitive():
    fp1 = BaseScraper.make_fingerprint("Senior Python Dev", "ACME CORP")
    fp2 = BaseScraper.make_fingerprint("senior python dev", "acme corp")
    assert fp1 == fp2


def test_clean_html_removes_tags():
    html = "<h1>Software Engineer</h1><p>We are <strong>hiring</strong>!</p>"
    result = BaseScraper.clean_html(html)
    assert "<" not in result
    assert "Software Engineer" in result
    assert "hiring" in result


def test_clean_html_collapses_whitespace():
    html = "<p>Hello   World</p>"
    result = BaseScraper.clean_html(html)
    assert "  " not in result


def test_parse_salary_usd_range():
    min_s, max_s, currency = BaseScraper.parse_salary("$120k-$160k")
    assert min_s == 120000
    assert max_s == 160000
    assert currency == "USD"


def test_parse_salary_eur():
    min_s, max_s, currency = BaseScraper.parse_salary("€80,000 - €100,000")
    assert currency == "EUR"
    assert min_s == 80000
    assert max_s == 100000


def test_parse_salary_single():
    min_s, max_s, currency = BaseScraper.parse_salary("$100k")
    assert min_s == 100000
    assert max_s == 100000


def test_parse_salary_empty():
    min_s, max_s, currency = BaseScraper.parse_salary("")
    assert min_s is None
    assert max_s is None
    assert currency == "USD"


def test_standardized_job_fingerprint():
    job = StandardizedJob(
        source_slug="remoteok",
        source_url="https://remoteok.com/1",
        title="Backend Engineer",
        company_name="TechCorp",
        apply_url="https://apply.com",
    )
    fp = job.compute_fingerprint()
    assert len(fp) == 32
    assert job.fingerprint == fp


def test_standardized_job_to_db_dict():
    job = StandardizedJob(
        source_slug="remoteok",
        source_url="https://remoteok.com/1",
        title="Backend Engineer",
        company_name="TechCorp",
        apply_url="https://apply.com",
        skills=["python", "go"],
        tags=["backend", "remote"],
    )
    d = job.to_db_dict()
    assert d["title"] == "Backend Engineer"
    assert d["company_name"] == "TechCorp"
    assert "python" in d["skills"]


def test_standardized_job_defaults():
    job = StandardizedJob(
        source_slug="test",
        source_url="https://example.com",
        title="Engineer",
        company_name="Corp",
        apply_url="https://example.com/apply",
    )
    assert job.is_remote is True
    assert job.remote_type == RemoteType.WORLDWIDE
    assert job.salary_currency == "USD"
    assert job.skills == []
    assert job.tags == []
