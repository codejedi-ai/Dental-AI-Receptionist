import { Tool } from "../ToolDefinition.ts";
import {
  asObj,
  getBookedTimes,
  getDentistsRaw,
  SATURDAY_SLOTS,
  torontoNow,
  toAmPm,
  WEEKDAY_SLOTS,
  weekdayFromDate,
} from "../lib/dentalHelpers.ts";

export const check_availability = new Tool(
  "check_availability",
  "Return available appointment slots for a dentist on a given date (YYYY-MM-DD).",
  "Must call before offering concrete times. Reject Sundays; Saturday uses shorter hours. Optionally filter by dentist name.",
  {
    type: "object",
    properties: {
      date: {
        type: "string",
        description: "Appointment date in YYYY-MM-DD (also accepts appointmentDate).",
      },
      dentist: {
        type: "string",
        description: "Optional dentist full name to filter slots.",
      },
    },
    required: ["date"],
  },
  async (parameters) => {
    try {
      const a = asObj(parameters);
      const date = String(a.date ?? a.appointmentDate ?? "").trim();
      const dentistName = String(a.dentist ?? "").trim();
      if (!date) {
        return {
          result: "I need a date in YYYY-MM-DD format to check availability.",
        };
      }
      if (!/^\d{4}-\d{2}-\d{2}$/.test(date)) {
        return {
          result: `Invalid date format: ${date}. Please use YYYY-MM-DD.`,
        };
      }

      const today = torontoNow().toISOString().slice(0, 10);
      if (date < today) {
        return {
          result: "That date is in the past. Please provide a future date.",
        };
      }

      const dow = weekdayFromDate(date);
      if (dow === 0) {
        return {
          result:
            "Sorry, the clinic is closed on Sundays. We are open Monday to Friday 8 AM to 6 PM and Saturday 9 AM to 2 PM.",
        };
      }
      const slots = dow === 6 ? SATURDAY_SLOTS : WEEKDAY_SLOTS;
      const dentists = await getDentistsRaw();
      const targetDentists = dentistName
        ? dentists.filter((d) => d.name.toLowerCase() === dentistName.toLowerCase())
        : dentists;

      if (!targetDentists.length) {
        return { result: `I could not find dentist: ${dentistName}` };
      }

      const available: Array<{ dentist: string; time: string }> = [];
      for (const dentist of targetDentists) {
        const booked = await getBookedTimes(date, dentist.id);
        for (const slot of slots) {
          if (!booked.has(slot)) {
            available.push({ dentist: dentist.name, time: slot });
          }
        }
      }

      if (!available.length) {
        const msg = dentistName
          ? `No available slots on ${date} with ${dentistName}.`
          : `No available slots on ${date}.`;
        return { result: `${msg} Would you like to try another date?` };
      }

      const sample = available.slice(0, 5).map((s) =>
        `${toAmPm(s.time)} with ${s.dentist}`
      ).join(", ");
      let result = `Available slots on ${date}: ${sample}.`;
      if (available.length > 5) {
        result += ` (${available.length - 5} more available)`;
      }
      return { result };
    } catch {
      return {
        result: "Sorry, I am having trouble checking availability right now.",
      };
    }
  },
);
