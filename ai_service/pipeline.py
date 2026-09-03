"""
Full AI processing pipeline.
Each job passes through all stages in sequence.
"""
import hashlib
import logging
import os
import re
from typing import Optional

import httpx

logger = logging.getLogger(__name__)

# ── Category taxonomy ─────────────────────────────────────────
CATEGORY_KEYWORDS = {
    "Frontend": ["frontend", "react", "vue", "angular", "next.js", "css", "html", "ui", "javascript", "typescript"],
    "Backend": ["backend", "api", "django", "flask", "fastapi", "rails", "spring", "node.js", "express", "golang", "rust", "java"],
    "Full Stack": ["full stack", "fullstack", "full-stack"],
    "DevOps & Cloud": ["devops", "kubernetes", "docker", "aws", "gcp", "azure", "terraform", "ci/cd", "infrastructure", "sre", "platform"],
    "AI & Machine Learning": ["machine learning", "ml", "ai ", "deep learning", "nlp", "llm", "pytorch", "tensorflow", "data science", "computer vision", "artificial intelligence"],
    "Data Science": ["data science", "data analyst", "analytics", "sql", "tableau", "power bi", "statistics", "pandas"],
    "Cybersecurity": ["security", "cybersecurity", "penetration", "soc", "siem", "vulnerability", "devsecops", "appsec", "infosec"],
    "Mobile": ["mobile", "ios", "android", "flutter", "react native", "swift", "kotlin", "xcode"],
    "Blockchain": ["blockchain", "web3", "solidity", "smart contract", "crypto", "defi", "nft"],
    "Embedded Systems": ["embedded", "firmware", "c++", "rtos", "microcontroller", "fpga", "iot"],
    "Product Management": ["product manager", "product management", "roadmap", "product owner"],
    "UI/UX Design": ["ux", "ui design", "figma", "sketch", "user experience", "user research", "interaction design"],
    "Graphic Design": ["graphic design", "illustrator", "photoshop", "branding", "visual design"],
    "Marketing": ["marketing", "seo", "content marketing", "growth", "paid ads", "social media"],
    "Sales": ["sales", "account executive", "business development", "revenue", "crm", "salesforce"],
    "Customer Support": ["customer support", "customer success", "customer service", "support engineer", "technical support"],
    "Finance": ["finance", "accounting", "cfo", "financial", "analyst", "fintech", "investment"],
    "HR & Recruiting": ["hr ", "human resources", "recruiting", "talent acquisition", "people operations"],
    "Legal": ["legal", "lawyer", "attorney", "compliance", "counsel"],
    "Writing & Content": ["writer", "content", "copywriter", "editor", "technical writer", "documentation"],
    "Healthcare": ["healthcare", "medical", "health", "biotech", "pharmaceutical", "clinical"],
    "Software Engineering": ["software engineer", "software developer", "engineer", "developer"],
}

# ── Title normalization map ───────────────────────────────────
TITLE_NORMALIZATION = {
    r"senior\s+(software|backend|frontend|full.?stack)": "Senior Software Engineer",
    r"(backend|back.?end)\s+(engineer|developer)": "Backend Engineer",
    r"(frontend|front.?end)\s+(engineer|developer)": "Frontend Engineer",
    r"full.?stack\s+(engineer|developer)": "Full Stack Engineer",
    r"software\s+engineer\s+(ii|iii|iv|2|3|4|l\d)": "Software Engineer",
    r"(sr\.?|senior)\s+swe": "Senior Software Engineer",
    r"(ml|machine learning)\s+engineer": "Machine Learning Engineer",
    r"(ai|artificial intelligence)\s+engineer": "AI Engineer",
    r"(devops|site reliability)\s+engineer": "DevOps Engineer",
    r"(cloud|infrastructure)\s+engineer": "Cloud Engineer",
    r"(data|analytics)\s+engineer": "Data Engineer",
    r"data\s+scientist": "Data Scientist",
    r"product\s+(manager|owner)": "Product Manager",
    r"(ux|ui|ux/ui)\s+(designer|researcher)": "UX Designer",
}

# ── Skills taxonomy ───────────────────────────────────────────
SKILL_KEYWORDS = {
    # Languages
    "python", "javascript", "typescript", "go", "golang", "rust", "java", "kotlin",
    "swift", "c++", "c#", "ruby", "php", "scala", "elixir", "haskell", "clojure",
    "r", "matlab", "dart", "perl",
    # Frontend
    "react", "vue", "angular", "next.js", "nuxt", "svelte", "html", "css", "tailwind",
    "webpack", "vite", "redux", "graphql",
    # Backend
    "node.js", "django", "flask", "fastapi", "rails", "spring", "express", "gin",
    "fiber", "actix", "nestjs", "laravel",
    # Databases
    "postgresql", "mysql", "mongodb", "redis", "elasticsearch", "cassandra",
    "dynamodb", "firebase", "supabase", "sqlite", "neo4j", "influxdb",
    # Cloud
    "aws", "gcp", "azure", "cloudflare", "vercel", "netlify", "heroku", "digitalocean",
    # DevOps
    "docker", "kubernetes", "k8s", "terraform", "ansible", "jenkins", "github actions",
    "gitlab ci", "circleci", "prometheus", "grafana", "datadog",
    # AI/ML
    "pytorch", "tensorflow", "scikit-learn", "pandas", "numpy", "langchain",
    "openai", "hugging face", "transformers", "spark", "kafka",
    # Mobile
    "flutter", "react native", "ios", "android", "xcode",
}


class AIPipeline:
    def __init__(self):
        self.openai_key = os.getenv("OPENAI_API_KEY", "")
        self.embedding_model = os.getenv("OPENAI_EMBEDDING_MODEL", "text-embedding-3-small")
        self.chat_model = os.getenv("OPENAI_CHAT_MODEL", "gpt-4o-mini")
        self._client: Optional[httpx.AsyncClient] = None

    async def initialize(self):
        self._client = httpx.AsyncClient(timeout=60.0)
        logger.info("AI Pipeline initialized")

    async def close(self):
        if self._client:
            await self._client.aclose()

    async def process(self, job_data: dict) -> dict:
        """Run the full AI pipeline on a job dict."""
        try:
            # Stage 1: Extract skills
            job_data = self._extract_skills(job_data)

            # Stage 2: Normalize title
            job_data = self._normalize_title(job_data)

            # Stage 3: Classify category
            job_data = self._classify_category(job_data)

            # Stage 4: Generate description summary
            if len(job_data.get("description", "")) > 200:
                job_data = await self._summarize(job_data)

            # Stage 5: Generate embedding
            job_data = await self._generate_embedding(job_data)

            # Stage 6: Fake job detection
            job_data = self._detect_fake(job_data)

            # Stage 7: Compute fingerprint if missing
            if not job_data.get("fingerprint"):
                title = job_data.get("title", "")
                company = job_data.get("company_name", "")
                fp = hashlib.sha256(f"{title.lower()}|{company.lower()}".encode()).hexdigest()[:32]
                job_data["fingerprint"] = fp

            job_data["ai_processed"] = True
            return job_data

        except Exception as e:
            logger.exception(f"Pipeline error: {e}")
            job_data["ai_processed"] = False
            return job_data

    def _extract_skills(self, job_data: dict) -> dict:
        """Extract skills and technologies from title + description."""
        text = (
            f"{job_data.get('title', '')} "
            f"{job_data.get('description', '')} "
            f"{' '.join(job_data.get('tags', []))}"
        ).lower()

        found_skills = []
        found_techs = []
        for skill in SKILL_KEYWORDS:
            if skill in text:
                found_skills.append(skill)
                found_techs.append(skill)

        # Merge with existing
        existing_skills = job_data.get("skills", [])
        existing_techs = job_data.get("technologies", [])
        job_data["skills"] = list(set(existing_skills + found_skills))
        job_data["technologies"] = list(set(existing_techs + found_techs))
        return job_data

    def _normalize_title(self, job_data: dict) -> dict:
        """Normalize job title to standard form."""
        title = job_data.get("title", "").lower()
        for pattern, normalized in TITLE_NORMALIZATION.items():
            if re.search(pattern, title, re.IGNORECASE):
                job_data["normalized_title"] = normalized
                return job_data
        # Default: title-case the original
        job_data["normalized_title"] = job_data.get("title", "").title()
        return job_data

    def _classify_category(self, job_data: dict) -> dict:
        """Classify job into a category based on keywords."""
        text = (
            f"{job_data.get('title', '')} "
            f"{' '.join(job_data.get('tags', []))} "
            f"{job_data.get('category', '')} "
            f"{job_data.get('description', '')[:500]}"
        ).lower()

        scores = {}
        for category, keywords in CATEGORY_KEYWORDS.items():
            score = sum(1 for kw in keywords if kw in text)
            if score > 0:
                scores[category] = score

        if scores:
            best = max(scores, key=scores.get)
            job_data["category"] = best
        elif not job_data.get("category"):
            job_data["category"] = "Software Engineering"

        return job_data

    async def _summarize(self, job_data: dict) -> dict:
        """Generate a 2-sentence summary of the job description."""
        if not self.openai_key:
            return job_data
        try:
            desc = job_data.get("description", "")[:3000]
            title = job_data.get("title", "")
            company = job_data.get("company_name", "")

            resp = await self._client.post(
                "https://api.openai.com/v1/chat/completions",
                headers={"Authorization": f"Bearer {self.openai_key}"},
                json={
                    "model": self.chat_model,
                    "messages": [
                        {"role": "system", "content": "You are a concise job description summarizer. Summarize in exactly 2 sentences."},
                        {"role": "user", "content": f"Job: {title} at {company}\n\nDescription:\n{desc}\n\nSummarize in 2 sentences:"},
                    ],
                    "max_tokens": 150,
                    "temperature": 0.3,
                },
            )
            result = resp.json()
            summary = result["choices"][0]["message"]["content"].strip()
            job_data["description_summary"] = summary
        except Exception as e:
            logger.warning(f"Summarization failed: {e}")
        return job_data

    async def _generate_embedding(self, job_data: dict) -> dict:
        """Generate embedding vector for the job."""
        if not self.openai_key:
            return job_data
        try:
            text = (
                f"{job_data.get('title', '')} "
                f"{job_data.get('company_name', '')} "
                f"{' '.join(job_data.get('skills', []))} "
                f"{job_data.get('description', '')[:1000]}"
            )

            resp = await self._client.post(
                "https://api.openai.com/v1/embeddings",
                headers={"Authorization": f"Bearer {self.openai_key}"},
                json={"model": self.embedding_model, "input": text},
            )
            result = resp.json()
            embedding = result["data"][0]["embedding"]
            job_data["embedding"] = embedding
        except Exception as e:
            logger.warning(f"Embedding failed: {e}")
        return job_data

    async def embed_text(self, text: str) -> Optional[list[float]]:
        """Embed arbitrary text (used by search)."""
        if not self.openai_key or not self._client:
            return None
        try:
            resp = await self._client.post(
                "https://api.openai.com/v1/embeddings",
                headers={"Authorization": f"Bearer {self.openai_key}"},
                json={"model": self.embedding_model, "input": text},
            )
            result = resp.json()
            return result["data"][0]["embedding"]
        except Exception as e:
            logger.warning(f"Embed text failed: {e}")
            return None

    def _detect_fake(self, job_data: dict) -> dict:
        """
        Detect suspicious/fake job listings using heuristics.
        Returns job_data with is_fake_suspected and fake_score.
        """
        score = 0.0
        title = job_data.get("title", "").lower()
        desc = job_data.get("description", "").lower()
        company = job_data.get("company_name", "").lower()

        # Red flags
        if any(kw in title for kw in ["make $", "earn from home", "unlimited income", "no experience required"]):
            score += 0.4
        if any(kw in desc for kw in ["mlm", "multi-level marketing", "pyramid", "investment required"]):
            score += 0.5
        if len(desc) < 100:
            score += 0.2
        if any(kw in company for kw in ["inc.", "llc"]) and len(company) < 5:
            score += 0.1
        if salary_min := job_data.get("salary_min"):
            if salary_min > 1_000_000:  # Suspiciously high
                score += 0.3

        job_data["fake_score"] = min(score, 1.0)
        job_data["is_fake_suspected"] = score >= 0.5
        return job_data
