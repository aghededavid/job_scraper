"""
RemoteScope Scraper — main entry point.
Delegates to the scheduler; supports --run-now and --source flags.

Usage:
    python -m scraper.main                    # Start scheduled mode
    python -m scraper.main --run-now          # Run all scrapers once and exit
    python -m scraper.main --source remoteok  # Run one source and exit
"""
import asyncio
import sys

from scraper.scheduler import main

if __name__ == "__main__":
    asyncio.run(main())
