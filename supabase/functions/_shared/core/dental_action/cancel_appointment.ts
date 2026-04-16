import { Tool } from "../ToolDefinition.ts";
import { asObj } from "../lib/dentalHelpers.ts";
import { restRequest } from "../lib/supabaseRest.ts";

export const cancel_appointment = new Tool(
  "cancel_appointment",
  "Cancel a confirmed appointment by patient name and appointment date.",
  "Use when the user asks to cancel. Resolves patient by fuzzy name match and date; confirms one row.",
  {
    type: "object",
    properties: {
      patientName: { type: "string", description: "Patient full name." },
      date: { type: "string", description: "Appointment date YYYY-MM-DD." },
    },
    required: ["patientName", "date"],
  },
  async (parameters) => {
    try {
      const a = asObj(parameters);
      const patientName = String(a.patientName ?? "").trim();
      const date = String(a.date ?? "").trim();
      if (!patientName) {
        return { result: "I need the patient name to cancel an appointment." };
      }
      if (!date) {
        return { result: "I need the appointment date (YYYY-MM-DD) to cancel." };
      }

      const patients = await restRequest<Array<{ id: number }>>(
        "GET",
        `patients?select=id&name=ilike.${encodeURIComponent(patientName)}&limit=5`,
      );
      if (!patients.length) {
        return { result: "I could not find a matching patient record." };
      }

      const ids = patients.map((p) => p.id);
      const target = await restRequest<Array<{ id: number }>>(
        "GET",
        `appointments?select=id&appointment_date=eq.${encodeURIComponent(date)}&status=eq.confirmed&patient_id=in.(${ids.join(",")})&limit=1`,
      );
      if (!target.length) {
        return {
          result: `No confirmed appointment found for ${patientName} on ${date}.`,
        };
      }

      await restRequest(
        "PATCH",
        `appointments?id=eq.${target[0].id}`,
        { status: "cancelled", cancelled_at: new Date().toISOString() },
      );

      return { result: `Appointment cancelled for ${patientName} on ${date}.` };
    } catch {
      return {
        result: "Error cancelling appointment. Please call the clinic directly.",
      };
    }
  },
);
