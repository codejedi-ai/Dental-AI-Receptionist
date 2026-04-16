import { Tool } from "../ToolDefinition.ts";
import {
  asObj,
  getBookedTimes,
  getDentistsRaw,
  SATURDAY_SLOTS,
  torontoNow,
  WEEKDAY_SLOTS,
} from "../lib/dentalHelpers.ts";

export const get_next_available_dates = new Tool(
  "get_next_available_dates",
  "Suggest up to five upcoming calendar dates that have at least one open slot (first dentist probe).",
  "Use when the user asks for earliest availability or next openings. Skips Sundays; respects clinic slot templates.",
  {
    type: "object",
    properties: {
      days: {
        type: "number",
        description:
          "Optional lookahead window in days (default 14, max 60).",
      },
    },
  },
  async (parameters) => {
    try {
      const a = asObj(parameters);
      const days = Number(a.days ?? 14);
      const lookAhead = Number.isFinite(days) && days > 0 ? Math.min(days, 60) : 14;

      const dentists = await getDentistsRaw();
      if (!dentists.length) {
        return { result: "I am having trouble checking availability." };
      }

      const firstDentist = dentists[0];
      const today = torontoNow();
      const openDates: string[] = [];

      for (let i = 1; i <= lookAhead && openDates.length < 5; i++) {
        const d = new Date(today);
        d.setUTCDate(today.getUTCDate() + i);
        const date = d.toISOString().slice(0, 10);
        const dow = d.getUTCDay();
        if (dow === 0) continue;
        const slots = dow === 6 ? SATURDAY_SLOTS : WEEKDAY_SLOTS;
        const booked = await getBookedTimes(date, firstDentist.id);
        if (booked.size < slots.length) {
          openDates.push(date);
        }
      }

      if (!openDates.length) {
        return {
          result: "No available dates found in the next two weeks.",
        };
      }
      return {
        result: `Next available dates: ${openDates.join(", ")}.`,
      };
    } catch {
      return { result: "I am having trouble checking availability." };
    }
  },
);
