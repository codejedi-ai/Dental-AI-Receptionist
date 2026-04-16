/**
 * `dispatch_dental_action` contract: operation + payload + optional requestId.
 * Normalizes AI-friendly aliases (phone→patientPhone, reason→service, provider→dentist).
 */

export const DISPATCH_SUPPORTED_OPERATIONS = [
  "health_check",
  "get_current_date",
  "get_clinic_info",
  "get_dentists",
  "parse_date",
  "get_next_available_dates",
  "check_availability",
  "book_appointment",
  "cancel_appointment",
  "send_booking_confirmation",
] as const;

export type DispatchOperation = (typeof DISPATCH_SUPPORTED_OPERATIONS)[number];

function asRecord(p: unknown): Record<string, unknown> {
  if (!p || typeof p !== "object") return {};
  return p as Record<string, unknown>;
}

/**
 * Map model-facing field names to what Edge tool handlers expect.
 */
export function normalizeOperationPayload(
  operation: string,
  payload: unknown,
): Record<string, unknown> {
  const p = { ...asRecord(payload) };

  if (operation === "book_appointment") {
    if (p.phone != null && p.patientPhone == null) {
      p.patientPhone = String(p.phone);
    }
    if (p.reason != null && p.service == null) {
      p.service = String(p.reason);
    }
    if (p.provider != null && p.dentist == null) {
      const pv = String(p.provider).toLowerCase();
      p.dentist = pv === "any" ? "" : String(p.provider);
    }
    delete p.phone;
    delete p.reason;
    delete p.provider;
    delete p.timezone;
  }

  if (operation === "check_availability") {
    if (p.provider != null && p.dentist == null) {
      const pv = String(p.provider).toLowerCase();
      p.dentist = pv === "any" ? "" : String(p.provider);
    }
    delete p.provider;
    delete p.timezone;
  }

  return p;
}

/** JSON Schema fragment for Vapi / OpenAI `dispatch_dental_action` tool (source of truth for copy). */
export const DISPATCH_DENTAL_ACTION_PARAMETER_SCHEMA = {
  type: "object",
  properties: {
    operation: {
      type: "string",
      enum: [...DISPATCH_SUPPORTED_OPERATIONS],
      description:
        "Backend operation to run. Use payload: {} when no fields are needed.",
    },
    payload: {
      type: "object",
      description:
        "Operation-specific fields. Examples: get_clinic_info → { topic: \"hours\" }; check_availability → { date, dentist? }; book_appointment → { patientName, patientPhone|phone, date, time, dentist, service|reason }; cancel_appointment → { patientName, date }; parse_date → { date_text }.",
    },
    requestId: {
      type: "string",
      description:
        "Optional idempotency key for retries (logged server-side; safe to omit).",
    },
  },
  required: ["operation"],
} as const;
