import { Tool } from "../ToolDefinition.ts";
import { asObj, parseDateText, torontoNow } from "../lib/dentalHelpers.ts";

export const parse_date = new Tool(
  "parse_date",
  "Convert natural-language or relative dates to YYYY-MM-DD (Toronto).",
  "Use before check_availability when the user says natural date text. If month/year are missing, default to current Toronto month/year.",
  {
    type: "object",
    properties: {
      date_text: {
        type: "string",
        description:
          "What the caller said (e.g. next Tuesday, tomorrow, 今天).",
      },
      dateText: {
        type: "string",
        description: "Alias for date_text.",
      },
      day: {
        type: "integer",
        description: "Day of month (1-31).",
      },
      month: {
        type: "string",
        description:
          "Month as number (1-12) or name (e.g. September, Sep). Defaults to current month when omitted.",
      },
      year: {
        type: "integer",
        description: "4-digit year. Defaults to current year when omitted.",
      },
    },
  },
  async (parameters) => {
    const a = asObj(parameters);
    const input = String(a.date_text ?? a.dateText ?? "");
    const now = torontoNow();
    const currentYear = now.getUTCFullYear();
    const currentMonth = now.getUTCMonth() + 1;

    const day = toInt(a.day);
    const month = resolveMonth(a.month);
    const year = toInt(a.year);

    if (day == null && (month != null || year != null)) {
      return {
        result:
          "Got it. What day of the month would you like for that appointment?",
      };
    }

    if (day != null) {
      const normalized = normalizeYmd(
        year ?? currentYear,
        month ?? currentMonth,
        day,
      );
      if (normalized) {
        return { result: normalized };
      }
      return {
        result:
          "I could not match that to a real calendar date. Could you confirm the day again?",
      };
    }

    if (isMonthOnlyText(input)) {
      return {
        result:
          "Sure. Which day in that month would you like for the appointment?",
      };
    }

    const parsed = parseDateText(input);
    if (!parsed) {
      return {
        result:
          "I could not understand that date yet. Please say it like 'May 6' or '2024-05-06'.",
      };
    }
    return { result: parsed };
  },
);

function toInt(v: unknown): number | null {
  if (typeof v === "number" && Number.isInteger(v)) return v;
  if (typeof v === "string" && v.trim() !== "") {
    const n = Number(v);
    if (Number.isInteger(n)) return n;
  }
  return null;
}

function resolveMonth(v: unknown): number | null {
  const n = toInt(v);
  if (n != null) return n >= 1 && n <= 12 ? n : null;
  if (typeof v !== "string") return null;
  const m = normalizeMonthTypos(v.trim().toLowerCase());
  const map: Record<string, number> = {
    jan: 1,
    january: 1,
    feb: 2,
    february: 2,
    mar: 3,
    march: 3,
    apr: 4,
    april: 4,
    may: 5,
    jun: 6,
    june: 6,
    jul: 7,
    july: 7,
    aug: 8,
    august: 8,
    sep: 9,
    sept: 9,
    september: 9,
    oct: 10,
    october: 10,
    nov: 11,
    november: 11,
    dec: 12,
    december: 12,
  };
  return map[m] ?? null;
}

function isMonthOnlyText(input: string): boolean {
  const s = normalizeMonthTypos(input.trim().toLowerCase());
  if (!s) return false;
  return /^(?:in\s+)?(jan(?:uary)?|feb(?:ruary)?|mar(?:ch)?|apr(?:il)?|may|jun(?:e)?|jul(?:y)?|aug(?:ust)?|sep(?:t)?(?:ember)?|oct(?:ober)?|nov(?:ember)?|dec(?:ember)?)$/
    .test(s);
}

function normalizeMonthTypos(input: string): string {
  return input
    .replace(/\bspetember\b/g, "september")
    .replace(/\bsetpember\b/g, "september");
}

function normalizeYmd(year: number, month: number, day: number): string | null {
  if (!Number.isInteger(year) || !Number.isInteger(month) || !Number.isInteger(day)) {
    return null;
  }
  if (month < 1 || month > 12 || day < 1 || day > 31) return null;
  const dt = new Date(Date.UTC(year, month - 1, day));
  if (
    dt.getUTCFullYear() !== year ||
    dt.getUTCMonth() + 1 !== month ||
    dt.getUTCDate() !== day
  ) {
    return null;
  }
  return dt.toISOString().slice(0, 10);
}
