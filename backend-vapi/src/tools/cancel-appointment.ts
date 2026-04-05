/**
 * Cancel an existing appointment.
 */

import { getBookings } from "./book-appointment";

interface CancelRequest {
  patientName: string;
  patientPhone?: string;
  date: string;
}

export function cancelAppointment(params: CancelRequest): string {
  const { patientName, date } = params;

  // In production: query your scheduling system
  // For demo: check in-memory bookings
  const bookings = getBookings();
  const found = bookings.find(
    (b) =>
      b.patientName.toLowerCase() === patientName.toLowerCase() &&
      b.date === date
  );

  if (!found) {
    return `I couldn't find an appointment for ${patientName} on ${date}. Could you double-check the name and date? Or I can look up by phone number.`;
  }

  // In production: actually delete from the system
  console.log(`❌ CANCELLED: ${patientName} | ${date} ${found.time} | ${found.dentist}`);

  return `I've cancelled ${patientName}'s ${found.service} appointment with ${found.dentist} on ${date} at ${found.time}. Would you like to rebook for a different time?`;
}
