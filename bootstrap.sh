#!/usr/bin/env sh
# RemoteScope — Bootstrap Script
# Run this once after cloning to generate go.sum files and install frontend deps.
# Usage: sh bootstrap.sh
set -e

echo "🚀 RemoteScope Bootstrap"
echo "========================="

# ── Go: backend ──────────────────────────────────────────────────────
if command -v go >/dev/null 2>&1; then
  echo "→ Running go mod tidy (backend)..."
  cd backend && go mod tidy && cd ..
  echo "✅ backend/go.sum generated"

  echo "→ Running go mod tidy (notification_service)..."
  cd notification_service && go mod tidy && cd ..
  echo "✅ notification_service/go.sum generated"
else
  echo "⚠  Go not found — skipping go mod tidy"
  echo "   The Dockerfiles use 'go.sum*' glob, so Docker build will still work."
  echo "   To generate go.sum locally: cd backend && go mod tidy"
fi

# ── Node: frontend ────────────────────────────────────────────────────
if command -v npm >/dev/null 2>&1; then
  echo "→ Installing frontend dependencies..."
  cd frontend && npm install && cd ..
  echo "✅ frontend/node_modules ready"
else
  echo "⚠  npm not found — skipping frontend install"
  echo "   Run: cd frontend && npm install"
fi

# ── Python: scraper ───────────────────────────────────────────────────
if command -v pip >/dev/null 2>&1; then
  echo "→ Installing scraper dependencies..."
  cd scraper && pip install -r requirements.txt -q && cd ..
  echo "✅ scraper dependencies installed"
fi

# ── Python: ai_service ────────────────────────────────────────────────
if command -v pip >/dev/null 2>&1; then
  echo "→ Installing ai_service dependencies..."
  cd ai_service && pip install -r requirements.txt -q && cd ..
  echo "✅ ai_service dependencies installed"
fi

echo ""
echo "✅ Bootstrap complete!"
echo "   Next: cp .env.example .env && fill in your secrets"
echo "   Then: make dev  (or: docker compose up --build -d)"
