import os
import sys
import csv

sys.stdout.reconfigure(encoding='utf-8')
BASE_DIR = os.path.dirname(os.path.abspath(__file__))

files = {
    'CMC -> CDN': 'BM2_CMC_CDN_requests.csv',
    '4G  -> CDN': 'BM2_4G_CDN_requests.csv',
    'CMC -> EC2 (V2)': 'BM2_CMC_EC2_V2_requests.csv',
    '4G  -> EC2 (V2)': 'BM2_4G_EC2_V2_requests.csv'
}

STATIC_PATTERNS = [
    '/static/img/products/[id].jpg',
    '/static/img/banner-large.jpg',
    '/static/docs/sample.pdf'
]

DYNAMIC_PATTERNS = [
    '/api/products',
    '/api/categories',
    '/api/time',
    '/api/cache/no-store',
    '/api/cache/public'
]

def analyze_file(fn):
    p = os.path.join(BASE_DIR, fn)
    with open(p, 'r', encoding='utf-8') as f:
        rows = list(csv.DictReader(f))
    agg = [r for r in rows if r['Name'] == 'Aggregated'][0]
    
    s_rows = [r for r in rows if r['Name'] in STATIC_PATTERNS]
    s_req = sum(int(r['Request Count']) for r in s_rows)
    s_avg = sum(int(r['Request Count']) * float(r['Average Response Time']) for r in s_rows) / s_req if s_req else 0
    s_p95 = sum(int(r['Request Count']) * float(r['95%']) for r in s_rows) / s_req if s_req else 0

    d_rows = [r for r in rows if r['Name'] in DYNAMIC_PATTERNS]
    d_req = sum(int(r['Request Count']) for r in d_rows)
    d_avg = sum(int(r['Request Count']) * float(r['Average Response Time']) for r in d_rows) / d_req if d_req else 0
    d_p95 = sum(int(r['Request Count']) * float(r['95%']) for r in d_rows) / d_req if d_req else 0

    tot_req = int(agg['Request Count'])
    tot_fail = int(agg['Failure Count'])
    return {
        'total_req': tot_req,
        'failures': tot_fail,
        'fail_pct': (tot_fail / tot_req * 100) if tot_req > 0 else 0,
        'rps': float(agg['Requests/s']),
        'avg_lat': float(agg['Average Response Time']),
        'p95_lat': float(agg['95%']),
        'min_lat': float(agg['Min Response Time']),
        'max_lat': float(agg['Max Response Time']),
        'static_req': s_req,
        'static_avg': s_avg,
        'static_p95': s_p95,
        'dynamic_req': d_req,
        'dynamic_avg': d_avg,
        'dynamic_p95': d_p95,
        'data_mb': (float(agg['Average Content Size']) * tot_req) / (1024 * 1024)
    }

results = {k: analyze_file(v) for k, v in files.items()}

print("\n" + "="*80)
print("  BM2: MULTI-NETWORK BENCHMARK REPORT (CMC FTTH vs MOBILE 4G)")
print("  Config: 30 CCU | Ramp-up: 10/s | Duration: 120s (2 minutes)")
print("="*80)

header = f"| {'Metric':<28} | " + " | ".join([f"{k:>12}" for k in results.keys()]) + " |"
print(header)
sep = "|:" + "-"*28 + ":|" + "|".join([":" + "-"*12 + ":" for _ in results.keys()]) + "|"
print(sep)

rows_to_print = [
    ("Total Requests", lambda r: f"{r['total_req']:,}"),
    ("Requests/sec (RPS)", lambda r: f"{r['rps']:.1f}"),
    ("Overall Avg Latency (ms)", lambda r: f"{r['avg_lat']:.1f} ms"),
    ("Overall P95 Latency (ms)", lambda r: f"{r['p95_lat']:.1f} ms"),
    ("Min Latency (ms)", lambda r: f"{r['min_lat']:.1f} ms"),
    ("Max Latency (ms)", lambda r: f"{r['max_lat']:.1f} ms"),
    ("Static Avg Latency (ms)", lambda r: f"{r['static_avg']:.1f} ms"),
    ("Static P95 Latency (ms)", lambda r: f"{r['static_p95']:.1f} ms"),
    ("Dynamic Avg Latency (ms)", lambda r: f"{r['dynamic_avg']:.1f} ms"),
    ("Dynamic P95 Latency (ms)", lambda r: f"{r['dynamic_p95']:.1f} ms"),
    ("Failure Ratio (%)", lambda r: f"{r['fail_pct']:.2f}%"),
    ("Data Transferred (MB)", lambda r: f"{r['data_mb']:.2f} MB")
]

for label, fn in rows_to_print:
    row_str = " | ".join([f"{fn(v):>12}" for v in results.values()])
    print(f"| {label:<28} | {row_str} |")
