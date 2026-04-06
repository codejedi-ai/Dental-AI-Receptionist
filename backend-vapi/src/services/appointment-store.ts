/**
 * Local JSON appointment store.
 *
 * Appointments are persisted in data/appointments/ — one JSON file per appointment.
 * File names: {YYYY-MM-DD}_{HH-MM}_{sanitised-patient-name}_{id}.json
 *
 * This is the source of truth for availability conflict checking when
 * Google Calendar is not configured.
 */

import fs from "fs";
import path from "path";
import crypto from "crypto";

const DATA_DIR = path.resolve(process.cwd(), "data", "appointments");

// Ensure the directory exists on startup
if (!fs.existsSync(DATA_DIR)) {
  fs.mkdirSync(DATA_DIR, { recursive: true });
}

export interface Appointment {
  id: string;
  patientName: string;
  patientPhone: string;
  patientEmail?: string;
  date: string;        // YYYY-MM-DD
  time: string;        // HH:MM (24h)
  dentist: string;
  service: string;
  notes?: string;
  status: "confirmed" | "cancelled";
  googleEventId?: string;
  createdAt: string;   // ISO
  cancelledAt?: string;
}

function sanitise(s: string): string {
  return s.replace(/[^a-zA-Z0-9]/g, "-").toLowerCase().slice(0, 20);
}

function filePath(appt: Appointment): string {
  const timeSlug = appt.time.replace(":", "-");
  const nameSlug = sanitise(appt.patientName);
  return path.join(DATA_DIR, `${appt.date}_${timeSlug}_${nameSlug}_${appt.id}.json`);
}

// ─────────────────────────────────────────────────────────────────
// Save a new appointment
// ─────────────────────────────────────────────────────────────────
export function saveAppointment(data: Omit<Appointment, "id" | "status" | "createdAt">): Appointment {
  const appt: Appointment = {
    ...data,
    id: crypto.randomBytes(6).toString("hex"),
    status: "confirmed",
    createdAt: new Date().toISOString(),
  };
  fs.writeFileSync(filePath(appt), JSON.stringify(appt, null, 2), "utf-8");
  console.log(`💾 Appointment saved: ${appt.id} (${appt.patientName} ${appt.date} ${appt.time})`);
  return appt;
}

// ─────────────────────────────────────────────────────────────────
// Load all appointments (optionally filter by date)
// ─────────────────────────────────────────────────────────────────
export function listAppointments(date?: string): Appointment[] {
  const files = fs.readdirSync(DATA_DIR).filter(f => f.endsWith(".json"));
  const results: Appointment[] = [];

  for (const file of files) {
    if (date && !file.startsWith(date)) continue;
    try {
      const raw = fs.readFileSync(path.join(DATA_DIR, file), "utf-8");
      results.push(JSON.parse(raw) as Appointment);
    } catch {
      // Corrupted file — skip
    }
  }

  return results.sort((a, b) => a.time.localeCompare(b.time));
}

// ─────────────────────────────────────────────────────────────────
// Find a single appointment by patient name + date
// ─────────────────────────────────────────────────────────────────
export function findAppointment(patientName: string, date: string): Appointment | null {
  const name = patientName.toLowerCase();
  const all = listAppointments(date);
  return all.find(a => a.status === "confirmed" && a.patientName.toLowerCase().includes(name)) ?? null;
}

// ─────────────────────────────────────────────────────────────────
// Cancel (marks as cancelled, rewrites file in-place)
// ─────────────────────────────────────────────────────────────────
export function cancelAppointmentById(id: string): Appointment | null {
  const files = fs.readdirSync(DATA_DIR).filter(f => f.includes(id) && f.endsWith(".json"));
  if (!files[0]) return null;

  const fp = path.join(DATA_DIR, files[0]);
  const appt: Appointment = JSON.parse(fs.readFileSync(fp, "utf-8"));
  appt.status = "cancelled";
  appt.cancelledAt = new Date().toISOString();
  fs.writeFileSync(fp, JSON.stringify(appt, null, 2), "utf-8");
  console.log(`💾 Appointment cancelled: ${appt.id}`);
  return appt;
}

// ─────────────────────────────────────────────────────────────────
// Check if a slot is already taken (for availability cross-check)
// ─────────────────────────────────────────────────────────────────
export function isSlotTaken(date: string, time: string, dentist: string): boolean {
  return listAppointments(date).some(
    a => a.status === "confirmed" && a.time === time && a.dentist === dentist
  );
}
