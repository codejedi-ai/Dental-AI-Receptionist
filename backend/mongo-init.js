// ─────────────────────────────────────────────────────────────
// MongoDB Init Script
// Creates collections and sets up initial document structure
// ─────────────────────────────────────────────────────────────

const dbName = process.env.MONGO_INITDB_DATABASE || "dental";
db = db.getSiblingDB(dbName);

// ── Email Verifications ──
db.createCollection("verifications");

db.verifications.createIndex({ "email": 1 }, { unique: false });
db.verifications.createIndex({ "token": 1 }, { unique: true });
db.verifications.createIndex({ "expiresAt": 1 }, { expireAfterSeconds: 0 });

print("✅ MongoDB: verifications collection created");

// ── Clinic Settings (document store) ──
db.createCollection("settings");

db.settings.insertOne({
  _id: "clinic",
  name: "Dental Clinic",
  phone: "(555) 555-0123",
  address: "123 Main Street",
  city: "City",
  province: "Province",
  postalCode: "A1A 1A1",
  hours: "Monday through Friday 8 AM to 6 PM, Saturday 9 AM to 2 PM. Closed on Sundays and statutory holidays.",
  insurance: "We accept most major dental insurance plans. Please bring your insurance card to your appointment.",
  emergency: "For after-hours dental emergencies, please go to your nearest emergency department.",
  cancellationPolicy: "We ask for at least 24 hours notice for cancellations. Late cancellations or no-shows may be subject to a fee.",
  newPatientInfo: "New patients are always welcome! For your first visit, please arrive 15 minutes early to complete paperwork. Bring your insurance card, a list of current medications, and any recent dental X-rays if you have them.",
  payment: "We accept cash, debit, Visa, Mastercard, and e-transfer. We also offer payment plans for larger treatments.",
  updatedAt: new Date()
});

print("✅ MongoDB: settings collection created with defaults");

// ── Call Logs (for Vapi call analytics) ──
db.createCollection("call_logs");

db.call_logs.createIndex({ "startedAt": -1 });
db.call_logs.createIndex({ "status": 1 });
db.call_logs.createIndex({ "patientPhone": 1 });

print("✅ MongoDB: call_logs collection created");

// ── Tool Call Audit ──
db.createCollection("tool_calls");

db.tool_calls.createIndex({ "timestamp": -1 });
db.tool_calls.createIndex({ "toolName": 1 });
db.tool_calls.createIndex({ "status": 1 });

print("✅ MongoDB: tool_calls collection created");

// ── Verification Seeds (default clinic info that Vapi assistant can query) ──
db.createCollection("clinic_info");

db.clinic_info.insertMany([
  { topic: "hours", content: db.settings.findOne({ _id: "clinic" }).hours },
  { topic: "location", content: db.settings.findOne({ _id: "clinic" }).address + ", " + db.settings.findOne({ _id: "clinic" }).city + ", " + db.settings.findOne({ _id: "clinic" }).province + " " + db.settings.findOne({ _id: "clinic" }).postalCode },
  { topic: "insurance", content: db.settings.findOne({ _id: "clinic" }).insurance },
  { topic: "emergency", content: db.settings.findOne({ _id: "clinic" }).emergency },
  { topic: "cancellation", content: db.settings.findOne({ _id: "clinic" }).cancellationPolicy },
  { topic: "newpatient", content: db.settings.findOne({ _id: "clinic" }).newPatientInfo },
  { topic: "payment", content: db.settings.findOne({ _id: "clinic" }).payment }
]);

db.clinic_info.createIndex({ "topic": 1 }, { unique: true });

print("✅ MongoDB: clinic_info collection created");

print("🎉 MongoDB initialization complete!");
