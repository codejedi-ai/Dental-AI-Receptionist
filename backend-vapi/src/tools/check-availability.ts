import { getAvailableSlots } from "../services/calendar";
import { isSlotTaken } from "../services/appointment-store";

interface AvailabilityRequest {
  date: string;
  dentist?: string;
  service?: string;
}

export async function checkAvailability(params: AvailabilityRequest): Promise<string> {
  const { date, dentist } = params;

  const requestedDate = new Date(date + "T00:00:00");
  const today = new Date();
  today.setHours(0, 0, 0, 0);

  if (requestedDate < today) {
    return "That date is in the past. Could you provide a future date?";
  }
  if (requestedDate.getDay() === 0) {
    return "Sorry, the clinic is closed on Sundays. We're open Monday–Friday 8 AM–6 PM and Saturday 9 AM–2 PM.";
  }

  const rawSlots = await getAvailableSlots(date, dentist);

  // Filter out any slots already booked in the local store
  const slots = rawSlots.filter(s => !isSlotTaken(date, s.time, s.dentist));

  if (slots.length === 0) {
    return `No available slots on ${date}${dentist ? ` with ${dentist}` : ""}. Would you like to try another date?`;
  }

  const top = slots.slice(0, 5);
  const list = top.map((s) => `${s.time} with ${s.dentist}`).join(", ");
  const more = slots.length > 5 ? ` (${slots.length - 5} more available)` : "";

  return `Available slots on ${date}: ${list}.${more} Which time works for you?`;
}
