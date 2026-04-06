import { createCalendarEvent, BookingDetails } from "../services/calendar";
import { sendConfirmationEmail } from "../services/email";
import { saveAppointment } from "../services/appointment-store";

interface BookingRequest {
  patientName: string;
  patientPhone: string;
  patientEmail?: string;
  date: string;
  time: string;
  dentist: string;
  service: string;
  notes?: string;
}

export async function bookAppointment(params: BookingRequest): Promise<string> {
  const { patientName, patientPhone, patientEmail, date, time, dentist, service, notes } = params;

  if (!patientName || !date || !time || !dentist) {
    return "I'm missing some details. Could you confirm the patient name, date, time, and dentist?";
  }

  // Create Google Calendar event
  const details: BookingDetails = { patientName, patientPhone, patientEmail, date, time, dentist, service, notes };
  const eventId = await createCalendarEvent(details);

  // Persist to local JSON store
  saveAppointment({ patientName, patientPhone, patientEmail, date, time, dentist, service, notes, googleEventId: eventId ?? undefined });

  // Send confirmation email (non-blocking — don't fail the booking if email fails)
  if (patientEmail) {
    sendConfirmationEmail(details).catch((err) =>
      console.error("Email send failed:", err)
    );
  }

  const calendarNote = eventId ? " It's on the clinic's calendar." : "";
  const emailNote = patientEmail ? ` A confirmation email will be sent to ${patientEmail}.` : ` We'll confirm by phone at ${patientPhone}.`;

  return `Booked! ${patientName} has a ${service} appointment with ${dentist} on ${date} at ${time}.${calendarNote}${emailNote} Please arrive 15 minutes early and bring your insurance card. Anything else?`;
}
