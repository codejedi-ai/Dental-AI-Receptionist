#!/usr/bin/env python3
"""
Migrate dentist_schedule_6_months.xlsx into PostgreSQL appointments table.
Each appointment blocks all 30-min slots it spans (using ceiling division).
A shared placeholder patient "Existing Booking" is used since the Excel has no patient info.
"""
import math
import openpyxl
import psycopg2
from datetime import datetime, timedelta

DB_URL = "host=localhost port=5432 dbname=dental user=dental password=internal_pg_2024"

def slots_for(start_str, end_str):
    """Return list of HH:MM slot strings covered by this appointment."""
    fmt = "%H:%M"
    start = datetime.strptime(start_str, fmt)
    end = datetime.strptime(end_str, fmt)
    duration_mins = int((end - start).total_seconds() / 60)
    num_slots = max(1, math.ceil(duration_mins / 30))
    result = []
    for i in range(num_slots):
        t = start + timedelta(minutes=30 * i)
        result.append(t.strftime("%H:%M"))
    return result

def main():
    conn = psycopg2.connect(DB_URL)
    cur = conn.cursor()

    # Create placeholder patient for imported appointments
    cur.execute("""
        INSERT INTO patients (name, mobile, email)
        VALUES ('Existing Booking', '0000000000', NULL)
        ON CONFLICT (name, mobile) DO UPDATE SET name = EXCLUDED.name
        RETURNING id
    """)
    placeholder_patient_id = cur.fetchone()[0]
    print(f"Placeholder patient ID: {placeholder_patient_id}")

    # Load dentist IDs
    cur.execute("SELECT id, name FROM dentists WHERE is_active = true")
    dentist_map = {name: did for did, name in cur.fetchall()}
    print(f"Dentists: {dentist_map}")

    # Load Excel
    wb = openpyxl.load_workbook('/root/Dental-AI-Reception/dentist_schedule_6_months.xlsx')
    ws = wb.active
    rows = list(ws.iter_rows(min_row=2, values_only=True))
    print(f"Rows to process: {len(rows)}")

    inserted = 0
    skipped = 0

    for row in rows:
        date_val, start_time, end_time, dentist_name, service, resource = row

        dentist_id = dentist_map.get(dentist_name)
        if not dentist_id:
            skipped += 1
            continue

        # Convert date to string if needed
        if hasattr(date_val, 'strftime'):
            date_str = date_val.strftime('%Y-%m-%d')
        else:
            date_str = str(date_val)

        for slot in slots_for(str(start_time), str(end_time)):
            try:
                cur.execute("""
                    INSERT INTO appointments
                        (patient_id, dentist_id, service, appointment_date, appointment_time, status, notes)
                    VALUES (%s, %s, %s, %s, %s, 'confirmed', 'Imported from schedule')
                    ON CONFLICT (dentist_id, appointment_date, appointment_time) DO NOTHING
                """, (placeholder_patient_id, dentist_id, service, date_str, slot))
                if cur.rowcount > 0:
                    inserted += 1
            except Exception as e:
                print(f"Error on {date_str} {slot} {dentist_name}: {e}")
                conn.rollback()

    conn.commit()
    cur.close()
    conn.close()
    print(f"\nDone — inserted: {inserted}, skipped (duplicate/unknown): {skipped}")

if __name__ == "__main__":
    main()
