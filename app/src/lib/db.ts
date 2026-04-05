import Database from "better-sqlite3";
import path from "path";

const DB_PATH = path.join(process.cwd(), "smile-dental.db");

let db: Database.Database | null = null;

export function getDb(): Database.Database {
  if (db) return db;

  db = new Database(DB_PATH);
  db.pragma("journal_mode = WAL");
  db.pragma("foreign_keys = ON");

  // Create tables
  db.exec(`
    CREATE TABLE IF NOT EXISTS patients (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      name TEXT NOT NULL,
      email TEXT,
      phone TEXT,
      created_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );

    CREATE TABLE IF NOT EXISTS dentists (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      name TEXT NOT NULL,
      specialty TEXT,
      bio TEXT,
      avatar_url TEXT DEFAULT ''
    );

    CREATE TABLE IF NOT EXISTS appointments (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      patient_id INTEGER,
      dentist TEXT NOT NULL,
      service TEXT NOT NULL,
      date TEXT NOT NULL,
      time TEXT NOT NULL,
      status TEXT DEFAULT 'confirmed',
      notes TEXT DEFAULT '',
      created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
      FOREIGN KEY (patient_id) REFERENCES patients(id)
    );
  `);

  // Seed dentists if empty
  const count = db.prepare("SELECT COUNT(*) as count FROM dentists").get() as { count: number };
  if (count.count === 0) {
    const insert = db.prepare(
      "INSERT INTO dentists (name, specialty, bio, avatar_url) VALUES (?, ?, ?, ?)"
    );
    insert.run(
      "Dr. Sarah Chen",
      "General & Cosmetic Dentistry",
      "Dr. Chen has over 15 years of experience in general and cosmetic dentistry. She is passionate about helping patients achieve their best smile through personalized treatment plans and gentle care.",
      ""
    );
    insert.run(
      "Dr. Michael Park",
      "Orthodontics & Invisalign",
      "Dr. Park is a board-certified orthodontist specializing in Invisalign and modern braces. He has transformed thousands of smiles and is known for his meticulous attention to detail.",
      ""
    );
    insert.run(
      "Dr. Priya Sharma",
      "Pediatric Dentistry & Root Canals",
      "Dr. Sharma specializes in pediatric dentistry and endodontics. Her gentle approach makes children feel at ease, and her expertise in root canals ensures painless procedures for adults.",
      ""
    );
  }

  return db;
}

export interface Patient {
  id: number;
  name: string;
  email: string;
  phone: string;
  created_at: string;
}

export interface Appointment {
  id: number;
  patient_id: number;
  dentist: string;
  service: string;
  date: string;
  time: string;
  status: string;
  notes: string;
  created_at: string;
}

export interface Dentist {
  id: number;
  name: string;
  specialty: string;
  bio: string;
  avatar_url: string;
}

// Generate available time slots for a given date and dentist
export function getAvailableSlots(date: string, dentist: string): string[] {
  const db = getDb();
  const dayOfWeek = new Date(date + "T12:00:00").getDay();

  // Sunday = closed
  if (dayOfWeek === 0) return [];

  // Saturday: 9am-1pm, Weekdays: 9am-5pm
  const isSaturday = dayOfWeek === 6;
  const startHour = 9;
  const endHour = isSaturday ? 13 : 17;

  const allSlots: string[] = [];
  for (let h = startHour; h < endHour; h++) {
    allSlots.push(`${h.toString().padStart(2, "0")}:00`);
    allSlots.push(`${h.toString().padStart(2, "0")}:30`);
  }

  // Remove already booked slots
  const booked = db
    .prepare(
      "SELECT time FROM appointments WHERE date = ? AND dentist = ? AND status != 'cancelled'"
    )
    .all(date, dentist) as { time: string }[];

  const bookedTimes = new Set(booked.map((b) => b.time));
  return allSlots.filter((slot) => !bookedTimes.has(slot));
}
