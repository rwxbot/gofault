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
