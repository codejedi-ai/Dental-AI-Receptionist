import { restRequest } from "./supabaseRest.ts";

export type JsonRecord = Record<string, unknown>;

export const WEEKDAY_SLOTS = [
  "08:00", "08:30", "09:00", "09:30", "10:00", "10:30", "11:00", "11:30",
  "12:00", "12:30", "13:00", "13:30", "14:00", "14:30", "15:00", "15:30",
  "16:00", "16:30", "17:00", "17:30",
];

export const SATURDAY_SLOTS = [
  "09:00", "09:30", "10:00", "10:30", "11:00", "11:30",
  "12:00", "12:30", "13:00", "13:30",
];

export const CLINIC_INFO: Record<string, string> = {
  general:
    "Smile Dental Clinic is at 123 Main Street, Newmarket, ON L3Y 4Z1. Phone: (905) 555-0123. We are open Monday to Friday 8 AM to 6 PM, Saturday 9 AM to 2 PM, and closed Sunday.",
  hours:
    "Our hours are Monday to Friday 8 AM to 6 PM, Saturday 9 AM to 2 PM, and closed Sunday.",
  services:
    "We offer Consultation, Cleaning, Filling, Bridge, Crown, Root Canal, Extraction, Whitening, Implant, Invisalign, Pediatric, and Emergency dental care.",
  location:
    "Smile Dental Clinic is located at 123 Main Street, Newmarket, ON L3Y 4Z1.",
  insurance:
    "We accept most major dental insurance plans. Please bring your insurance card to your appointment.",
  emergency:
    "For severe pain, heavy bleeding, or knocked-out tooth, please seek immediate emergency care and call us for same-day triage when available.",
};

export function asObj(input: unknown): JsonRecord {
  if (!input) return {};
  if (typeof input === "string") {
    try {
      return JSON.parse(input) as JsonRecord;
    } catch {
      return {};
    }
  }
  if (typeof input === "object") return input as JsonRecord;
  return {};
}

export function torontoNow(): Date {
  const now = new Date();
  const fmt = new Intl.DateTimeFormat("en-CA", {
    timeZone: "America/Toronto",
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
  });
  const parts = fmt.formatToParts(now);
  const year = Number(parts.find((p) => p.type === "year")?.value ?? "0");
  const month = Number(parts.find((p) => p.type === "month")?.value ?? "1");
  const day = Number(parts.find((p) => p.type === "day")?.value ?? "1");
  return new Date(Date.UTC(year, month - 1, day));
}

export function toAmPm(hhmm: string): string {
  const [h, m] = hhmm.split(":").map((v) => Number(v));
  const suffix = h >= 12 ? "PM" : "AM";
  const hour12 = ((h + 11) % 12) + 1;
  return `${hour12}:${String(m).padStart(2, "0")} ${suffix}`;
}

export function weekdayFromDate(date: string): number | null {
  const d = new Date(`${date}T00:00:00Z`);
  if (Number.isNaN(d.getTime())) return null;
  return d.getUTCDay();
}

export function parseDateText(input: string): string | null {
  const raw = input.trim();
  if (!raw) return null;
  if (/^\d{4}-\d{2}-\d{2}$/.test(raw)) return raw;

  const lower = raw.toLowerCase();
  const today = torontoNow();
  const out = new Date(today);

  if (lower === "today" || lower === "今天") {
    return out.toISOString().slice(0, 10);
  }
  if (lower === "tomorrow" || lower === "明天") {
    out.setUTCDate(out.getUTCDate() + 1);
    return out.toISOString().slice(0, 10);
  }
  if (lower === "day after tomorrow" || lower === "后天") {
    out.setUTCDate(out.getUTCDate() + 2);
    return out.toISOString().slice(0, 10);
  }

  const nextDayMatch = lower.match(
    /^(next\s+)?(monday|tuesday|wednesday|thursday|friday|saturday|sunday)$/,
  );
  if (nextDayMatch) {
    const map: Record<string, number> = {
      sunday: 0,
      monday: 1,
      tuesday: 2,
      wednesday: 3,
      thursday: 4,
      friday: 5,
      saturday: 6,
    };
    const target = map[nextDayMatch[2]];
    let delta = target - out.getUTCDay();
    if (delta <= 0 || nextDayMatch[1]) delta += 7;
    out.setUTCDate(out.getUTCDate() + delta);
    return out.toISOString().slice(0, 10);
  }

  const parsed = new Date(raw);
  if (!Number.isNaN(parsed.getTime())) {
    return parsed.toISOString().slice(0, 10);
  }
  return null;
}

export async function getDentistsRaw(): Promise<Array<{ id: number; name: string }>> {
  return await restRequest<Array<{ id: number; name: string }>>(
    "GET",
    "dentists?select=id,name&is_active=eq.true&order=name.asc",
  );
}

export async function getBookedTimes(
  date: string,
  dentistId: number,
): Promise<Set<string>> {
  const rows = await restRequest<Array<{ appointment_time: string }>>(
    "GET",
    `appointments?select=appointment_time&appointment_date=eq.${encodeURIComponent(date)}&dentist_id=eq.${dentistId}&status=eq.confirmed`,
  );
  return new Set(rows.map((r) => r.appointment_time.slice(0, 5)));
}
