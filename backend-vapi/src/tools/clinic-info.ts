/**
 * Clinic information lookup.
 * This is mostly handled by the system prompt, but having it as a tool
 * gives the LLM a structured way to fetch specific info.
 */

interface InfoRequest {
  topic: string;
}

const CLINIC_INFO: Record<string, string> = {
  hours:
    "We're open Monday through Friday 8 AM to 6 PM, and Saturday 9 AM to 2 PM. Closed on Sundays and statutory holidays.",
  location:
    "We're at 123 Main Street, Newmarket, Ontario, L3Y 4Z1. We have free parking in the lot behind the building.",
  insurance:
    "We accept most major dental insurance plans including Sun Life, Manulife, Great-West Life, Blue Cross, and Green Shield. Please bring your insurance card to your appointment.",
  emergency:
    "For after-hours dental emergencies like severe pain, a knocked-out tooth, or heavy bleeding, please go to Southlake Regional Health Centre Emergency Department at 596 Davis Drive, Newmarket.",
  cancellation:
    "We ask for at least 24 hours notice for cancellations. Late cancellations or no-shows may be subject to a fee.",
  newpatient:
    "New patients are always welcome! For your first visit, please arrive 15 minutes early to complete paperwork. Bring your insurance card, a list of current medications, and any recent dental X-rays if you have them.",
  payment:
    "We accept cash, debit, Visa, Mastercard, and e-transfer. We also offer payment plans for larger treatments.",
};

export function getClinicInfo(params: InfoRequest): string {
  const topic = params.topic?.toLowerCase() || "";

  // Try exact match first
  if (CLINIC_INFO[topic]) {
    return CLINIC_INFO[topic];
  }

  // Try fuzzy match
  for (const [key, value] of Object.entries(CLINIC_INFO)) {
    if (topic.includes(key) || key.includes(topic)) {
      return value;
    }
  }

  return "I don't have specific information about that. I can help with hours, location, insurance, emergency info, cancellation policy, new patient info, or payment options. What would you like to know?";
}
