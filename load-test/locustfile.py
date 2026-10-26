"""
CMC Telecom LabShop - Locust Load & Stress Testing Scenarios
Distributed Load Testing Engine with Multi-Tier Defense Mechanics

Architecture Targets:
    Case 1 (Direct):    Client -> EC2 (Nginx + Go API + SQLite)
    Case 2 (WAF Only):  Client -> WAF -> EC2
    Case 3 (CDN Only):  Client -> CDN -> EC2
    Case 4 (Full):      Client -> CDN -> WAF -> EC2

Test Modes (Controlled via TEST_MODE env var):
    - performance: Focuses on CDN Edge Caching and Origin throughput (StaticAssetUser + DynamicApiUser).
                   WafRateLimitProbeUser is deactivated (weight=0) to prevent IP blast radius.
    - security:    Focuses on WAF rules and Rate Limiting defenses (WafRateLimitProbeUser only).
"""

import os
import random
import logging
from locust import HttpUser, task, between, tag, events

# Configure logger
logger = logging.getLogger("locustfile")

# Environment Configurations
TEST_MODE = os.getenv("TEST_MODE", "performance").strip().lower()
USE_RANGE_HEADER = os.getenv("USE_RANGE_HEADER", "true").strip().lower() in ("true", "1", "yes")

# Weight Allocation based on TEST_MODE
IS_SECURITY_MODE = (TEST_MODE == "security")
STATIC_WEIGHT = 0 if IS_SECURITY_MODE else 4
DYNAMIC_WEIGHT = 0 if IS_SECURITY_MODE else 3
WAF_WEIGHT = 1 if IS_SECURITY_MODE else 0

# When USE_RANGE_HEADER=True, explicitly specify Accept-Encoding: identity
# This prevents CDN from prioritizing Brotli/Gzip compression over Byte-Range serving (RFC 7233),
# ensuring CDN returns 206 Partial Content (1KB) instead of downloading full 2.78MB assets.
RANGE_HEADERS = {"Range": "bytes=0-1024", "Accept-Encoding": "identity"} if USE_RANGE_HEADER else {}


@events.init.add_listener
def on_locust_init(environment, **kwargs):
    """Log startup configuration on Master and Workers."""
    logger.info(
        f"[Engine Config] TEST_MODE='{TEST_MODE}' | USE_RANGE_HEADER={USE_RANGE_HEADER} | "
        f"Weights: [StaticAsset={STATIC_WEIGHT}, DynamicApi={DYNAMIC_WEIGHT}, WafProbe={WAF_WEIGHT}]"
    )


class StaticAssetUser(HttpUser):
    """
    [Scenario 1: CDN Edge Cache Stress Test]
    Simulates high-volume traffic on static assets.
    Expected: ~95-99% Cache HIT at CMC CDN Edge nodes, response time < 25ms.
    Origin EC2 CPU remains cool and network outbound is protected.

    Anti-Bottleneck Protection:
    When USE_RANGE_HEADER=true, requests for large assets send 'Range: bytes=0-1024'.
    This validates Cache HIT/MISS, byte serving, and disk I/O without saturating
    the client workstation's downlink network bandwidth (preventing artificial 80s latency).
    """
    wait_time = between(0.1, 0.5)
    weight = STATIC_WEIGHT

    @tag('performance', 'cdn', 'static')
    @task(4)
    def fetch_product_image(self):
        img_id = random.randint(1, 40)
        padded_id = f"p{img_id:02d}"
        url = f"/static/img/products/{padded_id}.jpg"
        with self.client.get(
            url,
            headers=RANGE_HEADERS,
            name="/static/img/products/[id].jpg",
            catch_response=True
        ) as resp:
            if resp.status_code in (200, 206, 304):
                resp.success()
            else:
                resp.failure(f"HTTP {resp.status_code} on {url}: {resp.text[:80]}")

    @tag('performance', 'cdn', 'static')
    @task(2)
    def fetch_banner(self):
        # 2.5MB heavy JPEG asset
        with self.client.get(
            "/static/img/banner-large.jpg",
            headers=RANGE_HEADERS,
            name="/static/img/banner-large.jpg",
            catch_response=True
        ) as resp:
            if resp.status_code in (200, 206, 304):
                resp.success()
            else:
                resp.failure(f"HTTP {resp.status_code} on banner: {resp.text[:80]}")

    @tag('performance', 'cdn', 'static')
    @task(2)
    def fetch_sample_pdf(self):
        # Document PDF asset
        with self.client.get(
            "/static/docs/sample.pdf",
            headers=RANGE_HEADERS,
            name="/static/docs/sample.pdf",
            catch_response=True
        ) as resp:
            if resp.status_code in (200, 206, 304):
                resp.success()
            else:
                resp.failure(f"HTTP {resp.status_code} on sample.pdf: {resp.text[:80]}")

    @tag('performance', 'cdn', 'static')
    @task(1)
    def fetch_public_cache_api(self):
        # Endpoint configured with Cache-Control: public, max-age=60
        with self.client.get(
            "/api/cache/public?v=1",
            name="/api/cache/public",
            catch_response=True
        ) as resp:
            if resp.status_code in (200, 304):
                resp.success()
            else:
                resp.failure(f"HTTP {resp.status_code} on /api/cache/public: {resp.text[:80]}")


class DynamicApiUser(HttpUser):
    """
    [Scenario 2: Origin Stress Test]
    Simulates traffic hitting dynamic API endpoints that require database querying
    or origin processing. Used to determine the max throughput (RPS capacity) of Go API + SQLite.
    All query strings are normalized with grouped name='...' to prevent Prometheus high-cardinality explosions.
    """
    wait_time = between(0.4, 1.2)
    weight = DYNAMIC_WEIGHT

    @tag('performance', 'origin', 'api')
    @task(3)
    def fetch_catalog(self):
        page = random.randint(1, 4)
        with self.client.get(
            f"/api/products?page={page}&page_size=10",
            name="/api/products",
            catch_response=True
        ) as resp:
            if resp.status_code == 200:
                resp.success()
            else:
                resp.failure(f"HTTP {resp.status_code} on /api/products: {resp.text[:80]}")

    @tag('performance', 'origin', 'api')
    @task(2)
    def fetch_categories(self):
        with self.client.get("/api/categories", name="/api/categories", catch_response=True) as resp:
            if resp.status_code == 200:
                resp.success()
            else:
                resp.failure(f"HTTP {resp.status_code} on /api/categories: {resp.text[:80]}")

    @tag('performance', 'origin', 'api')
    @task(2)
    def fetch_realtime_timestamp(self):
        # /api/time sends no Cache-Control, forcing pass-through to Origin
        with self.client.get("/api/time", name="/api/time", catch_response=True) as resp:
            if resp.status_code == 200:
                resp.success()
            else:
                resp.failure(f"HTTP {resp.status_code} on /api/time: {resp.text[:80]}")

    @tag('performance', 'origin', 'api')
    @task(1)
    def fetch_no_store_api(self):
        with self.client.get("/api/cache/no-store", name="/api/cache/no-store", catch_response=True) as resp:
            if resp.status_code == 200:
                resp.success()
            else:
                resp.failure(f"HTTP {resp.status_code} on /api/cache/no-store: {resp.text[:80]}")


class WafRateLimitProbeUser(HttpUser):
    """
    [Scenario 3: WAF Defense & Brute Force Rate Limit Test]
    Rapidly sends login attempts and probe requests to test whether CMC WAF
    or Origin Nginx triggers HTTP 429 Too Many Requests or HTTP 403 Forbidden.
    Active ONLY when TEST_MODE=security (weight=1), isolated from performance benchmarks.
    """
    wait_time = between(0.05, 0.2)
    weight = WAF_WEIGHT

    @tag('security', 'waf')
    @task(3)
    def brute_force_login_burst(self):
        # Spams login endpoint with invalid credentials to trigger Rate Limiter
        login_data = {
            "email": f"probe_{random.randint(1, 99999)}@malicious-test.local",
            "password": "wrong_password_attempt_attack"
        }
        with self.client.post(
            "/api/auth/login",
            json=login_data,
            name="/api/auth/login [Burst]",
            catch_response=True
        ) as resp:
            if resp.status_code in (429, 403):
                # 429 = Rate limited by WAF or Nginx; 403 = WAF IP defense! Mark as success
                resp.success()
            elif resp.status_code == 401:
                # Normal rejected credential before rate limit threshold is crossed
                resp.success()
            else:
                resp.failure(f"Unexpected response code for login burst: {resp.status_code}")

    @tag('security', 'waf')
    @task(1)
    def waf_attack_payload_probe(self):
        # Sends SQLi string to observe WAF rule engine blocking action
        with self.client.get(
            "/api/products?q=' OR '1'='1' --",
            name="/api/products [SQLi Probe]",
            catch_response=True
        ) as resp:
            if resp.status_code in (403, 400):
                # 403 = WAF Rule Engine Blocked; 400 = WAF/App payload validation rejected
                resp.success()
            else:
                resp.failure(f"Expected WAF 403/400 block, got HTTP {resp.status_code}")
