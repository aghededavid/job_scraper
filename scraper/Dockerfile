FROM python:3.12-slim

WORKDIR /app

# Install system dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc \
    && rm -rf /var/lib/apt/lists/*

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY . .

# Create package structure
RUN touch scraper/__init__.py scraper/adapters/__init__.py scraper/models/__init__.py

CMD ["python", "-m", "scraper.scheduler"]
