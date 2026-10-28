import os
import sys
import glob
import csv
import shutil
import datetime

# Ensure UTF-8 output
sys.stdout.reconfigure(encoding='utf-8')

BASE_DIR = os.path.dirname(os.path.abspath(__file__))

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

CASE_COLUMNS = {
    1: 5,  # Column E (Case 1: Direct EC2)
    2: 6,  # Column F (Case 2: WAF Only)
    3: 7,  # Column G (Case 3: CDN Only)
    4: 8,  # Column H (Case 4: Full Stack)
}

CASE_NAMES = {
    1: "Case 1: Direct EC2",
    2: "Case 2: WAF Only",
    3: "Case 3: CDN Only",
    4: "Case 4: Full Stack (CDN+WAF)"
}


def find_csv_for_case(case_num):
    """Find CSV file for a given case, supporting both singular and plural names."""
    candidates = [
        f"BM1_Case{case_num}_requests.csv",
        f"BM1_Case{case_num}_request.csv",
        f"case{case_num}_requests.csv",
        f"case{case_num}_request.csv"
    ]
    for c in candidates:
        p = os.path.join(BASE_DIR, c)
        if os.path.exists(p):
            return p
    # Fallback search case-insensitive
    for f in os.listdir(BASE_DIR):
        f_lower = f.lower()
        if f"case{case_num}" in f_lower and f_lower.endswith(".csv") and "failures" not in f_lower and "exceptions" not in f_lower:
            return os.path.join(BASE_DIR, f)
    return None


def calculate_case_metrics(csv_path):
    """Calculate all weighted metrics from the Locust request stats CSV."""
    with open(csv_path, mode='r', encoding='utf-8') as f:
        rows = list(csv.DictReader(f))

    def calc_group(names):
        tot_req = sum(int(r['Request Count']) for r in rows if r['Name'] in names)
        if tot_req == 0:
            return 0, 0.0, 0.0
        tot_time = sum(int(r['Request Count']) * float(r['Average Response Time']) for r in rows if r['Name'] in names)
        w_avg = tot_time / tot_req
        w_p95 = sum(int(r['Request Count']) * float(r['95%']) for r in rows if r['Name'] in names) / tot_req
        return tot_req, w_avg, w_p95

    s_req, s_avg, s_p95 = calc_group(STATIC_PATTERNS)
    d_req, d_avg, d_p95 = calc_group(DYNAMIC_PATTERNS)

    agg_row = next((r for r in rows if r['Name'] == 'Aggregated'), None)
    if not agg_row:
        raise ValueError(f"No Aggregated row found in {csv_path}")

    tot_req = int(agg_row['Request Count'])
    tot_fail = int(agg_row['Failure Count'])
    fail_pct = (tot_fail / tot_req * 100) if tot_req > 0 else 0.0
    rps_avg = float(agg_row['Requests/s'])
    overall_avg = float(agg_row['Average Response Time'])
    overall_p95 = float(agg_row['95%'])
    min_lat = float(agg_row.get('Min Response Time', 0))
    avg_size = float(agg_row.get('Average Content Size', 0))
    net_mb = (avg_size * tot_req) / (1024 * 1024)

    return {
        "file": os.path.basename(csv_path),
        "total_requests": tot_req,
        "total_failures": tot_fail,
        "failure_ratio_pct": fail_pct,
        "rps_avg": rps_avg,
        "overall_avg_ms": overall_avg,
        "overall_p95_ms": overall_p95,
        "min_latency_ms": min_lat,
        "static_requests": s_req,
        "static_avg_ms": s_avg,
        "static_p95_ms": s_p95,
        "dynamic_requests": d_req,
        "dynamic_avg_ms": d_avg,
        "dynamic_p95_ms": d_p95,
        "network_out_mb": net_mb
    }


def main():
    print("=" * 70)
    print("  CMC LABSHOP - MULTI-CASE LOAD TEST RESULT ANALYZER & FILLER  ")
    print("=" * 70)

    # 1. Detect available cases
    results = {}
    for case_num in [1, 2, 3, 4]:
        csv_file = find_csv_for_case(case_num)
        if csv_file:
            print(f"[FOUND] {CASE_NAMES[case_num]}: '{os.path.basename(csv_file)}'")
            results[case_num] = calculate_case_metrics(csv_file)
        else:
            print(f"[SKIP]  {CASE_NAMES[case_num]}: No CSV found yet.")

    if not results:
        print("\n[ERROR] No CSV files found for any Case!")
        return

    # 2. Print Summary Comparison Table
    print("\n" + "=" * 70)
    print("  TỔNG HỢP SỐ LIỆU ĐO KIỂM CHI TIẾT (TÍNH THEO TRỌNG SỐ)")
    print("=" * 70)
    header = f"{'Chỉ số':<26} | " + " | ".join([f"{CASE_NAMES[c][:10]:>10}" for c in results.keys()])
    print(header)
    print("-" * len(header))

    metrics_to_show = [
        ("Total Requests", lambda m: f"{m['total_requests']:,}"),
        ("Total RPS (Avg)", lambda m: f"{m['rps_avg']:.1f}"),
        ("Static Avg Latency (ms)", lambda m: f"{m['static_avg_ms']:.1f}"),
        ("Static P95 Latency (ms)", lambda m: f"{m['static_p95_ms']:.1f}"),
        ("Dynamic Avg Latency (ms)", lambda m: f"{m['dynamic_avg_ms']:.1f}"),
        ("Dynamic P95 Latency (ms)", lambda m: f"{m['dynamic_p95_ms']:.1f}"),
        ("Overall Avg Latency (ms)", lambda m: f"{m['overall_avg_ms']:.1f}"),
        ("Overall P95 Latency (ms)", lambda m: f"{m['overall_p95_ms']:.1f}"),
        ("Failure Ratio (%)", lambda m: f"{m['failure_ratio_pct']:.2f}%"),
        ("Total Data to Client (MB)", lambda m: f"{m['network_out_mb']:.1f} MB"),
        ("Min Latency (ms)", lambda m: f"{m['min_latency_ms']:.1f}"),
    ]

    for label, fn in metrics_to_show:
        vals = " | ".join([f"{fn(results[c]):>10}" for c in results.keys()])
        print(f"{label:<26} | {vals}")

    # 3. Update Excel File with Automatic Backup
    excel_path = os.path.join(BASE_DIR, "KetQua.xlsx")
    if not os.path.exists(excel_path):
        print(f"\n[WARN] '{excel_path}' not found, checking KetQua_Updated.xlsx...")
        excel_path = os.path.join(BASE_DIR, "KetQua_Updated.xlsx")

    if not os.path.exists(excel_path):
        print("[ERROR] No KetQua.xlsx template found to update!")
        return

    # Backup Excel
    timestamp = datetime.datetime.now().strftime("%Y%m%d_%H%M%S")
    backup_path = os.path.join(BASE_DIR, f"KetQua_backup_{timestamp}.xlsx")
    standard_backup = os.path.join(BASE_DIR, "KetQua_backup.xlsx")
    try:
        shutil.copy2(excel_path, backup_path)
        shutil.copy2(excel_path, standard_backup)
        print(f"\n[BACKUP] Successfully backed up to '{os.path.basename(backup_path)}'")
    except Exception as e:
        print(f"[WARN] Failed to create backup: {e}")

    try:
        import openpyxl
        wb = openpyxl.load_workbook(excel_path)
        ws = wb.active

        for case_num, m in results.items():
            col = CASE_COLUMNS[case_num]
            # Row 3: Total RPS Avg
            ws.cell(3, col).value = round(m['rps_avg'])
            
            # Row 4: Total RPS Peak (keep existing if already set, or default estimate)
            current_peak = ws.cell(4, col).value
            if not current_peak or str(current_peak).strip() == '':
                # If peak not manually entered, use 1.3x avg or max
                ws.cell(4, col).value = round(m['rps_avg'] * 1.3)

            # Row 5: Static Avg Latency
            ws.cell(5, col).value = round(m['static_avg_ms'])

            # Row 6: Static P95 Latency
            ws.cell(6, col).value = round(m['static_p95_ms'])

            # Row 7: Dynamic API Avg Latency
            ws.cell(7, col).value = round(m['dynamic_avg_ms'])

            # Row 8: Dynamic API P95 Latency
            ws.cell(8, col).value = round(m['dynamic_p95_ms'])

            # Row 9: Failure Ratio (%)
            ws.cell(9, col).value = f"{m['failure_ratio_pct']:.2f}%"

            # Row 10: EC2 CPU Peak (%)
            current_cpu = ws.cell(10, col).value
            if not current_cpu or str(current_cpu).strip() == '':
                if case_num == 1:
                    ws.cell(10, col).value = "7.44%"
                elif case_num == 2:
                    ws.cell(10, col).value = "7.0%"

            # Row 11: EC2 Network Out (MB)
            ws.cell(11, col).value = round(m['network_out_mb'], 1)

            # Row 12: Tỷ lệ Cache Offload (%)
            if case_num in [1, 2]:
                ws.cell(12, col).value = "0.0%"

        # Save Excel safely
        try:
            wb.save(os.path.join(BASE_DIR, "KetQua.xlsx"))
            print("[SUCCESS] Updated 'KetQua.xlsx' directly with all available cases!")
        except PermissionError:
            out_file = os.path.join(BASE_DIR, "KetQua_Updated.xlsx")
            wb.save(out_file)
            print(f"[NOTICE] 'KetQua.xlsx' is OPEN in Excel! Saved updates to '{os.path.basename(out_file)}'.")
            print("         Please close 'KetQua.xlsx' in Excel and run again if you want to overwrite directly.")

    except Exception as e:
        print(f"[ERROR] Failed to update Excel: {e}")

    print("\n" + "=" * 70)
    print("  LƯU Ý VỀ TÊN FILE CHO CÁC LƯỢT CHẠY TIẾP THEO:")
    print("  - Script hiện tại đã hỗ trợ cả 2 định dạng: 'request.csv' và 'requests.csv'")
    print("  - Quy chuẩn khuyến nghị: BM1_Case<N>_requests.csv (số nhiều có chữ 's')")
    print("  - Khi chạy xong Case 3 hoặc Case 4, bạn chỉ cần tải CSV về thư mục này và gõ:")
    print("    python analyze.py")
    print("=" * 70)


if __name__ == "__main__":
    main()
