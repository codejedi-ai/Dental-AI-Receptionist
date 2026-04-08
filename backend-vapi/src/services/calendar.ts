import { google } from "googleapis";
import dotenv from "dotenv";
dotenv.config();

// ─────────────────────────────────────────────────────────────────
// Auth — OAuth2 with stored refresh token
// Run `npm run auth:google` once to generate GOOGLE_REFRESH_TOKEN
// ─────────────────────────────────────────────────────────────────
function getAuthClient() {
  const client = new google.auth.OAuth2(
    process.env.GOOGLE_CLIENT_ID,
    process.env.GOOGLE_CLIENT_SECRET,
    "urn:ietf:wg:oauth:2.0:oob"  // desktop / out-of-band redirect
  );
  client.setCredentials({ refresh_token: process.env.GOOGLE_REFRESH_TOKEN });
  return client;
}

const CALENDAR_ID = process.env.GOOGLE_CALENDAR_ID || "primary";
const DENTISTS = ["Dr. Sarah Chen", "Dr. Michael Park", "Dr. Priya Sharma"];

// Clinic hours
const SLOT_DURATION_MINS = 30;
const WEEKDAY_SLOTS = generateTimes("08:00", "18:00", SLOT_DURATION_MINS);
const SATURDAY_SLOTS = generateTimes("09:00", "14:00", SLOT_DURATION_MINS);

function generateTimes(start: string, end: string, stepMins: number): string[] {
  const times: string[] = [];
  const [sh, sm] = start.split(":").map(Number);
  const [eh, em] = end.split(":").map(Number);
  let current = sh * 60 + sm;
  const endMins = eh * 60 + em;
  while (current < endMins) {
    const h = Math.floor(current / 60).toString().padStart(2, "0");
    const m = (current % 60).toString().padStart(2, "0");
    times.push(`${h}:${m}`);
    current += stepMins;
  }
  return times;
}

export interface Slot {
  time: string;
  dentist: string;
  dateTime: string; // ISO
}

// ─────────────────────────────────────────────────────────────────
// Check availability — queries Google Calendar for busy times
// ─────────────────────────────────────────────────────────────────
export async function getAvailableSlots(
  date: string,
  dentistFilter?: string,
): Promise<Slot[]> {
  if (!process.env.GOOGLE_CLIENT_ID) {
    // Fall back to mock if Calendar not configured yet
    return getMockSlots(date, dentistFilter);
  }

  const dayDate = new Date(date + "T00:00:00");
  const dow = dayDate.getDay();
  if (dow === 0) return []; // closed Sunday

  const slotTimes = dow === 6 ? SATURDAY_SLOTS : WEEKDAY_SLOTS;
  const dentists = dentistFilter ? [dentistFilter] : DENTISTS;

  try {
    const auth = getAuthClient();
    const calendar = google.calendar({ version: "v3", auth });

    // Query busy periods for the whole day
    const dayStart = new Date(date + "T00:00:00-05:00").toISOString();
    const dayEnd   = new Date(date + "T23:59:59-05:00").toISOString();

    const freeBusy = await calendar.freebusy.query({
      requestBody: {
        timeMin: dayStart,
        timeMax: dayEnd,
        items: [{ id: CALENDAR_ID }],
      },
    });

    const busyPeriods = freeBusy.data.calendars?.[CALENDAR_ID]?.busy ?? [];

    const available: Slot[] = [];

    for (const dentist of dentists) {
      for (const time of slotTimes) {
        const slotStart = new Date(`${date}T${time}:00-05:00`);
        const slotEnd   = new Date(slotStart.getTime() + SLOT_DURATION_MINS * 60_000);

        // Check if this slot overlaps with any busy period
        const isBusy = busyPeriods.some((b: { start?: string | null; end?: string | null }) => {
          const bs = new Date(b.start!);
          const be = new Date(b.end!);
          return slotStart < be && slotEnd > bs;
        });

        if (!isBusy) {
          available.push({ time, dentist, dateTime: slotStart.toISOString() });
        }
      }
    }

    return available;
  } catch (err) {
    console.error("Calendar API error:", err);
    // Degrade gracefully to mock
    return getMockSlots(date, dentistFilter);
  }
}

// ─────────────────────────────────────────────────────────────────
// Book — creates a Google Calendar event
// ─────────────────────────────────────────────────────────────────
export interface BookingDetails {
  patientName: string;
  patientPhone: string;
  patientEmail?: string;
  date: string;
  time: string;
  dentist: string;
  service: string;
  notes?: string;
}

export async function createCalendarEvent(b: BookingDetails): Promise<string | null> {
  if (!process.env.GOOGLE_CLIENT_ID) {
    console.log("📅 [MOCK] Booking saved (no Calendar configured):", b);
    return null; // no event ID in mock mode
  }

  const auth = getAuthClient();
  const calendar = google.calendar({ version: "v3", auth });

  const startDT = new Date(`${b.date}T${b.time}:00-05:00`);
  const endDT   = new Date(startDT.getTime() + SLOT_DURATION_MINS * 60_000);

  const attendees = b.patientEmail ? [{ email: b.patientEmail }] : [];

  const event = await calendar.events.insert({
    calendarId: CALENDAR_ID,
    requestBody: {
      summary: `${b.service} — ${b.patientName}`,
      description: [
        `Patient: ${b.patientName}`,
        `Phone: ${b.patientPhone}`,
        `Service: ${b.service}`,
        b.notes ? `Notes: ${b.notes}` : "",
      ].filter(Boolean).join("\n"),
      start: { dateTime: startDT.toISOString(), timeZone: "America/Toronto" },
      end:   { dateTime: endDT.toISOString(),   timeZone: "America/Toronto" },
      attendees,
      location: "123 Main Street, Newmarket, ON L3Y 4Z1",
    },
    sendUpdates: b.patientEmail ? "all" : "none",
  });

  console.log(`📅 Calendar event created: ${event.data.id}`);
  return event.data.id ?? null;
}

// ─────────────────────────────────────────────────────────────────
// Cancel — deletes the Google Calendar event
// ─────────────────────────────────────────────────────────────────
export async function deleteCalendarEvent(eventId: string): Promise<void> {
  if (!process.env.GOOGLE_CLIENT_ID) return;
  const auth = getAuthClient();
  const calendar = google.calendar({ version: "v3", auth });
  await calendar.events.delete({ calendarId: CALENDAR_ID, eventId });
  console.log(`📅 Calendar event deleted: ${eventId}`);
}

// ─────────────────────────────────────────────────────────────────
// Find event by patient name + date (for cancellations)
// ─────────────────────────────────────────────────────────────────
export async function findEvent(patientName: string, date: string) {
  if (!process.env.GOOGLE_CLIENT_ID) return null;
  const auth = getAuthClient();
  const calendar = google.calendar({ version: "v3", auth });

  const dayStart = new Date(date + "T00:00:00-05:00").toISOString();
  const dayEnd   = new Date(date + "T23:59:59-05:00").toISOString();

  const events = await calendar.events.list({
    calendarId: CALENDAR_ID,
    timeMin: dayStart,
    timeMax: dayEnd,
    q: patientName,
    singleEvents: true,
  });

  return events.data.items?.[0] ?? null;
}

// ─────────────────────────────────────────────────────────────────
// Mock fallback (used when GOOGLE_CLIENT_ID not set)
// ─────────────────────────────────────────────────────────────────
function getMockSlots(date: string, dentistFilter?: string): Slot[] {
  const dayDate = new Date(date + "T00:00:00");
  const dow = dayDate.getDay();
  if (dow === 0) return [];

  const slotTimes = dow === 6 ? SATURDAY_SLOTS : WEEKDAY_SLOTS;
  const dentists = dentistFilter ? [dentistFilter] : DENTISTS;
  const slots: Slot[] = [];

  for (const dentist of dentists) {
    for (const time of slotTimes) {
      const seed = (date + dentist + time)
        .split("").reduce((a, c) => a + c.charCodeAt(0), 0);
      if (seed % 3 !== 0) {
        slots.push({
          time,
          dentist,
          dateTime: new Date(`${date}T${time}:00-05:00`).toISOString(),
        });
      }
    }
  }
  return slots;
}
