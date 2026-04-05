/**
 * Check available appointment slots.
 *
 * In production, this would query your clinic's scheduling system
 * (e.g., Dentrix, Open Dental, Google Calendar, etc.)
 */

interface AvailabilityRequest {
  date: string;
  dentist?: string;
  service?: string;
}

interface TimeSlot {
  time: string;
  dentist: string;
  available: boolean;
}

// Mock schedule — replace with real database/API calls
const MOCK_SCHEDULE: Record<string, TimeSlot[]> = {};

function generateMockSlots(date: string): TimeSlot[] {
  const dentists = ["Dr. Chen", "Dr. Park", "Dr. Sharma"];
  const times = [
    "09:00", "09:30", "10:00", "10:30", "11:00", "11:30",
    "13:00", "13:30", "14:00", "14:30", "15:00", "15:30",
    "16:00", "16:30",
  ];

  const slots: TimeSlot[] = [];
  for (const dentist of dentists) {
    for (const time of times) {
      // Randomly mark ~30% as unavailable for realism
      const seed = (date + dentist + time).split("").reduce((a, c) => a + c.charCodeAt(0), 0);
      slots.push({
        time,
        dentist,
        available: seed % 3 !== 0,
      });
    }
  }
  return slots;
}

export function checkAvailability(params: AvailabilityRequest): string {
  const { date, dentist, service } = params;

  // Validate date
  const requestedDate = new Date(date);
  const today = new Date();
  today.setHours(0, 0, 0, 0);

  if (requestedDate < today) {
    return "That date is in the past. Could you provide a future date?";
  }

  // Check if it's a Sunday
  if (requestedDate.getDay() === 0) {
    return "Sorry, the clinic is closed on Sundays. We're open Monday through Friday 8 AM to 6 PM, and Saturday 9 AM to 2 PM.";
  }

  // Generate or retrieve slots
  if (!MOCK_SCHEDULE[date]) {
    MOCK_SCHEDULE[date] = generateMockSlots(date);
  }

  let slots = MOCK_SCHEDULE[date].filter((s) => s.available);

  // Filter by dentist if specified
  if (dentist) {
    slots = slots.filter((s) => s.dentist === dentist);
  }

  // Saturday has reduced hours
  if (requestedDate.getDay() === 6) {
    slots = slots.filter((s) => {
      const hour = parseInt(s.time.split(":")[0]);
      return hour >= 9 && hour < 14;
    });
  }

  if (slots.length === 0) {
    return `No available slots on ${date}${dentist ? ` with ${dentist}` : ""}. Would you like to try another date?`;
  }

  // Return top 5 slots
  const topSlots = slots.slice(0, 5);
  const slotList = topSlots
    .map((s) => `${s.time} with ${s.dentist}`)
    .join(", ");

  return `Available slots on ${date}: ${slotList}. ${slots.length > 5 ? `(${slots.length - 5} more available)` : ""} Which time works for you?`;
}
