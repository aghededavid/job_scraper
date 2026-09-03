"""
Tests for the AI pipeline.
Run with: cd ai_service && python -m pytest tests/ -v
"""
import pytest
from ai_service.pipeline import AIPipeline


@pytest.fixture
def pipeline():
    return AIPipeline()


def test_extract_skills_basic(pipeline):
    job = {
        "title": "Senior Python Developer",
        "description": "We need someone with Python, Django, PostgreSQL, and Docker experience.",
        "tags": [],
    }
    result = pipeline._extract_skills(job)
    assert "python" in result["skills"]
    assert "docker" in result["technologies"]


def test_extract_skills_merges_existing(pipeline):
    job = {
        "title": "Backend Engineer",
        "description": "Go, Rust, and Kubernetes.",
        "tags": [],
        "skills": ["java"],
        "technologies": ["terraform"],
    }
    result = pipeline._extract_skills(job)
    assert "java" in result["skills"]
    assert "terraform" in result["technologies"]
    assert "go" in result["skills"]
    assert "kubernetes" in result["technologies"]


def test_normalize_title_senior(pipeline):
    job = {"title": "Senior Software Engineer"}
    result = pipeline._normalize_title(job)
    assert result["normalized_title"] == "Senior Software Engineer"


def test_normalize_title_ml(pipeline):
    job = {"title": "ML Engineer II"}
    result = pipeline._normalize_title(job)
    assert result["normalized_title"] == "Machine Learning Engineer"


def test_classify_category_frontend(pipeline):
    job = {
        "title": "React Frontend Developer",
        "description": "Build React, Vue, and TypeScript applications.",
        "tags": ["react", "typescript"],
        "category": "",
    }
    result = pipeline._classify_category(job)
    assert result["category"] == "Frontend"


def test_classify_category_devops(pipeline):
    job = {
        "title": "DevOps Engineer",
        "description": "Kubernetes, Terraform, AWS, CI/CD pipelines.",
        "tags": ["kubernetes", "aws"],
        "category": "",
    }
    result = pipeline._classify_category(job)
    assert result["category"] == "DevOps & Cloud"


def test_classify_category_default(pipeline):
    job = {"title": "Something Unusual", "description": "", "tags": [], "category": ""}
    result = pipeline._classify_category(job)
    assert result["category"] == "Software Engineering"


def test_detect_fake_clean_job(pipeline):
    job = {
        "title": "Software Engineer",
        "description": "We are building a world-class platform for remote teams. Join our distributed engineering team.",
        "company_name": "Acme Corp",
    }
    result = pipeline._detect_fake(job)
    assert result["fake_score"] < 0.5
    assert result["is_fake_suspected"] is False


def test_detect_fake_suspicious_job(pipeline):
    job = {
        "title": "Earn from home unlimited income",
        "description": "MLM investment required join now",
        "company_name": "X",
    }
    result = pipeline._detect_fake(job)
    assert result["is_fake_suspected"] is True
    assert result["fake_score"] >= 0.5


def test_full_pipeline_sync_stages(pipeline):
    """Test all sync stages of the pipeline together."""
    job = {
        "title": "Senior Python Backend Engineer",
        "description": "Django, PostgreSQL, Redis, Docker. 3+ years experience.",
        "tags": ["python", "django"],
        "company_name": "TechCorp",
    }
    job = pipeline._extract_skills(job)
    job = pipeline._normalize_title(job)
    job = pipeline._classify_category(job)
    job = pipeline._detect_fake(job)

    assert "python" in job["skills"]
    assert "normalized_title" in job
    assert "category" in job
    assert "fake_score" in job
    assert job["is_fake_suspected"] is False
