import { Router } from "express";
import { getDb } from "../db";

const router = Router();

router.get("/", (_req, res) => {
  const db = getDb();
  res.json({ patients: db.prepare("SELECT * FROM patients ORDER BY created_at DESC").all() });
});

router.post("/", (req, res) => {
  const db = getDb();
  const { name, email, phone } = req.body;
  if (!name) return res.status(400).json({ error: "name required" });
  const r = db.prepare("INSERT INTO patients (name, email, phone) VALUES (?, ?, ?)").run(name, email || "", phone || "");
  res.status(201).json({ patient: db.prepare("SELECT * FROM patients WHERE id = ?").get(r.lastInsertRowid) });
});

export default router;
