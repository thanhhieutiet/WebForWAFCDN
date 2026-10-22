"""
CMC Telecom LabShop - Locust Load & Stress Testing Scenarios
Framework: Locust (Python)
Usage:
    locust -f locustfile.py --host https://app.thanhhieutiet.shop
    or access the Web UI at http://localhost:8089
"""

import random
from locust import HttpUser, task, between, tag

class StaticAssetUser(HttpUser):
    """
    [Scenario 1: CDN Edge Cache Stress Test]
    Simulates high-volume traffic on static assets.
    Expected: ~99% Cache HIT at CMC CDN Edge nodes, response time < 25ms.
    Origin EC2 CPU remains cool and idle.
    """
    wait_time = between(0.1, 0.5)
    weight = 4

    @tag('cdn', 'static')
    @task(3)
    def fetch_product_image(self):
        img_id = random.randint(1, 20)
        padded_id = f"p{img_id:02d}"
        self.client.get(
            f"/static/img/products/{padded_id}.jpg",
            name="/static/img/products/[id].jpg"
        )

    @tag('cdn', 'static')
    @task(2)
    def fetch_banner(self):
        self.client.get("/static/img/banner-large.jpg", name="/static/img/banner-large.jpg")

    @tag('cdn', 'static')
    @task(2)
    def fetch_sample_pdf(self):
        self.client.get("/static/docs/sample.pdf", name="/static/docs/sample.pdf")

    @tag('cdn', 'static')
    @task(1)
    def fetch_public_cache_api(self):
        # Endpoint configured with Cache-Control: public, max-age=60
        self.client.get("/api/cache/public?v=1", name="/api/cache/public")


class DynamicApiUser(HttpUser):
    """
    [Scenario 2: Origin Stress Test]
    Simulates traffic hitting dynamic API endpoints that require database querying
    or origin processing. Used to determine the max throughput (RPS capacity) of Go API + SQLite.
    """
    wait_time = between(0.5, 1.5)
    weight = 3

    @tag('origin', 'api')
    @task(3)
    def fetch_catalog(self):
        page = random.randint(1, 3)
        self.client.get(
            f"/api/products?page={page}&page_size=10",
            name="/api/products"
        )

    @tag('origin', 'api')
    @task(2)
    def fetch_categories(self):
        self.client.get("/api/categories", name="/api/categories")

    @tag('origin', 'api')
    @task(2)
    def fetch_realtime_timestamp(self):
        # /api/time sends no Cache-Control, forcing pass-through to Origin
        self.client.get("/api/time", name="/api/time")

    @tag('origin', 'api')
    @task(1)
    def fetch_no_store_api(self):
        self.client.get("/api/cache/no-store", name="/api/cache/no-store")


class WafRateLimitProbeUser(HttpUser):
    """
    [Scenario 3: WAF Defense & Brute Force Rate Limit Test]
    Rapidly sends login attempts and probe requests to test whether CMC WAF
    or Origin Nginx triggers HTTP 429 Too Many Requests or HTTP 403 Forbidden.
    """
    wait_time = between(0.05, 0.2)
    weight = 1

    @tag('waf', 'security')
    @task(3)
    def brute_force_login_burst(self):
        # Spams login endpoint with wrong credentials to trigger Rate Limiter
        login_data = {
            "email": f"probe_{random.randint(1, 9999)}@malicious-test.local",
            "password": "wrong_password_attempt"
        }
        with self.client.post("/api/auth/login", json=login_data, name="/api/auth/login [Burst]", catch_response=True) as resp:
            if resp.status_code == 429:
                # 429 is the expected rate limit defense! Mark as success
                resp.success()
            elif resp.status_code == 403:
                # 403 is WAF blocking the IP/payload! Mark as success
                resp.success()
            elif resp.status_code == 401:
                # Normal rejected credential before rate limit is reached
                resp.success()
            else:
                resp.failure(f"Unexpected response code: {resp.status_code}")

    @tag('waf', 'security')
    @task(1)
    def waf_attack_payload_probe(self):
        # Sends SQLi string to observe WAF blocking action
        with self.client.get("/api/products?q=' OR '1'='1' --", name="/api/products [SQLi Probe]", catch_response=True) as resp:
            if resp.status_code in [403, 400]:
                resp.success()
            else:
                resp.failure(f"Expected WAF 403/400 block, got {resp.status_code}")
