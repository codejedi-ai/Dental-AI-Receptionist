import nodemailer from "nodemailer";
import dotenv from "dotenv";
dotenv.config();

const transporter = nodemailer.createTransport({
  host: "smtp.gmail.com",
  port: 465,
  secure: true,
  auth: {
    user: process.env.SMTP_EMAIL,
    pass: process.env.SMTP_PASSWORD,
  },
});

export interface AppointmentEmailData {
  patientName: string;
  patientEmail?: string;
  patientPhone: string;
  date: string;       // YYYY-MM-DD
  time: string;       // HH:MM  (24h, Toronto time)
  dentist: string;
  service: string;
  notes?: string;
  eventId?: string;   // Google Calendar event ID (optional)
}

// ─────────────────────────────────────────────────────────────────
// Generate .ics calendar invite content
// ─────────────────────────────────────────────────────────────────
function buildICS(data: AppointmentEmailData): string {
  // Convert Toronto time → UTC for the .ics
  const localStart = new Date(`${data.date}T${data.time}:00-05:00`);
  const localEnd   = new Date(localStart.getTime() + 30 * 60_000); // 30 min slot

  const fmt = (d: Date) =>
    d.toISOString().replace(/[-:]/g, "").replace(/\.\d{3}/, "");

  const uid = `${data.date}-${data.time}-${data.patientName.replace(/\s/g, "")}@smiledental.ca`;

  return [
    "BEGIN:VCALENDAR",
    "VERSION:2.0",
    "PRODID:-//Smile Dental Clinic//EN",
    "METHOD:REQUEST",
    "BEGIN:VEVENT",
    `UID:${uid}`,
    `DTSTAMP:${fmt(new Date())}`,
    `DTSTART:${fmt(localStart)}`,
    `DTEND:${fmt(localEnd)}`,
    `SUMMARY:${data.service} — Smile Dental Clinic`,
    `DESCRIPTION:Patient: ${data.patientName}\\nDentist: ${data.dentist}\\nService: ${data.service}${data.notes ? `\\nNotes: ${data.notes}` : ""}`,
    "LOCATION:123 Main Street\\, Newmarket\\, ON L3Y 4Z1",
    `ORGANIZER;CN="Smile Dental Clinic":mailto:${process.env.SMTP_EMAIL}`,
    `ATTENDEE;CN="${data.patientName}";RSVP=TRUE:mailto:${data.patientEmail}`,
    "STATUS:CONFIRMED",
    "SEQUENCE:0",
    "BEGIN:VALARM",
    "TRIGGER:-PT24H",
    "ACTION:EMAIL",
    "DESCRIPTION:Reminder: Dental appointment tomorrow",
    "END:VALARM",
    "BEGIN:VALARM",
    "TRIGGER:-PT1H",
    "ACTION:DISPLAY",
    "DESCRIPTION:Dental appointment in 1 hour",
    "END:VALARM",
    "END:VEVENT",
    "END:VCALENDAR",
  ].join("\r\n");
}

// ─────────────────────────────────────────────────────────────────
// Send booking confirmation + .ics invite
// ─────────────────────────────────────────────────────────────────
export async function sendConfirmationEmail(data: AppointmentEmailData): Promise<void> {
  if (!data.patientEmail) return;
  if (!process.env.SMTP_EMAIL || !process.env.SMTP_PASSWORD) {
    console.warn("⚠️  SMTP credentials not set — skipping confirmation email");
    return;
  }

  const icsContent = buildICS(data);

  await transporter.sendMail({
    from: `"Smile Dental Clinic" <${process.env.SMTP_EMAIL}>`,
    to: data.patientEmail,
    subject: `Appointment Confirmed — Smile Dental Clinic`,
    text: `
Hi ${data.patientName},

Your appointment has been confirmed!

  Date:    ${data.date}
  Time:    ${data.time}
  Dentist: ${data.dentist}
  Service: ${data.service}
${data.notes ? `  Notes:   ${data.notes}` : ""}

📍 Smile Dental Clinic
   123 Main Street, Newmarket, ON L3Y 4Z1
   (905) 555-0123

Please arrive 15 minutes early and bring your insurance card.
The calendar invite is attached — click it to add to your calendar.

— Smile Dental Clinic
    `.trim(),
    attachments: [
      {
        filename: "appointment.ics",
        content: Buffer.from(icsContent, "utf-8"),
        contentType: "text/calendar; method=REQUEST; charset=utf-8",
      },
    ],
  });

  console.log(`📧 Confirmation + .ics sent to ${data.patientEmail}`);
}

// ─────────────────────────────────────────────────────────────────
// Send email verification link
// ─────────────────────────────────────────────────────────────────
export async function sendVerificationEmail(email: string, verifyUrl: string): Promise<void> {
  if (!process.env.SMTP_EMAIL || !process.env.SMTP_PASSWORD) {
    console.warn("⚠️  SMTP credentials not set — skipping verification email");
    return;
  }

  await transporter.sendMail({
    from: `"Smile Dental Clinic" <${process.env.SMTP_EMAIL}>`,
    to: email,
    subject: `Verify your email — Smile Dental Clinic`,
    text: `
Hi there,

Please click the link below to verify your email address so we can confirm your appointment:

${verifyUrl}

This link expires in 10 minutes.

If you didn't request this, you can ignore this email.

— Smile Dental Clinic
    `.trim(),
    html: `
<!DOCTYPE html>
<html>
<body style="font-family:sans-serif;max-width:480px;margin:40px auto;color:#333">
  <h2 style="color:#1a6fc4">Smile Dental Clinic</h2>
  <p>Hi there,</p>
  <p>Please verify your email address to confirm your appointment booking:</p>
  <a href="${verifyUrl}" style="display:inline-block;padding:12px 24px;background:#1a6fc4;color:#fff;border-radius:6px;text-decoration:none;font-weight:bold;margin:16px 0">
    Verify My Email
  </a>
  <p style="color:#666;font-size:13px">This link expires in 10 minutes.<br>If you didn't request this, ignore this email.</p>
  <hr style="border:none;border-top:1px solid #eee;margin-top:32px">
  <p style="color:#999;font-size:12px">Smile Dental Clinic · 123 Main Street, Newmarket, ON L3Y 4Z1</p>
</body>
</html>
    `.trim(),
  });

  console.log(`📧 Verification email sent to ${email}`);
}

// ─────────────────────────────────────────────────────────────────
// Send cancellation email
// ─────────────────────────────────────────────────────────────────
export async function sendCancellationEmail(data: Partial<AppointmentEmailData>): Promise<void> {
  if (!data.patientEmail) return;
  if (!process.env.SMTP_EMAIL || !process.env.SMTP_PASSWORD) return;

  await transporter.sendMail({
    from: `"Smile Dental Clinic" <${process.env.SMTP_EMAIL}>`,
    to: data.patientEmail,
    subject: `Appointment Cancelled — Smile Dental Clinic`,
    text: `
Hi ${data.patientName},

Your ${data.service} appointment with ${data.dentist} on ${data.date} at ${data.time} has been cancelled.

To rebook, call us at (905) 555-0123 or reply to this email.

— Smile Dental Clinic
    `.trim(),
  });

  console.log(`📧 Cancellation email sent to ${data.patientEmail}`);
}
