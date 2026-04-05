import { NextRequest, NextResponse } from "next/server";
import { getDb } from "@/lib/db";

export async function GET(req: NextRequest) {
  const db = getDb();
  const url = new URL(req.url);
  const status = url.searchParams.get("status");
  const date = url.searchParams.get("date");
  const search = url.searchParams.get("search");

  let query = `
    SELECT a.*, p.name as patient_name, p.email as patient_email, p.phone as patient_phone
    FROM appointments a
    LEFT JOIN patients p ON a.patient_id = p.id
    WHERE 1=1
  `;
  const params: any[] = [];

  if (status && status !== "all") {
    query += " AND a.status = ?";
    params.push(status);
  }
  if (date) {
    query += " AND a.date = ?";
    params.push(date);
  }
  if (search) {
    query += " AND (p.name LIKE ? OR a.service LIKE ? OR a.dentist LIKE ?)";
    const s = `%${search}%`;
    params.push(s, s, s);
  }

  query += " ORDER BY a.date DESC, a.time ASC";

  const appointments = db.prepare(query).all(...params);
  return NextResponse.json({ appointments });
}

export async function POST(req: NextRequest) {
  const db = getDb();
  const body = await req.json();
  const { service, dentist, date, time, name, email, phone, notes } = body;

  if (!service || !dentist || !date || !time || !name) {
    return NextResponse.json({ error: "Missing required fields" }, { status: 400 });
  }

  // Find or create patient
  let patient = db
    .prepare("SELECT * FROM patients WHERE email = ? OR (name = ? AND phone = ?)")
    .get(email, name, phone) as any;

  if (!patient) {
    const result = db
      .prepare("INSERT INTO patients (name, email, phone) VALUES (?, ?, ?)")
      .run(name, email || "", phone || "");
    patient = { id: result.lastInsertRowid };
  }

  // Create appointment
  const result = db
    .prepare(
      "INSERT INTO appointments (patient_id, dentist, service, date, time, status, notes) VALUES (?, ?, ?, ?, ?, 'confirmed', ?)"
    )
    .run(patient.id, dentist, service, date, time, notes || "");

  const appointment = db.prepare("SELECT * FROM appointments WHERE id = ?").get(result.lastInsertRowid);

  return NextResponse.json({ appointment }, { status: 201 });
}
