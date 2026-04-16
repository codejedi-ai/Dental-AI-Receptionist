#!/usr/bin/env python3
"""
Seed dentist_schedule_6_months.xlsx into PostgreSQL appointments.
Run on the host with DB_HOST=localhost (or inside Docker postgres hostname "postgres").
Usage: python3 seed_schedule.py [excel_path]
"""
import sys, math, os
import openpyxl
import psycopg2
from datetime import datetime, timedelta

EXCEL_PATH = sys.argv[1] if len(sys.argv) > 1 else "/tmp/schedule.xlsx"
# Use "postgres" when running inside Docker, fallback to "localhost" for local dev
DB_HOST = os.environ.get("DB_HOST", "localhost")
DB_PORT = os.environ.get("PGPORT", "5432")
DSN = f"host={DB_HOST} port={DB_PORT} dbname=dental user=dental password=internal_pg_2024"

def slots_for(start_str, end_str):
    fmt = "%H:%M"
    start = datetime.strptime(start_str, fmt)
    end = datetime.strptime(end_str, fmt)
    duration_mins = int((end - start).total_seconds() / 60)
    num_slots = max(1, math.ceil(duration_mins / 30))
    return [(start + timedelta(minutes=30 * i)).strftime("%H:%M") for i in range(num_slots)]

def main():
    if not os.path.exists(EXCEL_PATH):
        print(f"ERROR: Excel not found: {EXCEL_PATH}")
        sys.exit(1)

    conn = psycopg2.connect(DSN)
    cur = conn.cursor()

    # Placeholder patient for imported bookings
    cur.execute("""
        INSERT INTO patients (name, mobile, email)
        VALUES ('Existing Booking', '0000000000', NULL)
        ON CONFLICT (name, mobile) DO UPDATE SET name = EXCLUDED.name
        RETURNING id
    """)
    placeholder_id = cur.fetchone()[0]

    # Dentists
    cur.execute("SELECT id, name FROM dentists WHERE is_active = true")
    dentist_map = {name: did for did, name in cur.fetchall()}
    print(f"  Dentists: {list(dentist_map.keys())}")

    # Before count
    cur.execute("SELECT COUNT(*) FROM appointments")
    before = cur.fetchone()[0]
    print(f"  Appointments before seed: {before}")

    # Load Excel
    wb = openpyxl.load_workbook(EXCEL_PATH)
    ws = wb.active
    rows = list(ws.iter_rows(min_row=2, values_only=True))
    print(f"  Excel rows to process: {len(rows)}")

    inserted = 0
    skipped = 0
    for row in rows:
        date_val, start_time, end_time, dentist_name, service, resource = row
        dentist_id = dentist_map.get(dentist_name)
        if not dentist_id:
            skipped += 1
            continue
        date_str = date_val.strftime("%Y-%m-%d") if hasattr(date_val, "strftime") else str(date_val)
        for slot in slots_for(str(start_time), str(end_time)):
            try:
                cur.execute("""
                    INSERT INTO appointments
                        (patient_id, dentist_id, service, appointment_date, appointment_time, status, notes)
                    VALUES (%s, %s, %s, %s, %s, 'confirmed', 'Imported from schedule')
                    ON CONFLICT (dentist_id, appointment_date, appointment_time) DO NOTHING
                """, (placeholder_id, dentist_id, service, date_str, slot))
                if cur.rowcount > 0:
                    inserted += 1
            except Exception:
                pass

    conn.commit()

    cur.execute("SELECT COUNT(*) FROM appointments")
    after = cur.fetchone()[0]
    cur.execute("SELECT COUNT(DISTINCT appointment_date) FROM appointments")
    dates = cur.fetchone()[0]
    cur.close()
    conn.close()

    print(f"  Appointments: {before} -> {after} (+{inserted} new)")
    print(f"  Distinct booked dates: {dates}")
    print(f"  Skipped (unknown dentist): {skipped}")

if __name__ == "__main__":
    main()
