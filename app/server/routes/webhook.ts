import { Router } from "express";
import { getDb, getAvailableSlots, formatTime } from "../db";

const router = Router();

const DENTISTS = ["Dr. Sarah Chen", "Dr. Michael Park", "Dr. Priya Sharma"];

router.post("/", (req, res) => {
  const { message } = req.body;
  if (!message || message.type !== "function-call") return res.status(400).json({ error: "Not a function call" });

  const { name, parameters } = message.functionCall;

  try {
    switch (name) {
      case "check_availability": {
        const dentist = parameters.dentist || DENTISTS[0];
        const slots = getAvailableSlots(parameters.date, dentist);
        if (slots.length === 0) return res.json({ results: [{ result: `No available slots on ${parameters.date} with ${dentist}. Try another date?` }] });
        const list = slots.slice(0, 6).map(formatTime).join(", ");
        return res.json({ results: [{ result: `Available on ${parameters.date} with ${dentist}: ${list}${slots.length > 6 ? ` (+${slots.length - 6} more)` : ""}. Which time?` }] });
      }
      case "book_appointment": {
        const db = getDb();
        const { date, time, dentist, service, patient_name, patient_phone, patient_email, notes } = parameters;
        let patient: any = patient_email ? db.prepare("SELECT * FROM patients WHERE email = ?").get(patient_email) : null;
        if (!patient && patient_name) patient = db.prepare("SELECT * FROM patients WHERE name = ? AND phone = ?").get(patient_name, patient_phone || "");
        if (!patient) {
          const r = db.prepare("INSERT INTO patients (name, email, phone) VALUES (?, ?, ?)").run(patient_name || "Voice Patient", patient_email || "", patient_phone || "");
          patient = { id: r.lastInsertRowid };
        }
        const d = dentist || DENTISTS[0], s = service || "Checkup";
        const avail = getAvailableSlots(date, d);
        if (!avail.includes(time)) return res.json({ results: [{ result: `Sorry, ${formatTime(time)} on ${date} is taken. Pick another time?` }] });
        const r = db.prepare("INSERT INTO appointments (patient_id, dentist, service, date, time, status, notes) VALUES (?, ?, ?, ?, ?, 'confirmed', ?)").run(patient.id, d, s, date, time, notes || "Booked via voice");
        return res.json({ results: [{ result: `Booked! ${s} with ${d} on ${date} at ${formatTime(time)}. Appointment #${r.lastInsertRowid}. Anything else?` }] });
      }
      case "cancel_appointment": {
        const db = getDb();
        const { appointment_id, patient_name } = parameters;
        let appt: any = appointment_id ? db.prepare("SELECT * FROM appointments WHERE id = ? AND status = 'confirmed'").get(appointment_id) : null;
        if (!appt && patient_name) appt = db.prepare("SELECT a.* FROM appointments a JOIN patients p ON a.patient_id = p.id WHERE p.name LIKE ? AND a.status = 'confirmed' ORDER BY a.date ASC LIMIT 1").get(`%${patient_name}%`);
        if (!appt) return res.json({ results: [{ result: "Couldn't find that appointment. Can you give me the appointment number or full name?" }] });
        db.prepare("UPDATE appointments SET status = 'cancelled' WHERE id = ?").run(appt.id);
        return res.json({ results: [{ result: `Cancelled appointment #${appt.id} on ${appt.date} at ${formatTime(appt.time)}. Want to rebook?` }] });
      }
      case "get_clinic_info":
        return res.json({ results: [{ result: "Smile Dental Clinic, 123 Main St, Newmarket ON. Phone: (905) 555-0123. Mon-Fri 8am-6pm, Sat 9am-2pm, Sun closed." }] });
      case "get_services":
        return res.json({ results: [{ result: "We offer: Checkups, Cleanings, Whitening, Fillings, Crowns, Root Canals, Implants, Invisalign, Pediatric, and Emergency care. Which are you interested in?" }] });
      default:
        return res.json({ results: [{ result: "I can help you book an appointment or check availability. What would you like?" }] });
    }
  } catch (err) {
    console.error("Webhook error:", err);
    return res.json({ results: [{ result: "Something went wrong. Please try again or call us directly." }] });
  }
});

export default router;
