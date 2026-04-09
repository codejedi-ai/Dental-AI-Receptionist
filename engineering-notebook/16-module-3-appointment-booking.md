# Page 16 — Module 3: Appointment Booking

---

## 16.1 Overview

Module 3 handles the **appointment booking conversation flow**. It receives classified intent from **Module 2** and guides the patient through collecting all required fields before booking.

> **Status:** Placeholder — the Go implementation exists in `vapi-backend/internal/modules/appointmentbooking/booking.go`. See [Page 15 — Evaluation](15-evaluation-technical-challenges.md) for known issues and fixes.

---

## 16.2 Booking State Machine

```
collecting_service → collecting_dentist → collecting_date → selecting_time → collecting_patient_info → confirming_details → complete
```

### Required Fields

| Field | Required | Notes |
|-------|----------|-------|
| `service` | ✅ Yes | e.g., Cleaning, Crown, Bridge, Consultation, Root Canal |
| `dentist` | ✅ Yes | Dr. Sarah Chen, Dr. Michael Park, Dr. Priya Sharma, or "any" |
| `date` | ✅ Yes | YYYY-MM-DD format |
| `time` | ✅ Yes | HH:MM format |
| `patient_name` | ✅ Yes | Full name |
| `patient_phone` | ✅ Yes | Mobile number (cell only) |
| `email` | ❌ No | Only as fallback if no mobile number |
| `address` | ❌ No | Never collected |
| `date_of_birth` | ❌ No | Never collected |

---

## 16.3 Tools

| Tool | Purpose |
|------|---------|
| `get_booking_step` | Determine next field to collect |
| `fill_booking_fields` | Extract fields from patient utterance |
| `is_booking_complete` | Check if all required fields collected |
| `get_confirm_message` | Generate confirmation read-back |
| `check_availability` | Query PostgreSQL for open slots |
| `book_appointment` | Save appointment to database |
| `send_booking_confirmation` | Send SMS confirmation |

---

*Previous: [Page 15 — Evaluation: Technical Challenges](15-evaluation-technical-challenges.md)*
