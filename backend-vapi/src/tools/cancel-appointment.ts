import { findEvent, deleteCalendarEvent } from "../services/calendar";
import { sendCancellationEmail } from "../services/email";
import { findAppointment, cancelAppointmentById } from "../services/appointment-store";

interface CancelRequest {
  patientName: string;
  patientPhone?: string;
  date: string;
}

export async function cancelAppointment(params: CancelRequest): Promise<string> {
  const { patientName, date } = params;

  // Look up in local store first (always available), then fall back to Google Calendar
  const localAppt = findAppointment(patientName, date);
  const calEvent = await findEvent(patientName, date);

  if (!localAppt && !calEvent) {
    return `I couldn't find an appointment for ${patientName} on ${date}. Could you double-check the name and date?`;
  }

  const service = localAppt?.service ?? calEvent?.summary?.split("—")[0]?.trim() ?? "appointment";
  const startTime = localAppt?.time ?? (calEvent?.start?.dateTime
    ? new Date(calEvent.start.dateTime).toLocaleTimeString("en-CA", { hour: "2-digit", minute: "2-digit" })
    : "");
  const patientEmail = localAppt?.patientEmail ?? calEvent?.attendees?.[0]?.email;

  // Cancel in local store
  if (localAppt) {
    cancelAppointmentById(localAppt.id);
  }

  // Delete from Google Calendar
  const eventId = localAppt?.googleEventId ?? calEvent?.id;
  if (eventId) {
    await deleteCalendarEvent(eventId).catch((err) =>
      console.error("Calendar delete failed:", err)
    );
  }

  // Email the patient
  if (patientEmail) {
    sendCancellationEmail({
      patientName,
      patientEmail,
      date,
      time: startTime,
      service,
    }).catch((err) => console.error("Cancellation email failed:", err));
  }

  return `Cancelled! ${patientName}'s ${service} on ${date}${startTime ? ` at ${startTime}` : ""} has been removed.${patientEmail ? " A cancellation email is on its way." : ""} Would you like to rebook?`;
}
