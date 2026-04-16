import { Tool } from "../ToolDefinition.ts";
import { asObj, getDentistsRaw, toAmPm } from "../lib/dentalHelpers.ts";
import { restRequest } from "../lib/supabaseRest.ts";

export const book_appointment = new Tool(
  "book_appointment",
  "Create a confirmed appointment after availability was verified and patient details collected.",
  "Requires patient name, date, time, dentist, service. Phone or email fallback per clinic policy. Never call without a prior successful check_availability for that slot.",
  {
    type: "object",
    properties: {
      patientName: { type: "string", description: "Patient full name." },
      patientPhone: { type: "string", description: "Cell phone; optional if placeholder." },
      patientEmail: { type: "string", description: "Used when no mobile." },
      date: { type: "string", description: "YYYY-MM-DD." },
      time: { type: "string", description: "HH:MM 24h slot string matching availability." },
      dentist: { type: "string", description: "Dentist full name as in directory." },
      service: { type: "string", description: "Service type (e.g. Cleaning)." },
      notes: { type: "string", description: "Optional notes." },
    },
    required: ["patientName", "date", "time", "dentist", "service"],
  },
  async (parameters) => {
    try {
      const a = asObj(parameters);
      const patientName = String(a.patientName ?? "").trim();
      const patientPhone = String(a.patientPhone ?? "").trim();
      const patientEmail = String(a.patientEmail ?? "").trim();
      const date = String(a.date ?? "").trim();
      const time = String(a.time ?? "").trim();
      const dentist = String(a.dentist ?? "").trim();
      const service = String(a.service ?? "").trim();
      const notes = String(a.notes ?? "").trim();

      if (!patientName || !date || !time || !dentist || !service) {
        return {
          result:
            "I need the patient name, date, time, dentist, and service to book. Please provide all required details.",
        };
      }

      const dentists = await getDentistsRaw();
      const match = dentists.find((d) => d.name.toLowerCase() === dentist.toLowerCase());
      if (!match) {
        return { result: `I could not find dentist: ${dentist}` };
      }

      const occupied = await restRequest<Array<{ id: number }>>(
        "GET",
        `appointments?select=id&appointment_date=eq.${encodeURIComponent(date)}&appointment_time=eq.${encodeURIComponent(time)}&dentist_id=eq.${match.id}&status=eq.confirmed&limit=1`,
      );
      if (occupied.length > 0) {
        return {
          result:
            `The slot ${date} at ${time} with ${dentist} is already booked. Please choose another time.`,
        };
      }

      const fallbackPhone = patientPhone ||
        `NO-MOBILE-${crypto.randomUUID().slice(0, 8)}`;

      const patientRows = await restRequest<Array<{ id: number; name: string }>>(
        "POST",
        "patients?on_conflict=name,mobile",
        [{
          name: patientName,
          mobile: fallbackPhone,
          email: patientEmail || null,
        }],
        { "Prefer": "resolution=merge-duplicates,return=representation" },
      );

      if (!patientRows.length) {
        return { result: "Error registering patient. Please try again." };
      }

      const appointmentRows = await restRequest<Array<{ id: number }>>(
        "POST",
        "appointments",
        [{
          patient_id: patientRows[0].id,
          dentist_id: match.id,
          service,
          appointment_date: date,
          appointment_time: time,
          status: "confirmed",
          notes: notes || null,
        }],
        { "Prefer": "return=representation" },
      );

      const appointmentId = appointmentRows[0]?.id;
      return {
        result:
          `Appointment booked successfully! Confirmation ID: ${appointmentId}. ${patientName} on ${date} at ${toAmPm(time)} with ${match.name} for ${service}.`,
      };
    } catch {
      return { result: "Error booking appointment. Please try again." };
    }
  },
);
