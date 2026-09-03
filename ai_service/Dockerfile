# ─────────────────────────────────────────────
# Stage 1: Development (hot-reload via uvicorn)
# ─────────────────────────────────────────────
FROM python:3.11-slim AS development

WORKDIR /app

# System dependencies for sentence-transformers / numpy
RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc g++ libpq-dev curl \
    && rm -rf /var/lib/apt/lists/*

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY . .

EXPOSE 8001

ENV PYTHONUNBUFFERED=1 \
    PYTHONDONTWRITEBYTECODE=1 \
    PYTHONPATH=/app

CMD ["uvicorn", "ai_service.main:app", "--host", "0.0.0.0", "--port", "8001", "--reload"]

# ─────────────────────────────────────────────
# Stage 2: Production
# ─────────────────────────────────────────────
FROM python:3.11-slim AS production

WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc g++ libpq-dev \
    && rm -rf /var/lib/apt/lists/*

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY . .

# Non-root user for security
RUN useradd -m appuser && chown -R appuser /app
USER appuser

EXPOSE 8001

ENV PYTHONUNBUFFERED=1 \
    PYTHONDONTWRITEBYTECODE=1 \
    PYTHONPATH=/app

CMD ["uvicorn", "ai_service.main:app", "--host", "0.0.0.0", "--port", "8001", "--workers", "2"]
