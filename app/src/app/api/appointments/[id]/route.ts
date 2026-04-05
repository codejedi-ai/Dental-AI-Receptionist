import { NextRequest, NextResponse } from "next/server";
import { getDb } from "@/lib/db";

export async function GET(req: NextRequest, { params }: { params: { id: string } }) {
  const db = getDb();
  const appointment = db
    .prepare(
      `SELECT a.*, p.name as patient_name, p.email as patient_email, p.phone as patient_phone
       FROM appointments a
       LEFT JOIN patients p ON a.patient_id = p.id
       WHERE a.id = ?`
    )
    .get(params.id);

  if (!appointment) {
    return NextResponse.json({ error: "Appointment not found" }, { status: 404 });
  }

  return NextResponse.json({ appointment });
}

export async function PATCH(req: NextRequest, { params }: { params: { id: string } }) {
  const db = getDb();
  const body = await req.json();
  const { status, notes } = body;

  const existing = db.prepare("SELECT * FROM appointments WHERE id = ?").get(params.id);
  if (!existing) {
    return NextResponse.json({ error: "Appointment not found" }, { status: 404 });
  }

  if (status) {
    db.prepare("UPDATE appointments SET status = ? WHERE id = ?").run(status, params.id);
  }
  if (notes !== undefined) {
    db.prepare("UPDATE appointments SET notes = ? WHERE id = ?").run(notes, params.id);
  }

  const updated = db.prepare("SELECT * FROM appointments WHERE id = ?").get(params.id);
  return NextResponse.json({ appointment: updated });
}
