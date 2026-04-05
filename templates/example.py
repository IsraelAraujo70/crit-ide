#!/usr/bin/env python3
"""A simple example showcasing Python syntax highlighting."""

import os
from dataclasses import dataclass, field
from typing import Optional

# Constants
MAX_RETRIES = 3
DEFAULT_TIMEOUT = 30.0

@dataclass
class Config:
    """Application configuration."""
    host: str = "localhost"
    port: int = 8080
    debug: bool = False
    tags: list[str] = field(default_factory=list)

class Server:
    """A minimal HTTP server wrapper."""

    def __init__(self, config: Optional[Config] = None):
        self.config = config or Config()
        self._running = False

    async def start(self):
        """Start the server."""
        self._running = True
        print(f"Server starting on {self.config.host}:{self.config.port}")

        for attempt in range(MAX_RETRIES):
            try:
                await self._bind()
                break
            except OSError as e:
                if attempt == MAX_RETRIES - 1:
                    raise RuntimeError("Failed to bind") from e
                print(f"Retry {attempt + 1}/{MAX_RETRIES}...")

    async def _bind(self):
        """Bind to the configured address."""
        pass  # Placeholder

    def stop(self):
        """Stop the server gracefully."""
        self._running = False
        print("Server stopped")

    @property
    def is_running(self) -> bool:
        return self._running


def calculate_stats(numbers: list[float]) -> dict:
    """Calculate basic statistics for a list of numbers."""
    if not numbers:
        return {"count": 0, "sum": 0.0, "mean": None}

    total = sum(numbers)
    count = len(numbers)
    mean = total / count
    sorted_nums = sorted(numbers)
    median = sorted_nums[count // 2]

    return {
        "count": count,
        "sum": total,
        "mean": round(mean, 2),
        "median": median,
        "min": min(numbers),
        "max": max(numbers),
    }


if __name__ == "__main__":
    config = Config(host="0.0.0.0", port=3000, debug=True)
    server = Server(config)

    data = [23.5, 18.0, 42.1, 7.8, 31.4]
    stats = calculate_stats(data)
    print(f"Stats: {stats}")

    # Triple-quoted string
    query = '''
    SELECT name, email
    FROM users
    WHERE active = True
    '''
    print(query)
