import { Router } from "express";
import { getDb, getAvailableSlots } from "../db";

const router = Router();

router.get("/available", (req, res) => {
  const { date, dentist } = req.query as { date: string; dentist: string };
  if (!date || !dentist) return res.status(400).json({ error: "date and dentist required" });
  res.json({ slots: getAvailableSlots(date, dentist), date, dentist });
});

router.get("/:id", (req, res) => {
  const db = getDb();
  const appt = db.prepare("SELECT a.*, p.name as patient_name, p.email as patient_email, p.phone as patient_phone FROM appointments a LEFT JOIN patients p ON a.patient_id = p.id WHERE a.id = ?").get(req.params.id);
  if (!appt) return res.status(404).json({ error: "Not found" });
  res.json({ appointment: appt });
});

router.patch("/:id", (req, res) => {
  const db = getDb();
  const { status } = req.body;
  if (status) db.prepare("UPDATE appointments SET status = ? WHERE id = ?").run(status, req.params.id);
  const appt = db.prepare("SELECT * FROM appointments WHERE id = ?").get(req.params.id);
  res.json({ appointment: appt });
});

router.get("/", (req, res) => {
  const db = getDb();
  const { status, date, search } = req.query as Record<string, string>;
  let q = "SELECT a.*, p.name as patient_name, p.email as patient_email, p.phone as patient_phone FROM appointments a LEFT JOIN patients p ON a.patient_id = p.id WHERE 1=1";
  const params: any[] = [];
  if (status && status !== "all") { q += " AND a.status = ?"; params.push(status); }
  if (date) { q += " AND a.date = ?"; params.push(date); }
  if (search) { const s = `%${search}%`; q += " AND (p.name LIKE ? OR a.service LIKE ? OR a.dentist LIKE ?)"; params.push(s, s, s); }
  q += " ORDER BY a.date DESC, a.time ASC";
  res.json({ appointments: db.prepare(q).all(...params) });
});

router.post("/", (req, res) => {
  const db = getDb();
  const { service, dentist, date, time, name, email, phone, notes } = req.body;
  if (!service || !dentist || !date || !time || !name) return res.status(400).json({ error: "Missing required fields" });

  let patient = db.prepare("SELECT * FROM patients WHERE email = ? OR (name = ? AND phone = ?)").get(email, name, phone) as any;
  if (!patient) {
    const r = db.prepare("INSERT INTO patients (name, email, phone) VALUES (?, ?, ?)").run(name, email || "", phone || "");
    patient = { id: r.lastInsertRowid };
  }
  const r = db.prepare("INSERT INTO appointments (patient_id, dentist, service, date, time, status, notes) VALUES (?, ?, ?, ?, ?, 'confirmed', ?)").run(patient.id, dentist, service, date, time, notes || "");
  const appointment = db.prepare("SELECT * FROM appointments WHERE id = ?").get(r.lastInsertRowid);
  res.status(201).json({ appointment });
});

export default router;
