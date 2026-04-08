import Database from "better-sqlite3";
import path from "path";

const DB_PATH = path.join(process.cwd(), "smile-dental.db");
let db: Database.Database | null = null;

export function getDb(): Database.Database {
  if (db) return db;
  db = new Database(DB_PATH);
  db.pragma("journal_mode = WAL");
  db.pragma("foreign_keys = ON");

  db.exec(`
    CREATE TABLE IF NOT EXISTS patients (
      id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL, email TEXT, phone TEXT, created_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );
    CREATE TABLE IF NOT EXISTS dentists (
      id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL, specialty TEXT, bio TEXT, avatar_url TEXT DEFAULT ''
    );
    CREATE TABLE IF NOT EXISTS appointments (
      id INTEGER PRIMARY KEY AUTOINCREMENT, patient_id INTEGER, dentist TEXT NOT NULL, service TEXT NOT NULL,
      date TEXT NOT NULL, time TEXT NOT NULL, status TEXT DEFAULT 'confirmed', notes TEXT DEFAULT '',
      created_at DATETIME DEFAULT CURRENT_TIMESTAMP, FOREIGN KEY (patient_id) REFERENCES patients(id)
    );
  `);

  const count = db.prepare("SELECT COUNT(*) as count FROM dentists").get() as { count: number };
  if (count.count === 0) {
    const ins = db.prepare("INSERT INTO dentists (name, specialty, bio) VALUES (?, ?, ?)");
    ins.run("Dr. Sarah Chen", "General & Cosmetic Dentistry", "Dr. Chen has over 15 years of experience in general and cosmetic dentistry.");
    ins.run("Dr. Michael Park", "Orthodontics & Invisalign", "Dr. Park is a board-certified orthodontist specializing in Invisalign and modern braces.");
    ins.run("Dr. Priya Sharma", "Pediatric Dentistry & Root Canals", "Dr. Sharma specializes in pediatric dentistry and endodontics.");
  }
  return db;
}

export function getAvailableSlots(date: string, dentist: string): string[] {
  const d = getDb();
  const dow = new Date(date + "T12:00:00").getDay();
  if (dow === 0) return [];
  const isSat = dow === 6;
  const slots: string[] = [];
  for (let h = 9; h < (isSat ? 13 : 17); h++) {
    slots.push(`${h.toString().padStart(2, "0")}:00`);
    slots.push(`${h.toString().padStart(2, "0")}:30`);
  }
  const booked = d.prepare("SELECT time FROM appointments WHERE date = ? AND dentist = ? AND status != 'cancelled'").all(date, dentist) as { time: string }[];
  const taken = new Set(booked.map((b) => b.time));
  return slots.filter((s) => !taken.has(s));
}

export function formatTime(time: string): string {
  const [h, m] = time.split(":").map(Number);
  return `${h % 12 || 12}:${m.toString().padStart(2, "0")} ${h >= 12 ? "PM" : "AM"}`;
}
