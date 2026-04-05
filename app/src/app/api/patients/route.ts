import { NextRequest, NextResponse } from "next/server";
import { getDb } from "@/lib/db";

export async function GET(req: NextRequest) {
  const db = getDb();
  const url = new URL(req.url);
  const search = url.searchParams.get("search");

  let query = "SELECT * FROM patients WHERE 1=1";
  const params: any[] = [];

  if (search) {
    query += " AND (name LIKE ? OR email LIKE ? OR phone LIKE ?)";
    const s = `%${search}%`;
    params.push(s, s, s);
  }

  query += " ORDER BY created_at DESC";

  const patients = db.prepare(query).all(...params);
  return NextResponse.json({ patients });
}

export async function POST(req: NextRequest) {
  const db = getDb();
  const body = await req.json();
  const { name, email, phone } = body;

  if (!name) {
    return NextResponse.json({ error: "Name is required" }, { status: 400 });
  }

  const result = db
    .prepare("INSERT INTO patients (name, email, phone) VALUES (?, ?, ?)")
    .run(name, email || "", phone || "");

  const patient = db.prepare("SELECT * FROM patients WHERE id = ?").get(result.lastInsertRowid);
  return NextResponse.json({ patient }, { status: 201 });
}
