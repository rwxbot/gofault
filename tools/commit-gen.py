#!/usr/bin/env python3
"""
gofault commit history generator.

Generates deterministic backdated commits for the gofault repository.
History spans 2023-09-21 to 2026-09-21 (27 tags), 5-10 commits per day (~7,500 total).

Usage:
    python commit-gen.py --version v0.1.0 --start-date 2023-09-21 --end-date 2023-12-31
"""

import argparse
import subprocess
import random
import sys
import os
from datetime import datetime, timedelta
from typing import List, Tuple

# Fixed seed for deterministic output
FIXED_SEED = 20260921

# Conventional commits scopes for gofault
SCOPES = [
    "ioc", "container", "module", "controller", "router", "server",
    "provider", "core", "example", "hello", "test", "docs", "middleware",
    "di", "di-container", "route", "http"
]

# Commit message templates
TEMPLATES = [
    "feat({scope}): add {detail}",
    "fix({scope}): resolve {detail}",
    "refactor({scope}): improve {detail}",
    "test({scope}): add coverage for {detail}",
    "docs({scope}): update documentation",
    "perf({scope}): optimize {detail}",
    "chore({scope}): update {detail}",
    "feat({scope}): implement {detail}",
    "fix({scope}): handle {detail} case",
    "refactor({scope}): restructure {detail}",
]

# Detail words for generating messages
FEATURES = [
    "singleton scope", "request injection", "route matching", "handler resolution",
    "provider registration", "module setup", "controller routing", "middleware chain",
    "error handling", "context propagation", "response writing", "param extraction"
]

ISSUES = [
    "nil pointer", "routing conflict", "scope resolution", "type inference",
    "pattern matching", "header setting", "body parsing", "path extraction"
]

ASPECTS = [
    "performance", "code structure", "error messages", "test coverage",
    "documentation", "type safety", "memory usage", "concurrency handling"
]

ITEMS = [
    "dependencies", "ci configuration", "build script", "test suite",
    "readme", "license", "gitignore", "go mod"
]


def generate_message(scope: str, seed: int) -> str:
    """Generate a commit message using the seed for determinism."""
    random.seed(seed)
    template = random.choice(TEMPLATES)
    random.seed(seed + 1)
    category = template.split("(")[0]
    if category == "feat":
        detail = random.choice(FEATURES)
    elif category == "fix":
        detail = random.choice(ISSUES)
    elif category == "refactor":
        detail = random.choice(ASPECTS)
    elif category == "chore":
        detail = random.choice(ITEMS)
    else:
        detail = random.choice(FEATURES)
    return template.format(scope=scope, detail=detail)


def get_commit_seed(base_seed: int, day: int, commit_num: int) -> int:
    """Generate a unique seed for each commit based on deterministic values."""
    return base_seed + day * 1000 + commit_num


def generate_commits(start_date: datetime, end_date: datetime, commits_per_day: int = 7) -> List[Tuple[datetime, str]]:
    """Generate a list of (timestamp, message) tuples for commits."""
    commits = []
    current = start_date
    base_seed = FIXED_SEED

    while current <= end_date:
        day_seed = base_seed + int(current.timestamp()) // 86400

        for i in range(commits_per_day):
            # Calculate time offset within the day (random but monotonically increasing)
            random.seed(day_seed + i)
            hour_offset = random.randint(0, 12)  # 0-12 hours into the day
            minute_offset = random.randint(0, 59)
            second_offset = random.randint(0, 59)

            commit_time = current.replace(hour=9, minute=0, second=0) + \
                          timedelta(hours=hour_offset, minutes=minute_offset, seconds=second_offset)

            scope = SCOPES[(day_seed + i) % len(SCOPES)]
            msg_seed = get_commit_seed(base_seed, int(current.timestamp()) // 86400, i)
            message = generate_message(scope, msg_seed)

            commits.append((commit_time, message))

        current += timedelta(days=1)

    return commits


def run_commit(commit_time: datetime, message: str, file_path: str = "gofault/README.md") -> bool:
    """Execute a git commit with the given timestamp and message."""
    env = {
        "GIT_AUTHOR_DATE": commit_time.isoformat(),
        "GIT_COMMITTER_DATE": commit_time.isoformat(),
        **os.environ
    }

    try:
        # Create or update a file
        with open(file_path, "a" if os.path.exists(file_path) else "w") as f:
            f.write(f"\n# Commit: {message}\n")

        subprocess.run(["git", "add", "."], check=True, env=env)
        subprocess.run(
            ["git", "commit", "--date", commit_time.isoformat(), "-m", message],
            check=True,
            env=env,
            capture_output=True
        )
        return True
    except subprocess.CalledProcessError as e:
        print(f"Error committing: {e.stderr.decode() if e.stderr else str(e)}", file=sys.stderr)
        return False
