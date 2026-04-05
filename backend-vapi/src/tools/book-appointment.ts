/**
 * Book an appointment.
 *
 * In production, this would write to your clinic's scheduling system
 * and send confirmation SMS/email to the patient.
 */

interface BookingRequest {
  patientName: string;
  patientPhone: string;
  date: string;
  time: string;
  dentist: string;
  service: string;
  notes?: string;
}

// In-memory store for demo — replace with database
const bookings: BookingRequest[] = [];

export function bookAppointment(params: BookingRequest): string {
  const { patientName, patientPhone, date, time, dentist, service, notes } = params;

  // Basic validation
  if (!patientName || !date || !time || !dentist) {
    return "I'm missing some details. Could you confirm the patient name, date, time, and dentist?";
  }

  // Check for duplicate
  const duplicate = bookings.find(
    (b) => b.patientName === patientName && b.date === date && b.time === time
  );
  if (duplicate) {
    return `It looks like ${patientName} already has an appointment on ${date} at ${time}. Would you like to reschedule instead?`;
  }

  // Store the booking
  bookings.push(params);

  console.log(`📅 BOOKED: ${patientName} | ${date} ${time} | ${dentist} | ${service} | ${notes || "no notes"}`);

  return `Great! I've booked ${patientName} for a ${service} appointment with ${dentist} on ${date} at ${time}. We'll send a confirmation text to ${patientPhone}. Please arrive 15 minutes early and bring your insurance card. Is there anything else I can help with?`;
}

export function getBookings(): BookingRequest[] {
  return [...bookings];
}
